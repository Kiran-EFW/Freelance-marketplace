package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// ContextKeyUserID is the Fiber locals key that stores the authenticated user ID.
const ContextKeyUserID = "user_id"

// NewJWTAuth returns middleware that validates the Authorization header and
// stores the authenticated user's ID in Fiber locals.
func NewJWTAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]
		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			log.Warn().Err(err).Msg("invalid JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "malformed token subject",
			})
		}

		c.Locals(ContextKeyUserID, userID)
		return c.Next()
	}
}

// GetUserID extracts the authenticated user ID from Fiber locals.
// Returns uuid.Nil if not present.
func GetUserID(c *fiber.Ctx) uuid.UUID {
	id, ok := c.Locals(ContextKeyUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
