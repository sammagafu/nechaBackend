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

// CommerceService owns CRUD for the partner-economy entities: influencers, booking
// referrals, suppliers, the central catalogue + property visibility, commission rules,
// commission records, and the rewards ledger (brief §3, §8, §17).
type CommerceService struct {
	influencers *repository.InfluencerRepository
	bookings    *repository.BookingReferralRepository
	suppliers   *repository.SupplierRepository
	catalogue   *repository.CatalogueRepository
	commissions *repository.CommissionRepository
	rewards     *repository.RewardRepository
	hotels      *repository.HotelRepository
	events      *repository.EventLogRepository
}

func NewCommerceService(
	influencers *repository.InfluencerRepository,
	bookings *repository.BookingReferralRepository,
	suppliers *repository.SupplierRepository,
	catalogue *repository.CatalogueRepository,
	commissions *repository.CommissionRepository,
	rewards *repository.RewardRepository,
	hotels *repository.HotelRepository,
	events *repository.EventLogRepository,
) *CommerceService {
	return &CommerceService{
		influencers: influencers,
		bookings:    bookings,
		suppliers:   suppliers,
		catalogue:   catalogue,
		commissions: commissions,
		rewards:     rewards,
		hotels:      hotels,
		events:      events,
	}
}

func notFound(msg string) error {
	return apperrors.New(apperrors.ErrNotFound.Code, msg, apperrors.ErrNotFound.Status)
}

func internal(err error, msg string) error {
	return apperrors.Wrap(err, apperrors.ErrInternal.Code, msg, apperrors.ErrInternal.Status)
}

func badRequest(msg string) error {
	return apperrors.New(apperrors.ErrBadRequest.Code, msg, apperrors.ErrBadRequest.Status)
}

// Influencers ----------------------------------------------------------------

func (s *CommerceService) ListInfluencers() ([]models.Influencer, error) {
	out, err := s.influencers.List()
	if err != nil {
		return nil, internal(err, "failed to list influencers")
	}
	return out, nil
}

func (s *CommerceService) CreateInfluencer(req dto.CreateInfluencerRequest) (*models.Influencer, error) {
	pct := req.CommissionPct
	if pct <= 0 {
		pct = 0.30
	}
	status := req.Status
	if status == "" {
		status = models.InfluencerStatusActive
	}
	m := &models.Influencer{
		Name:                strings.TrimSpace(req.Name),
		Email:               strings.TrimSpace(req.Email),
		Phone:               strings.TrimSpace(req.Phone),
		ReferralCode:        strings.TrimSpace(req.ReferralCode),
		CommissionPct:       pct,
		SelcomPayoutAccount: req.SelcomPayoutAccount,
		AgreementRef:        req.AgreementRef,
		Status:              status,
	}
	if err := s.influencers.Create(m); err != nil {
		return nil, internal(err, "failed to create influencer")
	}
	return m, nil
}

func (s *CommerceService) UpdateInfluencer(id string, req dto.UpdateInfluencerRequest) (*models.Influencer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, badRequest("invalid influencer id")
	}
	m, err := s.influencers.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("influencer not found")
		}
		return nil, internal(err, "failed to load influencer")
	}
	if req.Name != nil {
		m.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		m.Email = strings.TrimSpace(*req.Email)
	}
	if req.Phone != nil {
		m.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.ReferralCode != nil {
		m.ReferralCode = strings.TrimSpace(*req.ReferralCode)
	}
	if req.CommissionPct != nil {
		m.CommissionPct = *req.CommissionPct
	}
	if req.SelcomPayoutAccount != nil {
		m.SelcomPayoutAccount = *req.SelcomPayoutAccount
	}
	if req.AgreementRef != nil {
		m.AgreementRef = *req.AgreementRef
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.influencers.Update(m); err != nil {
		return nil, internal(err, "failed to update influencer")
	}
	return m, nil
}

// Booking referrals ----------------------------------------------------------

func (s *CommerceService) ListBookingReferrals() ([]models.BookingReferral, error) {
	out, err := s.bookings.List()
	if err != nil {
		return nil, internal(err, "failed to list booking referrals")
	}
	return out, nil
}

// CreateBookingReferral is called from the partner form. The referring partner is resolved
// by code and must be a non-hotel partner type.
func (s *CommerceService) CreateBookingReferral(req dto.CreateBookingReferralRequest) (*models.BookingReferral, error) {
	partner, err := s.hotels.FindByCode(strings.TrimSpace(req.ReferringPartnerCode))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("referring partner not found")
		}
		return nil, internal(err, "failed to load partner")
	}
	switch partner.PartnerType {
	case models.PartnerTypeTourOperator, models.PartnerTypeTravelAgent, models.PartnerTypeAirline:
		// ok
	default:
		return nil, badRequest("referring partner must be a tour operator, travel agent, or airline")
	}
	m := &models.BookingReferral{
		ReferringPartnerID: partner.ID,
		TravellerName:      strings.TrimSpace(req.TravellerName),
		TravellerPhone:     strings.TrimSpace(req.TravellerPhone),
		TravellerEmail:     strings.TrimSpace(req.TravellerEmail),
		TripContext:        strings.TrimSpace(req.TripContext),
		ReferralToken:      strings.ReplaceAll(uuid.New().String(), "-", "")[:16],
		Status:             models.BookingReferralStatusPending,
	}
	m.TripStartDate = parseOptionalDate(req.TripStartDate)
	m.TripEndDate = parseOptionalDate(req.TripEndDate)
	if err := s.bookings.Create(m); err != nil {
		return nil, internal(err, "failed to create booking referral")
	}
	return m, nil
}

