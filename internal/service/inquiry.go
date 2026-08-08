package service

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
)

type InquiryService struct {
	inquiries *repository.InquiryRepository
	events    *EventService
	platform  *PlatformService
}

func NewInquiryService(inquiries *repository.InquiryRepository, events *EventService, platform *PlatformService) *InquiryService {
	return &InquiryService{inquiries: inquiries, events: events, platform: platform}
}

var allowedInquiryTypes = map[string]bool{
	models.InquiryTypeHotelPartner:     true,
	models.InquiryTypeBrandPartner:     true,
	models.InquiryTypeAffiliatePartner: true,
	models.InquiryTypeContact:          true,
	models.InquiryTypeNewsletter:       true,
	models.InquiryTypeEventListing:     true,
	models.InquiryTypePartnerReferral:  true,
	models.InquiryTypeDiscoveryBooking: true,
}

func (s *InquiryService) Submit(req dto.SubmitInquiryRequest) (*dto.InquiryResponse, error) {
	inquiryType := strings.TrimSpace(req.Type)
	if !allowedInquiryTypes[inquiryType] {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid inquiry type", apperrors.ErrBadRequest.Status)
	}
	if inquiryType == models.InquiryTypeDiscoveryBooking && s.platform != nil &&
		!s.platform.FeatureEnabled(models.ConfigKeyFeatureDiscoveryTicketingEnabled, false) {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "discovery booking is not enabled", apperrors.ErrBadRequest.Status)
	}

	meta := map[string]int{}
	if req.Rooms > 0 {
		meta["rooms"] = req.Rooms
	}
	if req.Units > 0 {
		meta["units"] = req.Units
	}
	if req.Price > 0 {
		meta["price_tzs"] = req.Price
	}
	metaJSON := ""
	if len(meta) > 0 {
		b, _ := json.Marshal(meta)
		metaJSON = string(b)
	}

	item := &models.Inquiry{
		Type:     inquiryType,
		Status:   models.InquiryStatusNew,
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Company:  strings.TrimSpace(req.Company),
		Role:     strings.TrimSpace(req.Role),
		Location: strings.TrimSpace(req.Location),
		Category: strings.TrimSpace(req.Category),
		Message:  strings.TrimSpace(req.Message),
		Metadata: metaJSON,
	}

	if err := s.inquiries.Create(item); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save inquiry", apperrors.ErrInternal.Status)
	}

	if s.events != nil {
		s.events.InquiryCreated(item)
	}

	resp := toInquiryResponse(item)
	return &resp, nil
}

func (s *InquiryService) AdminList(status, inquiryType string) ([]dto.InquiryResponse, error) {
	items, err := s.inquiries.List(status, inquiryType)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list inquiries", apperrors.ErrInternal.Status)
	}
	out := make([]dto.InquiryResponse, 0, len(items))
	for i := range items {
		out = append(out, toInquiryResponse(&items[i]))
	}
	return out, nil
}

func (s *InquiryService) AdminUpdateStatus(id string, status string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid inquiry id", apperrors.ErrBadRequest.Status)
	}
	switch status {
	case models.InquiryStatusNew, models.InquiryStatusRead, models.InquiryStatusArchived:
	default:
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid status", apperrors.ErrBadRequest.Status)
	}
	if err := s.inquiries.UpdateStatus(parsed, status); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update inquiry", apperrors.ErrInternal.Status)
	}
	return nil
}

func toInquiryResponse(item *models.Inquiry) dto.InquiryResponse {
	return dto.InquiryResponse{
		ID:        item.ID.String(),
		Type:      item.Type,
		Status:    item.Status,
		Name:      item.Name,
		Email:     item.Email,
		Phone:     item.Phone,
		Company:   item.Company,
		Role:      item.Role,
		Location:  item.Location,
		Category:  item.Category,
		Message:   item.Message,
		Metadata:  item.Metadata,
		CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
