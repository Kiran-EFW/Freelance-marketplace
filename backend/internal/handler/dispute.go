package handler

import (
	"context"
	"fmt"
	stdmime "mime"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/storage"
	"github.com/seva-platform/backend/internal/middleware"
)

// Dispute represents a dispute raised on a job.
type Dispute struct {
	ID          uuid.UUID `json:"id"`
	JobID       uuid.UUID `json:"job_id"`
	RaisedBy    uuid.UUID `json:"raised_by"`
	Type        string    `json:"type"` // quality, no_show, pricing, other
	Description string    `json:"description"`
	Status      string    `json:"status"` // open, in_review, resolved, closed
	Resolution  *string   `json:"resolution,omitempty"`
	ResolvedBy  *uuid.UUID `json:"resolved_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DisputeEvidence represents evidence attached to a dispute.
type DisputeEvidence struct {
	ID        uuid.UUID `json:"id"`
	DisputeID uuid.UUID `json:"dispute_id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"` // photo, document, text
	FileURL   *string   `json:"file_url,omitempty"`
	Text      *string   `json:"text,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DisputeService defines the business operations required by DisputeHandler.
type DisputeService interface {
	Create(ctx context.Context, dispute *Dispute) error
	GetByID(ctx context.Context, id uuid.UUID) (*Dispute, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Dispute, int, error)
	AddEvidence(ctx context.Context, evidence *DisputeEvidence) error
	Respond(ctx context.Context, disputeID, userID uuid.UUID, response string) error
	Resolve(ctx context.Context, disputeID, resolvedBy uuid.UUID, resolution string) error
}

// DisputeHandler handles dispute endpoints.
type DisputeHandler struct {
	service DisputeService
	storage storage.StorageProvider
}

// NewDisputeHandler creates a new DisputeHandler.
func NewDisputeHandler(svc DisputeService, store storage.StorageProvider) *DisputeHandler {
	return &DisputeHandler{service: svc, storage: store}
}

// RegisterRoutes mounts dispute routes on the given Fiber router group.
func (h *DisputeHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateDispute)
	rg.Get("/", h.ListMyDisputes)
	rg.Get("/:id", h.GetDispute)
	rg.Post("/:id/evidence", h.AddEvidence)
	rg.Post("/:id/respond", h.RespondToDispute)
}

// RegisterAdminRoutes mounts admin dispute routes.
func (h *DisputeHandler) RegisterAdminRoutes(rg fiber.Router) {
	rg.Post("/:id/resolve", h.AdminResolveDispute)
}

// createDisputeRequest is the payload for POST /api/v1/disputes.
type createDisputeRequest struct {
	JobID       string   `json:"job_id"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"`
}

func (r *createDisputeRequest) validate() error {
	if r.JobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}
	if r.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "type is required")
	}
	validTypes := map[string]bool{"quality": true, "no_show": true, "pricing": true, "other": true}
	if !validTypes[r.Type] {
		return fiber.NewError(fiber.StatusBadRequest, "type must be one of: quality, no_show, pricing, other")
	}
	if r.Description == "" {
		return fiber.NewError(fiber.StatusBadRequest, "description is required")
	}
	return nil
}

// CreateDispute creates a new dispute.
// POST /api/v1/disputes
func (h *DisputeHandler) CreateDispute(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createDisputeRequest
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

	now := time.Now().UTC()
	dispute := &Dispute{
		ID:          uuid.New(),
		JobID:       jobID,
		RaisedBy:    userID,
		Type:        req.Type,
		Description: req.Description,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.service.Create(c.UserContext(), dispute); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create dispute")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create dispute",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": dispute,
	})
}

// GetDispute returns a single dispute by ID.
// GET /api/v1/disputes/:id
func (h *DisputeHandler) GetDispute(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid dispute ID format",
			},
		})
	}

	dispute, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("dispute_id", id.String()).Msg("failed to get dispute")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve dispute",
			},
		})
	}

	if dispute == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "dispute not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": dispute,
	})
}

