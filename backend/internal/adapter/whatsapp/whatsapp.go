package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// whatsappAPIBase is the base URL for the Meta WhatsApp Cloud API. It is a
// variable so that tests can override it.
var whatsappAPIBase = "https://graph.facebook.com/v18.0"

// ---------------------------------------------------------------------------
// Provider interface
// ---------------------------------------------------------------------------

// Provider defines the contract for sending WhatsApp messages via the Cloud API.
type Provider interface {
	SendTextMessage(ctx context.Context, to, message string) error
	SendTemplateMessage(ctx context.Context, to, templateName, languageCode string, params []string) error
	SendInteractiveMessage(ctx context.Context, to string, msg InteractiveMessage) error
}

// InteractiveMessage represents a WhatsApp interactive message (button or list).
type InteractiveMessage struct {
	Type    string   // "button" or "list"
	Header  string
	Body    string
	Footer  string
	Buttons []Button // max 3 for button type
}

// Button represents a reply button in an interactive message.
type Button struct {
	ID    string
	Title string
}

// IncomingMessage represents a parsed inbound WhatsApp message from the webhook.
type IncomingMessage struct {
	From      string // sender phone number
	MessageID string
	Timestamp string
	Type      string // "text", "button", "interactive", "image", etc.
	Text      string // body text (for text and button messages)
	ButtonID  string // reply button id (for interactive replies)
}

// ---------------------------------------------------------------------------
// Cloud API provider (Meta's official API)
// ---------------------------------------------------------------------------

// CloudProvider sends messages via Meta's WhatsApp Cloud API.
type CloudProvider struct {
	phoneNumberID string
	accessToken   string
	businessID    string
	httpClient    *http.Client
}

// NewCloudProvider constructs a CloudProvider from the application config.
func NewCloudProvider(cfg *config.Config) *CloudProvider {
	return &CloudProvider{
		phoneNumberID: cfg.WhatsAppPhoneNumberID,
		accessToken:   cfg.WhatsAppAccessToken,
		businessID:    cfg.WhatsAppBusinessID,
		httpClient:    &http.Client{},
	}
}

// sendRequest is a helper that POSTs a JSON payload to the messages endpoint.
func (c *CloudProvider) sendRequest(ctx context.Context, payload interface{}) error {
	url := fmt.Sprintf("%s/%s/messages", whatsappAPIBase, c.phoneNumberID)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("whatsapp: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("whatsapp: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp: API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	log.Debug().
		Str("provider", "whatsapp").
		Int("status", resp.StatusCode).
		Msg("WhatsApp message sent successfully")

	return nil
}

// SendTextMessage sends a plain text message to the specified phone number.
func (c *CloudProvider) SendTextMessage(ctx context.Context, to, message string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}

	if err := c.sendRequest(ctx, payload); err != nil {
		return err
	}

	log.Info().
		Str("provider", "whatsapp").
		Str("to", to).
		Msg("text message sent")

	return nil
}

// SendTemplateMessage sends a pre-approved template message.
func (c *CloudProvider) SendTemplateMessage(ctx context.Context, to, templateName, languageCode string, params []string) error {
	// Build template parameter components.
	var parameters []map[string]string
	for _, p := range params {
		parameters = append(parameters, map[string]string{
			"type": "text",
			"text": p,
		})
	}

	var components []map[string]interface{}
	if len(parameters) > 0 {
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": parameters,
		})
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": languageCode,
			},
			"components": components,
		},
	}

	if err := c.sendRequest(ctx, payload); err != nil {
		return err
	}

	log.Info().
		Str("provider", "whatsapp").
		Str("to", to).
		Str("template", templateName).
		Msg("template message sent")

	return nil
}

// SendInteractiveMessage sends an interactive message with buttons or a list.
func (c *CloudProvider) SendInteractiveMessage(ctx context.Context, to string, msg InteractiveMessage) error {
	interactive := map[string]interface{}{
		"type": msg.Type,
		"body": map[string]string{
			"text": msg.Body,
		},
	}

	if msg.Header != "" {
		interactive["header"] = map[string]string{
			"type": "text",
			"text": msg.Header,
		}
	}

	if msg.Footer != "" {
		interactive["footer"] = map[string]string{
			"text": msg.Footer,
		}
	}

	if msg.Type == "button" && len(msg.Buttons) > 0 {
		// Enforce max 3 buttons per WhatsApp API spec.
		buttons := msg.Buttons
		if len(buttons) > 3 {
			buttons = buttons[:3]
		}

		var replyButtons []map[string]interface{}
		for _, btn := range buttons {
			replyButtons = append(replyButtons, map[string]interface{}{
				"type": "reply",
				"reply": map[string]string{
					"id":    btn.ID,
					"title": btn.Title,
				},
			})
		}
		interactive["action"] = map[string]interface{}{
			"buttons": replyButtons,
		}
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive":       interactive,
	}

	if err := c.sendRequest(ctx, payload); err != nil {
		return err
	}

	log.Info().
		Str("provider", "whatsapp").
		Str("to", to).
		Str("interactive_type", msg.Type).
		Msg("interactive message sent")

	return nil
}

// ---------------------------------------------------------------------------
// Template message helpers for key Seva events
// ---------------------------------------------------------------------------

