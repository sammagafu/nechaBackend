package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

// InfluencerRepository -------------------------------------------------------

type InfluencerRepository struct {
	db *gorm.DB
}

func NewInfluencerRepository(db *gorm.DB) *InfluencerRepository {
	return &InfluencerRepository{db: db}
}

func (r *InfluencerRepository) Create(m *models.Influencer) error { return r.db.Create(m).Error }
func (r *InfluencerRepository) Update(m *models.Influencer) error { return r.db.Save(m).Error }

func (r *InfluencerRepository) FindByID(id uuid.UUID) (*models.Influencer, error) {
	var m models.Influencer
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *InfluencerRepository) FindByReferralCode(code string) (*models.Influencer, error) {
	var m models.Influencer
	if err := r.db.Where("referral_code = ? AND status = ?", code, models.InfluencerStatusActive).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *InfluencerRepository) List() ([]models.Influencer, error) {
	var out []models.Influencer
	err := r.db.Order("created_at DESC").Find(&out).Error
	return out, err
}

// BookingReferralRepository --------------------------------------------------

type BookingReferralRepository struct {
	db *gorm.DB
}

func NewBookingReferralRepository(db *gorm.DB) *BookingReferralRepository {
	return &BookingReferralRepository{db: db}
}

func (r *BookingReferralRepository) Create(m *models.BookingReferral) error { return r.db.Create(m).Error }
func (r *BookingReferralRepository) Update(m *models.BookingReferral) error { return r.db.Save(m).Error }

