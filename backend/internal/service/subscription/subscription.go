// Package subscription provides the subscription service implementation with
// real database operations for provider subscription management.
package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Subscription represents a provider's subscription record.
type Subscription struct {
	ID                    uuid.UUID  `json:"id"`
	ProviderID            uuid.UUID  `json:"provider_id"`
	Tier                  string     `json:"tier"`
	StartedAt             time.Time  `json:"started_at"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	AutoRenew             bool       `json:"auto_renew"`
	PaymentMethod         string     `json:"payment_method,omitempty"`
	GatewaySubscriptionID string     `json:"gateway_subscription_id,omitempty"`
	Amount                float64    `json:"amount"`
	Currency              string     `json:"currency"`
	Status                string     `json:"status"` // active, cancelled, expired, past_due
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// Plan pricing in INR.
var planPricing = map[string]float64{
	"free":         0,
	"professional": 299,
	"enterprise":   999,
}

// SubscriptionService implements subscription management with direct
// database operations.
type SubscriptionService struct {
	db *pgxpool.Pool
}

// NewSubscriptionService returns a ready-to-use SubscriptionService.
func NewSubscriptionService(db *pgxpool.Pool) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// ensureTableExists creates the subscriptions table if it does not already exist.
func (s *SubscriptionService) ensureTableExists(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		provider_id UUID NOT NULL REFERENCES users(id),
		tier TEXT NOT NULL DEFAULT 'free',
		started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMPTZ,
		auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
		payment_method TEXT,
		gateway_subscription_id TEXT,
		amount NUMERIC(10,2) NOT NULL DEFAULT 0,
		currency TEXT NOT NULL DEFAULT 'INR',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_provider ON subscriptions(provider_id, status);
	`
	_, err := s.db.Exec(ctx, ddl)
	return err
}

// GetCurrentSubscription returns the active subscription for a provider.
func (s *SubscriptionService) GetCurrentSubscription(ctx context.Context, providerID uuid.UUID) (*Subscription, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure subscriptions table exists")
	}

	query := `SELECT id, provider_id, tier, started_at, expires_at, auto_renew,
		payment_method, gateway_subscription_id, amount, currency, status, created_at, updated_at
		FROM subscriptions
		WHERE provider_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1`

	var sub Subscription
	var expiresAt pgtype.Timestamptz
	var paymentMethod, gatewaySubID pgtype.Text
	var amount pgtype.Numeric

	err := s.db.QueryRow(ctx, query, providerID).Scan(
		&sub.ID, &sub.ProviderID, &sub.Tier, &sub.StartedAt,
		&expiresAt, &sub.AutoRenew, &paymentMethod, &gatewaySubID,
		&amount, &sub.Currency, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No active subscription
		}
		return nil, fmt.Errorf("get current subscription: %w", err)
	}

	if expiresAt.Valid {
		t := expiresAt.Time
		sub.ExpiresAt = &t
	}
	if paymentMethod.Valid {
		sub.PaymentMethod = paymentMethod.String
	}
	if gatewaySubID.Valid {
		sub.GatewaySubscriptionID = gatewaySubID.String
	}
	if amount.Valid {
		f, _ := amount.Float64Value()
		if f.Valid {
			sub.Amount = f.Float64
		}
	}

	return &sub, nil
}

