package payment

import (
	"context"

	"github.com/rs/zerolog/log"
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

// NewPaymentGateway returns a concrete PaymentGateway based on the provider name.
func NewPaymentGateway(provider, apiKey string) PaymentGateway {
	switch provider {
	case "razorpay":
		return &RazorpayGateway{apiKey: apiKey}
	case "stripe":
		return &StripeGateway{apiKey: apiKey}
	default:
		log.Warn().Str("provider", provider).Msg("unknown payment provider, using noop gateway")
		return &NoopGateway{}
	}
}

// RazorpayGateway integrates with the Razorpay payment platform.
type RazorpayGateway struct {
	apiKey string
	// TODO: add key secret, webhook secret
}

// CreateOrder creates a Razorpay order.
func (r *RazorpayGateway) CreateOrder(ctx context.Context, amount float64, currency string, metadata map[string]string) (*Order, error) {
	// TODO: POST /v1/orders to Razorpay API
	log.Info().Str("gateway", "razorpay").Float64("amount", amount).Msg("creating order")
	return &Order{ID: "rzp_stub_order", Amount: amount, Currency: currency, Status: "created", Gateway: "razorpay"}, nil
}

// VerifyPayment verifies a Razorpay payment signature.
func (r *RazorpayGateway) VerifyPayment(ctx context.Context, orderID, paymentID, signature string) (*VerificationResult, error) {
	// TODO: verify HMAC-SHA256 signature
	log.Info().Str("gateway", "razorpay").Str("order_id", orderID).Msg("verifying payment")
	return &VerificationResult{Verified: true, OrderID: orderID, PaymentID: paymentID}, nil
}

// Refund processes a refund through Razorpay.
func (r *RazorpayGateway) Refund(ctx context.Context, paymentID string, amount float64) (*RefundResult, error) {
	// TODO: POST /v1/payments/{id}/refund to Razorpay API
	log.Info().Str("gateway", "razorpay").Str("payment_id", paymentID).Msg("processing refund")
	return &RefundResult{RefundID: "rzp_stub_refund", Amount: amount, Status: "processed"}, nil
}

// StripeGateway integrates with the Stripe payment platform.
type StripeGateway struct {
	apiKey string
	// TODO: add webhook signing secret
}

// CreateOrder creates a Stripe PaymentIntent.
func (s *StripeGateway) CreateOrder(ctx context.Context, amount float64, currency string, metadata map[string]string) (*Order, error) {
	// TODO: create Stripe PaymentIntent
	log.Info().Str("gateway", "stripe").Float64("amount", amount).Msg("creating payment intent")
	return &Order{ID: "pi_stub_order", Amount: amount, Currency: currency, Status: "created", Gateway: "stripe"}, nil
}

// VerifyPayment verifies a Stripe payment.
func (s *StripeGateway) VerifyPayment(ctx context.Context, orderID, paymentID, signature string) (*VerificationResult, error) {
	// TODO: verify via Stripe webhook or retrieve PaymentIntent
	log.Info().Str("gateway", "stripe").Str("order_id", orderID).Msg("verifying payment")
	return &VerificationResult{Verified: true, OrderID: orderID, PaymentID: paymentID}, nil
}

// Refund processes a refund through Stripe.
func (s *StripeGateway) Refund(ctx context.Context, paymentID string, amount float64) (*RefundResult, error) {
	// TODO: create Stripe Refund
	log.Info().Str("gateway", "stripe").Str("payment_id", paymentID).Msg("processing refund")
	return &RefundResult{RefundID: "re_stub_refund", Amount: amount, Status: "succeeded"}, nil
}

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
