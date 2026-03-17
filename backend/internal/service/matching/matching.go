package matching

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/pkg/geo"
)

// Scoring weights — must sum to 1.0.
const (
	weightDistance         = 0.25
	weightTrustScore      = 0.25
	weightResponseTime    = 0.15
	weightPriceCompetitive = 0.15
	weightCompletionRate  = 0.10
	weightLanguageMatch   = 0.05
	weightPlatformLoyalty = 0.05
)

// Boost multipliers applied on top of the base score.
const (
	boostUrgent       = 1.15
	boostBudget       = 1.10
	boostQuality      = 1.10
	boostSubscription = 1.20
)

// Service defines the matching service interface.
type Service interface {
	FindProviders(ctx context.Context, job *domain.Job) ([]ScoredProvider, error)
	MatchAndNotifyTopN(ctx context.Context, job *domain.Job, n int) error
}

// ScoredProvider wraps a search result with a match score.
type ScoredProvider struct {
	domain.ProviderSearchResult
	Score float64 `json:"score"`
}

// MatchingService implements the weighted provider-matching algorithm.
type MatchingService struct {
	providers     domain.ProviderRepository
	reviews       domain.ReviewRepository
	users         domain.UserRepository
	notifications NotificationSender
}

// NotificationSender is the subset of the notification service needed by matching.
type NotificationSender interface {
	Send(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, title, body string, data map[string]interface{}) error
	SendSMS(ctx context.Context, phone, message string) error
}

// NewMatchingService returns a ready-to-use MatchingService.
func NewMatchingService(
	providers domain.ProviderRepository,
	reviews domain.ReviewRepository,
	users domain.UserRepository,
	notifications NotificationSender,
) *MatchingService {
	return &MatchingService{
		providers:     providers,
		reviews:       reviews,
		users:         users,
		notifications: notifications,
	}
}

