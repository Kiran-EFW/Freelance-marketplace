package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserType represents the kind of account a user holds.
type UserType string

const (
	UserTypeCustomer UserType = "customer"
	UserTypeProvider UserType = "provider"
	UserTypeCompany  UserType = "company"
)

// User is the core identity in the marketplace.
type User struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Type              UserType  `json:"type" db:"type"`
	Phone             string    `json:"phone" db:"phone"`
	Email             *string   `json:"email,omitempty" db:"email"`
	Name              string    `json:"name" db:"name"`
	JurisdictionID    string    `json:"jurisdiction_id" db:"jurisdiction_id"`
	PreferredLanguage string    `json:"preferred_language" db:"preferred_language"`
	DeviceType        string    `json:"device_type,omitempty" db:"device_type"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// VerificationStatus tracks the provider's identity verification.
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationApproved VerificationStatus = "approved"
	VerificationRejected VerificationStatus = "rejected"
)

// SubscriptionTier represents the provider's current subscription level.
type SubscriptionTier string

const (
	SubscriptionFree    SubscriptionTier = "free"
	SubscriptionBasic   SubscriptionTier = "basic"
	SubscriptionPremium SubscriptionTier = "premium"
)

// ProviderProfile extends a user account with provider-specific fields.
type ProviderProfile struct {
	UserID               uuid.UUID          `json:"user_id" db:"user_id"`
	Skills               []string           `json:"skills" db:"skills"`
	ServiceRadiusKM      float64            `json:"service_radius_km" db:"service_radius_km"`
	Postcode             string             `json:"postcode" db:"postcode"`
	Latitude             float64            `json:"latitude" db:"latitude"`
	Longitude            float64            `json:"longitude" db:"longitude"`
	TrustScore           float64            `json:"trust_score" db:"trust_score"`
	Level                int                `json:"level" db:"level"`
	VerificationStatus   VerificationStatus `json:"verification_status" db:"verification_status"`
	SubscriptionTier     SubscriptionTier   `json:"subscription_tier" db:"subscription_tier"`
	AvailabilitySchedule json.RawMessage    `json:"availability_schedule,omitempty" db:"availability_schedule"`
	BankAccountID        *string            `json:"bank_account_id,omitempty" db:"bank_account_id"`
	JobsCompleted        int                `json:"jobs_completed" db:"total_jobs_completed"`
	ResponseTimeAvg      int                `json:"response_time_avg" db:"avg_response_time_minutes"`
	Bio                  string             `json:"bio" db:"description"`
	IsAvailable          bool               `json:"is_available" db:"is_available"`
	CreatedAt            time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at" db:"updated_at"`
}

// ProviderSearchFilters holds optional filters for searching providers.
type ProviderSearchFilters struct {
	CategoryID       *uuid.UUID
	CategorySlug     *string
	Postcode         *string
	Latitude         *float64
	Longitude        *float64
	RadiusKM         *float64
	MinRating        *float64
	MinTrustScore    *float64
	VerificationOnly bool
	Skills           []string
	Available        *bool
	SortBy           string // "distance", "rating", "trust_score", "response_time"
	Limit            int
	Offset           int
}

// ProviderSearchResult combines a provider profile with contextual search data.
type ProviderSearchResult struct {
	ProviderProfile
	UserName  string  `json:"user_name"`
	UserPhone string  `json:"user_phone"`
	Distance  float64 `json:"distance_km"`
	AvgRating float64 `json:"avg_rating"`
}

// UpdateProfileParams holds the fields that can be updated on a user profile.
type UpdateProfileParams struct {
	Name              *string
	Email             *string
	PreferredLanguage *string
	DeviceType        *string
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProviderRepository defines persistence operations for provider profiles.
type ProviderRepository interface {
	Create(ctx context.Context, profile *ProviderProfile) error
	GetByID(ctx context.Context, userID uuid.UUID) (*ProviderProfile, error)
	Update(ctx context.Context, profile *ProviderProfile) error
	Delete(ctx context.Context, userID uuid.UUID) error
	Search(ctx context.Context, filters ProviderSearchFilters) ([]ProviderSearchResult, error)
	ListByPostcode(ctx context.Context, postcode string, limit, offset int) ([]ProviderProfile, error)
	UpdateTrustScore(ctx context.Context, userID uuid.UUID, score float64) error
	IncrementJobsCompleted(ctx context.Context, userID uuid.UUID) error
}
