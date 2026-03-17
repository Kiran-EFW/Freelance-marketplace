package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// Order represents a payment order created with the gateway.
type Order struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	Gateway  string  `json:"gateway"`
}

// VerificationResult holds the outcome of a payment verification.
type VerificationResult struct {
	Verified  bool   `json:"verified"`
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
}

// RefundResult holds the outcome of a refund request.
type RefundResult struct {
	RefundID string  `json:"refund_id"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status"`
}

// PaymentGateway defines the interface for payment processing.
type PaymentGateway interface {
	CreateOrder(ctx context.Context, amount float64, currency string, metadata map[string]string) (*Order, error)
	VerifyPayment(ctx context.Context, orderID, paymentID, signature string) (*VerificationResult, error)
	Refund(ctx context.Context, paymentID string, amount float64) (*RefundResult, error)
}

// NewPaymentGateway returns a concrete PaymentGateway based on the application
// configuration. It uses the config to determine the provider and retrieve
// the necessary credentials.
func NewPaymentGateway(cfg *config.Config) PaymentGateway {
	switch cfg.PaymentProvider {
	case "razorpay":
		if cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "" {
			log.Warn().Msg("razorpay credentials not configured, using noop gateway")
			return &NoopGateway{}
		}
		return NewRazorpayGateway(cfg.RazorpayKeyID, cfg.RazorpayKeySecret, cfg.RazorpayWebhookSecret)
	case "stripe":
		if cfg.StripeAPIKey == "" {
			log.Warn().Msg("stripe API key not configured, using noop gateway")
			return &NoopGateway{}
		}
		return NewStripeGateway(cfg.StripeAPIKey)
	default:
		log.Warn().Str("provider", cfg.PaymentProvider).Msg("unknown payment provider, using noop gateway")
		return &NoopGateway{}
	}
}

// ===========================================================================
// Razorpay
// ===========================================================================

// razorpayAPIBase is the base URL for the Razorpay API. It is a variable so
// that tests can override it.
var razorpayAPIBase = "https://api.razorpay.com/v1"

// RazorpayGateway integrates with the Razorpay payment platform.
type RazorpayGateway struct {
	keyID         string
	keySecret     string
	webhookSecret string
	httpClient    *http.Client
}

