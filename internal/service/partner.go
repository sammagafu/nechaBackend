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

// PartnerService exposes property-scoped operations for partner-role users (brief §11).
// Partners see only their own hotel's dashboard, orders, and products — never platform-wide data.
type PartnerService struct {
	users *repository.UserRepository
	admin *AdminService
}

func NewPartnerService(users *repository.UserRepository, admin *AdminService) *PartnerService {
	return &PartnerService{users: users, admin: admin}
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
	orders, err := s.admin.ListOrders(200, 0)
	if err != nil {
		return nil, err
	}
	filtered := make([]dto.AdminOrderResponse, 0)
	hid := hotelID.String()
	for _, o := range orders {
		if o.HotelID == hid {
			filtered = append(filtered, o)
		}
	}
	return filtered, nil
}

func (s *PartnerService) ListProducts(userID string) ([]dto.AdminProductResponse, error) {
	hotelID, err := s.hotelIDForUser(userID)
	if err != nil {
		return nil, err
	}
	return s.admin.ListProducts(hotelID.String())
}
