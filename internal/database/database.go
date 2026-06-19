package database

import (
	"fmt"
	"time"

	"github.com/nechaafrica/backend/internal/domain/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dsn string) (*gorm.DB, error) {
	return ConnectWithRetry(dsn, 15, 2*time.Second)
}

func ConnectWithRetry(dsn string, attempts int, delay time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= attempts; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			return db, nil
		}
		if i < attempts {
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("database connection failed after %d attempts: %w", attempts, err)
}

// migrateLegacy backfills new storefront columns on databases created before v2.0.
func migrateLegacy(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Hotel{}) {
		return nil
	}

	hotelSteps := []struct {
		ddl      string
		backfill string
	}{
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS slug text`, `UPDATE hotels SET slug = LOWER(REPLACE(REPLACE(code, '_', '-'), ' ', '-')) WHERE slug IS NULL OR slug = ''`},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS location text`, `UPDATE hotels SET location = city WHERE location IS NULL OR location = ''`},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS zone text`, ``},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS initials text`, `UPDATE hotels SET initials = UPPER(LEFT(name, 2)) WHERE initials IS NULL OR initials = ''`},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS logo_url text`, ``},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS referral_code text`, `UPDATE hotels SET referral_code = code WHERE referral_code IS NULL OR referral_code = ''`},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS services jsonb DEFAULT '[]'`, `UPDATE hotels SET services = '[]' WHERE services IS NULL`},
		{`ALTER TABLE hotels ADD COLUMN IF NOT EXISTS is_verified boolean DEFAULT true`, `UPDATE hotels SET is_verified = true WHERE is_verified IS NULL`},
	}

	for _, step := range hotelSteps {
		if err := db.Exec(step.ddl).Error; err != nil {
			return fmt.Errorf("legacy migration ddl: %w", err)
		}
		if step.backfill != "" {
			if err := db.Exec(step.backfill).Error; err != nil {
				return fmt.Errorf("legacy migration backfill: %w", err)
			}
		}
	}

	if db.Migrator().HasTable(&models.Conversation{}) {
		steps := []struct {
			ddl      string
			backfill string
		}{
			{`ALTER TABLE conversations ADD COLUMN IF NOT EXISTS category text`, `UPDATE conversations SET category = COALESCE(NULLIF(subject, ''), 'other') WHERE category IS NULL OR category = ''`},
			{`ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider text DEFAULT 'email'`, `UPDATE users SET auth_provider = 'email' WHERE auth_provider IS NULL OR auth_provider = ''`},
			{`ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id text`, ``},
			{`ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL`, ``},
		}
		for _, step := range steps {
			if err := db.Exec(step.ddl).Error; err != nil {
				return fmt.Errorf("legacy conversation migration ddl: %w", err)
			}
			if step.backfill != "" {
				if err := db.Exec(step.backfill).Error; err != nil {
					return fmt.Errorf("legacy conversation migration backfill: %w", err)
				}
			}
		}
	}

	if !db.Migrator().HasTable(&models.Product{}) {
		return nil
	}

	productSteps := []struct {
		ddl      string
		backfill string
	}{
		{`ALTER TABLE products ADD COLUMN IF NOT EXISTS slug text`, `UPDATE products SET slug = LOWER(REPLACE(REPLACE(name, ' ', '-'), '_', '-')) WHERE slug IS NULL OR slug = ''`},
		{`ALTER TABLE products ADD COLUMN IF NOT EXISTS brand_name text`, `UPDATE products SET brand_name = 'NECHA NATURALS' WHERE brand_name IS NULL OR brand_name = ''`},
		{`ALTER TABLE products ADD COLUMN IF NOT EXISTS badge text`, ``},
		{`ALTER TABLE products ADD COLUMN IF NOT EXISTS is_featured boolean DEFAULT false`, `UPDATE products SET is_featured = false WHERE is_featured IS NULL`},
	}

	for _, step := range productSteps {
		if err := db.Exec(step.ddl).Error; err != nil {
			return fmt.Errorf("legacy product migration: %w", err)
		}
		if step.backfill != "" {
			if err := db.Exec(step.backfill).Error; err != nil {
				return fmt.Errorf("legacy product backfill: %w", err)
			}
		}
	}

	return nil
}

func Migrate(db *gorm.DB) error {
	if err := migrateLegacy(db); err != nil {
		return err
	}
	return db.AutoMigrate(
		&models.User{},
		&models.Hotel{},
		&models.Product{},
		&models.Reservation{},
		&models.Order{},
		&models.OrderItem{},
		&models.DiscoveryItem{},
		&models.Notification{},
		&models.SystemAlert{},
		&models.Conversation{},
		&models.Message{},
		&models.WebhookEndpoint{},
		&models.WebhookDelivery{},
		&models.GuestStay{},
		&models.HotelRoom{},
		&models.HotelCategory{},
		&models.HotelMenuItem{},
	)
}
