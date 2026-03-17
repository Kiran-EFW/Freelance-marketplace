package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// CropWork mirrors the service layer CropWork type for the handler responses.
type CropWork struct {
	CropName  string         `json:"crop_name"`
	CropSlug  string         `json:"crop_slug"`
	WorkTypes []CropWorkType `json:"work_types"`
}

// CropWorkType mirrors the service layer WorkType for the handler responses.
type CropWorkType struct {
	Slug         string        `json:"slug"`
	Name         string        `json:"name"`
	PricingModel string        `json:"pricing_model"`
	TypicalPrice CropPriceRange `json:"typical_price"`
	IsInSeason   bool          `json:"is_in_season"`
}

// CropPriceRange represents a price range for a work type.
type CropPriceRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Currency string  `json:"currency"`
}

// CropCatalogEntry mirrors the service layer CropCatalogEntry for handler responses.
type CropCatalogEntry struct {
	CropSlug         string              `json:"crop_slug"`
	Name             map[string]string   `json:"name"`
	WorkTypes        []CropWorkType      `json:"work_types"`
	SeasonalCalendar map[string][]string `json:"seasonal_calendar"`
	IsActive         bool                `json:"is_active"`
}

// CropCalendarService defines the business operations for crop calendar.
type CropCalendarService interface {
	GetSeasonalCalendar(ctx context.Context, jurisdictionID string, month int) ([]CropWork, error)
	GetCropsByJurisdiction(ctx context.Context, jurisdictionID string) ([]CropCatalogEntry, error)
}

// CropHandler handles crop calendar endpoints.
type CropHandler struct {
	service CropCalendarService
}

// NewCropHandler creates a new CropHandler.
func NewCropHandler(svc CropCalendarService) *CropHandler {
	return &CropHandler{service: svc}
}

// RegisterRoutes mounts crop calendar routes on the given Fiber router group.
func (h *CropHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/calendar", h.GetSeasonalCalendar)
	rg.Get("/", h.ListCrops)
}

// GetSeasonalCalendar returns available crop work for a jurisdiction and month.
// GET /api/v1/crops/calendar?jurisdiction=in&month=3
func (h *CropHandler) GetSeasonalCalendar(c *fiber.Ctx) error {
	jurisdiction := c.Query("jurisdiction", "in")

	monthStr := c.Query("month")
	var month int
	if monthStr == "" {
		month = int(time.Now().Month())
	} else {
		var err error
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "month must be a number between 1 and 12",
				},
			})
		}
	}

	calendar, err := h.service.GetSeasonalCalendar(c.UserContext(), jurisdiction, month)
	if err != nil {
		log.Error().Err(err).Str("jurisdiction", jurisdiction).Int("month", month).Msg("failed to get seasonal calendar")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve seasonal calendar",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"month":        month,
			"jurisdiction": jurisdiction,
			"crops":        calendar,
		},
	})
}

// ListCrops returns all crops catalogued for a jurisdiction.
// GET /api/v1/crops?jurisdiction=in
func (h *CropHandler) ListCrops(c *fiber.Ctx) error {
	jurisdiction := c.Query("jurisdiction", "in")

	crops, err := h.service.GetCropsByJurisdiction(c.UserContext(), jurisdiction)
	if err != nil {
		log.Error().Err(err).Str("jurisdiction", jurisdiction).Msg("failed to list crops")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list crops",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": crops,
	})
}
