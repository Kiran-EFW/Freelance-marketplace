package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// RecurringSchedule represents a recurring job schedule.
type RecurringSchedule struct {
	ID               uuid.UUID  `json:"id"`
	CustomerID       uuid.UUID  `json:"customer_id"`
	ProviderID       uuid.UUID  `json:"provider_id"`
	CategoryID       uuid.UUID  `json:"category_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	Frequency        string     `json:"frequency"` // daily, weekly, biweekly, monthly, quarterly
	DayOfWeek        *int       `json:"day_of_week,omitempty"`
	DayOfMonth       *int       `json:"day_of_month,omitempty"`
	PreferredTime    string     `json:"preferred_time"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"` // active, paused, cancelled
	NextOccurrence   *time.Time `json:"next_occurrence,omitempty"`
	LastOccurrence   *time.Time `json:"last_occurrence,omitempty"`
	TotalOccurrences int        `json:"total_occurrences"`
	MaxOccurrences   *int       `json:"max_occurrences,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// RecurringService defines the business operations required by RecurringHandler.
type RecurringService interface {
	Create(ctx context.Context, schedule *RecurringSchedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*RecurringSchedule, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]RecurringSchedule, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]RecurringSchedule, error)
	Update(ctx context.Context, schedule *RecurringSchedule) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RecurringHandler handles recurring schedule endpoints.
type RecurringHandler struct {
	service RecurringService
}

// NewRecurringHandler creates a new RecurringHandler.
func NewRecurringHandler(svc RecurringService) *RecurringHandler {
	return &RecurringHandler{service: svc}
}

// RegisterRoutes mounts recurring schedule routes on the given Fiber router group.
func (h *RecurringHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateSchedule)
	rg.Get("/", h.ListSchedules)
	rg.Get("/:id", h.GetSchedule)
	rg.Put("/:id", h.UpdateSchedule)
	rg.Put("/:id/pause", h.PauseSchedule)
	rg.Put("/:id/resume", h.ResumeSchedule)
	rg.Delete("/:id", h.CancelSchedule)
}

// createScheduleRequest is the payload for POST /api/v1/recurring.
type createScheduleRequest struct {
	ProviderID     string  `json:"provider_id"`
	CategoryID     string  `json:"category_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Frequency      string  `json:"frequency"`
	DayOfWeek      *int    `json:"day_of_week"`
	DayOfMonth     *int    `json:"day_of_month"`
	PreferredTime  string  `json:"preferred_time"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	MaxOccurrences *int    `json:"max_occurrences"`
}

func (r *createScheduleRequest) validate() error {
	if r.ProviderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider_id is required")
	}
	if r.CategoryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category_id is required")
	}
	if r.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	validFrequencies := map[string]bool{"daily": true, "weekly": true, "biweekly": true, "monthly": true, "quarterly": true}
	if !validFrequencies[r.Frequency] {
		return fiber.NewError(fiber.StatusBadRequest, "frequency must be one of: daily, weekly, biweekly, monthly, quarterly")
	}
	if r.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}
	return nil
}

// CreateSchedule creates a new recurring schedule.
// POST /api/v1/recurring
func (h *RecurringHandler) CreateSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createScheduleRequest
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

	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider_id format",
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

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	preferredTime := req.PreferredTime
	if preferredTime == "" {
		preferredTime = "09:00"
	}

	now := time.Now().UTC()
	schedule := &RecurringSchedule{
		ID:             uuid.New(),
		CustomerID:     userID,
		ProviderID:     providerID,
		CategoryID:     categoryID,
		Title:          req.Title,
		Description:    req.Description,
		Frequency:      req.Frequency,
		DayOfWeek:      req.DayOfWeek,
		DayOfMonth:     req.DayOfMonth,
		PreferredTime:  preferredTime,
		Amount:         req.Amount,
		Currency:       currency,
		Status:         "active",
		MaxOccurrences: req.MaxOccurrences,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.service.Create(c.UserContext(), schedule); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create recurring schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create recurring schedule",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": schedule,
	})
}

// ListSchedules lists the authenticated user's recurring schedules.
// GET /api/v1/recurring?role=customer|provider
func (h *RecurringHandler) ListSchedules(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	page, limit := parsePagination(c)
	offset := (page - 1) * limit
	role := c.Query("role", "customer")

	var schedules []RecurringSchedule
	var err error

	if role == "provider" {
		schedules, err = h.service.ListByProvider(c.UserContext(), userID, limit, offset)
	} else {
		schedules, err = h.service.ListByCustomer(c.UserContext(), userID, limit, offset)
	}

	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list recurring schedules")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list recurring schedules",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": schedules,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// GetSchedule returns a recurring schedule by ID.
// GET /api/v1/recurring/:id
func (h *RecurringHandler) GetSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid schedule ID format",
			},
		})
	}

	schedule, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("schedule_id", id.String()).Msg("failed to get recurring schedule")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "recurring schedule not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// updateScheduleRequest is the payload for PUT /api/v1/recurring/:id.
type updateScheduleRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Frequency      string  `json:"frequency"`
	DayOfWeek      *int    `json:"day_of_week"`
	DayOfMonth     *int    `json:"day_of_month"`
	PreferredTime  string  `json:"preferred_time"`
	Amount         float64 `json:"amount"`
	MaxOccurrences *int    `json:"max_occurrences"`
}

// UpdateSchedule updates a recurring schedule.
// PUT /api/v1/recurring/:id
func (h *RecurringHandler) UpdateSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid schedule ID format",
			},
		})
	}

	var req updateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	schedule := &RecurringSchedule{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		Frequency:      req.Frequency,
		DayOfWeek:      req.DayOfWeek,
		DayOfMonth:     req.DayOfMonth,
		PreferredTime:  req.PreferredTime,
		Amount:         req.Amount,
		MaxOccurrences: req.MaxOccurrences,
		UpdatedAt:      time.Now().UTC(),
	}

	if err := h.service.Update(c.UserContext(), schedule); err != nil {
		log.Error().Err(err).Str("schedule_id", id.String()).Msg("failed to update recurring schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update recurring schedule",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// PauseSchedule pauses a recurring schedule.
// PUT /api/v1/recurring/:id/pause
func (h *RecurringHandler) PauseSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid schedule ID format",
			},
		})
	}

	if err := h.service.UpdateStatus(c.UserContext(), id, "paused"); err != nil {
		log.Error().Err(err).Str("schedule_id", id.String()).Msg("failed to pause recurring schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to pause schedule",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":     id,
			"status": "paused",
		},
	})
}

// ResumeSchedule resumes a paused recurring schedule.
// PUT /api/v1/recurring/:id/resume
func (h *RecurringHandler) ResumeSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid schedule ID format",
			},
		})
	}

	if err := h.service.UpdateStatus(c.UserContext(), id, "active"); err != nil {
		log.Error().Err(err).Str("schedule_id", id.String()).Msg("failed to resume recurring schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to resume schedule",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":     id,
			"status": "active",
		},
	})
}

// CancelSchedule cancels a recurring schedule.
// DELETE /api/v1/recurring/:id
func (h *RecurringHandler) CancelSchedule(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid schedule ID format",
			},
		})
	}

	if err := h.service.Delete(c.UserContext(), id); err != nil {
		log.Error().Err(err).Str("schedule_id", id.String()).Msg("failed to cancel recurring schedule")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to cancel schedule",
			},
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
