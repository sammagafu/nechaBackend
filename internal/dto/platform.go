package dto

type SubmitGuestRequestRequest struct {
	GuestName  string `json:"guest_name" validate:"required"`
	GuestPhone string `json:"guest_phone" validate:"required"`
	GuestEmail string `json:"guest_email"`
	RoomNumber string `json:"room_number"`
	Category   string `json:"category"`
	Body       string `json:"body" validate:"required"`
}

type GuestRequestResponse struct {
	ID         string `json:"id"`
	HotelID    string `json:"hotel_id"`
	Category   string `json:"category"`
	Status     string `json:"status"`
	GuestName  string `json:"guest_name"`
	GuestPhone string `json:"guest_phone"`
	GuestEmail string `json:"guest_email"`
	RoomNumber string `json:"room_number"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
}

type DeliveryZoneResponse struct {
	Code             string `json:"code"`
	Label            string `json:"label"`
	DeliveryFeeTZS   int64  `json:"delivery_fee_tzs"`
	FreeThresholdTZS int64  `json:"free_threshold_tzs"`
}

type PlatformSettingsResponse struct {
	TzsToUsdRate             float64                `json:"tzs_to_usd_rate"`
	FreeDeliveryThresholdTZS int64                  `json:"free_delivery_threshold_tzs"`
	DefaultDeliveryFeeTZS    int64                  `json:"default_delivery_fee_tzs"`
	DeliveryBaseFeeTZS       int64                  `json:"delivery_base_fee_tzs"`
	DeliveryPerKmTZS         int64                  `json:"delivery_per_km_tzs"`
	Features                 PlatformFeatures       `json:"features"`
	Zones                    []DeliveryZoneResponse `json:"zones"`
}

type PlatformFeatures struct {
	RewardsEnabled               bool `json:"rewards_enabled"`
	RewardsRedeemEnabled         bool `json:"rewards_redeem_enabled"`
	DiscoveryTicketingEnabled    bool `json:"discovery_ticketing_enabled"`
	PartnerPortalEnabled         bool `json:"partner_portal_enabled"`
	PartnerProductsManageEnabled bool `json:"partner_products_manage_enabled"`
	DualCurrencyEnabled          bool `json:"dual_currency_enabled"`
	DistanceDeliveryEnabled      bool `json:"distance_delivery_enabled"`
	B2CShopEnabled               bool `json:"b2c_shop_enabled"`
}

type UpdatePlatformSettingsRequest struct {
	TzsToUsdRate             *float64          `json:"tzs_to_usd_rate"`
	FreeDeliveryThresholdTZS *int64            `json:"free_delivery_threshold_tzs"`
	DefaultDeliveryFeeTZS    *int64            `json:"default_delivery_fee_tzs"`
	DeliveryBaseFeeTZS       *int64            `json:"delivery_base_fee_tzs"`
	DeliveryPerKmTZS         *int64            `json:"delivery_per_km_tzs"`
	Features                 *PlatformFeatures `json:"features"`
}

type MenuItemResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	Tag         string `json:"tag"`
	MenuKind    string `json:"menu_kind"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

type CreateMenuItemRequest struct {
	Slug        string `json:"slug" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int64  `json:"price" validate:"required"`
	Currency    string `json:"currency"`
	Tag         string `json:"tag"`
	MenuKind    string `json:"menu_kind"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

type UpdateMenuItemRequest struct {
	Category    *string `json:"category"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	Currency    *string `json:"currency"`
	Tag         *string `json:"tag"`
	MenuKind    *string `json:"menu_kind"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}
