package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type GuestRequestRepository struct {
	db *gorm.DB
}

func NewGuestRequestRepository(db *gorm.DB) *GuestRequestRepository {
	return &GuestRequestRepository{db: db}
}

func (r *GuestRequestRepository) Create(item *models.GuestRequest) error {
	return r.db.Create(item).Error
}

func (r *GuestRequestRepository) ListByHotel(hotelID uuid.UUID, limit int) ([]models.GuestRequest, error) {
	if limit <= 0 {
		limit = 50
	}
	var items []models.GuestRequest
	return items, r.db.Where("hotel_id = ?", hotelID).Order("created_at DESC").Limit(limit).Find(&items).Error
}

func (r *GuestRequestRepository) ListAll(status string, limit int) ([]models.GuestRequest, error) {
	if limit <= 0 {
		limit = 100
	}
	var items []models.GuestRequest
	q := r.db.Order("created_at DESC").Limit(limit)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return items, q.Find(&items).Error
}
