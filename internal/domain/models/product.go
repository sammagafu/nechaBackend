package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductBadge string

const (
	BadgeAfricanBrand ProductBadge = "african_brand"
	BadgeBestSeller   ProductBadge = "best_seller"
	BadgeNew          ProductBadge = "new"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	HotelID     uuid.UUID `gorm:"type:uuid;not null;index" json:"hotel_id"`
	Hotel       Hotel     `gorm:"foreignKey:HotelID" json:"-"`
	Slug        string    `gorm:"index" json:"slug"`
	BrandName   string    `json:"brand_name"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Category    string    `gorm:"not null;index" json:"category"`
	Badge       string    `json:"badge"`
	Price       int64     `gorm:"not null" json:"price"`
	Currency    string    `gorm:"not null;default:TZS" json:"currency"`
	ImageURL    string    `json:"image_url"`
	Stock       int       `gorm:"not null;default:0" json:"stock"`
	IsFeatured  bool      `gorm:"default:false" json:"is_featured"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
