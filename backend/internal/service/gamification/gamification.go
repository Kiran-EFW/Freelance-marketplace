package gamification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// Point values for each action.
const (
	PointsForJobCompleted      = 10
	PointsForReviewGiven       = 3
	PointsForFiveStarReview    = 5
	PointsForReferralProvider  = 20
	PointsForReferralCustomer  = 10
	PointsForFastResponse      = 2
	PointsForEarlyAdopter      = 50
	PointsForProviderOnboarded = 15
	PointsForJobPosted         = 1
)

// predefined levels ordered by progression.
var levels = []domain.UserLevel{
	{Level: 1, Name: "New", MinJobs: 0, MinRating: 0.0},
	{Level: 2, Name: "Active", MinJobs: 5, MinRating: 3.0},
	{Level: 3, Name: "Trusted", MinJobs: 20, MinRating: 3.5},
	{Level: 4, Name: "Expert", MinJobs: 50, MinRating: 4.0},
	{Level: 5, Name: "Local Champion", MinJobs: 100, MinRating: 4.5},
}

// Service defines the gamification service interface.
type Service interface {
	AwardPoints(ctx context.Context, userID uuid.UUID, reason domain.PointsReason, referenceID *uuid.UUID) error
	GetBalance(ctx context.Context, userID uuid.UUID) (int, error)
	GetLevel(ctx context.Context, userID uuid.UUID) (*domain.UserLevel, error)
	SpendPoints(ctx context.Context, userID uuid.UUID, amount int, reason string) error
	GetLeaderboard(ctx context.Context, postcode string, limit int) ([]domain.LeaderboardEntry, error)
}

// GamificationService implements points, levels, and achievements.
type GamificationService struct {
	gamification domain.GamificationRepository
	providers    domain.ProviderRepository
	reviews      domain.ReviewRepository
}

// NewGamificationService returns a ready-to-use GamificationService.
func NewGamificationService(
	gamification domain.GamificationRepository,
	providers domain.ProviderRepository,
	reviews domain.ReviewRepository,
) *GamificationService {
	return &GamificationService{
		gamification: gamification,
		providers:    providers,
		reviews:      reviews,
	}
}

// AwardPoints credits the appropriate number of points to a user based on
// the reason.
func (s *GamificationService) AwardPoints(ctx context.Context, userID uuid.UUID, reason domain.PointsReason, referenceID *uuid.UUID) error {
	points := pointsForReason(reason)
	if points == 0 {
		return fmt.Errorf("%w: unknown points reason %s", domain.ErrInvalidInput, reason)
	}

	balance, err := s.gamification.GetBalance(ctx, userID)
	if err != nil {
		balance = 0
	}

	newBalance := balance + points

	entry := &domain.PointsEntry{
		ID:            uuid.New(),
		UserID:        userID,
		Points:        points,
		Reason:        reason,
		ReferenceType: string(reason),
		ReferenceID:   referenceID,
		Balance:       newBalance,
	}

	if err := s.gamification.AddEntry(ctx, entry); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to award points")
		return fmt.Errorf("award points: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("points", points).
		Str("reason", string(reason)).
		Int("new_balance", newBalance).
		Msg("points awarded")

	return nil
}

// GetBalance returns the user's current points balance.
func (s *GamificationService) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	balance, err := s.gamification.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}

// GetLevel calculates the user's current level based on completed jobs and
// average rating.
func (s *GamificationService) GetLevel(ctx context.Context, userID uuid.UUID) (*domain.UserLevel, error) {
	provider, err := s.providers.GetByID(ctx, userID)
	if err != nil {
		// Non-provider users default to level 1.
		return &levels[0], nil
	}

	avgRating, _, err := s.reviews.GetAverageRating(ctx, userID)
	if err != nil {
		avgRating = 0
	}

	// Find the highest level the provider qualifies for.
	result := &levels[0]
	for i := len(levels) - 1; i >= 0; i-- {
		if provider.JobsCompleted >= levels[i].MinJobs && avgRating >= levels[i].MinRating {
			result = &levels[i]
			break
		}
	}

	return result, nil
}

// SpendPoints deducts points from a user's balance (e.g., to redeem a reward).
func (s *GamificationService) SpendPoints(ctx context.Context, userID uuid.UUID, amount int, reason string) error {
	if amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}

	balance, err := s.gamification.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("%w: insufficient points (have %d, need %d)", domain.ErrInvalidInput, balance, amount)
	}

	newBalance := balance - amount

	entry := &domain.PointsEntry{
		ID:            uuid.New(),
		UserID:        userID,
		Points:        -amount,
		Reason:        domain.PointsSpent,
		ReferenceType: reason,
		Balance:       newBalance,
	}

	if err := s.gamification.AddEntry(ctx, entry); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to spend points")
		return fmt.Errorf("spend points: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("spent", amount).
		Int("new_balance", newBalance).
		Msg("points spent")

	return nil
}

// GetLeaderboard returns the top users by points within a postcode area.
func (s *GamificationService) GetLeaderboard(ctx context.Context, postcode string, limit int) ([]domain.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	entries, err := s.gamification.GetLeaderboard(ctx, postcode, limit)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	return entries, nil
}

// pointsForReason maps a reason to its point value.
func pointsForReason(reason domain.PointsReason) int {
	switch reason {
	case domain.PointsJobCompleted:
		return PointsForJobCompleted
	case domain.PointsReviewGiven:
		return PointsForReviewGiven
	case domain.PointsFiveStarReview:
		return PointsForFiveStarReview
	case domain.PointsReferralProvider:
		return PointsForReferralProvider
	case domain.PointsReferralCustomer:
		return PointsForReferralCustomer
	case domain.PointsFastResponse:
		return PointsForFastResponse
	case domain.PointsEarlyAdopter:
		return PointsForEarlyAdopter
	case domain.PointsProviderOnboarded:
		return PointsForProviderOnboarded
	case domain.PointsJobPosted:
		return PointsForJobPosted
	default:
		return 0
	}
}
