package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/whatsapp"
	"github.com/seva-platform/backend/internal/config"
)

// WhatsAppMessageHandler is called for each incoming WhatsApp message.
// Implementations can create conversations, route messages, etc.
type WhatsAppMessageHandler interface {
	HandleIncoming(ctx context.Context, msg whatsapp.IncomingMessage) error
}

// WhatsAppWebhookHandler handles Meta WhatsApp Cloud API webhooks.
type WhatsAppWebhookHandler struct {
	cfg        *config.Config
	msgHandler WhatsAppMessageHandler
}

// NewWhatsAppWebhookHandler creates a new WhatsAppWebhookHandler.
func NewWhatsAppWebhookHandler(cfg *config.Config, msgHandler WhatsAppMessageHandler) *WhatsAppWebhookHandler {
	return &WhatsAppWebhookHandler{
		cfg:        cfg,
		msgHandler: msgHandler,
	}
}

// RegisterRoutes mounts WhatsApp webhook routes on the given Fiber router group.
// Expected mount point: /webhooks/whatsapp
func (h *WhatsAppWebhookHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/whatsapp", h.Verify)
	rg.Post("/whatsapp", h.Receive)
}

// Verify handles the WhatsApp webhook verification challenge.
// Meta sends a GET request with hub.mode, hub.verify_token, and hub.challenge.
//
// GET /webhooks/whatsapp?hub.mode=subscribe&hub.verify_token=...&hub.challenge=...
func (h *WhatsAppWebhookHandler) Verify(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode != "subscribe" {
		log.Warn().
			Str("mode", mode).
			Msg("whatsapp webhook: unexpected hub.mode")
		return c.Status(fiber.StatusForbidden).SendString("invalid mode")
	}

	if token != h.cfg.WhatsAppWebhookVerifyToken {
		log.Warn().
			Str("token", token).
			Msg("whatsapp webhook: verify token mismatch")
		return c.Status(fiber.StatusForbidden).SendString("invalid verify token")
	}

	log.Info().Msg("whatsapp webhook: verification successful")
	return c.Status(fiber.StatusOK).SendString(challenge)
}

// Receive handles incoming WhatsApp messages and status updates.
//
// POST /webhooks/whatsapp
func (h *WhatsAppWebhookHandler) Receive(c *fiber.Ctx) error {
	// Verify webhook signature if app secret is configured.
	signature := c.Get("X-Hub-Signature-256")
	if signature != "" && h.cfg.WhatsAppBusinessID != "" {
		if !whatsapp.VerifyWebhookSignature(c.Body(), signature, h.cfg.WhatsAppBusinessID) {
			log.Warn().Msg("whatsapp webhook: invalid signature")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_SIGNATURE",
					"message": "webhook signature verification failed",
				},
			})
		}
	}

	// Parse the incoming payload.
	messages, err := whatsapp.ParseWebhookPayload(c.Body())
	if err != nil {
		log.Error().Err(err).Msg("whatsapp webhook: failed to parse payload")
		// Return 200 to Meta to avoid retries for malformed payloads.
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "error",
		})
	}

	// Process each message.
	for _, msg := range messages {
		log.Info().
			Str("from", msg.From).
			Str("type", msg.Type).
			Str("message_id", msg.MessageID).
			Str("text", msg.Text).
			Msg("whatsapp webhook: incoming message")

		if h.msgHandler != nil {
			if err := h.msgHandler.HandleIncoming(c.UserContext(), msg); err != nil {
				log.Error().Err(err).
					Str("from", msg.From).
					Str("message_id", msg.MessageID).
					Msg("whatsapp webhook: failed to handle message")
			}
		}
	}

	// Always return 200 to Meta to acknowledge receipt.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
	})
}

// ---------------------------------------------------------------------------
// Default message handler (logs only, for initial integration)
// ---------------------------------------------------------------------------

// DefaultWhatsAppMessageHandler is a basic handler that logs incoming messages.
// Replace with a real implementation that creates conversations, updates jobs, etc.
type DefaultWhatsAppMessageHandler struct{}

// HandleIncoming logs the incoming message. Override this for real logic.
func (d *DefaultWhatsAppMessageHandler) HandleIncoming(ctx context.Context, msg whatsapp.IncomingMessage) error {
	log.Info().
		Str("from", msg.From).
		Str("type", msg.Type).
		Str("text", msg.Text).
		Str("button_id", msg.ButtonID).
		Msg("default whatsapp handler: received message")
	return nil
}
