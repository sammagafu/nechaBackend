package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommissionRule stores the commercial terms per category (brief §3.12, §8.1).
// Rates live here rather than hardcoded so the founding/standard split recomputes if the
// base margin or premium pool ever changes.
type CommissionRule struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Category string    `gorm:"uniqueIndex;not null" json:"category"`
	// Product categories use the premium-split model.
	PremiumApplicable bool    `gorm:"default:false" json:"premium_applicable"`
	BaseMarginPct     float64 `gorm:"default:0" json:"base_margin_pct"`       // e.g. 0.20
	PremiumPoolPct    float64 `gorm:"default:0" json:"premium_pool_pct"`      // e.g. 0.13
	NechaPoolShareFounding float64 `gorm:"default:0" json:"necha_pool_share_founding"` // e.g. 0.50
	NechaPoolShareStandard float64 `gorm:"default:0" json:"necha_pool_share_standard"` // e.g. 0.70
	// Flat categories (tours, events, dining/spa external) use a single Necha commission pct.
	NechaCommissionPct float64 `gorm:"default:0" json:"necha_commission_pct"`
	// Referral modifier: fraction of Necha's share paid to an influencer/partner referral.
	InfluencerPctOfNecha float64 `gorm:"default:0.30" json:"influencer_pct_of_necha"`
	IsActive             bool      `gorm:"default:true" json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (c *CommissionRule) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Commission record lifecycle (brief §3.11, §8.2, §12.3).
const (
	CommissionStatusPending  = "pending"  // fulfilment not yet confirmed
	CommissionStatusEligible = "eligible" // eligible for payout batching
	CommissionStatusPaid     = "paid"     // included in a released payout batch
	CommissionStatusVoid     = "void"     // reversed by a refund (never deleted)
)

// CommissionRecord is created per order at payment capture (brief §8.2). Shares always sum
// to GMV. Referral shares are deducted from necha_share only, never from property/supplier.
type CommissionRecord struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID  uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	HotelID  uuid.UUID `gorm:"type:uuid;index" json:"hotel_id"`
	Category string    `gorm:"index" json:"category"`
	Currency string    `gorm:"default:TZS" json:"currency"`

	GMV                  int64 `json:"gmv"`
	NechaShare           int64 `json:"necha_share"`
	PropertyShare        int64 `json:"property_share"`
	SupplierShare        int64 `json:"supplier_share"`
	InfluencerShare      int64 `json:"influencer_share"`
	PartnerReferralShare int64 `json:"partner_referral_share"`
	WithholdingTaxAmount int64 `json:"withholding_tax_amount"`

	CommissionTier   string     `json:"commission_tier"`
	InfluencerID     *uuid.UUID `gorm:"type:uuid;index" json:"influencer_id,omitempty"`
	PartnerReferralID *uuid.UUID `gorm:"type:uuid;index" json:"partner_referral_id,omitempty"`

	Status          string     `gorm:"index;default:pending" json:"status"`
	PayoutEligibleAt *time.Time `json:"payout_eligible_at,omitempty"`
	PayoutBatchID   *uuid.UUID `gorm:"type:uuid;index" json:"payout_batch_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (c *CommissionRecord) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Payout batch payee types (brief §3.13, §8.3).
const (
	PayeeTypeProperty            = "property"
	PayeeTypeBrand               = "brand"
	PayeeTypeTourOperator        = "tour_operator"
	PayeeTypeEventOrganiser      = "event_organiser"
	PayeeTypeExternalSpaProvider = "external_spa_provider"
	PayeeTypeInfluencer          = "influencer"
	PayeeTypePartnerReferral     = "partner_referral"

	PayoutBatchStatusDraft    = "draft"
	PayoutBatchStatusReleased = "released"
)

// PayoutBatch groups eligible commission records for one payee into a settlement run.
type PayoutBatch struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	PayeeType   string     `gorm:"index;not null" json:"payee_type"`
	PayeeID     uuid.UUID  `gorm:"type:uuid;index" json:"payee_id"`
	Currency    string     `gorm:"default:TZS" json:"currency"`
	TotalAmount int64      `json:"total_amount"`
	RecordCount int        `json:"record_count"`
	Status      string     `gorm:"index;default:draft" json:"status"`
	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`
	ReleasedAt          *time.Time `json:"released_at,omitempty"`
	DisbursementRef     string     `json:"disbursement_ref,omitempty"`
	DisbursementStatus  string     `json:"disbursement_status,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (p *PayoutBatch) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PayoutBatchItem links a commission record to a payout batch and records the exact amount
// paid to that batch's payee. A single commission record can appear in several batches (e.g.
// a property batch for property_share and an influencer batch for influencer_share).
type PayoutBatchItem struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PayoutBatchID      uuid.UUID `gorm:"type:uuid;index;not null" json:"payout_batch_id"`
	CommissionRecordID uuid.UUID `gorm:"type:uuid;index;not null" json:"commission_record_id"`
	PayeeType          string    `gorm:"index" json:"payee_type"`
	Amount             int64     `json:"amount"`
	CreatedAt          time.Time `json:"created_at"`
}

func (p *PayoutBatchItem) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// EventLog is an append-only audit trail (brief §3.15). Rows are never updated or deleted.
type EventLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	EventType  string     `gorm:"index;not null" json:"event_type"`
	EntityType string     `gorm:"index" json:"entity_type"`
	EntityID   *uuid.UUID `gorm:"type:uuid;index" json:"entity_id,omitempty"`
	Detail     string     `json:"detail"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (e *EventLog) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// InventoryReservation closes the checkout race for capacity-constrained categories
// (brief §3.14) — applied only to tour/event/dining, not products/room service.
const (
	InventoryReservationStatusActive    = "active"
	InventoryReservationStatusConfirmed = "confirmed"
	InventoryReservationStatusExpired   = "expired"
	InventoryReservationStatusReleased  = "released"
)

type InventoryReservation struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CatalogueItemID uuid.UUID `gorm:"type:uuid;index;not null" json:"catalogue_item_id"`
	GuestID   *uuid.UUID `gorm:"type:uuid;index" json:"guest_id,omitempty"`
	Quantity  int        `gorm:"not null" json:"quantity"`
	Status    string     `gorm:"index;default:active" json:"status"`
	ReservedAt time.Time `json:"reserved_at"`
	ExpiresAt time.Time  `gorm:"index" json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PaymentRefund is an immutable mock or live refund attempt. Duplicate callbacks
// reuse IdempotencyKey so the same refund cannot reverse commission twice.
type PaymentRefund struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID         uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	Amount          int64     `gorm:"not null" json:"amount"`
	IdempotencyKey  string    `gorm:"uniqueIndex;not null" json:"idempotency_key"`
	Reason          string    `json:"reason,omitempty"`
	IsDemo          bool      `gorm:"default:true" json:"is_demo"`
	Status          string    `gorm:"default:completed" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

func (p *PaymentRefund) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (r *InventoryReservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.ReservedAt.IsZero() {
		r.ReservedAt = time.Now()
	}
	return nil
}
