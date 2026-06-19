package repository

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) CreateEndpoint(e *models.WebhookEndpoint) error {
	return r.db.Create(e).Error
}

func (r *WebhookRepository) UpdateEndpoint(e *models.WebhookEndpoint) error {
	return r.db.Save(e).Error
}

func (r *WebhookRepository) FindEndpoint(id uuid.UUID) (*models.WebhookEndpoint, error) {
	var e models.WebhookEndpoint
	err := r.db.First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *WebhookRepository) ListEndpoints() ([]models.WebhookEndpoint, error) {
	var items []models.WebhookEndpoint
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *WebhookRepository) ListActiveForEvent(event string) ([]models.WebhookEndpoint, error) {
	var items []models.WebhookEndpoint
	err := r.db.Where("is_active = ?", true).Find(&items).Error
	if err != nil {
		return nil, err
	}
	matched := make([]models.WebhookEndpoint, 0)
	for _, e := range items {
		for _, ev := range e.Events {
			if ev == event || ev == "*" {
				matched = append(matched, e)
				break
			}
		}
	}
	return matched, nil
}

func (r *WebhookRepository) CreateDelivery(d *models.WebhookDelivery) error {
	return r.db.Create(d).Error
}

func (r *WebhookRepository) UpdateDelivery(d *models.WebhookDelivery) error {
	return r.db.Save(d).Error
}

func (r *WebhookRepository) ListDeliveries(limit int) ([]models.WebhookDelivery, error) {
	var items []models.WebhookDelivery
	q := r.db.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}
