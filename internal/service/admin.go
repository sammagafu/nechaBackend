package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type AdminService struct {
	hotels       *repository.HotelRepository
	orders       *repository.OrderRepository
	reservations *repository.ReservationRepository
	events       *EventService
	guestStays   *GuestStayService
}

func NewAdminService(hotels *repository.HotelRepository, orders *repository.OrderRepository, reservations *repository.ReservationRepository, events *EventService, guestStays *GuestStayService) *AdminService {
	return &AdminService{hotels: hotels, orders: orders, reservations: reservations, events: events, guestStays: guestStays}
}

func (s *AdminService) ListGuestStays(limit, offset int) ([]dto.AdminGuestStayResponse, error) {
	if s.guestStays == nil {
		return []dto.AdminGuestStayResponse{}, nil
	}
	return s.guestStays.List(limit, offset)
}

func (s *AdminService) Dashboard() (*dto.DashboardStats, error) {
	hotels, err := s.hotels.CountHotels()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count hotels", apperrors.ErrInternal.Status)
	}
	products, err := s.hotels.CountAllProducts()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count products", apperrors.ErrInternal.Status)
	}
	orders, err := s.orders.Count()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count orders", apperrors.ErrInternal.Status)
	}
	reservations, err := s.reservations.Count()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count reservations", apperrors.ErrInternal.Status)
	}
	pendingOrders, err := s.orders.CountByStatus(models.OrderStatusPending)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count pending orders", apperrors.ErrInternal.Status)
	}
	pendingReservations, err := s.reservations.CountByStatus(models.ReservationStatusPending)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count pending reservations", apperrors.ErrInternal.Status)
	}
	return &dto.DashboardStats{
		Hotels: hotels, Products: products, Orders: orders, Reservations: reservations,
		PendingOrders: pendingOrders, PendingReservations: pendingReservations,
	}, nil
}

func (s *AdminService) Analytics() (*dto.AnalyticsOverview, error) {
	since := time.Now().AddDate(0, 0, -13)
	revenue, err := s.orders.SumRevenue()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to sum revenue", apperrors.ErrInternal.Status)
	}
	orders30, err := s.orders.CountSince(time.Now().AddDate(0, 0, -30))
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count recent orders", apperrors.ErrInternal.Status)
	}
	reservations30, err := s.reservations.CountSince(time.Now().AddDate(0, 0, -30))
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count recent reservations", apperrors.ErrInternal.Status)
	}
	trendRows, err := s.orders.DailyMetricsSince(since)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order trend", apperrors.ErrInternal.Status)
	}
	statusRows, err := s.orders.CountByStatusGrouped()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order statuses", apperrors.ErrInternal.Status)
	}
	topRows, err := s.orders.TopHotelsByOrders(5)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load top hotels", apperrors.ErrInternal.Status)
	}

	trend := fillDailyMetrics(since, 14, trendRows)
	byStatus := make([]dto.StatusCount, 0, len(statusRows))
	for _, row := range statusRows {
		byStatus = append(byStatus, dto.StatusCount{Status: row.Status, Count: row.Count})
	}
	topHotels := make([]dto.HotelPerformance, 0, len(topRows))
	for _, row := range topRows {
		topHotels = append(topHotels, dto.HotelPerformance{
			HotelID: row.HotelID.String(), HotelName: row.HotelName,
			OrderCount: row.OrderCount, Revenue: row.Revenue,
		})
	}

	return &dto.AnalyticsOverview{
		TotalRevenue: revenue, Currency: "TZS",
		OrdersLast30Days: orders30, ReservationsLast30Days: reservations30,
		OrderTrend: trend, OrdersByStatus: byStatus, TopHotels: topHotels,
	}, nil
}

