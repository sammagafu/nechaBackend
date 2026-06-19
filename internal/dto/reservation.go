package dto

type HotelReservationRequest struct {
	HotelCode       string `json:"hotel_code" validate:"required"`
	GuestName       string `json:"guest_name" validate:"required"`
	GuestEmail      string `json:"guest_email" validate:"required,email"`
	GuestPhone      string `json:"guest_phone" validate:"required"`
	CheckIn         string `json:"check_in" validate:"required"`
	CheckOut        string `json:"check_out" validate:"required"`
	RoomType        string `json:"room_type" validate:"required"`
	GuestCount      int    `json:"guest_count" validate:"required,min=1"`
	SpecialRequests string `json:"special_requests"`
}

type TableReservationRequest struct {
	HotelCode       string `json:"hotel_code" validate:"required"`
	GuestName       string `json:"guest_name" validate:"required"`
	GuestEmail      string `json:"guest_email" validate:"required,email"`
	GuestPhone      string `json:"guest_phone" validate:"required"`
	ReservationDate string `json:"reservation_date" validate:"required"`
	PartySize       int    `json:"party_size" validate:"required,min=1"`
	TableNumber     string `json:"table_number"`
	SpecialRequests string `json:"special_requests"`
}

type ReservationResponse struct {
	ID              string `json:"id"`
	HotelID         string `json:"hotel_id"`
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
	PartySize       int    `json:"party_size,omitempty"`
	SpecialRequests string `json:"special_requests,omitempty"`
	CreatedAt       string `json:"created_at"`
}