// SendOTPMessage sends an OTP verification template message.
func SendOTPMessage(p Provider, ctx context.Context, to, otp string) error {
	return p.SendTemplateMessage(ctx, to, "otp_verification", "en", []string{otp})
}

// SendJobAlertToProvider sends a new job alert to a provider.
func SendJobAlertToProvider(p Provider, ctx context.Context, to, category, area string) error {
	return p.SendTemplateMessage(ctx, to, "new_job_alert", "en", []string{category, area})
}

// SendQuoteNotification sends a quote notification to a customer.
func SendQuoteNotification(p Provider, ctx context.Context, to, providerName, amount string) error {
	return p.SendTemplateMessage(ctx, to, "quote_received", "en", []string{providerName, amount})
}

// SendJobStatusUpdate sends a job status update to the customer.
func SendJobStatusUpdate(p Provider, ctx context.Context, to, status, jobTitle string) error {
	return p.SendTemplateMessage(ctx, to, "job_status_update", "en", []string{jobTitle, status})
}

// ---------------------------------------------------------------------------
// Webhook signature verification
// ---------------------------------------------------------------------------

// VerifyWebhookSignature verifies the X-Hub-Signature-256 header from Meta.
func VerifyWebhookSignature(payload []byte, signature, appSecret string) bool {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ParseWebhookPayload extracts incoming messages from the webhook JSON payload.
func ParseWebhookPayload(body []byte) ([]IncomingMessage, error) {
	var payload struct {
		Object string `json:"object"`
		Entry  []struct {
			ID      string `json:"id"`
			Changes []struct {
				Value struct {
					MessagingProduct string `json:"messaging_product"`
					Metadata         struct {
						DisplayPhoneNumber string `json:"display_phone_number"`
						PhoneNumberID      string `json:"phone_number_id"`
					} `json:"metadata"`
					Messages []struct {
						From      string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      *struct {
							Body string `json:"body"`
						} `json:"text,omitempty"`
						Interactive *struct {
							Type        string `json:"type"`
							ButtonReply *struct {
								ID    string `json:"id"`
								Title string `json:"title"`
							} `json:"button_reply,omitempty"`
							ListReply *struct {
								ID          string `json:"id"`
								Title       string `json:"title"`
								Description string `json:"description"`
							} `json:"list_reply,omitempty"`
						} `json:"interactive,omitempty"`
						Button *struct {
							Text    string `json:"text"`
							Payload string `json:"payload"`
						} `json:"button,omitempty"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("whatsapp: failed to parse webhook payload: %w", err)
	}

	var messages []IncomingMessage
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				incoming := IncomingMessage{
					From:      msg.From,
					MessageID: msg.ID,
					Timestamp: msg.Timestamp,
					Type:      msg.Type,
				}

				switch msg.Type {
				case "text":
					if msg.Text != nil {
						incoming.Text = msg.Text.Body
					}
				case "interactive":
					if msg.Interactive != nil {
						if msg.Interactive.ButtonReply != nil {
							incoming.ButtonID = msg.Interactive.ButtonReply.ID
							incoming.Text = msg.Interactive.ButtonReply.Title
						} else if msg.Interactive.ListReply != nil {
							incoming.ButtonID = msg.Interactive.ListReply.ID
							incoming.Text = msg.Interactive.ListReply.Title
						}
					}
				case "button":
					if msg.Button != nil {
						incoming.Text = msg.Button.Text
						incoming.ButtonID = msg.Button.Payload
					}
				}

				messages = append(messages, incoming)
			}
		}
	}

	return messages, nil
}

// ---------------------------------------------------------------------------
// Noop provider (development / testing)
// ---------------------------------------------------------------------------

// NoopProvider is a no-op implementation used during development and testing.
type NoopProvider struct{}

// SendTextMessage logs the message without actually sending it.
func (n *NoopProvider) SendTextMessage(ctx context.Context, to, message string) error {
	log.Info().
		Str("provider", "whatsapp-noop").
		Str("to", to).
		Str("message", message).
		Msg("text message (noop)")
	return nil
}

// SendTemplateMessage logs the template without actually sending it.
func (n *NoopProvider) SendTemplateMessage(ctx context.Context, to, templateName, languageCode string, params []string) error {
	log.Info().
		Str("provider", "whatsapp-noop").
		Str("to", to).
		Str("template", templateName).
		Str("language", languageCode).
		Strs("params", params).
		Msg("template message (noop)")
	return nil
}

// SendInteractiveMessage logs the interactive message without actually sending it.
func (n *NoopProvider) SendInteractiveMessage(ctx context.Context, to string, msg InteractiveMessage) error {
	log.Info().
		Str("provider", "whatsapp-noop").
		Str("to", to).
		Str("type", msg.Type).
		Str("body", msg.Body).
		Msg("interactive message (noop)")
	return nil
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

// NewProvider returns a WhatsApp Provider based on the application configuration.
// Falls back to NoopProvider if credentials are not configured.
func NewProvider(cfg *config.Config) Provider {
	if cfg.WhatsAppPhoneNumberID == "" || cfg.WhatsAppAccessToken == "" {
		log.Warn().Msg("WhatsApp credentials not configured, using noop provider")
		return &NoopProvider{}
	}
	return NewCloudProvider(cfg)
}