func (s *AdminService) StoreDashboard(hotelID string) (*dto.StoreDashboard, error) {
	uid, err := uuid.Parse(hotelID)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
	}
	hotel, err := s.hotels.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	since := time.Now().AddDate(0, 0, -13)
	thirtyDays := time.Now().AddDate(0, 0, -30)
	productCount, _ := s.hotels.CountProducts(uid)
	totalRevenue, err := s.orders.SumRevenueByHotel(uid)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to sum store revenue", apperrors.ErrInternal.Status)
	}
	orderCount, err := s.orders.CountByHotel(uid)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count store orders", apperrors.ErrInternal.Status)
	}
	reservationCount, err := s.reservations.CountByHotel(uid)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count store reservations", apperrors.ErrInternal.Status)
	}
	pendingOrders, _ := s.orders.CountByHotelAndStatus(uid, models.OrderStatusPending)
	pendingReservations, _ := s.reservations.CountByHotelAndStatus(uid, models.ReservationStatusPending)
	orders30, _ := s.orders.CountByHotelSince(uid, thirtyDays)
	revenue30, _ := s.orders.SumRevenueByHotelSince(uid, thirtyDays)
	trendRows, err := s.orders.DailyMetricsByHotelSince(uid, since)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load store trend", apperrors.ErrInternal.Status)
	}
	productRows, err := s.orders.TopProductsByHotel(uid, 5)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load top products", apperrors.ErrInternal.Status)
	}
	recentOrders, err := s.orders.ListByHotel(uid, 8)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load recent orders", apperrors.ErrInternal.Status)
	}
	recentReservations, err := s.reservations.ListByHotel(uid, 8)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load recent reservations", apperrors.ErrInternal.Status)
	}

	topProducts := make([]dto.ProductPerformance, 0, len(productRows))
	for _, row := range productRows {
		topProducts = append(topProducts, dto.ProductPerformance{
			ProductID: row.ProductID.String(), Name: row.Name,
			Quantity: row.Quantity, Revenue: row.Revenue,
		})
	}
	recent := make([]dto.AdminOrderResponse, 0, len(recentOrders))
	for _, order := range recentOrders {
		recent = append(recent, toAdminOrderResponse(&order))
	}
	recentRes := make([]dto.AdminReservationResponse, 0, len(recentReservations))
	for _, reservation := range recentReservations {
		recentRes = append(recentRes, toAdminReservationResponse(&reservation))
	}

	return &dto.StoreDashboard{
		Hotel: toAdminHotelResponse(hotel, productCount),
		TotalRevenue: totalRevenue, Currency: "TZS",
		Orders: orderCount, Reservations: reservationCount,
		PendingOrders: pendingOrders, PendingReservations: pendingReservations,
		OrdersLast30Days: orders30, RevenueLast30Days: revenue30,
		OrderTrend: fillDailyMetrics(since, 14, trendRows),
		TopProducts: topProducts, RecentOrders: recent, RecentReservations: recentRes,
	}, nil
}

func fillDailyMetrics(start time.Time, days int, rows []repository.DailyMetricRow) []dto.DailyMetric {
	byDate := make(map[string]repository.DailyMetricRow, len(rows))
	for _, row := range rows {
		byDate[row.Date] = row
	}
	result := make([]dto.DailyMetric, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		if row, ok := byDate[day]; ok {
			result = append(result, dto.DailyMetric{Date: day, Count: row.Count, Revenue: row.Revenue})
			continue
		}
		result = append(result, dto.DailyMetric{Date: day, Count: 0, Revenue: 0})
	}
	return result
}

func (s *AdminService) ListHotels() ([]dto.AdminHotelResponse, error) {
	hotels, err := s.hotels.ListAll()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list hotels", apperrors.ErrInternal.Status)
	}
	result := make([]dto.AdminHotelResponse, 0, len(hotels))
	for _, h := range hotels {
		pc, _ := s.hotels.CountProducts(h.ID)
		result = append(result, toAdminHotelResponse(&h, pc))
	}
	return result, nil
}

func (s *AdminService) GetHotel(id string) (*dto.AdminHotelResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
	}
	hotel, err := s.hotels.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	pc, _ := s.hotels.CountProducts(hotel.ID)
	resp := toAdminHotelResponse(hotel, pc)
	return &resp, nil
}

func (s *AdminService) CreateHotel(req dto.CreateHotelRequest) (*dto.AdminHotelResponse, error) {
	ref := req.ReferralCode
	if ref == "" {
		ref = req.Code
	}
	initials := req.Initials
	if initials == "" && len(req.Name) >= 2 {
		initials = string([]rune(req.Name)[0:2])
	}
	hotel := &models.Hotel{
		Code: req.Code, Slug: req.Slug, Name: req.Name, Description: req.Description,
		Address: req.Address, City: req.City, Location: req.Location, Country: req.Country,
		Zone: req.Zone, Phone: req.Phone, Initials: initials, LogoURL: req.LogoURL,
		ReferralCode: ref, Services: models.StringSlice(req.Services),
		IsVerified: req.IsVerified, KkooappID: req.KkooappID, IsActive: true,
	}
	if err := s.hotels.Create(hotel); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create hotel", apperrors.ErrInternal.Status)
	}
	resp := toAdminHotelResponse(hotel, 0)
	return &resp, nil
}

