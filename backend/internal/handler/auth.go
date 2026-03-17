package handler

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/config"
	"github.com/seva-platform/backend/internal/repository/postgres"
	rediscache "github.com/seva-platform/backend/internal/repository/redis"
)

const (
	otpLength       = 6
	otpTTL          = 5 * time.Minute
	otpRateWindow   = 10 * time.Minute
	otpRateMax      = 3
	otpMaxAttempts  = 3
	refreshTokenMul = 7 // refresh token lives 7x the access token expiry
)

// phoneRegex matches Indian phone numbers:
//   - +91 followed by 10 digits
//   - or bare 10-digit number starting with 6-9
var phoneRegex = regexp.MustCompile(`^(?:\+91)?([6-9]\d{9})$`)

// AuthHandler manages OTP-based authentication flows.
type AuthHandler struct {
	cfg     *config.Config
	db      *postgres.Queries
	cache   *rediscache.CacheStore
	smsSvc  sms.SMSProvider
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	cfg *config.Config,
	db *postgres.Queries,
	cache *rediscache.CacheStore,
	smsSvc sms.SMSProvider,
) *AuthHandler {
	return &AuthHandler{
		cfg:   cfg,
		db:    db,
		cache: cache,
		smsSvc: smsSvc,
	}
}

// RegisterRoutes mounts auth routes on the given Fiber router group.
func (h *AuthHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/send-otp", h.SendOTP)
	rg.Post("/verify-otp", h.VerifyOTP)
	rg.Post("/refresh", h.RefreshToken)
}

// ---------------------------------------------------------------------------
// Request / Response types
// ---------------------------------------------------------------------------

// sendOTPRequest is the payload for POST /auth/send-otp.
type sendOTPRequest struct {
	Phone string `json:"phone" validate:"required"`
}

// verifyOTPRequest is the payload for POST /auth/verify-otp.
type verifyOTPRequest struct {
	Phone string `json:"phone" validate:"required"`
	OTP   string `json:"otp"   validate:"required,len=6"`
}

// refreshTokenRequest is the payload for POST /auth/refresh.
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ---------------------------------------------------------------------------
// SendOTP
// ---------------------------------------------------------------------------

// SendOTP generates a one-time password and dispatches it via the SMS adapter.
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req sendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "phone number is required",
		})
	}

	// 1. Validate & normalise the phone number to +91XXXXXXXXXX.
	phone, err := validatePhone(req.Phone)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := c.Context()

	// 2. Rate-limit: max 3 OTP requests per phone per 10-minute window.
	rateLimitKey := fmt.Sprintf("otp_rate:%s", phone)
	count, err := h.cache.IncrementRateLimit(ctx, rateLimitKey, otpRateWindow)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to check OTP rate limit")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process request",
		})
	}
	if count > otpRateMax {
		log.Warn().Str("phone", phone).Int64("count", count).Msg("OTP rate limit exceeded")
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "too many OTP requests, please try again later",
		})
	}

	// 3. Generate a cryptographically secure 6-digit OTP.
	code, err := generateOTP()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate OTP",
		})
	}

	// 4. Store OTP in the database (with 5-minute expiry set by the SQL query).
	_, err = h.db.CreateOTP(ctx, postgres.CreateOTPParams{
		Phone: phone,
		Code:  code,
	})
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to store OTP in database")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate OTP",
		})
	}

	// 5. Also cache in Redis for fast lookup during verification.
	otpCacheKey := fmt.Sprintf("otp:%s", phone)
	if err := h.cache.Set(ctx, otpCacheKey, code, otpTTL); err != nil {
		// Non-fatal: we can fall back to database lookup during verification.
		log.Warn().Err(err).Str("phone", phone).Msg("failed to cache OTP in Redis")
	}

	// 6. Send OTP via SMS.
	smsMessage := fmt.Sprintf("Your Seva verification code is: %s. Valid for 5 minutes.", code)
	log.Info().Str("phone", phone).Msg("sending OTP")

	if err := h.smsSvc.SendSMS(phone, smsMessage); err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to send OTP via SMS")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to send OTP",
		})
	}

	// 7. Return success without leaking the OTP.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent successfully",
	})
}

// ---------------------------------------------------------------------------
// VerifyOTP
// ---------------------------------------------------------------------------

