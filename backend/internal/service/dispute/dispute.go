package dispute

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// autoResolvableTypes are dispute types that the system can attempt to resolve
// automatically at tier 1.
var autoResolvableTypes = map[domain.DisputeType]bool{
	domain.DisputeTypeNoShow:     true,
	domain.DisputeTypeOvercharge: true,
	domain.DisputeTypeLateArrival: true,
}

// Service defines the dispute service interface.
type Service interface {
	Create(ctx context.Context, jobID, raisedBy uuid.UUID, disputeType domain.DisputeType, description string) (*domain.Dispute, error)
	DetermineSeverity(disputeType domain.DisputeType, description string) domain.DisputeSeverity
	AutoResolve(ctx context.Context, disputeID uuid.UUID) (bool, error)
	Escalate(ctx context.Context, disputeID uuid.UUID, reason string) error
	Resolve(ctx context.Context, disputeID uuid.UUID, resolution domain.Resolution) error
	ListByStatus(ctx context.Context, status domain.DisputeStatus, page int) ([]domain.Dispute, error)
}

// DisputeService implements dispute lifecycle management.
type DisputeService struct {
	disputes domain.DisputeRepository
	jobs     domain.JobRepository
}

// NewDisputeService returns a ready-to-use DisputeService.
func NewDisputeService(disputes domain.DisputeRepository, jobs domain.JobRepository) *DisputeService {
	return &DisputeService{
		disputes: disputes,
		jobs:     jobs,
	}
}

// Create opens a new dispute for a job.
func (s *DisputeService) Create(ctx context.Context, jobID, raisedBy uuid.UUID, disputeType domain.DisputeType, description string) (*domain.Dispute, error) {
	if description == "" {
		return nil, fmt.Errorf("%w: description is required", domain.ErrInvalidInput)
	}

	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("%w: job %s", domain.ErrNotFound, jobID)
	}

	// Determine the opposing party.
	var against uuid.UUID
	if raisedBy == job.CustomerID && job.ProviderID != nil {
		against = *job.ProviderID
	} else if job.ProviderID != nil && raisedBy == *job.ProviderID {
		against = job.CustomerID
	} else {
		return nil, fmt.Errorf("%w: raiser is not part of this job", domain.ErrUnauthorized)
	}

	severity := s.DetermineSeverity(disputeType, description)

	dispute := &domain.Dispute{
		ID:          uuid.New(),
		JobID:       jobID,
		RaisedBy:    raisedBy,
		Against:     against,
		Type:        disputeType,
		Severity:    severity,
		Status:      domain.DisputeStatusOpen,
		Description: description,
	}

	if err := s.disputes.Create(ctx, dispute); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to create dispute")
		return nil, fmt.Errorf("create dispute: %w", err)
	}

	// Move the job to disputed status.
	if err := s.jobs.UpdateStatus(ctx, jobID, domain.JobStatusDisputed); err != nil {
		log.Warn().Err(err).Str("job_id", jobID.String()).Msg("failed to update job status to disputed")
	}

	log.Info().
		Str("dispute_id", dispute.ID.String()).
		Str("job_id", jobID.String()).
		Str("type", string(disputeType)).
		Str("severity", string(severity)).
		Msg("dispute created")

	return dispute, nil
}

// DetermineSeverity auto-classifies the severity tier based on dispute type
// and contextual signals.
func (s *DisputeService) DetermineSeverity(disputeType domain.DisputeType, description string) domain.DisputeSeverity {
	switch disputeType {
	case domain.DisputeTypeHarassment:
		return domain.DisputeSeverityCritical
	case domain.DisputeTypePropertyDamage:
		return domain.DisputeSeverityHigh
	case domain.DisputeTypeNoShow, domain.DisputeTypeNonPayment:
		return domain.DisputeSeverityMedium
	case domain.DisputeTypeOvercharge, domain.DisputeTypeLateArrival:
		return domain.DisputeSeverityLow
	case domain.DisputeTypeQuality:
		// Quality disputes can vary; default to medium.
		return domain.DisputeSeverityMedium
	default:
		return domain.DisputeSeverityLow
	}
}

