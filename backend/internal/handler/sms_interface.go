package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
	smssession "github.com/seva-platform/backend/internal/service/sms_session"
)

// SMSInterfaceHandler handles incoming SMS messages via Twilio webhooks
// and provides a text-based interface for basic phone users.
type SMSInterfaceHandler struct {
	cfg            *config.Config
	sessionManager *smssession.SMSSessionManager
}

// NewSMSInterfaceHandler creates a new SMSInterfaceHandler.
func NewSMSInterfaceHandler(cfg *config.Config, sessionManager *smssession.SMSSessionManager) *SMSInterfaceHandler {
	return &SMSInterfaceHandler{
		cfg:            cfg,
		sessionManager: sessionManager,
	}
}

// RegisterRoutes mounts SMS webhook routes on the given Fiber router group.
// Expected mount point: /webhooks/sms
func (h *SMSInterfaceHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/incoming", h.HandleIncomingSMS)
}

// ---------------------------------------------------------------------------
// Twilio SMS webhook
// ---------------------------------------------------------------------------

// HandleIncomingSMS handles POST /webhooks/sms/incoming from Twilio.
// It parses the incoming SMS body as a command, processes it through the
// session manager, and returns a TwiML response.
func (h *SMSInterfaceHandler) HandleIncomingSMS(c *fiber.Ctx) error {
	// Optionally verify Twilio request signature.
	if h.cfg.TwilioAuthToken != "" {
		signature := c.Get("X-Twilio-Signature")
		if signature != "" {
			requestURL := h.buildRequestURL(c)
			if !verifyTwilioSignature(h.cfg.TwilioAuthToken, signature, requestURL, c.Body()) {
				log.Warn().Msg("sms webhook: invalid Twilio signature")
				return c.Status(fiber.StatusForbidden).SendString("invalid signature")
			}
		}
	}

	// Parse the Twilio webhook form data.
	from := c.FormValue("From")
	body := c.FormValue("Body")

	if from == "" || body == "" {
		log.Warn().Str("from", from).Msg("sms webhook: missing From or Body")
		return c.Status(fiber.StatusBadRequest).SendString("missing required fields")
	}

	log.Info().
		Str("from", from).
		Str("body", body).
		Msg("sms webhook: incoming message")

	// Process the message through the session manager.
	ctx := c.UserContext()
	reply := h.sessionManager.ProcessMessage(ctx, from, body)

	// Return TwiML response.
	twiml := formatTwiMLResponse(reply)
	c.Set("Content-Type", "application/xml")
	return c.Status(fiber.StatusOK).SendString(twiml)
}

// ---------------------------------------------------------------------------
// TwiML helpers
// ---------------------------------------------------------------------------

// formatTwiMLResponse wraps a reply message in TwiML <Response><Message> XML.
func formatTwiMLResponse(message string) string {
	// Escape XML special characters in the message.
	escaped := xmlEscape(message)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Message>%s</Message>
</Response>`, escaped)
}

// xmlEscape escapes XML special characters in a string.
func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// ---------------------------------------------------------------------------
// Twilio signature verification
// ---------------------------------------------------------------------------

// buildRequestURL reconstructs the full URL for signature verification.
func (h *SMSInterfaceHandler) buildRequestURL(c *fiber.Ctx) string {
	scheme := "https"
	if c.Get("X-Forwarded-Proto") != "" {
		scheme = c.Get("X-Forwarded-Proto")
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Hostname(), c.OriginalURL())
}

// verifyTwilioSignature validates a Twilio webhook request signature.
// See: https://www.twilio.com/docs/usage/security#validating-requests
func verifyTwilioSignature(authToken, signature, requestURL string, body []byte) bool {
	// Parse the form-encoded body to get parameters.
	params, err := url.ParseQuery(string(body))
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse form body for signature verification")
		return false
	}

	// Sort parameter keys alphabetically and concatenate key=value pairs.
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataBuilder strings.Builder
	dataBuilder.WriteString(requestURL)
	for _, k := range keys {
		dataBuilder.WriteString(k)
		dataBuilder.WriteString(params.Get(k))
	}

	// Compute HMAC-SHA1 of the data using the auth token.
	mac := hmac.New(sha1.New, []byte(authToken))
	mac.Write([]byte(dataBuilder.String()))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}
