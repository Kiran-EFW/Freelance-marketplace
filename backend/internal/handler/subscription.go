package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// ---------------------------------------------------------------------------
// Plan definitions
// ---------------------------------------------------------------------------

// Plan describes a subscription tier with its pricing and features.
type Plan struct {
	Tier               string   `json:"tier"`
	Name               string   `json:"name"`
	PriceINR           float64  `json:"price_inr"`
	MaxLeadsPerMonth   int      `json:"max_leads_per_month"` // -1 = unlimited
	CommissionDiscount float64  `json:"commission_discount"` // e.g. 0.02 = 2%
	Features           []string `json:"features"`
	Description        string   `json:"description"`
}

// AvailablePlans holds the subscription plans offered by Seva.
var AvailablePlans = map[string]Plan{
	"free": {
		Tier:               "free",
		Name:               "Free",
		PriceINR:           0,
		MaxLeadsPerMonth:   5,
		CommissionDiscount: 0,
		Features:           []string{"basic_search", "5_leads_per_month"},
		Description:        "Get started with basic features",
	},
	"professional": {
		Tier:               "professional",
		Name:               "Pro",
		PriceINR:           299,
		MaxLeadsPerMonth:   -1,
		CommissionDiscount: 0.02,
		Features:           []string{"priority_search", "pro_badge", "analytics", "unlimited_leads"},
		Description:        "Unlimited leads and priority search",
	},
	"enterprise": {
		Tier:               "enterprise",
		Name:               "Business",
		PriceINR:           999,
		MaxLeadsPerMonth:   -1,
		CommissionDiscount: 0.03,
		Features:           []string{"team_profiles", "branded_page", "quote_templates", "invoice_generation", "priority_search", "pro_badge", "analytics", "unlimited_leads"},
		Description:        "Team profiles and advanced tools",
	},
}

// ---------------------------------------------------------------------------
// Subscription domain types (handler-level)
// ---------------------------------------------------------------------------

// Subscription represents a provider's subscription record.
type Subscription struct {
	ID                    uuid.UUID  `json:"id"`
	ProviderID            uuid.UUID  `json:"provider_id"`
	Tier                  string     `json:"tier"`
	StartedAt             time.Time  `json:"started_at"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	AutoRenew             bool       `json:"auto_renew"`
	PaymentMethod         string     `json:"payment_method,omitempty"`
	GatewaySubscriptionID string     `json:"gateway_subscription_id,omitempty"`
	Amount                float64    `json:"amount"`
	Currency              string     `json:"currency"`
	Status                string     `json:"status"` // active, cancelled, expired, past_due
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// SubscriptionService defines the business operations for subscriptions.
type SubscriptionService interface {
	GetCurrentSubscription(ctx context.Context, providerID uuid.UUID) (*Subscription, error)
	Subscribe(ctx context.Context, providerID uuid.UUID, tier, paymentMethod string) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID, providerID uuid.UUID) error
	HandlePaymentWebhook(ctx context.Context, payload []byte, signature string) error
	ListBillingHistory(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]Subscription, int, error)
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// SubscriptionHandler handles subscription management endpoints.
type SubscriptionHandler struct {
	service SubscriptionService
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(svc SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: svc}
}

// RegisterRoutes mounts subscription routes on the given Fiber router group.
// Expected mount point: /api/v1/subscriptions (authenticated)
func (h *SubscriptionHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/plans", h.ListPlans)
	rg.Post("/", h.Subscribe)
	rg.Get("/current", h.GetCurrentSubscription)
	rg.Get("/history", h.GetBillingHistory)
	rg.Put("/:id/cancel", h.CancelSubscription)
}

// RegisterWebhookRoutes mounts the payment gateway webhook for subscription
// renewals. Expected mount point: /webhooks
func (h *SubscriptionHandler) RegisterWebhookRoutes(rg fiber.Router) {
	rg.Post("/subscription", h.PaymentWebhook)
}

// ---------------------------------------------------------------------------
// Endpoint handlers
// ---------------------------------------------------------------------------

// ListPlans returns all available subscription plans with pricing.
// GET /api/v1/subscriptions/plans
func (h *SubscriptionHandler) ListPlans(c *fiber.Ctx) error {
	// Return plans in a stable order.
	plans := []Plan{
		AvailablePlans["free"],
		AvailablePlans["professional"],
		AvailablePlans["enterprise"],
	}

	return c.JSON(fiber.Map{
		"data": plans,
	})
}

// subscribeRequest is the payload for POST /api/v1/subscriptions.
type subscribeRequest struct {
	Tier          string `json:"tier"`
	PaymentMethod string `json:"payment_method"`
}

func (r *subscribeRequest) validate() error {
	if r.Tier == "" {
		return fiber.NewError(fiber.StatusBadRequest, "tier is required")
	}
	if _, ok := AvailablePlans[r.Tier]; !ok {
		return fiber.NewError(fiber.StatusBadRequest, "invalid subscription tier; valid options are: free, professional, enterprise")
	}
	return nil
}

// Subscribe creates a new subscription for the authenticated provider.
// POST /api/v1/subscriptions
func (h *SubscriptionHandler) Subscribe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req subscribeRequest
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

	sub, err := h.service.Subscribe(c.UserContext(), userID, req.Tier, req.PaymentMethod)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Str("tier", req.Tier).Msg("failed to create subscription")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create subscription",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": sub,
	})
}

// GetCurrentSubscription returns the provider's active subscription.
// GET /api/v1/subscriptions/current
func (h *SubscriptionHandler) GetCurrentSubscription(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	sub, err := h.service.GetCurrentSubscription(c.UserContext(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get current subscription")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve subscription",
			},
		})
	}

	if sub == nil {
		// No active subscription; return the default free tier.
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"tier":   "free",
				"status": "active",
				"plan":   AvailablePlans["free"],
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": sub,
	})
}

// CancelSubscription cancels a provider's subscription.
// PUT /api/v1/subscriptions/:id/cancel
func (h *SubscriptionHandler) CancelSubscription(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	subID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid subscription ID format",
			},
		})
	}

	if err := h.service.CancelSubscription(c.UserContext(), subID, userID); err != nil {
		log.Error().Err(err).Str("subscription_id", subID.String()).Msg("failed to cancel subscription")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to cancel subscription",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"subscription_id": subID,
			"status":          "cancelled",
			"message":         "subscription cancelled successfully",
		},
	})
}

// GetBillingHistory returns the provider's subscription billing history.
// GET /api/v1/subscriptions/history?page=1&limit=20
func (h *SubscriptionHandler) GetBillingHistory(c *fiber.Ctx) error {
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

	history, total, err := h.service.ListBillingHistory(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get billing history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve billing history",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": history,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// PaymentWebhook handles payment gateway webhooks for subscription renewals.
// POST /webhooks/subscription
func (h *SubscriptionHandler) PaymentWebhook(c *fiber.Ctx) error {
	signature := c.Get("X-Razorpay-Signature")
	if signature == "" {
		signature = c.Get("Stripe-Signature")
	}

	if err := h.service.HandlePaymentWebhook(c.UserContext(), c.Body(), signature); err != nil {
		log.Error().Err(err).Msg("subscription webhook processing failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "WEBHOOK_ERROR",
				"message": "failed to process subscription webhook",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}
