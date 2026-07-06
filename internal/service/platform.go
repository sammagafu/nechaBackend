package service

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/dto"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

type PlatformService struct {
	zones  *repository.DeliveryZoneRepository
	config *repository.PlatformConfigRepository
}

func NewPlatformService(zones *repository.DeliveryZoneRepository, config *repository.PlatformConfigRepository) *PlatformService {
	return &PlatformService{zones: zones, config: config}
}

func (s *PlatformService) Settings() (*dto.PlatformSettingsResponse, error) {
	rate := 2625.0
	if raw, err := s.config.Get(models.ConfigKeyTzsToUsdRate); err == nil && raw != "" {
		if parsed, parseErr := strconv.ParseFloat(raw, 64); parseErr == nil && parsed > 0 {
			rate = parsed
		}
	}

	zones, err := s.zones.ListActive()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load delivery zones", apperrors.ErrInternal.Status)
	}

	out := &dto.PlatformSettingsResponse{
		TzsToUsdRate:             rate,
		FreeDeliveryThresholdTZS: 180000,
		DefaultDeliveryFeeTZS:    5000,
		DeliveryBaseFeeTZS:       s.configInt64(models.ConfigKeyDeliveryBaseFeeTZS, 3000),
		DeliveryPerKmTZS:         s.configInt64(models.ConfigKeyDeliveryPerKmTZS, 1000),
		Features:                 s.loadFeatures(),
		Zones:                    make([]dto.DeliveryZoneResponse, 0, len(zones)),
	}
	for _, z := range zones {
		out.Zones = append(out.Zones, dto.DeliveryZoneResponse{
			Code:             z.Code,
			Label:            z.Label,
			DeliveryFeeTZS:   z.DeliveryFeeTZS,
			FreeThresholdTZS: z.FreeThresholdTZS,
		})
		if out.FreeDeliveryThresholdTZS == 180000 && z.FreeThresholdTZS > 0 {
			out.FreeDeliveryThresholdTZS = z.FreeThresholdTZS
		}
		if out.DefaultDeliveryFeeTZS == 5000 && z.DeliveryFeeTZS > 0 {
			out.DefaultDeliveryFeeTZS = z.DeliveryFeeTZS
		}
	}
	return out, nil
}

// ComputeDeliveryFee returns the delivery fee for a subtotal in the hotel's zone.
func (s *PlatformService) ComputeDeliveryFee(zoneCode string, subtotal int64) (int64, error) {
	code := strings.TrimSpace(zoneCode)
	if code == "" {
		code = "A"
	}
	zone, err := s.zones.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			settings, settingsErr := s.Settings()
			if settingsErr != nil {
				return 0, settingsErr
			}
			if subtotal >= settings.FreeDeliveryThresholdTZS {
				return 0, nil
			}
			return settings.DefaultDeliveryFeeTZS, nil
		}
		return 0, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load delivery zone", apperrors.ErrInternal.Status)
	}
	if subtotal >= zone.FreeThresholdTZS {
		return 0, nil
	}
	return zone.DeliveryFeeTZS, nil
}

// ComputeDeliveryFeeByDistance derives a delivery fee from the straight-line (haversine)
// distance between the hotel and the guest's delivery coordinates (brief §4.5 — the stored
// Google Maps location is used for distance-based pricing). Falls back to the flat zone fee
// when coordinates are unavailable. Config keys drive the base fee and per-km rate so the
// pricing can be tuned without a code change.
func (s *PlatformService) ComputeDeliveryFeeByDistance(hotel *models.Hotel, destLat, destLng *float64, subtotal int64) (int64, error) {
	if !s.FeatureEnabled(models.ConfigKeyFeatureDistanceDeliveryEnabled, true) {
		zone := ""
		if hotel != nil {
			zone = hotel.Zone
		}
		return s.ComputeDeliveryFee(zone, subtotal)
	}
	if hotel == nil || hotel.Latitude == nil || hotel.Longitude == nil || destLat == nil || destLng == nil {
		zone := ""
		if hotel != nil {
			zone = hotel.Zone
		}
		return s.ComputeDeliveryFee(zone, subtotal)
	}
	settings, err := s.Settings()
	if err != nil {
		return 0, err
	}
	if subtotal >= settings.FreeDeliveryThresholdTZS {
		return 0, nil
	}

	baseFee := int64(3000)
	if raw, err := s.config.Get(models.ConfigKeyDeliveryBaseFeeTZS); err == nil && raw != "" {
		if parsed, perr := strconv.ParseInt(raw, 10, 64); perr == nil && parsed >= 0 {
			baseFee = parsed
		}
	}
	perKm := int64(1000)
	if raw, err := s.config.Get(models.ConfigKeyDeliveryPerKmTZS); err == nil && raw != "" {
		if parsed, perr := strconv.ParseInt(raw, 10, 64); perr == nil && parsed >= 0 {
			perKm = parsed
		}
	}

	km := haversineKm(*hotel.Latitude, *hotel.Longitude, *destLat, *destLng)
	fee := baseFee + int64(math.Ceil(km))*perKm
	return fee, nil
}

// haversineKm returns the great-circle distance between two lat/lng points in kilometres.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func (s *PlatformService) FeatureEnabled(key string, defaultVal bool) bool {
	raw, err := s.config.Get(key)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaultVal
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultVal
	}
}

func (s *PlatformService) setFeature(key string, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	return s.config.Set(key, val)
}

