package dto

// HotelContextInput is optional context from the hotel QR storefront (register / social sign-in).
type HotelContextInput struct {
	HotelCode    string `json:"hotel_code"`
	HotelSlug    string `json:"hotel_slug"`
	Channel      string `json:"channel"`
	RoomNumber   string `json:"room_number"`
	ReferralCode string `json:"referral_code"`
	ScannedAt    string `json:"scanned_at"`
}

type HotelScanRequest struct {
	Ref       string `json:"ref"`
	Channel   string `json:"channel" validate:"required,oneof=room poster lobby"`
	ScannedAt string `json:"scanned_at"`
}

type AdminGuestStayResponse struct {
	ID              string  `json:"id"`
	UserID          *string `json:"user_id,omitempty"`
	UserEmail       string  `json:"user_email,omitempty"`
	UserName        string  `json:"user_name,omitempty"`
	HotelID         string  `json:"hotel_id"`
	HotelName       string  `json:"hotel_name"`
	Channel         string  `json:"channel,omitempty"`
	RoomNumber      string  `json:"room_number,omitempty"`
	ReferralCode    string  `json:"referral_code,omitempty"`
	Source          string  `json:"source"`
	OrderID         *string `json:"order_id,omitempty"`
	ReservationID   *string `json:"reservation_id,omitempty"`
	ScannedAt       string  `json:"scanned_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	ItemsSummary    string  `json:"items_summary,omitempty"`
	TotalAmount     *int64  `json:"total_amount,omitempty"`
	Currency        string  `json:"currency,omitempty"`
}
