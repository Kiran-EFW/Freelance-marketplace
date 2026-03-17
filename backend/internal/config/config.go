package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application.
type Config struct {
	// Server
	ServerPort  string
	Environment string // dev, staging, prod

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret string
	JWTExpiry time.Duration

	// SMS
	SMSProvider string
	SMSAPIKey   string

	// Twilio
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Payment - Razorpay
	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string

	// Payment - Stripe
	StripeAPIKey string

	// Payment
	PaymentProvider string

	// SMTP / Email
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// WhatsApp Business API (Meta Cloud API)
	WhatsAppPhoneNumberID      string
	WhatsAppAccessToken        string
	WhatsAppBusinessID         string
	WhatsAppWebhookVerifyToken string

	// FCM Push Notifications
	FCMProjectID             string
	FCMServiceAccountKeyPath string // path to the service account JSON key file
	FCMServiceAccountKey     string // inline JSON key (alternative to file path)

	// Cloudflare R2 Object Storage
	R2AccountID       string
	R2AccessKeyID     string
	R2AccessKeySecret string
	R2Bucket          string
	R2PublicURL       string // public base URL for serving uploaded files

	// Meilisearch
	MeilisearchURL string
	MeilisearchKey string

	// Database Migrations
	MigrationsPath string // path to migration files
	AutoMigrate    bool   // run migrations on startup
}

// Load reads configuration from environment variables with sensible defaults.
// It attempts to load a .env file first but does not fail if one is not found.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("no .env file found, reading configuration from environment")
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "dev"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/seva?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry:     jwtExpiry,
		SMSProvider:           getEnv("SMS_PROVIDER", "twilio"),
		SMSAPIKey:             getEnv("SMS_API_KEY", ""),
		TwilioAccountSID:      getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:       getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber:      getEnv("TWILIO_FROM_NUMBER", ""),
		RazorpayKeyID:         getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:     getEnv("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		StripeAPIKey:          getEnv("STRIPE_API_KEY", ""),
		PaymentProvider:       getEnv("PAYMENT_PROVIDER", "razorpay"),
		SMTPHost:          getEnv("SMTP_HOST", "localhost"),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@seva.in"),
		WhatsAppPhoneNumberID:      getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppAccessToken:        getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppBusinessID:         getEnv("WHATSAPP_BUSINESS_ID", ""),
		WhatsAppWebhookVerifyToken: getEnv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "seva-webhook-verify"),
		FCMProjectID:             getEnv("FCM_PROJECT_ID", ""),
		FCMServiceAccountKeyPath: getEnv("FCM_SERVICE_ACCOUNT_KEY_PATH", ""),
		FCMServiceAccountKey:     getEnv("FCM_SERVICE_ACCOUNT_KEY", ""),
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2AccessKeySecret: getEnv("R2_ACCESS_KEY_SECRET", ""),
		R2Bucket:          getEnv("R2_BUCKET", "seva-uploads"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		MeilisearchURL:    getEnv("MEILISEARCH_URL", ""),
		MeilisearchKey:    getEnv("MEILISEARCH_KEY", ""),
		MigrationsPath:    getEnv("MIGRATIONS_PATH", "migrations"),
		AutoMigrate:       getEnv("AUTO_MIGRATE", "true") == "true",
	}

	return cfg, nil
}

// IsProd returns true when running in the production environment.
func (c *Config) IsProd() bool {
	return c.Environment == "prod"
}

// RateLimitMax returns the per-second rate limit based on environment.
func (c *Config) RateLimitMax() int {
	v := getEnv("RATE_LIMIT_MAX", "")
	if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	if c.IsProd() {
		return 30
	}
	return 100
}

// GetFCMServiceAccountKey returns the FCM service account key JSON bytes.
// It first checks for an inline key, then falls back to reading from file path.
// Returns nil if neither is configured.
func (c *Config) GetFCMServiceAccountKey() []byte {
	if c.FCMServiceAccountKey != "" {
		return []byte(c.FCMServiceAccountKey)
	}
	if c.FCMServiceAccountKeyPath != "" {
		data, err := os.ReadFile(c.FCMServiceAccountKeyPath)
		if err != nil {
			log.Error().Err(err).Str("path", c.FCMServiceAccountKeyPath).Msg("failed to read FCM service account key file")
			return nil
		}
		return data
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
