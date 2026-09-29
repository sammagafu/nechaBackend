package dto

type SubmitInquiryRequest struct {
	Type              string `json:"type" validate:"required"`
	Name              string `json:"name"`
	Email             string `json:"email" validate:"required,email"`
	Phone             string `json:"phone"`
	Company           string `json:"company"`
	Role              string `json:"role"`
	Location          string `json:"location"`
	Category          string `json:"category"`
	Message           string `json:"message"`
	Rooms             int    `json:"rooms"`
	Units             int    `json:"units"`
	Price             int    `json:"price"`
	PropertyType      string `json:"property_type"`
	PreferredCallTime string `json:"preferred_call_time"`
	BrandStory        string `json:"brand_story"`
	SampleProducts    string `json:"sample_products"`
	ProposedCapacity  int    `json:"proposed_capacity"`
	ProposedDates     string `json:"proposed_dates"`
	PartnerType       string `json:"partner_type"`
	ClientVolume      string `json:"client_volume"`
	Destinations      string `json:"destinations"`
}

type InquiryResponse struct {
	ID                string  `json:"id"`
	Type              string  `json:"type"`
	Status            string  `json:"status"`
	Name              string  `json:"name"`
	Email             string  `json:"email"`
	Phone             string  `json:"phone"`
	Company           string  `json:"company"`
	Role              string  `json:"role"`
	Location          string  `json:"location"`
	Category          string  `json:"category"`
	Message           string  `json:"message"`
	Metadata          string  `json:"metadata"`
	SubmissionPayload string  `json:"submission_payload"`
	AssignedTo        *string `json:"assigned_to,omitempty"`
	ConvertedToID     *string `json:"converted_to_id,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

type UpdateInquiryRequest struct {
	Status        string  `json:"status" validate:"required"`
	AssignedTo    *string `json:"assigned_to"`
	ConvertedToID *string `json:"converted_to_id"`
}
