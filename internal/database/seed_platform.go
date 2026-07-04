package database

import (
	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/gorm"
)

func seedPlatformDefaults(db *gorm.DB) error {
	zones := []models.DeliveryZone{
		{Code: "A", Label: "Zone A — City Centre", DeliveryFeeTZS: 5000, FreeThresholdTZS: 180000, SortOrder: 1, IsActive: true},
		{Code: "B", Label: "Zone B — Masaki & Oyster Bay", DeliveryFeeTZS: 5000, FreeThresholdTZS: 180000, SortOrder: 2, IsActive: true},
		{Code: "C", Label: "Zone C — Mikocheni & Upanga", DeliveryFeeTZS: 7000, FreeThresholdTZS: 180000, SortOrder: 3, IsActive: true},
		{Code: "D", Label: "Zone D — Mbezi & Tegeta", DeliveryFeeTZS: 8000, FreeThresholdTZS: 180000, SortOrder: 4, IsActive: true},
	}
	for _, z := range zones {
		var existing models.DeliveryZone
		if err := db.Where("code = ?", z.Code).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&z).Error; err != nil {
			return err
		}
	}

	var cfg models.PlatformConfig
	if err := db.Where("key = ?", models.ConfigKeyTzsToUsdRate).First(&cfg).Error; err != nil {
		if err := db.Create(&models.PlatformConfig{Key: models.ConfigKeyTzsToUsdRate, Value: "2625"}).Error; err != nil {
			return err
		}
	}
	defaults := map[string]string{
		models.ConfigKeyDeliveryBaseFeeTZS: "3000",
		models.ConfigKeyDeliveryPerKmTZS:   "1000",
		models.ConfigKeyFoundingTierMonths: "13",
	}
	for key, value := range defaults {
		var row models.PlatformConfig
		if err := db.Where("key = ?", key).First(&row).Error; err != nil {
			if err := db.Create(&models.PlatformConfig{Key: key, Value: value}).Error; err != nil {
				return err
			}
		}
	}

	var reward models.RewardRule
	if err := db.Where("code = ?", "default").First(&reward).Error; err != nil {
		if err := db.Create(&models.RewardRule{
			Code:                  "default",
			Name:                  "Necha Rewards",
			PointsPerCurrencyUnit: 0.01,
			RedeemValuePerPoint:   10,
			IsActive:              true,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
