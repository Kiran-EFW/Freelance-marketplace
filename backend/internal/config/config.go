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

	// Object Storage
	StorageBucket string
	StorageRegion string
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
		SMSProvider:   getEnv("SMS_PROVIDER", "twilio"),
		SMSAPIKey:     getEnv("SMS_API_KEY", ""),
		StorageBucket: getEnv("STORAGE_BUCKET", "seva-uploads"),
		StorageRegion: getEnv("STORAGE_REGION", "ap-south-1"),
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

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
