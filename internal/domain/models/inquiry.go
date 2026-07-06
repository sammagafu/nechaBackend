package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	InquiryTypeHotelPartner      = "hotel_partner"
	InquiryTypeBrandPartner      = "brand_partner"
	InquiryTypeContact           = "contact"
	InquiryTypeNewsletter        = "newsletter"
	InquiryTypeEventListing      = "event_listing"
	InquiryTypePartnerReferral   = "partner_referral"
	InquiryTypeDiscoveryBooking  = "discovery_booking"

	InquiryStatusNew      = "new"
	InquiryStatusRead     = "read"
	InquiryStatusArchived = "archived"
)

type Inquiry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Type      string    `gorm:"not null;index" json:"type"`
	Status    string    `gorm:"not null;default:new;index" json:"status"`
	Name      string    `json:"name"`
	Email     string    `gorm:"index" json:"email"`
	Phone     string    `json:"phone"`
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	Location  string    `json:"location"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i *Inquiry) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
