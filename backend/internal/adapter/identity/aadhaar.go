package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// VerificationResult holds the outcome of an Aadhaar verification.
type VerificationResult struct {
	Valid   bool   `json:"valid"`
	Name    string `json:"name"`
	DOB     string `json:"dob"`
	Gender  string `json:"gender"`
	Address string `json:"address"`
}

// AadhaarProvider defines the interface for Aadhaar verification operations.
type AadhaarProvider interface {
	GenerateOTP(ctx context.Context, aadhaarNumber string) error
	VerifyOTP(ctx context.Context, aadhaarNumber, otp string) (*VerificationResult, error)
}

// AadhaarVerifier provides Aadhaar verification via an external API
// (e.g., UIDAI-approved service provider).
type AadhaarVerifier struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewAadhaarVerifier creates a new Aadhaar verification client.
func NewAadhaarVerifier(apiKey, apiURL string) *AadhaarVerifier {
	return &AadhaarVerifier{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// aadhaarOTPRequest is the request body for Aadhaar OTP generation.
type aadhaarOTPRequest struct {
	AadhaarNumber string `json:"aadhaar_number"`
}

// aadhaarVerifyRequest is the request body for Aadhaar OTP verification.
type aadhaarVerifyRequest struct {
	AadhaarNumber string `json:"aadhaar_number"`
	OTP           string `json:"otp"`
}

// aadhaarResponse is the response body from the Aadhaar verification API.
type aadhaarResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Name    string `json:"name"`
		DOB     string `json:"dob"`
		Gender  string `json:"gender"`
		Address string `json:"address"`
	} `json:"data"`
}

// GenerateOTP sends an OTP to the mobile number registered with the Aadhaar.
func (a *AadhaarVerifier) GenerateOTP(ctx context.Context, aadhaarNumber string) error {
	reqBody := aadhaarOTPRequest{
		AadhaarNumber: aadhaarNumber,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("aadhaar generate otp marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.apiURL+"/otp/generate", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("aadhaar generate otp create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("aadhaar generate otp http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("aadhaar generate otp read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("aadhaar generate otp API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var aadhaarResp aadhaarResponse
	if err := json.Unmarshal(respBody, &aadhaarResp); err != nil {
		return fmt.Errorf("aadhaar generate otp unmarshal: %w", err)
	}

	if !aadhaarResp.Success {
		return fmt.Errorf("aadhaar generate otp failed: %s", aadhaarResp.Message)
	}

	log.Info().Msg("aadhaar OTP generated successfully")

	return nil
}

// VerifyOTP validates the OTP against the Aadhaar number and returns the
// verification result containing the user's details.
func (a *AadhaarVerifier) VerifyOTP(ctx context.Context, aadhaarNumber, otp string) (*VerificationResult, error) {
	reqBody := aadhaarVerifyRequest{
		AadhaarNumber: aadhaarNumber,
		OTP:           otp,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("aadhaar verify otp marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.apiURL+"/otp/verify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("aadhaar verify otp create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aadhaar verify otp http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("aadhaar verify otp read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aadhaar verify otp API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var aadhaarResp aadhaarResponse
	if err := json.Unmarshal(respBody, &aadhaarResp); err != nil {
		return nil, fmt.Errorf("aadhaar verify otp unmarshal: %w", err)
	}

	if !aadhaarResp.Success {
		return &VerificationResult{Valid: false}, nil
	}

	result := &VerificationResult{
		Valid:   true,
		Name:    aadhaarResp.Data.Name,
		DOB:     aadhaarResp.Data.DOB,
		Gender:  aadhaarResp.Data.Gender,
		Address: aadhaarResp.Data.Address,
	}

	log.Info().Bool("valid", result.Valid).Msg("aadhaar OTP verification completed")

	return result, nil
}

// --- NoopAadhaarVerifier ---

// NoopAadhaarVerifier is a no-op implementation of AadhaarProvider for
// development and testing.
type NoopAadhaarVerifier struct{}

// NewNoopAadhaarVerifier creates a new no-op Aadhaar verifier.
func NewNoopAadhaarVerifier() *NoopAadhaarVerifier {
	return &NoopAadhaarVerifier{}
}

// GenerateOTP simulates OTP generation without making an API call.
func (n *NoopAadhaarVerifier) GenerateOTP(_ context.Context, aadhaarNumber string) error {
	log.Debug().Str("aadhaar", aadhaarNumber).Msg("noop aadhaar generate OTP")
	return nil
}

// VerifyOTP returns a mock successful verification result.
func (n *NoopAadhaarVerifier) VerifyOTP(_ context.Context, aadhaarNumber, otp string) (*VerificationResult, error) {
	log.Debug().Str("aadhaar", aadhaarNumber).Str("otp", otp).Msg("noop aadhaar verify OTP")
	return &VerificationResult{
		Valid:   true,
		Name:    "Noop User",
		DOB:     "01/01/1990",
		Gender:  "M",
		Address: "123 Test Street, Mumbai, Maharashtra 400001",
	}, nil
}
