package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type InquiryRepository struct {
	db *gorm.DB
}

func NewInquiryRepository(db *gorm.DB) *InquiryRepository {
	return &InquiryRepository{db: db}
}

func (r *InquiryRepository) Create(inquiry *models.Inquiry) error {
	return r.db.Create(inquiry).Error
}

func (r *InquiryRepository) List(status, inquiryType string) ([]models.Inquiry, error) {
	var items []models.Inquiry
	q := r.db.Model(&models.Inquiry{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if inquiryType != "" {
		q = q.Where("type = ?", inquiryType)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *InquiryRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Inquiry{}).Where("id = ?", id).Update("status", status).Error
}
