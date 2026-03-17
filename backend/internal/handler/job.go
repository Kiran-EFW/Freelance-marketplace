package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/middleware"
)

// Quote represents a provider's quote on a job.
type Quote struct {
	ID         uuid.UUID  `json:"id"`
	JobID      uuid.UUID  `json:"job_id"`
	ProviderID uuid.UUID  `json:"provider_id"`
	Amount     float64    `json:"amount"`
	Currency   string     `json:"currency"`
	Message    string     `json:"message"`
	Status     string     `json:"status"` // pending, accepted, rejected
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// JobService defines the business operations required by JobHandler.
type JobService interface {
	Create(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]domain.Job, int, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]domain.Job, int, error)
	ListByStatus(ctx context.Context, userID uuid.UUID, role string, status *domain.JobStatus, limit, offset int) ([]domain.Job, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status domain.JobStatus) error
	SubmitQuote(ctx context.Context, quote *Quote) error
	ListQuotes(ctx context.Context, jobID uuid.UUID) ([]Quote, error)
	AcceptQuote(ctx context.Context, jobID, quoteID, customerID uuid.UUID) error
}

// JobHandler handles job lifecycle endpoints.
type JobHandler struct {
	service JobService
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(svc JobService) *JobHandler {
	return &JobHandler{service: svc}
}

// RegisterRoutes mounts job routes on the given Fiber router group.
func (h *JobHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateJob)
	rg.Get("/", h.ListMyJobs)
	rg.Get("/:id", h.GetJob)
	rg.Patch("/:id/status", h.UpdateJobStatus)
	rg.Post("/:id/quotes", h.SubmitQuote)
	rg.Get("/:id/quotes", h.ListQuotes)
	rg.Post("/:id/quotes/:quoteId/accept", h.AcceptQuote)
}

// createJobRequest is the payload for POST /api/v1/jobs.
type createJobRequest struct {
	CategoryID    string   `json:"category_id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Postcode      string   `json:"postcode"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	ScheduledAt   *string  `json:"scheduled_at"`
	BudgetMin     *float64 `json:"budget_min"`
	BudgetMax     *float64 `json:"budget_max"`
	PaymentMethod string   `json:"payment_method"`
	Photos        []string `json:"photos"`
}

func (r *createJobRequest) validate() error {
	if r.CategoryID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category_id is required")
	}
	if r.Description == "" {
		return fiber.NewError(fiber.StatusBadRequest, "description is required")
	}
	if r.Postcode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "postcode is required")
	}
	return nil
}

// CreateJob creates a new job request.
// POST /api/v1/jobs
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createJobRequest
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

	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_DATE",
					"message": "scheduled_at must be RFC3339 format",
				},
			})
		}
		scheduledAt = &t
	}

	paymentMethod := domain.PaymentMethodOnline
	switch req.PaymentMethod {
	case "cash":
		paymentMethod = domain.PaymentMethodCash
	case "wallet":
		paymentMethod = domain.PaymentMethodWallet
	case "online", "":
		paymentMethod = domain.PaymentMethodOnline
	}

	now := time.Now().UTC()
	job := &domain.Job{
		ID:            uuid.New(),
		CustomerID:    userID,
		CategoryID:    categoryID,
		Postcode:      req.Postcode,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Status:        domain.JobStatusPosted,
		Description:   req.Description,
		ScheduledAt:   scheduledAt,
		Currency:      "INR",
		PaymentMethod: paymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.service.Create(c.UserContext(), job); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create job")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create job",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": job,
	})
}

// GetJob returns a single job by ID.
// GET /api/v1/jobs/:id
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	job, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("job_id", idStr).Msg("failed to get job")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve job",
			},
		})
	}

	if job == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "job not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": job,
	})
}

// ListMyJobs lists jobs for the authenticated user.
// GET /api/v1/jobs?role=customer|provider&status=posted&page=1&limit=20
func (h *JobHandler) ListMyJobs(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	role := c.Query("role", "customer")
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	var statusFilter *domain.JobStatus
	if s := c.Query("status"); s != "" {
		st := domain.JobStatus(s)
		statusFilter = &st
	}

	jobs, total, err := h.service.ListByStatus(c.UserContext(), userID, role, statusFilter, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list jobs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list jobs",
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

// updateJobStatusRequest is the payload for PATCH /api/v1/jobs/:id/status.
type updateJobStatusRequest struct {
	Status string `json:"status"`
}

func (r *updateJobStatusRequest) validate() error {
	validStatuses := map[string]bool{
		"accepted":    true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatuses[r.Status] {
		return fiber.NewError(fiber.StatusBadRequest, "status must be one of: accepted, in_progress, completed, cancelled")
	}
	return nil
}

// UpdateJobStatus changes the status of a job.
// PATCH /api/v1/jobs/:id/status
func (h *JobHandler) UpdateJobStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	var req updateJobStatusRequest
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

	if err := h.service.UpdateStatus(c.UserContext(), jobID, userID, domain.JobStatus(req.Status)); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Str("status", req.Status).Msg("failed to update job status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update job status",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":     jobID,
			"status": req.Status,
		},
	})
}

// submitQuoteRequest is the payload for POST /api/v1/jobs/:id/quotes.
type submitQuoteRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Message  string  `json:"message"`
}

func (r *submitQuoteRequest) validate() error {
	if r.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}
	return nil
}

// SubmitQuote allows a provider to submit a quote for a job.
// POST /api/v1/jobs/:id/quotes
func (h *JobHandler) SubmitQuote(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	var req submitQuoteRequest
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

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	now := time.Now().UTC()
	quote := &Quote{
		ID:         uuid.New(),
		JobID:      jobID,
		ProviderID: userID,
		Amount:     req.Amount,
		Currency:   currency,
		Message:    req.Message,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.service.SubmitQuote(c.UserContext(), quote); err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to submit quote")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to submit quote",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": quote,
	})
}

// ListQuotes returns all quotes for a given job.
// GET /api/v1/jobs/:id/quotes
func (h *JobHandler) ListQuotes(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	quotes, err := h.service.ListQuotes(c.UserContext(), jobID)
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to list quotes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list quotes",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": quotes,
	})
}

// AcceptQuote allows a customer to accept a quote on their job.
// POST /api/v1/jobs/:id/quotes/:quoteId/accept
func (h *JobHandler) AcceptQuote(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	quoteID, err := uuid.Parse(c.Params("quoteId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid quote ID format",
			},
		})
	}

	if err := h.service.AcceptQuote(c.UserContext(), jobID, quoteID, userID); err != nil {
		log.Error().Err(err).
			Str("job_id", jobID.String()).
			Str("quote_id", quoteID.String()).
			Msg("failed to accept quote")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to accept quote",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"job_id":   jobID,
			"quote_id": quoteID,
			"status":   "accepted",
		},
	})
}

// parsePagination extracts page and limit from query parameters with defaults.
func parsePagination(c *fiber.Ctx) (page, limit int) {
	page = 1
	limit = 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	// Cap limit at 100.
	if limit > 100 {
		limit = 100
	}

	return page, limit
}
