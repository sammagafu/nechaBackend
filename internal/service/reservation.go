package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/integration/kkooapp"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type ReservationService struct {
	hotels       *repository.HotelRepository
	reservations *repository.ReservationRepository
	kkooapp      kkooapp.Client
	events       *EventService
	guestStays   *GuestStayService
}

func NewReservationService(
	hotels *repository.HotelRepository,
	reservations *repository.ReservationRepository,
	kkooappClient kkooapp.Client,
	events *EventService,
	guestStays *GuestStayService,
) *ReservationService {
	return &ReservationService{
		hotels:       hotels,
		reservations: reservations,
		kkooapp:      kkooappClient,
		events:       events,
		guestStays:   guestStays,
	}
}

func (s *ReservationService) CreateHotelReservation(ctx context.Context, req dto.HotelReservationRequest, userID *uuid.UUID) (*dto.ReservationResponse, error) {
	hotel, err := s.hotels.FindByCode(req.HotelCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	if hotel.KkooappID == "" {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "hotel is not linked to kkooapp", apperrors.ErrBadRequest.Status)
	}

	checkIn, err := parseDate(req.CheckIn)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "invalid check_in date (use YYYY-MM-DD)", apperrors.ErrValidation.Status)
	}
	checkOut, err := parseDate(req.CheckOut)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "invalid check_out date (use YYYY-MM-DD)", apperrors.ErrValidation.Status)
	}
	if !checkOut.After(checkIn) {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "check_out must be after check_in", apperrors.ErrValidation.Status)
	}

	reservation := &models.Reservation{
		HotelID:         hotel.ID,
		UserID:          userID,
		Type:            models.ReservationTypeHotel,
		Status:          models.ReservationStatusPending,
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		CheckIn:         &checkIn,
		CheckOut:        &checkOut,
		RoomType:        req.RoomType,
		GuestCount:      req.GuestCount,
		SpecialRequests: req.SpecialRequests,
	}

	if err := s.reservations.Create(reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create reservation", apperrors.ErrInternal.Status)
	}

	kkResp, err := s.kkooapp.ReserveHotelRoom(ctx, kkooapp.HotelReservationRequest{
		PropertyID:      hotel.KkooappID,
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		CheckIn:         req.CheckIn,
		CheckOut:        req.CheckOut,
		RoomType:        req.RoomType,
		GuestCount:      req.GuestCount,
		SpecialRequests: req.SpecialRequests,
		ExternalRef:     reservation.ID.String(),
	})
	if err != nil {
		reservation.Status = models.ReservationStatusFailed
		reservation.Notes = err.Error()
		_ = s.reservations.Update(reservation)
		return nil, err
	}

	reservation.KkooappRef = kkResp.Reference
	reservation.Status = mapKkooappReservationStatus(kkResp.Status)
	if err := s.reservations.Update(reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update reservation", apperrors.ErrInternal.Status)
	}

	if s.events != nil {
		s.events.ReservationCreated(reservation, hotel.Name)
	}
	if s.guestStays != nil {
		s.guestStays.RecordFromReservation(reservation, req.RoomType, hotel.ReferralCode, models.GuestStaySourceHotelReservation)
	}
	return toReservationResponse(reservation), nil
}

func (s *ReservationService) CreateTableReservation(ctx context.Context, req dto.TableReservationRequest, userID *uuid.UUID) (*dto.ReservationResponse, error) {
	hotel, err := s.hotels.FindByCode(req.HotelCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	if hotel.KkooappID == "" {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "hotel is not linked to kkooapp", apperrors.ErrBadRequest.Status)
	}

	reservationDate, err := parseDateTime(req.ReservationDate)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrValidation.Code, "invalid reservation_date (use RFC3339)", apperrors.ErrValidation.Status)
	}

	reservation := &models.Reservation{
		HotelID:         hotel.ID,
		UserID:          userID,
		Type:            models.ReservationTypeTable,
		Status:          models.ReservationStatusPending,
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		ReservationDate: &reservationDate,
		PartySize:       req.PartySize,
		TableNumber:     req.TableNumber,
		SpecialRequests: req.SpecialRequests,
	}

	if err := s.reservations.Create(reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to create reservation", apperrors.ErrInternal.Status)
	}

	kkResp, err := s.kkooapp.ReserveTable(ctx, kkooapp.TableReservationRequest{
		PropertyID:      hotel.KkooappID,
		GuestName:       req.GuestName,
		GuestEmail:      req.GuestEmail,
		GuestPhone:      req.GuestPhone,
		ReservationDate: req.ReservationDate,
		PartySize:       req.PartySize,
		TableNumber:     req.TableNumber,
		SpecialRequests: req.SpecialRequests,
		ExternalRef:     reservation.ID.String(),
	})
	if err != nil {
		reservation.Status = models.ReservationStatusFailed
		reservation.Notes = err.Error()
		_ = s.reservations.Update(reservation)
		return nil, err
	}

	reservation.KkooappRef = kkResp.Reference
	reservation.Status = mapKkooappReservationStatus(kkResp.Status)
	if err := s.reservations.Update(reservation); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to update reservation", apperrors.ErrInternal.Status)
	}

	if s.events != nil {
		s.events.ReservationCreated(reservation, hotel.Name)
	}
	if s.guestStays != nil {
		s.guestStays.RecordFromReservation(reservation, req.TableNumber, hotel.ReferralCode, models.GuestStaySourceTableReservation)
	}
	return toReservationResponse(reservation), nil
}

func (s *ReservationService) GetByID(id uuid.UUID, viewerID *uuid.UUID, isAdmin bool) (*dto.ReservationResponse, error) {
	reservation, err := s.reservations.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "reservation not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load reservation", apperrors.ErrInternal.Status)
	}
	resp := toReservationResponse(reservation)
	if !canViewReservationPII(reservation, viewerID, isAdmin) {
		resp.GuestEmail = ""
		resp.GuestPhone = ""
	}
	return resp, nil
}

func canViewReservationPII(reservation *models.Reservation, viewerID *uuid.UUID, isAdmin bool) bool {
	if isAdmin {
		return true
	}
	if viewerID == nil || reservation.UserID == nil {
		return false
	}
	return *reservation.UserID == *viewerID
}

func toReservationResponse(r *models.Reservation) *dto.ReservationResponse {
	resp := &dto.ReservationResponse{
		ID:              r.ID.String(),
		HotelID:         r.HotelID.String(),
		Type:            string(r.Type),
		Status:          string(r.Status),
		KkooappRef:      r.KkooappRef,
		GuestName:       r.GuestName,
		GuestEmail:      r.GuestEmail,
		GuestPhone:      r.GuestPhone,
		RoomType:        r.RoomType,
		GuestCount:      r.GuestCount,
		PartySize:       r.PartySize,
		SpecialRequests: r.SpecialRequests,
		CreatedAt:       r.CreatedAt.UTC().Format(time.RFC3339),
	}
	if r.CheckIn != nil {
		resp.CheckIn = r.CheckIn.Format("2006-01-02")
	}
	if r.CheckOut != nil {
		resp.CheckOut = r.CheckOut.Format("2006-01-02")
	}
	if r.ReservationDate != nil {
		resp.ReservationDate = r.ReservationDate.UTC().Format(time.RFC3339)
	}
	return resp
}

func mapKkooappReservationStatus(status string) models.ReservationStatus {
	switch status {
	case "confirmed":
		return models.ReservationStatusConfirmed
	case "cancelled":
		return models.ReservationStatusCancelled
	default:
		return models.ReservationStatusPending
	}
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseDateTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
