package repository

import (
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type PlatformConfigRepository struct {
	db *gorm.DB
}

func NewPlatformConfigRepository(db *gorm.DB) *PlatformConfigRepository {
	return &PlatformConfigRepository{db: db}
}

func (r *PlatformConfigRepository) Get(key string) (string, error) {
	var row models.PlatformConfig
	err := r.db.Where("key = ?", key).First(&row).Error
	if err != nil {
		return "", err
	}
	return row.Value, nil
}

func (r *PlatformConfigRepository) Set(key, value string) error {
	var row models.PlatformConfig
	err := r.db.Where("key = ?", key).First(&row).Error
	if err == nil {
		row.Value = value
		return r.db.Save(&row).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.Create(&models.PlatformConfig{Key: key, Value: value}).Error
}
