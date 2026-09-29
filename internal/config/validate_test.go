package config

import "testing"

func TestValidateDevelopmentAllowsDefaults(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Env: "development"},
		JWT:      JWTConfig{Secret: defaultJWTSecret},
		Database: DatabaseConfig{Password: "postgres"},
		Selcom:   SelcomConfig{APIKey: "live-key"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error in development, got %v", err)
	}
}

func TestValidateProductionRejectsWeakJWT(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Env: "production"},
		JWT:      JWTConfig{Secret: defaultJWTSecret},
		Database: DatabaseConfig{Password: "strong-db-password"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected JWT validation error in production")
	}
}

func TestValidateProductionRequiresSelcomWebhookSecret(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Env: "production"},
		JWT:      JWTConfig{Secret: "this-is-a-long-enough-production-secret"},
		Database: DatabaseConfig{Password: "strong-db-password"},
		Selcom: SelcomConfig{
			APIKey:        "live-key",
			WebhookSecret: "",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Selcom webhook secret validation error")
	}
}