// AutoResolve attempts tier 1 automatic resolution for simple dispute types.
// Returns true if the dispute was resolved, false if it needs human review.
func (s *DisputeService) AutoResolve(ctx context.Context, disputeID uuid.UUID) (bool, error) {
	dispute, err := s.disputes.GetByID(ctx, disputeID)
	if err != nil {
		return false, fmt.Errorf("%w: dispute %s", domain.ErrNotFound, disputeID)
	}

	if dispute.Status != domain.DisputeStatusOpen {
		return false, fmt.Errorf("%w: dispute is not open", domain.ErrInvalidState)
	}

	// Only certain types are eligible for auto-resolution.
	if !autoResolvableTypes[dispute.Type] {
		log.Info().
			Str("dispute_id", disputeID.String()).
			Str("type", string(dispute.Type)).
			Msg("dispute type not eligible for auto-resolution")
		return false, nil
	}

	// Critical and high severity always need human review.
	if dispute.Severity == domain.DisputeSeverityCritical || dispute.Severity == domain.DisputeSeverityHigh {
		return false, nil
	}

	// Simple auto-resolution logic:
	// - No-show: auto-refund the customer.
	// - Overcharge: flag for review but mark as investigating.
	// - Late arrival: issue a warning, close dispute.

	switch dispute.Type {
	case domain.DisputeTypeNoShow:
		resolution := domain.Resolution{
			DisputeID:      disputeID,
			ResolutionType: "refund",
			Notes:          "Auto-resolved: provider no-show confirmed. Full refund issued.",
		}
		if err := s.disputes.Resolve(ctx, disputeID, resolution); err != nil {
			return false, fmt.Errorf("auto-resolve no-show: %w", err)
		}
		log.Info().Str("dispute_id", disputeID.String()).Msg("auto-resolved: no-show")
		return true, nil

	case domain.DisputeTypeLateArrival:
		resolution := domain.Resolution{
			DisputeID:      disputeID,
			ResolutionType: "warning",
			Notes:          "Auto-resolved: provider warned about late arrival.",
		}
		if err := s.disputes.Resolve(ctx, disputeID, resolution); err != nil {
			return false, fmt.Errorf("auto-resolve late arrival: %w", err)
		}
		log.Info().Str("dispute_id", disputeID.String()).Msg("auto-resolved: late arrival")
		return true, nil

	default:
		// Move to investigating for human review.
		if err := s.disputes.UpdateStatus(ctx, disputeID, domain.DisputeStatusInvestigating); err != nil {
			return false, fmt.Errorf("update status: %w", err)
		}
		return false, nil
	}
}

// Escalate moves a dispute to the escalated tier for senior review.
func (s *DisputeService) Escalate(ctx context.Context, disputeID uuid.UUID, reason string) error {
	dispute, err := s.disputes.GetByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("%w: dispute %s", domain.ErrNotFound, disputeID)
	}

	if dispute.Status == domain.DisputeStatusResolved || dispute.Status == domain.DisputeStatusClosed {
		return fmt.Errorf("%w: cannot escalate a resolved or closed dispute", domain.ErrInvalidState)
	}

	now := time.Now()
	if err := s.disputes.Escalate(ctx, disputeID, now); err != nil {
		return fmt.Errorf("escalate dispute: %w", err)
	}

	if err := s.disputes.UpdateStatus(ctx, disputeID, domain.DisputeStatusEscalated); err != nil {
		return fmt.Errorf("update escalated status: %w", err)
	}

	log.Info().
		Str("dispute_id", disputeID.String()).
		Str("reason", reason).
		Msg("dispute escalated")

	return nil
}

// Resolve closes a dispute with a resolution.
func (s *DisputeService) Resolve(ctx context.Context, disputeID uuid.UUID, resolution domain.Resolution) error {
	dispute, err := s.disputes.GetByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("%w: dispute %s", domain.ErrNotFound, disputeID)
	}

	if dispute.Status == domain.DisputeStatusResolved || dispute.Status == domain.DisputeStatusClosed {
		return fmt.Errorf("%w: dispute is already resolved or closed", domain.ErrInvalidState)
	}

	if err := s.disputes.Resolve(ctx, disputeID, resolution); err != nil {
		return fmt.Errorf("resolve dispute: %w", err)
	}

	log.Info().
		Str("dispute_id", disputeID.String()).
		Str("resolution_type", resolution.ResolutionType).
		Msg("dispute resolved")

	return nil
}

// ListByStatus returns a paginated list of disputes filtered by status.
func (s *DisputeService) ListByStatus(ctx context.Context, status domain.DisputeStatus, page int) ([]domain.Dispute, error) {
	const pageSize = 20
	offset := page * pageSize

	disputes, err := s.disputes.ListByStatus(ctx, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list disputes: %w", err)
	}
	return disputes, nil
}
