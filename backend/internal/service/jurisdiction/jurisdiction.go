package jurisdiction

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/repository/postgres"
	rediscache "github.com/seva-platform/backend/internal/repository/redis"
)

const (
	// cacheTTL is how long jurisdiction configs are cached in Redis.
	cacheTTL = 1 * time.Hour

	// cacheKeyPrefix is the Redis key prefix for jurisdiction configs.
	cacheKeyPrefix = "jurisdiction:"
)

// JurisdictionConfig holds the parsed JSONB config for a jurisdiction.
type JurisdictionConfig struct {
	CommissionRates      map[string]float64 `json:"commission_rates"`
	CommissionThresholds map[string]float64 `json:"commission_thresholds"`
	SubscriptionPrices   map[string]float64 `json:"subscription_prices"`
	LeadFees             map[string]float64 `json:"lead_fees"`
	UrgentBookingFee     float64            `json:"urgent_booking_fee"`
	PaymentMethods       []string           `json:"payment_methods"`
	DefaultPaymentGateway string            `json:"default_payment_gateway"`
	IDVerification       string             `json:"id_verification"`
	PhoneFormat          string             `json:"phone_format"`
	PostcodeFormat       string             `json:"postcode_format"`
	SupportedLanguages   []string           `json:"supported_languages"`
	CategoriesEnabled    []string           `json:"service_categories_enabled"`
	SeasonalCalendar     bool               `json:"seasonal_calendar_enabled"`
	SMSTemplates         map[string]string  `json:"sms_templates"`
}

// JurisdictionInfo is the public-facing jurisdiction details (without raw JSONB).
type JurisdictionInfo struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	DefaultLanguage string             `json:"default_language"`
	Currency        string             `json:"currency"`
	CurrencySymbol  string             `json:"currency_symbol"`
	PhonePrefix     string             `json:"phone_prefix"`
	Timezone        string             `json:"timezone"`
	IsActive        bool               `json:"is_active"`
	Config          *JurisdictionConfig `json:"config,omitempty"`
}

// phonePrefixes maps known phone prefixes to jurisdiction IDs for auto-detection.
var phonePrefixes = []struct {
	Prefix       string
	Jurisdiction string
}{
	{"+91", "in"},
	{"+44", "uk"},
	{"+1", "us"},
	{"+49", "de"},
	{"+33", "fr"},
}

// JurisdictionService provides business logic for jurisdiction configuration.
type JurisdictionService struct {
	queries *postgres.Queries
	cache   *rediscache.CacheStore
}

// NewJurisdictionService returns a ready-to-use JurisdictionService.
func NewJurisdictionService(queries *postgres.Queries, cache *rediscache.CacheStore) *JurisdictionService {
	return &JurisdictionService{
		queries: queries,
		cache:   cache,
	}
}

// GetConfig returns the full config for a jurisdiction, cached for 1 hour.
func (s *JurisdictionService) GetConfig(ctx context.Context, jurisdictionID string) (*JurisdictionConfig, error) {
	jurisdictionID = strings.ToLower(jurisdictionID)
	cacheKey := cacheKeyPrefix + jurisdictionID

	// Try cache first.
	var cfg JurisdictionConfig
	if err := s.cache.GetJSON(ctx, cacheKey, &cfg); err == nil {
		return &cfg, nil
	}

	// Fetch from database.
	j, err := s.queries.GetJurisdiction(ctx, jurisdictionID)
	if err != nil {
		return nil, fmt.Errorf("get jurisdiction %s: %w", jurisdictionID, err)
	}

	if err := json.Unmarshal(j.Config, &cfg); err != nil {
		return nil, fmt.Errorf("parse jurisdiction config for %s: %w", jurisdictionID, err)
	}

	// Cache the parsed config.
	if err := s.cache.SetJSON(ctx, cacheKey, &cfg, cacheTTL); err != nil {
		log.Warn().Err(err).Str("jurisdiction", jurisdictionID).Msg("failed to cache jurisdiction config")
	}

	return &cfg, nil
}

// GetJurisdiction returns the full jurisdiction information including parsed config.
func (s *JurisdictionService) GetJurisdiction(ctx context.Context, jurisdictionID string) (*JurisdictionInfo, error) {
	jurisdictionID = strings.ToLower(jurisdictionID)

	j, err := s.queries.GetJurisdiction(ctx, jurisdictionID)
	if err != nil {
		return nil, fmt.Errorf("get jurisdiction %s: %w", jurisdictionID, err)
	}

	var cfg JurisdictionConfig
	if err := json.Unmarshal(j.Config, &cfg); err != nil {
		return nil, fmt.Errorf("parse jurisdiction config for %s: %w", jurisdictionID, err)
	}

	return &JurisdictionInfo{
		ID:              j.ID,
		Name:            j.Name,
		DefaultLanguage: j.DefaultLanguage,
		Currency:        j.Currency,
		CurrencySymbol:  j.CurrencySymbol,
		PhonePrefix:     j.PhonePrefix,
		Timezone:        j.Timezone,
		IsActive:        j.IsActive,
		Config:          &cfg,
	}, nil
}

