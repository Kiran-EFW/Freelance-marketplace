package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

// NewRequestID returns middleware that generates a UUID v4 for each request,
// stores it in c.Locals("requestid"), and sets the X-Request-ID response header.
func NewRequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.New().String()
		},
		ContextKey: "requestid",
	})
}
