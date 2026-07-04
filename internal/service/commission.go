package service

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/repository"
	"gorm.io/gorm"
)

// CommissionService encapsulates the commercial split rules (brief §8). All rates come from
// CommissionRule rows, never hardcoded, so the founding/standard product split recomputes
// automatically if the base margin or premium pool is ever adjusted.
type CommissionService struct {
	commissions *repository.CommissionRepository
	influencers *repository.InfluencerRepository
	events      *repository.EventLogRepository
}

func NewCommissionService(
	commissions *repository.CommissionRepository,
	influencers *repository.InfluencerRepository,
	events *repository.EventLogRepository,
) *CommissionService {
	return &CommissionService{commissions: commissions, influencers: influencers, events: events}
}

// shareBreakdown is the computed split before persistence.
type shareBreakdown struct {
	GMV                  int64
	NechaShare           int64
	PropertyShare        int64
	SupplierShare        int64
	InfluencerShare      int64
	PartnerReferralShare int64
}

func roundShare(v float64) int64 { return int64(math.Round(v)) }

// computeSplit implements the §8.1 category math. tier is the property's commission tier.
func computeSplit(rule *models.CommissionRule, gmv int64, tier string) shareBreakdown {
	b := shareBreakdown{GMV: gmv}
	if gmv <= 0 {
		return b
	}
	if rule.PremiumApplicable {
		// Product: 20% base retail margin + 13% premium pool, split by tier.
		poolShare := rule.NechaPoolShareStandard
		if tier == models.CommissionTierFounding {
			poolShare = rule.NechaPoolShareFounding
		}
		premiumPool := float64(gmv) * rule.PremiumPoolPct
		nechaFromPool := premiumPool * poolShare
		b.PropertyShare = roundShare(premiumPool - nechaFromPool)
		baseMargin := float64(gmv) * rule.BaseMarginPct
		b.NechaShare = roundShare(baseMargin + nechaFromPool)
		b.SupplierShare = gmv - b.NechaShare - b.PropertyShare
		return b
	}
	// Flat categories (tours, events, dining/spa external): single Necha commission pct.
	b.NechaShare = roundShare(float64(gmv) * rule.NechaCommissionPct)
	b.SupplierShare = gmv - b.NechaShare
	return b
}

// Generate creates the commission record for an order at payment capture (brief §8.2).
// Free-utility categories (room service, standard reservations) produce no record.
func (s *CommissionService) Generate(order *models.Order, hotel *models.Hotel) error {
	if order == nil || hotel == nil {
		return nil
	}
	category := order.Category
	if category == "" {
		category = models.OrderCategoryProduct
	}
	// Free-utility categories never create a commission record.
	if category == models.OrderCategoryRoomService || category == models.OrderCategoryReservation {
		return nil
	}
	// Idempotency: never double-create for the same order.
	if _, err := s.commissions.FindRecordByOrder(order.ID); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	rule, err := s.commissions.FindRuleByCategory(category)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // no rule configured for this category yet
		}
		return err
	}

	tier := hotel.CommissionTier
	if tier == "" {
		tier = models.CommissionTierStandard
	}
	// GMV excludes any delivery fee bundled into the order total is out of scope here;
	// TotalAmount is treated as the commissionable GMV.
	split := computeSplit(rule, order.TotalAmount, tier)

	rec := &models.CommissionRecord{
		OrderID:        order.ID,
		HotelID:        order.HotelID,
		Category:       category,
		Currency:       order.Currency,
		GMV:            split.GMV,
		NechaShare:     split.NechaShare,
		PropertyShare:  split.PropertyShare,
		SupplierShare:  split.SupplierShare,
		CommissionTier: tier,
		Status:         models.CommissionStatusPending,
	}

	// Referral modifier — deducted from Necha's share only, never property/supplier.
	// Influencer and partner referral are mutually exclusive (brief §8.1.1).
	if order.ReferredByInfluencerID != nil {
		pct := rule.InfluencerPctOfNecha
		if inf, err := s.influencers.FindByID(*order.ReferredByInfluencerID); err == nil && inf.CommissionPct > 0 {
			pct = inf.CommissionPct
		}
		rec.InfluencerShare = roundShare(float64(rec.NechaShare) * pct)
		rec.NechaShare -= rec.InfluencerShare
		rec.InfluencerID = order.ReferredByInfluencerID
	} else if order.ReferredByPartnerID != nil {
		pct := rule.InfluencerPctOfNecha // partner rate reuses the same modifier default
		rec.PartnerReferralShare = roundShare(float64(rec.NechaShare) * pct)
		rec.NechaShare -= rec.PartnerReferralShare
		rec.PartnerReferralID = order.ReferredByPartnerID
	}

	if err := s.commissions.CreateRecord(rec); err != nil {
		return err
	}
	s.logEvent("commission_generated", "order", &order.ID, category)
	return nil
}

