package review

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// Trust score weights — must sum to 1.0.
const (
	trustWeightOnTime      = 0.25
	trustWeightCompletion  = 0.25
	trustWeightRating      = 0.30
	trustWeightDispute     = 0.15
	trustWeightResponse    = 0.05
)

// Service defines the review service interface.
type Service interface {
	Create(ctx context.Context, jobID, reviewerID uuid.UUID, rating int, comment string) (*domain.Review, error)
	Respond(ctx context.Context, reviewID, providerID uuid.UUID, response string) error
	GetProviderStats(ctx context.Context, providerID uuid.UUID) (*domain.ProviderReviewStats, error)
	CalculateTrustScore(ctx context.Context, providerID uuid.UUID) (float64, error)
	Moderate(ctx context.Context, reviewID uuid.UUID, action domain.ModerationStatus) error
}

// ReviewService implements review and trust-score business logic.
type ReviewService struct {
	reviews   domain.ReviewRepository
	jobs      domain.JobRepository
	disputes  domain.DisputeRepository
	providers domain.ProviderRepository
}

// NewReviewService returns a ready-to-use ReviewService.
func NewReviewService(
	reviews domain.ReviewRepository,
	jobs domain.JobRepository,
	disputes domain.DisputeRepository,
	providers domain.ProviderRepository,
) *ReviewService {
	return &ReviewService{
		reviews:   reviews,
		jobs:      jobs,
		disputes:  disputes,
		providers: providers,
	}
}

// Create adds a review for a completed job. Both the customer and provider can
// review each other (bidirectional).
func (s *ReviewService) Create(ctx context.Context, jobID, reviewerID uuid.UUID, rating int, comment string) (*domain.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("%w: rating must be between 1 and 5", domain.ErrInvalidInput)
	}

	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if job.Status != domain.JobStatusCompleted {
		return nil, fmt.Errorf("%w: can only review completed jobs", domain.ErrInvalidState)
	}

	// Determine who is being reviewed.
	var revieweeID uuid.UUID
	if reviewerID == job.CustomerID {
		if job.ProviderID == nil {
			return nil, fmt.Errorf("%w: job has no provider", domain.ErrInvalidState)
		}
		revieweeID = *job.ProviderID
	} else if job.ProviderID != nil && reviewerID == *job.ProviderID {
		revieweeID = job.CustomerID
	} else {
		return nil, fmt.Errorf("%w: reviewer is not part of this job", domain.ErrUnauthorized)
	}

	// Prevent duplicate reviews.
	existing, _ := s.reviews.GetByJobAndReviewer(ctx, jobID, reviewerID)
	if existing != nil {
		return nil, fmt.Errorf("%w: already reviewed this job", domain.ErrAlreadyExists)
	}

	review := &domain.Review{
		ID:               uuid.New(),
		JobID:            jobID,
		ReviewerID:       reviewerID,
		RevieweeID:       revieweeID,
		Rating:           rating,
		Comment:          comment,
		Language:         "en",
		ModerationStatus: domain.ModerationPending,
	}

	if err := s.reviews.Create(ctx, review); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to create review")
		return nil, fmt.Errorf("create review: %w", err)
	}

	log.Info().
		Str("review_id", review.ID.String()).
		Str("job_id", jobID.String()).
		Int("rating", rating).
		Msg("review created")

	// Recalculate the provider's trust score asynchronously.
	// In production this would be dispatched to a background worker.
	go func() {
		bgCtx := context.Background()
		if _, err := s.CalculateTrustScore(bgCtx, revieweeID); err != nil {
			log.Warn().Err(err).Str("provider_id", revieweeID.String()).Msg("trust score recalculation failed")
		}
	}()

	return review, nil
}

