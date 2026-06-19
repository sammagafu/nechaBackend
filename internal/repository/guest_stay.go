package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type GuestStayRepository struct {
	db *gorm.DB
}

func NewGuestStayRepository(db *gorm.DB) *GuestStayRepository {
	return &GuestStayRepository{db: db}
}

func (r *GuestStayRepository) Create(stay *models.GuestStay) error {
	return r.db.Create(stay).Error
}

func (r *GuestStayRepository) List(limit, offset int) ([]models.GuestStay, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var stays []models.GuestStay
	err := r.db.
		Preload("Hotel").
		Preload("User").
		Preload("Order.Items").
		Preload("Reservation").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&stays).Error
	return stays, err
}

func (r *GuestStayRepository) ListByUser(userID uuid.UUID, limit int) ([]models.GuestStay, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var stays []models.GuestStay
	err := r.db.
		Preload("Hotel").
		Preload("Order.Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&stays).Error
	return stays, err
}
