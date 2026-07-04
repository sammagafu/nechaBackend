package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeliveryZone struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code               string    `gorm:"uniqueIndex;not null" json:"code"`
	Label              string    `gorm:"not null" json:"label"`
	DeliveryFeeTZS     int64     `gorm:"not null;default:5000" json:"delivery_fee_tzs"`
	FreeThresholdTZS   int64     `gorm:"not null;default:180000" json:"free_threshold_tzs"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	SortOrder          int       `gorm:"default:0" json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (d *DeliveryZone) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
