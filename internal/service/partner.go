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

// PartnerService exposes property-scoped operations for partner-role users (brief §11).
type PartnerService struct {
	users       *repository.UserRepository
	admin       *AdminService
	commissions *repository.CommissionRepository
	guestStays  *repository.GuestStayRepository
	hotels      *repository.HotelRepository
	bookings    *repository.BookingReferralRepository
}

func NewPartnerService(
	users *repository.UserRepository,
	admin *AdminService,
	commissions *repository.CommissionRepository,
	guestStays *repository.GuestStayRepository,
	hotels *repository.HotelRepository,
	bookings *repository.BookingReferralRepository,
) *PartnerService {
	return &PartnerService{
		users:       users,
		admin:       admin,
		commissions: commissions,
		guestStays:  guestStays,
		hotels:      hotels,
		bookings:    bookings,
	}
}

func (s *PartnerService) hotelIDForUser(userID string) (uuid.UUID, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, apperrors.New(apperrors.ErrBadRequest.Code, "invalid user id", apperrors.ErrBadRequest.Status)
	}
	user, err := s.users.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.New(apperrors.ErrNotFound.Code, "user not found", apperrors.ErrNotFound.Status)
		}
		return uuid.Nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load user", apperrors.ErrInternal.Status)
	}
	if user.Role != models.UserRolePartner && user.Role != models.UserRoleAdmin {
		return uuid.Nil, apperrors.ErrForbidden
	}
	if user.HotelID == nil {
		return uuid.Nil, apperrors.New(apperrors.ErrForbidden.Code, "partner account is not linked to a property", apperrors.ErrForbidden.Status)
	}
	return *user.HotelID, nil
}

func (s *PartnerService) Dashboard(userID string) (*dto.StoreDashboard, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.StoreDashboard(hotelID.String())
}

func (s *PartnerService) ListOrders(userID string) ([]dto.AdminOrderResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.ListOrdersByHotel(hotelID.String(), 200)
}

func (s *PartnerService) ListProducts(userID string) ([]dto.AdminProductResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.ListProducts(hotelID.String())
}

func (s *PartnerService) UpdateProduct(userID, productID string, req dto.UpdateProductRequest) (*dto.AdminProductResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	products, err := s.admin.ListProducts(hotelID.String())
	if err != nil {
		return nil, err
	}
	allowed := false
	for _, p := range products {
		if p.ID == productID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, apperrors.New(apperrors.ErrForbidden.Code, "product does not belong to your property", apperrors.ErrForbidden.Status)
	}
	return s.admin.UpdateProduct(productID, req)
}

func (s *PartnerService) ListMenu(userID, menuKind string) ([]dto.MenuItemResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.ListMenuItems(hotelID.String(), menuKind)
}

func (s *PartnerService) CreateMenuItem(userID string, req dto.CreateMenuItemRequest) (*dto.MenuItemResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.CreateMenuItem(hotelID.String(), req)
}

func (s *PartnerService) UpdateMenuItem(userID, itemID string, req dto.UpdateMenuItemRequest) (*dto.MenuItemResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	items, err := s.admin.ListMenuItems(hotelID.String(), "")
	if err != nil {
		return nil, err
	}
	allowed := false
	for _, item := range items {
		if item.ID == itemID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, apperrors.New(apperrors.ErrForbidden.Code, "menu item does not belong to your property", apperrors.ErrForbidden.Status)
	}
	return s.admin.UpdateMenuItem(itemID, req)
}

func (s *PartnerService) ListGuestStays(userID string) ([]dto.AdminGuestStayResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	stays, err := s.guestStays.ListByHotel(hotelID, 50)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list guest stays", apperrors.ErrInternal.Status)
	}
	out := make([]dto.AdminGuestStayResponse, 0, len(stays))
	for i := range stays {
		out = append(out, toAdminGuestStayResponse(&stays[i]))
	}
	return out, nil
}

