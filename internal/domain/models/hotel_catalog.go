package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryKind string

const (
	CategoryKindProduct CategoryKind = "product"
	CategoryKindMenu    CategoryKind = "menu"
)

type HotelRoom struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_hotel_room" json:"hotel_id"`
	Hotel      Hotel     `gorm:"foreignKey:HotelID" json:"-"`
	RoomNumber string    `gorm:"not null;uniqueIndex:idx_hotel_room" json:"room_number"`
	RoomType   string    `json:"room_type,omitempty"`
	Floor      string    `json:"floor,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HotelCategory struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID   uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:idx_hotel_category" json:"hotel_id"`
	Hotel     Hotel        `gorm:"foreignKey:HotelID" json:"-"`
	Slug      string       `gorm:"not null;uniqueIndex:idx_hotel_category" json:"slug"`
	Label     string       `gorm:"not null" json:"label"`
	Kind      CategoryKind `gorm:"not null;uniqueIndex:idx_hotel_category" json:"kind"`
	SortOrder int          `gorm:"default:0" json:"sort_order"`
	IsActive  bool         `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type HotelMenuItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_hotel_menu_item" json:"hotel_id"`
	Hotel       Hotel     `gorm:"foreignKey:HotelID" json:"-"`
	Slug        string    `gorm:"not null;uniqueIndex:idx_hotel_menu_item" json:"slug"`
	Category    string    `gorm:"not null;index" json:"category"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Price       int64     `gorm:"not null" json:"price"`
	Currency    string    `gorm:"not null;default:TZS" json:"currency"`
	Tag         string    `json:"tag,omitempty"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r *HotelRoom) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (c *HotelCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (m *HotelMenuItem) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
