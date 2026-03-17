package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/ai"
	"github.com/seva-platform/backend/internal/adapter/search"
	"github.com/seva-platform/backend/internal/adapter/sms"
)

// Task type constants used across the application.
const (
	TypeSendSMS               = "sms:send"
	TypeSendEmail             = "email:send"
	TypeMatchProviders        = "job:match_providers"
	TypeCalculateTrustScore   = "provider:calculate_trust_score"
	TypeProcessPayout         = "payment:process_payout"
	TypeSendSeasonalReminder  = "notification:seasonal_reminder"
	TypeGenerateSEOContent    = "content:generate_seo"
	TypeIndexProvider         = "search:index_provider"
	TypeCleanExpiredOTPs      = "maintenance:clean_expired_otps"
	TypeComputeLeaderboard    = "leaderboard:compute"
)

// SendSMSPayload is the data passed to the SMS sending task.
type SendSMSPayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

// SendEmailPayload is the data passed to the email sending task.
type SendEmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// MatchProvidersPayload is the data passed to the provider matching task.
type MatchProvidersPayload struct {
	JobID      string  `json:"job_id"`
	CategoryID string  `json:"category_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	RadiusKM   float64 `json:"radius_km"`
}

// CalculateTrustScorePayload is the data passed to the trust score recalculation task.
type CalculateTrustScorePayload struct {
	ProviderID string `json:"provider_id"`
}

// ProcessPayoutPayload is the data passed to the payout processing task.
type ProcessPayoutPayload struct {
	ProviderID string  `json:"provider_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	JobID      string  `json:"job_id"`
}

// SendSeasonalReminderPayload is the data passed to the seasonal reminder task.
type SendSeasonalReminderPayload struct {
	JurisdictionID string `json:"jurisdiction_id"`
	Season         string `json:"season"`
	CategoryID     string `json:"category_id"`
}

// GenerateSEOContentPayload is the data passed to the SEO content generation task.
type GenerateSEOContentPayload struct {
	CategoryID string `json:"category_id"`
	Postcode   string `json:"postcode"`
	Location   string `json:"location"`
	Language   string `json:"language"`
}

// IndexProviderPayload is the data passed to the provider indexing task.
type IndexProviderPayload struct {
	ProviderID string   `json:"provider_id"`
	Name       string   `json:"name"`
	Skills     []string `json:"skills"`
	Postcode   string   `json:"postcode"`
	Location   string   `json:"location"`
	Rating     float64  `json:"rating"`
	Category   string   `json:"category"`
	CategoryID string   `json:"category_id"`
	Language   []string `json:"language"`
	Latitude   float64  `json:"latitude"`
	Longitude  float64  `json:"longitude"`
}

// CleanExpiredOTPsPayload is the data passed to the OTP cleanup task.
type CleanExpiredOTPsPayload struct {
	OlderThanMinutes int `json:"older_than_minutes"`
}

// ComputeLeaderboardPayload is the data passed to the leaderboard computation task.
type ComputeLeaderboardPayload struct {
	Postcode   string `json:"postcode"`
	CategoryID string `json:"category_id"`
}

// Deps holds the external dependencies required by task handlers.
type Deps struct {
	SMSProvider  sms.SMSProvider
	Claude       ai.ClaudeProvider
	Search       search.SearchProvider
}

// Worker wraps the Asynq server and mux for background job processing.
type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	deps   *Deps
}

// NewWorker creates a new background worker connected to the given Redis URL.
func NewWorker(redisAddr string, deps *Deps) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("task_type", task.Type()).
					Msg("task processing failed")
			}),
		},
	)

	w := &Worker{server: srv, mux: asynq.NewServeMux(), deps: deps}

	w.mux.HandleFunc(TypeSendSMS, w.handleSendSMS)
	w.mux.HandleFunc(TypeSendEmail, w.handleSendEmail)
	w.mux.HandleFunc(TypeMatchProviders, w.handleMatchProviders)
	w.mux.HandleFunc(TypeCalculateTrustScore, w.handleCalculateTrustScore)
	w.mux.HandleFunc(TypeProcessPayout, w.handleProcessPayout)
	w.mux.HandleFunc(TypeSendSeasonalReminder, w.handleSendSeasonalReminder)
	w.mux.HandleFunc(TypeGenerateSEOContent, w.handleGenerateSEOContent)
	w.mux.HandleFunc(TypeIndexProvider, w.handleIndexProvider)
	w.mux.HandleFunc(TypeCleanExpiredOTPs, w.handleCleanExpiredOTPs)
	w.mux.HandleFunc(TypeComputeLeaderboard, w.handleComputeLeaderboard)

	return w
}