// FindProviders searches for providers that match a job and ranks them using
// the weighted scoring algorithm.
func (s *MatchingService) FindProviders(ctx context.Context, job *domain.Job) ([]ScoredProvider, error) {
	filters := domain.ProviderSearchFilters{
		CategoryID: &job.CategoryID,
		Postcode:   &job.Postcode,
		Latitude:   &job.Latitude,
		Longitude:  &job.Longitude,
	}

	// Default search radius: 15 km.
	defaultRadius := 15.0
	filters.RadiusKM = &defaultRadius
	available := true
	filters.Available = &available

	results, err := s.providers.Search(ctx, filters)
	if err != nil {
		log.Error().Err(err).Str("job_id", job.ID.String()).Msg("provider search failed")
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Fetch the customer to determine language for matching.
	customer, _ := s.users.GetByID(ctx, job.CustomerID)

	scored := make([]ScoredProvider, 0, len(results))
	for i := range results {
		score := s.scoreProvider(&results[i], job, customer)
		scored = append(scored, ScoredProvider{
			ProviderSearchResult: results[i],
			Score:                score,
		})
	}

	// Apply boosts.
	s.applyBoosts(scored, job)

	// Sort by descending score.
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	log.Info().
		Str("job_id", job.ID.String()).
		Int("candidates", len(scored)).
		Msg("provider matching completed")

	return scored, nil
}

// scoreProvider calculates a normalised 0-1 score for a provider/job pair.
func (s *MatchingService) scoreProvider(provider *domain.ProviderSearchResult, job *domain.Job, customer *domain.User) float64 {
	var score float64

	// --- Distance (closer = higher) ---
	maxDistance := 30.0 // km ceiling
	dist := provider.Distance
	if dist <= 0 {
		dist = geo.DistanceKM(job.Latitude, job.Longitude, provider.Latitude, provider.Longitude)
	}
	distanceScore := 1.0 - math.Min(dist/maxDistance, 1.0)
	score += distanceScore * weightDistance

	// --- Trust score (0-5 normalised to 0-1) ---
	trustScore := provider.TrustScore / 5.0
	score += trustScore * weightTrustScore

	// --- Response time (lower minutes = higher score) ---
	maxResponseMinutes := 120.0
	responseScore := 1.0
	if provider.ResponseTimeAvg > 0 {
		responseScore = 1.0 - math.Min(float64(provider.ResponseTimeAvg)/maxResponseMinutes, 1.0)
	}
	score += responseScore * weightResponseTime

	// --- Price competitiveness ---
	// If the job has a quoted price we prefer providers whose past quotes are
	// close to the budget; otherwise give a neutral 0.5.
	priceScore := 0.5
	if job.QuotedPrice != nil && *job.QuotedPrice > 0 {
		// Providers with more completed jobs tend to offer fairer prices.
		priceScore = math.Min(float64(provider.JobsCompleted)/50.0, 1.0)
	}
	score += priceScore * weightPriceCompetitive

	// --- Completion rate (based on jobs completed vs level) ---
	completionScore := math.Min(float64(provider.JobsCompleted)/100.0, 1.0)
	score += completionScore * weightCompletionRate

	// --- Language match ---
	languageScore := 0.0
	if customer != nil && customer.PreferredLanguage != "" {
		// In a full implementation we would check provider.Languages.
		// For now, same jurisdiction is a proxy for language overlap.
		if customer.JurisdictionID == "in" {
			languageScore = 0.5 // baseline for same country
		}
	}
	score += languageScore * weightLanguageMatch

	// --- Platform loyalty (level as a proxy) ---
	loyaltyScore := math.Min(float64(provider.Level)/5.0, 1.0)
	score += loyaltyScore * weightPlatformLoyalty

	return score
}

// applyBoosts modifies scores based on job urgency, budget, quality preference,
// and provider subscription tier.
func (s *MatchingService) applyBoosts(results []ScoredProvider, job *domain.Job) {
	for i := range results {
		// Urgent jobs boost providers with fast response times.
		// The job's scheduled_at being within 24 hours is treated as urgent.
		if job.ScheduledAt != nil && time.Until(*job.ScheduledAt).Hours() < 24 && results[i].ResponseTimeAvg > 0 && results[i].ResponseTimeAvg < 30 {
			results[i].Score *= boostUrgent
		}

		// Budget jobs boost cheaper providers (higher completion, lower level).
		if job.QuotedPrice != nil && *job.QuotedPrice < 500 {
			results[i].Score *= boostBudget
		}

		// Quality boost for high-trust providers.
		if results[i].TrustScore >= 4.0 {
			results[i].Score *= boostQuality
		}

		// Subscription boost for premium providers.
		if results[i].SubscriptionTier == domain.SubscriptionPremium {
			results[i].Score *= boostSubscription
		}
	}
}

// MatchAndNotifyTopN finds the best N providers for a job and sends them
// notifications. For SMS-only providers the notification is sent as a text
// message.
func (s *MatchingService) MatchAndNotifyTopN(ctx context.Context, job *domain.Job, n int) error {
	scored, err := s.FindProviders(ctx, job)
	if err != nil {
		return err
	}

	if n > len(scored) {
		n = len(scored)
	}

	for i := 0; i < n; i++ {
		provider := scored[i]

		user, err := s.users.GetByID(ctx, provider.UserID)
		if err != nil {
			log.Warn().Err(err).Str("provider_id", provider.UserID.String()).Msg("could not fetch provider user")
			continue
		}

		title := "New job near you"
		body := "A customer needs help. Tap to view details."

		if user.DeviceType == "basic_phone" {
			// SMS fallback for feature-phone users.
			msg := "Seva: New job available near " + job.Postcode + ". Reply YES to accept."
			if err := s.notifications.SendSMS(ctx, user.Phone, msg); err != nil {
				log.Warn().Err(err).Str("phone", user.Phone).Msg("SMS notification failed")
			}
			continue
		}

		data := map[string]interface{}{
			"job_id":   job.ID.String(),
			"postcode": job.Postcode,
		}
		if err := s.notifications.Send(ctx, provider.UserID, domain.NotifJobNew, title, body, data); err != nil {
			log.Warn().Err(err).Str("provider_id", provider.UserID.String()).Msg("push notification failed")
		}
	}

	log.Info().
		Str("job_id", job.ID.String()).
		Int("notified", n).
		Msg("top providers notified")

	return nil
}
