package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationType string

const (
	ReservationTypeHotel ReservationType = "hotel"
	ReservationTypeTable ReservationType = "table"
)

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusFailed    ReservationStatus = "failed"
)

type Reservation struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID         uuid.UUID         `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Hotel           Hotel             `gorm:"foreignKey:HotelID" json:"-"`
	UserID          *uuid.UUID        `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User            *User             `gorm:"foreignKey:UserID" json:"-"`
	Type            ReservationType   `gorm:"not null" json:"type"`
	Status          ReservationStatus `gorm:"not null;default:pending" json:"status"`
	KkooappRef      string            `gorm:"index" json:"kkooapp_ref"`
	GuestName       string            `gorm:"not null" json:"guest_name"`
	GuestEmail      string            `json:"guest_email"`
	GuestPhone      string            `json:"guest_phone"`
	CheckIn         *time.Time        `json:"check_in,omitempty"`
	CheckOut        *time.Time        `json:"check_out,omitempty"`
	RoomType        string            `json:"room_type,omitempty"`
	GuestCount      int               `json:"guest_count"`
	ReservationDate *time.Time        `json:"reservation_date,omitempty"`
	TableNumber     string            `json:"table_number,omitempty"`
	PartySize       int               `json:"party_size,omitempty"`
	SpecialRequests string            `json:"special_requests,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (r *Reservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
