package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// Route represents a provider's service route with stops.
type Route struct {
	ID         uuid.UUID   `json:"id"`
	ProviderID uuid.UUID   `json:"provider_id"`
	Name       string      `json:"name"`
	DayOfWeek  int         `json:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime  string      `json:"start_time"`  // HH:MM format
	EndTime    string      `json:"end_time"`
	Stops      []RouteStop `json:"stops,omitempty"`
	Status     string      `json:"status"` // active, paused, completed
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// RouteStop represents a single stop on a provider's route.
type RouteStop struct {
	ID          uuid.UUID  `json:"id"`
	RouteID     uuid.UUID  `json:"route_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	Address     string     `json:"address"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Order       int        `json:"order"`
	Duration    int        `json:"estimated_duration_min"`
	Notes       string     `json:"notes,omitempty"`
	Status      string     `json:"status"` // scheduled, completed, skipped
	CreatedAt   time.Time  `json:"created_at"`
}

// WeeklyScheduleEntry represents a single visit in the weekly schedule.
type WeeklyScheduleEntry struct {
	RouteID    uuid.UUID `json:"route_id"`
	RouteName  string    `json:"route_name"`
	DayOfWeek  int       `json:"day_of_week"`
	StartTime  string    `json:"start_time"`
	StopCount  int       `json:"stop_count"`
}

// RouteGap represents a geographic area with demand but no provider coverage.
type RouteGap struct {
	Postcode    string  `json:"postcode"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	DemandCount int     `json:"demand_count"`
}

// OptimizeResult wraps the optimized stops with distance and time metadata.
type OptimizeResult struct {
	Stops          []RouteStop `json:"stops"`
	TotalDistKM    float64     `json:"total_distance_km"`
	EstimatedMins  int         `json:"estimated_time_min"`
}

// RouteService defines the business operations required by RouteHandler.
type RouteService interface {
	Create(ctx context.Context, route *Route) error
	ListByProvider(ctx context.Context, providerID uuid.UUID) ([]Route, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Route, error)
	AddStop(ctx context.Context, stop *RouteStop) error
	RemoveStop(ctx context.Context, routeID, stopID uuid.UUID) error
	OptimizeRoute(ctx context.Context, routeID uuid.UUID) (*OptimizeResult, error)
	GetWeeklySchedule(ctx context.Context, providerID uuid.UUID) ([]WeeklyScheduleEntry, error)
	RequestRouteService(ctx context.Context, customerID uuid.UUID, postcode string, categoryID uuid.UUID, notes string) error
	FindGaps(ctx context.Context, category string, jurisdiction string) ([]RouteGap, error)
}

// RouteHandler handles route management endpoints.
type RouteHandler struct {
	service RouteService
}

// NewRouteHandler creates a new RouteHandler.
func NewRouteHandler(svc RouteService) *RouteHandler {
	return &RouteHandler{service: svc}
}

// RegisterRoutes mounts route management routes on the given Fiber router group.
func (h *RouteHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateRoute)
	rg.Get("/", h.GetMyRoutes)
	rg.Get("/schedule", h.GetWeeklySchedule)
	rg.Get("/gaps", h.FindGaps)
	rg.Post("/request", h.RequestRouteService)
	rg.Get("/:id", h.GetRoute)
	rg.Post("/:id/stops", h.AddStop)
	rg.Delete("/:id/stops/:stopId", h.RemoveStop)
	rg.Post("/:id/optimize", h.OptimizeRoute)
}

// createRouteRequest is the payload for POST /api/v1/routes.
type createRouteRequest struct {
	Name      string `json:"name"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (r *createRouteRequest) validate() error {
	if r.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if r.DayOfWeek < 0 || r.DayOfWeek > 6 {
		return fiber.NewError(fiber.StatusBadRequest, "day_of_week must be between 0 (Sunday) and 6 (Saturday)")
	}
	if r.StartTime == "" || r.EndTime == "" {
		return fiber.NewError(fiber.StatusBadRequest, "start_time and end_time are required (HH:MM format)")
	}
	return nil
}

// CreateRoute creates a new route for the authenticated provider.
// POST /api/v1/routes
func (h *RouteHandler) CreateRoute(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createRouteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	now := time.Now().UTC()
	route := &Route{
		ID:         uuid.New(),
		ProviderID: userID,
		Name:       req.Name,
		DayOfWeek:  req.DayOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.service.Create(c.UserContext(), route); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create route")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create route",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": route,
	})
}

