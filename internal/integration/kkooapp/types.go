package kkooapp

import "time"

// External API request/response types mirror Kkooapp partner contract.

type HotelReservationRequest struct {
	PropertyID      string `json:"property_id"`
	GuestName       string `json:"guest_name"`
	GuestEmail      string `json:"guest_email"`
	GuestPhone      string `json:"guest_phone"`
	CheckIn         string `json:"check_in"`
	CheckOut        string `json:"check_out"`
	RoomType        string `json:"room_type"`
	GuestCount      int    `json:"guest_count"`
	SpecialRequests string `json:"special_requests,omitempty"`
	ExternalRef     string `json:"external_ref"`
}

type TableReservationRequest struct {
	PropertyID      string `json:"property_id"`
	GuestName       string `json:"guest_name"`
	GuestEmail      string `json:"guest_email"`
	GuestPhone      string `json:"guest_phone"`
	ReservationDate string `json:"reservation_date"`
	PartySize       int    `json:"party_size"`
	TableNumber     string `json:"table_number,omitempty"`
	SpecialRequests string `json:"special_requests,omitempty"`
	ExternalRef     string `json:"external_ref"`
}

type FoodOrderItem struct {
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
	Notes     string `json:"notes,omitempty"`
}

type FoodOrderRequest struct {
	PropertyID    string          `json:"property_id"`
	CustomerName  string          `json:"customer_name"`
	CustomerPhone string          `json:"customer_phone"`
	TableNumber   string          `json:"table_number,omitempty"`
	RoomNumber    string          `json:"room_number,omitempty"`
	Items         []FoodOrderItem `json:"items"`
	Notes         string          `json:"notes,omitempty"`
	ExternalRef   string          `json:"external_ref"`
}

type ReservationResponse struct {
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderResponse struct {
	Reference   string    `json:"reference"`
	Status      string    `json:"status"`
	TotalAmount int64     `json:"total_amount"`
	Currency    string    `json:"currency"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderStatusResponse struct {
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIErrorResponse struct {
	Error APIError `json:"error"`
}
