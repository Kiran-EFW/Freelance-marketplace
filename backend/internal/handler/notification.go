package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// Notification represents a user notification.
type Notification struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      string     `json:"type"` // job_update, payment, review, system, promotion
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Data      *string    `json:"data,omitempty"` // JSON payload for deep linking
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationPreferences holds a user's notification settings.
type NotificationPreferences struct {
	UserID      uuid.UUID `json:"user_id"`
	PushEnabled bool      `json:"push_enabled"`
	SMSEnabled  bool      `json:"sms_enabled"`
	EmailEnabled bool     `json:"email_enabled"`
	JobUpdates  bool      `json:"job_updates"`
	Promotions  bool      `json:"promotions"`
	Reviews     bool      `json:"reviews"`
}

// NotificationService defines the business operations required by NotificationHandler.
type NotificationService interface {
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, int, error)
	MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *NotificationPreferences) error
}

// NotificationHandler handles notification endpoints.
type NotificationHandler struct {
	service NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(svc NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

// RegisterRoutes mounts notification routes on the given Fiber router group.
func (h *NotificationHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/", h.ListNotifications)
	rg.Get("/count", h.GetUnreadCount)
	rg.Post("/read-all", h.MarkAllRead)
	rg.Put("/preferences", h.UpdatePreferences)
	rg.Patch("/:id/read", h.MarkRead)
}

// ListNotifications returns paginated notifications, unread first.
// GET /api/v1/notifications?page=1&limit=20
func (h *NotificationHandler) ListNotifications(c *fiber.Ctx) error {
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

	notifications, total, err := h.service.List(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list notifications")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list notifications",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": notifications,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// MarkRead marks a single notification as read.
// PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid notification ID format",
			},
		})
	}

	if err := h.service.MarkRead(c.UserContext(), notifID, userID); err != nil {
		log.Error().Err(err).Str("notification_id", notifID.String()).Msg("failed to mark notification as read")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to mark notification as read",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":      notifID,
			"is_read": true,
		},
	})
}

// MarkAllRead marks all notifications as read for the authenticated user.
// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	if err := h.service.MarkAllRead(c.UserContext(), userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to mark all notifications as read")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to mark all notifications as read",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "all notifications marked as read",
		},
	})
}

// GetUnreadCount returns the count of unread notifications.
// GET /api/v1/notifications/count
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	count, err := h.service.GetUnreadCount(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get unread count")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve unread count",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"unread_count": count,
		},
	})
}

// updatePreferencesRequest is the payload for PUT /api/v1/notifications/preferences.
type updatePreferencesRequest struct {
	PushEnabled  *bool `json:"push_enabled"`
	SMSEnabled   *bool `json:"sms_enabled"`
	EmailEnabled *bool `json:"email_enabled"`
	JobUpdates   *bool `json:"job_updates"`
	Promotions   *bool `json:"promotions"`
	Reviews      *bool `json:"reviews"`
}

// UpdatePreferences updates the user's notification preferences.
// PUT /api/v1/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req updatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	// Fetch existing preferences to apply partial updates.
	existing, err := h.service.GetPreferences(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get notification preferences")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve preferences",
			},
		})
	}

	prefs := existing
	if prefs == nil {
		prefs = &NotificationPreferences{
			UserID:       userID,
			PushEnabled:  true,
			SMSEnabled:   true,
			EmailEnabled: true,
			JobUpdates:   true,
			Promotions:   true,
			Reviews:      true,
		}
	}

	if req.PushEnabled != nil {
		prefs.PushEnabled = *req.PushEnabled
	}
	if req.SMSEnabled != nil {
		prefs.SMSEnabled = *req.SMSEnabled
	}
	if req.EmailEnabled != nil {
		prefs.EmailEnabled = *req.EmailEnabled
	}
	if req.JobUpdates != nil {
		prefs.JobUpdates = *req.JobUpdates
	}
	if req.Promotions != nil {
		prefs.Promotions = *req.Promotions
	}
	if req.Reviews != nil {
		prefs.Reviews = *req.Reviews
	}

	if err := h.service.UpdatePreferences(c.UserContext(), prefs); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to update notification preferences")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update preferences",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": prefs,
	})
}