// Respond allows the reviewee (typically a provider) to reply to a review.
func (s *ReviewService) Respond(ctx context.Context, reviewID, providerID uuid.UUID, response string) error {
	review, err := s.reviews.GetByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("%w: review %s", domain.ErrNotFound, reviewID)
	}

	if review.RevieweeID != providerID {
		return fmt.Errorf("%w: only the reviewee can respond", domain.ErrUnauthorized)
	}

	if review.ProviderResponse != nil {
		return fmt.Errorf("%w: response already submitted", domain.ErrAlreadyExists)
	}

	now := time.Now()
	if err := s.reviews.UpdateResponse(ctx, reviewID, response, now); err != nil {
		return fmt.Errorf("update response: %w", err)
	}

	log.Info().Str("review_id", reviewID.String()).Msg("provider response added")
	return nil
}

// GetProviderStats returns aggregate review statistics for a provider.
func (s *ReviewService) GetProviderStats(ctx context.Context, providerID uuid.UUID) (*domain.ProviderReviewStats, error) {
	stats, err := s.reviews.GetProviderStats(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("get provider stats: %w", err)
	}
	return stats, nil
}

// CalculateTrustScore computes the weighted trust score for a provider and
// persists the result.
//
// Formula:
//
//	On-time rate:       25%
//	Completion rate:    25%
//	Customer ratings:   30%
//	Dispute rate:       15% (inverted — fewer disputes = higher score)
//	Response time:       5%
func (s *ReviewService) CalculateTrustScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	provider, err := s.providers.GetByID(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("%w: provider %s", domain.ErrNotFound, providerID)
	}

	// --- Avg rating (normalise 0-5 to 0-1) ---
	avgRating, totalReviews, err := s.reviews.GetAverageRating(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("get avg rating: %w", err)
	}
	ratingFactor := avgRating / 5.0

	// --- Completion rate ---
	completionRate := 1.0
	if provider.JobsCompleted > 0 {
		// Simplistic: assume 95% completion if they have a track record.
		// A full implementation would track cancelled vs completed.
		completionRate = 0.95
	}

	// --- On-time rate ---
	// Placeholder: would be calculated from scheduled vs actual completion times.
	onTimeRate := 0.90

	// --- Dispute rate (inverted) ---
	disputeCount, err := s.disputes.CountByProvider(ctx, providerID)
	if err != nil {
		disputeCount = 0
	}
	disputeRate := 1.0
	if totalReviews > 0 {
		disputeRate = 1.0 - (float64(disputeCount) / float64(totalReviews))
		if disputeRate < 0 {
			disputeRate = 0
		}
	}

	// --- Response time (normalise) ---
	responseTimeFactor := 0.5
	if provider.ResponseTimeAvg > 0 && provider.ResponseTimeAvg < 30 {
		responseTimeFactor = 1.0
	} else if provider.ResponseTimeAvg > 0 && provider.ResponseTimeAvg < 60 {
		responseTimeFactor = 0.7
	}

	// Weighted sum (produces a 0-1 value).
	score := onTimeRate*trustWeightOnTime +
		completionRate*trustWeightCompletion +
		ratingFactor*trustWeightRating +
		disputeRate*trustWeightDispute +
		responseTimeFactor*trustWeightResponse

	// Scale to 0-5 for storage (matches trust_score DECIMAL(3,2) column).
	trustScore := score * 5.0

	// Persist.
	if err := s.providers.UpdateTrustScore(ctx, providerID, trustScore); err != nil {
		log.Warn().Err(err).Str("provider_id", providerID.String()).Msg("failed to persist trust score")
	}

	log.Info().
		Str("provider_id", providerID.String()).
		Float64("trust_score", trustScore).
		Msg("trust score recalculated")

	return trustScore, nil
}

// Moderate updates the moderation status of a review (admin action).
func (s *ReviewService) Moderate(ctx context.Context, reviewID uuid.UUID, action domain.ModerationStatus) error {
	if _, err := s.reviews.GetByID(ctx, reviewID); err != nil {
		return fmt.Errorf("%w: review %s", domain.ErrNotFound, reviewID)
	}

	if err := s.reviews.UpdateModerationStatus(ctx, reviewID, action); err != nil {
		return fmt.Errorf("moderate review: %w", err)
	}

	log.Info().
		Str("review_id", reviewID.String()).
		Str("action", string(action)).
		Msg("review moderated")

	return nil
}
