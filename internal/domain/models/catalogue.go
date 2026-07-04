package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Supplier types unify five supplier tables under one discriminator (brief §3.6). A
// property fulfilling its own in-house service is represented by supplier_type = property
// with supplier_ref_id pointing back at the Hotel, avoiding a redundant profile table.
const (
	SupplierTypeBrand               = "brand"
	SupplierTypeTourOperator        = "tour_operator"
	SupplierTypeEventOrganiser      = "event_organiser"
	SupplierTypeExternalSpaProvider = "external_spa_provider"
	SupplierTypeProperty            = "property"
)

type Supplier struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SupplierType string    `gorm:"index;not null" json:"supplier_type"`
	Name         string    `gorm:"not null" json:"name"`
	ContactPhone string    `json:"contact_phone"`
	ContactEmail string    `json:"contact_email"`
	// SupplierRefID links a supplier_type=property row back to its Hotel.
	SupplierRefID       *uuid.UUID `gorm:"type:uuid;index" json:"supplier_ref_id,omitempty"`
	LicenseReference    string     `json:"license_reference,omitempty"`
	SelcomPayoutAccount string     `json:"selcom_payout_account,omitempty"`
	AgreementNotes      string     `json:"agreement_notes,omitempty"`
	IsActive            bool       `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CatalogueItem is the central catalogue row (brief §2.5, §3.5). One item can be exposed to
// many properties via PropertyVisibility. capacity_or_stock is nullable/zero for free
// utility categories. Fulfilment + commission routing derive from category + supplier_type.
type CatalogueItem struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SupplierID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"supplier_id"`
	Category     string     `gorm:"index;not null" json:"category"`
	Name         string     `gorm:"not null" json:"name"`
	Description  string     `json:"description"`
	Currency     string     `gorm:"default:TZS" json:"currency"`
	BasePrice    int64      `gorm:"default:0" json:"base_price"`
	CapacityOrStock int64   `gorm:"default:0" json:"capacity_or_stock"`
	IsDigital    bool       `gorm:"default:false" json:"is_digital"`
	// EventDate applies to event/tour categories to drive payout-after-event timing.
	EventDate *time.Time `json:"event_date,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (c *CatalogueItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// PropertyVisibility controls which catalogue items appear on a given property's storefront
// and lets a property override display price (brief §2.5). Row-level tenancy: a guest at
// Property A must never see Property B's catalogue or pricing.
type PropertyVisibility struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID         uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_prop_item" json:"hotel_id"`
	CatalogueItemID uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_prop_item" json:"catalogue_item_id"`
	IsVisible       bool      `gorm:"default:true" json:"is_visible"`
	PriceOverride   *int64    `json:"price_override,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (p *PropertyVisibility) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