func (s *PlatformService) configInt64(key string, defaultVal int64) int64 {
	raw, err := s.config.Get(key)
	if err != nil || raw == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func (s *PlatformService) loadFeatures() dto.PlatformFeatures {
	return dto.PlatformFeatures{
		RewardsEnabled:               s.FeatureEnabled(models.ConfigKeyFeatureRewardsEnabled, true),
		RewardsRedeemEnabled:         s.FeatureEnabled(models.ConfigKeyFeatureRewardsRedeemEnabled, false),
		DiscoveryTicketingEnabled:    s.FeatureEnabled(models.ConfigKeyFeatureDiscoveryTicketingEnabled, false),
		PartnerPortalEnabled:         s.FeatureEnabled(models.ConfigKeyFeaturePartnerPortalEnabled, true),
		PartnerProductsManageEnabled: s.FeatureEnabled(models.ConfigKeyFeaturePartnerProductsManageEnabled, false),
		DualCurrencyEnabled:          s.FeatureEnabled(models.ConfigKeyFeatureDualCurrencyEnabled, true),
		DistanceDeliveryEnabled:      s.FeatureEnabled(models.ConfigKeyFeatureDistanceDeliveryEnabled, true),
	}
}

func (s *PlatformService) UpdateSettings(req dto.UpdatePlatformSettingsRequest) (*dto.PlatformSettingsResponse, error) {
	if req.TzsToUsdRate != nil && *req.TzsToUsdRate > 0 {
		if err := s.config.Set(models.ConfigKeyTzsToUsdRate, strconv.FormatFloat(*req.TzsToUsdRate, 'f', -1, 64)); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save rate", apperrors.ErrInternal.Status)
		}
	}
	if req.DeliveryBaseFeeTZS != nil && *req.DeliveryBaseFeeTZS >= 0 {
		if err := s.config.Set(models.ConfigKeyDeliveryBaseFeeTZS, strconv.FormatInt(*req.DeliveryBaseFeeTZS, 10)); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save delivery base fee", apperrors.ErrInternal.Status)
		}
	}
	if req.DeliveryPerKmTZS != nil && *req.DeliveryPerKmTZS >= 0 {
		if err := s.config.Set(models.ConfigKeyDeliveryPerKmTZS, strconv.FormatInt(*req.DeliveryPerKmTZS, 10)); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save delivery per-km fee", apperrors.ErrInternal.Status)
		}
	}
	if req.Features != nil {
		f := *req.Features
		flags := map[string]bool{
			models.ConfigKeyFeatureRewardsEnabled:               f.RewardsEnabled,
			models.ConfigKeyFeatureRewardsRedeemEnabled:         f.RewardsRedeemEnabled,
			models.ConfigKeyFeatureDiscoveryTicketingEnabled:    f.DiscoveryTicketingEnabled,
			models.ConfigKeyFeaturePartnerPortalEnabled:         f.PartnerPortalEnabled,
			models.ConfigKeyFeaturePartnerProductsManageEnabled: f.PartnerProductsManageEnabled,
			models.ConfigKeyFeatureDualCurrencyEnabled:          f.DualCurrencyEnabled,
			models.ConfigKeyFeatureDistanceDeliveryEnabled:      f.DistanceDeliveryEnabled,
		}
		for key, enabled := range flags {
			if err := s.setFeature(key, enabled); err != nil {
				return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save feature flag", apperrors.ErrInternal.Status)
			}
		}
	}
	return s.Settings()
}

type GuestRequestService struct {
	requests *repository.GuestRequestRepository
	hotels   *repository.HotelRepository
	events   *EventService
}

func NewGuestRequestService(
	requests *repository.GuestRequestRepository,
	hotels *repository.HotelRepository,
	events *EventService,
) *GuestRequestService {
	return &GuestRequestService{requests: requests, hotels: hotels, events: events}
}

func (s *GuestRequestService) Submit(slug string, req dto.SubmitGuestRequestRequest) (*dto.GuestRequestResponse, error) {
	hotel, err := s.hotels.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, apperrors.ErrNotFound.Code, "hotel not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load hotel", apperrors.ErrInternal.Status)
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = models.GuestRequestCategoryGeneral
	}

	item := &models.GuestRequest{
		HotelID:    hotel.ID,
		Category:   category,
		Status:     models.GuestRequestStatusNew,
		GuestName:  strings.TrimSpace(req.GuestName),
		GuestPhone: strings.TrimSpace(req.GuestPhone),
		GuestEmail: strings.TrimSpace(req.GuestEmail),
		RoomNumber: strings.TrimSpace(req.RoomNumber),
		Body:       strings.TrimSpace(req.Body),
	}
	if item.Body == "" {
		return nil, apperrors.New(apperrors.ErrBadRequest.Code, "request body is required", apperrors.ErrBadRequest.Status)
	}
	if err := s.requests.Create(item); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to save guest request", apperrors.ErrInternal.Status)
	}
	if s.events != nil {
		s.events.GuestRequestCreated(item, hotel.Name)
	}
	resp := toGuestRequestResponse(item)
	return &resp, nil
}

func toGuestRequestResponse(item *models.GuestRequest) dto.GuestRequestResponse {
	return dto.GuestRequestResponse{
		ID:         item.ID.String(),
		HotelID:    item.HotelID.String(),
		Category:   item.Category,
		Status:     item.Status,
		GuestName:  item.GuestName,
		GuestPhone: item.GuestPhone,
		GuestEmail: item.GuestEmail,
		RoomNumber: item.RoomNumber,
		Body:       item.Body,
		CreatedAt:  item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
