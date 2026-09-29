package service

import (
	"log"
	"strconv"
	"time"

	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/repository"
	"gorm.io/gorm"
)

// MaintenanceService runs the platform's scheduled background jobs (brief §3.14, §4.2):
//   - founding-tier auto-transition: founding-cohort properties revert to the standard
//     commission split once their tier window elapses.
//   - inventory reservation sweeper: capacity holds for tours/events/dining that were never
//     confirmed are expired so the seats/stock return to the pool.
//
// Jobs are idempotent and safe to run repeatedly, so a simple ticker is sufficient.
type MaintenanceService struct {
	hotels   *repository.HotelRepository
	config   *repository.PlatformConfigRepository
	events   *repository.EventLogRepository
	db       *gorm.DB
	interval time.Duration
}

func NewMaintenanceService(
	hotels *repository.HotelRepository,
	config *repository.PlatformConfigRepository,
	events *repository.EventLogRepository,
	db *gorm.DB,
) *MaintenanceService {
	return &MaintenanceService{
		hotels:   hotels,
		config:   config,
		events:   events,
		db:       db,
		interval: time.Hour,
	}
}

// Start launches the background loop. It runs one pass immediately, then on every tick.
func (s *MaintenanceService) Start() {
	go func() {
		s.RunOnce()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for range ticker.C {
			s.RunOnce()
		}
	}()
}

// RunOnce executes every maintenance job a single time.
func (s *MaintenanceService) RunOnce() {
	if n, err := s.TransitionFoundingTiers(time.Now()); err != nil {
		log.Printf("maintenance: founding-tier transition failed: %v", err)
	} else if n > 0 {
		log.Printf("maintenance: transitioned %d propert(ies) from founding to standard tier", n)
	}
	if n, err := s.ExpireInventoryReservations(time.Now()); err != nil {
		log.Printf("maintenance: inventory reservation sweep failed: %v", err)
	} else if n > 0 {
		log.Printf("maintenance: expired %d stale inventory reservation(s)", n)
	}
	if n, err := s.ExpireUnpaidProductHolds(time.Now()); err != nil {
		log.Printf("maintenance: unpaid product hold sweep failed: %v", err)
	} else if n > 0 {
		log.Printf("maintenance: released stock for %d expired unpaid product order(s)", n)
	}
}

// foundingTierMonths reads the configurable tier window, defaulting to 14 months (brief §4.2).
func (s *MaintenanceService) foundingTierMonths() int {
	months := 14
	if raw, err := s.config.Get(models.ConfigKeyFoundingTierMonths); err == nil && raw != "" {
		if parsed, perr := strconv.Atoi(raw); perr == nil && parsed > 0 {
			months = parsed
		}
	}
	return months
}

// TransitionFoundingTiers reverts founding-cohort properties to the standard split once their
// tier start date is older than the configured window. Returns the number transitioned.
func (s *MaintenanceService) TransitionFoundingTiers(now time.Time) (int, error) {
	hotels, err := s.hotels.ListAll()
	if err != nil {
		return 0, err
	}
	cutoff := now.AddDate(0, -s.foundingTierMonths(), 0)
	transitioned := 0
	for i := range hotels {
		h := hotels[i]
		if h.CommissionTier != models.CommissionTierFounding {
			continue
		}
		if h.CommissionTierStartDate == nil || h.CommissionTierStartDate.After(cutoff) {
			continue
		}
		h.CommissionTier = models.CommissionTierStandard
		if err := s.hotels.Update(&h); err != nil {
			return transitioned, err
		}
		transitioned++
		if s.events != nil {
			id := h.ID
			_ = s.events.Append(&models.EventLog{
				EventType:  "commission_tier_transitioned",
				EntityType: "hotel",
				EntityID:   &id,
				Detail:     "founding→standard",
			})
		}
	}
	return transitioned, nil
}

// ExpireInventoryReservations marks active holds whose expiry has passed as expired, returning
// the reserved capacity to the pool. Returns the number of reservations expired.
func (s *MaintenanceService) ExpireInventoryReservations(now time.Time) (int, error) {
	res := s.db.Model(&models.InventoryReservation{}).
		Where("status = ? AND expires_at < ?", models.InventoryReservationStatusActive, now).
		Update("status", models.InventoryReservationStatusExpired)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}

// ExpireUnpaidProductHolds cancels pending product orders whose payment never completed
// and restores the stock hold (brief §2.4 / §3.14).
func (s *MaintenanceService) ExpireUnpaidProductHolds(now time.Time) (int, error) {
	cutoff := unpaidHoldCutoff(now)
	var orders []models.Order
	if err := s.db.Preload("Items").
		Where("type = ? AND status = ? AND stock_held = ? AND created_at < ?",
			models.OrderTypeProduct, models.OrderStatusPending, true, cutoff).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	released := 0
	for i := range orders {
		order := &orders[i]
		for _, item := range order.Items {
			if item.ProductID == nil || item.Quantity <= 0 {
				continue
			}
			if err := s.hotels.RestoreProductStock(*item.ProductID, item.Quantity); err != nil {
				return released, err
			}
		}
		order.StockHeld = false
		order.Status = models.OrderStatusCancelled
		order.PaymentStatus = "expired"
		if err := s.db.Save(order).Error; err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}
