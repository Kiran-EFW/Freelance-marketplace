package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
)

// Default rate limit settings.
const (
	defaultWindowDuration = 1 * time.Minute
	defaultMaxRequests    = 60
	authMaxRequests       = 10 // stricter limit for auth endpoints
	authWindowDuration    = 5 * time.Minute
)

// RateLimiter provides Redis-backed rate limiting using a sliding window
// algorithm.
type RateLimiter struct {
	redis         *redis.Client
	maxRequests   int
	windowSeconds int
}

// NewRateLimiter creates a new rate limiter backed by Redis.
func NewRateLimiter(redisClient *redis.Client, cfg *config.Config) *RateLimiter {
	maxReq := cfg.RateLimitMax()
	if maxReq <= 0 {
		maxReq = defaultMaxRequests
	}

	return &RateLimiter{
		redis:         redisClient,
		maxRequests:   maxReq,
		windowSeconds: int(defaultWindowDuration.Seconds()),
	}
}

// Middleware returns a Fiber handler that enforces per-IP rate limiting using
// a Redis-backed sliding window counter.
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		path := c.Path()

		// Determine rate limits. Auth endpoints get stricter limits.
		maxReq := rl.maxRequests
		window := rl.windowSeconds

		if isAuthEndpoint(path) {
			maxReq = authMaxRequests
			window = int(authWindowDuration.Seconds())
		}

		key := fmt.Sprintf("ratelimit:%s:%s", ip, rateLimitBucket(path))

		allowed, remaining, resetAt, err := rl.checkLimit(c.Context(), key, maxReq, window)
		if err != nil {
			// On Redis errors, allow the request through rather than blocking
			// legitimate traffic.
			log.Error().Err(err).Str("ip", ip).Msg("rate limiter redis error, allowing request")
			return c.Next()
		}

		// Set rate limit headers.
		c.Set("X-RateLimit-Limit", strconv.Itoa(maxReq))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Set("Retry-After", strconv.FormatInt(retryAfter, 10))

			log.Warn().
				Str("ip", ip).
				Str("path", path).
				Int("limit", maxReq).
				Msg("rate limit exceeded")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
		}

		return c.Next()
	}
}

// checkLimit uses a Redis sliding window to determine if the request is allowed.
// It returns whether the request is allowed, the number of remaining requests,
// and the Unix timestamp when the window resets.
func (rl *RateLimiter) checkLimit(ctx context.Context, key string, maxReq, windowSeconds int) (allowed bool, remaining int, resetAt int64, err error) {
	now := time.Now()
	windowStart := now.Add(-time.Duration(windowSeconds) * time.Second)
	resetAt = now.Add(time.Duration(windowSeconds) * time.Second).Unix()

	pipe := rl.redis.Pipeline()

	// Remove entries outside the current window.
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixMicro(), 10))

	// Count current entries in the window.
	countCmd := pipe.ZCard(ctx, key)

	// Add the current request timestamp.
	nowMicro := now.UnixMicro()
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(nowMicro),
		Member: nowMicro,
	})

	// Set expiry on the key to auto-cleanup.
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, fmt.Errorf("rate limit pipeline exec: %w", err)
	}

	count := int(countCmd.Val())

	if count >= maxReq {
		return false, 0, resetAt, nil
	}

	remaining = maxReq - count - 1 // -1 for the request we just added
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetAt, nil
}

// isAuthEndpoint returns true if the path is an authentication-related endpoint
// that should have stricter rate limits.
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/otp",
		"/api/v1/auth/verify",
		"/api/v1/auth/refresh",
	}

	for _, p := range authPaths {
		if path == p {
			return true
		}
	}

	return false
}

// rateLimitBucket returns a bucket identifier for rate limiting. Auth endpoints
// share one bucket; all other endpoints share another.
func rateLimitBucket(path string) string {
	if isAuthEndpoint(path) {
		return "auth"
	}
	return "general"
}
