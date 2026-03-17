package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
	"github.com/seva-platform/backend/internal/repository/postgres"
)

// DeviceTokenHandler handles device token registration endpoints for push
// notifications.
type DeviceTokenHandler struct {
	queries *postgres.Queries
}

// NewDeviceTokenHandler creates a new DeviceTokenHandler.
func NewDeviceTokenHandler(queries *postgres.Queries) *DeviceTokenHandler {
	return &DeviceTokenHandler{queries: queries}
}

// RegisterRoutes mounts device token routes on the given Fiber router group.
// These routes are mounted under /api/v1/notifications alongside the existing
// notification routes.
func (h *DeviceTokenHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/device-token", h.RegisterToken)
	rg.Delete("/device-token/:token", h.DeactivateToken)
}

// registerDeviceTokenRequest is the payload for POST /api/v1/notifications/device-token.
type registerDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"` // android, ios, web
}

func (r *registerDeviceTokenRequest) validate() error {
	if r.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token is required")
	}
	switch r.Platform {
	case "android", "ios", "web":
		// valid
	default:
		return fiber.NewError(fiber.StatusBadRequest, "platform must be one of: android, ios, web")
	}
	return nil
}

// RegisterToken registers or updates a device token for push notifications.
// POST /api/v1/notifications/device-token
func (h *DeviceTokenHandler) RegisterToken(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req registerDeviceTokenRequest
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

	dt, err := h.queries.RegisterDeviceToken(c.UserContext(), postgres.RegisterDeviceTokenParams{
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to register device token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to register device token",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":       dt.ID,
			"token":    dt.Token,
			"platform": dt.Platform,
			"message":  "device token registered successfully",
		},
	})
}

// DeactivateToken deactivates a device token.
// DELETE /api/v1/notifications/device-token/:token
func (h *DeviceTokenHandler) DeactivateToken(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "device token is required",
			},
		})
	}

	if err := h.queries.DeactivateDeviceToken(c.UserContext(), token); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to deactivate device token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to deactivate device token",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "device token deactivated successfully",
		},
	})
}
