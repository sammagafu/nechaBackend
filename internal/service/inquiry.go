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
	models.InquiryTypeHotelPartnership:      true,
	models.InquiryTypeBrandOnboarding:       true,
	models.InquiryTypeEventListing:          true,
	models.InquiryTypePartnerReferralSignup: true,
	models.InquiryTypeHotelPartner:          true,
	models.InquiryTypeBrandPartner:          true,
	models.InquiryTypeAffiliatePartner:      true,
	models.InquiryTypePartnerReferral:       true,
	models.InquiryTypeContact:               true,
	models.InquiryTypeNewsletter:            true,
	models.InquiryTypeDiscoveryBooking:      true,
}

func (s *InquiryService) Submit(req dto.SubmitInquiryRequest) (*dto.InquiryResponse, error) {
	inquiryType := models.NormalizeInquiryType(strings.TrimSpace(req.Type))
	if !allowedInquiryTypes[inquiryType] && !allowedInquiryTypes[strings.TrimSpace(req.Type)] {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid inquiry type", apperrors.ErrBadRequest.Status)
	}
	if !allowedInquiryTypes[inquiryType] {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid inquiry type", apperrors.ErrBadRequest.Status)
	}
	if inquiryType == models.InquiryTypeDiscoveryBooking && s.platform != nil &&
		!s.platform.FeatureEnabled(models.ConfigKeyFeatureDiscoveryTicketingEnabled, false) {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "discovery booking is not enabled", apperrors.ErrBadRequest.Status)
	}

	payload := map[string]any{}
	if req.Rooms > 0 {
		payload["rooms"] = req.Rooms
	}
	if req.Units > 0 {
		payload["units"] = req.Units
	}
	if req.Price > 0 {
		payload["price_tzs"] = req.Price
	}
	if req.PropertyType != "" {
		payload["property_type"] = strings.TrimSpace(req.PropertyType)
	}
	if req.PreferredCallTime != "" {
		payload["preferred_call_time"] = strings.TrimSpace(req.PreferredCallTime)
	}
	if req.BrandStory != "" {
		payload["brand_story"] = strings.TrimSpace(req.BrandStory)
	}
	if req.SampleProducts != "" {
		payload["sample_products"] = strings.TrimSpace(req.SampleProducts)
	}
	if req.ProposedCapacity > 0 {
		payload["proposed_capacity"] = req.ProposedCapacity
	}
	if req.ProposedDates != "" {
		payload["proposed_dates"] = strings.TrimSpace(req.ProposedDates)
	}
	if req.PartnerType != "" {
		payload["partner_type"] = strings.TrimSpace(req.PartnerType)
	}
	if req.ClientVolume != "" {
		payload["client_volume"] = strings.TrimSpace(req.ClientVolume)
	}
	if req.Destinations != "" {
		payload["destinations"] = strings.TrimSpace(req.Destinations)
	}

	payloadJSON := ""
	if len(payload) > 0 {
		b, _ := json.Marshal(payload)
		payloadJSON = string(b)
	}

	item := &models.Inquiry{
		Type:              inquiryType,
		Status:            models.InquiryStatusNew,
		Name:              strings.TrimSpace(req.Name),
		Email:             strings.TrimSpace(req.Email),
		Phone:             strings.TrimSpace(req.Phone),
		Company:           strings.TrimSpace(req.Company),
		Role:              strings.TrimSpace(req.Role),
		Location:          strings.TrimSpace(req.Location),
		Category:          strings.TrimSpace(req.Category),
		Message:           strings.TrimSpace(req.Message),
		Metadata:          payloadJSON,
		SubmissionPayload: payloadJSON,
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
	items, err := s.inquiries.List(status, models.NormalizeInquiryType(inquiryType))
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list inquiries", apperrors.ErrInternal.Status)
	}
	out := make([]dto.InquiryResponse, 0, len(items))
	for i := range items {
		out = append(out, toInquiryResponse(&items[i]))
	}
	return out, nil
}

func (s *InquiryService) AdminUpdateStatus(id string, req dto.UpdateInquiryRequest) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid inquiry id", apperrors.ErrBadRequest.Status)
	}
	status := strings.TrimSpace(req.Status)
	switch status {
	case models.InquiryStatusRead:
		status = models.InquiryStatusContacted
	case models.InquiryStatusNew, models.InquiryStatusContacted, models.InquiryStatusQualified,
		models.InquiryStatusConverted, models.InquiryStatusArchived:
	default:
		return apperrors.New(apperrors.ErrBadRequest.Code, "invalid status", apperrors.ErrBadRequest.Status)
	}

	updates := map[string]any{"status": status}
	if req.AssignedTo != nil {
		if strings.TrimSpace(*req.AssignedTo) == "" {
			updates["assigned_to"] = nil
		} else {
			aid, err := uuid.Parse(strings.TrimSpace(*req.AssignedTo))
			if err != nil {
				return apperrors.New(apperrors.ErrBadRequest.Code, "invalid assigned_to", apperrors.ErrBadRequest.Status)
			}
			updates["assigned_to"] = aid
		}
	}
	if req.ConvertedToID != nil {
		if strings.TrimSpace(*req.ConvertedToID) == "" {
			updates["converted_to_id"] = nil
		} else {
			cid, err := uuid.Parse(strings.TrimSpace(*req.ConvertedToID))
			if err != nil {
				return apperrors.New(apperrors.ErrBadRequest.Code, "invalid converted_to_id", apperrors.ErrBadRequest.Status)
			}
			updates["converted_to_id"] = cid
		}
	}
	if status == models.InquiryStatusConverted && req.ConvertedToID == nil {
		// Architecture §9.7: strongly prompt for converted_to_id — do not block if ops
		// is still creating the target entity in the same session.
	}

	if err := s.inquiries.UpdateFields(parsed, updates); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update inquiry", apperrors.ErrInternal.Status)
	}
	return nil
}

func toInquiryResponse(item *models.Inquiry) dto.InquiryResponse {
	resp := dto.InquiryResponse{
		ID:                item.ID.String(),
		Type:              item.Type,
		Status:            item.Status,
		Name:              item.Name,
		Email:             item.Email,
		Phone:             item.Phone,
		Company:           item.Company,
		Role:              item.Role,
		Location:          item.Location,
		Category:          item.Category,
		Message:           item.Message,
		Metadata:          item.Metadata,
		SubmissionPayload: item.SubmissionPayload,
		CreatedAt:         item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if item.AssignedTo != nil {
		s := item.AssignedTo.String()
		resp.AssignedTo = &s
	}
	if item.ConvertedToID != nil {
		s := item.ConvertedToID.String()
		resp.ConvertedToID = &s
	}
	return resp
}
