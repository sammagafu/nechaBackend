package dto

type DashboardStats struct {
	Hotels              int64 `json:"hotels"`
	Products            int64 `json:"products"`
	Orders              int64 `json:"orders"`
	Reservations        int64 `json:"reservations"`
	PendingOrders       int64 `json:"pending_orders"`
	PendingReservations int64 `json:"pending_reservations"`
}

type AnalyticsOverview struct {
	TotalRevenue           int64                  `json:"total_revenue"`
	Currency               string                 `json:"currency"`
	OrdersLast30Days       int64                  `json:"orders_last_30_days"`
	ReservationsLast30Days int64                  `json:"reservations_last_30_days"`
	OrderTrend             []DailyMetric          `json:"order_trend"`
	OrdersByStatus         []StatusCount          `json:"orders_by_status"`
	TopHotels              []HotelPerformance     `json:"top_hotels"`
}

type DailyMetric struct {
	Date    string `json:"date"`
	Count   int64  `json:"count"`
	Revenue int64  `json:"revenue"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type HotelPerformance struct {
	HotelID    string `json:"hotel_id"`
	HotelName  string `json:"hotel_name"`
	OrderCount int64  `json:"order_count"`
	Revenue    int64  `json:"revenue"`
}

type StoreDashboard struct {
	Hotel                  AdminHotelResponse   `json:"hotel"`
	TotalRevenue           int64                `json:"total_revenue"`
	Currency               string               `json:"currency"`
	Orders                 int64                `json:"orders"`
	Reservations           int64                `json:"reservations"`
	PendingOrders          int64                `json:"pending_orders"`
	PendingReservations    int64                `json:"pending_reservations"`
	OrdersLast30Days       int64                `json:"orders_last_30_days"`
	RevenueLast30Days      int64                `json:"revenue_last_30_days"`
	OrderTrend             []DailyMetric        `json:"order_trend"`
	TopProducts            []ProductPerformance `json:"top_products"`
	RecentOrders           []AdminOrderResponse       `json:"recent_orders"`
	RecentReservations     []AdminReservationResponse `json:"recent_reservations"`
}

type ProductPerformance struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Quantity  int64  `json:"quantity"`
	Revenue   int64  `json:"revenue"`
}

type AdminHotelResponse struct {
	ID           string   `json:"id"`
	Code         string   `json:"code"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Address      string   `json:"address"`
	City         string   `json:"city"`
	Location     string   `json:"location"`
	Country      string   `json:"country"`
	Zone         string   `json:"zone"`
	Phone        string   `json:"phone"`
	Initials     string   `json:"initials"`
	LogoURL      string   `json:"logo_url"`
	ReferralCode string   `json:"referral_code"`
	Services     []string `json:"services"`
	IsVerified   bool     `json:"is_verified"`
	KkooappID    string   `json:"kkooapp_id"`
	IsActive     bool     `json:"is_active"`
	ProductCount int64    `json:"product_count"`
	CreatedAt    string   `json:"created_at"`
}

type CreateHotelRequest struct {
	Code         string   `json:"code" validate:"required"`
	Slug         string   `json:"slug" validate:"required"`
	Name         string   `json:"name" validate:"required"`
	Description  string   `json:"description"`
	Address      string   `json:"address"`
	City         string   `json:"city"`
	Location     string   `json:"location"`
	Country      string   `json:"country"`
	Zone         string   `json:"zone"`
	Phone        string   `json:"phone"`
	Initials     string   `json:"initials"`
	LogoURL      string   `json:"logo_url"`
	ReferralCode string   `json:"referral_code"`
	Services     []string `json:"services"`
	IsVerified   bool     `json:"is_verified"`
	KkooappID    string   `json:"kkooapp_id"`
}

type UpdateHotelRequest struct {
	Code         *string  `json:"code"`
	Slug         *string  `json:"slug"`
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Address      *string  `json:"address"`
	City         *string  `json:"city"`
	Location     *string  `json:"location"`
	Country      *string  `json:"country"`
	Zone         *string  `json:"zone"`
	Phone        *string  `json:"phone"`
	Initials     *string  `json:"initials"`
	LogoURL      *string  `json:"logo_url"`
	ReferralCode *string  `json:"referral_code"`
	Services     []string `json:"services"`
	IsVerified   *bool    `json:"is_verified"`
	KkooappID    *string  `json:"kkooapp_id"`
	IsActive     *bool    `json:"is_active"`
}

