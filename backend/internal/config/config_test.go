package config

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// TestIsProd
// ---------------------------------------------------------------------------

func TestIsProd(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{
			name:        "prod environment",
			environment: "prod",
			want:        true,
		},
		{
			name:        "dev environment",
			environment: "dev",
			want:        false,
		},
		{
			name:        "staging environment",
			environment: "staging",
			want:        false,
		},
		{
			name:        "empty string",
			environment: "",
			want:        false,
		},
		{
			name:        "production (not 'prod')",
			environment: "production",
			want:        false,
		},
		{
			name:        "PROD uppercase",
			environment: "PROD",
			want:        false, // case-sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			if got := cfg.IsProd(); got != tt.want {
				t.Errorf("IsProd() = %v, want %v (environment=%q)", got, tt.want, tt.environment)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRateLimitMax
// ---------------------------------------------------------------------------

func TestRateLimitMax(t *testing.T) {
	// Clear any env override for RATE_LIMIT_MAX.
	os.Unsetenv("RATE_LIMIT_MAX")

	tests := []struct {
		name        string
		environment string
		envOverride string
		want        int
	}{
		{
			name:        "dev environment defaults to 100",
			environment: "dev",
			want:        100,
		},
		{
			name:        "prod environment defaults to 30",
			environment: "prod",
			want:        30,
		},
		{
			name:        "staging defaults to 100 (non-prod)",
			environment: "staging",
			want:        100,
		},
		{
			name:        "env override takes priority in dev",
			environment: "dev",
			envOverride: "50",
			want:        50,
		},
		{
			name:        "env override takes priority in prod",
			environment: "prod",
			envOverride: "10",
			want:        10,
		},
		{
			name:        "invalid env override falls back to default",
			environment: "dev",
			envOverride: "not-a-number",
			want:        100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or clear the env override.
			if tt.envOverride != "" {
				os.Setenv("RATE_LIMIT_MAX", tt.envOverride)
				defer os.Unsetenv("RATE_LIMIT_MAX")
			} else {
				os.Unsetenv("RATE_LIMIT_MAX")
			}

			cfg := &Config{Environment: tt.environment}
			if got := cfg.RateLimitMax(); got != tt.want {
				t.Errorf("RateLimitMax() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetEnv
// ---------------------------------------------------------------------------

func TestGetEnv(t *testing.T) {
	// Set a test variable.
	os.Setenv("SEVA_TEST_VAR", "hello")
	defer os.Unsetenv("SEVA_TEST_VAR")

	// Variable that exists should be returned.
	val := getEnv("SEVA_TEST_VAR", "default")
	if val != "hello" {
		t.Errorf("getEnv(SEVA_TEST_VAR) = %q, want %q", val, "hello")
	}

	// Variable that does not exist should return fallback.
	val = getEnv("SEVA_NONEXISTENT_VAR", "fallback")
	if val != "fallback" {
		t.Errorf("getEnv(SEVA_NONEXISTENT_VAR) = %q, want %q", val, "fallback")
	}

	// Empty env variable should return fallback.
	os.Setenv("SEVA_EMPTY_VAR", "")
	defer os.Unsetenv("SEVA_EMPTY_VAR")
	val = getEnv("SEVA_EMPTY_VAR", "fallback")
	if val != "fallback" {
		t.Errorf("getEnv(SEVA_EMPTY_VAR) = %q, want %q (empty env should use fallback)", val, "fallback")
	}
}

// ---------------------------------------------------------------------------
// TestGetFCMServiceAccountKey
// ---------------------------------------------------------------------------

func TestGetFCMServiceAccountKeyInline(t *testing.T) {
	cfg := &Config{
		FCMServiceAccountKey: `{"type":"service_account","project_id":"test"}`,
	}

	key := cfg.GetFCMServiceAccountKey()
	if key == nil {
		t.Fatal("GetFCMServiceAccountKey() returned nil, expected inline key")
	}
	if string(key) != `{"type":"service_account","project_id":"test"}` {
		t.Errorf("GetFCMServiceAccountKey() = %q, want inline JSON", string(key))
	}
}

func TestGetFCMServiceAccountKeyEmpty(t *testing.T) {
	cfg := &Config{}

	key := cfg.GetFCMServiceAccountKey()
	if key != nil {
		t.Errorf("GetFCMServiceAccountKey() should return nil when nothing is configured, got %q", string(key))
	}
}

func TestGetFCMServiceAccountKeyBadPath(t *testing.T) {
	cfg := &Config{
		FCMServiceAccountKeyPath: "/nonexistent/path/to/key.json",
	}

	key := cfg.GetFCMServiceAccountKey()
	if key != nil {
		t.Errorf("GetFCMServiceAccountKey() should return nil for bad path, got %q", string(key))
	}
}

// ---------------------------------------------------------------------------
// TestConfigDefaults
// ---------------------------------------------------------------------------

func TestConfigDefaults(t *testing.T) {
	// Test that a Config with defaults has sensible values.
	cfg := &Config{
		ServerPort:  "8080",
		Environment: "dev",
	}

	if cfg.IsProd() {
		t.Error("dev environment should not be prod")
	}

	// Rate limit should be 100 in dev.
	os.Unsetenv("RATE_LIMIT_MAX")
	if got := cfg.RateLimitMax(); got != 100 {
		t.Errorf("RateLimitMax() in dev = %d, want 100", got)
	}
}
