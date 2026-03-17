package job

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// validTransitions defines which status changes are permitted.
var validTransitions = map[domain.JobStatus][]domain.JobStatus{
	domain.JobStatusPosted:     {domain.JobStatusMatched, domain.JobStatusCancelled},
	domain.JobStatusMatched:    {domain.JobStatusAccepted, domain.JobStatusPosted, domain.JobStatusCancelled},
	domain.JobStatusAccepted:   {domain.JobStatusInProgress, domain.JobStatusCancelled},
	domain.JobStatusInProgress: {domain.JobStatusCompleted, domain.JobStatusDisputed, domain.JobStatusCancelled},
	domain.JobStatusCompleted:  {domain.JobStatusDisputed},
	domain.JobStatusDisputed:   {domain.JobStatusResolved},
	domain.JobStatusCancelled:  {},
}

// JobStatusResolved is used only after a dispute is resolved; it returns the
// job to a terminal state.
const JobStatusResolved domain.JobStatus = "resolved"

// CreateJobParams holds the inputs for creating a new job.
type CreateJobParams struct {
	CategoryID     uuid.UUID
	Postcode       string
	Latitude       float64
	Longitude      float64
	Description    string
	ScheduledAt    *time.Time
	QuotedPrice    *float64
	Currency       string
	PaymentMethod  domain.PaymentMethod
	IsRecurring    bool
	RecurrenceRule *string
	JurisdictionID string
}

// QuoteParams holds the inputs for submitting a provider quote.
type QuoteParams struct {
	Amount  float64
	Message string
}

// Service defines the job service interface.
type Service interface {
	Create(ctx context.Context, customerID uuid.UUID, params CreateJobParams) (*domain.Job, error)
	GetByID(ctx context.Context, jobID uuid.UUID) (*domain.Job, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, filters domain.JobSearchFilters) ([]domain.Job, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, filters domain.JobSearchFilters) ([]domain.Job, error)
	Accept(ctx context.Context, jobID, providerID uuid.UUID) error
	Decline(ctx context.Context, jobID, providerID uuid.UUID) error
	Start(ctx context.Context, jobID uuid.UUID) error
	Complete(ctx context.Context, jobID uuid.UUID) error
	Cancel(ctx context.Context, jobID uuid.UUID, cancelledBy uuid.UUID, reason string) error
	SubmitQuote(ctx context.Context, jobID, providerID uuid.UUID, quote QuoteParams) error
}

// JobService implements job lifecycle management.
type JobService struct {
	jobs domain.JobRepository
}

// NewJobService returns a ready-to-use JobService.
func NewJobService(jobs domain.JobRepository) *JobService {
	return &JobService{jobs: jobs}
}

// Create posts a new job on behalf of a customer.
func (s *JobService) Create(ctx context.Context, customerID uuid.UUID, params CreateJobParams) (*domain.Job, error) {
	if params.Description == "" {
		return nil, fmt.Errorf("%w: description is required", domain.ErrInvalidInput)
	}
	if params.Postcode == "" {
		return nil, fmt.Errorf("%w: postcode is required", domain.ErrInvalidInput)
	}

	currency := params.Currency
	if currency == "" {
		currency = "INR"
	}

	job := &domain.Job{
		ID:             uuid.New(),
		CustomerID:     customerID,
		CategoryID:     params.CategoryID,
		Postcode:       params.Postcode,
		Latitude:       params.Latitude,
		Longitude:      params.Longitude,
		Status:         domain.JobStatusPosted,
		Description:    params.Description,
		ScheduledAt:    params.ScheduledAt,
		QuotedPrice:    params.QuotedPrice,
		Currency:       currency,
		PaymentMethod:  params.PaymentMethod,
		IsRecurring:    params.IsRecurring,
		RecurrenceRule: params.RecurrenceRule,
		JurisdictionID: params.JurisdictionID,
	}

	if err := s.jobs.Create(ctx, job); err != nil {
		log.Error().Err(err).Str("customer_id", customerID.String()).Msg("failed to create job")
		return nil, fmt.Errorf("create job: %w", err)
	}

	log.Info().
		Str("job_id", job.ID.String()).
		Str("customer_id", customerID.String()).
		Str("postcode", params.Postcode).
		Msg("job created")

	return job, nil
}