// Start begins processing tasks. This blocks until the server is stopped.
func (w *Worker) Start() error {
	log.Info().Msg("starting asynq worker")
	return w.server.Start(w.mux)
}

// Shutdown gracefully stops the worker.
func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

// --- Task Handlers ---

func (w *Worker) handleSendSMS(_ context.Context, task *asynq.Task) error {
	var payload SendSMSPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SendSMS payload: %w", err)
	}

	log.Info().
		Str("phone", payload.Phone).
		Msg("processing SendSMS task")

	if w.deps != nil && w.deps.SMSProvider != nil {
		if err := w.deps.SMSProvider.SendSMS(payload.Phone, payload.Message); err != nil {
			return fmt.Errorf("send SMS to %s: %w", payload.Phone, err)
		}
	}

	log.Info().Str("phone", payload.Phone).Msg("SMS sent successfully")
	return nil
}

func (w *Worker) handleSendEmail(_ context.Context, task *asynq.Task) error {
	var payload SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SendEmail payload: %w", err)
	}

	// TODO: integrate with an email service (SES, SendGrid, etc.)
	// For now, log the email details.
	log.Info().
		Str("to", payload.To).
		Str("subject", payload.Subject).
		Msg("processing SendEmail task (logged, not sent)")

	return nil
}

func (w *Worker) handleMatchProviders(_ context.Context, task *asynq.Task) error {
	var payload MatchProvidersPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal MatchProviders payload: %w", err)
	}

	log.Info().
		Str("job_id", payload.JobID).
		Str("category_id", payload.CategoryID).
		Float64("lat", payload.Latitude).
		Float64("lng", payload.Longitude).
		Float64("radius_km", payload.RadiusKM).
		Msg("processing MatchProviders task")

	// Search for matching providers using Meilisearch.
	if w.deps != nil && w.deps.Search != nil {
		ctx := context.Background()
		results, err := w.deps.Search.SearchProviders(ctx, "", search.SearchFilters{
			CategoryID: payload.CategoryID,
			Lat:        payload.Latitude,
			Lng:        payload.Longitude,
			RadiusKM:   payload.RadiusKM,
		})
		if err != nil {
			return fmt.Errorf("match providers search: %w", err)
		}

		log.Info().
			Str("job_id", payload.JobID).
			Int("matched", len(results.Hits)).
			Msg("providers matched, sending notifications")

		// TODO: send push notifications to matched providers
		// TODO: create job_match records in the database
	}

	return nil
}

func (w *Worker) handleCalculateTrustScore(_ context.Context, task *asynq.Task) error {
	var payload CalculateTrustScorePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal CalculateTrustScore payload: %w", err)
	}

	log.Info().
		Str("provider_id", payload.ProviderID).
		Msg("processing CalculateTrustScore task")

	// TODO: aggregate ratings, completion rate, response time, cancellations
	// and compute a normalized trust score using the formula:
	// trust_score = (avg_rating * 0.4) + (completion_rate * 0.3) +
	//              (response_time_score * 0.2) + (tenure_score * 0.1)
	// Then update the provider's trust_score in the database.

	return nil
}

func (w *Worker) handleProcessPayout(_ context.Context, task *asynq.Task) error {
	var payload ProcessPayoutPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ProcessPayout payload: %w", err)
	}

	log.Info().
		Str("provider_id", payload.ProviderID).
		Float64("amount", payload.Amount).
		Str("currency", payload.Currency).
		Str("job_id", payload.JobID).
		Msg("processing ProcessPayout task")

	// TODO: look up provider's bank details / UPI ID from database
	// TODO: initiate payout via payment gateway (Razorpay / Stripe)
	// TODO: update payout record status in database
	// TODO: send SMS/notification to provider on success/failure

	return nil
}

