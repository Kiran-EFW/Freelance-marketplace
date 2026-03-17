package push

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// Provider defines the interface for sending push notifications.
type Provider interface {
	SendToDevice(ctx context.Context, token string, notification Notification) error
	SendToTopic(ctx context.Context, topic string, notification Notification) error
}

// Notification represents a push notification payload.
type Notification struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	ImageURL string            `json:"image_url,omitempty"`
}

// ErrInvalidToken is returned when FCM reports the device token is invalid.
var ErrInvalidToken = fmt.Errorf("push: device token is invalid or unregistered")

// ---------------------------------------------------------------------------
// FCM HTTP v1 Provider
// ---------------------------------------------------------------------------

// fcmAPIBase is the base URL for FCM HTTP v1 API. It is a variable so tests
// can override it.
var fcmAPIBase = "https://fcm.googleapis.com/v1"

// tokenEndpoint is the Google OAuth2 token endpoint.
var tokenEndpoint = "https://oauth2.googleapis.com/token"

// ServiceAccountKey represents the fields we need from a Google service account
// JSON key file.
type ServiceAccountKey struct {
	Type                string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url"`
}

// FCMProvider sends push notifications via FCM HTTP v1 API using service
// account credentials.
type FCMProvider struct {
	projectID  string
	saKey      *ServiceAccountKey
	privateKey *rsa.PrivateKey
	httpClient *http.Client

	// OAuth2 token caching
	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

// NewFCMProvider creates a new FCMProvider from a service account key JSON.
// The keyJSON should be the raw contents of a Google Cloud service account key file.
func NewFCMProvider(projectID string, keyJSON []byte) (*FCMProvider, error) {
	var saKey ServiceAccountKey
	if err := json.Unmarshal(keyJSON, &saKey); err != nil {
		return nil, fmt.Errorf("fcm: failed to parse service account key: %w", err)
	}

	if saKey.PrivateKey == "" {
		return nil, fmt.Errorf("fcm: service account key is missing private_key field")
	}

	// Parse the PEM-encoded RSA private key.
	block, _ := pem.Decode([]byte(saKey.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("fcm: failed to decode PEM block from private key")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("fcm: failed to parse private key: %w", err)
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("fcm: private key is not RSA")
	}

	// Use projectID from parameter, fallback to the one in the key file.
	if projectID == "" {
		projectID = saKey.ProjectID
	}

	return &FCMProvider{
		projectID:  projectID,
		saKey:      &saKey,
		privateKey: rsaKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SendToDevice sends a notification to a specific device token.
func (f *FCMProvider) SendToDevice(ctx context.Context, token string, notification Notification) error {
	msg := fcmMessage{
		Message: fcmMessageBody{
			Token:        token,
			Notification: &fcmNotification{Title: notification.Title, Body: notification.Body},
			Data:         notification.Data,
			Android:      &fcmAndroid{Priority: "HIGH"},
			APNS: &fcmAPNS{
				Payload: &fcmAPNSPayload{
					APS: &fcmAPS{Sound: "default"},
				},
			},
		},
	}

	if notification.ImageURL != "" {
		msg.Message.Notification.Image = notification.ImageURL
	}

	return f.send(ctx, msg)
}

// SendToTopic sends a notification to all subscribers of a topic.
func (f *FCMProvider) SendToTopic(ctx context.Context, topic string, notification Notification) error {
	msg := fcmMessage{
		Message: fcmMessageBody{
			Topic:        topic,
			Notification: &fcmNotification{Title: notification.Title, Body: notification.Body},
			Data:         notification.Data,
			Android:      &fcmAndroid{Priority: "HIGH"},
			APNS: &fcmAPNS{
				Payload: &fcmAPNSPayload{
					APS: &fcmAPS{Sound: "default"},
				},
			},
		},
	}

	if notification.ImageURL != "" {
		msg.Message.Notification.Image = notification.ImageURL
	}

	return f.send(ctx, msg)
}

// send dispatches a message to the FCM HTTP v1 API.
func (f *FCMProvider) send(ctx context.Context, msg fcmMessage) error {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("fcm: failed to get access token: %w", err)
	}

	apiURL := fmt.Sprintf("%s/projects/%s/messages:send", fcmAPIBase, f.projectID)

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("fcm: failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fcm: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fcm: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp fcmErrorResponse
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil {
			// Check if the error indicates an invalid/unregistered token.
			if errResp.Error.Code == 404 || errResp.Error.Code == 400 {
				for _, detail := range errResp.Error.Details {
					if detail.ErrorCode == "UNREGISTERED" || detail.ErrorCode == "INVALID_ARGUMENT" {
						log.Warn().
							Str("error_code", detail.ErrorCode).
							Msg("FCM token is invalid or unregistered")
						return ErrInvalidToken
					}
				}
			}

			return fmt.Errorf("fcm: API error (HTTP %d): code=%d status=%s message=%s",
				resp.StatusCode, errResp.Error.Code, errResp.Error.Status, errResp.Error.Message)
		}
		return fmt.Errorf("fcm: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Debug().
		Str("provider", "fcm").
		Int("status", resp.StatusCode).
		Msg("push notification sent successfully")

	return nil
}

// getAccessToken returns a cached OAuth2 access token or generates a new one
// using the service account credentials.
func (f *FCMProvider) getAccessToken(ctx context.Context) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Return cached token if still valid (with 5-minute buffer).
	if f.accessToken != "" && time.Now().Before(f.tokenExpiry.Add(-5*time.Minute)) {
		return f.accessToken, nil
	}

	// Generate a new JWT assertion.
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    f.saKey.ClientEmail,
		Subject:   f.saKey.ClientEmail,
		Audience:  jwt.ClaimStrings{tokenEndpoint},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
	}

	// Add the scope as a custom claim.
	type scopedClaims struct {
		jwt.RegisteredClaims
		Scope string `json:"scope"`
	}

	sc := scopedClaims{
		RegisteredClaims: claims,
		Scope:            "https://www.googleapis.com/auth/firebase.messaging",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, sc)
	token.Header["kid"] = f.saKey.PrivateKeyID

	signedJWT, err := token.SignedString(f.privateKey)
	if err != nil {
		return "", fmt.Errorf("fcm: failed to sign JWT: %w", err)
	}

	// Exchange the JWT for an access token.
	formData := fmt.Sprintf("grant_type=urn%%3Aietf%%3Aparams%%3Aoauth%%3Agrant-type%%3Ajwt-bearer&assertion=%s", signedJWT)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, bytes.NewBufferString(formData))
	if err != nil {
		return "", fmt.Errorf("fcm: failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fcm: token exchange HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fcm: failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fcm: token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("fcm: failed to parse token response: %w", err)
	}

	f.accessToken = tokenResp.AccessToken
	f.tokenExpiry = now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	log.Info().
		Str("provider", "fcm").
		Int("expires_in", tokenResp.ExpiresIn).
		Msg("obtained new FCM access token")

	return f.accessToken, nil
}

// ---------------------------------------------------------------------------
// FCM API types
// ---------------------------------------------------------------------------

type fcmMessage struct {
	Message fcmMessageBody `json:"message"`
}

type fcmMessageBody struct {
	Token        string           `json:"token,omitempty"`
	Topic        string           `json:"topic,omitempty"`
	Notification *fcmNotification `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Android      *fcmAndroid      `json:"android,omitempty"`
	APNS         *fcmAPNS         `json:"apns,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Image string `json:"image,omitempty"`
}

type fcmAndroid struct {
	Priority string `json:"priority,omitempty"`
}

type fcmAPNS struct {
	Payload *fcmAPNSPayload `json:"payload,omitempty"`
}

type fcmAPNSPayload struct {
	APS *fcmAPS `json:"aps,omitempty"`
}

type fcmAPS struct {
	Sound string `json:"sound,omitempty"`
}

type fcmErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
		Details []struct {
			ErrorCode string `json:"errorCode"`
		} `json:"details"`
	} `json:"error"`
}

// ---------------------------------------------------------------------------
// Noop Provider (development / testing)
// ---------------------------------------------------------------------------

// NoopProvider is a no-op implementation used during development and testing.
type NoopProvider struct{}

// SendToDevice logs the notification without actually sending it.
func (n *NoopProvider) SendToDevice(_ context.Context, token string, notification Notification) error {
	log.Info().
		Str("provider", "push_noop").
		Str("token", token).
		Str("title", notification.Title).
		Str("body", notification.Body).
		Msg("push notification (noop)")
	return nil
}

// SendToTopic logs the notification without actually sending it.
func (n *NoopProvider) SendToTopic(_ context.Context, topic string, notification Notification) error {
	log.Info().
		Str("provider", "push_noop").
		Str("topic", topic).
		Str("title", notification.Title).
		Str("body", notification.Body).
		Msg("push notification to topic (noop)")
	return nil
}

// NewPushProvider creates a push Provider based on the given configuration.
// If the FCM service account key is provided, it creates a real FCM provider;
// otherwise it falls back to the NoopProvider.
func NewPushProvider(projectID string, serviceAccountKeyJSON []byte) Provider {
	if len(serviceAccountKeyJSON) == 0 {
		log.Warn().Msg("FCM service account key not configured, using noop push provider")
		return &NoopProvider{}
	}

	provider, err := NewFCMProvider(projectID, serviceAccountKeyJSON)
	if err != nil {
		log.Error().Err(err).Msg("failed to create FCM provider, falling back to noop")
		return &NoopProvider{}
	}

	log.Info().Str("project_id", projectID).Msg("FCM push provider initialized")
	return provider
}
