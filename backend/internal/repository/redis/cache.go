package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// CacheStore provides a typed wrapper around a Redis client for caching operations.
type CacheStore struct {
	client *redis.Client
}

// NewCacheStore creates a CacheStore backed by the given Redis client.
func NewCacheStore(client *redis.Client) *CacheStore {
	return &CacheStore{client: client}
}

// Set stores a string value with the given key and TTL.
func (c *CacheStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("redis: failed to set value")
	}
	return err
}

// Get retrieves a string value by key. Returns redis.Nil if the key does not exist.
func (c *CacheStore) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", err
		}
		log.Error().Err(err).Str("key", key).Msg("redis: failed to get value")
		return "", err
	}
	return val, nil
}

// Delete removes a key from the cache.
func (c *CacheStore) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("redis: failed to delete key")
	}
	return err
}

// SetJSON marshals the value to JSON and stores it with the given key and TTL.
func (c *CacheStore) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("redis: failed to marshal JSON")
		return fmt.Errorf("redis: marshal JSON for key %s: %w", key, err)
	}
	return c.Set(ctx, key, string(data), ttl)
}

// GetJSON retrieves a value by key and unmarshals the JSON into dest.
// Returns redis.Nil if the key does not exist.
func (c *CacheStore) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		log.Error().Err(err).Str("key", key).Msg("redis: failed to unmarshal JSON")
		return fmt.Errorf("redis: unmarshal JSON for key %s: %w", key, err)
	}
	return nil
}

// IncrementRateLimit atomically increments a counter for rate limiting.
// It sets the expiry to the given window duration on first increment.
// Returns the current count after incrementing.
func (c *CacheStore) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("redis: failed to increment rate limit")
		return 0, err
	}
	return incr.Val(), nil
}

// sessionKey returns the Redis key for a user session.
func sessionKey(userID string) string {
	return fmt.Sprintf("session:%s", userID)
}

// SetSession stores session data for a user with the given TTL.
func (c *CacheStore) SetSession(ctx context.Context, userID string, sessionData SessionData, ttl time.Duration) error {
	return c.SetJSON(ctx, sessionKey(userID), sessionData, ttl)
}

// GetSession retrieves session data for a user.
// Returns redis.Nil if no session exists.
func (c *CacheStore) GetSession(ctx context.Context, userID string) (SessionData, error) {
	var session SessionData
	err := c.GetJSON(ctx, sessionKey(userID), &session)
	if err != nil {
		return session, err
	}
	return session, nil
}

// DeleteSession removes a user's session from Redis.
func (c *CacheStore) DeleteSession(ctx context.Context, userID string) error {
	return c.Delete(ctx, sessionKey(userID))
}
