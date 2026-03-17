package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// SMSProvider is the interface for sending text messages.
type SMSProvider interface {
	SendSMS(to, message string) error
}

// NewSMSProvider returns a concrete SMSProvider based on the provider name
// and the application configuration.
func NewSMSProvider(cfg *config.Config) (SMSProvider, error) {
	switch cfg.SMSProvider {
	case "twilio":
		if cfg.TwilioAccountSID == "" || cfg.TwilioAuthToken == "" || cfg.TwilioFromNumber == "" {
			return nil, fmt.Errorf("twilio provider requires TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, and TWILIO_FROM_NUMBER")
		}
		return NewTwilioProvider(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber), nil
	case "msg91":
		if cfg.SMSAPIKey == "" {
			return nil, fmt.Errorf("msg91 provider requires SMS_API_KEY")
		}
		return NewMSG91Provider(cfg.SMSAPIKey), nil
	case "noop":
		return &NoopProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", cfg.SMSProvider)
	}
}

// ---------------------------------------------------------------------------
// Twilio
// ---------------------------------------------------------------------------

// twilioAPIBase is the base URL for the Twilio REST API. It is a variable so
// that tests can override it.
var twilioAPIBase = "https://api.twilio.com/2010-04-01"

// TwilioProvider sends SMS via the Twilio REST API.
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// NewTwilioProvider constructs a TwilioProvider with the given credentials.
func NewTwilioProvider(accountSID, authToken, fromNumber string) *TwilioProvider {
	return &TwilioProvider{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{},
	}
}

// twilioErrorResponse represents an error returned by the Twilio API.
type twilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// twilioSuccessResponse contains the fields we care about on a successful send.
type twilioSuccessResponse struct {
	SID    string `json:"sid"`
	Status string `json:"status"`
}

// SendSMS dispatches a message through the Twilio Messages API.
func (t *TwilioProvider) SendSMS(to, message string) error {
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", twilioAPIBase, t.accountSID)

	// Build form-encoded body.
	formData := url.Values{}
	formData.Set("To", to)
	formData.Set("From", t.fromNumber)
	formData.Set("Body", message)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("twilio: failed to create request: %w", err)
	}

	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("twilio: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("twilio: failed to read response body: %w", err)
	}

	// Twilio returns 201 on success.
	if resp.StatusCode != http.StatusCreated {
		var errResp twilioErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return fmt.Errorf("twilio: API error (HTTP %d): code=%d message=%s info=%s",
				resp.StatusCode, errResp.Code, errResp.Message, errResp.MoreInfo)
		}
		return fmt.Errorf("twilio: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var success twilioSuccessResponse
	if err := json.Unmarshal(body, &success); err != nil {
		return fmt.Errorf("twilio: failed to parse success response: %w", err)
	}

	log.Info().
		Str("provider", "twilio").
		Str("to", to).
		Str("sid", success.SID).
		Str("status", success.Status).
		Msg("SMS sent successfully")

	return nil
}

// ---------------------------------------------------------------------------
// MSG91
// ---------------------------------------------------------------------------

// msg91APIBase is the base URL for the MSG91 API. It is a variable so that
// tests can override it.
var msg91APIBase = "https://control.msg91.com/api/v5"

// MSG91Provider sends SMS via the MSG91 Flow API.
type MSG91Provider struct {
	apiKey     string
	httpClient *http.Client
}

// NewMSG91Provider constructs an MSG91Provider with the given API key.
func NewMSG91Provider(apiKey string) *MSG91Provider {
	return &MSG91Provider{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// msg91Request is the JSON body sent to the MSG91 Flow API.
type msg91Request struct {
	FlowID  string `json:"flow_id"`
	Mobiles string `json:"mobiles"`
	Var     string `json:"var"`
}

// msg91Response represents the MSG91 API response.
type msg91Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SendSMS dispatches a message through the MSG91 Flow API.
// The message is expected to be in the format "flow_id:variable_value" so that
// the flow template and variable are separated. If no colon is present the
// entire message is treated as the variable value and a default flow ID is used.
func (m *MSG91Provider) SendSMS(to, message string) error {
	flowID := "default"
	varValue := message

	if parts := strings.SplitN(message, ":", 2); len(parts) == 2 {
		flowID = parts[0]
		varValue = parts[1]
	}

	apiURL := fmt.Sprintf("%s/flow/", msg91APIBase)

	payload := msg91Request{
		FlowID:  flowID,
		Mobiles: to,
		Var:     varValue,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msg91: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("msg91: failed to create request: %w", err)
	}

	req.Header.Set("authkey", m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("msg91: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("msg91: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp msg91Response
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return fmt.Errorf("msg91: API error (HTTP %d): type=%s message=%s",
				resp.StatusCode, errResp.Type, errResp.Message)
		}
		return fmt.Errorf("msg91: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result msg91Response
	if err := json.Unmarshal(body, &result); err != nil {
		log.Warn().Err(err).Msg("msg91: could not parse response body")
	}

	log.Info().
		Str("provider", "msg91").
		Str("to", to).
		Str("type", result.Type).
		Str("message", result.Message).
		Msg("SMS sent successfully")

	return nil
}

// ---------------------------------------------------------------------------
// Noop (development / testing)
// ---------------------------------------------------------------------------

// NoopProvider is a no-op implementation used during development and testing.
type NoopProvider struct{}

// SendSMS logs the message without actually sending it.
func (n *NoopProvider) SendSMS(to, message string) error {
	log.Info().Str("provider", "noop").Str("to", to).Str("message", message).Msg("SMS (noop)")
	return nil
}
