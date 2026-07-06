package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Key          string    `gorm:"uniqueIndex;not null" json:"key"`
	Value        string    `gorm:"not null" json:"value"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (p *PlatformConfig) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

const (
	ConfigKeyTzsToUsdRate = "tzs_to_usd_rate"
	// Distance-based delivery pricing (brief §4.5).
	ConfigKeyDeliveryBaseFeeTZS = "delivery_base_fee_tzs"
	ConfigKeyDeliveryPerKmTZS   = "delivery_per_km_tzs"
	// Commission tier transition window in months (founding → standard, brief §4.2, §8).
	ConfigKeyFoundingTierMonths = "founding_tier_months"

	// Feature flags — admin toggles (stored as "true" / "false").
	ConfigKeyFeatureRewardsEnabled              = "feature_rewards_enabled"
	ConfigKeyFeatureRewardsRedeemEnabled        = "feature_rewards_redeem_enabled"
	ConfigKeyFeatureDiscoveryTicketingEnabled   = "feature_discovery_ticketing_enabled"
	ConfigKeyFeaturePartnerPortalEnabled        = "feature_partner_portal_enabled"
	ConfigKeyFeaturePartnerProductsManageEnabled = "feature_partner_products_manage_enabled"
	ConfigKeyFeatureDualCurrencyEnabled         = "feature_dual_currency_enabled"
	ConfigKeyFeatureDistanceDeliveryEnabled     = "feature_distance_delivery_enabled"
)