func (r *BookingReferralRepository) FindByToken(token string) (*models.BookingReferral, error) {
	var m models.BookingReferral
	if err := r.db.Where("referral_token = ?", token).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// FindActiveByPhone returns a pending/converted referral whose trip window still covers now.
func (r *BookingReferralRepository) FindActiveByPhone(phone string, at time.Time) (*models.BookingReferral, error) {
	var m models.BookingReferral
	err := r.db.Where("traveller_phone = ? AND status <> ?", phone, models.BookingReferralStatusExpired).
		Where("trip_end_date IS NULL OR trip_end_date >= ?", at).
		Order("created_at DESC").First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *BookingReferralRepository) List() ([]models.BookingReferral, error) {
	var out []models.BookingReferral
	err := r.db.Order("created_at DESC").Find(&out).Error
	return out, err
}

func (r *BookingReferralRepository) ListByPartner(partnerID uuid.UUID, limit int) ([]models.BookingReferral, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []models.BookingReferral
	err := r.db.Where("referring_partner_id = ?", partnerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&out).Error
	return out, err
}

// SupplierRepository ---------------------------------------------------------

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository { return &SupplierRepository{db: db} }

func (r *SupplierRepository) Create(m *models.Supplier) error { return r.db.Create(m).Error }
func (r *SupplierRepository) Update(m *models.Supplier) error { return r.db.Save(m).Error }

func (r *SupplierRepository) FindByID(id uuid.UUID) (*models.Supplier, error) {
	var m models.Supplier
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SupplierRepository) FindByNameAndType(name, supplierType string) (*models.Supplier, error) {
	var m models.Supplier
	if err := r.db.Where("LOWER(name) = LOWER(?) AND supplier_type = ?", name, supplierType).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// CatalogueRepository --------------------------------------------------------

type CatalogueRepository struct {
	db *gorm.DB
}

func NewCatalogueRepository(db *gorm.DB) *CatalogueRepository { return &CatalogueRepository{db: db} }

func (r *CatalogueRepository) Create(m *models.CatalogueItem) error { return r.db.Create(m).Error }
func (r *CatalogueRepository) Update(m *models.CatalogueItem) error { return r.db.Save(m).Error }

func (r *CatalogueRepository) FindByID(id uuid.UUID) (*models.CatalogueItem, error) {
	var m models.CatalogueItem
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *CatalogueRepository) List() ([]models.CatalogueItem, error) {
	var out []models.CatalogueItem
	err := r.db.Order("created_at DESC").Find(&out).Error
	return out, err
}

// ListVisibleForHotel returns catalogue items exposed to a property via PropertyVisibility.
func (r *CatalogueRepository) ListVisibleForHotel(hotelID uuid.UUID) ([]models.CatalogueItem, error) {
	var out []models.CatalogueItem
	err := r.db.
		Joins("JOIN property_visibilities pv ON pv.catalogue_item_id = catalogue_items.id").
		Where("pv.hotel_id = ? AND pv.is_visible = ? AND catalogue_items.is_active = ?", hotelID, true, true).
		Order("catalogue_items.created_at DESC").Find(&out).Error
	return out, err
}

func (r *CatalogueRepository) SetVisibility(m *models.PropertyVisibility) error {
	return r.db.Where("hotel_id = ? AND catalogue_item_id = ?", m.HotelID, m.CatalogueItemID).
		Assign(map[string]interface{}{"is_visible": m.IsVisible, "price_override": m.PriceOverride}).
		FirstOrCreate(m).Error
}

func (r *CatalogueRepository) ListVisibilityForHotel(hotelID uuid.UUID) ([]models.PropertyVisibility, error) {
	var out []models.PropertyVisibility
	err := r.db.Where("hotel_id = ?", hotelID).Find(&out).Error
	return out, err
}

// CommissionRepository -------------------------------------------------------

type CommissionRepository struct {
	db *gorm.DB
}

func NewCommissionRepository(db *gorm.DB) *CommissionRepository {
	return &CommissionRepository{db: db}
}

func (r *CommissionRepository) FindRuleByCategory(category string) (*models.CommissionRule, error) {
	var m models.CommissionRule
	if err := r.db.Where("category = ? AND is_active = ?", category, true).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *CommissionRepository) UpsertRule(m *models.CommissionRule) error {
	var existing models.CommissionRule
	err := r.db.Where("category = ?", m.Category).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(m).Error
	}
	if err != nil {
		return err
	}
	m.ID = existing.ID
	return r.db.Save(m).Error
}

func (r *CommissionRepository) ListRules() ([]models.CommissionRule, error) {
	var out []models.CommissionRule
	err := r.db.Order("category ASC").Find(&out).Error
	return out, err
}

func (r *CommissionRepository) CreateRecord(m *models.CommissionRecord) error {
	return r.db.Create(m).Error
}

func (r *CommissionRepository) UpdateRecord(m *models.CommissionRecord) error {
	return r.db.Save(m).Error
}

func (r *CommissionRepository) FindRecordByOrder(orderID uuid.UUID) (*models.CommissionRecord, error) {
	var m models.CommissionRecord
	if err := r.db.Where("order_id = ?", orderID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *CommissionRepository) ListRecordsByHotel(hotelID uuid.UUID, limit int) ([]models.CommissionRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var out []models.CommissionRecord
	err := r.db.Where("hotel_id = ?", hotelID).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

func (r *CommissionRepository) ListRecords() ([]models.CommissionRecord, error) {
	var out []models.CommissionRecord
	err := r.db.Order("created_at DESC").Find(&out).Error
	return out, err
}

// MarkEligible flips pending records for an order to eligible once fulfilment is confirmed.
func (r *CommissionRepository) MarkEligible(orderID uuid.UUID, at time.Time) error {
	return r.db.Model(&models.CommissionRecord{}).
		Where("order_id = ? AND status = ?", orderID, models.CommissionStatusPending).
		Updates(map[string]interface{}{"status": models.CommissionStatusEligible, "payout_eligible_at": at}).Error
}

// VoidByOrder reverses commission for a refunded order (marked void, never deleted).
func (r *CommissionRepository) VoidByOrder(orderID uuid.UUID) error {
	return r.db.Model(&models.CommissionRecord{}).
		Where("order_id = ? AND status <> ?", orderID, models.CommissionStatusPaid).
		Update("status", models.CommissionStatusVoid).Error
}

// ListEligibleRecords returns eligible, un-batched records up to the given time.
func (r *CommissionRepository) ListEligibleRecords(until time.Time) ([]models.CommissionRecord, error) {
	var out []models.CommissionRecord
	err := r.db.Where("status = ? AND payout_batch_id IS NULL AND payout_eligible_at <= ?",
		models.CommissionStatusEligible, until).Find(&out).Error
	return out, err
}

func (r *CommissionRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// PayoutRepository -----------------------------------------------------------

type PayoutRepository struct {
	db *gorm.DB
}

func NewPayoutRepository(db *gorm.DB) *PayoutRepository { return &PayoutRepository{db: db} }

func (r *PayoutRepository) Create(m *models.PayoutBatch) error { return r.db.Create(m).Error }
func (r *PayoutRepository) Update(m *models.PayoutBatch) error { return r.db.Save(m).Error }

func (r *PayoutRepository) FindByID(id uuid.UUID) (*models.PayoutBatch, error) {
	var m models.PayoutBatch
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PayoutRepository) List() ([]models.PayoutBatch, error) {
	var out []models.PayoutBatch
	err := r.db.Order("created_at DESC").Find(&out).Error
	return out, err
}

func (r *PayoutRepository) RecordsForBatch(batchID uuid.UUID) ([]models.CommissionRecord, error) {
	var out []models.CommissionRecord
	err := r.db.Where("payout_batch_id = ?", batchID).Find(&out).Error
	return out, err
}

func (r *PayoutRepository) CreateItem(item *models.PayoutBatchItem) error {
	return r.db.Create(item).Error
}

func (r *PayoutRepository) ItemsForBatch(batchID uuid.UUID) ([]models.PayoutBatchItem, error) {
	var out []models.PayoutBatchItem
	err := r.db.Where("payout_batch_id = ?", batchID).Find(&out).Error
	return out, err
}

func (r *PayoutRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// EventLogRepository ---------------------------------------------------------

type EventLogRepository struct {
	db *gorm.DB
}

func NewEventLogRepository(db *gorm.DB) *EventLogRepository { return &EventLogRepository{db: db} }

func (r *EventLogRepository) Append(entry *models.EventLog) error { return r.db.Create(entry).Error }

func (r *EventLogRepository) List(limit int) ([]models.EventLog, error) {
	if limit <= 0 {
		limit = 200
	}
	var out []models.EventLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

// RewardRepository -----------------------------------------------------------

type RewardRepository struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) *RewardRepository { return &RewardRepository{db: db} }

func (r *RewardRepository) ActiveRule() (*models.RewardRule, error) {
	var m models.RewardRule
	if err := r.db.Where("is_active = ?", true).Order("created_at DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *RewardRepository) UpsertRule(m *models.RewardRule) error {
	var existing models.RewardRule
	err := r.db.Where("code = ?", m.Code).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(m).Error
	}
	if err != nil {
		return err
	}
	m.ID = existing.ID
	return r.db.Save(m).Error
}

func (r *RewardRepository) AppendEntry(entry *models.RewardLedgerEntry) error {
	return r.db.Create(entry).Error
}

func (r *RewardRepository) BalanceForUser(userID uuid.UUID) (int64, error) {
	var balance int64
	err := r.db.Model(&models.RewardLedgerEntry{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(points),0)").Scan(&balance).Error
	return balance, err
}

func (r *RewardRepository) LedgerForUser(userID uuid.UUID) ([]models.RewardLedgerEntry, error) {
	var out []models.RewardLedgerEntry
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&out).Error
	return out, err
}
