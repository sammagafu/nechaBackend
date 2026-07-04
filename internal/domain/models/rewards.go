package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Rewards programme scaffolding (brief §5.5, §17.2, Phase 2). Reward earning wires into the
// order confirmation flow as a side effect, the same way commission records are created.
const (
	RewardEntryTypeEarn   = "earn"
	RewardEntryTypeRedeem = "redeem"
	RewardEntryTypeExpire = "expire"
)

// RewardRule defines how points are earned/redeemed. Kept configurable, not hardcoded.
type RewardRule struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code string    `gorm:"uniqueIndex;not null" json:"code"`
	Name string    `gorm:"not null" json:"name"`
	// PointsPerCurrencyUnit: points earned per 1 unit of order currency spent.
	PointsPerCurrencyUnit float64 `gorm:"default:0" json:"points_per_currency_unit"`
	// RedeemValuePerPoint: currency value of one point when redeemed.
	RedeemValuePerPoint float64   `gorm:"default:0" json:"redeem_value_per_point"`
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (r *RewardRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RewardLedgerEntry is an append-only ledger of points movements per guest/user.
type RewardLedgerEntry struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	GuestPhone string    `gorm:"index" json:"guest_phone,omitempty"`
	OrderID   *uuid.UUID `gorm:"type:uuid;index" json:"order_id,omitempty"`
	EntryType string     `gorm:"index;not null" json:"entry_type"`
	Points    int64      `json:"points"`
	Note      string     `json:"note,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (r *RewardLedgerEntry) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
