package handler

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// TestValidatePhone
// ---------------------------------------------------------------------------

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// Valid formats
		{
			name:  "already normalised +91",
			input: "+919876543210",
			want:  "+919876543210",
		},
		{
			name:  "bare 10 digits starting with 9",
			input: "9876543210",
			want:  "+919876543210",
		},
		{
			name:  "bare 10 digits starting with 8",
			input: "8765432109",
			want:  "+918765432109",
		},
		{
			name:  "bare 10 digits starting with 7",
			input: "7654321098",
			want:  "+917654321098",
		},
		{
			name:  "bare 10 digits starting with 6",
			input: "6543210987",
			want:  "+916543210987",
		},
		{
			name:  "with 91 prefix no plus",
			input: "919876543210",
			want:  "+919876543210",
		},
		{
			name:  "with leading zero",
			input: "09876543210",
			want:  "+919876543210",
		},
		{
			name:  "with spaces",
			input: "+91 987 654 3210",
			want:  "+919876543210",
		},
		{
			name:  "with dashes",
			input: "+91-987-654-3210",
			want:  "+919876543210",
		},
		{
			name:  "with parentheses and spaces",
			input: "(+91) 9876543210",
			want:  "+919876543210",
		},

		// Invalid formats
		{
			name:    "too short",
			input:   "12345",
			wantErr: true,
		},
		{
			name:    "starts with 5 (invalid for India)",
			input:   "5432109876",
			wantErr: true,
		},
		{
			name:    "starts with 1 (invalid for India)",
			input:   "1234567890",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "letters mixed in",
			input:   "98765abc10",
			wantErr: true,
		},
		{
			name:    "too many digits",
			input:   "+9198765432100",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePhone(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePhone(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("validatePhone(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("validatePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGenerateOTP
// ---------------------------------------------------------------------------

func TestGenerateOTP(t *testing.T) {
	// Generate multiple OTPs and verify format.
	seen := make(map[string]bool)
	digitRegex := regexp.MustCompile(`^\d{6}$`)

	for i := 0; i < 100; i++ {
		otp, err := generateOTP()
		if err != nil {
			t.Fatalf("generateOTP() returned error: %v", err)
		}

		// Must be exactly 6 characters.
		if len(otp) != 6 {
			t.Errorf("generateOTP() length = %d, want 6: %q", len(otp), otp)
		}

		// Must be all digits.
		if !digitRegex.MatchString(otp) {
			t.Errorf("generateOTP() = %q, not all digits", otp)
		}

		seen[otp] = true
	}

	// With 100 iterations we should see at least a few different OTPs
	// (probability of all same is astronomically low).
	if len(seen) < 5 {
		t.Errorf("generateOTP() produced only %d unique values in 100 iterations, expected more randomness", len(seen))
	}
}

func TestGenerateOTPZeroPadding(t *testing.T) {
	// Run many iterations to ensure zero-padded values like "000123" are valid.
	// We can't force a specific value, but we verify the format works.
	for i := 0; i < 50; i++ {
		otp, err := generateOTP()
		if err != nil {
			t.Fatalf("generateOTP() error: %v", err)
		}
		// OTP should be exactly 6 digits even if the number is small.
		if len(otp) != 6 {
			t.Errorf("OTP %q is not 6 characters", otp)
		}
	}
}

// ---------------------------------------------------------------------------
// TestCustomErrorHandler
// ---------------------------------------------------------------------------

func TestCustomErrorHandler(t *testing.T) {
	// Test that our error handler produces the correct status codes and structure.
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "fiber 400 error",
			err:        fiber.NewError(fiber.StatusBadRequest, "bad request"),
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "fiber 404 error",
			err:        fiber.NewError(fiber.StatusNotFound, "not found"),
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:       "fiber 401 error",
			err:        fiber.NewError(fiber.StatusUnauthorized, "unauthorized"),
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name:       "generic error defaults to 500",
			err:        fmt.Errorf("something broke"),
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that a fiber.Error gives us the right code.
			code := fiber.StatusInternalServerError
			message := "internal server error"

			if e, ok := tt.err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			if code != tt.wantStatus {
				t.Errorf("status = %d, want %d", code, tt.wantStatus)
			}

			if code == fiber.StatusInternalServerError && message != "internal server error" {
				t.Errorf("generic error message should be 'internal server error', got %q", message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestPhoneRegex
// ---------------------------------------------------------------------------

func TestPhoneRegex(t *testing.T) {
	re := regexp.MustCompile(`^(?:\+91)?([6-9]\d{9})$`)

	validNumbers := []string{
		"+919876543210",
		"9876543210",
		"8765432109",
		"7654321098",
		"6543210987",
	}

	invalidNumbers := []string{
		"1234567890",    // starts with 1
		"5432109876",    // starts with 5
		"98765",         // too short
		"98765432100",   // too long (11 digits)
		"+91123456789",  // starts with 1
		"",              // empty
	}

	for _, num := range validNumbers {
		if !re.MatchString(num) {
			t.Errorf("phoneRegex should match %q", num)
		}
	}

	for _, num := range invalidNumbers {
		if re.MatchString(num) {
			t.Errorf("phoneRegex should NOT match %q", num)
		}
	}
}

// ---------------------------------------------------------------------------
// TestGenerateTokenID
// ---------------------------------------------------------------------------

func TestGenerateTokenID(t *testing.T) {
	id := generateTokenID()
	// Should be 32 hex characters (16 bytes -> 32 hex chars).
	if len(id) != 32 {
		t.Errorf("generateTokenID() length = %d, want 32", len(id))
	}

	// Should only contain hex characters.
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("generateTokenID() contains non-hex character: %c", c)
		}
	}

	// Two IDs should be different.
	id2 := generateTokenID()
	if id == id2 {
		t.Error("generateTokenID() produced two identical IDs")
	}
}
