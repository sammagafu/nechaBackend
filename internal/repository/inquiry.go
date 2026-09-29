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

func (r *InquiryRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Inquiry{}).Where("id = ?", id).Update("status", status).Error
}

func (r *InquiryRepository) UpdateFields(id uuid.UUID, updates map[string]any) error {
	return r.db.Model(&models.Inquiry{}).Where("id = ?", id).Updates(updates).Error
}

func (r *InquiryRepository) List(status, inquiryType string) ([]models.Inquiry, error) {
	var items []models.Inquiry
	q := r.db.Model(&models.Inquiry{})
	if status != "" {
		if status == models.InquiryStatusContacted {
			q = q.Where("status IN ?", []string{models.InquiryStatusContacted, models.InquiryStatusRead})
		} else {
			q = q.Where("status = ?", status)
		}
	}
	if inquiryType != "" {
		aliases := []string{inquiryType}
		switch inquiryType {
		case models.InquiryTypeHotelPartnership:
			aliases = append(aliases, models.InquiryTypeHotelPartner)
		case models.InquiryTypeBrandOnboarding:
			aliases = append(aliases, models.InquiryTypeBrandPartner)
		case models.InquiryTypePartnerReferralSignup:
			aliases = append(aliases, models.InquiryTypeAffiliatePartner, models.InquiryTypePartnerReferral)
		}
		q = q.Where("type IN ?", aliases)
	}
	err := q.Order("created_at DESC").Find(&items).Error
	return items, err
}
