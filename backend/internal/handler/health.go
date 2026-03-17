package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const appVersion = "0.1.0"

// HealthHandler exposes liveness and readiness endpoints.
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: rdb}
}

// RegisterRoutes mounts health routes on the given Fiber router group.
func (h *HealthHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/", h.Liveness)
	rg.Get("/ready", h.Readiness)
}

// Liveness returns a simple OK response to indicate the process is running.
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"version":   appVersion,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Readiness verifies that downstream dependencies (database, Redis) are reachable.
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Check PostgreSQL
	if err := h.db.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unavailable",
			"database": "unreachable",
			"error":    err.Error(),
		})
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unavailable",
			"redis":  "unreachable",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "ready",
		"database":  "connected",
		"redis":     "connected",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