// VerifyOTP validates the OTP and returns a JWT pair on success.
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req verifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Phone == "" || req.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "phone and otp are required",
		})
	}

	if len(req.OTP) != otpLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP must be 6 digits",
		})
	}

	// Validate & normalise the phone number.
	phone, err := validatePhone(req.Phone)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := c.Context()
	log.Info().Str("phone", phone).Msg("verifying OTP")

	// Look up the latest valid (non-expired, non-verified, under max attempts) OTP from the DB.
	// We pass the user-supplied code in the query so the DB can match it directly.
	otpRecord, err := h.db.GetValidOTP(ctx, postgres.GetValidOTPParams{
		Phone: phone,
		Code:  req.OTP,
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			// Before returning, we still want to increment attempts on the most recent
			// OTP for this phone to prevent brute force. We do a broader lookup.
			h.incrementLatestOTPAttempt(ctx, phone)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired OTP",
			})
		}
		log.Error().Err(err).Str("phone", phone).Msg("failed to look up OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "verification failed",
		})
	}

	// Increment attempts for this OTP record.
	if err := h.db.IncrementOTPAttempts(ctx, otpRecord.ID); err != nil {
		log.Error().Err(err).Str("otp_id", otpRecord.ID.String()).Msg("failed to increment OTP attempts")
		// Non-fatal: continue with verification.
	}

	// Constant-time comparison of the OTP codes.
	if subtle.ConstantTimeCompare([]byte(req.OTP), []byte(otpRecord.Code)) != 1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired OTP",
		})
	}

	// Mark OTP as verified.
	if err := h.db.MarkOTPVerified(ctx, otpRecord.ID); err != nil {
		log.Error().Err(err).Str("otp_id", otpRecord.ID.String()).Msg("failed to mark OTP verified")
		// Non-fatal: continue.
	}

	// Delete OTP from Redis now that it has been used.
	otpCacheKey := fmt.Sprintf("otp:%s", phone)
	if err := h.cache.Delete(ctx, otpCacheKey); err != nil {
		log.Warn().Err(err).Str("phone", phone).Msg("failed to delete OTP from Redis")
	}

	// Upsert user: look up by phone; create if first login.
	user, err := h.getOrCreateUser(ctx, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to upsert user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create account",
		})
	}

	// Generate JWT token pair.
	accessToken, err := h.generateToken(user.ID.String(), string(user.Type), h.cfg.JWTExpiry, "access")
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	refreshExpiry := h.cfg.JWTExpiry * refreshTokenMul
	refreshToken, err := h.generateToken(user.ID.String(), string(user.Type), refreshExpiry, "refresh")
	if err != nil {
		log.Error().Err(err).Msg("failed to generate refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.cfg.JWTExpiry.Seconds()),
		"user": fiber.Map{
			"id":    user.ID.String(),
			"phone": user.Phone,
			"type":  string(user.Type),
		},
	})
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

// RefreshToken issues a new access token given a valid refresh token.
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req refreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh_token is required",
		})
	}

	log.Info().Msg("refreshing access token")

	// Parse and validate the refresh token.
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired refresh token",
		})
	}

	// Verify this is actually a refresh token.
	if claims.TokenType != "refresh" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token type",
		})
	}

	ctx := c.Context()

	// Check if the refresh token has been revoked (blacklisted in Redis).
	blacklistKey := fmt.Sprintf("token_blacklist:%s", claims.ID)
	_, err = h.cache.Get(ctx, blacklistKey)
	if err == nil {
		// Key exists in Redis => token has been revoked.
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "token has been revoked",
		})
	}
	// If err is redis.Nil, the token is NOT blacklisted, so continue.
	// If err is something else, log it but allow the refresh to proceed.

	userID := claims.Subject
	role := claims.Role

	// Issue a new access token.
	accessToken, err := h.generateToken(userID, role, h.cfg.JWTExpiry, "access")
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.cfg.JWTExpiry.Seconds()),
	})
}

// ---------------------------------------------------------------------------
// Custom JWT Claims
// ---------------------------------------------------------------------------

// CustomClaims extends the standard JWT claims with application-specific fields.
type CustomClaims struct {
	jwt.RegisteredClaims
	Role      string `json:"role,omitempty"`
	TokenType string `json:"token_type,omitempty"`
}

// ---------------------------------------------------------------------------
// Helper: generate JWT
// ---------------------------------------------------------------------------

