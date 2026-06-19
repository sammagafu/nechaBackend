package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type DiscoveryService struct {
	discovery *repository.DiscoveryRepository
	hotels    *repository.HotelRepository
}

func NewDiscoveryService(discovery *repository.DiscoveryRepository, hotels *repository.HotelRepository) *DiscoveryService {
	return &DiscoveryService{discovery: discovery, hotels: hotels}
}

func (s *DiscoveryService) PortalByHotelSlug(slug string) (*dto.DiscoveryPortalResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	events, err := s.discovery.ListActiveForHotel(hotel.ID, models.DiscoverySectionEvents)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load events", apperrors.ErrInternal.Status)
	}
	restaurants, err := s.discovery.ListActiveForHotel(hotel.ID, models.DiscoverySectionRestaurants)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load restaurants", apperrors.ErrInternal.Status)
	}
	tours, err := s.discovery.ListActiveForHotel(hotel.ID, models.DiscoverySectionTours)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load tours", apperrors.ErrInternal.Status)
	}

	return &dto.DiscoveryPortalResponse{
		Events:      toDiscoveryResponses(events),
		Restaurants:   toDiscoveryResponses(restaurants),
		Tours:         toDiscoveryResponses(tours),
	}, nil
}

func (s *DiscoveryService) SubmitEvent(req dto.SubmitDiscoveryEventRequest) (*dto.DiscoveryItemResponse, error) {
	startsAt, err := parseDiscoveryTime(req.EventStartsAt)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid event start date", apperrors.ErrBadRequest.Status)
	}
	var endsAt *time.Time
	if req.EventEndsAt != "" {
		t, err := parseDiscoveryTime(req.EventEndsAt)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid event end date", apperrors.ErrBadRequest.Status)
		}
		endsAt = &t
	}

	ticketMode := req.TicketMode
	if ticketMode == "" {
		if req.TicketURL != "" {
			ticketMode = models.TicketModeReferral
		} else {
			ticketMode = models.TicketModeNone
		}
	}

	var hotelID *uuid.UUID
	if req.HotelSlug != "" {
		hotel, err := s.hotels.FindBySlug(req.HotelSlug)
		if err == nil {
			hotelID = &hotel.ID
		}
	}

	slug := slugifyDiscovery(req.Name)
	item := &models.DiscoveryItem{
		HotelID:        hotelID,
		Section:        models.DiscoverySectionEvents,
		Subcategory:    req.Subcategory,
		Slug:           slug,
		Name:           req.Name,
		Description:    req.Description,
		Venue:          req.Venue,
		Location:       req.Location,
		Phone:          req.OrganizerPhone,
		Website:        req.Website,
		EventStartsAt:  &startsAt,
		EventEndsAt:    endsAt,
		TicketURL:      req.TicketURL,
		TicketMode:     ticketMode,
		OrganizerName:  req.OrganizerName,
		OrganizerEmail: req.OrganizerEmail,
		Status:         models.DiscoveryStatusPending,
	}

	if err := s.discovery.Create(item); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to submit event", apperrors.ErrInternal.Status)
	}

	resp := toDiscoveryResponse(item)
	return &resp, nil
}

func (s *DiscoveryService) AdminList(section, status string) ([]dto.AdminDiscoveryItemResponse, error) {
	items, err := s.discovery.ListAll(section, status)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list discovery items", apperrors.ErrInternal.Status)
	}
	result := make([]dto.AdminDiscoveryItemResponse, 0, len(items))
	for i := range items {
		result = append(result, toAdminDiscoveryResponse(&items[i]))
	}
	return result, nil
}

func (s *DiscoveryService) AdminGet(id string) (*dto.AdminDiscoveryItemResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid id", apperrors.ErrBadRequest.Status)
	}
	item, err := s.discovery.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "item not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load item", apperrors.ErrInternal.Status)
	}
	resp := toAdminDiscoveryResponse(item)
	return &resp, nil
}

