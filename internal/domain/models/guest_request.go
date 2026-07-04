package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GuestRequestCategoryGeneral  = "general"
	GuestRequestCategoryProduct  = "product"
	GuestRequestCategoryService  = "service"
	GuestRequestStatusNew        = "new"
	GuestRequestStatusInProgress = "in_progress"
	GuestRequestStatusResolved   = "resolved"
)

type GuestRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Hotel         Hotel      `gorm:"foreignKey:HotelID" json:"-"`
	Category      string     `gorm:"not null;default:general;index" json:"category"`
	Status        string     `gorm:"not null;default:new;index" json:"status"`
	GuestName     string     `json:"guest_name"`
	GuestPhone    string     `json:"guest_phone"`
	GuestEmail    string     `json:"guest_email"`
	RoomNumber    string     `json:"room_number"`
	Body          string     `gorm:"not null" json:"body"`
	UserID        *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (g *GuestRequest) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}