// GetByID retrieves a single job.
func (s *JobService) GetByID(ctx context.Context, jobID uuid.UUID) (*domain.Job, error) {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}
	return job, nil
}

// ListByCustomer returns jobs created by a customer.
func (s *JobService) ListByCustomer(ctx context.Context, customerID uuid.UUID, filters domain.JobSearchFilters) ([]domain.Job, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	return s.jobs.ListByCustomer(ctx, customerID, limit, filters.Offset)
}

// ListByProvider returns jobs assigned to a provider.
func (s *JobService) ListByProvider(ctx context.Context, providerID uuid.UUID, filters domain.JobSearchFilters) ([]domain.Job, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	return s.jobs.ListByProvider(ctx, providerID, limit, filters.Offset)
}

// Accept lets a provider accept a matched or posted job.
func (s *JobService) Accept(ctx context.Context, jobID, providerID uuid.UUID) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if err := s.validateTransition(job.Status, domain.JobStatusAccepted); err != nil {
		return err
	}

	job.ProviderID = &providerID
	if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusAccepted); err != nil {
		return fmt.Errorf("accept job: %w", err)
	}

	log.Info().
		Str("job_id", jobID.String()).
		Str("provider_id", providerID.String()).
		Msg("job accepted")

	return nil
}

// Decline removes a provider's interest in a job.
func (s *JobService) Decline(ctx context.Context, jobID, providerID uuid.UUID) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	// If the provider was assigned, move the job back to posted.
	if job.ProviderID != nil && *job.ProviderID == providerID {
		if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusPosted); err != nil {
			return fmt.Errorf("decline job: %w", err)
		}
	}

	log.Info().
		Str("job_id", jobID.String()).
		Str("provider_id", providerID.String()).
		Msg("job declined")

	return nil
}

// Start marks a job as in progress.
func (s *JobService) Start(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if err := s.validateTransition(job.Status, domain.JobStatusInProgress); err != nil {
		return err
	}

	if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusInProgress); err != nil {
		return fmt.Errorf("start job: %w", err)
	}

	log.Info().Str("job_id", jobID.String()).Msg("job started")
	return nil
}

// Complete marks a job as finished and triggers post-completion flows
// (review requests, payments, gamification points).
func (s *JobService) Complete(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if err := s.validateTransition(job.Status, domain.JobStatusCompleted); err != nil {
		return err
	}

	if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusCompleted); err != nil {
		return fmt.Errorf("complete job: %w", err)
	}

	log.Info().Str("job_id", jobID.String()).Msg("job completed")

	// TODO: trigger review request notification
	// TODO: trigger payment release from escrow
	// TODO: award gamification points

	return nil
}

// Cancel aborts a job with a reason. Only allowed from certain states.
func (s *JobService) Cancel(ctx context.Context, jobID uuid.UUID, cancelledBy uuid.UUID, reason string) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if err := s.validateTransition(job.Status, domain.JobStatusCancelled); err != nil {
		return err
	}

	if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusCancelled); err != nil {
		return fmt.Errorf("cancel job: %w", err)
	}

	log.Info().
		Str("job_id", jobID.String()).
		Str("cancelled_by", cancelledBy.String()).
		Str("reason", reason).
		Msg("job cancelled")

	return nil
}

// SubmitQuote records a provider's price offer for a job.
func (s *JobService) SubmitQuote(ctx context.Context, jobID, providerID uuid.UUID, quote QuoteParams) error {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	if job.Status != domain.JobStatusPosted && job.Status != domain.JobStatusMatched {
		return fmt.Errorf("%w: cannot quote on job in status %s", domain.ErrInvalidState, job.Status)
	}

	if quote.Amount <= 0 {
		return fmt.Errorf("%w: quote amount must be positive", domain.ErrInvalidInput)
	}

	// In a full implementation this would write to the job_quotes table.
	// For now we log the intent.
	log.Info().
		Str("job_id", jobID.String()).
		Str("provider_id", providerID.String()).
		Float64("amount", quote.Amount).
		Msg("quote submitted")

	return nil
}

// validateTransition checks whether moving from `from` to `to` is allowed.
func (s *JobService) validateTransition(from, to domain.JobStatus) error {
	allowed, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("%w: no transitions from status %s", domain.ErrInvalidState, from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("%w: cannot move from %s to %s", domain.ErrInvalidState, from, to)
}