// ListMyDisputes lists the authenticated user's disputes.
// GET /api/v1/disputes?page=1&limit=20
func (h *DisputeHandler) ListMyDisputes(c *fiber.Ctx) error {
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

	disputes, total, err := h.service.ListByUser(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list disputes")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list disputes",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": disputes,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// AddEvidence uploads evidence for a dispute.
// POST /api/v1/disputes/:id/evidence
func (h *DisputeHandler) AddEvidence(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	disputeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid dispute ID format",
			},
		})
	}

	evidenceType := c.FormValue("type", "photo")
	text := c.FormValue("text")

	var fileURL *string
	file, fileErr := c.FormFile("file")
	if fileErr == nil && file != nil {
		// Validate file size (max 10MB).
		if file.Size > 10*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "file size must not exceed 10MB",
				},
			})
		}

		// Open the uploaded file for reading.
		src, openErr := file.Open()
		if openErr != nil {
			log.Error().Err(openErr).Msg("failed to open uploaded evidence file")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "failed to process uploaded file",
				},
			})
		}
		defer src.Close()

		// Detect content type from the file extension.
		ext := filepath.Ext(file.Filename)
		contentType := stdmime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Generate a unique storage key to avoid filename collisions.
		uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		key := storage.DisputeEvidenceKey(disputeID.String(), uniqueName)

		uploadedURL, uploadErr := h.storage.Upload(c.UserContext(), key, src, contentType)
		if uploadErr != nil {
			log.Error().Err(uploadErr).Str("dispute_id", disputeID.String()).Str("key", key).Msg("failed to upload evidence to storage")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "STORAGE_ERROR",
					"message": "failed to upload evidence to storage",
				},
			})
		}
		fileURL = &uploadedURL
	}

	var textPtr *string
	if text != "" {
		textPtr = &text
	}

	evidence := &DisputeEvidence{
		ID:        uuid.New(),
		DisputeID: disputeID,
		UserID:    userID,
		Type:      evidenceType,
		FileURL:   fileURL,
		Text:      textPtr,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.service.AddEvidence(c.UserContext(), evidence); err != nil {
		log.Error().Err(err).Str("dispute_id", disputeID.String()).Msg("failed to add evidence")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to add evidence",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": evidence,
	})
}

// respondToDisputeRequest is the payload for POST /api/v1/disputes/:id/respond.
type respondToDisputeRequest struct {
	Response string `json:"response"`
}

// RespondToDispute allows the other party to respond to a dispute.
// POST /api/v1/disputes/:id/respond
func (h *DisputeHandler) RespondToDispute(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	disputeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid dispute ID format",
			},
		})
	}

	var req respondToDisputeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.Response == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "response is required",
			},
		})
	}

	if err := h.service.Respond(c.UserContext(), disputeID, userID, req.Response); err != nil {
		log.Error().Err(err).Str("dispute_id", disputeID.String()).Msg("failed to respond to dispute")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to respond to dispute",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"dispute_id": disputeID,
			"message":    "response submitted successfully",
		},
	})
}

// adminResolveDisputeRequest is the payload for POST /api/v1/disputes/:id/resolve.
type adminResolveDisputeRequest struct {
	Resolution string `json:"resolution"`
}

// AdminResolveDispute allows an admin/mediator to resolve a dispute.
// POST /api/v1/admin/disputes/:id/resolve
func (h *DisputeHandler) AdminResolveDispute(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	disputeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid dispute ID format",
			},
		})
	}

	var req adminResolveDisputeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.Resolution == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "resolution is required",
			},
		})
	}

	if err := h.service.Resolve(c.UserContext(), disputeID, userID, req.Resolution); err != nil {
		log.Error().Err(err).Str("dispute_id", disputeID.String()).Msg("failed to resolve dispute")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to resolve dispute",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"dispute_id": disputeID,
			"status":     "resolved",
			"message":    "dispute resolved successfully",
		},
	})
}