func (s *AdminService) UpdateHotel(id string, req dto.UpdateHotelRequest) (*dto.AdminHotelResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
	}
	hotel, err := s.hotels.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	applyHotelUpdate(hotel, req)
	if err := s.hotels.Update(hotel); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update hotel", apperrors.ErrInternal.Status)
	}
	pc, _ := s.hotels.CountProducts(hotel.ID)
	resp := toAdminHotelResponse(hotel, pc)
	return &resp, nil
}

func (s *AdminService) ListProducts(hotelID string) ([]dto.AdminProductResponse, error) {
	uid, err := uuid.Parse(hotelID)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
	}
	products, err := s.hotels.ListAllProducts(uid)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list products", apperrors.ErrInternal.Status)
	}
	return toAdminProductResponses(products), nil
}

func (s *AdminService) CreateProduct(hotelID string, req dto.CreateProductRequest) (*dto.AdminProductResponse, error) {
	uid, err := uuid.Parse(hotelID)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
	}
	if _, err := s.hotels.FindByID(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	currency := req.Currency
	if currency == "" {
		currency = "TZS"
	}
	product := &models.Product{
		HotelID: uid, Slug: req.Slug, BrandName: req.BrandName, Name: req.Name,
		Description: req.Description, Category: req.Category, Badge: req.Badge,
		Price: req.Price, Currency: currency, ImageURL: req.ImageURL,
		Stock: req.Stock, IsFeatured: req.IsFeatured, IsActive: true,
	}
	if err := s.hotels.CreateProduct(product); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create product", apperrors.ErrInternal.Status)
	}
	resp := toAdminProductResponse(product)
	return &resp, nil
}

func (s *AdminService) UpdateProduct(id string, req dto.UpdateProductRequest) (*dto.AdminProductResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid product id", apperrors.ErrBadRequest.Status)
	}
	product, err := s.hotels.FindProductByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "product not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load product", apperrors.ErrInternal.Status)
	}
	applyProductUpdate(product, req)
	if err := s.hotels.UpdateProduct(product); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update product", apperrors.ErrInternal.Status)
	}
	resp := toAdminProductResponse(product)
	return &resp, nil
}

func (s *AdminService) OrderSummary() (*dto.OrderSummary, error) {
	total, err := s.orders.Count()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count orders", apperrors.ErrInternal.Status)
	}
	pending, _ := s.orders.CountByStatus(models.OrderStatusPending)
	inProgress, _ := s.orders.CountInStatuses([]models.OrderStatus{
		models.OrderStatusConfirmed,
		models.OrderStatusPreparing,
		models.OrderStatusReady,
	})
	delivered, _ := s.orders.CountByStatus(models.OrderStatusDelivered)
	productOrders, _ := s.orders.CountByType(models.OrderTypeProduct)
	foodOrders, _ := s.orders.CountByType(models.OrderTypeFood)
	ordersToday, _ := s.orders.CountToday()
	thirtyDays := time.Now().AddDate(0, 0, -30)
	orders30, _ := s.orders.CountSince(thirtyDays)
	totalRevenue, _ := s.orders.SumRevenue()
	revenue30, _ := s.orders.SumRevenueSince(thirtyDays)
	return &dto.OrderSummary{
		Total: total, Pending: pending, InProgress: inProgress, Delivered: delivered,
		ProductOrders: productOrders, FoodOrders: foodOrders,
		OrdersToday: ordersToday, OrdersLast30Days: orders30,
		TotalRevenue: totalRevenue, RevenueLast30Days: revenue30,
		Currency: "TZS",
	}, nil
}

func (s *AdminService) GetOrder(id string) (*dto.AdminOrderDetailResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status)
	}
	order, err := s.orders.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	resp := toAdminOrderDetailResponse(order)
	return &resp, nil
}

func (s *AdminService) ListOrders(limit, offset int) ([]dto.AdminOrderResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	orders, err := s.orders.List(limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list orders", apperrors.ErrInternal.Status)
	}
	result := make([]dto.AdminOrderResponse, 0, len(orders))
	for _, o := range orders {
		result = append(result, toAdminOrderResponse(&o))
	}
	return result, nil
}