// GetMyRoutes lists all routes for the authenticated provider.
// GET /api/v1/routes
func (h *RouteHandler) GetMyRoutes(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	routes, err := h.service.ListByProvider(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list routes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list routes",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": routes,
	})
}

// GetRoute returns a single route with its stops.
// GET /api/v1/routes/:id
func (h *RouteHandler) GetRoute(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid route ID format",
			},
		})
	}

	route, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("route_id", id.String()).Msg("failed to get route")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve route",
			},
		})
	}

	if route == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "route not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": route,
	})
}

// addStopRequest is the payload for POST /api/v1/routes/:id/stops.
type addStopRequest struct {
	CustomerID string  `json:"customer_id"`
	Address    string  `json:"address"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Duration   int     `json:"estimated_duration_min"`
	Notes      string  `json:"notes"`
}

func (r *addStopRequest) validate() error {
	if r.CustomerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "customer_id is required")
	}
	if r.Address == "" {
		return fiber.NewError(fiber.StatusBadRequest, "address is required")
	}
	return nil
}

// AddStop adds a customer stop to a route.
// POST /api/v1/routes/:id/stops
func (h *RouteHandler) AddStop(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	routeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid route ID format",
			},
		})
	}

	var req addStopRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid customer_id format",
			},
		})
	}

	stop := &RouteStop{
		ID:         uuid.New(),
		RouteID:    routeID,
		CustomerID: customerID,
		Address:    req.Address,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Duration:   req.Duration,
		Notes:      req.Notes,
		Status:     "scheduled",
		CreatedAt:  time.Now().UTC(),
	}

	if err := h.service.AddStop(c.UserContext(), stop); err != nil {
		log.Error().Err(err).Str("route_id", routeID.String()).Msg("failed to add stop")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to add stop to route",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": stop,
	})
}

// RemoveStop removes a stop from a route.
// DELETE /api/v1/routes/:id/stops/:stopId
func (h *RouteHandler) RemoveStop(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	routeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid route ID format",
			},
		})
	}

	stopID, err := uuid.Parse(c.Params("stopId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid stop ID format",
			},
		})
	}

	if err := h.service.RemoveStop(c.UserContext(), routeID, stopID); err != nil {
		log.Error().Err(err).Str("route_id", routeID.String()).Str("stop_id", stopID.String()).Msg("failed to remove stop")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to remove stop",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"message": "stop removed successfully",
		},
	})
}

// OptimizeRoute reorders stops for optimal routing.
// POST /api/v1/routes/:id/optimize
func (h *RouteHandler) OptimizeRoute(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	routeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid route ID format",
			},
		})
	}

	optimized, err := h.service.OptimizeRoute(c.UserContext(), routeID)
	if err != nil {
		log.Error().Err(err).Str("route_id", routeID.String()).Msg("failed to optimize route")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to optimize route",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": optimized,
	})
}

// GetWeeklySchedule returns the provider's weekly visit schedule.
// GET /api/v1/routes/schedule
func (h *RouteHandler) GetWeeklySchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	schedule, err := h.service.GetWeeklySchedule(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get weekly schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve weekly schedule",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// requestRouteServiceRequest is the payload for POST /api/v1/routes/request.
type requestRouteServiceRequest struct {
	Postcode   string `json:"postcode"`
	CategoryID string `json:"category_id"`
	Notes      string `json:"notes"`
}

func (r *requestRouteServiceRequest) validate() error {
	if r.Postcode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "postcode is required")
	}
	if r.CategoryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category_id is required")
	}
	return nil
}

// RequestRouteService allows a customer to request being added to a provider's route.
// POST /api/v1/routes/request
func (h *RouteHandler) RequestRouteService(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req requestRouteServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid category_id format",
			},
		})
	}

	if err := h.service.RequestRouteService(c.UserContext(), userID, req.Postcode, categoryID, req.Notes); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to request route service")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to submit route service request",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"message": "route service request submitted successfully",
		},
	})
}

// FindGaps returns areas where demand exists but no provider coverage is available.
// GET /api/v1/routes/gaps?category=plumbing&jurisdiction=in
func (h *RouteHandler) FindGaps(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	category := c.Query("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "category query parameter is required",
			},
		})
	}

	jurisdiction := c.Query("jurisdiction", "in")

	gaps, err := h.service.FindGaps(c.UserContext(), category, jurisdiction)
	if err != nil {
		log.Error().Err(err).Str("category", category).Msg("failed to find service gaps")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to find service gaps",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"gaps":         gaps,
			"category":     category,
			"jurisdiction": jurisdiction,
		},
	})
}
