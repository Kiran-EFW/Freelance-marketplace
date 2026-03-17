package crop

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// Service defines the crop calendar service interface.
type Service interface {
	GetSeasonalCalendar(ctx context.Context, jurisdictionID string, month int) ([]CropWork, error)
	GetCropsByJurisdiction(ctx context.Context, jurisdictionID string) ([]CropCatalogEntry, error)
}

// PriceRange represents a min/max price range with currency.
type PriceRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Currency string  `json:"currency"`
}

// WorkType describes a specific type of agricultural work.
type WorkType struct {
	Slug         string     `json:"slug"`
	Name         string     `json:"name"`
	PricingModel string     `json:"pricing_model"`
	TypicalPrice PriceRange `json:"typical_price"`
	IsInSeason   bool       `json:"is_in_season"`
}

// CropWork aggregates a crop and its available work types for a given month.
type CropWork struct {
	CropName  string     `json:"crop_name"`
	CropSlug  string     `json:"crop_slug"`
	WorkTypes []WorkType `json:"work_types"`
}

// CropCatalogEntry represents a full crop catalog record for a jurisdiction.
type CropCatalogEntry struct {
	CropSlug         string            `json:"crop_slug"`
	Name             map[string]string `json:"name"`
	WorkTypes        []WorkType        `json:"work_types"`
	SeasonalCalendar map[string][]string `json:"seasonal_calendar"`
	IsActive         bool              `json:"is_active"`
}

// CropService implements crop calendar business logic.
type CropService struct {
	crops domain.CropRepository
}

// NewCropService returns a ready-to-use CropService.
func NewCropService(crops domain.CropRepository) *CropService {
	return &CropService{crops: crops}
}

// workTypeEntry is the raw JSON shape stored in the crop_catalog.work_types column.
type workTypeEntry struct {
	Slug         string            `json:"slug"`
	Name         map[string]string `json:"name"`
	PricingModel string            `json:"pricing_model"`
	TypicalPrice PriceRange        `json:"typical_price"`
}

// GetSeasonalCalendar returns the available crop work types for a given
// jurisdiction and calendar month (1-12). Work types that are active in the
// requested month are marked with IsInSeason = true.
func (s *CropService) GetSeasonalCalendar(ctx context.Context, jurisdictionID string, month int) ([]CropWork, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("%w: month must be between 1 and 12", domain.ErrInvalidInput)
	}
	if jurisdictionID == "" {
		jurisdictionID = "in"
	}

	entries, err := s.crops.ListByJurisdiction(ctx, jurisdictionID)
	if err != nil {
		return nil, fmt.Errorf("list crops: %w", err)
	}

	monthKey := strconv.Itoa(month)
	var results []CropWork

	for _, entry := range entries {
		if !entry.IsActive {
			continue
		}

		// Parse work_types JSON
		var rawWorkTypes []workTypeEntry
		if err := json.Unmarshal(entry.WorkTypes, &rawWorkTypes); err != nil {
			log.Warn().Err(err).Str("crop", entry.CropSlug).Msg("failed to parse work_types")
			continue
		}

		// Parse seasonal_calendar JSON
		var calendar map[string][]string
		if err := json.Unmarshal(entry.SeasonalCalendar, &calendar); err != nil {
			log.Warn().Err(err).Str("crop", entry.CropSlug).Msg("failed to parse seasonal_calendar")
			continue
		}

		// Determine which work types are in season this month
		inSeasonSlugs := make(map[string]bool)
		if slugs, ok := calendar[monthKey]; ok {
			for _, slug := range slugs {
				inSeasonSlugs[slug] = true
			}
		}

		// Build the work type list
		var workTypes []WorkType
		for _, wt := range rawWorkTypes {
			name := wt.Name["en"]
			if name == "" {
				name = wt.Slug
			}
			workTypes = append(workTypes, WorkType{
				Slug:         wt.Slug,
				Name:         name,
				PricingModel: wt.PricingModel,
				TypicalPrice: wt.TypicalPrice,
				IsInSeason:   inSeasonSlugs[wt.Slug],
			})
		}

		// Parse the crop name
		var nameMap map[string]string
		if err := json.Unmarshal(entry.Name, &nameMap); err != nil {
			log.Warn().Err(err).Str("crop", entry.CropSlug).Msg("failed to parse crop name")
			continue
		}

		cropName := nameMap["en"]
		if cropName == "" {
			cropName = entry.CropSlug
		}

		results = append(results, CropWork{
			CropName:  cropName,
			CropSlug:  entry.CropSlug,
			WorkTypes: workTypes,
		})
	}

	log.Info().
		Str("jurisdiction", jurisdictionID).
		Int("month", month).
		Int("crops", len(results)).
		Msg("seasonal calendar retrieved")

	return results, nil
}

// GetCropsByJurisdiction returns all crops catalogued for a given jurisdiction.
func (s *CropService) GetCropsByJurisdiction(ctx context.Context, jurisdictionID string) ([]CropCatalogEntry, error) {
	if jurisdictionID == "" {
		jurisdictionID = "in"
	}

	entries, err := s.crops.ListByJurisdiction(ctx, jurisdictionID)
	if err != nil {
		return nil, fmt.Errorf("list crops: %w", err)
	}

	var results []CropCatalogEntry
	for _, entry := range entries {
		var nameMap map[string]string
		if err := json.Unmarshal(entry.Name, &nameMap); err != nil {
			nameMap = map[string]string{"en": entry.CropSlug}
		}

		var rawWorkTypes []workTypeEntry
		if err := json.Unmarshal(entry.WorkTypes, &rawWorkTypes); err != nil {
			rawWorkTypes = nil
		}

		var calendar map[string][]string
		if err := json.Unmarshal(entry.SeasonalCalendar, &calendar); err != nil {
			calendar = make(map[string][]string)
		}

		var workTypes []WorkType
		for _, wt := range rawWorkTypes {
			name := wt.Name["en"]
			if name == "" {
				name = wt.Slug
			}
			workTypes = append(workTypes, WorkType{
				Slug:         wt.Slug,
				Name:         name,
				PricingModel: wt.PricingModel,
				TypicalPrice: wt.TypicalPrice,
			})
		}

		results = append(results, CropCatalogEntry{
			CropSlug:         entry.CropSlug,
			Name:             nameMap,
			WorkTypes:        workTypes,
			SeasonalCalendar: calendar,
			IsActive:         entry.IsActive,
		})
	}

	return results, nil
}