func (s *AdminService) UpdateOrderStatus(id, status string) (*dto.AdminOrderResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid order id", apperrors.ErrBadRequest.Status)
	}
	order, err := s.orders.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "order not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load order", apperrors.ErrInternal.Status)
	}
	previousStatus := string(order.Status)
	order.Status = models.OrderStatus(status)
	if err := s.orders.Update(order); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update order", apperrors.ErrInternal.Status)
	}
	reloaded, _ := s.orders.FindByID(uid)
	if s.events != nil && reloaded != nil {
		hotel, _ := s.hotels.FindByID(reloaded.HotelID)
		hotelName := ""
		if hotel != nil {
			hotelName = hotel.Name
		}
		s.events.OrderStatusUpdated(reloaded, hotelName, previousStatus)
	}
	resp := toAdminOrderResponse(reloaded)
	return &resp, nil
}

func (s *AdminService) GetReservation(id string) (*dto.AdminReservationResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid reservation id", apperrors.ErrBadRequest.Status)
	}
	reservation, err := s.reservations.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "reservation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load reservation", apperrors.ErrInternal.Status)
	}
	resp := toAdminReservationResponse(reservation)
	return &resp, nil
}

func (s *AdminService) ListReservations(limit, offset int) ([]dto.AdminReservationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	reservations, err := s.reservations.List(limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list reservations", apperrors.ErrInternal.Status)
	}
	result := make([]dto.AdminReservationResponse, 0, len(reservations))
	for _, r := range reservations {
		result = append(result, toAdminReservationResponse(&r))
	}
	return result, nil
}

func (s *AdminService) UpdateReservationStatus(id, status string) (*dto.AdminReservationResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid reservation id", apperrors.ErrBadRequest.Status)
	}
	reservation, err := s.reservations.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "reservation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load reservation", apperrors.ErrInternal.Status)
	}
	previousStatus := string(reservation.Status)
	reservation.Status = models.ReservationStatus(status)
	if err := s.reservations.Update(reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update reservation", apperrors.ErrInternal.Status)
	}
	reloaded, _ := s.reservations.FindByID(uid)
	if s.events != nil && reloaded != nil {
		hotel, _ := s.hotels.FindByID(reloaded.HotelID)
		hotelName := ""
		if hotel != nil {
			hotelName = hotel.Name
		}
		s.events.ReservationStatusUpdated(reloaded, hotelName, previousStatus)
	}
	resp := toAdminReservationResponse(reloaded)
	return &resp, nil
}

func applyHotelUpdate(h *models.Hotel, req dto.UpdateHotelRequest) {
	if req.Code != nil {
		h.Code = *req.Code
	}
	if req.Slug != nil {
		h.Slug = *req.Slug
	}
	if req.Name != nil {
		h.Name = *req.Name
	}
	if req.Description != nil {
		h.Description = *req.Description
	}
	if req.Address != nil {
		h.Address = *req.Address
	}
	if req.City != nil {
		h.City = *req.City
	}
	if req.Location != nil {
		h.Location = *req.Location
	}
	if req.Country != nil {
		h.Country = *req.Country
	}
	if req.Zone != nil {
		h.Zone = *req.Zone
	}
	if req.Phone != nil {
		h.Phone = *req.Phone
	}
	if req.Initials != nil {
		h.Initials = *req.Initials
	}
	if req.LogoURL != nil {
		h.LogoURL = *req.LogoURL
	}
	if req.ReferralCode != nil {
		h.ReferralCode = *req.ReferralCode
	}
	if req.Services != nil {
		h.Services = models.StringSlice(req.Services)
	}
	if req.IsVerified != nil {
		h.IsVerified = *req.IsVerified
	}
	if req.KkooappID != nil {
		h.KkooappID = *req.KkooappID
	}
	if req.IsActive != nil {
		h.IsActive = *req.IsActive
	}
}

