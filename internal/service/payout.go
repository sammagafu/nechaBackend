package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"github.com/nechaafrica/backend/internal/integration/selcom"
	"github.com/nechaafrica/backend/internal/repository"
	apperrors "github.com/nechaafrica/backend/pkg/errors"
	"gorm.io/gorm"
)

const (
	DisbursementStatusSkipped   = "skipped"
	DisbursementStatusCompleted = "completed"
	DisbursementStatusFailed    = "failed"
)

// PayoutService groups eligible commission records into settlement batches per payee and
// releases them (brief §8.3). Payout is only ever computed from records already marked
// eligible (i.e. fulfilment confirmed), enforcing the hold/release principle (§2.3).
type PayoutService struct {
	commissions *repository.CommissionRepository
	payouts     *repository.PayoutRepository
	events      *repository.EventLogRepository
	hotels      *repository.HotelRepository
	influencers *repository.InfluencerRepository
	selcom      selcom.Client
}

func NewPayoutService(
	commissions *repository.CommissionRepository,
	payouts *repository.PayoutRepository,
	events *repository.EventLogRepository,
	hotels *repository.HotelRepository,
	influencers *repository.InfluencerRepository,
	selcomClient selcom.Client,
) *PayoutService {
	return &PayoutService{
		commissions: commissions,
		payouts:     payouts,
		events:      events,
		hotels:      hotels,
		influencers: influencers,
		selcom:      selcomClient,
	}
}

// GenerateBatches builds draft payout batches for all eligible, un-batched commission
// records up to `until`. One property batch is created per hotel (canonical owner of each
// record), plus influencer and partner-referral batches attributed via batch items.
func (s *PayoutService) GenerateBatches(until time.Time) ([]models.PayoutBatch, error) {
	if until.IsZero() {
		until = time.Now()
	}
	records, err := s.commissions.ListEligibleRecords(until)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load eligible commissions", apperrors.ErrInternal.Status)
	}
	if len(records) == 0 {
		return []models.PayoutBatch{}, nil
	}

	byProperty := map[uuid.UUID][]models.CommissionRecord{}
	byInfluencer := map[uuid.UUID][]models.CommissionRecord{}
	byPartner := map[uuid.UUID][]models.CommissionRecord{}
	for _, rec := range records {
		byProperty[rec.HotelID] = append(byProperty[rec.HotelID], rec)
		if rec.InfluencerID != nil && rec.InfluencerShare > 0 {
			byInfluencer[*rec.InfluencerID] = append(byInfluencer[*rec.InfluencerID], rec)
		}
		if rec.PartnerReferralID != nil && rec.PartnerReferralShare > 0 {
			byPartner[*rec.PartnerReferralID] = append(byPartner[*rec.PartnerReferralID], rec)
		}
	}

	created := []models.PayoutBatch{}
	err = s.payouts.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		for hotelID, recs := range byProperty {
			batch := &models.PayoutBatch{
				PayeeType: models.PayeeTypeProperty,
				PayeeID:   hotelID,
				Currency:  currencyOf(recs),
				Status:    models.PayoutBatchStatusDraft,
				PeriodEnd: &until,
			}
			var total int64
			for _, rec := range recs {
				total += rec.PropertyShare
			}
			batch.TotalAmount = total
			batch.RecordCount = len(recs)
			if err := tx.Create(batch).Error; err != nil {
				return err
			}
			for _, rec := range recs {
				if err := tx.Create(&models.PayoutBatchItem{
					PayoutBatchID:      batch.ID,
					CommissionRecordID: rec.ID,
					PayeeType:          models.PayeeTypeProperty,
					Amount:             rec.PropertyShare,
				}).Error; err != nil {
					return err
				}
				if err := tx.Model(&models.CommissionRecord{}).Where("id = ?", rec.ID).
					Updates(map[string]interface{}{"payout_batch_id": batch.ID, "updated_at": now}).Error; err != nil {
					return err
				}
			}
			created = append(created, *batch)
		}

		for influencerID, recs := range byInfluencer {
			batch, err := createShareBatch(tx, models.PayeeTypeInfluencer, influencerID, recs, until, func(r models.CommissionRecord) int64 { return r.InfluencerShare })
			if err != nil {
				return err
			}
			created = append(created, *batch)
		}

		for partnerID, recs := range byPartner {
			batch, err := createShareBatch(tx, models.PayeeTypePartnerReferral, partnerID, recs, until, func(r models.CommissionRecord) int64 { return r.PartnerReferralShare })
			if err != nil {
				return err
			}
			created = append(created, *batch)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to generate payout batches", apperrors.ErrInternal.Status)
	}
	s.logEvent("payout_batches_generated", "payout_batch", nil, "")
	return created, nil
}

