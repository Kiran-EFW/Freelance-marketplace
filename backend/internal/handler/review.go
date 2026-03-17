package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// Review represents a customer review for a completed job.
type Review struct {
	ID         uuid.UUID  `json:"id"`
	JobID      uuid.UUID  `json:"job_id"`
	ReviewerID uuid.UUID  `json:"reviewer_id"`
	ProviderID uuid.UUID  `json:"provider_id"`
	Rating     int        `json:"rating"` // 1-5
	Comment    string     `json:"comment"`
	Response   *string    `json:"response,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// RatingStats holds aggregate rating data for a provider.
type RatingStats struct {
	ProviderID   uuid.UUID      `json:"provider_id"`
	AverageRating float64       `json:"average_rating"`
	TotalReviews  int           `json:"total_reviews"`
	Distribution  map[int]int   `json:"distribution"` // rating -> count
}

// ReviewService defines the business operations required by ReviewHandler.
type ReviewService interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*Review, error)
	ListByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]Review, int, error)
	RespondToReview(ctx context.Context, reviewID, providerID uuid.UUID, response string) error
	GetRatingStats(ctx context.Context, providerID uuid.UUID) (*RatingStats, error)
}

// ReviewHandler handles review endpoints.
type ReviewHandler struct {
	service ReviewService
}

// NewReviewHandler creates a new ReviewHandler.
func NewReviewHandler(svc ReviewService) *ReviewHandler {
	return &ReviewHandler{service: svc}
}

// RegisterRoutes mounts review routes on the given Fiber router group.
func (h *ReviewHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateReview)
	rg.Get("/:id", h.GetReview)
	rg.Post("/:id/respond", h.RespondToReview)
}

// RegisterProviderReviewRoutes mounts provider review routes.
// These are nested under /api/v1/providers/:id/reviews and /api/v1/providers/:id/ratings.
func (h *ReviewHandler) RegisterProviderReviewRoutes(rg fiber.Router) {
	rg.Get("/:id/reviews", h.ListProviderReviews)
	rg.Get("/:id/ratings", h.GetProviderRatingStats)
}

// createReviewRequest is the payload for POST /api/v1/reviews.
type createReviewRequest struct {
	JobID   string `json:"job_id"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

func (r *createReviewRequest) validate() error {
	if r.JobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}
	if r.Rating < 1 || r.Rating > 5 {
		return fiber.NewError(fiber.StatusBadRequest, "rating must be between 1 and 5")
	}
	return nil
}

// CreateReview creates a review for a completed job.
// POST /api/v1/reviews
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createReviewRequest
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
	review := &Review{
		ID:         uuid.New(),
		JobID:      jobID,
		ReviewerID: userID,
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.service.Create(c.UserContext(), review); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("job_id", jobID.String()).Msg("failed to create review")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create review",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": review,
	})
}

// GetReview returns a single review by ID.
// GET /api/v1/reviews/:id
func (h *ReviewHandler) GetReview(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid review ID format",
			},
		})
	}

	review, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("review_id", id.String()).Msg("failed to get review")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve review",
			},
		})
	}

	if review == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "review not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": review,
	})
}

// ListProviderReviews returns paginated reviews for a provider.
// GET /api/v1/providers/:id/reviews?page=1&limit=20
func (h *ReviewHandler) ListProviderReviews(c *fiber.Ctx) error {
	providerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider ID format",
			},
		})
	}

	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	reviews, total, err := h.service.ListByProvider(c.UserContext(), providerID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID.String()).Msg("failed to list provider reviews")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list reviews",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": reviews,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// respondToReviewRequest is the payload for POST /api/v1/reviews/:id/respond.
type respondToReviewRequest struct {
	Response string `json:"response"`
}

func (r *respondToReviewRequest) validate() error {
	if r.Response == "" {
		return fiber.NewError(fiber.StatusBadRequest, "response is required")
	}
	return nil
}

// RespondToReview allows a provider to respond to a review.
// POST /api/v1/reviews/:id/respond
func (h *ReviewHandler) RespondToReview(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid review ID format",
			},
		})
	}

	var req respondToReviewRequest
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

	if err := h.service.RespondToReview(c.UserContext(), reviewID, userID, req.Response); err != nil {
		log.Error().Err(err).Str("review_id", reviewID.String()).Msg("failed to respond to review")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to respond to review",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"review_id": reviewID,
			"message":   "response added successfully",
		},
	})
}

// GetProviderRatingStats returns aggregate rating statistics for a provider.
// GET /api/v1/providers/:id/ratings
func (h *ReviewHandler) GetProviderRatingStats(c *fiber.Ctx) error {
	providerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider ID format",
			},
		})
	}

	stats, err := h.service.GetRatingStats(c.UserContext(), providerID)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID.String()).Msg("failed to get rating stats")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve rating statistics",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}
