package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// EscrowTransaction represents an escrow payment held until job completion.
type EscrowTransaction struct {
	ID               uuid.UUID  `json:"id"`
	JobID            uuid.UUID  `json:"job_id"`
	CustomerID       uuid.UUID  `json:"customer_id"`
	ProviderID       uuid.UUID  `json:"provider_id"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"` // held, released, refunded, disputed
	GatewayPaymentID *string    `json:"gateway_payment_id,omitempty"`
	HeldAt           time.Time  `json:"held_at"`
	ReleasedAt       *time.Time `json:"released_at,omitempty"`
	RefundedAt       *time.Time `json:"refunded_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// EscrowService defines the business operations required by EscrowHandler.
type EscrowService interface {
	Create(ctx context.Context, escrow *EscrowTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*EscrowTransaction, error)
	GetByJobID(ctx context.Context, jobID uuid.UUID) (*EscrowTransaction, error)
	Release(ctx context.Context, id uuid.UUID) (*EscrowTransaction, error)
	Refund(ctx context.Context, id uuid.UUID) (*EscrowTransaction, error)
	Dispute(ctx context.Context, id uuid.UUID) (*EscrowTransaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]EscrowTransaction, error)
}

// EscrowHandler handles escrow payment endpoints.
type EscrowHandler struct {
	service EscrowService
}

// NewEscrowHandler creates a new EscrowHandler.
func NewEscrowHandler(svc EscrowService) *EscrowHandler {
	return &EscrowHandler{service: svc}
}

// RegisterRoutes mounts escrow routes on the given Fiber router group.
func (h *EscrowHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateEscrow)
	rg.Get("/", h.ListMyEscrow)
	rg.Get("/:jobId", h.GetEscrowByJob)
	rg.Post("/:id/release", h.ReleaseEscrow)
	rg.Post("/:id/refund", h.RefundEscrow)
	rg.Post("/:id/dispute", h.DisputeEscrow)
}

// createEscrowRequest is the payload for POST /api/v1/escrow.
type createEscrowRequest struct {
	JobID            string  `json:"job_id"`
	ProviderID       string  `json:"provider_id"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	GatewayPaymentID string  `json:"gateway_payment_id"`
}

func (r *createEscrowRequest) validate() error {
	if r.JobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}
	if r.ProviderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "provider_id is required")
	}
	if r.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}
	return nil
}

// CreateEscrow creates a new escrow transaction for a job.
// POST /api/v1/escrow
func (h *EscrowHandler) CreateEscrow(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createEscrowRequest
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

	jobID, err := uuid.Parse(req.JobID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job_id format",
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

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	var gatewayID *string
	if req.GatewayPaymentID != "" {
		gatewayID = &req.GatewayPaymentID
	}

	now := time.Now().UTC()
	escrow := &EscrowTransaction{
		ID:               uuid.New(),
		JobID:            jobID,
		CustomerID:       userID,
		ProviderID:       providerID,
		Amount:           req.Amount,
		Currency:         currency,
		Status:           "held",
		GatewayPaymentID: gatewayID,
		HeldAt:           now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.service.Create(c.UserContext(), escrow); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("job_id", jobID.String()).Msg("failed to create escrow")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create escrow transaction",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": escrow,
	})
}

// GetEscrowByJob returns the escrow transaction for a given job.
// GET /api/v1/escrow/:jobId
func (h *EscrowHandler) GetEscrowByJob(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid job ID format",
			},
		})
	}

	escrow, err := h.service.GetByJobID(c.UserContext(), jobID)
	if err != nil {
		log.Error().Err(err).Str("job_id", jobID.String()).Msg("failed to get escrow by job")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve escrow transaction",
			},
		})
	}

	if escrow == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "escrow transaction not found for this job",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": escrow,
	})
}

// ListMyEscrow lists the authenticated user's escrow transactions.
// GET /api/v1/escrow?page=1&limit=20
func (h *EscrowHandler) ListMyEscrow(c *fiber.Ctx) error {
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

	escrows, err := h.service.ListByUser(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list escrow transactions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list escrow transactions",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": escrows,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// ReleaseEscrow releases funds to the provider upon job completion.
// POST /api/v1/escrow/:id/release
func (h *EscrowHandler) ReleaseEscrow(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	escrowID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid escrow ID format",
			},
		})
	}

	escrow, err := h.service.Release(c.UserContext(), escrowID)
	if err != nil {
		log.Error().Err(err).Str("escrow_id", escrowID.String()).Msg("failed to release escrow")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to release escrow funds",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": escrow,
	})
}

// RefundEscrow refunds the escrow funds to the customer.
// POST /api/v1/escrow/:id/refund
func (h *EscrowHandler) RefundEscrow(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	escrowID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid escrow ID format",
			},
		})
	}

	escrow, err := h.service.Refund(c.UserContext(), escrowID)
	if err != nil {
		log.Error().Err(err).Str("escrow_id", escrowID.String()).Msg("failed to refund escrow")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to refund escrow",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": escrow,
	})
}

// DisputeEscrow marks an escrow transaction as disputed.
// POST /api/v1/escrow/:id/dispute
func (h *EscrowHandler) DisputeEscrow(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	escrowID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid escrow ID format",
			},
		})
	}

	escrow, err := h.service.Dispute(c.UserContext(), escrowID)
	if err != nil {
		log.Error().Err(err).Str("escrow_id", escrowID.String()).Msg("failed to dispute escrow")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to mark escrow as disputed",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": escrow,
	})
}
