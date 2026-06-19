package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

const (
	PaymentProviderSelcom   = "selcom"
	PaymentStatusPending    = "pending"
	PaymentStatusCompleted  = "completed"
	PaymentStatusCancelled  = "cancelled"
	PaymentStatusFailed     = "failed"
)

type PaymentService struct {
	cfg     config.SelcomConfig
	selcom  selcom.Client
	orders  *repository.OrderRepository
	hotels  *repository.HotelRepository
	events  *EventService
	enabled bool
}

func NewPaymentService(
	cfg config.SelcomConfig,
	selcomClient selcom.Client,
	orders *repository.OrderRepository,
	hotels *repository.HotelRepository,
	events *EventService,
) *PaymentService {
	enabled := cfg.MockMode || (cfg.APIKey != "" && cfg.APISecret != "" && cfg.Vendor != "")
	return &PaymentService{
		cfg:     cfg,
		selcom:  selcomClient,
		orders:  orders,
		hotels:  hotels,
		events:  events,
		enabled: enabled,
	}
}

func (s *PaymentService) Enabled() bool {
	return s.enabled
}

type CheckoutInput struct {
	Order         *models.Order
	HotelName     string
	BuyerEmail    string
	BuyerName     string
	BuyerPhone    string
	ReturnURL     string
	CancelURL     string
	BuyerRemarks  string
	ItemCount     int
}

type CheckoutSession struct {
	PaymentRequired bool
	PaymentURL      string
	PaymentStatus   string
	PaymentProvider string
}

func (s *PaymentService) StartCheckout(ctx context.Context, input CheckoutInput) (*CheckoutSession, error) {
	if !s.enabled {
		return &CheckoutSession{PaymentRequired: false}, nil
	}

	returnURL := withOrderID(input.ReturnURL, input.Order.ID.String())
	if returnURL == "" {
		returnURL = strings.TrimRight(s.cfg.PublicAppURL, "/") + "/payment/return?order_id=" + input.Order.ID.String()
	}
	cancelURL := withOrderID(input.CancelURL, input.Order.ID.String())
	if cancelURL == "" {
		cancelURL = strings.TrimRight(s.cfg.PublicAppURL, "/") + "/payment/cancel?order_id=" + input.Order.ID.String()
	}
	webhookURL := strings.TrimRight(s.cfg.PublicAPIURL, "/") + "/api/webhooks/selcom"

	itemCount := input.ItemCount
	if itemCount <= 0 {
		itemCount = len(input.Order.Items)
	}
	if itemCount <= 0 {
		itemCount = 1
	}

	result, err := s.selcom.CreateOrderMinimal(ctx, selcom.CreateOrderMinimalInput{
		OrderID:         input.Order.ID.String(),
		BuyerEmail:      fallbackEmail(input.BuyerEmail),
		BuyerName:       input.BuyerName,
		BuyerPhone:      input.BuyerPhone,
		Amount:          input.Order.TotalAmount,
		Currency:        input.Order.Currency,
		RedirectURL:     returnURL,
		CancelURL:       cancelURL,
		WebhookURL:      webhookURL,
		BuyerRemarks:    input.BuyerRemarks,
		MerchantRemarks: fmt.Sprintf("Necha order %s", input.Order.ID.String()[:8]),
		NoOfItems:       itemCount,
	})
	if err != nil {
		return nil, err
	}

	input.Order.PaymentProvider = PaymentProviderSelcom
	input.Order.PaymentStatus = PaymentStatusPending
	input.Order.PaymentRef = result.Reference
	if err := s.orders.Update(input.Order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save payment session", apperrors.ErrInternal.Status)
	}

	return &CheckoutSession{
		PaymentRequired: true,
		PaymentURL:      result.PaymentGatewayURL,
		PaymentStatus:   PaymentStatusPending,
		PaymentProvider: PaymentProviderSelcom,
	}, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload selcom.WebhookPayload) error {
	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status)
	}

	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	hotel, err := s.hotels.FindByID(order.HotelID)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	previousStatus := string(order.Status)
	if payload.Amount != "" {
		if amount, ok := selcom.ParseWebhookAmount(payload.Amount); ok && amount != order.TotalAmount {
			return apperrors.New(apperrors.ErrBadRequest.Code, "payment amount mismatch", apperrors.ErrBadRequest.Status)
		}
	}
	if order.PaymentRef != "" && payload.Reference != "" && order.PaymentRef != payload.Reference {
		return apperrors.New(apperrors.ErrBadRequest.Code, "payment reference mismatch", apperrors.ErrBadRequest.Status)
	}

	switch strings.ToUpper(payload.PaymentStatus) {
	case "COMPLETED":
		order.PaymentStatus = PaymentStatusCompleted
		order.Status = models.OrderStatusConfirmed
		if payload.Reference != "" {
			order.PaymentRef = payload.Reference
		}
	case "CANCELLED", "USERCANCELED":
		order.PaymentStatus = PaymentStatusCancelled
		order.Status = models.OrderStatusCancelled
	default:
		order.PaymentStatus = PaymentStatusPending
	}

	if err := s.orders.Update(order); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order payment", apperrors.ErrInternal.Status)
	}

	if s.events != nil {
		if order.PaymentStatus == PaymentStatusCompleted && previousStatus == string(models.OrderStatusPending) {
			s.events.OrderCreated(order, hotel.Name)
		} else if previousStatus != string(order.Status) {
			s.events.OrderStatusUpdated(order, hotel.Name, previousStatus)
		}
	}
	return nil
}

func (s *PaymentService) CompleteMockPayment(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	if order.PaymentStatus == PaymentStatusCompleted {
		return order, nil
	}

	hotel, err := s.hotels.FindByID(order.HotelID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	previousStatus := string(order.Status)
	order.PaymentStatus = PaymentStatusCompleted
	order.Status = models.OrderStatusConfirmed
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order", apperrors.ErrInternal.Status)
	}
	if s.events != nil && previousStatus == string(models.OrderStatusPending) {
		s.events.OrderCreated(order, hotel.Name)
	}
	return order, nil
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderID uuid.UUID) (string, string, error) {
	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return "", "", apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	return order.PaymentStatus, string(order.Status), nil
}

func withOrderID(raw, orderID string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "PLACEHOLDER", orderID)
	if strings.Contains(raw, "order_id=") {
		return raw
	}
	sep := "?"
	if strings.Contains(raw, "?") {
		sep = "&"
	}
	return raw + sep + "order_id=" + orderID
}

func fallbackEmail(email string) string {
	email = strings.TrimSpace(email)
	if email != "" {
		return email
	}
	return "orders@necha.africa"
}
