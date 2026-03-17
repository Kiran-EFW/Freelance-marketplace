package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/adapter/sms"
	"github.com/seva-platform/backend/internal/config"
	"github.com/seva-platform/backend/internal/domain"
)

// AuthHandler manages OTP-based authentication flows.
type AuthHandler struct {
	cfg      *config.Config
	users    domain.UserRepository
	smsSvc   sms.SMSProvider
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(cfg *config.Config, users domain.UserRepository, smsSvc sms.SMSProvider) *AuthHandler {
	return &AuthHandler{
		cfg:    cfg,
		users:  users,
		smsSvc: smsSvc,
	}
}

// RegisterRoutes mounts auth routes on the given Fiber router group.
func (h *AuthHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/send-otp", h.SendOTP)
	rg.Post("/verify-otp", h.VerifyOTP)
	rg.Post("/refresh", h.RefreshToken)
}

// sendOTPRequest is the payload for POST /auth/send-otp.
type sendOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

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

	// TODO: generate a cryptographically random OTP
	// TODO: store OTP in Redis with a TTL (e.g., 5 minutes)
	// TODO: rate-limit OTP requests per phone number

	log.Info().Str("phone", req.Phone).Msg("sending OTP")

	if err := h.smsSvc.SendSMS(req.Phone, "Your verification code is: 000000"); err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("failed to send OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to send OTP",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent successfully",
	})
}

// verifyOTPRequest is the payload for POST /auth/verify-otp.
type verifyOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

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

	// TODO: retrieve stored OTP from Redis and compare
	// TODO: delete OTP from Redis after successful verification
	// TODO: create user if first login (upsert by phone)

	log.Info().Str("phone", req.Phone).Msg("verifying OTP")

	// Placeholder: generate JWT tokens
	accessToken, err := h.generateToken(uuid.New(), req.Phone, h.cfg.JWTExpiry)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	refreshToken, err := h.generateToken(uuid.New(), req.Phone, h.cfg.JWTExpiry*7)
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
	})
}

// refreshTokenRequest is the payload for POST /auth/refresh.
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

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

	// TODO: validate the refresh token signature and expiry
	// TODO: check that the refresh token has not been revoked
	// TODO: issue a new access token

	log.Info().Msg("refreshing access token")

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired refresh token",
		})
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "malformed token subject",
		})
	}

	accessToken, err := h.generateToken(userID, "", h.cfg.JWTExpiry)
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

// generateToken creates a signed JWT with standard claims.
func (h *AuthHandler) generateToken(userID uuid.UUID, phone string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		Issuer:    "seva",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
