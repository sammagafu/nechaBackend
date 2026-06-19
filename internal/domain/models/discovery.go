package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DiscoverySectionEvents      = "events"
	DiscoverySectionRestaurants = "restaurants"
	DiscoverySectionTours       = "tours"

	DiscoveryStatusPending = "pending"
	DiscoveryStatusActive  = "active"
	DiscoveryStatusRemoved = "removed"

	TicketModeNone     = "none"
	TicketModeReferral = "referral"
	TicketModePlatform = "platform"
)

type DiscoveryItem struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID        *uuid.UUID `gorm:"type:uuid;index" json:"hotel_id"`
	Section        string     `gorm:"not null;index" json:"section"`
	Subcategory    string     `gorm:"index" json:"subcategory"`
	Slug           string     `gorm:"index" json:"slug"`
	Name           string     `gorm:"not null" json:"name"`
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
	TicketMode     string     `gorm:"default:none" json:"ticket_mode"`
	OrganizerName  string     `json:"organizer_name"`
	OrganizerEmail string     `json:"organizer_email"`
	IsFeatured     bool       `gorm:"default:false" json:"is_featured"`
	Status         string     `gorm:"not null;default:active;index" json:"status"`
	SortOrder      int        `gorm:"default:0" json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (d *DiscoveryItem) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
