package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Category represents a service category in the marketplace (e.g., plumbing, electrical).
type Category struct {
	ID              uuid.UUID          `json:"id" db:"id"`
	Slug            string             `json:"slug" db:"slug"`
	Name            map[string]string  `json:"name" db:"name"`                // i18n: {"en": "Plumbing", "hi": "..."}
	ParentID        *uuid.UUID         `json:"parent_id,omitempty" db:"parent_id"`
	Icon            string             `json:"icon" db:"icon"`
	IsActive        bool               `json:"is_active" db:"is_active"`
	RequiresLicense bool               `json:"requires_license" db:"requires_license"`
	Metadata        json.RawMessage    `json:"metadata,omitempty" db:"metadata"`
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListActive(ctx context.Context) ([]Category, error)
}
