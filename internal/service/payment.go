package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/config"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

const (
	PaymentProviderSelcom     = "selcom"
	PaymentProviderSelcomMock = "selcom_mock"
	PaymentStatusPending      = "pending"
	PaymentStatusCompleted    = "completed"
	PaymentStatusCancelled    = "cancelled"
	PaymentStatusFailed       = "failed"
	PaymentStatusExpired      = "expired"
	PaymentStatusRefunded     = "refunded"
	PaymentStatusPartialRefund = "partially_refunded"
	DemoPaymentDisclaimer     = "Demo payment — no money charged."
)

type PaymentService struct {
	cfg       config.SelcomConfig
	selcom    selcom.Client
	orders    *repository.OrderRepository
	hotels    *repository.HotelRepository
	discovery *repository.DiscoveryRepository
	events    *EventService
	enabled   bool
}

func NewPaymentService(
	cfg config.SelcomConfig,
	selcomClient selcom.Client,
	orders *repository.OrderRepository,
	hotels *repository.HotelRepository,
	discovery *repository.DiscoveryRepository,
	events *EventService,
) *PaymentService {
	enabled := cfg.MockMode || (cfg.APIKey != "" && cfg.APISecret != "" && cfg.Vendor != "")
	return &PaymentService{
		cfg:       cfg,
		selcom:    selcomClient,
		orders:    orders,
		hotels:    hotels,
		discovery: discovery,
		events:    events,
		enabled:   enabled,
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

	if s.isDemo() {
		input.Order.PaymentProvider = PaymentProviderSelcomMock
		input.Order.PaymentIsDemo = true
	} else {
		input.Order.PaymentProvider = PaymentProviderSelcom
		input.Order.PaymentIsDemo = false
	}
	input.Order.PaymentStatus = PaymentStatusPending
	input.Order.PaymentRef = result.Reference
	if err := s.orders.Update(input.Order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save payment session", apperrors.ErrInternal.Status)
	}

	return &CheckoutSession{
		PaymentRequired: true,
		PaymentURL:      result.PaymentGatewayURL,
		PaymentStatus:   PaymentStatusPending,
		PaymentProvider: input.Order.PaymentProvider,
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
	if order.PaymentStatus == PaymentStatusCompleted && strings.EqualFold(payload.PaymentStatus, "COMPLETED") {
		return nil
	}
	if orderNoLongerPayable(order) {
		return apperrors.New(apperrors.ErrBadRequest.Code, "order is no longer payable", apperrors.ErrBadRequest.Status)
	}
	if payload.Amount != "" {
		if amount, ok := selcom.ParseWebhookAmount(payload.Amount); ok && amount != order.TotalAmount {
			return apperrors.New(apperrors.ErrBadRequest.Code, "payment amount mismatch", apperrors.ErrBadRequest.Status)
		}
	}
	if order.PaymentRef != "" && payload.Reference != "" && order.PaymentRef != payload.Reference {
		return apperrors.New(apperrors.ErrBadRequest.Code, "payment reference mismatch", apperrors.ErrBadRequest.Status)
	}

	paymentStatus, orderStatus, releaseStock := webhookPaymentOutcome(payload.PaymentStatus)
	order.PaymentStatus = paymentStatus
	if orderStatus != "" {
		order.Status = models.OrderStatus(orderStatus)
	}
	if paymentStatus == PaymentStatusCompleted && payload.Reference != "" {
		order.PaymentRef = payload.Reference
	}
	if releaseStock {
		s.releaseHeldStock(order)
	}

	if err := s.orders.Update(order); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order payment", apperrors.ErrInternal.Status)
	}

	if s.events != nil {
		if order.PaymentStatus == PaymentStatusCompleted && previousStatus == string(models.OrderStatusPending) {
			s.finalizeDiscoveryTickets(order)
			s.events.OrderCreated(order, hotel)
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
	if orderNoLongerPayable(order) {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "order is no longer payable", apperrors.ErrBadRequest.Status)
	}

	hotel, err := s.hotels.FindByID(order.HotelID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	previousStatus := string(order.Status)
	order.PaymentStatus = PaymentStatusCompleted
	order.Status = models.OrderStatusConfirmed
	order.PaymentIsDemo = true
	if order.PaymentProvider == "" || order.PaymentProvider == PaymentProviderSelcom {
		order.PaymentProvider = PaymentProviderSelcomMock
	}
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order", apperrors.ErrInternal.Status)
	}
	if s.events != nil && previousStatus == string(models.OrderStatusPending) {
		s.finalizeDiscoveryTickets(order)
		s.events.OrderCreated(order, hotel)
	}
	return order, nil
}

func (s *PaymentService) finalizeDiscoveryTickets(order *models.Order) {
	if s.discovery == nil || order == nil || !strings.Contains(order.Notes, "discovery_ticket:") {
		return
	}
	var itemID uuid.UUID
	qty := 1
	for _, part := range strings.Split(order.Notes, "|") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "item:") {
			if id, err := uuid.Parse(strings.TrimPrefix(part, "item:")); err == nil {
				itemID = id
			}
		}
		if strings.HasPrefix(part, "qty:") {
			if n, err := strconv.Atoi(strings.TrimPrefix(part, "qty:")); err == nil && n > 0 {
				qty = n
			}
		}
	}
	if itemID == uuid.Nil {
		return
	}
	_ = s.discovery.IncrementTicketsSold(itemID, qty)
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderID uuid.UUID) (string, string, bool, error) {
	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", false, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return "", "", false, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	return order.PaymentStatus, string(order.Status), order.PaymentIsDemo || s.isDemo(), nil
}

