package repository

import (
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type DeliveryZoneRepository struct {
	db *gorm.DB
}

func NewDeliveryZoneRepository(db *gorm.DB) *DeliveryZoneRepository {
	return &DeliveryZoneRepository{db: db}
}

func (r *DeliveryZoneRepository) ListActive() ([]models.DeliveryZone, error) {
	var zones []models.DeliveryZone
	return zones, r.db.Where("is_active = ?", true).Order("sort_order ASC, label ASC").Find(&zones).Error
}

func (r *DeliveryZoneRepository) FindByCode(code string) (*models.DeliveryZone, error) {
	var zone models.DeliveryZone
	err := r.db.Where("code = ? AND is_active = ?", code, true).First(&zone).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *DeliveryZoneRepository) Upsert(zone *models.DeliveryZone) error {
	var existing models.DeliveryZone
	err := r.db.Where("code = ?", zone.Code).First(&existing).Error
	if err == nil {
		zone.ID = existing.ID
		return r.db.Save(zone).Error
	}
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(zone).Error
	}
	return err
}
