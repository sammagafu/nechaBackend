package dto

import "time"

type DiscoveryItemResponse struct {
	ID             string     `json:"id"`
	Section        string     `json:"section"`
	Subcategory    string     `json:"subcategory"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       string     `json:"image_url"`
	Venue          string     `json:"venue"`
	Location       string     `json:"location"`
	Distance       string     `json:"distance"`
	Phone          string     `json:"phone"`
	Website        string     `json:"website"`
	PriceHint      string     `json:"price_hint"`
	EventStartsAt  *time.Time `json:"event_starts_at"`
	EventEndsAt    *time.Time `json:"event_ends_at"`
	TicketURL      string     `json:"ticket_url"`
	TicketMode     string     `json:"ticket_mode"`
	OrganizerName  string     `json:"organizer_name,omitempty"`
	IsFeatured     bool       `json:"is_featured"`
}

type DiscoveryPortalResponse struct {
	Events      []DiscoveryItemResponse `json:"events"`
	Restaurants []DiscoveryItemResponse `json:"restaurants"`
	Tours       []DiscoveryItemResponse `json:"tours"`
}

type SubmitDiscoveryEventRequest struct {
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description" validate:"required"`
	Subcategory    string `json:"subcategory" validate:"required"`
	Venue          string `json:"venue" validate:"required"`
	Location       string `json:"location"`
	EventStartsAt  string `json:"event_starts_at" validate:"required"`
	EventEndsAt    string `json:"event_ends_at"`
	OrganizerName  string `json:"organizer_name" validate:"required"`
	OrganizerEmail string `json:"organizer_email" validate:"required,email"`
	OrganizerPhone string `json:"organizer_phone"`
	Website        string `json:"website"`
	TicketURL      string `json:"ticket_url"`
	TicketMode     string `json:"ticket_mode"`
	HotelSlug      string `json:"hotel_slug"`
}

type AdminDiscoveryItemResponse struct {
	ID             string     `json:"id"`
	HotelID        *string    `json:"hotel_id"`
	Section        string     `json:"section"`
	Subcategory    string     `json:"subcategory"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ImageURL       string     `json:"image_url"`
	Venue          string     `json:"venue"`
	Location       string     `json:"location"`
	Distance       string     `json:"distance"`
	Phone          string     `json:"phone"`
	Website        string     `json:"website"`
	PriceHint      string     `json:"price_hint"`
	EventStartsAt  *time.Time `json:"event_starts_at"`
	EventEndsAt    *time.Time `json:"event_ends_at"`
	TicketURL      string     `json:"ticket_url"`
	TicketMode     string     `json:"ticket_mode"`
	OrganizerName  string     `json:"organizer_name"`
	OrganizerEmail string     `json:"organizer_email"`
	IsFeatured     bool       `json:"is_featured"`
	Status         string     `json:"status"`
	SortOrder      int        `json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateDiscoveryItemRequest struct {
	HotelID        *string `json:"hotel_id"`
	Section        string  `json:"section" validate:"required"`
	Subcategory    string  `json:"subcategory"`
	Slug           string  `json:"slug"`
	Name           string  `json:"name" validate:"required"`
	Description    string  `json:"description"`
	ImageURL       string  `json:"image_url"`
	Venue          string  `json:"venue"`
	Location       string  `json:"location"`
	Distance       string  `json:"distance"`
	Phone          string  `json:"phone"`
	Website        string  `json:"website"`
	PriceHint      string  `json:"price_hint"`
	EventStartsAt  string  `json:"event_starts_at"`
	EventEndsAt    string  `json:"event_ends_at"`
	TicketURL      string  `json:"ticket_url"`
	TicketMode     string  `json:"ticket_mode"`
	OrganizerName  string  `json:"organizer_name"`
	OrganizerEmail string  `json:"organizer_email"`
	IsFeatured     bool    `json:"is_featured"`
	Status         string  `json:"status"`
	SortOrder      int     `json:"sort_order"`
}

type UpdateDiscoveryItemRequest struct {
	HotelID        *string `json:"hotel_id"`
	Section        *string `json:"section"`
	Subcategory    *string `json:"subcategory"`
	Slug           *string `json:"slug"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	ImageURL       *string `json:"image_url"`
	Venue          *string `json:"venue"`
	Location       *string `json:"location"`
	Distance       *string `json:"distance"`
	Phone          *string `json:"phone"`
	Website        *string `json:"website"`
	PriceHint      *string `json:"price_hint"`
	EventStartsAt  *string `json:"event_starts_at"`
	EventEndsAt    *string `json:"event_ends_at"`
	TicketURL      *string `json:"ticket_url"`
	TicketMode     *string `json:"ticket_mode"`
	OrganizerName  *string `json:"organizer_name"`
	OrganizerEmail *string `json:"organizer_email"`
	IsFeatured     *bool   `json:"is_featured"`
	Status         *string `json:"status"`
	SortOrder      *int    `json:"sort_order"`
}
