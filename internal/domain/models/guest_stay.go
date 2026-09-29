package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuestStaySource string

const (
	GuestStaySourceRegister          GuestStaySource = "register"
	GuestStaySourceScan              GuestStaySource = "scan"
	GuestStaySourceProductOrder      GuestStaySource = "product_order"
	GuestStaySourceFoodOrder         GuestStaySource = "food_order"
	GuestStaySourceHotelReservation  GuestStaySource = "hotel_reservation"
	GuestStaySourceTableReservation  GuestStaySource = "table_reservation"
)

const (
	GuestStayChannelRoom   = "room"
	GuestStayChannelPoster = "poster"
	GuestStayChannelLobby  = "lobby"
)

// GuestStay links a user (optional) to a hotel visit: QR scan time, room, channel, and booking.
type GuestStay struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          *uuid.UUID      `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User            *User           `gorm:"foreignKey:UserID" json:"-"`
	HotelID         uuid.UUID       `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Hotel           Hotel           `gorm:"foreignKey:HotelID" json:"-"`
	Channel         string          `gorm:"index" json:"channel,omitempty"`
	RoomNumber      string          `json:"room_number,omitempty"`
	ReferralCode    string          `json:"referral_code,omitempty"`
	Source          GuestStaySource `gorm:"not null;index" json:"source"`
	OrderID         *uuid.UUID      `gorm:"type:uuid;index" json:"order_id,omitempty"`
	Order           *Order          `gorm:"foreignKey:OrderID" json:"-"`
	ReservationID   *uuid.UUID      `gorm:"type:uuid;index" json:"reservation_id,omitempty"`
	Reservation     *Reservation    `gorm:"foreignKey:ReservationID" json:"-"`
	ScannedAt       *time.Time      `json:"scanned_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (g *GuestStay) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}
