package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderType string

const (
	OrderTypeFood    OrderType = "food"
	OrderTypeProduct OrderType = "product"
)

// Order categories map to the commission table (brief §2.2, §8.1). Every order must be
// classifiable into exactly one category so the engine knows who is paid and how.
const (
	OrderCategoryProduct        = "product"
	OrderCategoryTour           = "tour"
	OrderCategoryEvent          = "event"
	OrderCategoryDiningExternal = "dining_experience"
	OrderCategorySpaExternal    = "spa_service"
	OrderCategoryRoomService    = "room_service"
	OrderCategoryReservation    = "reservation_utility"
)

// Sales channels; B2C (no hotel premium) is a Phase 2 addition (brief §5.5).
const (
	SalesChannelHotelStorefront = "hotel_storefront"
	SalesChannelB2C             = "b2c"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusFailed     OrderStatus = "failed"
)

type Order struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Hotel        Hotel       `gorm:"foreignKey:HotelID" json:"-"`
	UserID       *uuid.UUID  `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User         *User       `gorm:"foreignKey:UserID" json:"-"`
	Type         OrderType   `gorm:"not null" json:"type"`
	Status       OrderStatus `gorm:"not null;default:pending" json:"status"`
	KkooappRef   string      `gorm:"index" json:"kkooapp_ref"`
	CustomerName string      `gorm:"not null" json:"customer_name"`
	CustomerPhone string     `json:"customer_phone"`
	TableNumber  string      `json:"table_number,omitempty"`
	RoomNumber   string      `json:"room_number,omitempty"`
	TotalAmount  int64       `gorm:"not null;default:0" json:"total_amount"`
	Currency     string      `gorm:"not null;default:USD" json:"currency"`
	Items        []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	Notes           string `json:"notes,omitempty"`
	// B2C street delivery (Phase 2 §17.1.2) — distinct from hotel room delivery.
	DeliveryAddress string `json:"delivery_address,omitempty"`
	DeliveryCity    string `json:"delivery_city,omitempty"`
	DeliveryCountry string `json:"delivery_country,omitempty"`
	DeliveryFee     int64  `gorm:"default:0" json:"delivery_fee"`
	// StockHeld is true while product stock was decremented as a payment hold.
	StockHeld bool `gorm:"default:false" json:"stock_held"`
	PaymentProvider string `json:"payment_provider,omitempty"`
	PaymentStatus   string `json:"payment_status,omitempty"`
	PaymentRef      string `gorm:"index" json:"payment_ref,omitempty"`
	PaymentIsDemo   bool   `gorm:"default:false" json:"payment_is_demo"`
	RefundedAmount  int64  `gorm:"default:0" json:"refunded_amount"`
	// Commission classification + attribution, captured once at placement (brief §3.7).
	Category     string `gorm:"index;default:product" json:"category"`
	SalesChannel string `gorm:"index;default:hotel_storefront" json:"sales_channel"`
	ReferralCode string `gorm:"index" json:"referral_code,omitempty"`
	// referred_by_influencer_id and referred_by_partner_id are mutually exclusive.
	ReferredByInfluencerID *uuid.UUID `gorm:"type:uuid;index" json:"referred_by_influencer_id,omitempty"`
	ReferredByPartnerID    *uuid.UUID `gorm:"type:uuid;index" json:"referred_by_partner_id,omitempty"`
	// Tax fields exist early (brief §3.16) — populated with real values only once confirmed.
	TaxAmount int64  `gorm:"default:0" json:"tax_amount"`
	TaxRate   float64 `gorm:"default:0" json:"tax_rate"`
	TaxType   string `json:"tax_type,omitempty"`
	FulfilledAt     *time.Time `json:"fulfilled_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID  *uuid.UUID `gorm:"type:uuid" json:"product_id,omitempty"`
	Name       string     `gorm:"not null" json:"name"`
	Quantity   int        `gorm:"not null" json:"quantity"`
	UnitPrice  int64      `gorm:"not null" json:"unit_price"`
	TotalPrice int64      `gorm:"not null" json:"total_price"`
	Notes      string     `json:"notes,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}