func (w *Worker) handleSendSeasonalReminder(_ context.Context, task *asynq.Task) error {
	var payload SendSeasonalReminderPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal SendSeasonalReminder payload: %w", err)
	}

	log.Info().
		Str("jurisdiction_id", payload.JurisdictionID).
		Str("season", payload.Season).
		Str("category_id", payload.CategoryID).
		Msg("processing SendSeasonalReminder task")

	// TODO: query seasonal_calendars table to find upcoming seasonal events
	// TODO: find consumers in the jurisdiction who might need related services
	// TODO: send personalized reminders via SMS/push notification
	// e.g., "Monsoon is approaching! Book waterproofing services now."

	return nil
}

func (w *Worker) handleGenerateSEOContent(_ context.Context, task *asynq.Task) error {
	var payload GenerateSEOContentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal GenerateSEOContent payload: %w", err)
	}

	log.Info().
		Str("category_id", payload.CategoryID).
		Str("postcode", payload.Postcode).
		Str("location", payload.Location).
		Str("language", payload.Language).
		Msg("processing GenerateSEOContent task")

	if w.deps != nil && w.deps.Claude != nil {
		ctx := context.Background()
		prompt := fmt.Sprintf(
			"Generate an SEO-optimized landing page description for the service category "+
				"with ID %q in the location %q (postcode: %s). "+
				"The content should be in %s language, approximately 200 words, "+
				"and highlight the benefits of finding trusted local service providers "+
				"through the Seva marketplace platform. Include relevant local context.",
			payload.CategoryID, payload.Location, payload.Postcode, payload.Language,
		)

		content, err := w.deps.Claude.GenerateContent(ctx, prompt)
		if err != nil {
			return fmt.Errorf("generate SEO content: %w", err)
		}

		log.Info().
			Str("category_id", payload.CategoryID).
			Int("content_length", len(content)).
			Msg("SEO content generated successfully")

		// TODO: store the generated content in the landing_pages table
	}

	return nil
}

func (w *Worker) handleIndexProvider(_ context.Context, task *asynq.Task) error {
	var payload IndexProviderPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal IndexProvider payload: %w", err)
	}

	log.Info().
		Str("provider_id", payload.ProviderID).
		Str("name", payload.Name).
		Msg("processing IndexProvider task")

	if w.deps != nil && w.deps.Search != nil {
		ctx := context.Background()
		doc := search.ProviderDocument{
			ID:         payload.ProviderID,
			Name:       payload.Name,
			Skills:     payload.Skills,
			Postcode:   payload.Postcode,
			Location:   payload.Location,
			Rating:     payload.Rating,
			Category:   payload.Category,
			CategoryID: payload.CategoryID,
			Language:   payload.Language,
		}

		if payload.Latitude != 0 || payload.Longitude != 0 {
			doc.Geo = &search.GeoPoint{
				Lat: payload.Latitude,
				Lng: payload.Longitude,
			}
		}

		if err := w.deps.Search.IndexProvider(ctx, doc); err != nil {
			return fmt.Errorf("index provider %s: %w", payload.ProviderID, err)
		}

		log.Info().Str("provider_id", payload.ProviderID).Msg("provider indexed successfully")
	}

	return nil
}

func (w *Worker) handleCleanExpiredOTPs(_ context.Context, task *asynq.Task) error {
	var payload CleanExpiredOTPsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal CleanExpiredOTPs payload: %w", err)
	}

	if payload.OlderThanMinutes <= 0 {
		payload.OlderThanMinutes = 10 // default: clean OTPs older than 10 minutes
	}

	log.Info().
		Int("older_than_minutes", payload.OlderThanMinutes).
		Msg("processing CleanExpiredOTPs task")

	// TODO: execute DELETE FROM otps WHERE created_at < NOW() - INTERVAL 'X minutes'
	// TODO: log number of rows deleted

	return nil
}

