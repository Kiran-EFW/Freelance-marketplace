package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// PointsBalance holds the user's current points balance and recent history.
type PointsBalance struct {
	UserID  uuid.UUID      `json:"user_id"`
	Balance int            `json:"balance"`
	Recent  []PointsEntry  `json:"recent"`
}

// PointsEntry represents a single entry in the points ledger.
type PointsEntry struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Amount      int       `json:"amount"` // positive = earned, negative = spent
	Type        string    `json:"type"`   // earned, spent, bonus, penalty
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserLevel holds the user's current level and progress info.
type UserLevel struct {
	UserID         uuid.UUID `json:"user_id"`
	Level          int       `json:"level"`
	CurrentPoints  int       `json:"current_points"`
	PointsToNext   int       `json:"points_to_next_level"`
	ProgressPct    float64   `json:"progress_percent"`
	Title          string    `json:"title"` // e.g., "Bronze", "Silver", "Gold"
}

// LeaderboardEntry represents a position on the leaderboard.
type LeaderboardEntry struct {
	Rank     int       `json:"rank"`
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Points   int       `json:"points"`
	Level    int       `json:"level"`
	Postcode string    `json:"postcode"`
}

// GamificationService defines the business operations required by GamificationHandler.
type GamificationService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*PointsBalance, error)
	GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]PointsEntry, int, error)
	GetLevel(ctx context.Context, userID uuid.UUID) (*UserLevel, error)
	GetLeaderboard(ctx context.Context, postcode string, limit, offset int) ([]LeaderboardEntry, error)
	SpendPoints(ctx context.Context, userID uuid.UUID, amount int, purpose string) error
}

// GamificationHandler handles gamification / points endpoints.
type GamificationHandler struct {
	service GamificationService
}

// NewGamificationHandler creates a new GamificationHandler.
func NewGamificationHandler(svc GamificationService) *GamificationHandler {
	return &GamificationHandler{service: svc}
}

// RegisterRoutes mounts gamification routes on the given Fiber router group.
func (h *GamificationHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/", h.GetMyPoints)
	rg.Get("/history", h.GetPointsHistory)
	rg.Get("/level", h.GetMyLevel)
	rg.Get("/leaderboard", h.GetLeaderboard)
	rg.Post("/spend", h.SpendPoints)
}

// GetMyPoints returns the authenticated user's current points balance.
// GET /api/v1/points
func (h *GamificationHandler) GetMyPoints(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	balance, err := h.service.GetBalance(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get points balance")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve points balance",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": balance,
	})
}

// GetPointsHistory returns the authenticated user's points ledger.
// GET /api/v1/points/history?page=1&limit=20
func (h *GamificationHandler) GetPointsHistory(c *fiber.Ctx) error {
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

	entries, total, err := h.service.GetHistory(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get points history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve points history",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": entries,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetMyLevel returns the authenticated user's current level and progress.
// GET /api/v1/points/level
func (h *GamificationHandler) GetMyLevel(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	level, err := h.service.GetLevel(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get user level")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve level information",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": level,
	})
}

// GetLeaderboard returns the postcode leaderboard.
// GET /api/v1/points/leaderboard?postcode=560001&page=1&limit=20
func (h *GamificationHandler) GetLeaderboard(c *fiber.Ctx) error {
	postcode := c.Query("postcode", "")
	page, limit := parsePagination(c)
	offset := (page - 1) * limit

	entries, err := h.service.GetLeaderboard(c.UserContext(), postcode, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("postcode", postcode).Msg("failed to get leaderboard")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve leaderboard",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": entries,
		"meta": fiber.Map{
			"page":     page,
			"limit":    limit,
			"postcode": postcode,
		},
	})
}

// spendPointsRequest is the payload for POST /api/v1/points/spend.
type spendPointsRequest struct {
	Amount  int    `json:"amount"`
	Purpose string `json:"purpose"` // boost, badge, highlight, etc.
}

func (r *spendPointsRequest) validate() error {
	if r.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}
	if r.Purpose == "" {
		return fiber.NewError(fiber.StatusBadRequest, "purpose is required")
	}
	validPurposes := map[string]bool{
		"boost":     true,
		"badge":     true,
		"highlight": true,
		"premium":   true,
	}
	if !validPurposes[r.Purpose] {
		return fiber.NewError(fiber.StatusBadRequest, "purpose must be one of: boost, badge, highlight, premium")
	}
	return nil
}

// SpendPoints deducts points from the user's balance.
// POST /api/v1/points/spend
func (h *GamificationHandler) SpendPoints(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req spendPointsRequest
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

	if err := h.service.SpendPoints(c.UserContext(), userID, req.Amount, req.Purpose); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Int("amount", req.Amount).Msg("failed to spend points")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to spend points",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "points spent successfully",
			"amount":  req.Amount,
			"purpose": req.Purpose,
		},
	})
}
