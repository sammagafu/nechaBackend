package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type WebhookService struct {
	repo       *repository.WebhookRepository
	httpClient *http.Client
	inboundKey string
}

func NewWebhookService(repo *repository.WebhookRepository, inboundKey string) *WebhookService {
	return &WebhookService{
		repo:       repo,
		inboundKey: inboundKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *WebhookService) ListEndpoints() ([]dto.WebhookEndpointResponse, error) {
	items, err := s.repo.ListEndpoints()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list webhooks", apperrors.ErrInternal.Status)
	}
	result := make([]dto.WebhookEndpointResponse, 0, len(items))
	for _, e := range items {
		result = append(result, toWebhookEndpointResponse(&e))
	}
	return result, nil
}

func (s *WebhookService) CreateEndpoint(req dto.CreateWebhookRequest) (*dto.WebhookEndpointResponse, string, error) {
	secret, err := randomSecret()
	if err != nil {
		return nil, "", apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to generate secret", apperrors.ErrInternal.Status)
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	endpoint := &models.WebhookEndpoint{
		URL:         req.URL,
		Secret:      secret,
		Description: req.Description,
		Events:      models.StringSlice(req.Events),
		IsActive:    active,
	}
	if err := s.repo.CreateEndpoint(endpoint); err != nil {
		return nil, "", apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create webhook", apperrors.ErrInternal.Status)
	}
	resp := toWebhookEndpointResponse(endpoint)
	return &resp, secret, nil
}

func (s *WebhookService) UpdateEndpoint(id string, req dto.UpdateWebhookRequest) (*dto.WebhookEndpointResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid webhook id", apperrors.ErrBadRequest.Status)
	}
	endpoint, err := s.repo.FindEndpoint(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "webhook not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load webhook", apperrors.ErrInternal.Status)
	}
	if req.URL != nil {
		endpoint.URL = *req.URL
	}
	if req.Description != nil {
		endpoint.Description = *req.Description
	}
	if len(req.Events) > 0 {
		endpoint.Events = models.StringSlice(req.Events)
	}
	if req.IsActive != nil {
		endpoint.IsActive = *req.IsActive
	}
	if err := s.repo.UpdateEndpoint(endpoint); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update webhook", apperrors.ErrInternal.Status)
	}
	resp := toWebhookEndpointResponse(endpoint)
	return &resp, nil
}

func (s *WebhookService) ListDeliveries(limit int) ([]dto.WebhookDeliveryResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items, err := s.repo.ListDeliveries(limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list deliveries", apperrors.ErrInternal.Status)
	}
	result := make([]dto.WebhookDeliveryResponse, 0, len(items))
	for _, d := range items {
		result = append(result, toWebhookDeliveryResponse(&d))
	}
	return result, nil
}

func (s *WebhookService) DispatchEvent(ctx context.Context, event string, payload map[string]interface{}) {
	endpoints, err := s.repo.ListActiveForEvent(event)
	if err != nil || len(endpoints) == 0 {
		return
	}
	body, _ := json.Marshal(map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	})
	for _, endpoint := range endpoints {
		go s.deliver(context.Background(), endpoint, event, string(body))
	}
}

func (s *WebhookService) deliver(ctx context.Context, endpoint models.WebhookEndpoint, event, body string) {
	delivery := &models.WebhookDelivery{
		EndpointID: endpoint.ID,
		EventType:  event,
		Payload:    body,
		Status:     models.WebhookDeliveryPending,
		Attempts:   1,
	}
	_ = s.repo.CreateDelivery(delivery)

	sig := signPayload(endpoint.Secret, body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewBufferString(body))
	if err != nil {
		delivery.Status = models.WebhookDeliveryFailed
		delivery.ResponseBody = err.Error()
		_ = s.repo.UpdateDelivery(delivery)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Necha-Event", event)
	req.Header.Set("X-Necha-Signature", sig)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		delivery.Status = models.WebhookDeliveryFailed
		delivery.ResponseBody = err.Error()
		_ = s.repo.UpdateDelivery(delivery)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	delivery.ResponseCode = resp.StatusCode
	delivery.ResponseBody = string(respBody)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = models.WebhookDeliverySuccess
	} else {
		delivery.Status = models.WebhookDeliveryFailed
	}
	_ = s.repo.UpdateDelivery(delivery)
}

func (s *WebhookService) VerifyInbound(secret string) bool {
	if s.inboundKey == "" {
		return false
	}
	return secret != "" && hmac.Equal([]byte(secret), []byte(s.inboundKey))
}

func (s *WebhookService) HandleInbound(req dto.InboundWebhookRequest) error {
	switch req.Event {
	case "order.status_updated", "reservation.status_updated":
		return nil
	default:
		return nil
	}
}

func signPayload(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func randomSecret() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toWebhookEndpointResponse(e *models.WebhookEndpoint) dto.WebhookEndpointResponse {
	return dto.WebhookEndpointResponse{
		ID:          e.ID.String(),
		URL:         e.URL,
		Description: e.Description,
		Events:      []string(e.Events),
		IsActive:    e.IsActive,
		CreatedAt:   e.CreatedAt,
	}
}

func toWebhookDeliveryResponse(d *models.WebhookDelivery) dto.WebhookDeliveryResponse {
	return dto.WebhookDeliveryResponse{
		ID:           d.ID.String(),
		EndpointID:   d.EndpointID.String(),
		EventType:    d.EventType,
		Status:       string(d.Status),
		ResponseCode: d.ResponseCode,
		Attempts:     d.Attempts,
		CreatedAt:    d.CreatedAt,
	}
}