// Suppliers ------------------------------------------------------------------

func (s *CommerceService) ListSuppliers() ([]models.Supplier, error) {
	out, err := s.suppliers.List()
	if err != nil {
		return nil, internal(err, "failed to list suppliers")
	}
	return out, nil
}

func (s *CommerceService) CreateSupplier(req dto.CreateSupplierRequest) (*models.Supplier, error) {
	m := &models.Supplier{
		SupplierType:        strings.TrimSpace(req.SupplierType),
		Name:                strings.TrimSpace(req.Name),
		ContactPhone:        req.ContactPhone,
		ContactEmail:        req.ContactEmail,
		LicenseReference:    req.LicenseReference,
		SelcomPayoutAccount: req.SelcomPayoutAccount,
		AgreementNotes:      req.AgreementNotes,
		IsActive:            true,
	}
	if err := s.suppliers.Create(m); err != nil {
		return nil, internal(err, "failed to create supplier")
	}
	return m, nil
}

func (s *CommerceService) UpdateSupplier(id string, req dto.UpdateSupplierRequest) (*models.Supplier, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, badRequest("invalid supplier id")
	}
	m, err := s.suppliers.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("supplier not found")
		}
		return nil, internal(err, "failed to load supplier")
	}
	if req.Name != nil {
		m.Name = strings.TrimSpace(*req.Name)
	}
	if req.ContactPhone != nil {
		m.ContactPhone = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		m.ContactEmail = *req.ContactEmail
	}
	if req.LicenseReference != nil {
		m.LicenseReference = *req.LicenseReference
	}
	if req.SelcomPayoutAccount != nil {
		m.SelcomPayoutAccount = *req.SelcomPayoutAccount
	}
	if req.AgreementNotes != nil {
		m.AgreementNotes = *req.AgreementNotes
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}
	if err := s.suppliers.Update(m); err != nil {
		return nil, internal(err, "failed to update supplier")
	}
	return m, nil
}

// Catalogue ------------------------------------------------------------------

func (s *CommerceService) ListCatalogue() ([]models.CatalogueItem, error) {
	out, err := s.catalogue.List()
	if err != nil {
		return nil, internal(err, "failed to list catalogue")
	}
	return out, nil
}

func (s *CommerceService) CreateCatalogueItem(req dto.CreateCatalogueItemRequest) (*models.CatalogueItem, error) {
	supplierID, err := uuid.Parse(req.SupplierID)
	if err != nil {
		return nil, badRequest("invalid supplier id")
	}
	if _, err := s.suppliers.FindByID(supplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("supplier not found")
		}
		return nil, internal(err, "failed to load supplier")
	}
	currency := req.Currency
	if currency == "" {
		currency = "TZS"
	}
	m := &models.CatalogueItem{
		SupplierID:      supplierID,
		Category:        strings.TrimSpace(req.Category),
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		Currency:        currency,
		BasePrice:       req.BasePrice,
		CapacityOrStock: req.CapacityOrStock,
		IsDigital:       req.IsDigital,
		EventDate:       parseOptionalDate(req.EventDate),
		IsActive:        true,
	}
	if err := s.catalogue.Create(m); err != nil {
		return nil, internal(err, "failed to create catalogue item")
	}
	return m, nil
}

func (s *CommerceService) UpdateCatalogueItem(id string, req dto.UpdateCatalogueItemRequest) (*models.CatalogueItem, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, badRequest("invalid catalogue item id")
	}
	m, err := s.catalogue.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound("catalogue item not found")
		}
		return nil, internal(err, "failed to load catalogue item")
	}
	if req.Name != nil {
		m.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.Currency != nil {
		m.Currency = *req.Currency
	}
	if req.BasePrice != nil {
		m.BasePrice = *req.BasePrice
	}
	if req.CapacityOrStock != nil {
		m.CapacityOrStock = *req.CapacityOrStock
	}
	if req.IsDigital != nil {
		m.IsDigital = *req.IsDigital
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}
	if err := s.catalogue.Update(m); err != nil {
		return nil, internal(err, "failed to update catalogue item")
	}
	return m, nil
}

