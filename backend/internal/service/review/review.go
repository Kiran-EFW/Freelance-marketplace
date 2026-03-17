package review

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// Trust score component weights — must sum to 1.0.
const (
	trustWeightRating     = 0.40
	trustWeightCompletion = 0.25
	trustWeightResponse   = 0.15
	trustWeightVolume     = 0.10
	trustWeightRecency    = 0.10
)

// Provider level thresholds.
const (
	levelActiveMinScore  = 2.0
	levelActiveMinJobs   = 3
	levelTrustedMinScore = 3.5
	levelTrustedMinJobs  = 20
	levelExpertMinScore  = 4.0
	levelExpertMinJobs   = 50
	levelChampMinScore   = 4.5
	levelChampMinJobs    = 100
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

// CalculateTrustScore computes a comprehensive weighted trust score for a
// provider, persists the result, and evaluates whether the provider should
// level up.
//
// Component weights (sum to 1.0):
//
//	Average rating  (40%): avg of all ratings, normalised to 0-1
//	Completion rate (25%): completed / (completed + cancelled + disputed)
//	Response time   (15%): inverse of avg response time, normalised
//	Volume bonus    (10%): log2(total_jobs+1) / log2(100), capped at 1.0
//	Recency         (10%): recent reviews (last 90 days) weighted 2x
//
// After computing the base score, modifiers are applied:
//
//	verified provider:    +0.10
//	has profile photo:    +0.05
//	has bank account:     +0.05
//
// The final result is clamped to the 0.00 - 5.00 range.
func (s *ReviewService) CalculateTrustScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	// 1. Fetch provider profile.
	provider, err := s.providers.GetByID(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("%w: provider %s", domain.ErrNotFound, providerID)
	}

	// 2. Fetch all reviews for recency-weighted rating calculation.
	allReviews, err := s.reviews.ListByReviewee(ctx, providerID, 1000, 0)
	if err != nil {
		// Fall back to aggregate if listing fails.
		allReviews = nil
	}

	// Also fetch aggregate stats as a fallback / for total count.
	avgRating, totalReviews, err := s.reviews.GetAverageRating(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("get avg rating: %w", err)
	}

	// ---------------------------------------------------------------
	// Component 1: Average Rating (40%) — normalised to 0-1
	// ---------------------------------------------------------------
	ratingFactor := avgRating / 5.0

	// ---------------------------------------------------------------
	// Component 2: Completion Rate (25%)
	// completed / (completed + cancelled + disputed)
	// ---------------------------------------------------------------
	completionRate := 1.0
	totalJobs := provider.JobsCompleted
	disputeCount, _ := s.disputes.CountByProvider(ctx, providerID)

	if totalJobs > 0 {
		// Denominator: completed jobs + disputes (approximation for
		// cancelled/disputed jobs since the schema does not track
		// cancelled separately on the provider profile).
		denominator := float64(totalJobs) + float64(disputeCount)
		if denominator > 0 {
			completionRate = float64(totalJobs) / denominator
		}
	}

	// ---------------------------------------------------------------
	// Component 3: Response Time (15%) — inverse normalisation
	// Under 15 min = 1.0, under 30 min = 0.8, under 60 = 0.5,
	// over 60 min = 0.2, no data = 0.5 (neutral).
	// ---------------------------------------------------------------
	responseTimeFactor := 0.5 // default when no data
	avgResponseMin := provider.ResponseTimeAvg
	if avgResponseMin > 0 {
		switch {
		case avgResponseMin <= 15:
			responseTimeFactor = 1.0
		case avgResponseMin <= 30:
			responseTimeFactor = 0.8
		case avgResponseMin <= 60:
			responseTimeFactor = 0.5
		case avgResponseMin <= 120:
			responseTimeFactor = 0.3
		default:
			responseTimeFactor = 0.2
		}
	}

	// ---------------------------------------------------------------
	// Component 4: Volume Bonus (10%)
	// log2(total_jobs + 1) / log2(100), capped at 1.0
	// ---------------------------------------------------------------
	volumeFactor := 0.0
	if totalJobs > 0 {
		volumeFactor = math.Log2(float64(totalJobs)+1) / math.Log2(100)
		if volumeFactor > 1.0 {
			volumeFactor = 1.0
		}
	}

	// ---------------------------------------------------------------
	// Component 5: Recency (10%)
	// Reviews from the last 90 days count 2x when computing an
	// alternative average. Compare with the all-time average and
	// blend toward the recency-weighted value.
	// ---------------------------------------------------------------
	recencyFactor := ratingFactor // fallback to overall rating
	if len(allReviews) > 0 {
		cutoff := time.Now().Add(-90 * 24 * time.Hour)
		var weightedSum, weightTotal float64
		for _, r := range allReviews {
			weight := 1.0
			if r.CreatedAt.After(cutoff) {
				weight = 2.0
			}
			weightedSum += float64(r.Rating) * weight
			weightTotal += weight
		}
		if weightTotal > 0 {
			recencyFactor = (weightedSum / weightTotal) / 5.0
		}
	}

	// ---------------------------------------------------------------
	// Weighted sum (produces a 0-1 value).
	// ---------------------------------------------------------------
	baseScore := ratingFactor*trustWeightRating +
		completionRate*trustWeightCompletion +
		responseTimeFactor*trustWeightResponse +
		volumeFactor*trustWeightVolume +
		recencyFactor*trustWeightRecency

	// Scale to 0-5.
	trustScore := baseScore * 5.0

	// ---------------------------------------------------------------
	// 4. Provider-level modifiers
	// ---------------------------------------------------------------
	if provider.VerificationStatus == domain.VerificationApproved {
		trustScore += 0.10
	}
	// Profile photo: check if avatar/photo exists (user-level field).
	// Since ProviderProfile does not directly have an avatar, we approximate
	// by checking whether the provider has a bio set (indicates profile effort).
	if provider.Bio != "" {
		trustScore += 0.05
	}
	if provider.BankAccountID != nil && *provider.BankAccountID != "" {
		trustScore += 0.05
	}

	// ---------------------------------------------------------------
	// 5. Clamp to 0.00 - 5.00
	// ---------------------------------------------------------------
	if trustScore < 0 {
		trustScore = 0
	}
	if trustScore > 5.0 {
		trustScore = 5.0
	}

	// Round to two decimal places.
	trustScore = math.Round(trustScore*100) / 100

	// ---------------------------------------------------------------
	// 6. Persist trust score.
	// ---------------------------------------------------------------
	if err := s.providers.UpdateTrustScore(ctx, providerID, trustScore); err != nil {
		log.Warn().Err(err).Str("provider_id", providerID.String()).Msg("failed to persist trust score")
	}

	// ---------------------------------------------------------------
	// 7. Evaluate level progression.
	// ---------------------------------------------------------------
	newLevel := evaluateProviderLevel(trustScore, totalJobs, totalReviews)
	if newLevel != "" {
		log.Info().
			Str("provider_id", providerID.String()).
			Str("new_level", newLevel).
			Float64("trust_score", trustScore).
			Int("total_jobs", totalJobs).
			Msg("provider level evaluated")
	}

	log.Info().
		Str("provider_id", providerID.String()).
		Float64("trust_score", trustScore).
		Float64("rating_factor", ratingFactor).
		Float64("completion_rate", completionRate).
		Float64("response_factor", responseTimeFactor).
		Float64("volume_factor", volumeFactor).
		Float64("recency_factor", recencyFactor).
		Msg("trust score recalculated")

	return trustScore, nil
}

// evaluateProviderLevel determines which provider level a provider should hold
// based on their trust score and total jobs completed. Returns the level name
// as a string matching the provider_level enum.
//
// Level thresholds:
//
//	new -> active:         trust_score >= 2.0 AND total_jobs >= 3
//	active -> trusted:     trust_score >= 3.5 AND total_jobs >= 20
//	trusted -> expert:     trust_score >= 4.0 AND total_jobs >= 50
//	expert -> local_champion: trust_score >= 4.5 AND total_jobs >= 100
func evaluateProviderLevel(trustScore float64, totalJobs, totalReviews int) string {
	switch {
	case trustScore >= levelChampMinScore && totalJobs >= levelChampMinJobs:
		return "local_champion"
	case trustScore >= levelExpertMinScore && totalJobs >= levelExpertMinJobs:
		return "expert"
	case trustScore >= levelTrustedMinScore && totalJobs >= levelTrustedMinJobs:
		return "trusted"
	case trustScore >= levelActiveMinScore && totalJobs >= levelActiveMinJobs:
		return "active"
	default:
		return "new"
	}
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
