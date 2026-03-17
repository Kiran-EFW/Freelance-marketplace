package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
	fraudsvc "github.com/seva-platform/backend/internal/service/fraud"
)

// FraudService defines the business operations required by FraudHandler.
type FraudService interface {
	GetFlaggedAccounts(ctx context.Context, limit, offset int) ([]fraudsvc.FlaggedAccount, int, error)
	CalculateRiskScore(ctx context.Context, userID uuid.UUID) (float64, error)
	GetUserRiskProfile(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	DetectFakeReviewRing(ctx context.Context, providerID uuid.UUID) (*fraudsvc.ReviewRingResult, error)
	UpdateFlagStatus(ctx context.Context, userID uuid.UUID, newStatus string, adminID uuid.UUID) error
	FlagAccount(ctx context.Context, userID uuid.UUID, reason string, riskScore float64) error
}

// FraudHandler handles fraud detection and management endpoints for admin
// users. All routes are mounted under the admin route group and require
// admin-level authentication.
type FraudHandler struct {
	service FraudService
}

// NewFraudHandler creates a new FraudHandler.
func NewFraudHandler(svc FraudService) *FraudHandler {
	return &FraudHandler{service: svc}
}

// RegisterRoutes mounts fraud management routes on the given Fiber router
// group. Expected to be called with adminGroup.Group("/fraud").
func (h *FraudHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/flagged", h.ListFlaggedAccounts)
	rg.Get("/risk/:userId", h.GetRiskScore)
	rg.Post("/review-ring/:providerId", h.CheckReviewRing)
	rg.Post("/clear/:userId", h.ClearAccount)
	rg.Post("/suspend/:userId", h.SuspendAccount)
	rg.Post("/flag/:userId", h.FlagAccount)
}

// ListFlaggedAccounts returns a paginated list of flagged accounts sorted by
// risk score.
// GET /api/v1/admin/fraud/flagged?page=1&limit=20
func (h *FraudHandler) ListFlaggedAccounts(c *fiber.Ctx) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	accounts, total, err := h.service.GetFlaggedAccounts(c.UserContext(), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("fraud: failed to list flagged accounts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list flagged accounts",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": accounts,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetRiskScore returns the computed risk score and risk profile for a user.
// GET /api/v1/admin/fraud/risk/:userId
func (h *FraudHandler) GetRiskScore(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	profile, err := h.service.GetUserRiskProfile(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to get risk profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to calculate risk score",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": profile,
	})
}

// CheckReviewRing runs fake review ring analysis for a provider.
// POST /api/v1/admin/fraud/review-ring/:providerId
func (h *FraudHandler) CheckReviewRing(c *fiber.Ctx) error {
	providerID, err := uuid.Parse(c.Params("providerId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid provider ID format",
			},
		})
	}

	result, err := h.service.DetectFakeReviewRing(c.UserContext(), providerID)
	if err != nil {
		log.Error().Err(err).Str("provider_id", providerID.String()).Msg("fraud: review ring check failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to check review ring",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

// ClearAccount clears a flagged account, marking it as reviewed and safe.
// POST /api/v1/admin/fraud/clear/:userId
func (h *FraudHandler) ClearAccount(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	if err := h.service.UpdateFlagStatus(c.UserContext(), userID, "cleared", adminID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to clear account")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to clear flagged account",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user_id": userID,
			"status":  "cleared",
			"message": "account cleared successfully",
		},
	})
}

// SuspendAccount suspends a flagged account, deactivating the user.
// POST /api/v1/admin/fraud/suspend/:userId
func (h *FraudHandler) SuspendAccount(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	if err := h.service.UpdateFlagStatus(c.UserContext(), userID, "suspended", adminID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to suspend account")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to suspend flagged account",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user_id": userID,
			"status":  "suspended",
			"message": "account suspended successfully",
		},
	})
}

// flagAccountRequest is the payload for POST /api/v1/admin/fraud/flag/:userId.
type flagAccountRequest struct {
	Reason    string  `json:"reason"`
	RiskScore float64 `json:"risk_score"`
}

// FlagAccount manually flags a user account for review.
// POST /api/v1/admin/fraud/flag/:userId
func (h *FraudHandler) FlagAccount(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	var req flagAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "reason is required",
			},
		})
	}

	// If no risk score provided, compute one automatically.
	riskScore := req.RiskScore
	if riskScore == 0 {
		computed, err := h.service.CalculateRiskScore(c.UserContext(), userID)
		if err != nil {
			log.Warn().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to auto-compute risk score")
			riskScore = 50 // default to moderate risk
		} else {
			riskScore = computed
		}
	}

	if err := h.service.FlagAccount(c.UserContext(), userID, req.Reason, riskScore); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("fraud: failed to flag account")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to flag account",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"user_id":    userID,
			"risk_score": riskScore,
			"reason":     req.Reason,
			"status":     "flagged",
			"message":    "account flagged successfully",
		},
	})
}
