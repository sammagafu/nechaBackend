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
	PaymentProvider string `json:"payment_provider,omitempty"`
	PaymentStatus   string `json:"payment_status,omitempty"`
	PaymentRef      string `gorm:"index" json:"payment_ref,omitempty"`
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