// MarkFulfilled flips an order's commission records to eligible for payout, computing the
// eligibility timestamp per category timing (brief §2.3): delivery for goods, and a window
// after the event/tour date for experiences.
func (s *CommissionService) MarkFulfilled(order *models.Order) error {
	if order == nil {
		return nil
	}
	eligibleAt := time.Now()
	switch order.Category {
	case models.OrderCategoryTour, models.OrderCategorySpaExternal:
		eligibleAt = eligibleAt.Add(48 * time.Hour) // no-show/dispute window
	case models.OrderCategoryEvent:
		// Events settle only after the event date; approximated here as +48h.
		eligibleAt = eligibleAt.Add(48 * time.Hour)
	}
	if err := s.commissions.MarkEligible(order.ID, eligibleAt); err != nil {
		return err
	}
	s.logEvent("delivery_confirmed", "order", &order.ID, string(order.Status))
	return nil
}

// Void reverses commission for a refunded order (brief §12.3) — marked void, never deleted.
func (s *CommissionService) Void(orderID uuid.UUID) error {
	if err := s.commissions.VoidByOrder(orderID); err != nil {
		return err
	}
	s.logEvent("commission_reversed", "order", &orderID, "")
	return nil
}

func (s *CommissionService) ListRecords() ([]models.CommissionRecord, error) {
	return s.commissions.ListRecords()
}

func (s *CommissionService) ListRules() ([]models.CommissionRule, error) {
	return s.commissions.ListRules()
}

func (s *CommissionService) logEvent(eventType, entityType string, entityID *uuid.UUID, detail string) {
	if s.events == nil {
		return
	}
	_ = s.events.Append(&models.EventLog{
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Detail:     detail,
	})
}

// SeedDefaultCommissionRules inserts the launch commercial terms (brief §8.1) if absent.
func SeedDefaultCommissionRules(repo *repository.CommissionRepository) error {
	defaults := []models.CommissionRule{
		{
			Category:               models.OrderCategoryProduct,
			PremiumApplicable:      true,
			BaseMarginPct:          0.20,
			PremiumPoolPct:         0.13,
			NechaPoolShareFounding: 0.50,
			NechaPoolShareStandard: 0.70,
			InfluencerPctOfNecha:   0.30,
			IsActive:               true,
		},
		{Category: models.OrderCategoryTour, NechaCommissionPct: 0.15, InfluencerPctOfNecha: 0.30, IsActive: true},
		{Category: models.OrderCategoryEvent, NechaCommissionPct: 0.12, InfluencerPctOfNecha: 0.30, IsActive: true},
		{Category: models.OrderCategoryDiningExternal, NechaCommissionPct: 0.125, InfluencerPctOfNecha: 0.30, IsActive: true},
		{Category: models.OrderCategorySpaExternal, NechaCommissionPct: 0.125, InfluencerPctOfNecha: 0.30, IsActive: true},
	}
	for i := range defaults {
		rule := defaults[i]
		if _, err := repo.FindRuleByCategory(rule.Category); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := repo.UpsertRule(&rule); err != nil {
			return err
		}
	}
	return nil
}