func applyProductUpdate(p *models.Product, req dto.UpdateProductRequest) {
	if req.Slug != nil {
		p.Slug = *req.Slug
	}
	if req.BrandName != nil {
		p.BrandName = *req.BrandName
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Badge != nil {
		p.Badge = *req.Badge
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Currency != nil {
		p.Currency = *req.Currency
	}
	if req.ImageURL != nil {
		p.ImageURL = *req.ImageURL
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	if req.IsFeatured != nil {
		p.IsFeatured = *req.IsFeatured
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
}

func toAdminHotelResponse(h *models.Hotel, productCount int64) dto.AdminHotelResponse {
	services := []string(h.Services)
	if services == nil {
		services = []string{}
	}
	return dto.AdminHotelResponse{
		ID: h.ID.String(), Code: h.Code, Slug: h.Slug, Name: h.Name,
		Description: h.Description, Address: h.Address, City: h.City,
		Location: h.Location, Country: h.Country, Zone: h.Zone, Phone: h.Phone,
		Initials: h.Initials, LogoURL: h.LogoURL, ReferralCode: h.ReferralCode,
		Services: services, IsVerified: h.IsVerified, KkooappID: h.KkooappID,
		IsActive: h.IsActive, ProductCount: productCount,
		CreatedAt: h.CreatedAt.Format(time.RFC3339),
	}
}

func toAdminProductResponse(p *models.Product) dto.AdminProductResponse {
	return dto.AdminProductResponse{
		ID: p.ID.String(), HotelID: p.HotelID.String(), Slug: p.Slug,
		BrandName: p.BrandName, Name: p.Name, Description: p.Description,
		Category: p.Category, Badge: p.Badge, Price: p.Price, Currency: p.Currency,
		ImageURL: p.ImageURL, Stock: p.Stock, IsFeatured: p.IsFeatured,
		IsActive: p.IsActive, CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

func toAdminProductResponses(products []models.Product) []dto.AdminProductResponse {
	result := make([]dto.AdminProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, toAdminProductResponse(&p))
	}
	return result
}

func toAdminOrderResponse(o *models.Order) dto.AdminOrderResponse {
	hotelName := ""
	if o.Hotel.ID != uuid.Nil {
		hotelName = o.Hotel.Name
	}
	return dto.AdminOrderResponse{
		ID: o.ID.String(), HotelID: o.HotelID.String(), HotelName: hotelName,
		Type: string(o.Type), Status: string(o.Status), KkooappRef: o.KkooappRef,
		CustomerName: o.CustomerName, CustomerPhone: o.CustomerPhone,
		RoomNumber: o.RoomNumber, TotalAmount: o.TotalAmount, Currency: o.Currency,
		ItemCount: len(o.Items), CreatedAt: o.CreatedAt.Format(time.RFC3339),
	}
}

func toAdminOrderDetailResponse(o *models.Order) dto.AdminOrderDetailResponse {
	base := toAdminOrderResponse(o)
	items := make([]dto.AdminOrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, dto.AdminOrderItemResponse{
			ID: item.ID.String(), Name: item.Name, Quantity: item.Quantity,
			UnitPrice: item.UnitPrice, TotalPrice: item.TotalPrice, Notes: item.Notes,
		})
	}
	return dto.AdminOrderDetailResponse{
		AdminOrderResponse: base,
		TableNumber: o.TableNumber, Notes: o.Notes,
		PaymentProvider: o.PaymentProvider, PaymentStatus: o.PaymentStatus,
		PaymentRef: o.PaymentRef, UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
		Items: items,
	}
}

func toAdminReservationResponse(r *models.Reservation) dto.AdminReservationResponse {
	hotelName := ""
	if r.Hotel.ID != uuid.Nil {
		hotelName = r.Hotel.Name
	}
	resp := dto.AdminReservationResponse{
		ID: r.ID.String(), HotelID: r.HotelID.String(), HotelName: hotelName,
		Type: string(r.Type), Status: string(r.Status), KkooappRef: r.KkooappRef,
		GuestName: r.GuestName, GuestEmail: r.GuestEmail, GuestPhone: r.GuestPhone,
		RoomType: r.RoomType, GuestCount: r.GuestCount, TableNumber: r.TableNumber,
		PartySize: r.PartySize, SpecialRequests: r.SpecialRequests, Notes: r.Notes,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
	if r.CheckIn != nil {
		resp.CheckIn = r.CheckIn.Format(time.RFC3339)
	}
	if r.CheckOut != nil {
		resp.CheckOut = r.CheckOut.Format(time.RFC3339)
	}
	if r.ReservationDate != nil {
		resp.ReservationDate = r.ReservationDate.Format(time.RFC3339)
	}
	return resp
}
