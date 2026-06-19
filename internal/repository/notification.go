package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepository) CreateBatch(items []models.Notification) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *NotificationRepository) ListByUser(userID uuid.UUID, limit int) ([]models.Notification, error) {
	var items []models.Notification
	q := r.db.Where("user_id = ? AND read_at IS NULL", userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}

func (r *NotificationRepository) CountUnread(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkRead(userID uuid.UUID, ids []uuid.UUID) error {
	q := r.db.Where("user_id = ?", userID)
	if len(ids) > 0 {
		q = q.Where("id IN ?", ids)
	} else {
		q = q.Where("read_at IS NULL")
	}
	return q.Delete(&models.Notification{}).Error
}

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) ListActive() ([]models.SystemAlert, error) {
	now := time.Now()
	var items []models.SystemAlert
	err := r.db.Where("is_active = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *AlertRepository) ListAll() ([]models.SystemAlert, error) {
	var items []models.SystemAlert
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *AlertRepository) Create(a *models.SystemAlert) error {
	return r.db.Create(a).Error
}

func (r *AlertRepository) FindByID(id uuid.UUID) (*models.SystemAlert, error) {
	var item models.SystemAlert
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AlertRepository) Update(a *models.SystemAlert) error {
	return r.db.Save(a).Error
}