type AdminProductResponse struct {
	ID          string `json:"id"`
	HotelID     string `json:"hotel_id"`
	Slug        string `json:"slug"`
	BrandName   string `json:"brand_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Badge       string `json:"badge"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	ImageURL    string `json:"image_url"`
	Stock       int    `json:"stock"`
	IsFeatured  bool   `json:"is_featured"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type CreateProductRequest struct {
	Slug        string `json:"slug" validate:"required"`
	BrandName   string `json:"brand_name"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Category    string `json:"category" validate:"required"`
	Badge       string `json:"badge"`
	Price       int64  `json:"price" validate:"required"`
	Currency    string `json:"currency"`
	ImageURL    string `json:"image_url"`
	Stock       int    `json:"stock"`
	IsFeatured  bool   `json:"is_featured"`
}

type UpdateProductRequest struct {
	Slug        *string `json:"slug"`
	BrandName   *string `json:"brand_name"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Badge       *string `json:"badge"`
	Price       *int64  `json:"price"`
	Currency    *string `json:"currency"`
	ImageURL    *string `json:"image_url"`
	Stock       *int    `json:"stock"`
	IsFeatured  *bool   `json:"is_featured"`
	IsActive    *bool   `json:"is_active"`
}

type AdminOrderResponse struct {
	ID            string `json:"id"`
	HotelID       string `json:"hotel_id"`
	HotelName     string `json:"hotel_name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	KkooappRef    string `json:"kkooapp_ref"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	RoomNumber    string `json:"room_number"`
	TotalAmount   int64  `json:"total_amount"`
	Currency      string `json:"currency"`
	ItemCount     int    `json:"item_count"`
	CreatedAt     string `json:"created_at"`
}

type AdminOrderItemResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	TotalPrice int64  `json:"total_price"`
	Notes      string `json:"notes,omitempty"`
}

type AdminOrderDetailResponse struct {
	AdminOrderResponse
	TableNumber     string                   `json:"table_number,omitempty"`
	Notes           string                   `json:"notes,omitempty"`
	PaymentProvider string                   `json:"payment_provider,omitempty"`
	PaymentStatus   string                   `json:"payment_status,omitempty"`
	PaymentRef      string                   `json:"payment_ref,omitempty"`
	UpdatedAt       string                   `json:"updated_at"`
	Items           []AdminOrderItemResponse `json:"items"`
}

type AdminReservationResponse struct {
	ID              string `json:"id"`
	HotelID         string `json:"hotel_id"`
	HotelName       string `json:"hotel_name"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	KkooappRef      string `json:"kkooapp_ref"`
	GuestName       string `json:"guest_name"`
	GuestEmail      string `json:"guest_email"`
	GuestPhone      string `json:"guest_phone"`
	CheckIn         string `json:"check_in,omitempty"`
	CheckOut        string `json:"check_out,omitempty"`
	RoomType        string `json:"room_type,omitempty"`
	GuestCount      int    `json:"guest_count,omitempty"`
	ReservationDate string `json:"reservation_date,omitempty"`
	TableNumber     string `json:"table_number,omitempty"`
	PartySize       int    `json:"party_size,omitempty"`
	SpecialRequests string `json:"special_requests,omitempty"`
	Notes           string `json:"notes,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type OrderSummary struct {
	Total           int64  `json:"total"`
	Pending         int64  `json:"pending"`
	InProgress      int64  `json:"in_progress"`
	Delivered       int64  `json:"delivered"`
	ProductOrders   int64  `json:"product_orders"`
	FoodOrders      int64  `json:"food_orders"`
	OrdersToday     int64  `json:"orders_today"`
	OrdersLast30Days int64 `json:"orders_last_30_days"`
	TotalRevenue    int64  `json:"total_revenue"`
	RevenueLast30Days int64 `json:"revenue_last_30_days"`
	Currency        string `json:"currency"`
}
