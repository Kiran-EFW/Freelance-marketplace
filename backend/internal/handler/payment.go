package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/middleware"
)

// PaymentOrder represents a payment order in the system.
type PaymentOrder struct {
	ID            uuid.UUID `json:"id"`
	JobID         uuid.UUID `json:"job_id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	GatewayID     string    `json:"gateway_id"`
	Status        string    `json:"status"` // created, paid, failed, refunded
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Transaction represents a financial transaction record.
type Transaction struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"` // payment, refund, payout
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentService defines the business operations required by PaymentHandler.
type PaymentService interface {
	CreateOrder(ctx context.Context, order *PaymentOrder) error
	VerifyPayment(ctx context.Context, orderID, paymentID, signature string) error
	HandleWebhook(ctx context.Context, gateway string, payload []byte, signature string) error
	GetPaymentStatus(ctx context.Context, id uuid.UUID) (*PaymentOrder, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Transaction, int, error)
	RequestRefund(ctx context.Context, paymentID, userID uuid.UUID, reason string) error
}

// PaymentHandler handles payment endpoints.
type PaymentHandler struct {
	service PaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// RegisterRoutes mounts payment routes on the given Fiber router group.
func (h *PaymentHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/orders", h.CreatePaymentOrder)
	rg.Post("/verify", h.VerifyPayment)
	rg.Get("/history", h.GetTransactionHistory)
	rg.Get("/:id", h.GetPaymentStatus)
	rg.Post("/:id/refund", h.RequestRefund)
}

// RegisterWebhookRoutes mounts webhook routes (no JWT auth, signature-verified).
func (h *PaymentHandler) RegisterWebhookRoutes(rg fiber.Router) {
	rg.Post("/payment", h.PaymentWebhook)
}

// createPaymentOrderRequest is the payload for POST /api/v1/payments/orders.
type createPaymentOrderRequest struct {
	JobID         string  `json:"job_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
}

func (r *createPaymentOrderRequest) validate() error {
	if r.JobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}
	if r.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}
	return nil
}

// CreatePaymentOrder creates a payment order for a job.
// POST /api/v1/payments/orders
func (h *PaymentHandler) CreatePaymentOrder(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req createPaymentOrderRequest
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

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	now := time.Now().UTC()
	order := &PaymentOrder{
		ID:            uuid.New(),
		JobID:         jobID,
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      currency,
		Status:        "created",
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.service.CreateOrder(c.UserContext(), order); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create payment order")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create payment order",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": order,
	})
}

// verifyPaymentRequest is the payload for POST /api/v1/payments/verify.
type verifyPaymentRequest struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Signature string `json:"signature"`
}

func (r *verifyPaymentRequest) validate() error {
	if r.OrderID == "" || r.PaymentID == "" || r.Signature == "" {
		return fiber.NewError(fiber.StatusBadRequest, "order_id, payment_id, and signature are required")
	}
	return nil
}

// VerifyPayment verifies a payment gateway callback.
// POST /api/v1/payments/verify
func (h *PaymentHandler) VerifyPayment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	var req verifyPaymentRequest
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

	if err := h.service.VerifyPayment(c.UserContext(), req.OrderID, req.PaymentID, req.Signature); err != nil {
		log.Error().Err(err).Str("order_id", req.OrderID).Msg("payment verification failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VERIFICATION_FAILED",
				"message": "payment verification failed",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"verified": true,
			"order_id": req.OrderID,
		},
	})
}

// PaymentWebhook handles payment gateway webhooks (Razorpay/Stripe).
// POST /webhooks/payment
func (h *PaymentHandler) PaymentWebhook(c *fiber.Ctx) error {
	signature := c.Get("X-Razorpay-Signature")
	if signature == "" {
		signature = c.Get("Stripe-Signature")
	}

	gateway := "razorpay"
	if c.Get("Stripe-Signature") != "" {
		gateway = "stripe"
	}

	if err := h.service.HandleWebhook(c.UserContext(), gateway, c.Body(), signature); err != nil {
		log.Error().Err(err).Str("gateway", gateway).Msg("webhook processing failed")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "WEBHOOK_ERROR",
				"message": "failed to process webhook",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

// GetPaymentStatus returns the status of a payment.
// GET /api/v1/payments/:id
func (h *PaymentHandler) GetPaymentStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid payment ID format",
			},
		})
	}

	order, err := h.service.GetPaymentStatus(c.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Str("payment_id", id.String()).Msg("failed to get payment status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve payment status",
			},
		})
	}

	if order == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "payment not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": order,
	})
}

// GetTransactionHistory returns the user's transaction history.
// GET /api/v1/payments/history?page=1&limit=20
func (h *PaymentHandler) GetTransactionHistory(c *fiber.Ctx) error {
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

	transactions, total, err := h.service.GetTransactionHistory(c.UserContext(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get transaction history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve transaction history",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": transactions,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// requestRefundRequest is the payload for POST /api/v1/payments/:id/refund.
type requestRefundRequest struct {
	Reason string `json:"reason"`
}

// RequestRefund requests a refund for a payment.
// POST /api/v1/payments/:id/refund
func (h *PaymentHandler) RequestRefund(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
	}

	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid payment ID format",
			},
		})
	}

	var req requestRefundRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := h.service.RequestRefund(c.UserContext(), paymentID, userID, req.Reason); err != nil {
		log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("failed to request refund")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to process refund request",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"payment_id": paymentID,
			"status":     "refund_requested",
		},
	})
}