// SetCatalogueVisibility exposes/hides a catalogue item for a specific property.
func (s *CommerceService) SetCatalogueVisibility(itemID string, req dto.SetVisibilityRequest) error {
	catItemID, err := uuid.Parse(itemID)
	if err != nil {
		return badRequest("invalid catalogue item id")
	}
	hotelID, err := uuid.Parse(req.HotelID)
	if err != nil {
		return badRequest("invalid hotel id")
	}
	pv := &models.PropertyVisibility{
		HotelID:         hotelID,
		CatalogueItemID: catItemID,
		IsVisible:       req.IsVisible,
		PriceOverride:   req.PriceOverride,
	}
	if err := s.catalogue.SetVisibility(pv); err != nil {
		return internal(err, "failed to set visibility")
	}
	return nil
}

func (s *CommerceService) ListVisibleCatalogueForHotel(hotelID uuid.UUID) ([]models.CatalogueItem, error) {
	out, err := s.catalogue.ListVisibleForHotel(hotelID)
	if err != nil {
		return nil, internal(err, "failed to list visible catalogue")
	}
	return out, nil
}

// Commission rules + records -------------------------------------------------

func (s *CommerceService) ListCommissionRules() ([]models.CommissionRule, error) {
	out, err := s.commissions.ListRules()
	if err != nil {
		return nil, internal(err, "failed to list commission rules")
	}
	return out, nil
}

func (s *CommerceService) UpsertCommissionRule(req dto.UpsertCommissionRuleRequest) (*models.CommissionRule, error) {
	m := &models.CommissionRule{
		Category:               strings.TrimSpace(req.Category),
		PremiumApplicable:      req.PremiumApplicable,
		BaseMarginPct:          req.BaseMarginPct,
		PremiumPoolPct:         req.PremiumPoolPct,
		NechaPoolShareFounding: req.NechaPoolShareFounding,
		NechaPoolShareStandard: req.NechaPoolShareStandard,
		NechaCommissionPct:     req.NechaCommissionPct,
		InfluencerPctOfNecha:   req.InfluencerPctOfNecha,
		IsActive:               req.IsActive,
	}
	if err := s.commissions.UpsertRule(m); err != nil {
		return nil, internal(err, "failed to upsert commission rule")
	}
	return m, nil
}

func (s *CommerceService) ListCommissionRecords() ([]models.CommissionRecord, error) {
	out, err := s.commissions.ListRecords()
	if err != nil {
		return nil, internal(err, "failed to list commission records")
	}
	return out, nil
}

// Rewards --------------------------------------------------------------------

func (s *CommerceService) UpsertRewardRule(req dto.UpsertRewardRuleRequest) (*models.RewardRule, error) {
	m := &models.RewardRule{
		Code:                  strings.TrimSpace(req.Code),
		Name:                  strings.TrimSpace(req.Name),
		PointsPerCurrencyUnit: req.PointsPerCurrencyUnit,
		RedeemValuePerPoint:   req.RedeemValuePerPoint,
		IsActive:              req.IsActive,
	}
	if err := s.rewards.UpsertRule(m); err != nil {
		return nil, internal(err, "failed to upsert reward rule")
	}
	return m, nil
}

func (s *CommerceService) RewardBalance(userID uuid.UUID) (int64, []models.RewardLedgerEntry, error) {
	balance, err := s.rewards.BalanceForUser(userID)
	if err != nil {
		return 0, nil, internal(err, "failed to compute balance")
	}
	ledger, err := s.rewards.LedgerForUser(userID)
	if err != nil {
		return 0, nil, internal(err, "failed to load ledger")
	}
	return balance, ledger, nil
}

func (s *CommerceService) ActiveRewardRule() (*models.RewardRule, error) {
	return s.rewards.ActiveRule()
}

func (s *CommerceService) RedeemRewards(userID uuid.UUID, points int64) (int64, error) {
	if points <= 0 {
		return 0, apperrors.New(apperrors.ErrBadRequest.Code, "points must be positive", apperrors.ErrBadRequest.Status)
	}
	_, err := s.rewards.ActiveRule()
	if err != nil {
		return 0, internal(err, "rewards rule not configured")
	}
	balance, err := s.rewards.BalanceForUser(userID)
	if err != nil {
		return 0, internal(err, "failed to compute balance")
	}
	if points > balance {
		return 0, apperrors.New(apperrors.ErrBadRequest.Code, "insufficient points", apperrors.ErrBadRequest.Status)
	}
	if err := s.rewards.AppendEntry(&models.RewardLedgerEntry{
		UserID:    &userID,
		EntryType: models.RewardEntryTypeRedeem,
		Points:    -points,
		Note:      "Points redeemed",
	}); err != nil {
		return 0, internal(err, "failed to redeem points")
	}
	newBalance, err := s.rewards.BalanceForUser(userID)
	if err != nil {
		return 0, internal(err, "failed to compute balance")
	}
	return newBalance, nil
}

// EventLog -------------------------------------------------------------------

func (s *CommerceService) ListEventLog(limit int) ([]models.EventLog, error) {
	out, err := s.events.List(limit)
	if err != nil {
		return nil, internal(err, "failed to load event log")
	}
	return out, nil
}

func parseOptionalDate(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t
		}
	}
	return nil
}
