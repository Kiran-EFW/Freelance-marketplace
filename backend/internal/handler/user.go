package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/domain"
	"github.com/seva-platform/backend/internal/middleware"
)

// UserService defines the business operations required by UserHandler.
type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Deactivate(ctx context.Context, id uuid.UUID) error
}

// UserHandler handles user profile endpoints.
type UserHandler struct {
	service UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// RegisterRoutes mounts user routes on the given Fiber router group.
func (h *UserHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/me", h.GetProfile)
	rg.Put("/me", h.UpdateProfile)
	rg.Delete("/me", h.DeactivateAccount)
	rg.Get("/:id", h.GetUserByID)
}

// GetProfile returns the authenticated user's profile.
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	user, err := h.service.GetByID(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve profile",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

// updateProfileRequest is the payload for PUT /api/v1/users/me.
type updateProfileRequest struct {
	Name              *string `json:"name"`
	Email             *string `json:"email"`
	PreferredLanguage *string `json:"preferred_language"`
	Postcode          *string `json:"postcode"`
	Address           *string `json:"address"`
}

// validate checks that the update request has at least one field set.
func (r *updateProfileRequest) validate() error {
	if r.Name == nil && r.Email == nil && r.PreferredLanguage == nil && r.Postcode == nil && r.Address == nil {
		return fiber.NewError(fiber.StatusBadRequest, "at least one field must be provided")
	}
	return nil
}

// UpdateProfile updates the authenticated user's profile.
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req updateProfileRequest
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

	// Fetch existing user to apply partial updates.
	user, err := h.service.GetByID(c.UserContext(), userID)
	if err != nil || user == nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get user for update")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		})
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.PreferredLanguage != nil {
		user.PreferredLanguage = *req.PreferredLanguage
	}
	user.UpdatedAt = time.Now().UTC()

	if err := h.service.Update(c.UserContext(), user); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update profile",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

// GetUserByID returns a user by ID. Admin-only endpoint.
// GET /api/v1/users/:id
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid user ID format",
			},
		})
	}

	user, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("target_user_id", idStr).Msg("failed to get user by ID")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

// DeactivateAccount soft-deletes the authenticated user's account.
// DELETE /api/v1/users/me
func (h *UserHandler) DeactivateAccount(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	if err := h.service.Deactivate(c.UserContext(), userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to deactivate account")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to deactivate account",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"message": "account deactivated successfully",
		},
	})
}