func (s *DiscoveryService) AdminCreate(req dto.CreateDiscoveryItemRequest) (*dto.AdminDiscoveryItemResponse, error) {
	item, err := s.buildDiscoveryItem(req)
	if err != nil {
		return nil, err
	}
	if item.Status == "" {
		item.Status = models.DiscoveryStatusActive
	}
	if err := s.discovery.Create(item); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create item", apperrors.ErrInternal.Status)
	}
	resp := toAdminDiscoveryResponse(item)
	return &resp, nil
}

func (s *DiscoveryService) AdminUpdate(id string, req dto.UpdateDiscoveryItemRequest) (*dto.AdminDiscoveryItemResponse, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid id", apperrors.ErrBadRequest.Status)
	}
	item, err := s.discovery.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "item not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load item", apperrors.ErrInternal.Status)
	}

	if err := applyDiscoveryUpdate(item, req); err != nil {
		return nil, err
	}
	if err := s.discovery.Save(item); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update item", apperrors.ErrInternal.Status)
	}
	resp := toAdminDiscoveryResponse(item)
	return &resp, nil
}

func (s *DiscoveryService) PendingCount() (int64, error) {
	return s.discovery.CountByStatus(models.DiscoveryStatusPending)
}

func (s *DiscoveryService) buildDiscoveryItem(req dto.CreateDiscoveryItemRequest) (*models.DiscoveryItem, error) {
	var hotelID *uuid.UUID
	if req.HotelID != nil && *req.HotelID != "" {
		uid, err := uuid.Parse(*req.HotelID)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
		}
		hotelID = &uid
	}

	slug := req.Slug
	if slug == "" {
		slug = slugifyDiscovery(req.Name)
	}

	ticketMode := req.TicketMode
	if ticketMode == "" {
		ticketMode = models.TicketModeNone
	}

	item := &models.DiscoveryItem{
		HotelID:        hotelID,
		Section:        req.Section,
		Subcategory:    req.Subcategory,
		Slug:           slug,
		Name:           req.Name,
		Description:    req.Description,
		ImageURL:       req.ImageURL,
		Venue:          req.Venue,
		Location:       req.Location,
		Distance:       req.Distance,
		Phone:          req.Phone,
		Website:        req.Website,
		PriceHint:      req.PriceHint,
		TicketURL:      req.TicketURL,
		TicketMode:     ticketMode,
		OrganizerName:  req.OrganizerName,
		OrganizerEmail: req.OrganizerEmail,
		IsFeatured:     req.IsFeatured,
		Status:         req.Status,
		SortOrder:      req.SortOrder,
	}

	if req.EventStartsAt != "" {
		t, err := parseDiscoveryTime(req.EventStartsAt)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid event start date", apperrors.ErrBadRequest.Status)
		}
		item.EventStartsAt = &t
	}
	if req.EventEndsAt != "" {
		t, err := parseDiscoveryTime(req.EventEndsAt)
		if err != nil {
			return nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid event end date", apperrors.ErrBadRequest.Status)
		}
		item.EventEndsAt = &t
	}

	return item, nil
}

