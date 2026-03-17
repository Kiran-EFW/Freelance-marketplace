package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
)

// ProviderSearchResult represents a provider result with distance and rating info.
type ProviderSearchResult struct {
	UserID        uuid.UUID `json:"user_id"`
	Name          string    `json:"name"`
	Skills        []string  `json:"skills"`
	Postcode      string    `json:"postcode"`
	DistanceKM    float64   `json:"distance_km"`
	AverageRating float64   `json:"average_rating"`
	TotalReviews  int       `json:"total_reviews"`
	TrustScore    float64   `json:"trust_score"`
	IsOnline      bool      `json:"is_online"`
	Level         int       `json:"level"`
}

// SearchService defines the business operations required by SearchHandler.
type SearchService interface {
	SearchProviders(ctx context.Context, filters domain.ProviderSearchFilters, sortBy string) ([]ProviderSearchResult, int, error)
	SearchJobs(ctx context.Context, filters domain.JobSearchFilters) ([]domain.Job, int, error)
	SearchCategories(ctx context.Context, query string) ([]domain.Category, error)
	GetCategoryTree(ctx context.Context) ([]domain.Category, error)
}

// SearchHandler handles search-related endpoints.
type SearchHandler struct {
	service SearchService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(svc SearchService) *SearchHandler {
	return &SearchHandler{service: svc}
}

// RegisterRoutes mounts search routes on the given Fiber router groups.
func (h *SearchHandler) RegisterSearchRoutes(rg fiber.Router) {
	rg.Get("/providers", h.SearchProviders)
	rg.Get("/jobs", h.SearchJobs)
	rg.Get("/categories", h.SearchCategories)
}

// RegisterCategoryRoutes mounts category routes on the given Fiber router group.
func (h *SearchHandler) RegisterCategoryRoutes(rg fiber.Router) {
	rg.Get("/", h.GetCategoryTree)
}

// SearchProviders searches for providers based on filters.
// GET /api/v1/search/providers?postcode=&category=&lat=&lng=&radius_km=&min_rating=&sort_by=&page=&limit=
func (h *SearchHandler) SearchProviders(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	filters := domain.ProviderSearchFilters{
		Limit:  limit,
		Offset: offset,
	}

	if v := c.Query("postcode"); v != "" {
		filters.Postcode = &v
	}
	if v := c.Query("category"); v != "" {
		filters.CategorySlug = &v
	}
	if v := c.Query("lat"); v != "" {
		if lat, err := strconv.ParseFloat(v, 64); err == nil {
			filters.Latitude = &lat
		}
	}
	if v := c.Query("lng"); v != "" {
		if lng, err := strconv.ParseFloat(v, 64); err == nil {
			filters.Longitude = &lng
		}
	}
	if v := c.Query("radius_km"); v != "" {
		if r, err := strconv.ParseFloat(v, 64); err == nil {
			filters.RadiusKM = &r
		}
	}
	if v := c.Query("min_rating"); v != "" {
		if r, err := strconv.ParseFloat(v, 64); err == nil {
			filters.MinTrustScore = &r
		}
	}
	if c.Query("verified_only") == "true" {
		filters.VerificationOnly = true
	}

	sortBy := c.Query("sort_by", "distance")

	results, total, err := h.service.SearchProviders(c.UserContext(), filters, sortBy)
	if err != nil {
		log.Error().Err(err).Msg("failed to search providers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to search providers",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": results,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// SearchJobs searches for available jobs (for providers looking for work).
// GET /api/v1/search/jobs?postcode=&category_id=&lat=&lng=&radius_km=&status=&page=&limit=
func (h *SearchHandler) SearchJobs(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	filters := domain.JobSearchFilters{
		Limit:  limit,
		Offset: offset,
	}

	if v := c.Query("postcode"); v != "" {
		filters.Postcode = &v
	}
	if v := c.Query("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filters.CategoryID = &id
		}
	}
	if v := c.Query("lat"); v != "" {
		if lat, err := strconv.ParseFloat(v, 64); err == nil {
			filters.Latitude = &lat
		}
	}
	if v := c.Query("lng"); v != "" {
		if lng, err := strconv.ParseFloat(v, 64); err == nil {
			filters.Longitude = &lng
		}
	}
	if v := c.Query("radius_km"); v != "" {
		if r, err := strconv.ParseFloat(v, 64); err == nil {
			filters.RadiusKM = &r
		}
	}
	if v := c.Query("status"); v != "" {
		st := domain.JobStatus(v)
		filters.Status = &st
	}

	jobs, total, err := h.service.SearchJobs(c.UserContext(), filters)
	if err != nil {
		log.Error().Err(err).Msg("failed to search jobs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to search jobs",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": jobs,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// SearchCategories searches categories with autocomplete.
// GET /api/v1/search/categories?q=plumb
func (h *SearchHandler) SearchCategories(c *fiber.Ctx) error {
	query := c.Query("q", "")

	categories, err := h.service.SearchCategories(c.UserContext(), query)
	if err != nil {
		log.Error().Err(err).Msg("failed to search categories")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to search categories",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": categories,
	})
}

// GetCategoryTree returns the full category hierarchy.
// GET /api/v1/categories
func (h *SearchHandler) GetCategoryTree(c *fiber.Ctx) error {
	categories, err := h.service.GetCategoryTree(c.UserContext())
	if err != nil {
		log.Error().Err(err).Msg("failed to get category tree")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve categories",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": categories,
	})
}
