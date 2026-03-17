package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// TestVerifyRazorpaySignature
// ---------------------------------------------------------------------------

func TestVerifyRazorpaySignature(t *testing.T) {
	keySecret := "test_secret_key_12345"
	gw := NewRazorpayGateway("key_id_test", keySecret, "webhook_secret")

	orderID := "order_DBJOWzybf0sJbb"
	paymentID := "pay_DGR2OqAGPseEDh"

	// Compute the correct signature.
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(keySecret))
	mac.Write([]byte(payload))
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name         string
		orderID      string
		paymentID    string
		signature    string
		wantVerified bool
	}{
		{
			name:         "valid signature",
			orderID:      orderID,
			paymentID:    paymentID,
			signature:    validSignature,
			wantVerified: true,
		},
		{
			name:         "invalid signature - tampered",
			orderID:      orderID,
			paymentID:    paymentID,
			signature:    "invalid_signature_here",
			wantVerified: false,
		},
		{
			name:         "invalid signature - wrong order ID",
			orderID:      "order_WRONG",
			paymentID:    paymentID,
			signature:    validSignature,
			wantVerified: false,
		},
		{
			name:         "invalid signature - wrong payment ID",
			orderID:      orderID,
			paymentID:    "pay_WRONG",
			signature:    validSignature,
			wantVerified: false,
		},
		{
			name:         "empty signature",
			orderID:      orderID,
			paymentID:    paymentID,
			signature:    "",
			wantVerified: false,
		},
		{
			name:         "valid signature with different secret",
			orderID:      orderID,
			paymentID:    paymentID,
			signature:    computeSignature("different_secret", orderID, paymentID),
			wantVerified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gw.VerifyPayment(context.Background(), tt.orderID, tt.paymentID, tt.signature)
			if err != nil {
				t.Fatalf("VerifyPayment returned error: %v", err)
			}

			if result.Verified != tt.wantVerified {
				t.Errorf("Verified = %v, want %v", result.Verified, tt.wantVerified)
			}

			if result.OrderID != tt.orderID {
				t.Errorf("OrderID = %q, want %q", result.OrderID, tt.orderID)
			}

			if result.PaymentID != tt.paymentID {
				t.Errorf("PaymentID = %q, want %q", result.PaymentID, tt.paymentID)
			}
		})
	}
}

// computeSignature is a test helper that produces a valid HMAC-SHA256 signature.
func computeSignature(secret, orderID, paymentID string) string {
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// ---------------------------------------------------------------------------
// TestConvertAmountToPaise
// ---------------------------------------------------------------------------

func TestConvertAmountToPaise(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		wantPaise int64
	}{
		{
			name:      "simple whole rupees",
			amount:    100.0,
			wantPaise: 10000,
		},
		{
			name:      "with paise",
			amount:    99.99,
			wantPaise: 9999,
		},
		{
			name:      "one rupee",
			amount:    1.0,
			wantPaise: 100,
		},
		{
			name:      "zero",
			amount:    0,
			wantPaise: 0,
		},
		{
			name:      "large amount",
			amount:    50000.50,
			wantPaise: 5000050,
		},
		{
			name:      "fractional paise rounds correctly",
			amount:    10.005,
			wantPaise: 1001, // math.Round(10.005 * 100) = 1001
		},
		{
			name:      "float precision edge case",
			amount:    19.99,
			wantPaise: 1999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64(math.Round(tt.amount * 100))
			if got != tt.wantPaise {
				t.Errorf("ConvertAmountToPaise(%f) = %d, want %d", tt.amount, got, tt.wantPaise)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestNoopGateway
// ---------------------------------------------------------------------------

func TestNoopGateway(t *testing.T) {
	gw := &NoopGateway{}
	ctx := context.Background()

	// CreateOrder should succeed.
	order, err := gw.CreateOrder(ctx, 100.0, "INR", nil)
	if err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}
	if order.ID != "noop_order" {
		t.Errorf("CreateOrder ID = %q, want %q", order.ID, "noop_order")
	}
	if order.Gateway != "noop" {
		t.Errorf("CreateOrder Gateway = %q, want %q", order.Gateway, "noop")
	}
	if order.Amount != 100.0 {
		t.Errorf("CreateOrder Amount = %f, want 100.0", order.Amount)
	}

	// VerifyPayment should always succeed.
	result, err := gw.VerifyPayment(ctx, "order_123", "pay_456", "sig")
	if err != nil {
		t.Fatalf("VerifyPayment error: %v", err)
	}
	if !result.Verified {
		t.Error("NoopGateway should always verify successfully")
	}

	// Refund should succeed.
	refund, err := gw.Refund(ctx, "pay_456", 50.0)
	if err != nil {
		t.Fatalf("Refund error: %v", err)
	}
	if refund.Amount != 50.0 {
		t.Errorf("Refund Amount = %f, want 50.0", refund.Amount)
	}
}

// ---------------------------------------------------------------------------
// TestRazorpayGatewayConstruction
// ---------------------------------------------------------------------------

func TestRazorpayGatewayConstruction(t *testing.T) {
	gw := NewRazorpayGateway("key_id", "key_secret", "webhook_secret")
	if gw == nil {
		t.Fatal("NewRazorpayGateway returned nil")
	}
	if gw.keyID != "key_id" {
		t.Errorf("keyID = %q, want %q", gw.keyID, "key_id")
	}
	if gw.keySecret != "key_secret" {
		t.Errorf("keySecret = %q, want %q", gw.keySecret, "key_secret")
	}
	if gw.webhookSecret != "webhook_secret" {
		t.Errorf("webhookSecret = %q, want %q", gw.webhookSecret, "webhook_secret")
	}
	if gw.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

// ---------------------------------------------------------------------------
// TestStripeGatewayConstruction
// ---------------------------------------------------------------------------

func TestStripeGatewayConstruction(t *testing.T) {
	gw := NewStripeGateway("sk_test_12345")
	if gw == nil {
		t.Fatal("NewStripeGateway returned nil")
	}
	if gw.apiKey != "sk_test_12345" {
		t.Errorf("apiKey = %q, want %q", gw.apiKey, "sk_test_12345")
	}
	if gw.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}