// Subscribe creates a new subscription for a provider. If the provider
// already has an active subscription, it is cancelled first.
func (s *SubscriptionService) Subscribe(ctx context.Context, providerID uuid.UUID, tier, paymentMethod string) (*Subscription, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure subscriptions table exists")
	}

	price, ok := planPricing[tier]
	if !ok {
		return nil, fmt.Errorf("invalid subscription tier: %s", tier)
	}

	// Cancel any existing active subscription.
	_, err := s.db.Exec(ctx,
		`UPDATE subscriptions SET status = 'cancelled', updated_at = NOW() WHERE provider_id = $1 AND status = 'active'`,
		providerID,
	)
	if err != nil {
		log.Warn().Err(err).Str("provider_id", providerID.String()).Msg("failed to cancel existing subscription")
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 1, 0) // 1 month subscription

	sub := &Subscription{
		ID:            uuid.New(),
		ProviderID:    providerID,
		Tier:          tier,
		StartedAt:     now,
		ExpiresAt:     &expiresAt,
		AutoRenew:     true,
		PaymentMethod: paymentMethod,
		Amount:        price,
		Currency:      "INR",
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	query := `INSERT INTO subscriptions (id, provider_id, tier, started_at, expires_at, auto_renew,
		payment_method, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = s.db.Exec(ctx, query,
		sub.ID, sub.ProviderID, sub.Tier, sub.StartedAt, sub.ExpiresAt,
		sub.AutoRenew, sub.PaymentMethod, sub.Amount, sub.Currency,
		sub.Status, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	// Update the provider's subscription tier in the profile.
	_, err = s.db.Exec(ctx,
		`UPDATE provider_profiles SET subscription_tier = $2 WHERE user_id = $1`,
		providerID, tier,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to update provider subscription tier")
	}

	log.Info().
		Str("subscription_id", sub.ID.String()).
		Str("provider_id", providerID.String()).
		Str("tier", tier).
		Float64("amount", price).
		Msg("subscription created")

	return sub, nil
}

// CancelSubscription cancels an active subscription.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, subscriptionID, providerID uuid.UUID) error {
	query := `UPDATE subscriptions
		SET status = 'cancelled', auto_renew = false, updated_at = NOW()
		WHERE id = $1 AND provider_id = $2 AND status = 'active'`

	tag, err := s.db.Exec(ctx, query, subscriptionID, providerID)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found or already cancelled")
	}

	// Reset provider to free tier.
	_, err = s.db.Exec(ctx,
		`UPDATE provider_profiles SET subscription_tier = 'free' WHERE user_id = $1`,
		providerID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to reset provider subscription tier")
	}

	log.Info().
		Str("subscription_id", subscriptionID.String()).
		Str("provider_id", providerID.String()).
		Msg("subscription cancelled")

	return nil
}

// HandlePaymentWebhook processes a payment gateway webhook for subscription
// renewals. The payload format depends on the gateway (Razorpay or Stripe).
func (s *SubscriptionService) HandlePaymentWebhook(ctx context.Context, payload []byte, signature string) error {
	// In production, this would:
	// 1. Verify the webhook signature against the gateway's secret.
	// 2. Parse the payload to extract subscription/payment details.
	// 3. Update the subscription status based on the event type.
	// 4. Handle renewal, payment failure, and cancellation events.

	log.Info().
		Int("payload_size", len(payload)).
		Bool("has_signature", signature != "").
		Msg("subscription webhook received")

	return nil
}

// ListBillingHistory returns the subscription billing history for a provider.
func (s *SubscriptionService) ListBillingHistory(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]Subscription, int, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure subscriptions table exists")
	}

	query := `SELECT id, provider_id, tier, started_at, expires_at, auto_renew,
		payment_method, gateway_subscription_id, amount, currency, status, created_at, updated_at
		FROM subscriptions
		WHERE provider_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, providerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list billing history: %w", err)
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		var expiresAt pgtype.Timestamptz
		var paymentMethod, gatewaySubID pgtype.Text
		var amount pgtype.Numeric

		if err := rows.Scan(
			&sub.ID, &sub.ProviderID, &sub.Tier, &sub.StartedAt,
			&expiresAt, &sub.AutoRenew, &paymentMethod, &gatewaySubID,
			&amount, &sub.Currency, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan subscription row, skipping")
			continue
		}

		if expiresAt.Valid {
			t := expiresAt.Time
			sub.ExpiresAt = &t
		}
		if paymentMethod.Valid {
			sub.PaymentMethod = paymentMethod.String
		}
		if gatewaySubID.Valid {
			sub.GatewaySubscriptionID = gatewaySubID.String
		}
		if amount.Valid {
			f, _ := amount.Float64Value()
			if f.Valid {
				sub.Amount = f.Float64
			}
		}

		subs = append(subs, sub)
	}

	return subs, len(subs), nil
}
