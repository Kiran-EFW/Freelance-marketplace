package logger

import (
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup initialises the global zerolog logger.
// In development, output is a pretty-printed console writer;
// in production, output is structured JSON on stdout.
func Setup(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	var writer io.Writer
	if env == "dev" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		writer = os.Stdout
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Logger = zerolog.New(writer).
		With().
		Timestamp().
		Caller().
		Str("service", "seva-api").
		Logger()
}

// FiberLogger returns Fiber middleware that logs every request via zerolog.
func FiberLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		elapsed := time.Since(start)
		status := c.Response().StatusCode()

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", elapsed).
			Str("ip", c.IP()).
			Str("request_id", c.Locals("requestid").(string)).
			Msg("request")

		return err
	}
}

// WithRequestID returns a sub-logger that includes the request ID from the
// Fiber context, useful for correlating log entries to a single request.
func WithRequestID(c *fiber.Ctx) zerolog.Logger {
	reqID, _ := c.Locals("requestid").(string)
	return log.With().Str("request_id", reqID).Logger()
}
