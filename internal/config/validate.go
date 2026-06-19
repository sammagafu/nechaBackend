package config

import (
	"fmt"
	"strings"
)

const defaultJWTSecret = "change-me-in-production"

func (c *Config) Validate() error {
	if c.Server.Env != "production" {
		return nil
	}
	if c.JWT.Secret == defaultJWTSecret || len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be a strong secret (32+ chars) in production")
	}
	if strings.EqualFold(c.Database.Password, "postgres") {
		return fmt.Errorf("DB_PASSWORD must not use the default in production")
	}
	if !c.Selcom.MockMode && c.Selcom.APIKey != "" && c.Selcom.WebhookSecret == "" {
		return fmt.Errorf("SELCOM_WEBHOOK_SECRET is required when Selcom live payments are enabled")
	}
	return nil
}

func (c *Config) SeedDemoUsers() bool {
	if c.Server.Env != "production" {
		return true
	}
	return strings.EqualFold(getEnv("SEED_DEMO_DATA", ""), "true")
}