func (s *PaymentService) isDemo() bool {
	return s.cfg.MockMode || strings.TrimSpace(s.cfg.APIKey) == "" || strings.TrimSpace(s.cfg.APISecret) == "" || strings.TrimSpace(s.cfg.Vendor) == ""
}

type RefundResult struct {
	OrderID        string
	Amount         int64
	RefundedTotal  int64
	PaymentStatus  string
	Idempotent     bool
	IsDemo         bool
	Disclaimer     string
	Order          *models.Order `json:"-"`
}

func (s *PaymentService) ApplyRefund(orderID uuid.UUID, amount int64, key, reason string) (*RefundResult, error) {
	if amount <= 0 {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "refund amount must be greater than zero", apperrors.ErrValidation.Status)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "idempotency key is required", apperrors.ErrValidation.Status)
	}

	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	if order.PaymentStatus != PaymentStatusCompleted && order.PaymentStatus != PaymentStatusPartialRefund {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "only completed payments can be refunded", apperrors.ErrBadRequest.Status)
	}

	var existing models.PaymentRefund
	if err := s.orders.DB().Where("idempotency_key = ?", key).First(&existing).Error; err == nil {
		return &RefundResult{
			OrderID:       order.ID.String(),
			Amount:        existing.Amount,
			RefundedTotal: order.RefundedAmount,
			PaymentStatus: order.PaymentStatus,
			Idempotent:    true,
			IsDemo:        existing.IsDemo,
			Disclaimer:    DemoPaymentDisclaimer,
			Order:         order,
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load refund", apperrors.ErrInternal.Status)
	}

	remaining := order.TotalAmount - order.RefundedAmount
	if amount > remaining {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "refund exceeds remaining captured amount", apperrors.ErrBadRequest.Status)
	}

	isDemo := order.PaymentIsDemo || s.isDemo()
	refund := &models.PaymentRefund{
		OrderID:        order.ID,
		Amount:         amount,
		IdempotencyKey: key,
		Reason:         reason,
		IsDemo:         isDemo,
		Status:         "completed",
	}
	order.RefundedAmount += amount
	if order.RefundedAmount >= order.TotalAmount {
		order.PaymentStatus = PaymentStatusRefunded
	} else {
		order.PaymentStatus = PaymentStatusPartialRefund
	}

	if err := s.orders.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(refund).Error; err != nil {
			return err
		}
		return tx.Save(order).Error
	}); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to record refund", apperrors.ErrInternal.Status)
	}

	return &RefundResult{
		OrderID:       order.ID.String(),
		Amount:        amount,
		RefundedTotal: order.RefundedAmount,
		PaymentStatus: order.PaymentStatus,
		IsDemo:        isDemo,
		Disclaimer:    DemoPaymentDisclaimer,
		Order:         order,
	}, nil
}

