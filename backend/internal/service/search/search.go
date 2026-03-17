// Package search provides the search service implementation with real
// database queries for finding providers, jobs, and categories.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// SearchService implements provider, job, and category search operations
// using direct database queries.
type SearchService struct {
	queries *postgres.Queries
	db      *pgxpool.Pool
}

// NewSearchService returns a ready-to-use SearchService.
func NewSearchService(queries *postgres.Queries, db *pgxpool.Pool) *SearchService {
	return &SearchService{queries: queries, db: db}
}

// SearchProviders finds providers matching the given filters. It supports
// location-based, postcode-based, and skill-based search modes with sorting.
func (s *SearchService) SearchProviders(ctx context.Context, filters domain.ProviderSearchFilters, sortBy string) ([]domain.ProviderSearchResult, int, error) {
	limit := int32(20)
	offset := int32(0)
	if filters.Limit > 0 {
		limit = int32(filters.Limit)
	}
	if filters.Offset > 0 {
		offset = int32(filters.Offset)
	}

	// Location-based search takes priority when lat/lng are provided.
	if filters.Latitude != nil && filters.Longitude != nil {
		radiusMeters := 25000.0 // default 25km
		if filters.RadiusKM != nil {
			radiusMeters = *filters.RadiusKM * 1000
		}

		rows, err := s.queries.SearchProvidersByLocation(ctx, postgres.SearchProvidersByLocationParams{
			Longitude:    *filters.Longitude,
			Latitude:     *filters.Latitude,
			RadiusMeters: radiusMeters,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("search providers by location: %w", err)
		}

		results := make([]domain.ProviderSearchResult, 0, len(rows))
		for _, row := range rows {
			profile := sqlcProviderToDomain(row.ProviderProfile)
			results = append(results, domain.ProviderSearchResult{
				ProviderProfile: *profile,
				UserName:        pgTextToString(row.UserName),
				UserPhone:       row.UserPhone,
				Distance:        row.DistanceMeters / 1000, // convert to km
			})
		}

		log.Info().
			Float64("lat", *filters.Latitude).
			Float64("lng", *filters.Longitude).
			Int("results", len(results)).
			Msg("provider location search completed")

		return results, len(results), nil
	}

	// Postcode-based search.
	if filters.Postcode != nil && *filters.Postcode != "" {
		rows, err := s.queries.SearchProvidersByPostcode(ctx, postgres.SearchProvidersByPostcodeParams{
			Postcode: *filters.Postcode,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("search providers by postcode: %w", err)
		}

		results := make([]domain.ProviderSearchResult, 0, len(rows))
		for _, row := range rows {
			profile := sqlcProviderToDomain(row.ProviderProfile)
			results = append(results, domain.ProviderSearchResult{
				ProviderProfile: *profile,
				UserName:        pgTextToString(row.UserName),
				UserPhone:       row.UserPhone,
			})
		}

		return results, len(results), nil
	}

	// Skill-based search.
	if len(filters.Skills) > 0 {
		rows, err := s.queries.SearchProvidersBySkill(ctx, postgres.SearchProvidersBySkillParams{
			Skills: filters.Skills,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("search providers by skill: %w", err)
		}

		results := make([]domain.ProviderSearchResult, 0, len(rows))
		for _, row := range rows {
			profile := sqlcProviderToDomain(row.ProviderProfile)
			results = append(results, domain.ProviderSearchResult{
				ProviderProfile: *profile,
				UserName:        pgTextToString(row.UserName),
				UserPhone:       row.UserPhone,
			})
		}

		return results, len(results), nil
	}

	// Fallback: no specific filter. Return empty result rather than scanning
	// entire table.
	log.Debug().Msg("search providers called with no filters")
	return []domain.ProviderSearchResult{}, 0, nil
}

// SearchJobs finds jobs matching the given filters using raw SQL to support
// the full set of filter combinations.
func (s *SearchService) SearchJobs(ctx context.Context, filters domain.JobSearchFilters) ([]domain.Job, int, error) {
	query := `SELECT id, customer_id, provider_id, category_id, description,
		postcode, status, scheduled_at, quoted_price, final_price, currency,
		payment_method, is_recurring, jurisdiction_id, created_at, updated_at
		FROM jobs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filters.Status))
		argIdx++
	}
	if filters.CategoryID != nil {
		query += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, *filters.CategoryID)
		argIdx++
	}
	if filters.Postcode != nil && *filters.Postcode != "" {
		query += fmt.Sprintf(" AND postcode = $%d", argIdx)
		args = append(args, *filters.Postcode)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	lim := 20
	if filters.Limit > 0 {
		lim = filters.Limit
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, lim)
	argIdx++

	off := 0
	if filters.Offset > 0 {
		off = filters.Offset
	}
	query += fmt.Sprintf(" OFFSET $%d", argIdx)
	args = append(args, off)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		var description, postcode pgtype.Text
		var scheduledAt pgtype.Timestamptz
		var quotedPrice, finalPrice pgtype.Numeric

		if err := rows.Scan(
			&j.ID, &j.CustomerID, &j.ProviderID, &j.CategoryID,
			&description,
			&postcode, &j.Status, &scheduledAt, &quotedPrice, &finalPrice,
			&j.Currency, &j.PaymentMethod, &j.IsRecurring,
			&j.JurisdictionID, &j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("failed to scan job search row, skipping")
			continue
		}

		if description.Valid {
			j.Description = description.String
		}
		if postcode.Valid {
			j.Postcode = postcode.String
		}
		if scheduledAt.Valid {
			t := scheduledAt.Time
			j.ScheduledAt = &t
		}
		if quotedPrice.Valid {
			f, _ := quotedPrice.Float64Value()
			if f.Valid {
				v := f.Float64
				j.QuotedPrice = &v
			}
		}
		if finalPrice.Valid {
			f, _ := finalPrice.Float64Value()
			if f.Valid {
				v := f.Float64
				j.FinalPrice = &v
			}
		}

		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("job search rows: %w", err)
	}

	return jobs, len(jobs), nil
}

// SearchCategories searches active categories whose name contains the query
// string (case-insensitive).
func (s *SearchService) SearchCategories(ctx context.Context, query string) ([]domain.Category, error) {
	allCategories, err := s.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	queryLower := strings.ToLower(query)
	var results []domain.Category

	for _, cat := range allCategories {
		// Parse the JSON name map and check if any language matches.
		var nameMap map[string]string
		if err := json.Unmarshal(cat.Name, &nameMap); err != nil {
			continue
		}

		matched := query == "" // empty query matches all
		if !matched {
			for _, name := range nameMap {
				if strings.Contains(strings.ToLower(name), queryLower) {
					matched = true
					break
				}
			}
		}

		if matched {
			results = append(results, sqlcCategoryToDomain(cat))
		}
	}

	return results, nil
}

// GetCategoryTree returns the full category hierarchy with parent-child
// relationships preserved.
func (s *SearchService) GetCategoryTree(ctx context.Context) ([]domain.Category, error) {
	categories, err := s.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	results := make([]domain.Category, len(categories))
	for i, cat := range categories {
		results[i] = sqlcCategoryToDomain(cat)
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func pgTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func pgTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func pgNumericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	if f.Valid {
		return f.Float64
	}
	return 0
}

func pgInt4ToInt(n pgtype.Int4) int {
	if !n.Valid {
		return 0
	}
	return int(n.Int32)
}

func sqlcProviderToDomain(p postgres.ProviderProfile) *domain.ProviderProfile {
	return &domain.ProviderProfile{
		UserID:               p.UserID,
		Skills:               p.Skills,
		ServiceRadiusKM:      pgNumericToFloat(p.ServiceRadiusKM),
		Postcode:             pgTextToString(p.Postcode),
		TrustScore:           pgNumericToFloat(p.TrustScore),
		Level:                int(p.TotalJobsCompleted / 10), // approximate
		VerificationStatus:   domain.VerificationStatus(p.VerificationStatus),
		SubscriptionTier:     domain.SubscriptionTier(p.SubscriptionTier),
		AvailabilitySchedule: p.AvailabilitySchedule,
		BankAccountID:        pgTextToStringPtr(p.BankAccountID),
		JobsCompleted:        int(p.TotalJobsCompleted),
		ResponseTimeAvg:      pgInt4ToInt(p.AvgResponseTimeMinutes),
		Bio:                  pgTextToString(p.Description),
		IsAvailable:          p.IsAvailable,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func sqlcCategoryToDomain(c postgres.Category) domain.Category {
	var nameMap map[string]string
	if err := json.Unmarshal(c.Name, &nameMap); err != nil {
		nameMap = map[string]string{"en": c.Slug}
	}
	return domain.Category{
		ID:              c.ID,
		Slug:            c.Slug,
		Name:            nameMap,
		ParentID:        c.ParentID,
		Icon:            pgTextToString(c.Icon),
		IsActive:        c.IsActive,
		RequiresLicense: c.RequiresLicense,
		Metadata:        c.Metadata,
	}
}