// NewRazorpayGateway constructs a RazorpayGateway with the given credentials.
func NewRazorpayGateway(keyID, keySecret, webhookSecret string) *RazorpayGateway {
	return &RazorpayGateway{
		keyID:         keyID,
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// razorpayOrderRequest is the JSON body sent to create an order.
type razorpayOrderRequest struct {
	Amount   int64             `json:"amount"`
	Currency string            `json:"currency"`
	Receipt  string            `json:"receipt,omitempty"`
	Notes    map[string]string `json:"notes,omitempty"`
}

// razorpayOrderResponse is the relevant portion of the Razorpay order response.
type razorpayOrderResponse struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

// razorpayErrorResponse represents the error structure returned by Razorpay.
type razorpayErrorResponse struct {
	Error struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Step        string `json:"step"`
		Reason      string `json:"reason"`
	} `json:"error"`
}

// razorpayRefundResponse holds the relevant refund fields.
type razorpayRefundResponse struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

// CreateOrder creates a Razorpay order via POST /v1/orders.
// The amount parameter is in the major currency unit (e.g. rupees) and is
// converted to the smallest unit (paise) before sending to Razorpay.
func (r *RazorpayGateway) CreateOrder(ctx context.Context, amount float64, currency string, metadata map[string]string) (*Order, error) {
	apiURL := fmt.Sprintf("%s/orders", razorpayAPIBase)

	// Convert to smallest currency unit (paise for INR, cents for USD).
	amountInSmallest := int64(math.Round(amount * 100))

	receipt := ""
	if v, ok := metadata["receipt"]; ok {
		receipt = v
	} else {
		receipt = fmt.Sprintf("order_%d", time.Now().UnixMilli())
	}

	reqBody := razorpayOrderRequest{
		Amount:   amountInSmallest,
		Currency: strings.ToUpper(currency),
		Receipt:  receipt,
		Notes:    metadata,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to create request: %w", err)
	}

	req.SetBasicAuth(r.keyID, r.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("razorpay: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp razorpayErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return nil, fmt.Errorf("razorpay: API error (HTTP %d): code=%s description=%s reason=%s",
				resp.StatusCode, errResp.Error.Code, errResp.Error.Description, errResp.Error.Reason)
		}
		return nil, fmt.Errorf("razorpay: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var orderResp razorpayOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("razorpay: failed to parse order response: %w", err)
	}

	log.Info().
		Str("gateway", "razorpay").
		Str("order_id", orderResp.ID).
		Int64("amount_paise", orderResp.Amount).
		Str("currency", orderResp.Currency).
		Str("status", orderResp.Status).
		Msg("order created successfully")

	return &Order{
		ID:       orderResp.ID,
		Amount:   amount,
		Currency: orderResp.Currency,
		Status:   orderResp.Status,
		Gateway:  "razorpay",
	}, nil
}

// VerifyPayment verifies a Razorpay payment by computing the expected
// HMAC-SHA256 signature of "razorpay_order_id|razorpay_payment_id" using the
// key secret and comparing it to the signature provided by the client.
func (r *RazorpayGateway) VerifyPayment(_ context.Context, orderID, paymentID, signature string) (*VerificationResult, error) {
	// The payload to sign is "order_id|payment_id".
	payload := orderID + "|" + paymentID

	mac := hmac.New(sha256.New, []byte(r.keySecret))
	mac.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	verified := hmac.Equal([]byte(expectedSignature), []byte(signature))

	if !verified {
		log.Warn().
			Str("gateway", "razorpay").
			Str("order_id", orderID).
			Str("payment_id", paymentID).
			Msg("payment signature verification failed")

		return &VerificationResult{
			Verified:  false,
			OrderID:   orderID,
			PaymentID: paymentID,
		}, nil
	}

	log.Info().
		Str("gateway", "razorpay").
		Str("order_id", orderID).
		Str("payment_id", paymentID).
		Msg("payment signature verified successfully")

	return &VerificationResult{
		Verified:  true,
		OrderID:   orderID,
		PaymentID: paymentID,
	}, nil
}

// Refund processes a refund through Razorpay via POST /v1/payments/{id}/refund.
// The amount parameter is in the major currency unit and is converted to the
// smallest unit (paise) before sending.
func (r *RazorpayGateway) Refund(ctx context.Context, paymentID string, amount float64) (*RefundResult, error) {
	apiURL := fmt.Sprintf("%s/payments/%s/refund", razorpayAPIBase, paymentID)

	amountInSmallest := int64(math.Round(amount * 100))

	reqBody := map[string]int64{
		"amount": amountInSmallest,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to marshal refund request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to create refund request: %w", err)
	}

	req.SetBasicAuth(r.keyID, r.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("razorpay: refund HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("razorpay: failed to read refund response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp razorpayErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return nil, fmt.Errorf("razorpay: refund API error (HTTP %d): code=%s description=%s reason=%s",
				resp.StatusCode, errResp.Error.Code, errResp.Error.Description, errResp.Error.Reason)
		}
		return nil, fmt.Errorf("razorpay: refund unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var refundResp razorpayRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("razorpay: failed to parse refund response: %w", err)
	}

	log.Info().
		Str("gateway", "razorpay").
		Str("payment_id", paymentID).
		Str("refund_id", refundResp.ID).
		Int64("amount_paise", refundResp.Amount).
		Str("status", refundResp.Status).
		Msg("refund processed successfully")

	return &RefundResult{
		RefundID: refundResp.ID,
		Amount:   amount,
		Status:   refundResp.Status,
	}, nil
}

// ===========================================================================
// Stripe
// ===========================================================================

// stripeAPIBase is the base URL for the Stripe API. It is a variable so that
// tests can override it.
var stripeAPIBase = "https://api.stripe.com/v1"

// StripeGateway integrates with the Stripe payment platform.
type StripeGateway struct {
	apiKey     string
	httpClient *http.Client
}

// NewStripeGateway constructs a StripeGateway with the given secret API key.
func NewStripeGateway(apiKey string) *StripeGateway {
	return &StripeGateway{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// stripePaymentIntentResponse represents the relevant portion of a Stripe
// PaymentIntent response.
type stripePaymentIntentResponse struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

// stripeErrorResponse represents a Stripe API error.
type stripeErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// stripeRefundResponse represents the relevant portion of a Stripe Refund response.
type stripeRefundResponse struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
	Status string `json:"status"`
}

// CreateOrder creates a Stripe PaymentIntent via POST /v1/payment_intents.
// The amount is in the major currency unit and is converted to the smallest
// unit (cents) before sending.
func (s *StripeGateway) CreateOrder(ctx context.Context, amount float64, currency string, metadata map[string]string) (*Order, error) {
	apiURL := fmt.Sprintf("%s/payment_intents", stripeAPIBase)

	amountInSmallest := int64(math.Round(amount * 100))

	// Stripe uses form-encoded POST bodies.
	formData := url.Values{}
	formData.Set("amount", fmt.Sprintf("%d", amountInSmallest))
	formData.Set("currency", strings.ToLower(currency))
	for k, v := range metadata {
		formData.Set(fmt.Sprintf("metadata[%s]", k), v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp stripeErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return nil, fmt.Errorf("stripe: API error (HTTP %d): type=%s code=%s message=%s",
				resp.StatusCode, errResp.Error.Type, errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("stripe: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var piResp stripePaymentIntentResponse
	if err := json.Unmarshal(body, &piResp); err != nil {
		return nil, fmt.Errorf("stripe: failed to parse PaymentIntent response: %w", err)
	}

	log.Info().
		Str("gateway", "stripe").
		Str("payment_intent_id", piResp.ID).
		Int64("amount_cents", piResp.Amount).
		Str("currency", piResp.Currency).
		Str("status", piResp.Status).
		Msg("payment intent created successfully")

	return &Order{
		ID:       piResp.ID,
		Amount:   amount,
		Currency: piResp.Currency,
		Status:   piResp.Status,
		Gateway:  "stripe",
	}, nil
}

// VerifyPayment verifies a Stripe payment by retrieving the PaymentIntent
// from the Stripe API and checking its status.
func (s *StripeGateway) VerifyPayment(ctx context.Context, orderID, paymentID, _ string) (*VerificationResult, error) {
	// Retrieve the PaymentIntent to verify its status.
	apiURL := fmt.Sprintf("%s/payment_intents/%s", stripeAPIBase, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to create verification request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe: verification HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to read verification response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp stripeErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return nil, fmt.Errorf("stripe: verification API error (HTTP %d): type=%s code=%s message=%s",
				resp.StatusCode, errResp.Error.Type, errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("stripe: verification unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var piResp stripePaymentIntentResponse
	if err := json.Unmarshal(body, &piResp); err != nil {
		return nil, fmt.Errorf("stripe: failed to parse verification response: %w", err)
	}

	// A PaymentIntent is considered verified if its status is "succeeded".
	verified := piResp.Status == "succeeded"

	log.Info().
		Str("gateway", "stripe").
		Str("payment_intent_id", piResp.ID).
		Str("status", piResp.Status).
		Bool("verified", verified).
		Msg("payment verification completed")

	return &VerificationResult{
		Verified:  verified,
		OrderID:   orderID,
		PaymentID: paymentID,
	}, nil
}

// Refund processes a refund through Stripe via POST /v1/refunds.
// The amount is in the major currency unit and is converted to the smallest
// unit (cents) before sending.
func (s *StripeGateway) Refund(ctx context.Context, paymentID string, amount float64) (*RefundResult, error) {
	apiURL := fmt.Sprintf("%s/refunds", stripeAPIBase)

	amountInSmallest := int64(math.Round(amount * 100))

	formData := url.Values{}
	formData.Set("payment_intent", paymentID)
	formData.Set("amount", fmt.Sprintf("%d", amountInSmallest))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to create refund request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe: refund HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("stripe: failed to read refund response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp stripeErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil {
			return nil, fmt.Errorf("stripe: refund API error (HTTP %d): type=%s code=%s message=%s",
				resp.StatusCode, errResp.Error.Type, errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("stripe: refund unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var refundResp stripeRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, fmt.Errorf("stripe: failed to parse refund response: %w", err)
	}

	log.Info().
		Str("gateway", "stripe").
		Str("payment_id", paymentID).
		Str("refund_id", refundResp.ID).
		Int64("amount_cents", refundResp.Amount).
		Str("status", refundResp.Status).
		Msg("refund processed successfully")

	return &RefundResult{
		RefundID: refundResp.ID,
		Amount:   amount,
		Status:   refundResp.Status,
	}, nil
}

// ===========================================================================
// Noop (development / testing)
// ===========================================================================

// NoopGateway is a no-op implementation for development and testing.
type NoopGateway struct{}

func (n *NoopGateway) CreateOrder(_ context.Context, amount float64, currency string, _ map[string]string) (*Order, error) {
	return &Order{ID: "noop_order", Amount: amount, Currency: currency, Status: "created", Gateway: "noop"}, nil
}

func (n *NoopGateway) VerifyPayment(_ context.Context, orderID, paymentID, _ string) (*VerificationResult, error) {
	return &VerificationResult{Verified: true, OrderID: orderID, PaymentID: paymentID}, nil
}

func (n *NoopGateway) Refund(_ context.Context, paymentID string, amount float64) (*RefundResult, error) {
	return &RefundResult{RefundID: "noop_refund", Amount: amount, Status: "processed"}, nil
}
