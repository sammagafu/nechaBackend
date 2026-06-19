package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type AlertService struct {
	repo *repository.AlertRepository
}

func NewAlertService(repo *repository.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

func (s *AlertService) ListActive() ([]dto.SystemAlertResponse, error) {
	items, err := s.repo.ListActive()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list alerts", apperrors.ErrInternal.Status)
	}
	result := make([]dto.SystemAlertResponse, 0, len(items))
	for _, a := range items {
		result = append(result, toAlertResponse(&a))
	}
	return result, nil
}

func (s *AlertService) ListAll() ([]dto.SystemAlertResponse, error) {
	items, err := s.repo.ListAll()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list alerts", apperrors.ErrInternal.Status)
	}
	result := make([]dto.SystemAlertResponse, 0, len(items))
	for _, a := range items {
		result = append(result, toAlertResponse(&a))
	}
	return result, nil
}

func (s *AlertService) Create(req dto.CreateSystemAlertRequest) (*dto.SystemAlertResponse, error) {
	severity := models.NotificationSeverityInfo
	if req.Severity != "" {
		severity = models.NotificationSeverity(req.Severity)
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	alert := &models.SystemAlert{
		Title:    req.Title,
		Body:     req.Body,
		Severity: severity,
		Link:     req.Link,
		IsActive: active,
		StartsAt: req.StartsAt,
		EndsAt:   req.EndsAt,
	}
	if err := s.repo.Create(alert); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create alert", apperrors.ErrInternal.Status)
	}
	resp := toAlertResponse(alert)
	return &resp, nil
}

func (s *AlertService) Update(id string, req dto.CreateSystemAlertRequest) (*dto.SystemAlertResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid alert id", apperrors.ErrBadRequest.Status)
	}
	alert, err := s.repo.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "alert not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load alert", apperrors.ErrInternal.Status)
	}
	if req.Title != "" {
		alert.Title = req.Title
	}
	if req.Body != "" {
		alert.Body = req.Body
	}
	if req.Severity != "" {
		alert.Severity = models.NotificationSeverity(req.Severity)
	}
	alert.Link = req.Link
	if req.IsActive != nil {
		alert.IsActive = *req.IsActive
	}
	if req.StartsAt != nil {
		alert.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		alert.EndsAt = req.EndsAt
	}
	if err := s.repo.Update(alert); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update alert", apperrors.ErrInternal.Status)
	}
	resp := toAlertResponse(alert)
	return &resp, nil
}

func toAlertResponse(a *models.SystemAlert) dto.SystemAlertResponse {
	return dto.SystemAlertResponse{
		ID:       a.ID.String(),
		Title:    a.Title,
		Body:     a.Body,
		Severity: string(a.Severity),
		Link:     a.Link,
		IsActive: a.IsActive,
		StartsAt: a.StartsAt,
		EndsAt:   a.EndsAt,
	}
}