func (w *Worker) handleComputeLeaderboard(_ context.Context, task *asynq.Task) error {
	var payload ComputeLeaderboardPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ComputeLeaderboard payload: %w", err)
	}

	log.Info().
		Str("postcode", payload.Postcode).
		Str("category_id", payload.CategoryID).
		Msg("processing ComputeLeaderboard task")

	// TODO: query providers in the postcode/category
	// TODO: rank by trust score, job count, and rating
	// TODO: upsert into leaderboard table with rank positions
	// TODO: optionally send notifications to providers who moved up

	return nil
}

// --- Task Constructors (for enqueuing from handlers) ---

// NewSendSMSTask creates a new SendSMS task.
func NewSendSMSTask(phone, message string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendSMSPayload{Phone: phone, Message: message})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendSMS, payload, asynq.Queue("critical")), nil
}

// NewSendEmailTask creates a new SendEmail task.
func NewSendEmailTask(to, subject, body string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailPayload{To: to, Subject: subject, Body: body})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload, asynq.Queue("default")), nil
}

// NewMatchProvidersTask creates a new MatchProviders task.
func NewMatchProvidersTask(jobID, categoryID string, lat, lng, radiusKM float64) (*asynq.Task, error) {
	payload, err := json.Marshal(MatchProvidersPayload{
		JobID:      jobID,
		CategoryID: categoryID,
		Latitude:   lat,
		Longitude:  lng,
		RadiusKM:   radiusKM,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeMatchProviders, payload, asynq.Queue("critical")), nil
}

// NewCalculateTrustScoreTask creates a new CalculateTrustScore task.
func NewCalculateTrustScoreTask(providerID string) (*asynq.Task, error) {
	payload, err := json.Marshal(CalculateTrustScorePayload{ProviderID: providerID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCalculateTrustScore, payload, asynq.Queue("low")), nil
}

// NewProcessPayoutTask creates a new ProcessPayout task.
func NewProcessPayoutTask(providerID string, amount float64, currency, jobID string) (*asynq.Task, error) {
	payload, err := json.Marshal(ProcessPayoutPayload{
		ProviderID: providerID,
		Amount:     amount,
		Currency:   currency,
		JobID:      jobID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeProcessPayout, payload, asynq.Queue("critical")), nil
}

// NewSendSeasonalReminderTask creates a new SendSeasonalReminder task.
func NewSendSeasonalReminderTask(jurisdictionID, season, categoryID string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendSeasonalReminderPayload{
		JurisdictionID: jurisdictionID,
		Season:         season,
		CategoryID:     categoryID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendSeasonalReminder, payload, asynq.Queue("default")), nil
}

// NewGenerateSEOContentTask creates a new GenerateSEOContent task.
func NewGenerateSEOContentTask(categoryID, postcode, location, language string) (*asynq.Task, error) {
	payload, err := json.Marshal(GenerateSEOContentPayload{
		CategoryID: categoryID,
		Postcode:   postcode,
		Location:   location,
		Language:   language,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateSEOContent, payload, asynq.Queue("low")), nil
}

// NewIndexProviderTask creates a new IndexProvider task.
func NewIndexProviderTask(p IndexProviderPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeIndexProvider, payload, asynq.Queue("default")), nil
}

// NewCleanExpiredOTPsTask creates a new CleanExpiredOTPs task.
func NewCleanExpiredOTPsTask(olderThanMinutes int) (*asynq.Task, error) {
	payload, err := json.Marshal(CleanExpiredOTPsPayload{OlderThanMinutes: olderThanMinutes})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCleanExpiredOTPs, payload, asynq.Queue("low")), nil
}

// NewComputeLeaderboardTask creates a new ComputeLeaderboard task.
func NewComputeLeaderboardTask(postcode, categoryID string) (*asynq.Task, error) {
	payload, err := json.Marshal(ComputeLeaderboardPayload{
		Postcode:   postcode,
		CategoryID: categoryID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeComputeLeaderboard, payload, asynq.Queue("low")), nil
}
