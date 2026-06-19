package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for StringSlice")
	}
	return json.Unmarshal(bytes, s)
}

type Hotel struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Code         string      `gorm:"uniqueIndex;not null" json:"code"`
	Slug         string      `gorm:"uniqueIndex;not null" json:"slug"`
	Name         string      `gorm:"not null" json:"name"`
	Description  string      `json:"description"`
	Address      string      `json:"address"`
	City         string      `json:"city"`
	Location     string      `json:"location"`
	Country      string      `json:"country"`
	Zone         string      `json:"zone"`
	Phone        string      `json:"phone"`
	Initials     string      `json:"initials"`
	LogoURL      string      `json:"logo_url"`
	ReferralCode string      `gorm:"index" json:"referral_code"`
	Services     StringSlice `gorm:"type:jsonb" json:"services"`
	IsVerified   bool        `gorm:"default:true" json:"is_verified"`
	KkooappID    string      `gorm:"index" json:"kkooapp_id"`
	IsActive     bool        `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (h *Hotel) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
