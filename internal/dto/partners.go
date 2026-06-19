package dto

const FoundingCohortSize = 10

type PartnerHotelCard struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Initials string `json:"initials"`
	Slug     string `json:"slug"`
}

type PartnersLandingResponse struct {
	FoundingSpotsTotal int                `json:"founding_spots_total"`
	FoundingSpotsTaken int                `json:"founding_spots_taken"`
	PartnerHotels      []PartnerHotelCard `json:"partner_hotels"`
	FeaturedProducts   []ProductResponse  `json:"featured_products"`
	ActiveHotelCount   int                `json:"active_hotel_count"`
}
