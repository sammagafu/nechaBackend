package dto

type PartnerSettingsResponse struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Address             string   `json:"address"`
	City                string   `json:"city"`
	Location            string   `json:"location"`
	Phone               string   `json:"phone"`
	Email               string   `json:"email"`
	GoogleMapsURL       string   `json:"google_maps_url"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`
	SelcomPayoutAccount string   `json:"selcom_payout_account"`
}

type PartnerUpdateSettingsRequest struct {
	Description         *string  `json:"description"`
	Phone               *string  `json:"phone"`
	Email               *string  `json:"email"`
	GoogleMapsURL       *string  `json:"google_maps_url"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	SelcomPayoutAccount *string  `json:"selcom_payout_account"`
}

type PartnerCommissionResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	Category      string `json:"category"`
	GMV           int64  `json:"gmv"`
	PropertyShare int64  `json:"property_share"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type PartnerReferralResponse struct {
	ID             string `json:"id"`
	TravellerName  string `json:"traveller_name"`
	TravellerPhone string `json:"traveller_phone"`
	TravellerEmail string `json:"traveller_email"`
	Destination    string `json:"destination"`
	TripStartDate  string `json:"trip_start_date,omitempty"`
	TripEndDate    string `json:"trip_end_date,omitempty"`
	TripContext    string `json:"trip_context"`
	ReferralToken  string `json:"referral_token"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type CreatePartnerReferralRequest struct {
	TravellerName  string `json:"traveller_name" validate:"required"`
	TravellerPhone string `json:"traveller_phone" validate:"required"`
	TravellerEmail string `json:"traveller_email"`
	Destination    string `json:"destination"`
	TripStartDate  string `json:"trip_start_date"`
	TripEndDate    string `json:"trip_end_date"`
	TripContext    string `json:"trip_context"`
}

type DiscoveryOrderRequest struct {
	DiscoverySlug string `json:"discovery_slug" validate:"required"`
	HotelCode     string `json:"hotel_code"`
	CustomerName  string `json:"customer_name" validate:"required"`
	CustomerPhone string `json:"customer_phone" validate:"required"`
	CustomerEmail string `json:"customer_email"`
	Quantity      int    `json:"quantity" validate:"required,min=1"`
	ReturnURL     string `json:"return_url"`
	CancelURL     string `json:"cancel_url"`
}
