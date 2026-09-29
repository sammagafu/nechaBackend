package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Canonical inquiry types from architecture §9.2. Legacy aliases are accepted on write
// and normalised to these values.
const (
	InquiryTypeHotelPartnership      = "hotel_partnership"
	InquiryTypeBrandOnboarding       = "brand_onboarding"
	InquiryTypeEventListing          = "event_listing"
	InquiryTypePartnerReferralSignup = "partner_referral_signup"

	// Legacy aliases kept so existing rows and older clients keep working.
	InquiryTypeHotelPartner     = "hotel_partner"
	InquiryTypeBrandPartner     = "brand_partner"
	InquiryTypeAffiliatePartner = "affiliate_partner"
	InquiryTypePartnerReferral  = "partner_referral"

	InquiryTypeContact          = "contact"
	InquiryTypeNewsletter       = "newsletter"
	InquiryTypeDiscoveryBooking = "discovery_booking"

	InquiryStatusNew       = "new"
	InquiryStatusContacted = "contacted"
	InquiryStatusQualified = "qualified"
	InquiryStatusConverted = "converted"
	InquiryStatusRead      = "read"     // legacy → treated as contacted
	InquiryStatusArchived  = "archived" // retained for existing rows
)

type Inquiry struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Type              string     `gorm:"not null;index" json:"type"`
	Status            string     `gorm:"not null;default:new;index" json:"status"`
	Name              string     `json:"name"`
	Email             string     `gorm:"index" json:"email"`
	Phone             string     `json:"phone"`
	Company           string     `json:"company"`
	Role              string     `json:"role"`
	Location          string     `json:"location"`
	Category          string     `json:"category"`
	Message           string     `json:"message"`
	Metadata          string     `json:"metadata"`
	SubmissionPayload string     `gorm:"type:text" json:"submission_payload"`
	AssignedTo        *uuid.UUID `gorm:"type:uuid;index" json:"assigned_to,omitempty"`
	ConvertedToID     *uuid.UUID `gorm:"type:uuid;index" json:"converted_to_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (i *Inquiry) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// NormalizeInquiryType maps legacy form types onto architecture §9 values.
func NormalizeInquiryType(raw string) string {
	switch raw {
	case InquiryTypeHotelPartner:
		return InquiryTypeHotelPartnership
	case InquiryTypeBrandPartner:
		return InquiryTypeBrandOnboarding
	case InquiryTypeAffiliatePartner, InquiryTypePartnerReferral:
		return InquiryTypePartnerReferralSignup
	default:
		return raw
	}
}
