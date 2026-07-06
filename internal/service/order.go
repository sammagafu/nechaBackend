package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type OrderService struct {
	hotels      *repository.HotelRepository
	orders      *repository.OrderRepository
	discovery   *repository.DiscoveryRepository
	events      *EventService
	payments    *PaymentService
	guestStays  *GuestStayService
	platform    *PlatformService
	influencers *repository.InfluencerRepository
	bookings    *repository.BookingReferralRepository
	rewards     *repository.RewardRepository
}

func NewOrderService(
	hotels *repository.HotelRepository,
	orders *repository.OrderRepository,
	discovery *repository.DiscoveryRepository,
	events *EventService,
	payments *PaymentService,
	guestStays *GuestStayService,
	platform *PlatformService,
	influencers *repository.InfluencerRepository,
	bookings *repository.BookingReferralRepository,
	rewards *repository.RewardRepository,
) *OrderService {
	return &OrderService{
		hotels:      hotels,
		orders:      orders,
		discovery:   discovery,
		events:      events,
		payments:    payments,
		guestStays:  guestStays,
		platform:    platform,
		influencers: influencers,
		bookings:    bookings,
		rewards:     rewards,
	}
}

// resolveAttribution captures referral attribution once, at order placement (brief §3.7).
// Influencer (by referral code) takes precedence; otherwise an open partner booking referral
// matched on the customer's phone. The two are mutually exclusive.
func (s *OrderService) resolveAttribution(order *models.Order, referralCode, customerPhone string) {
	code := strings.TrimSpace(referralCode)
	if code != "" {
		order.ReferralCode = code
		if s.influencers != nil {
			if inf, err := s.influencers.FindByReferralCode(code); err == nil {
				order.ReferredByInfluencerID = &inf.ID
				return
			}
		}
	}
	if s.bookings != nil && strings.TrimSpace(customerPhone) != "" {
		if ref, err := s.bookings.FindActiveByPhone(strings.TrimSpace(customerPhone), time.Now()); err == nil {
			order.ReferredByPartnerID = &ref.ReferringPartnerID
		}
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

	var total int64
	items := make([]models.OrderItem, 0, len(req.Items))
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
		Category:      models.OrderCategoryRoomService,
		SalesChannel:  models.SalesChannelHotelStorefront,
	}
	if req.RequirePayment && total > 0 {
		order.Category = models.OrderCategorySpaExternal
		if s.payments != nil && s.payments.Enabled() {
			order.Status = models.OrderStatusPending
		} else {
			order.Status = models.OrderStatusConfirmed
		}
	}

	if err := s.orders.Create(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create order", apperrors.ErrInternal.Status)
	}

	order.KkooappRef = nechaOrderRef(order.ID)
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order", apperrors.ErrInternal.Status)
	}

	loaded, err := s.orders.FindByID(order.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	resp := toOrderResponse(loaded)
	if req.RequirePayment && total > 0 && s.payments != nil && s.payments.Enabled() {
		session, err := s.payments.StartCheckout(ctx, CheckoutInput{
			Order:        loaded,
			HotelName:    hotel.Name,
			BuyerEmail:   req.CustomerEmail,
			BuyerName:    req.CustomerName,
			BuyerPhone:   req.CustomerPhone,
			ReturnURL:    req.ReturnURL,
			CancelURL:    req.CancelURL,
			BuyerRemarks: req.Notes,
			ItemCount:    len(items),
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
		if s.events != nil && !session.PaymentRequired {
			s.events.OrderCreated(loaded, hotel)
		}
	} else if s.events != nil {
		s.events.OrderCreated(loaded, hotel)
	}
	if s.guestStays != nil {
		s.guestStays.RecordFromOrder(loaded, referralFromNotes(req.Notes), models.GuestStaySourceFoodOrder)
	}
	return resp, nil
}

func nechaOrderRef(id uuid.UUID) string {
	s := strings.ReplaceAll(id.String(), "-", "")
	if len(s) > 10 {
		s = s[:10]
	}
	return "NA-" + strings.ToUpper(s)
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
	if req.DeliveryFee > 0 || s.platform != nil {
		zoneCode := strings.TrimSpace(req.DeliveryZoneCode)
		if zoneCode == "" {
			zoneCode = hotel.Zone
		}
		var expectedFee int64
		if s.platform != nil {
			// Prefer distance-based pricing when the guest supplied delivery coordinates and
			// the hotel has a stored location (brief §4.5); otherwise fall back to flat zone pricing.
			if req.DeliveryLatitude != nil && req.DeliveryLongitude != nil && hotel.Latitude != nil && hotel.Longitude != nil {
				expectedFee, err = s.platform.ComputeDeliveryFeeByDistance(hotel, req.DeliveryLatitude, req.DeliveryLongitude, total)
			} else {
				expectedFee, err = s.platform.ComputeDeliveryFee(zoneCode, total)
			}
			if err != nil {
				return nil, err
			}
		} else if req.DeliveryFee > 0 {
			expectedFee = req.DeliveryFee
		}
		if req.DeliveryFee != expectedFee {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "delivery fee does not match zone pricing", apperrors.ErrBadRequest.Status)
		}
		total += expectedFee
	} else if req.DeliveryFee > 0 {
		total += req.DeliveryFee
	}

	redeemPoints := int64(0)
	if req.RedeemPoints > 0 && userID != nil && s.rewards != nil && s.platform != nil &&
		s.platform.FeatureEnabled(models.ConfigKeyFeatureRewardsRedeemEnabled, false) {
		rule, err := s.rewards.ActiveRule()
		if err != nil {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "rewards not configured", apperrors.ErrBadRequest.Status)
		}
		balance, err := s.rewards.BalanceForUser(*userID)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load reward balance", apperrors.ErrInternal.Status)
		}
		if req.RedeemPoints > balance {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "insufficient reward points", apperrors.ErrBadRequest.Status)
		}
		discount := int64(float64(req.RedeemPoints) * rule.RedeemValuePerPoint)
		if discount > total {
			discount = total
		}
		total -= discount
		if discount > 0 {
			redeemPoints = req.RedeemPoints
		}
	}

	notes := req.Notes
	meta := []string{}
	if redeemPoints > 0 {
		meta = append(meta, "redeem_points:"+strconv.FormatInt(redeemPoints, 10))
	}
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
		Category:      models.OrderCategoryProduct,
		SalesChannel:  models.SalesChannelHotelStorefront,
	}
	if strings.TrimSpace(req.SalesChannel) == models.SalesChannelB2C {
		order.SalesChannel = models.SalesChannelB2C
	}
	refCode := strings.TrimSpace(req.ReferralCode)
	if refCode == "" {
		refCode = referralFromNotes(notes)
	}
	s.resolveAttribution(order, refCode, req.CustomerPhone)

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

	if redeemPoints > 0 && userID != nil && s.rewards != nil {
		orderID := loaded.ID
		_ = s.rewards.AppendEntry(&models.RewardLedgerEntry{
			UserID:    userID,
			OrderID:   &orderID,
			EntryType: models.RewardEntryTypeRedeem,
			Points:    -redeemPoints,
			Note:      "checkout order " + loaded.ID.String()[:8],
		})
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
		s.events.OrderCreated(loaded, hotel)
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

	return &dto.OrderTrackResponse{
		ID:         order.ID.String(),
		Status:     status,
		KkooappRef: order.KkooappRef,
		UpdatedAt:  updatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *OrderService) CreateDiscoveryOrder(ctx context.Context, req dto.DiscoveryOrderRequest, userID *uuid.UUID) (*dto.OrderResponse, error) {
	if s.platform != nil && !s.platform.FeatureEnabled(models.ConfigKeyFeatureDiscoveryTicketingEnabled, false) {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "discovery ticketing is not enabled", apperrors.ErrBadRequest.Status)
	}
	if s.discovery == nil {
		return nil, apperrors.New(apperrors.ErrInternal.Code, "discovery booking unavailable", apperrors.ErrInternal.Status)
	}

	item, err := s.discovery.FindActiveBySlug(strings.TrimSpace(req.DiscoverySlug))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "discovery item not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load discovery item", apperrors.ErrInternal.Status)
	}
	if item.TicketMode != models.TicketModePlatform {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "this listing does not support in-app ticketing", apperrors.ErrBadRequest.Status)
	}
	if item.PriceTZS <= 0 {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "ticket price is not configured", apperrors.ErrBadRequest.Status)
	}

	qty := req.Quantity
	if qty < 1 {
		qty = 1
	}
	if item.TicketCapacity > 0 && item.TicketsSold+qty > item.TicketCapacity {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "not enough tickets available", apperrors.ErrBadRequest.Status)
	}

	var hotel *models.Hotel
	if item.HotelID != nil {
		hotel, err = s.hotels.FindByID(*item.HotelID)
	} else {
		hotel, err = s.hotels.FindByCode("SEACLIFF24")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "booking hotel not configured", apperrors.ErrBadRequest.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	lineTotal := item.PriceTZS * int64(qty)
	category := models.OrderCategoryEvent
	if item.Section == models.DiscoverySectionTours {
		category = models.OrderCategoryTour
	}

	notes := "discovery_ticket:" + item.Slug + "|qty:" + strconv.Itoa(qty) + "|item:" + item.ID.String()
	order := &models.Order{
		HotelID:       hotel.ID,
		UserID:        userID,
		Type:          models.OrderTypeProduct,
		Status:        models.OrderStatusPending,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		TotalAmount:   lineTotal,
		Currency:      "TZS",
		Category:      category,
		SalesChannel:  models.SalesChannelB2C,
		Notes:         notes,
		Items: []models.OrderItem{{
			Name:       item.Name + " ticket",
			Quantity:   qty,
			UnitPrice:  item.PriceTZS,
			TotalPrice: lineTotal,
		}},
	}

	if err := s.orders.Create(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create order", apperrors.ErrInternal.Status)
	}

	loaded, err := s.orders.FindByID(order.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}

	resp := toOrderResponse(loaded)
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
			ItemCount:    qty,
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
		if s.events != nil && !session.PaymentRequired {
			s.events.OrderCreated(loaded, hotel)
		}
	} else if s.events != nil {
		s.events.OrderCreated(loaded, hotel)
	}
	return resp, nil
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

func referralFromNotes(notes string) string {
	for _, part := range strings.Split(notes, "|") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "referral:") {
			return strings.TrimSpace(strings.TrimPrefix(part, "referral:"))
		}
	}
	return ""
}
