package handler

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// ---------------------------------------------------------------------------
// Analytics types
// ---------------------------------------------------------------------------

// EarningsHistoryEntry represents one month of earnings data.
type EarningsHistoryEntry struct {
	Month    time.Time `json:"month"`
	Earnings float64   `json:"earnings"`
	JobCount int       `json:"job_count"`
}

// DemandByCategoryEntry represents demand for a single category.
type DemandByCategoryEntry struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategorySlug string    `json:"category_slug"`
	CategoryName string    `json:"category_name"`
	DemandCount  int       `json:"demand_count"`
}

// DemandByPostcodeEntry represents demand for a single postcode (heatmap).
type DemandByPostcodeEntry struct {
	Postcode    string  `json:"postcode"`
	DemandCount int     `json:"demand_count"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

// PerformanceMetrics holds the provider's key performance metrics.
type PerformanceMetrics struct {
	ResponseRate   float64 `json:"response_rate"`
	CompletionRate float64 `json:"completion_rate"`
	AvgRating      float64 `json:"avg_rating"`
	TotalReviews   int     `json:"total_reviews"`
	TotalEarnings  float64 `json:"total_earnings"`
}

// PeakDemandHourEntry represents demand for one hour of the day.
type PeakDemandHourEntry struct {
	HourOfDay   int `json:"hour_of_day"`
	DemandCount int `json:"demand_count"`
}

// CompetitorDensityEntry represents the count of providers per category in a postcode.
type CompetitorDensityEntry struct {
	Postcode      string `json:"postcode"`
	CategorySlug  string `json:"category_slug"`
	CategoryName  string `json:"category_name"`
	ProviderCount int    `json:"provider_count"`
}

// AnalyticsInsight represents a single actionable insight.
type AnalyticsInsight struct {
	Type    string `json:"type"`    // "opportunity", "performance", "trend"
	Title   string `json:"title"`
	Message string `json:"message"`
	Impact  string `json:"impact"` // "high", "medium", "low"
}

// ---------------------------------------------------------------------------
// Analytics service interface
// ---------------------------------------------------------------------------

// AnalyticsService defines the business operations for analytics.
type AnalyticsService interface {
	GetEarningsHistory(ctx context.Context, providerID uuid.UUID, months int) ([]EarningsHistoryEntry, error)
	GetDemandByCategory(ctx context.Context, lng, lat, radiusMeters float64) ([]DemandByCategoryEntry, error)
	GetDemandByPostcode(ctx context.Context, lng, lat, radiusMeters float64) ([]DemandByPostcodeEntry, error)
	GetPerformanceMetrics(ctx context.Context, providerID uuid.UUID) (*PerformanceMetrics, error)
	GetPeakDemandHours(ctx context.Context, providerID uuid.UUID) ([]PeakDemandHourEntry, error)
	GetCompetitorDensity(ctx context.Context, providerID uuid.UUID) ([]CompetitorDensityEntry, error)
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// AnalyticsHandler handles analytics dashboard endpoints.
type AnalyticsHandler struct {
	service AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(svc AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: svc}
}

// RegisterRoutes mounts analytics routes on the given Fiber router group.
func (h *AnalyticsHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/earnings", h.GetEarnings)
	rg.Get("/demand", h.GetDemandHeatmap)
	rg.Get("/performance", h.GetPerformance)
	rg.Get("/peak-hours", h.GetPeakHours)
	rg.Get("/competitors", h.GetCompetitors)
	rg.Get("/insights", h.GetInsights)
}

// GetEarnings returns monthly earnings chart data.
// GET /api/v1/analytics/earnings?period=12m
func (h *AnalyticsHandler) GetEarnings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	months := 12
	if period := c.Query("period"); period != "" {
		switch period {
		case "7d":
			months = 1
		case "30d":
			months = 1
		case "90d":
			months = 3
		case "12m":
			months = 12
		default:
			if v, err := strconv.Atoi(period); err == nil && v > 0 && v <= 24 {
				months = v
			}
		}
	}

	history, err := h.service.GetEarningsHistory(c.UserContext(), userID, months)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get earnings history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve earnings history",
			},
		})
	}

	if history == nil {
		history = []EarningsHistoryEntry{}
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"history": history,
			"period":  months,
		},
	})
}

// GetDemandHeatmap returns postcode-level demand data for heatmap display.
// GET /api/v1/analytics/demand?lat=X&lng=Y&radius=25
func (h *AnalyticsHandler) GetDemandHeatmap(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	lat, err := strconv.ParseFloat(c.Query("lat", "12.9716"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "invalid lat parameter",
			},
		})
	}

	lng, err := strconv.ParseFloat(c.Query("lng", "77.5946"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "invalid lng parameter",
			},
		})
	}

	radiusKM, err := strconv.ParseFloat(c.Query("radius", "25"), 64)
	if err != nil || radiusKM <= 0 {
		radiusKM = 25
	}
	if radiusKM > 100 {
		radiusKM = 100
	}
	radiusMeters := radiusKM * 1000

	// Get both postcode demand and category demand in parallel
	postcodes, err := h.service.GetDemandByPostcode(c.UserContext(), lng, lat, radiusMeters)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get demand by postcode")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve demand heatmap data",
			},
		})
	}

	categories, err := h.service.GetDemandByCategory(c.UserContext(), lng, lat, radiusMeters)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get demand by category")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve demand data",
			},
		})
	}

	if postcodes == nil {
		postcodes = []DemandByPostcodeEntry{}
	}
	if categories == nil {
		categories = []DemandByCategoryEntry{}
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"postcodes":  postcodes,
			"categories": categories,
			"center":     fiber.Map{"lat": lat, "lng": lng},
			"radius_km":  radiusKM,
		},
	})
}

// GetPerformance returns provider performance metrics.
// GET /api/v1/analytics/performance
func (h *AnalyticsHandler) GetPerformance(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	metrics, err := h.service.GetPerformanceMetrics(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get performance metrics")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve performance metrics",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": metrics,
	})
}

// GetPeakHours returns demand by hour of day.
// GET /api/v1/analytics/peak-hours
func (h *AnalyticsHandler) GetPeakHours(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	hours, err := h.service.GetPeakDemandHours(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get peak demand hours")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve peak demand hours",
			},
		})
	}

	if hours == nil {
		hours = []PeakDemandHourEntry{}
	}

	return c.JSON(fiber.Map{
		"data": hours,
	})
}

// GetCompetitors returns competitor density data.
// GET /api/v1/analytics/competitors
func (h *AnalyticsHandler) GetCompetitors(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	density, err := h.service.GetCompetitorDensity(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get competitor density")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve competitor density",
			},
		})
	}

	if density == nil {
		density = []CompetitorDensityEntry{}
	}

	return c.JSON(fiber.Map{
		"data": density,
	})
}

// GetInsights returns AI-generated actionable insights by combining metrics.
// GET /api/v1/analytics/insights
func (h *AnalyticsHandler) GetInsights(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	// Gather metrics to generate insights from
	metrics, err := h.service.GetPerformanceMetrics(c.UserContext(), userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to get metrics for insights")
		metrics = &PerformanceMetrics{}
	}

	peakHours, err := h.service.GetPeakDemandHours(c.UserContext(), userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to get peak hours for insights")
	}

	competitors, err := h.service.GetCompetitorDensity(c.UserContext(), userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to get competitors for insights")
	}

	insights := generateInsights(metrics, peakHours, competitors)

	return c.JSON(fiber.Map{
		"data": insights,
	})
}

// generateInsights produces actionable suggestions from metrics.
func generateInsights(
	metrics *PerformanceMetrics,
	peakHours []PeakDemandHourEntry,
	competitors []CompetitorDensityEntry,
) []AnalyticsInsight {
	var insights []AnalyticsInsight

	// Response rate insights
	if metrics != nil && metrics.ResponseRate < 0.5 {
		insights = append(insights, AnalyticsInsight{
			Type:    "performance",
			Title:   "Improve Your Response Rate",
			Message: fmt.Sprintf("Your response rate is %.0f%%. Responding to more job requests can increase your visibility and bookings. Aim for at least 70%%.", metrics.ResponseRate*100),
			Impact:  "high",
		})
	}

	// Completion rate insights
	if metrics != nil && metrics.CompletionRate > 0 && metrics.CompletionRate < 0.8 {
		insights = append(insights, AnalyticsInsight{
			Type:    "performance",
			Title:   "Boost Your Completion Rate",
			Message: fmt.Sprintf("Your completion rate is %.0f%%. Completing more accepted jobs builds trust and improves your ranking.", metrics.CompletionRate*100),
			Impact:  "high",
		})
	}

	// Rating insights
	if metrics != nil && metrics.AvgRating > 0 && metrics.AvgRating < 4.0 {
		insights = append(insights, AnalyticsInsight{
			Type:    "performance",
			Title:   "Improve Your Rating",
			Message: fmt.Sprintf("Your average rating is %.1f. Focus on communication and quality to increase customer satisfaction.", metrics.AvgRating),
			Impact:  "medium",
		})
	} else if metrics != nil && metrics.AvgRating >= 4.5 {
		insights = append(insights, AnalyticsInsight{
			Type:    "performance",
			Title:   "Excellent Rating",
			Message: fmt.Sprintf("Your %.1f star rating is outstanding! Keep up the great work to maintain your top-provider status.", metrics.AvgRating),
			Impact:  "low",
		})
	}

	// Peak hours insight
	if len(peakHours) > 0 {
		var peakHour PeakDemandHourEntry
		maxDemand := 0
		for _, h := range peakHours {
			if h.DemandCount > maxDemand {
				maxDemand = h.DemandCount
				peakHour = h
			}
		}
		if maxDemand > 0 {
			insights = append(insights, AnalyticsInsight{
				Type:    "opportunity",
				Title:   "Peak Demand Time",
				Message: fmt.Sprintf("Most job requests come in at %d:00. Being available and responsive during this time can help you win more jobs.", peakHour.HourOfDay),
				Impact:  "medium",
			})
		}
	}

	// Low competition areas
	if len(competitors) > 0 {
		// Find areas with low provider counts
		for _, comp := range competitors {
			if comp.ProviderCount <= 2 {
				insights = append(insights, AnalyticsInsight{
					Type:    "opportunity",
					Title:   fmt.Sprintf("Low Competition in %s", comp.Postcode),
					Message: fmt.Sprintf("Only %d providers offer %s in %s. Expanding your service area here could win you more jobs.", comp.ProviderCount, comp.CategoryName, comp.Postcode),
					Impact:  "high",
				})
				break // Only show one such insight
			}
		}

		// High competition warning
		for _, comp := range competitors {
			if comp.ProviderCount >= 10 {
				insights = append(insights, AnalyticsInsight{
					Type:    "trend",
					Title:   fmt.Sprintf("High Competition in %s", comp.Postcode),
					Message: fmt.Sprintf("%d providers offer %s in %s. Consider differentiating with faster response times or competitive pricing.", comp.ProviderCount, comp.CategoryName, comp.Postcode),
					Impact:  "medium",
				})
				break
			}
		}
	}

	// Earnings growth insight
	if metrics != nil && metrics.TotalEarnings > 0 {
		insights = append(insights, AnalyticsInsight{
			Type:    "trend",
			Title:   "Earnings Summary",
			Message: fmt.Sprintf("You have earned Rs. %s in total. Keep your profile active and respond quickly to maximize earnings.", formatCurrency(metrics.TotalEarnings)),
			Impact:  "low",
		})
	}

	// Always provide at least one insight
	if len(insights) == 0 {
		insights = append(insights, AnalyticsInsight{
			Type:    "trend",
			Title:   "Getting Started",
			Message: "Complete your profile and respond to job requests to start building your analytics data.",
			Impact:  "medium",
		})
	}

	return insights
}

// formatCurrency formats a number as a human-readable currency string.
func formatCurrency(amount float64) string {
	if amount >= 100000 {
		return fmt.Sprintf("%.1fL", amount/100000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("%.1fK", amount/1000)
	}
	return fmt.Sprintf("%.0f", math.Round(amount))
}