func (s *PaymentService) CreateDemoReservation(itemID uuid.UUID, qty int, alreadyExpired bool) (*models.InventoryReservation, error) {
	if qty <= 0 {
		qty = 1
	}
	expires := time.Now().Add(15 * time.Minute)
	if alreadyExpired {
		expires = time.Now().Add(-1 * time.Minute)
	}
	res := &models.InventoryReservation{
		CatalogueItemID: itemID,
		Quantity:        qty,
		Status:          models.InventoryReservationStatusActive,
		ReservedAt:      time.Now(),
		ExpiresAt:       expires,
	}
	if err := s.orders.DB().Create(res).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create demo reservation", apperrors.ErrInternal.Status)
	}
	return res, nil
}

func (s *PaymentService) CountActiveReservations(itemID uuid.UUID) (int64, error) {
	var n int64
	err := s.orders.DB().Model(&models.InventoryReservation{}).
		Where("catalogue_item_id = ? AND status = ?", itemID, models.InventoryReservationStatusActive).
		Count(&n).Error
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count reservations", apperrors.ErrInternal.Status)
	}
	return n, nil
}

func (s *PaymentService) ExpireStaleReservations() (int, error) {
	res := s.orders.DB().Model(&models.InventoryReservation{}).
		Where("status = ? AND expires_at < ?", models.InventoryReservationStatusActive, time.Now()).
		Update("status", models.InventoryReservationStatusExpired)
	if res.Error != nil {
		return 0, apperrors.Wrap(res.Error, apperrors.ErrInternal.Code, "failed to expire reservations", apperrors.ErrInternal.Status)
	}
	return int(res.RowsAffected), nil
}

func (s *PaymentService) ExpireHoldNow(orderID uuid.UUID) (*models.Order, error) {
	order, err := s.orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	if order.Status != models.OrderStatusPending || !order.StockHeld {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "order has no unpaid stock hold", apperrors.ErrBadRequest.Status)
	}
	s.releaseHeldStock(order)
	order.Status = models.OrderStatusCancelled
	order.PaymentStatus = PaymentStatusExpired
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to expire hold", apperrors.ErrInternal.Status)
	}
	return order, nil
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

func orderNoLongerPayable(order *models.Order) bool {
	if order == nil {
		return true
	}
	switch order.PaymentStatus {
	case PaymentStatusCancelled, PaymentStatusFailed, PaymentStatusExpired, PaymentStatusRefunded:
		return true
	}
	switch order.Status {
	case models.OrderStatusCancelled, models.OrderStatusFailed:
		return true
	}
	return false
}

func unpaidHoldCutoff(now time.Time) time.Time {
	return now.Add(-15 * time.Minute)
}

func webhookPaymentOutcome(status string) (paymentStatus string, orderStatus models.OrderStatus, releaseStock bool) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "COMPLETED":
		return PaymentStatusCompleted, models.OrderStatusConfirmed, false
	case "CANCELLED", "USERCANCELED":
		return PaymentStatusCancelled, models.OrderStatusCancelled, true
	case "FAILED", "REJECTED", "EXPIRED":
		return PaymentStatusFailed, models.OrderStatusFailed, true
	default:
		return PaymentStatusPending, "", false
	}
}

func fallbackEmail(email string) string {
	email = strings.TrimSpace(email)
	if email != "" {
		return email
	}
	return "orders@necha.africa"
}

func (s *PaymentService) releaseHeldStock(order *models.Order) {
	if order == nil || !order.StockHeld || s.hotels == nil {
		return
	}
	for _, item := range order.Items {
		if item.ProductID == nil || item.Quantity <= 0 {
			continue
		}
		_ = s.hotels.RestoreProductStock(*item.ProductID, item.Quantity)
	}
	order.StockHeld = false
}
