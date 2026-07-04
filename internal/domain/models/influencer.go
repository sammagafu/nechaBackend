package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	InfluencerStatusActive   = "active"
	InfluencerStatusInactive = "inactive"
)

// Influencer is an attribution + payout entity (brief §3.6.6) — it never lists catalogue
// items. Orders carry referred_by_influencer_id; the influencer is paid a share of Necha's
// margin through the same Commission Record / Payout Batch mechanism as suppliers.
type Influencer struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"index" json:"email"`
	Phone        string    `json:"phone"`
	ReferralCode string    `gorm:"uniqueIndex;not null" json:"referral_code"`
	// CommissionPct is stored per influencer since rates vary by agreement (default 30%
	// of Necha's base margin). A zero value falls back to the commission rule default.
	CommissionPct       float64   `gorm:"default:0.30" json:"commission_pct"`
	SelcomPayoutAccount string    `json:"selcom_payout_account,omitempty"`
	AgreementRef        string    `json:"agreement_ref,omitempty"`
	Status              string    `gorm:"index;default:active" json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (i *Influencer) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
