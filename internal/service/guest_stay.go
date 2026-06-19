package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	"gorm.io/gorm"
)

type GuestStayService struct {
	stays  *repository.GuestStayRepository
	hotels *repository.HotelRepository
}

func NewGuestStayService(stays *repository.GuestStayRepository, hotels *repository.HotelRepository) *GuestStayService {
	return &GuestStayService{stays: stays, hotels: hotels}
}

type guestStayRecord struct {
	UserID        *uuid.UUID
	HotelID       uuid.UUID
	Channel       string
	RoomNumber    string
	ReferralCode  string
	Source        models.GuestStaySource
	OrderID       *uuid.UUID
	ReservationID *uuid.UUID
	ScannedAt     *time.Time
}

func (s *GuestStayService) RecordScan(slug, ref, channel, scannedAtRaw string, userID *uuid.UUID) error {
	hotel, err := s.resolveHotel(ref, slug)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	scannedAt := parseScannedAt(scannedAtRaw)
	if scannedAt == nil {
		scannedAt = &now
	}
	return s.record(guestStayRecord{
		UserID:       userID,
		HotelID:      hotel.ID,
		Channel:      normalizeChannel(channel),
		ReferralCode: firstNonEmptyStr(ref, hotel.ReferralCode, hotel.Code),
		Source:       models.GuestStaySourceScan,
		ScannedAt:    scannedAt,
	})
}

func (s *GuestStayService) RecordFromContext(input *dto.HotelContextInput, userID uuid.UUID, source models.GuestStaySource) {
	if input == nil {
		return
	}
	hotel, err := s.resolveHotel(input.HotelCode, input.HotelSlug)
	if err != nil {
		return
	}
	_ = s.record(guestStayRecord{
		UserID:       &userID,
		HotelID:      hotel.ID,
		Channel:      normalizeChannel(input.Channel),
		RoomNumber:   strings.TrimSpace(input.RoomNumber),
		ReferralCode: firstNonEmptyStr(input.ReferralCode, hotel.ReferralCode, hotel.Code),
		Source:       source,
		ScannedAt:    parseScannedAt(input.ScannedAt),
	})
}

func (s *GuestStayService) RecordFromOrder(order *models.Order, referralCode string, source models.GuestStaySource) {
	if order == nil {
		return
	}
	scannedAt := order.CreatedAt
	_ = s.record(guestStayRecord{
		UserID:       order.UserID,
		HotelID:      order.HotelID,
		Channel:      models.GuestStayChannelRoom,
		RoomNumber:   strings.TrimSpace(order.RoomNumber),
		ReferralCode: strings.TrimSpace(referralCode),
		Source:       source,
		OrderID:      &order.ID,
		ScannedAt:    &scannedAt,
	})
}

func (s *GuestStayService) RecordFromReservation(reservation *models.Reservation, roomNumber, referralCode string, source models.GuestStaySource) {
	if reservation == nil {
		return
	}
	scannedAt := reservation.CreatedAt
	_ = s.record(guestStayRecord{
		UserID:        reservation.UserID,
		HotelID:       reservation.HotelID,
		Channel:       models.GuestStayChannelRoom,
		RoomNumber:    strings.TrimSpace(roomNumber),
		ReferralCode:  strings.TrimSpace(referralCode),
		Source:        source,
		ReservationID: &reservation.ID,
		ScannedAt:     &scannedAt,
	})
}

func (s *GuestStayService) List(limit, offset int) ([]dto.AdminGuestStayResponse, error) {
	stays, err := s.stays.List(limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AdminGuestStayResponse, 0, len(stays))
	for i := range stays {
		result = append(result, toAdminGuestStayResponse(&stays[i]))
	}
	return result, nil
}

func (s *GuestStayService) record(input guestStayRecord) error {
	stay := &models.GuestStay{
		UserID:        input.UserID,
		HotelID:       input.HotelID,
		Channel:       input.Channel,
		RoomNumber:    input.RoomNumber,
		ReferralCode:  input.ReferralCode,
		Source:        input.Source,
		OrderID:       input.OrderID,
		ReservationID: input.ReservationID,
		ScannedAt:     input.ScannedAt,
	}
	return s.stays.Create(stay)
}

func (s *GuestStayService) resolveHotel(code, slug string) (*models.Hotel, error) {
	code = strings.TrimSpace(code)
	slug = strings.TrimSpace(slug)
	if code != "" {
		hotel, err := s.hotels.FindByCode(code)
		if err == nil {
			return hotel, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if slug != "" {
		return s.hotels.FindBySlug(slug)
	}
	return nil, gorm.ErrRecordNotFound
}

func parseScannedAt(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return nil
		}
	}
	return &t
}

func firstNonEmptyStr(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizeChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case models.GuestStayChannelPoster, models.GuestStayChannelLobby:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return models.GuestStayChannelRoom
	}
}

func toAdminGuestStayResponse(stay *models.GuestStay) dto.AdminGuestStayResponse {
	resp := dto.AdminGuestStayResponse{
		ID:           stay.ID.String(),
		HotelID:      stay.HotelID.String(),
		HotelName:    stay.Hotel.Name,
		Channel:      stay.Channel,
		RoomNumber:   stay.RoomNumber,
		ReferralCode: stay.ReferralCode,
		Source:       string(stay.Source),
		CreatedAt:    stay.CreatedAt.Format(time.RFC3339),
	}
	if stay.UserID != nil {
		id := stay.UserID.String()
		resp.UserID = &id
		if stay.User != nil {
			resp.UserEmail = stay.User.Email
			resp.UserName = stay.User.FullName
		}
	}
	if stay.OrderID != nil {
		id := stay.OrderID.String()
		resp.OrderID = &id
	}
	if stay.ReservationID != nil {
		id := stay.ReservationID.String()
		resp.ReservationID = &id
	}
	if stay.ScannedAt != nil {
		resp.ScannedAt = stay.ScannedAt.Format(time.RFC3339)
	}
	if stay.Order != nil {
		total := stay.Order.TotalAmount
		resp.TotalAmount = &total
		resp.Currency = stay.Order.Currency
		resp.ItemsSummary = summarizeOrderItems(stay.Order.Items)
	}
	return resp
}

func summarizeOrderItems(items []models.OrderItem) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, formatItemLine(item.Quantity, item.Name))
	}
	return strings.Join(parts, ", ")
}

func formatItemLine(qty int, name string) string {
	if qty <= 1 {
		return name
	}
	return strconv.Itoa(qty) + "× " + name
}
