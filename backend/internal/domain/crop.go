package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CropCatalog represents a crop entry in the jurisdiction-specific catalog.
type CropCatalog struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	JurisdictionID   string          `json:"jurisdiction_id" db:"jurisdiction_id"`
	CropSlug         string          `json:"crop_slug" db:"crop_slug"`
	Name             json.RawMessage `json:"name" db:"name"`
	WorkTypes        json.RawMessage `json:"work_types" db:"work_types"`
	SeasonalCalendar json.RawMessage `json:"seasonal_calendar" db:"seasonal_calendar"`
	Metadata         json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	IsActive         bool            `json:"is_active" db:"is_active"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// CropRepository defines persistence operations for the crop catalog.
type CropRepository interface {
	ListByJurisdiction(ctx context.Context, jurisdictionID string) ([]CropCatalog, error)
	GetBySlug(ctx context.Context, jurisdictionID, cropSlug string) (*CropCatalog, error)
	Create(ctx context.Context, crop *CropCatalog) error
	Update(ctx context.Context, crop *CropCatalog) error
	Delete(ctx context.Context, id uuid.UUID) error
}
