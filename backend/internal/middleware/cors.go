package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/seva-platform/backend/internal/config"
)

// NewCORS returns a CORS middleware configured for the given environment.
func NewCORS(cfg *config.Config) fiber.Handler {
	allowOrigins := "*"
	if cfg.IsProd() {
		// In production, restrict to known front-end origins.
		allowOrigins = "https://app.seva.io,https://admin.seva.io"
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           3600,
	})
}