func createShareBatch(
	tx *gorm.DB,
	payeeType string,
	payeeID uuid.UUID,
	recs []models.CommissionRecord,
	until time.Time,
	share func(models.CommissionRecord) int64,
) (*models.PayoutBatch, error) {
	batch := &models.PayoutBatch{
		PayeeType:   payeeType,
		PayeeID:     payeeID,
		Currency:    currencyOf(recs),
		Status:      models.PayoutBatchStatusDraft,
		RecordCount: len(recs),
		PeriodEnd:   &until,
	}
	var total int64
	for _, rec := range recs {
		total += share(rec)
	}
	batch.TotalAmount = total
	if err := tx.Create(batch).Error; err != nil {
		return nil, err
	}
	for _, rec := range recs {
		if err := tx.Create(&models.PayoutBatchItem{
			PayoutBatchID:      batch.ID,
			CommissionRecordID: rec.ID,
			PayeeType:          payeeType,
			Amount:             share(rec),
		}).Error; err != nil {
			return nil, err
		}
	}
	return batch, nil
}

// Release marks a batch released and triggers Selcom wallet disbursement when configured.
func (s *PayoutService) Release(batchID uuid.UUID) (*models.PayoutBatch, error) {
	batch, err := s.payouts.FindByID(batchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrNotFound.Code, "payout batch not found", apperrors.ErrNotFound.Status)
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to load batch", apperrors.ErrInternal.Status)
	}
	if batch.Status == models.PayoutBatchStatusReleased {
		return batch, nil // idempotent — never disburse the same batch twice
	}

	disburseRef, disburseStatus := s.disburseBatch(batch)
	if disburseStatus == DisbursementStatusFailed {
		return nil, apperrors.New(apperrors.ErrInternal.Code, "disbursement failed; batch not released", apperrors.ErrInternal.Status)
	}

	now := time.Now()
	err = s.payouts.Transaction(func(tx *gorm.DB) error {
		batch.Status = models.PayoutBatchStatusReleased
		batch.ReleasedAt = &now
		batch.DisbursementRef = disburseRef
		batch.DisbursementStatus = disburseStatus
		if err := tx.Save(batch).Error; err != nil {
			return err
		}
		if batch.PayeeType == models.PayeeTypeProperty {
			if err := tx.Model(&models.CommissionRecord{}).
				Where("payout_batch_id = ?", batch.ID).
				Update("status", models.CommissionStatusPaid).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal.Code, "failed to release batch", apperrors.ErrInternal.Status)
	}
	s.logEvent("payout_released", "payout_batch", &batch.ID, batch.PayeeType+":"+disburseStatus)
	return batch, nil
}

func (s *PayoutService) disburseBatch(batch *models.PayoutBatch) (ref, status string) {
	if s.selcom == nil || batch.TotalAmount <= 0 {
		return "", DisbursementStatusSkipped
	}

	account, narrative := s.payoutAccountFor(batch)
	if account == "" {
		return "", DisbursementStatusSkipped
	}

	result, err := s.selcom.DisburseWallet(context.Background(), selcom.DisburseInput{
		Reference: batch.ID.String(),
		Account:   account,
		Amount:    batch.TotalAmount,
		Currency:  batch.Currency,
		Narrative: narrative,
	})
	if err != nil {
		return "", DisbursementStatusFailed
	}
	return result.Reference, disbursementStatusFromResult(result)
}

func disbursementStatusFromResult(result *selcom.DisburseResult) string {
	if result == nil {
		return DisbursementStatusFailed
	}
	if result.Status == "demo_completed" || strings.HasPrefix(result.Reference, "MOCK-DISB") {
		return "demo_completed"
	}
	return DisbursementStatusCompleted
}

func (s *PayoutService) payoutAccountFor(batch *models.PayoutBatch) (account, narrative string) {
	switch batch.PayeeType {
	case models.PayeeTypeProperty, models.PayeeTypePartnerReferral, models.PayeeTypeTourOperator:
		if s.hotels == nil {
			return "", ""
		}
		hotel, err := s.hotels.FindByID(batch.PayeeID)
		if err != nil {
			return "", ""
		}
		label := "payout"
		if batch.PayeeType != models.PayeeTypeProperty {
			label = "referral payout"
		}
		return hotel.SelcomPayoutAccount, fmt.Sprintf("Necha %s %s", label, hotel.Name)
	case models.PayeeTypeInfluencer:
		if s.influencers == nil {
			return "", ""
		}
		inf, err := s.influencers.FindByID(batch.PayeeID)
		if err != nil {
			return "", ""
		}
		return inf.SelcomPayoutAccount, fmt.Sprintf("Necha influencer payout %s", inf.Name)
	default:
		return "", ""
	}
}

func (s *PayoutService) List() ([]models.PayoutBatch, error) { return s.payouts.List() }

func currencyOf(recs []models.CommissionRecord) string {
	for _, r := range recs {
		if r.Currency != "" {
			return r.Currency
		}
	}
	return "TZS"
}

func (s *PayoutService) logEvent(eventType, entityType string, entityID *uuid.UUID, detail string) {
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
