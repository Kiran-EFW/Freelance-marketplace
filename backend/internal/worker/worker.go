package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/ai"
	"github.com/seva-platform/backend/internal/adapter/payment"
	"github.com/seva-platform/backend/internal/adapter/search"
	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/config"
)

// Task type constants used across the application.
const (
	TypeSendSMS              = "sms:send"
	TypeSendEmail            = "email:send"
	TypeMatchProviders       = "job:match_providers"
	TypeCalculateTrustScore  = "provider:calculate_trust_score"
	TypeProcessPayout        = "payment:process_payout"
	TypeSendSeasonalReminder = "notification:seasonal_reminder"
	TypeGenerateSEOContent   = "content:generate_seo"
	TypeIndexProvider        = "search:index_provider"
	TypeCleanExpiredOTPs       = "maintenance:clean_expired_otps"
	TypeComputeLeaderboard     = "leaderboard:compute"
	TypeProcessRecurringJobs   = "recurring:process_due"
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
	CropID         string `json:"crop_id"`
	Activity       string `json:"activity"`
	Month          string `json:"month"`
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

// EmailConfig holds SMTP configuration for sending emails.
type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// Deps holds the external dependencies required by task handlers.
type Deps struct {
	SMSProvider    sms.SMSProvider
	Claude         ai.ClaudeProvider
	Search         search.SearchProvider
	DB             *pgxpool.Pool
	Redis          *redis.Client
	PaymentGateway payment.PaymentGateway
	Email          *EmailConfig
	Cfg            *config.Config
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
	w.mux.HandleFunc(TypeProcessRecurringJobs, w.handleProcessRecurringJobs)

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

	log.Info().
		Str("to", payload.To).
		Str("subject", payload.Subject).
		Msg("processing SendEmail task")

	emailCfg := w.getEmailConfig()
	if emailCfg == nil || emailCfg.Host == "" || emailCfg.Host == "localhost" {
		log.Warn().
			Str("to", payload.To).
			Str("subject", payload.Subject).
			Msg("SMTP not configured, email logged but not sent")
		return nil
	}

	// Build the email message in RFC 822 format.
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		emailCfg.From, payload.To, payload.Subject, payload.Body,
	)

	addr := fmt.Sprintf("%s:%s", emailCfg.Host, emailCfg.Port)
	var auth smtp.Auth
	if emailCfg.User != "" {
		auth = smtp.PlainAuth("", emailCfg.User, emailCfg.Password, emailCfg.Host)
	}

	if err := smtp.SendMail(addr, auth, emailCfg.From, []string{payload.To}, []byte(msg)); err != nil {
		return fmt.Errorf("send email to %s via SMTP: %w", payload.To, err)
	}

	log.Info().Str("to", payload.To).Str("subject", payload.Subject).Msg("email sent successfully")
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

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping trust score calculation")
		return nil
	}

	ctx := context.Background()

	// 1. Fetch average rating for the provider.
	var avgRating float64
	var reviewCount int
	err := w.deps.DB.QueryRow(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*)
		 FROM reviews
		 WHERE reviewee_id = $1`, payload.ProviderID,
	).Scan(&avgRating, &reviewCount)
	if err != nil {
		return fmt.Errorf("query reviews for provider %s: %w", payload.ProviderID, err)
	}

	// 2. Calculate completion rate.
	var totalJobs, completedJobs int
	err = w.deps.DB.QueryRow(ctx,
		`SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'completed')
		 FROM jobs
		 WHERE assigned_provider_id = $1`, payload.ProviderID,
	).Scan(&totalJobs, &completedJobs)
	if err != nil {
		return fmt.Errorf("query jobs for provider %s: %w", payload.ProviderID, err)
	}

	completionRate := 0.0
	if totalJobs > 0 {
		completionRate = float64(completedJobs) / float64(totalJobs)
	}

	// 3. Calculate response time score (inverse of avg response time in minutes).
	var avgResponseMinutes float64
	err = w.deps.DB.QueryRow(ctx,
		`SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (q.created_at - j.created_at)) / 60), 0)
		 FROM quotes q
		 JOIN jobs j ON j.id = q.job_id
		 WHERE q.provider_id = $1`, payload.ProviderID,
	).Scan(&avgResponseMinutes)
	if err != nil {
		// Non-fatal: if we cannot get response time, default to 0 score.
		log.Warn().Err(err).Str("provider_id", payload.ProviderID).Msg("failed to compute avg response time")
		avgResponseMinutes = 1440 // default to 24 hours
	}

	// Response time score: faster is better. Map 0-1440 minutes to 1.0-0.0.
	// Use an inverse formula: score = 1 / (1 + avgResponseMinutes/60)
	responseTimeScore := 1.0 / (1.0 + avgResponseMinutes/60.0)

	// 4. Volume score: log(total_jobs + 1) normalized to 0-1.
	// Use log base 10; 100 jobs maps to score ~1.0.
	volumeScore := math.Log10(float64(totalJobs)+1) / 2.0
	if volumeScore > 1.0 {
		volumeScore = 1.0
	}

	// 5. Compute weighted trust score:
	// avg_rating (40%) + completion_rate (25%) + response_time (20%) + volume (15%)
	// Normalize avg_rating from 0-5 scale to 0-1.
	normalizedRating := avgRating / 5.0

	trustScore := (normalizedRating * 0.40) +
		(completionRate * 0.25) +
		(responseTimeScore * 0.20) +
		(volumeScore * 0.15)

	// Scale to 0-100 for storage.
	trustScoreScaled := math.Round(trustScore * 100)

	// 6. Update provider_profiles.trust_score in the database.
	_, err = w.deps.DB.Exec(ctx,
		`UPDATE provider_profiles
		 SET trust_score = $1, updated_at = NOW()
		 WHERE id = $2`, trustScoreScaled, payload.ProviderID,
	)
	if err != nil {
		return fmt.Errorf("update trust_score for provider %s: %w", payload.ProviderID, err)
	}

	log.Info().
		Str("provider_id", payload.ProviderID).
		Float64("trust_score", trustScoreScaled).
		Float64("avg_rating", avgRating).
		Int("review_count", reviewCount).
		Float64("completion_rate", completionRate).
		Float64("response_time_score", responseTimeScore).
		Float64("volume_score", volumeScore).
		Msg("trust score calculated and updated")

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

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping payout processing")
		return nil
	}

	ctx := context.Background()

	// 1. Look up provider's bank details / fund account from provider_profiles.
	var bankAccountID, contactID, phone string
	err := w.deps.DB.QueryRow(ctx,
		`SELECT
			COALESCE(pp.bank_account_id, ''),
			COALESCE(pp.razorpay_contact_id, ''),
			COALESCE(u.phone, '')
		 FROM provider_profiles pp
		 JOIN users u ON u.id = pp.user_id
		 WHERE pp.id = $1`, payload.ProviderID,
	).Scan(&bankAccountID, &contactID, &phone)
	if err != nil {
		return fmt.Errorf("look up bank details for provider %s: %w", payload.ProviderID, err)
	}

	if bankAccountID == "" {
		log.Error().
			Str("provider_id", payload.ProviderID).
			Msg("provider has no bank account configured, cannot process payout")
		// Update payout status to failed.
		_, _ = w.deps.DB.Exec(ctx,
			`UPDATE payouts SET status = 'failed', failure_reason = 'no_bank_account', updated_at = NOW()
			 WHERE provider_id = $1 AND job_id = $2 AND status = 'pending'`,
			payload.ProviderID, payload.JobID,
		)
		return fmt.Errorf("provider %s has no bank account", payload.ProviderID)
	}

	// 2. Initiate payout via Razorpay Payouts API.
	amountPaise := int64(payload.Amount * 100)

	razorpayPayload := map[string]interface{}{
		"account_number":  bankAccountID,
		"fund_account_id": contactID,
		"amount":          amountPaise,
		"currency":        payload.Currency,
		"mode":            "NEFT",
		"purpose":         "payout",
		"queue_if_low_balance": true,
		"reference_id":    fmt.Sprintf("payout_%s_%s", payload.JobID, payload.ProviderID),
		"narration":       fmt.Sprintf("Seva payout for job %s", payload.JobID),
	}

	payoutJSON, err := json.Marshal(razorpayPayload)
	if err != nil {
		return fmt.Errorf("marshal razorpay payout payload: %w", err)
	}

	// Make the API call to Razorpay.
	var razorpayKeyID, razorpayKeySecret string
	if w.deps.Cfg != nil {
		razorpayKeyID = w.deps.Cfg.RazorpayKeyID
		razorpayKeySecret = w.deps.Cfg.RazorpayKeySecret
	}

	if razorpayKeyID == "" || razorpayKeySecret == "" {
		log.Warn().Msg("Razorpay credentials not configured, logging payout but not executing")
		// Mark as processing for dev environments.
		_, _ = w.deps.DB.Exec(ctx,
			`UPDATE payouts SET status = 'processing', updated_at = NOW()
			 WHERE provider_id = $1 AND job_id = $2 AND status = 'pending'`,
			payload.ProviderID, payload.JobID,
		)
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.razorpay.com/v1/payouts", bytes.NewReader(payoutJSON))
	if err != nil {
		return fmt.Errorf("create razorpay payout request: %w", err)
	}
	req.SetBasicAuth(razorpayKeyID, razorpayKeySecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("razorpay payout API call: %w", err)
	}
	defer resp.Body.Close()

	var razorpayResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Error  struct {
			Description string `json:"description"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&razorpayResp); err != nil {
		return fmt.Errorf("decode razorpay payout response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.Error().
			Int("status", resp.StatusCode).
			Str("error", razorpayResp.Error.Description).
			Str("provider_id", payload.ProviderID).
			Msg("razorpay payout failed")

		_, _ = w.deps.DB.Exec(ctx,
			`UPDATE payouts SET status = 'failed', failure_reason = $1, updated_at = NOW()
			 WHERE provider_id = $2 AND job_id = $3 AND status = 'pending'`,
			razorpayResp.Error.Description, payload.ProviderID, payload.JobID,
		)
		return fmt.Errorf("razorpay payout failed: %s", razorpayResp.Error.Description)
	}

	// 3. Update payout status in database.
	_, err = w.deps.DB.Exec(ctx,
		`UPDATE payouts
		 SET status = 'processing', gateway_payout_id = $1, updated_at = NOW()
		 WHERE provider_id = $2 AND job_id = $3 AND status = 'pending'`,
		razorpayResp.ID, payload.ProviderID, payload.JobID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to update payout status in DB after successful Razorpay call")
	}

	// 4. Send notification to provider.
	if w.deps.SMSProvider != nil && phone != "" {
		msg := fmt.Sprintf(
			"Your payout of %s %.2f for job %s has been initiated. You will receive it within 2-3 business days.",
			payload.Currency, payload.Amount, payload.JobID,
		)
		if err := w.deps.SMSProvider.SendSMS(phone, msg); err != nil {
			log.Warn().Err(err).Str("phone", phone).Msg("failed to send payout notification SMS")
		}
	}

	log.Info().
		Str("provider_id", payload.ProviderID).
		Str("razorpay_payout_id", razorpayResp.ID).
		Str("payout_status", razorpayResp.Status).
		Float64("amount", payload.Amount).
		Msg("payout initiated successfully")

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
		Str("crop_id", payload.CropID).
		Str("activity", payload.Activity).
		Str("month", payload.Month).
		Msg("processing SendSeasonalReminder task")

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping seasonal reminder processing")
		return nil
	}

	// Determine the crop identifier to search by. Prefer crop_id; fall back to category_id.
	cropSlug := payload.CropID
	if cropSlug == "" {
		cropSlug = payload.CategoryID
	}
	if cropSlug == "" {
		log.Warn().Msg("no crop_id or category_id provided, skipping seasonal reminder")
		return nil
	}

	ctx := context.Background()

	// Query users who have active jobs or subscriptions related to this crop/jurisdiction.
	rows, err := w.deps.DB.Query(ctx,
		`SELECT DISTINCT u.phone, u.name
		 FROM users u
		 JOIN jobs j ON (j.customer_id = u.id OR j.provider_id = u.id)
		 JOIN categories c ON j.category_id = c.id
		 WHERE c.slug LIKE '%' || $1 || '%'
		   AND u.phone IS NOT NULL
		   AND u.phone != ''
		 LIMIT 500`,
		cropSlug,
	)
	if err != nil {
		return fmt.Errorf("query users for seasonal reminder (crop=%s): %w", cropSlug, err)
	}
	defer rows.Close()

	type recipient struct {
		Phone string
		Name  string
	}

	var recipients []recipient
	for rows.Next() {
		var r recipient
		if err := rows.Scan(&r.Phone, &r.Name); err != nil {
			log.Error().Err(err).Msg("failed to scan seasonal reminder recipient row")
			continue
		}
		recipients = append(recipients, r)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate seasonal reminder recipient rows: %w", err)
	}

	if len(recipients) == 0 {
		log.Info().
			Str("crop_id", cropSlug).
			Msg("no users found for seasonal reminder, nothing to send")
		return nil
	}

	// Build the SMS message based on the activity and crop.
	activity := payload.Activity
	if activity == "" {
		activity = payload.Season
	}

	// Capitalize the crop name for display (e.g., "coconut" -> "Coconut").
	cropDisplay := cropSlug
	if len(cropDisplay) > 0 {
		cropDisplay = strings.ToUpper(cropDisplay[:1]) + cropDisplay[1:]
	}

	message := fmt.Sprintf(
		"Seva: %s %s season is approaching in your area. Book workers now to avoid the rush. Reply STOP to opt out.",
		cropDisplay, activity,
	)

	// Send SMS notifications to each recipient.
	sentCount := 0
	failedCount := 0

	if w.deps.SMSProvider == nil {
		log.Warn().
			Int("recipients", len(recipients)).
			Msg("SMS provider not configured, logging seasonal reminders but not sending")
		return nil
	}

	for _, r := range recipients {
		if err := w.deps.SMSProvider.SendSMS(r.Phone, message); err != nil {
			log.Warn().
				Err(err).
				Str("phone", r.Phone).
				Str("name", r.Name).
				Msg("failed to send seasonal reminder SMS")
			failedCount++
			continue
		}
		sentCount++
	}

	log.Info().
		Str("crop_id", cropSlug).
		Str("activity", activity).
		Int("total_recipients", len(recipients)).
		Int("sent", sentCount).
		Int("failed", failedCount).
		Msg("seasonal reminder processing complete")

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

	if w.deps == nil || w.deps.Claude == nil {
		log.Warn().Msg("Claude AI not configured, skipping SEO content generation")
		return nil
	}

	ctx := context.Background()

	// Generate title prompt.
	titlePrompt := fmt.Sprintf(
		"Generate a short SEO-optimized page title (under 60 characters) for the service category "+
			"with ID %q in the location %q (postcode: %s). Language: %s. "+
			"Format: '[Service] in [Location] - Seva'. Return only the title, nothing else.",
		payload.CategoryID, payload.Location, payload.Postcode, payload.Language,
	)

	title, err := w.deps.Claude.GenerateContent(ctx, titlePrompt)
	if err != nil {
		return fmt.Errorf("generate SEO title: %w", err)
	}

	// Generate description prompt.
	descPrompt := fmt.Sprintf(
		"Generate an SEO-optimized landing page description for the service category "+
			"with ID %q in the location %q (postcode: %s). "+
			"The content should be in %s language, approximately 200 words, "+
			"and highlight the benefits of finding trusted local service providers "+
			"through the Seva marketplace platform. Include relevant local context.",
		payload.CategoryID, payload.Location, payload.Postcode, payload.Language,
	)

	description, err := w.deps.Claude.GenerateContent(ctx, descPrompt)
	if err != nil {
		return fmt.Errorf("generate SEO description: %w", err)
	}

	// Store in Redis cache for the SEO landing pages to use.
	if w.deps.Redis != nil {
		cacheKey := fmt.Sprintf("seo:%s:%s:%s", payload.CategoryID, payload.Postcode, payload.Language)
		seoData := map[string]string{
			"title":       title,
			"description": description,
			"category_id": payload.CategoryID,
			"postcode":    payload.Postcode,
			"location":    payload.Location,
			"language":    payload.Language,
			"generated_at": time.Now().UTC().Format(time.RFC3339),
		}

		seoJSON, err := json.Marshal(seoData)
		if err != nil {
			return fmt.Errorf("marshal SEO data for cache: %w", err)
		}

		// Cache for 7 days.
		if err := w.deps.Redis.Set(ctx, cacheKey, seoJSON, 7*24*time.Hour).Err(); err != nil {
			log.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to cache SEO content in Redis")
		}
	}

	// Also store in the database if available.
	if w.deps.DB != nil {
		_, err = w.deps.DB.Exec(ctx,
			`INSERT INTO seo_landing_pages (category_id, postcode, location, language, title, description, generated_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			 ON CONFLICT (category_id, postcode, language)
			 DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description,
			              generated_at = NOW(), updated_at = NOW()`,
			payload.CategoryID, payload.Postcode, payload.Location, payload.Language, title, description,
		)
		if err != nil {
			log.Warn().Err(err).Msg("failed to store SEO content in database (table may not exist yet)")
		}
	}

	log.Info().
		Str("category_id", payload.CategoryID).
		Str("location", payload.Location).
		Int("title_length", len(title)).
		Int("description_length", len(description)).
		Msg("SEO content generated and cached successfully")

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

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping OTP cleanup")
		return nil
	}

	ctx := context.Background()

	// Delete expired OTPs from the otp_codes table.
	result, err := w.deps.DB.Exec(ctx,
		`DELETE FROM otp_codes WHERE expires_at < NOW()`,
	)
	if err != nil {
		return fmt.Errorf("delete expired OTPs: %w", err)
	}

	deletedCount := result.RowsAffected()

	log.Info().
		Int64("deleted_count", deletedCount).
		Msg("expired OTPs cleaned up successfully")

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

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping leaderboard computation")
		return nil
	}

	ctx := context.Background()

	// Query providers grouped by postcode/category, ranked by trust_score and total jobs.
	rows, err := w.deps.DB.Query(ctx,
		`SELECT
			pp.id,
			pp.user_id,
			COALESCE(u.name, ''),
			COALESCE(pp.trust_score, 0),
			COALESCE(pp.rating_average, 0),
			COUNT(j.id) FILTER (WHERE j.status = 'completed') AS total_completed
		 FROM provider_profiles pp
		 JOIN users u ON u.id = pp.user_id
		 LEFT JOIN jobs j ON j.assigned_provider_id = pp.id
		 WHERE ($1 = '' OR EXISTS (
			SELECT 1 FROM provider_service_areas psa
			WHERE psa.provider_id = pp.id AND psa.postcode = $1
		 ))
		 AND ($2 = '' OR EXISTS (
			SELECT 1 FROM provider_categories pc
			WHERE pc.provider_id = pp.id AND pc.category_id = $2
		 ))
		 GROUP BY pp.id, pp.user_id, u.name, pp.trust_score, pp.rating_average
		 ORDER BY COALESCE(pp.trust_score, 0) DESC,
		          COUNT(j.id) FILTER (WHERE j.status = 'completed') DESC
		 LIMIT 100`,
		payload.Postcode, payload.CategoryID,
	)
	if err != nil {
		return fmt.Errorf("query providers for leaderboard: %w", err)
	}
	defer rows.Close()

	type leaderboardEntry struct {
		ProviderID     string  `json:"provider_id"`
		UserID         string  `json:"user_id"`
		Name           string  `json:"name"`
		TrustScore     float64 `json:"trust_score"`
		Rating         float64 `json:"rating"`
		TotalCompleted int     `json:"total_completed"`
		Rank           int     `json:"rank"`
	}

	var entries []leaderboardEntry
	rank := 1
	for rows.Next() {
		var e leaderboardEntry
		if err := rows.Scan(&e.ProviderID, &e.UserID, &e.Name, &e.TrustScore, &e.Rating, &e.TotalCompleted); err != nil {
			return fmt.Errorf("scan leaderboard row: %w", err)
		}
		e.Rank = rank
		entries = append(entries, e)
		rank++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate leaderboard rows: %w", err)
	}

	// Store results in Redis as JSON cache.
	if w.deps.Redis != nil {
		cacheKey := fmt.Sprintf("leaderboard:%s:%s", payload.Postcode, payload.CategoryID)
		if cacheKey == "leaderboard::" {
			cacheKey = "leaderboard:global"
		}

		leaderboardJSON, err := json.Marshal(entries)
		if err != nil {
			return fmt.Errorf("marshal leaderboard data: %w", err)
		}

		// Cache for 1 hour.
		if err := w.deps.Redis.Set(ctx, cacheKey, leaderboardJSON, 1*time.Hour).Err(); err != nil {
			log.Warn().Err(err).Str("cache_key", cacheKey).Msg("failed to cache leaderboard in Redis")
		}
	}

	log.Info().
		Str("postcode", payload.Postcode).
		Str("category_id", payload.CategoryID).
		Int("entries", len(entries)).
		Msg("leaderboard computed and cached successfully")

	return nil
}

