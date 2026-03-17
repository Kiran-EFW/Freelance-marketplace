package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the lifecycle state of a job.
type JobStatus string

const (
	JobStatusPosted     JobStatus = "posted"
	JobStatusMatched    JobStatus = "matched"
	JobStatusAccepted   JobStatus = "accepted"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusCancelled  JobStatus = "cancelled"
	JobStatusDisputed   JobStatus = "disputed"
)

// PaymentMethod indicates how the customer intends to pay.
type PaymentMethod string

const (
	PaymentMethodOnline PaymentMethod = "online"
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodWallet PaymentMethod = "wallet"
)

// Job represents a service request created by a customer.
type Job struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	CustomerID     uuid.UUID     `json:"customer_id" db:"customer_id"`
	ProviderID     *uuid.UUID    `json:"provider_id,omitempty" db:"provider_id"`
	CategoryID     uuid.UUID     `json:"category_id" db:"category_id"`
	Postcode       string        `json:"postcode" db:"postcode"`
	Latitude       float64       `json:"latitude" db:"latitude"`
	Longitude      float64       `json:"longitude" db:"longitude"`
	Status         JobStatus     `json:"status" db:"status"`
	Description    string        `json:"description" db:"description"`
	ScheduledAt    *time.Time    `json:"scheduled_at,omitempty" db:"scheduled_at"`
	QuotedPrice    *float64      `json:"quoted_price,omitempty" db:"quoted_price"`
	FinalPrice     *float64      `json:"final_price,omitempty" db:"final_price"`
	Currency       string        `json:"currency" db:"currency"`
	PaymentMethod  PaymentMethod `json:"payment_method" db:"payment_method"`
	IsRecurring    bool          `json:"is_recurring" db:"is_recurring"`
	RecurrenceRule *string       `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	JurisdictionID string        `json:"jurisdiction_id" db:"jurisdiction_id"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// JobSearchFilters holds optional filters for listing or searching jobs.
type JobSearchFilters struct {
	Status     *JobStatus
	CategoryID *uuid.UUID
	Postcode   *string
	Latitude   *float64
	Longitude  *float64
	RadiusKM   *float64
	FromDate   *time.Time
	ToDate     *time.Time
	Limit      int
	Offset     int
}

// JobRepository defines persistence operations for jobs.
type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]Job, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]Job, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status JobStatus) error
	Search(ctx context.Context, filters JobSearchFilters) ([]Job, error)
}
