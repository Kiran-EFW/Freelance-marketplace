package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ModerationStatus represents the moderation state of user-generated content.
type ModerationStatus string

const (
	ModerationPending  ModerationStatus = "pending"
	ModerationApproved ModerationStatus = "approved"
	ModerationFlagged  ModerationStatus = "flagged"
	ModerationRemoved  ModerationStatus = "removed"
)

// Review represents a rating and comment left after a job is completed.
// Reviews are bidirectional: both customers and providers can review each other.
type Review struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	JobID            uuid.UUID        `json:"job_id" db:"job_id"`
	ReviewerID       uuid.UUID        `json:"reviewer_id" db:"reviewer_id"`
	RevieweeID       uuid.UUID        `json:"reviewee_id" db:"reviewee_id"`
	Rating           int              `json:"rating" db:"rating"` // 1-5
	Comment          string           `json:"comment,omitempty" db:"comment"`
	Language         string           `json:"language" db:"language"`
	ModerationStatus ModerationStatus `json:"moderation_status" db:"moderation_status"`
	ProviderResponse *string          `json:"provider_response,omitempty" db:"response"`
	RespondedAt      *time.Time       `json:"responded_at,omitempty" db:"responded_at"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// ProviderReviewStats aggregates review metrics for a provider.
type ProviderReviewStats struct {
	ProviderID   uuid.UUID      `json:"provider_id"`
	AvgRating    float64        `json:"avg_rating"`
	TotalReviews int            `json:"total_reviews"`
	Distribution map[int]int    `json:"distribution"` // rating -> count (1-5)
}

// TrustScoreFactors holds the individual components that make up a provider's
// trust score. Each factor is a value between 0 and 1.
type TrustScoreFactors struct {
	OnTimeRate     float64 `json:"on_time_rate"`
	CompletionRate float64 `json:"completion_rate"`
	AvgRating      float64 `json:"avg_rating"`      // normalised to 0-1
	DisputeRate    float64 `json:"dispute_rate"`     // inverted: lower is better
	ResponseTime   float64 `json:"response_time"`    // normalised to 0-1
}

// ReviewRepository defines persistence operations for reviews.
type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	GetByJobAndReviewer(ctx context.Context, jobID, reviewerID uuid.UUID) (*Review, error)
	ListByReviewee(ctx context.Context, revieweeID uuid.UUID, limit, offset int) ([]Review, error)
	ListByJob(ctx context.Context, jobID uuid.UUID) ([]Review, error)
	UpdateResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error
	UpdateModerationStatus(ctx context.Context, id uuid.UUID, status ModerationStatus) error
	GetProviderStats(ctx context.Context, providerID uuid.UUID) (*ProviderReviewStats, error)
	GetAverageRating(ctx context.Context, userID uuid.UUID) (float64, int, error)
}