// --- Helper Methods ---

// getEmailConfig returns the email configuration from deps.
func (w *Worker) getEmailConfig() *EmailConfig {
	if w.deps != nil && w.deps.Email != nil {
		return w.deps.Email
	}
	// Fall back to config if Email dep not set.
	if w.deps != nil && w.deps.Cfg != nil {
		return &EmailConfig{
			Host:     w.deps.Cfg.SMTPHost,
			Port:     w.deps.Cfg.SMTPPort,
			User:     w.deps.Cfg.SMTPUser,
			Password: w.deps.Cfg.SMTPPassword,
			From:     w.deps.Cfg.SMTPFrom,
		}
	}
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
func NewSendSeasonalReminderTask(jurisdictionID, season, categoryID, cropID, activity, month string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendSeasonalReminderPayload{
		JurisdictionID: jurisdictionID,
		Season:         season,
		CategoryID:     categoryID,
		CropID:         cropID,
		Activity:       activity,
		Month:          month,
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

// NewProcessRecurringJobsTask creates a new task that processes due recurring schedules.
func NewProcessRecurringJobsTask() (*asynq.Task, error) {
	return asynq.NewTask(TypeProcessRecurringJobs, nil, asynq.Queue("default")), nil
}

// handleProcessRecurringJobs queries recurring_schedules where next_occurrence <= NOW()
// and status = 'active', creates new job records for each, and updates next_occurrence
// based on the frequency.
func (w *Worker) handleProcessRecurringJobs(_ context.Context, _ *asynq.Task) error {
	log.Info().Msg("processing due recurring job schedules")

	if w.deps == nil || w.deps.DB == nil {
		log.Warn().Msg("DB not available, skipping recurring job processing")
		return nil
	}

	ctx := context.Background()

	// Fetch up to 100 due schedules at a time.
	rows, err := w.deps.DB.Query(ctx,
		`SELECT id, customer_id, provider_id, category_id, title, description,
		        frequency, day_of_week, day_of_month, preferred_time, amount, currency,
		        next_occurrence, total_occurrences, max_occurrences
		 FROM recurring_schedules
		 WHERE next_occurrence <= NOW()
		   AND status = 'active'
		 ORDER BY next_occurrence ASC
		 LIMIT 100`,
	)
	if err != nil {
		return fmt.Errorf("query due recurring schedules: %w", err)
	}
	defer rows.Close()

	type schedule struct {
		ID               string
		CustomerID       string
		ProviderID       string
		CategoryID       string
		Title            string
		Description      string
		Frequency        string
		DayOfWeek        *int
		DayOfMonth       *int
		PreferredTime    string
		Amount           float64
		Currency         string
		NextOccurrence   time.Time
		TotalOccurrences int
		MaxOccurrences   *int
	}

	var schedules []schedule
	for rows.Next() {
		var s schedule
		var dayOfWeek, dayOfMonth, maxOcc *int
		if err := rows.Scan(
			&s.ID, &s.CustomerID, &s.ProviderID, &s.CategoryID,
			&s.Title, &s.Description, &s.Frequency,
			&dayOfWeek, &dayOfMonth, &s.PreferredTime,
			&s.Amount, &s.Currency, &s.NextOccurrence,
			&s.TotalOccurrences, &maxOcc,
		); err != nil {
			log.Error().Err(err).Msg("failed to scan recurring schedule row in worker")
			continue
		}
		s.DayOfWeek = dayOfWeek
		s.DayOfMonth = dayOfMonth
		s.MaxOccurrences = maxOcc
		schedules = append(schedules, s)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate recurring schedule rows: %w", err)
	}

	processedCount := 0
	for _, s := range schedules {
		// Check max_occurrences limit.
		if s.MaxOccurrences != nil && s.TotalOccurrences >= *s.MaxOccurrences {
			// Max reached: set status to cancelled.
			_, _ = w.deps.DB.Exec(ctx,
				`UPDATE recurring_schedules SET status = 'cancelled' WHERE id = $1`, s.ID)
			log.Info().Str("schedule_id", s.ID).Msg("recurring schedule reached max occurrences, cancelled")
			continue
		}

		// Create a new job from this schedule.
		_, err := w.deps.DB.Exec(ctx,
			`INSERT INTO jobs (customer_id, assigned_provider_id, category_id, description, status, quoted_price, currency, is_recurring, scheduled_at)
			 VALUES ($1, $2, $3, $4, 'pending', $5, $6, true, $7)`,
			s.CustomerID, s.ProviderID, s.CategoryID,
			fmt.Sprintf("[Recurring] %s - %s", s.Title, s.Description),
			s.Amount, s.Currency, s.NextOccurrence,
		)
		if err != nil {
			log.Error().Err(err).Str("schedule_id", s.ID).Msg("failed to create job from recurring schedule")
			continue
		}

		// Calculate the next occurrence.
		now := time.Now().UTC()
		var nextOccurrence time.Time
		switch s.Frequency {
		case "daily":
			nextOccurrence = s.NextOccurrence.AddDate(0, 0, 1)
		case "weekly":
			nextOccurrence = s.NextOccurrence.AddDate(0, 0, 7)
		case "biweekly":
			nextOccurrence = s.NextOccurrence.AddDate(0, 0, 14)
		case "monthly":
			nextOccurrence = s.NextOccurrence.AddDate(0, 1, 0)
		case "quarterly":
			nextOccurrence = s.NextOccurrence.AddDate(0, 3, 0)
		default:
			nextOccurrence = s.NextOccurrence.AddDate(0, 0, 7) // fallback to weekly
		}

		// Ensure next occurrence is in the future.
		for nextOccurrence.Before(now) {
			switch s.Frequency {
			case "daily":
				nextOccurrence = nextOccurrence.AddDate(0, 0, 1)
			case "weekly":
				nextOccurrence = nextOccurrence.AddDate(0, 0, 7)
			case "biweekly":
				nextOccurrence = nextOccurrence.AddDate(0, 0, 14)
			case "monthly":
				nextOccurrence = nextOccurrence.AddDate(0, 1, 0)
			case "quarterly":
				nextOccurrence = nextOccurrence.AddDate(0, 3, 0)
			default:
				nextOccurrence = nextOccurrence.AddDate(0, 0, 7)
			}
		}

		// Update schedule with new next_occurrence and increment total_occurrences.
		_, err = w.deps.DB.Exec(ctx,
			`UPDATE recurring_schedules
			 SET next_occurrence = $2, last_occurrence = $3, total_occurrences = total_occurrences + 1
			 WHERE id = $1`,
			s.ID, nextOccurrence, now,
		)
		if err != nil {
			log.Error().Err(err).Str("schedule_id", s.ID).Msg("failed to update recurring schedule next occurrence")
			continue
		}

		processedCount++
		log.Info().
			Str("schedule_id", s.ID).
			Str("title", s.Title).
			Time("next_occurrence", nextOccurrence).
			Msg("recurring job created and schedule updated")
	}

	log.Info().Int("processed", processedCount).Int("total_due", len(schedules)).Msg("recurring job processing complete")
	return nil
}
