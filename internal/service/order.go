package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/integration/kkooapp"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type OrderService struct {
	hotels     *repository.HotelRepository
	orders     *repository.OrderRepository
	kkooapp    kkooapp.Client
	events     *EventService
	payments   *PaymentService
	guestStays *GuestStayService
}

func NewOrderService(
	hotels *repository.HotelRepository,
	orders *repository.OrderRepository,
	kkooappClient kkooapp.Client,
	events *EventService,
	payments *PaymentService,
	guestStays *GuestStayService,
) *OrderService {
	return &OrderService{
		hotels:     hotels,
		orders:     orders,
		kkooapp:    kkooappClient,
		events:     events,
		payments:   payments,
		guestStays: guestStays,
	}
}

func (s *OrderService) CreateFoodOrder(ctx context.Context, req dto.FoodOrderRequest, userID *uuid.UUID) (*dto.OrderResponse, error) {
	hotel, err := s.hotels.FindByCode(req.HotelCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	if hotel.KkooappID == "" {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "hotel is not linked to kkooapp", apperrors.ErrBadRequest.Status)
	}

	var total int64
	items := make([]models.OrderItem, 0, len(req.Items))
	kkItems := make([]kkooapp.FoodOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		lineTotal := item.UnitPrice * int64(item.Quantity)
		total += lineTotal
		items = append(items, models.OrderItem{
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: lineTotal,
			Notes:      item.Notes,
		})
		kkItems = append(kkItems, kkooapp.FoodOrderItem{
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Notes:     item.Notes,
		})
	}

	order := &models.Order{
		HotelID:       hotel.ID,
		UserID:        userID,
		Type:          models.OrderTypeFood,
		Status:        models.OrderStatusPending,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		TableNumber:   req.TableNumber,
		RoomNumber:    req.RoomNumber,
		TotalAmount:   total,
		Currency:      "TZS",
		Items:         items,
		Notes:         req.Notes,
	}

	if err := s.orders.Create(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create order", apperrors.ErrInternal.Status)
	}

	kkResp, err := s.kkooapp.PlaceFoodOrder(ctx, kkooapp.FoodOrderRequest{
		PropertyID:    hotel.KkooappID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		TableNumber:   req.TableNumber,
		RoomNumber:    req.RoomNumber,
		Items:         kkItems,
		Notes:         req.Notes,
		ExternalRef:   order.ID.String(),
	})
	if err != nil {
		order.Status = models.OrderStatusFailed
		order.Notes = err.Error()
		_ = s.orders.Update(order)
		return nil, err
	}

	order.KkooappRef = kkResp.Reference
	order.Status = mapKkooappOrderStatus(kkResp.Status)
	if kkResp.TotalAmount > 0 {
		order.TotalAmount = kkResp.TotalAmount
	}
	if kkResp.Currency != "" {
		order.Currency = kkResp.Currency
	}
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order", apperrors.ErrInternal.Status)
	}

	loaded, err := s.orders.FindByID(order.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	if s.events != nil {
		s.events.OrderCreated(loaded, hotel.Name)
	}
	if s.guestStays != nil {
		s.guestStays.RecordFromOrder(loaded, referralFromNotes(req.Notes), models.GuestStaySourceFoodOrder)
	}
	return toOrderResponse(loaded), nil
}

