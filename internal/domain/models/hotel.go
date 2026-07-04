package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for StringSlice")
	}
	return json.Unmarshal(bytes, s)
}

// Partner types govern which catalogue categories a partner may offer (brief §3.2, §3.5.3).
const (
	PartnerTypeHotel        = "hotel"
	PartnerTypeTourOperator = "tour_operator"
	PartnerTypeTravelAgent  = "travel_agent"
	PartnerTypeAirline      = "airline"
)

// Commission tiers drive which commission split applies (brief §4.2, §8).
const (
	CommissionTierFounding = "founding"
	CommissionTierStandard = "standard"
)

type Hotel struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Code         string      `gorm:"uniqueIndex;not null" json:"code"`
	Slug         string      `gorm:"uniqueIndex;not null" json:"slug"`
	Name         string      `gorm:"not null" json:"name"`
	Description  string      `json:"description"`
	Address      string      `json:"address"`
	City         string      `json:"city"`
	Location     string      `json:"location"`
	Country      string      `json:"country"`
	Zone           string      `json:"zone"`
	GoogleMapsURL  string      `json:"google_maps_url"`
	Latitude       *float64    `json:"latitude,omitempty"`
	Longitude      *float64    `json:"longitude,omitempty"`
	Phone          string      `json:"phone"`
	Email          string      `json:"email"`
	Initials     string      `json:"initials"`
	LogoURL      string      `json:"logo_url"`
	ReferralCode string      `gorm:"index" json:"referral_code"`
	Services     StringSlice `gorm:"type:jsonb" json:"services"`
	IsVerified   bool        `gorm:"default:true" json:"is_verified"`
	KkooappID    string      `gorm:"index" json:"kkooapp_id"`
	// PartnerType governs allowed catalogue categories. Only "hotel" behaves differently
	// today; other values exist so adding a partner type later is config, not migration.
	PartnerType string `gorm:"default:hotel;index" json:"partner_type"`
	// CommissionTier + CommissionTierStartDate let the founding cohort auto-revert to the
	// standard split after the configured tier window without a special-cased cohort.
	CommissionTier          string     `gorm:"default:standard" json:"commission_tier"`
	CommissionTierStartDate *time.Time `json:"commission_tier_start_date,omitempty"`
	SelcomPayoutAccount     string     `json:"selcom_payout_account,omitempty"`
	IsActive                bool       `gorm:"default:true" json:"is_active"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (h *Hotel) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