// generateToken creates a signed JWT with standard and custom claims.
func (h *AuthHandler) generateToken(userID, role string, expiry time.Duration, tokenType string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    "seva",
			ID:        generateTokenID(),
		},
		Role:      role,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

// generateTokenID produces a short random hex string for the JWT "jti" claim.
func generateTokenID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// ---------------------------------------------------------------------------
// Helper: generate OTP
// ---------------------------------------------------------------------------

// generateOTP produces a cryptographically secure 6-digit numeric code.
func generateOTP() (string, error) {
	// We generate a random number in [0, 999999] and zero-pad to 6 digits.
	max := big.NewInt(1000000) // exclusive upper bound
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// ---------------------------------------------------------------------------
// Helper: validate & normalise phone
// ---------------------------------------------------------------------------

// validatePhone normalises an Indian phone number to the +91XXXXXXXXXX format.
// It accepts:
//   - +91XXXXXXXXXX  (already normalised)
//   - 91XXXXXXXXXX   (missing '+')
//   - 0XXXXXXXXXX    (local with leading zero)
//   - XXXXXXXXXX     (bare 10 digits, first digit 6-9)
func validatePhone(phone string) (string, error) {
	// Strip spaces, dashes, and parentheses.
	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(phone)

	// Remove leading zero if present (common in Indian local dialling).
	if strings.HasPrefix(cleaned, "0") && len(cleaned) == 11 {
		cleaned = cleaned[1:]
	}

	// Remove leading "91" without "+" if present.
	if strings.HasPrefix(cleaned, "91") && len(cleaned) == 12 {
		cleaned = "+" + cleaned
	}

	// Try to match.
	matches := phoneRegex.FindStringSubmatch(cleaned)
	if matches == nil {
		return "", fmt.Errorf("invalid Indian phone number: must be +91XXXXXXXXXX or 10 digits starting with 6-9")
	}

	// matches[1] is the 10-digit capture group.
	return "+91" + matches[1], nil
}

// ---------------------------------------------------------------------------
// Helper: user upsert
// ---------------------------------------------------------------------------

// getOrCreateUser looks up a user by phone. If none exists, it creates a new
// customer account and returns it.
func (h *AuthHandler) getOrCreateUser(ctx context.Context, phone string) (postgres.User, error) {
	user, err := h.db.GetUserByPhone(ctx, phone)
	if err == nil {
		// Existing user found.
		return user, nil
	}

	if err != pgx.ErrNoRows {
		// Unexpected database error.
		return postgres.User{}, fmt.Errorf("get user by phone: %w", err)
	}

	// First-time login: create a new customer account.
	newUser, err := h.db.CreateUser(ctx, postgres.CreateUserParams{
		Type:              postgres.UserTypeCustomer,
		Phone:             phone,
		Email:             pgtype.Text{Valid: false},
		Name:              pgtype.Text{Valid: false},
		JurisdictionID:    "IN", // default jurisdiction for Indian phone numbers
		PreferredLanguage: "en",
		DeviceType:        postgres.DeviceTypeSmartphone,
	})
	if err != nil {
		return postgres.User{}, fmt.Errorf("create user: %w", err)
	}

	log.Info().Str("user_id", newUser.ID.String()).Str("phone", phone).Msg("created new user on first OTP login")
	return newUser, nil
}

// ---------------------------------------------------------------------------
// Helper: increment OTP attempts (best-effort for brute-force protection)
// ---------------------------------------------------------------------------

// incrementLatestOTPAttempt is a best-effort attempt to increment the attempt
// counter on the most recent OTP for a phone number even when the supplied code
// didn't match. This prevents unlimited brute-force guesses.
func (h *AuthHandler) incrementLatestOTPAttempt(ctx context.Context, phone string) {
	otp, err := h.db.GetLatestOTPByPhone(ctx, phone)
	if err != nil {
		log.Debug().Str("phone", phone).Msg("no active OTP to increment attempts for")
		return
	}
	if err := h.db.IncrementOTPAttempts(ctx, otp.ID); err != nil {
		log.Error().Err(err).Str("otp_id", otp.ID.String()).Msg("failed to increment OTP attempts on mismatch")
	}
	log.Debug().Str("phone", phone).Int32("attempts", otp.Attempts+1).Msg("incremented OTP attempts after failed verification")
}
