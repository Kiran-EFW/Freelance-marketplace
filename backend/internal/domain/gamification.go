package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PointsReason identifies why points were awarded or deducted.
type PointsReason string

const (
	PointsJobCompleted      PointsReason = "job_completed"
	PointsReviewGiven       PointsReason = "review_given"
	PointsFiveStarReview    PointsReason = "five_star_review"
	PointsReferralProvider  PointsReason = "referral_provider"
	PointsReferralCustomer  PointsReason = "referral_customer"
	PointsFastResponse      PointsReason = "fast_response"
	PointsEarlyAdopter      PointsReason = "early_adopter"
	PointsProviderOnboarded PointsReason = "provider_onboarded"
	PointsJobPosted         PointsReason = "job_posted"
	PointsSpent             PointsReason = "points_spent"
)

// PointsEntry records a single credit or debit in the user's points ledger.
type PointsEntry struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	UserID        uuid.UUID    `json:"user_id" db:"user_id"`
	Points        int          `json:"points" db:"points"`
	Reason        PointsReason `json:"reason" db:"reason"`
	ReferenceType string       `json:"reference_type,omitempty" db:"reference_type"`
	ReferenceID   *uuid.UUID   `json:"reference_id,omitempty" db:"reference_id"`
	Balance       int          `json:"balance_after" db:"balance_after"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
}

// UserLevel defines the requirements and label for each provider level.
type UserLevel struct {
	Level     int     `json:"level"`
	Name      string  `json:"name"`
	MinJobs   int     `json:"min_jobs"`
	MinRating float64 `json:"min_rating"`
}

// Achievement represents a one-time accomplishment earned by a user.
type Achievement struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	AchievementType string    `json:"achievement_type" db:"achievement_type"`
	EarnedAt        time.Time `json:"earned_at" db:"earned_at"`
}

// LeaderboardEntry represents a user's rank on the points leaderboard.
type LeaderboardEntry struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Points   int       `json:"points"`
	Rank     int       `json:"rank"`
	Postcode string    `json:"postcode"`
}

// GamificationRepository defines persistence operations for points, levels,
// and achievements.
type GamificationRepository interface {
	AddEntry(ctx context.Context, entry *PointsEntry) error
	GetBalance(ctx context.Context, userID uuid.UUID) (int, error)
	ListEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]PointsEntry, error)

	CreateAchievement(ctx context.Context, achievement *Achievement) error
	ListAchievements(ctx context.Context, userID uuid.UUID) ([]Achievement, error)
	HasAchievement(ctx context.Context, userID uuid.UUID, achievementType string) (bool, error)

	GetLeaderboard(ctx context.Context, postcode string, limit int) ([]LeaderboardEntry, error)
}
