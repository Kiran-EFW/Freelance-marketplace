package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DisputeType classifies the nature of a dispute.
type DisputeType string

const (
	DisputeTypeQuality        DisputeType = "quality"
	DisputeTypeNoShow         DisputeType = "no_show"
	DisputeTypeOvercharge     DisputeType = "overcharge"
	DisputeTypePropertyDamage DisputeType = "property_damage"
	DisputeTypeHarassment     DisputeType = "harassment"
	DisputeTypeNonPayment     DisputeType = "non_payment"
	DisputeTypeLateArrival    DisputeType = "late_arrival"
	DisputeTypeOther          DisputeType = "other"
)

// DisputeSeverity indicates the urgency and impact of a dispute.
type DisputeSeverity string

const (
	DisputeSeverityLow      DisputeSeverity = "low"
	DisputeSeverityMedium   DisputeSeverity = "medium"
	DisputeSeverityHigh     DisputeSeverity = "high"
	DisputeSeverityCritical DisputeSeverity = "critical"
)

// DisputeStatus represents the lifecycle state of a dispute.
type DisputeStatus string

const (
	DisputeStatusOpen          DisputeStatus = "open"
	DisputeStatusInvestigating DisputeStatus = "under_review"
	DisputeStatusMediation     DisputeStatus = "mediation"
	DisputeStatusResolved      DisputeStatus = "resolved"
	DisputeStatusEscalated     DisputeStatus = "escalated"
	DisputeStatusClosed        DisputeStatus = "closed"
)

// Dispute represents a formal complaint raised by either party after a job.
type Dispute struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	JobID            uuid.UUID       `json:"job_id" db:"job_id"`
	RaisedBy         uuid.UUID       `json:"raised_by" db:"raised_by"`
	Against          uuid.UUID       `json:"against" db:"against"`
	Type             DisputeType     `json:"type" db:"type"`
	Severity         DisputeSeverity `json:"severity" db:"severity"`
	Status           DisputeStatus   `json:"status" db:"status"`
	Description      string          `json:"description" db:"description"`
	Evidence         json.RawMessage `json:"evidence,omitempty" db:"evidence"`
	ResolutionNotes  *string         `json:"resolution,omitempty" db:"resolution"`
	ResolvedBy       *uuid.UUID      `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolutionAmount *float64        `json:"resolution_amount,omitempty" db:"resolution_amount"`
	EscalatedAt      *time.Time      `json:"escalated_at,omitempty" db:"escalated_at"`
	ResolvedAt       *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// Resolution captures how a dispute was resolved.
type Resolution struct {
	DisputeID      uuid.UUID `json:"dispute_id"`
	ResolvedBy     uuid.UUID `json:"resolved_by"`
	ResolutionType string    `json:"resolution_type"` // "refund", "warning", "ban", "dismissed"
	RefundAmount   float64   `json:"refund_amount"`
	Notes          string    `json:"notes"`
}

// DisputeRepository defines persistence operations for disputes.
type DisputeRepository interface {
	Create(ctx context.Context, dispute *Dispute) error
	GetByID(ctx context.Context, id uuid.UUID) (*Dispute, error)
	GetByJobID(ctx context.Context, jobID uuid.UUID) ([]Dispute, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status DisputeStatus) error
	UpdateSeverity(ctx context.Context, id uuid.UUID, severity DisputeSeverity) error
	Resolve(ctx context.Context, id uuid.UUID, resolution Resolution) error
	Escalate(ctx context.Context, id uuid.UUID, escalatedAt time.Time) error
	ListByStatus(ctx context.Context, status DisputeStatus, limit, offset int) ([]Dispute, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Dispute, error)
	CountByProvider(ctx context.Context, providerID uuid.UUID) (int, error)
}
