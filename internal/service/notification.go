package service

import (
	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ListForUser(userID uuid.UUID, limit int) (*dto.NotificationListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	items, err := s.repo.ListByUser(userID, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list notifications", apperrors.ErrInternal.Status)
	}
	unread, err := s.repo.CountUnread(userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to count notifications", apperrors.ErrInternal.Status)
	}
	resp := &dto.NotificationListResponse{
		Items:       make([]dto.NotificationResponse, 0, len(items)),
		UnreadCount: int(unread),
	}
	for _, n := range items {
		resp.Items = append(resp.Items, toNotificationResponse(&n))
	}
	return resp, nil
}

func (s *NotificationService) MarkRead(userID uuid.UUID, ids []string) error {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return apperrors.New(apperrors.ErrBadRequest.Code, "invalid notification id", apperrors.ErrBadRequest.Status)
		}
		parsed = append(parsed, uid)
	}
	if err := s.repo.MarkRead(userID, parsed); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to delete notifications", apperrors.ErrInternal.Status)
	}
	return nil
}

func (s *NotificationService) NotifyUser(userID uuid.UUID, nType, title, body, link string, severity models.NotificationSeverity) error {
	if severity == "" {
		severity = models.NotificationSeverityInfo
	}
	return s.repo.Create(&models.Notification{
		UserID:   userID,
		Type:     nType,
		Title:    title,
		Body:     body,
		Link:     link,
		Severity: severity,
	})
}

func (s *NotificationService) NotifyUsers(userIDs []uuid.UUID, nType, title, body, link string, severity models.NotificationSeverity) error {
	if len(userIDs) == 0 {
		return nil
	}
	if severity == "" {
		severity = models.NotificationSeverityInfo
	}
	items := make([]models.Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		items = append(items, models.Notification{
			UserID:   uid,
			Type:     nType,
			Title:    title,
			Body:     body,
			Link:     link,
			Severity: severity,
		})
	}
	return s.repo.CreateBatch(items)
}

func toNotificationResponse(n *models.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:        n.ID.String(),
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Link:      n.Link,
		Severity:  string(n.Severity),
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