func (s *PartnerService) ListCommissions(userID string) ([]dto.PartnerCommissionResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	records, err := s.commissions.ListRecordsByHotel(hotelID, 100)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list commissions", apperrors.ErrInternal.Status)
	}
	out := make([]dto.PartnerCommissionResponse, 0, len(records))
	for _, r := range records {
		out = append(out, dto.PartnerCommissionResponse{
			ID:            r.ID.String(),
			OrderID:       r.OrderID.String(),
			Category:      r.Category,
			GMV:           r.GMV,
			PropertyShare: r.PropertyShare,
			Status:        r.Status,
			CreatedAt:     r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *PartnerService) GetSettings(userID string) (*dto.PartnerSettingsResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	hotel, err := s.hotels.FindByID(hotelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}
	return toPartnerSettings(hotel), nil
}

func (s *PartnerService) UpdateSettings(userID string, req dto.PartnerUpdateSettingsRequest) (*dto.PartnerSettingsResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	update := dto.UpdateHotelRequest{
		Description:         req.Description,
		Phone:               req.Phone,
		Email:               req.Email,
		GoogleMapsURL:       req.GoogleMapsURL,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		SelcomPayoutAccount: req.SelcomPayoutAccount,
	}
	resp, err := s.admin.UpdateHotel(hotelID.String(), update)
	if err != nil {
		return nil, err
	}
	return &dto.PartnerSettingsResponse{
		ID:                  resp.ID,
		Name:                resp.Name,
		Description:         resp.Description,
		Address:             resp.Address,
		City:                resp.City,
		Location:            resp.Location,
		Phone:               resp.Phone,
		Email:               resp.Email,
		GoogleMapsURL:       resp.GoogleMapsURL,
		Latitude:            resp.Latitude,
		Longitude:           resp.Longitude,
		SelcomPayoutAccount: resp.SelcomPayoutAccount,
	}, nil
}

func (s *PartnerService) isAffiliatePartner(partnerType string) bool {
	switch partnerType {
	case models.PartnerTypeTourOperator, models.PartnerTypeTravelAgent, models.PartnerTypeAirline:
		return true
	default:
		return false
	}
}

func (s *PartnerService) ListReferrals(userID string) ([]dto.PartnerReferralResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	hotel, err := s.hotels.FindByID(hotelID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load partner", apperrors.ErrInternal.Status)
	}
	if !s.isAffiliatePartner(hotel.PartnerType) {
		return []dto.PartnerReferralResponse{}, nil
	}
	referrals, err := s.bookings.ListByPartner(hotelID, 50)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to list referrals", apperrors.ErrInternal.Status)
	}
	out := make([]dto.PartnerReferralResponse, 0, len(referrals))
	for _, r := range referrals {
		out = append(out, toPartnerReferral(r))
	}
	return out, nil
}

func (s *PartnerService) CreateReferral(userID string, req dto.CreatePartnerReferralRequest) (*dto.PartnerReferralResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	hotel, err := s.hotels.FindByID(hotelID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load partner", apperrors.ErrInternal.Status)
	}
	if !s.isAffiliatePartner(hotel.PartnerType) {
		return nil, apperrors.New(apperrors.ErrForbidden.Code, "only tour operators, travel agents and airlines can create referrals", apperrors.ErrForbidden.Status)
	}
	m := &models.BookingReferral{
		ReferringPartnerID: hotel.ID,
		TravellerName:      strings.TrimSpace(req.TravellerName),
		TravellerPhone:     strings.TrimSpace(req.TravellerPhone),
		TravellerEmail:     strings.TrimSpace(req.TravellerEmail),
		Destination:        strings.TrimSpace(req.Destination),
		TripContext:        strings.TrimSpace(req.TripContext),
		ReferralToken:      strings.ReplaceAll(uuid.New().String(), "-", "")[:16],
		Status:             models.BookingReferralStatusPending,
	}
	if req.TripStartDate != "" {
		if t, err := time.Parse("2006-01-02", req.TripStartDate); err == nil {
			m.TripStartDate = &t
		}
	}
	if req.TripEndDate != "" {
		if t, err := time.Parse("2006-01-02", req.TripEndDate); err == nil {
			m.TripEndDate = &t
		}
	}
	if err := s.bookings.Create(m); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create referral", apperrors.ErrInternal.Status)
	}
	resp := toPartnerReferral(*m)
	return &resp, nil
}

func toPartnerReferral(r models.BookingReferral) dto.PartnerReferralResponse {
	out := dto.PartnerReferralResponse{
		ID:             r.ID.String(),
		TravellerName:  r.TravellerName,
		TravellerPhone: r.TravellerPhone,
		TravellerEmail: r.TravellerEmail,
		Destination:    r.Destination,
		TripContext:    r.TripContext,
		ReferralToken:  r.ReferralToken,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339),
	}
	if r.TripStartDate != nil {
		out.TripStartDate = r.TripStartDate.Format("2006-01-02")
	}
	if r.TripEndDate != nil {
		out.TripEndDate = r.TripEndDate.Format("2006-01-02")
	}
	return out
}

func toPartnerSettings(hotel *models.Hotel) *dto.PartnerSettingsResponse {
	return &dto.PartnerSettingsResponse{
		ID:                  hotel.ID.String(),
		Name:                hotel.Name,
		Description:         hotel.Description,
		Address:             hotel.Address,
		City:                hotel.City,
		Location:            hotel.Location,
		Phone:               hotel.Phone,
		Email:               hotel.Email,
		GoogleMapsURL:       hotel.GoogleMapsURL,
		Latitude:            hotel.Latitude,
		Longitude:           hotel.Longitude,
		SelcomPayoutAccount: hotel.SelcomPayoutAccount,
	}
}