// ListActive returns all active jurisdictions.
func (s *JurisdictionService) ListActive(ctx context.Context) ([]JurisdictionInfo, error) {
	rows, err := s.queries.ListActiveJurisdictions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active jurisdictions: %w", err)
	}

	result := make([]JurisdictionInfo, 0, len(rows))
	for _, j := range rows {
		info := JurisdictionInfo{
			ID:              j.ID,
			Name:            j.Name,
			DefaultLanguage: j.DefaultLanguage,
			Currency:        j.Currency,
			CurrencySymbol:  j.CurrencySymbol,
			PhonePrefix:     j.PhonePrefix,
			Timezone:        j.Timezone,
			IsActive:        j.IsActive,
		}

		var cfg JurisdictionConfig
		if err := json.Unmarshal(j.Config, &cfg); err == nil {
			info.Config = &cfg
		}

		result = append(result, info)
	}

	return result, nil
}

// GetCommissionRate returns the commission rate based on job amount and jurisdiction.
func (s *JurisdictionService) GetCommissionRate(ctx context.Context, jurisdictionID string, amount float64) (float64, error) {
	cfg, err := s.GetConfig(ctx, jurisdictionID)
	if err != nil {
		return 0, err
	}

	thresholds := cfg.CommissionThresholds
	rates := cfg.CommissionRates

	// Default to highest rate if thresholds are not configured.
	if len(thresholds) == 0 || len(rates) == 0 {
		if r, ok := rates["mid"]; ok {
			return r, nil
		}
		return 0.05, nil // fallback default
	}

	midMax, hasMidMax := thresholds["mid_max"]
	lowMax, hasLowMax := thresholds["low_max"]

	switch {
	case hasMidMax && amount <= midMax:
		if r, ok := rates["low"]; ok {
			return r, nil
		}
	case hasLowMax && amount <= lowMax:
		if r, ok := rates["mid"]; ok {
			return r, nil
		}
	default:
		if r, ok := rates["high"]; ok {
			return r, nil
		}
	}

	// Fallback to mid rate.
	if r, ok := rates["mid"]; ok {
		return r, nil
	}
	return 0.05, nil
}

// GetSubscriptionPrice returns the subscription price for a tier in the jurisdiction's
// currency. It returns (price, currency, error).
func (s *JurisdictionService) GetSubscriptionPrice(ctx context.Context, jurisdictionID, tier string) (float64, string, error) {
	cfg, err := s.GetConfig(ctx, jurisdictionID)
	if err != nil {
		return 0, "", err
	}

	jurisdictionID = strings.ToLower(jurisdictionID)

	j, jErr := s.queries.GetJurisdiction(ctx, jurisdictionID)
	if jErr != nil {
		return 0, "", fmt.Errorf("get jurisdiction for currency: %w", jErr)
	}

	price, ok := cfg.SubscriptionPrices[tier]
	if !ok {
		return 0, "", fmt.Errorf("subscription tier %q not found for jurisdiction %s", tier, jurisdictionID)
	}

	return price, j.Currency, nil
}

// ValidatePhone validates a phone number against the jurisdiction's format.
func (s *JurisdictionService) ValidatePhone(ctx context.Context, jurisdictionID, phone string) (bool, error) {
	cfg, err := s.GetConfig(ctx, jurisdictionID)
	if err != nil {
		return false, err
	}

	if cfg.PhoneFormat == "" {
		// No format defined; accept any phone number.
		return true, nil
	}

	re, err := regexp.Compile(cfg.PhoneFormat)
	if err != nil {
		log.Error().Err(err).Str("jurisdiction", jurisdictionID).Str("pattern", cfg.PhoneFormat).Msg("invalid phone format regex")
		return false, fmt.Errorf("invalid phone format pattern for %s: %w", jurisdictionID, err)
	}

	return re.MatchString(phone), nil
}

// GetPaymentGateway returns the default payment gateway for a jurisdiction.
func (s *JurisdictionService) GetPaymentGateway(ctx context.Context, jurisdictionID string) (string, error) {
	cfg, err := s.GetConfig(ctx, jurisdictionID)
	if err != nil {
		return "", err
	}

	if cfg.DefaultPaymentGateway == "" {
		return "stripe", nil // fallback default
	}

	return cfg.DefaultPaymentGateway, nil
}

// GetEnabledCategories returns the list of enabled service categories for a jurisdiction.
func (s *JurisdictionService) GetEnabledCategories(ctx context.Context, jurisdictionID string) ([]string, error) {
	cfg, err := s.GetConfig(ctx, jurisdictionID)
	if err != nil {
		return nil, err
	}

	return cfg.CategoriesEnabled, nil
}

// DetectJurisdiction detects jurisdiction from phone prefix.
// It returns an empty string if no match is found.
func (s *JurisdictionService) DetectJurisdiction(phone string) string {
	phone = strings.TrimSpace(phone)

	// Match from longest prefix to shortest to avoid "+1" matching "+1x" prefixes.
	// The phonePrefixes slice is ordered with longer prefixes first where needed.
	for _, pp := range phonePrefixes {
		if strings.HasPrefix(phone, pp.Prefix) {
			return pp.Jurisdiction
		}
	}

	return ""
}