func applyDiscoveryUpdate(item *models.DiscoveryItem, req dto.UpdateDiscoveryItemRequest) error {
	if req.HotelID != nil {
		if *req.HotelID == "" {
			item.HotelID = nil
		} else {
			uid, err := uuid.Parse(*req.HotelID)
			if err != nil {
				return apperrors.New(apperrors.ErrBadRequest.Code, "invalid hotel id", apperrors.ErrBadRequest.Status)
			}
			item.HotelID = &uid
		}
	}
	if req.Section != nil {
		item.Section = *req.Section
	}
	if req.Subcategory != nil {
		item.Subcategory = *req.Subcategory
	}
	if req.Slug != nil {
		item.Slug = *req.Slug
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.ImageURL != nil {
		item.ImageURL = *req.ImageURL
	}
	if req.Venue != nil {
		item.Venue = *req.Venue
	}
	if req.Location != nil {
		item.Location = *req.Location
	}
	if req.Distance != nil {
		item.Distance = *req.Distance
	}
	if req.Phone != nil {
		item.Phone = *req.Phone
	}
	if req.Website != nil {
		item.Website = *req.Website
	}
	if req.PriceHint != nil {
		item.PriceHint = *req.PriceHint
	}
	if req.EventStartsAt != nil {
		if *req.EventStartsAt == "" {
			item.EventStartsAt = nil
		} else {
			t, err := parseDiscoveryTime(*req.EventStartsAt)
			if err != nil {
				return apperrors.New(apperrors.ErrBadRequest.Code, "invalid event start date", apperrors.ErrBadRequest.Status)
			}
			item.EventStartsAt = &t
		}
	}
	if req.EventEndsAt != nil {
		if *req.EventEndsAt == "" {
			item.EventEndsAt = nil
		} else {
			t, err := parseDiscoveryTime(*req.EventEndsAt)
			if err != nil {
				return apperrors.New(apperrors.ErrBadRequest.Code, "invalid event end date", apperrors.ErrBadRequest.Status)
			}
			item.EventEndsAt = &t
		}
	}
	if req.TicketURL != nil {
		item.TicketURL = *req.TicketURL
	}
	if req.TicketMode != nil {
		item.TicketMode = *req.TicketMode
	}
	if req.OrganizerName != nil {
		item.OrganizerName = *req.OrganizerName
	}
	if req.OrganizerEmail != nil {
		item.OrganizerEmail = *req.OrganizerEmail
	}
	if req.IsFeatured != nil {
		item.IsFeatured = *req.IsFeatured
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	return nil
}

func parseDiscoveryTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid time")
}

func slugifyDiscovery(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "&", "and")
	replacer := strings.NewReplacer(" ", "-", "/", "-", ",", "", ".", "")
	s = replacer.Replace(s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

func toDiscoveryResponses(items []models.DiscoveryItem) []dto.DiscoveryItemResponse {
	result := make([]dto.DiscoveryItemResponse, 0, len(items))
	for i := range items {
		result = append(result, toDiscoveryResponse(&items[i]))
	}
	return result
}

func toDiscoveryResponse(item *models.DiscoveryItem) dto.DiscoveryItemResponse {
	return dto.DiscoveryItemResponse{
		ID:            item.ID.String(),
		Section:       item.Section,
		Subcategory:   item.Subcategory,
		Slug:          item.Slug,
		Name:          item.Name,
		Description:   item.Description,
		ImageURL:      item.ImageURL,
		Venue:         item.Venue,
		Location:      item.Location,
		Distance:      item.Distance,
		Phone:         item.Phone,
		Website:       item.Website,
		PriceHint:     item.PriceHint,
		EventStartsAt: item.EventStartsAt,
		EventEndsAt:   item.EventEndsAt,
		TicketURL:     item.TicketURL,
		TicketMode:    item.TicketMode,
		OrganizerName: item.OrganizerName,
		IsFeatured:    item.IsFeatured,
	}
}

func toAdminDiscoveryResponse(item *models.DiscoveryItem) dto.AdminDiscoveryItemResponse {
	var hotelID *string
	if item.HotelID != nil {
		s := item.HotelID.String()
		hotelID = &s
	}
	return dto.AdminDiscoveryItemResponse{
		ID:             item.ID.String(),
		HotelID:        hotelID,
		Section:        item.Section,
		Subcategory:    item.Subcategory,
		Slug:           item.Slug,
		Name:           item.Name,
		Description:    item.Description,
		ImageURL:       item.ImageURL,
		Venue:          item.Venue,
		Location:       item.Location,
		Distance:       item.Distance,
		Phone:          item.Phone,
		Website:        item.Website,
		PriceHint:      item.PriceHint,
		EventStartsAt:  item.EventStartsAt,
		EventEndsAt:    item.EventEndsAt,
		TicketURL:      item.TicketURL,
		TicketMode:     item.TicketMode,
		OrganizerName:  item.OrganizerName,
		OrganizerEmail: item.OrganizerEmail,
		IsFeatured:     item.IsFeatured,
		Status:         item.Status,
		SortOrder:      item.SortOrder,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}
