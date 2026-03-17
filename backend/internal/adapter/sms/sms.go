package sms

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// SMSProvider is the interface for sending text messages.
type SMSProvider interface {
	SendSMS(to, message string) error
}

// NewSMSProvider returns a concrete SMSProvider based on the provider name.
func NewSMSProvider(providerName, apiKey string) (SMSProvider, error) {
	switch providerName {
	case "twilio":
		return &TwilioProvider{apiKey: apiKey}, nil
	case "msg91":
		return &MSG91Provider{apiKey: apiKey}, nil
	case "noop":
		return &NoopProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", providerName)
	}
}

// TwilioProvider sends SMS via the Twilio API.
type TwilioProvider struct {
	apiKey string
	// TODO: add account SID, auth token, from number
}

// SendSMS dispatches a message through Twilio.
func (t *TwilioProvider) SendSMS(to, message string) error {
	// TODO: implement Twilio REST API call
	// POST https://api.twilio.com/2010-04-01/Accounts/{AccountSid}/Messages.json
	log.Info().Str("provider", "twilio").Str("to", to).Msg("sending SMS")
	return nil
}

// MSG91Provider sends SMS via the MSG91 API.
type MSG91Provider struct {
	apiKey string
	// TODO: add sender ID, template ID, route
}

// SendSMS dispatches a message through MSG91.
func (m *MSG91Provider) SendSMS(to, message string) error {
	// TODO: implement MSG91 API call
	log.Info().Str("provider", "msg91").Str("to", to).Msg("sending SMS")
	return nil
}

// NoopProvider is a no-op implementation used during development and testing.
type NoopProvider struct{}

// SendSMS logs the message without actually sending it.
func (n *NoopProvider) SendSMS(to, message string) error {
	log.Info().Str("provider", "noop").Str("to", to).Str("message", message).Msg("SMS (noop)")
	return nil
}
