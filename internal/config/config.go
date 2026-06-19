package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Kkooapp  KkooappConfig
	Selcom   SelcomConfig
	Webhook  WebhookConfig
}

type WebhookConfig struct {
	InboundSecret string
}

type ServerConfig struct {
	Port          string
	Env           string
	AllowedOrigin string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type OAuthConfig struct {
	GoogleClientID string
	AppleClientID  string
}

type SelcomConfig struct {
	BaseURL        string
	APIKey         string
	APISecret      string
	Vendor         string
	WebhookSecret  string
	PublicAPIURL   string
	PublicAppURL   string
	TimeoutSeconds int
	MockMode       bool
}

type KkooappConfig struct {
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
	Endpoints      KkooappEndpoints
}

type KkooappEndpoints struct {
	HotelReservation string
	TableReservation string
	FoodOrder        string
	OrderStatus      string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:          getEnv("SERVER_PORT", "8080"),
			Env:           getEnv("APP_ENV", "development"),
			AllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "https://necha.africa"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "nechaafrica"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenTTL:  parseDuration(getEnv("JWT_ACCESS_TTL", "24h")),
			RefreshTokenTTL: parseDuration(getEnv("JWT_REFRESH_TTL", "168h")),
		},
		OAuth: OAuthConfig{
			GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
			AppleClientID:  getEnv("APPLE_CLIENT_ID", ""),
		},
		Selcom: SelcomConfig{
			BaseURL:        getEnv("SELCOM_BASE_URL", "https://apigw.selcommobile.com"),
			APIKey:         getEnv("SELCOM_API_KEY", ""),
			APISecret:      getEnv("SELCOM_API_SECRET", ""),
			Vendor:         getEnv("SELCOM_VENDOR", ""),
			WebhookSecret:  getEnv("SELCOM_WEBHOOK_SECRET", ""),
			PublicAPIURL:   getEnv("PUBLIC_API_URL", "http://localhost:8080"),
			PublicAppURL:   getEnv("PUBLIC_APP_URL", "http://localhost:3000"),
			TimeoutSeconds: parseInt(getEnv("SELCOM_TIMEOUT_SECONDS", "30")),
			MockMode:       getEnv("SELCOM_MOCK", "") == "true",
		},
		Kkooapp: KkooappConfig{
			BaseURL:        getEnv("KKOOAPP_BASE_URL", "https://api.kkooapp.com/v1"),
			APIKey:         getEnv("KKOOAPP_API_KEY", ""),
			TimeoutSeconds: parseInt(getEnv("KKOOAPP_TIMEOUT_SECONDS", "30")),
			Endpoints: KkooappEndpoints{
				HotelReservation: getEnv("KKOOAPP_ENDPOINT_HOTEL_RESERVATION", "/reservations/hotel"),
				TableReservation: getEnv("KKOOAPP_ENDPOINT_TABLE_RESERVATION", "/reservations/table"),
				FoodOrder:        getEnv("KKOOAPP_ENDPOINT_FOOD_ORDER", "/orders/food"),
				OrderStatus:      getEnv("KKOOAPP_ENDPOINT_ORDER_STATUS", "/orders/{id}/status"),
			},
		},
		Webhook: WebhookConfig{
			InboundSecret: getEnv("WEBHOOK_INBOUND_SECRET", ""),
		},
	}
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" port=" + d.Port +
		" sslmode=" + d.SSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