func (s *OrderService) CreateProductOrder(ctx context.Context, req dto.ProductOrderRequest, userID *uuid.UUID) (*dto.OrderResponse, error) {
	hotelCode := req.HotelCode
	if hotelCode == "" {
		hotelCode = "SEACLIFF24"
	}

	hotel, err := s.hotels.FindByCode(hotelCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	currency := req.Currency
	if currency == "" {
		currency = "TZS"
	}

	type pricedItem struct {
		productID  uuid.UUID
		name       string
		quantity   int
		unitPrice  int64
		lineTotal  int64
		trackStock bool
	}
	priced := make([]pricedItem, 0, len(req.Items))
	var total int64
	itemCount := 0

	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrValidation.Code, "invalid product_id", apperrors.ErrValidation.Status)
		}
		product, err := s.hotels.FindActiveProductForHotel(hotel.ID, productID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.New(apperrors.ErrNotFound.Code, "product not found", apperrors.ErrNotFound.Status)
			}
			return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load product", apperrors.ErrInternal.Status)
		}
		trackStock := product.Stock > 0
		if trackStock && item.Quantity > product.Stock {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "insufficient stock for "+product.Name, apperrors.ErrBadRequest.Status)
		}
		if product.Currency != "" {
			currency = product.Currency
		}
		lineTotal := product.Price * int64(item.Quantity)
		total += lineTotal
		itemCount += item.Quantity
		priced = append(priced, pricedItem{
			productID:  productID,
			name:       product.Name,
			quantity:   item.Quantity,
			unitPrice:  product.Price,
			lineTotal:  lineTotal,
			trackStock: trackStock,
		})
	}
	if req.DeliveryFee > 0 {
		total += req.DeliveryFee
	}

	notes := req.Notes
	meta := []string{}
	if req.CustomerEmail != "" {
		meta = append(meta, "email:"+req.CustomerEmail)
	}
	if req.Address != "" {
		meta = append(meta, "address:"+req.Address)
	}
	if req.City != "" {
		meta = append(meta, "city:"+req.City)
	}
	if req.Country != "" {
		meta = append(meta, "country:"+req.Country)
	}
	if req.PaymentMethod != "" {
		meta = append(meta, "payment:"+req.PaymentMethod)
	}
	if len(meta) > 0 {
		if notes != "" {
			notes += " | "
		}
		notes += strings.Join(meta, " | ")
	}

	orderStatus := models.OrderStatusConfirmed
	if s.payments != nil && s.payments.Enabled() {
		orderStatus = models.OrderStatusPending
	}

	order := &models.Order{
		HotelID:       hotel.ID,
		UserID:        userID,
		Type:          models.OrderTypeProduct,
		Status:        orderStatus,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		RoomNumber:    req.RoomNumber,
		TotalAmount:   total,
		Currency:      currency,
		Notes:         notes,
	}

	if err := s.orders.Transaction(func(tx *gorm.DB) error {
		for _, item := range priced {
			if !item.trackStock {
				continue
			}
			if err := s.hotels.DecrementProductStock(tx, item.productID, item.quantity); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.New(apperrors.ErrBadRequest.Code, "insufficient stock", apperrors.ErrBadRequest.Status)
				}
				return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update stock", apperrors.ErrInternal.Status)
			}
		}
		order.Items = make([]models.OrderItem, 0, len(priced))
		for _, item := range priced {
			pid := item.productID
			order.Items = append(order.Items, models.OrderItem{
				ProductID:  &pid,
				Name:       item.name,
				Quantity:   item.quantity,
				UnitPrice:  item.unitPrice,
				TotalPrice: item.lineTotal,
			})
		}
		return s.orders.CreateTx(tx, order)
	}); err != nil {
		if appErr, ok := apperrors.IsAppError(err); ok {
			return nil, appErr
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create order", apperrors.ErrInternal.Status)
	}

	loaded, err := s.orders.FindByID(order.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	resp := toOrderResponse(loaded)
	if s.guestStays != nil {
		s.guestStays.RecordFromOrder(loaded, referralFromNotes(notes), models.GuestStaySourceProductOrder)
	}
	if s.payments != nil && s.payments.Enabled() {
		session, err := s.payments.StartCheckout(ctx, CheckoutInput{
			Order:        loaded,
			HotelName:    hotel.Name,
			BuyerEmail:   req.CustomerEmail,
			BuyerName:    req.CustomerName,
			BuyerPhone:   req.CustomerPhone,
			ReturnURL:    req.ReturnURL,
			CancelURL:    req.CancelURL,
			BuyerRemarks: notes,
			ItemCount:    itemCount,
		})
		if err != nil {
			loaded.Status = models.OrderStatusFailed
			loaded.PaymentStatus = PaymentStatusFailed
			_ = s.orders.Update(loaded)
			return nil, err
		}
		resp.PaymentRequired = session.PaymentRequired
		resp.PaymentURL = session.PaymentURL
		resp.PaymentProvider = session.PaymentProvider
		resp.PaymentStatus = session.PaymentStatus
		return resp, nil
	}

	if s.events != nil {
		s.events.OrderCreated(loaded, hotel.Name)
	}
	return resp, nil
}

func (s *OrderService) Track(ctx context.Context, id uuid.UUID) (*dto.OrderTrackResponse, error) {
	order, err := s.orders.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	status := string(order.Status)
	updatedAt := order.UpdatedAt

	if order.KkooappRef != "" {
		kkStatus, err := s.kkooapp.GetOrderStatus(ctx, order.KkooappRef)
		if err == nil {
			status = kkStatus.Status
			updatedAt = kkStatus.UpdatedAt
			order.Status = mapKkooappOrderStatus(kkStatus.Status)
			_ = s.orders.Update(order)
		}
	}

	return &dto.OrderTrackResponse{
		ID:         order.ID.String(),
		Status:     status,
		KkooappRef: order.KkooappRef,
		UpdatedAt:  updatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func toOrderResponse(o *models.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, dto.OrderItemResponse{
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
			Notes:      item.Notes,
		})
	}
	return &dto.OrderResponse{
		ID:              o.ID.String(),
		HotelID:         o.HotelID.String(),
		Type:            string(o.Type),
		Status:          string(o.Status),
		KkooappRef:      o.KkooappRef,
		CustomerName:    o.CustomerName,
		CustomerPhone:   o.CustomerPhone,
		TableNumber:     o.TableNumber,
		RoomNumber:      o.RoomNumber,
		TotalAmount:     o.TotalAmount,
		Currency:        o.Currency,
		Items:           items,
		Notes:           o.Notes,
		PaymentProvider: o.PaymentProvider,
		PaymentStatus:   o.PaymentStatus,
		PaymentRef:      o.PaymentRef,
		CreatedAt:       o.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapKkooappOrderStatus(status string) models.OrderStatus {
	switch status {
	case "confirmed":
		return models.OrderStatusConfirmed
	case "preparing":
		return models.OrderStatusPreparing
	case "ready":
		return models.OrderStatusReady
	case "delivered":
		return models.OrderStatusDelivered
	case "cancelled":
		return models.OrderStatusCancelled
	default:
		return models.OrderStatusPending
	}
}

func referralFromNotes(notes string) string {
	for _, part := range strings.Split(notes, "|") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "referral:") {
			return strings.TrimSpace(strings.TrimPrefix(part, "referral:"))
		}
	}
	return ""
}
