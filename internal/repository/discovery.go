package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type DiscoveryRepository struct {
	db *gorm.DB
}

func NewDiscoveryRepository(db *gorm.DB) *DiscoveryRepository {
	return &DiscoveryRepository{db: db}
}

func (r *DiscoveryRepository) ListActiveForHotel(hotelID uuid.UUID, section string) ([]models.DiscoveryItem, error) {
	var items []models.DiscoveryItem
	q := r.db.Where("status = ?", models.DiscoveryStatusActive).
		Where("(hotel_id IS NULL OR hotel_id = ?)", hotelID)
	if section != "" {
		q = q.Where("section = ?", section)
	}
	err := q.Order("is_featured DESC, sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *DiscoveryRepository) FindByID(id uuid.UUID) (*models.DiscoveryItem, error) {
	var item models.DiscoveryItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DiscoveryRepository) ListAll(section, status string) ([]models.DiscoveryItem, error) {
	var items []models.DiscoveryItem
	q := r.db.Model(&models.DiscoveryItem{})
	if section != "" {
		q = q.Where("section = ?", section)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *DiscoveryRepository) Create(item *models.DiscoveryItem) error {
	return r.db.Create(item).Error
}

func (r *DiscoveryRepository) Save(item *models.DiscoveryItem) error {
	return r.db.Save(item).Error
}

func (r *DiscoveryRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.DiscoveryItem{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
