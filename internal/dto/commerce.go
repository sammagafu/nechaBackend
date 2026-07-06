package dto

// Influencer -----------------------------------------------------------------

type CreateInfluencerRequest struct {
	Name                string  `json:"name" validate:"required"`
	Email               string  `json:"email"`
	Phone               string  `json:"phone"`
	ReferralCode        string  `json:"referral_code" validate:"required"`
	CommissionPct       float64 `json:"commission_pct"`
	SelcomPayoutAccount string  `json:"selcom_payout_account"`
	AgreementRef        string  `json:"agreement_ref"`
	Status              string  `json:"status"`
}

type UpdateInfluencerRequest struct {
	Name                *string  `json:"name"`
	Email               *string  `json:"email"`
	Phone               *string  `json:"phone"`
	ReferralCode        *string  `json:"referral_code"`
	CommissionPct       *float64 `json:"commission_pct"`
	SelcomPayoutAccount *string  `json:"selcom_payout_account"`
	AgreementRef        *string  `json:"agreement_ref"`
	Status              *string  `json:"status"`
}

// BookingReferral ------------------------------------------------------------

// CreateBookingReferralRequest is submitted by a referral partner (tour operator, travel
// agent, airline) per traveller — never by the guest. traveller_phone becomes the guest key.
type CreateBookingReferralRequest struct {
	ReferringPartnerCode string `json:"referring_partner_code" validate:"required"`
	TravellerName        string `json:"traveller_name" validate:"required"`
	TravellerPhone       string `json:"traveller_phone" validate:"required"`
	TravellerEmail       string `json:"traveller_email"`
	TripStartDate        string `json:"trip_start_date"`
	TripEndDate          string `json:"trip_end_date"`
	TripContext          string `json:"trip_context"`
}

// Supplier -------------------------------------------------------------------

type CreateSupplierRequest struct {
	SupplierType        string `json:"supplier_type" validate:"required"`
	Name                string `json:"name" validate:"required"`
	ContactPhone        string `json:"contact_phone"`
	ContactEmail        string `json:"contact_email"`
	LicenseReference    string `json:"license_reference"`
	SelcomPayoutAccount string `json:"selcom_payout_account"`
	AgreementNotes      string `json:"agreement_notes"`
}

type UpdateSupplierRequest struct {
	Name                *string `json:"name"`
	ContactPhone        *string `json:"contact_phone"`
	ContactEmail        *string `json:"contact_email"`
	LicenseReference    *string `json:"license_reference"`
	SelcomPayoutAccount *string `json:"selcom_payout_account"`
	AgreementNotes      *string `json:"agreement_notes"`
	IsActive            *bool   `json:"is_active"`
}

// CatalogueItem --------------------------------------------------------------

type CreateCatalogueItemRequest struct {
	SupplierID      string `json:"supplier_id" validate:"required"`
	Category        string `json:"category" validate:"required"`
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description"`
	Currency        string `json:"currency"`
	BasePrice       int64  `json:"base_price"`
	CapacityOrStock int64  `json:"capacity_or_stock"`
	IsDigital       bool   `json:"is_digital"`
	EventDate       string `json:"event_date"`
}

type UpdateCatalogueItemRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Currency        *string `json:"currency"`
	BasePrice       *int64  `json:"base_price"`
	CapacityOrStock *int64  `json:"capacity_or_stock"`
	IsDigital       *bool   `json:"is_digital"`
	IsActive        *bool   `json:"is_active"`
}

type SetVisibilityRequest struct {
	HotelID       string `json:"hotel_id" validate:"required"`
	IsVisible     bool   `json:"is_visible"`
	PriceOverride *int64 `json:"price_override"`
}

// Commission rules -----------------------------------------------------------

type UpsertCommissionRuleRequest struct {
	Category               string  `json:"category" validate:"required"`
	PremiumApplicable      bool    `json:"premium_applicable"`
	BaseMarginPct          float64 `json:"base_margin_pct"`
	PremiumPoolPct         float64 `json:"premium_pool_pct"`
	NechaPoolShareFounding float64 `json:"necha_pool_share_founding"`
	NechaPoolShareStandard float64 `json:"necha_pool_share_standard"`
	NechaCommissionPct     float64 `json:"necha_commission_pct"`
	InfluencerPctOfNecha   float64 `json:"influencer_pct_of_necha"`
	IsActive               bool    `json:"is_active"`
}

// Payouts --------------------------------------------------------------------

type GenerateBatchesRequest struct {
	Until string `json:"until"`
}

// Rewards --------------------------------------------------------------------

type RedeemRewardsRequest struct {
	Points int64 `json:"points" validate:"required,min=1"`
}

type UpsertRewardRuleRequest struct {
	Code                  string  `json:"code" validate:"required"`
	Name                  string  `json:"name" validate:"required"`
	PointsPerCurrencyUnit float64 `json:"points_per_currency_unit"`
	RedeemValuePerPoint   float64 `json:"redeem_value_per_point"`
	IsActive              bool    `json:"is_active"`
}
