package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	BookingReferralStatusPending   = "pending"
	BookingReferralStatusConverted = "converted"
	BookingReferralStatusExpired   = "expired"
)

// BookingReferral captures how a tour operator, travel agent, or airline introduces a
// specific traveller (brief §3.6.7). It is created by the partner (not the guest), pre-fills
// trip context, and drives referred_by_partner_id attribution on the traveller's orders for
// the duration of that trip (the attribution window, brief §3.7 note).
type BookingReferral struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	// ReferringPartnerID must point to a Hotel with partner_type in
	// {tour_operator, travel_agent, airline}.
	ReferringPartnerID uuid.UUID `gorm:"type:uuid;index;not null" json:"referring_partner_id"`
	TravellerName      string    `json:"traveller_name"`
	TravellerPhone     string    `gorm:"index" json:"traveller_phone"`
	TravellerEmail     string    `json:"traveller_email"`
	TripStartDate      *time.Time `json:"trip_start_date,omitempty"`
	TripEndDate        *time.Time `json:"trip_end_date,omitempty"`
	TripContext        string    `json:"trip_context"`
	// ReferralToken is the unique link parameter sent to the traveller, functioning like a
	// hotel's QR-encoded URL but pre-filling referral + trip context.
	ReferralToken    string     `gorm:"uniqueIndex;not null" json:"referral_token"`
	ConvertedGuestID *uuid.UUID `gorm:"type:uuid;index" json:"converted_guest_id,omitempty"`
	Status           string     `gorm:"index;default:pending" json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (b *BookingReferral) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
