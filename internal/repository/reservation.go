package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(reservation *models.Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *ReservationRepository) FindByID(id uuid.UUID) (*models.Reservation, error) {
	var reservation models.Reservation
	err := r.db.Preload("Hotel").First(&reservation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) Update(reservation *models.Reservation) error {
	return r.db.Save(reservation).Error
}

func (r *ReservationRepository) List(limit, offset int) ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Preload("Hotel").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).Count(&count).Error
	return count, err
}

func (r *ReservationRepository) CountByStatus(status models.ReservationStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *ReservationRepository) CountSince(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).Where("created_at >= ?", since).Count(&count).Error
	return count, err
}

func (r *ReservationRepository) CountByHotel(hotelID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).Where("hotel_id = ?", hotelID).Count(&count).Error
	return count, err
}

func (r *ReservationRepository) CountByHotelAndStatus(hotelID uuid.UUID, status models.ReservationStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status = ?", hotelID, status).
		Count(&count).Error
	return count, err
}

func (r *ReservationRepository) ListByHotel(hotelID uuid.UUID, limit int) ([]models.Reservation, error) {
	if limit <= 0 || limit > 50 {
		limit = 8
	}
	var reservations []models.Reservation
	err := r.db.Preload("Hotel").
		Where("hotel_id = ?", hotelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&reservations).Error
	return reservations, err
}
