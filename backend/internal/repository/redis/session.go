package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// SessionData holds the authenticated session state for a user.
type SessionData struct {
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	JurisdictionID string `json:"jurisdiction_id"`
	ExpiresAt      int64  `json:"expires_at"` // Unix timestamp
}

// SessionStore provides session management operations on top of Redis.
type SessionStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewSessionStore creates a SessionStore with default settings.
// The default TTL is 24 hours and the key prefix is "session:".
func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{
		client: client,
		prefix: "session:",
		ttl:    24 * time.Hour,
	}
}

// sessionStoreKey returns the fully qualified Redis key for the given user.
func (s *SessionStore) sessionStoreKey(userID string) string {
	return fmt.Sprintf("%s%s", s.prefix, userID)
}

// Create persists a new session for the user, overwriting any existing session.
func (s *SessionStore) Create(ctx context.Context, data SessionData) error {
	if data.ExpiresAt == 0 {
		data.ExpiresAt = time.Now().Add(s.ttl).Unix()
	}

	key := s.sessionStoreKey(data.UserID)
	fields := map[string]any{
		"user_id":         data.UserID,
		"role":            data.Role,
		"jurisdiction_id": data.JurisdictionID,
		"expires_at":      data.ExpiresAt,
	}

	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.ExpireAt(ctx, key, time.Unix(data.ExpiresAt, 0))
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("user_id", data.UserID).Msg("session: failed to create")
		return fmt.Errorf("session: create for user %s: %w", data.UserID, err)
	}
	return nil
}

// Get retrieves the session for the given user.
// Returns an error wrapping redis.Nil if no session exists.
func (s *SessionStore) Get(ctx context.Context, userID string) (SessionData, error) {
	key := s.sessionStoreKey(userID)
	result, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("session: failed to get")
		return SessionData{}, fmt.Errorf("session: get for user %s: %w", userID, err)
	}
	if len(result) == 0 {
		return SessionData{}, redis.Nil
	}

	var expiresAt int64
	if v, ok := result["expires_at"]; ok {
		fmt.Sscanf(v, "%d", &expiresAt)
	}

	session := SessionData{
		UserID:         result["user_id"],
		Role:           result["role"],
		JurisdictionID: result["jurisdiction_id"],
		ExpiresAt:      expiresAt,
	}

	// Check if session has expired (belt-and-suspenders; Redis TTL is the primary guard).
	if session.ExpiresAt > 0 && time.Now().Unix() > session.ExpiresAt {
		_ = s.Delete(ctx, userID)
		return SessionData{}, redis.Nil
	}

	return session, nil
}

// Delete removes the session for the given user.
func (s *SessionStore) Delete(ctx context.Context, userID string) error {
	key := s.sessionStoreKey(userID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("session: failed to delete")
		return fmt.Errorf("session: delete for user %s: %w", userID, err)
	}
	return nil
}

// Refresh extends the session expiry by the default TTL.
// This is typically called on each authenticated request to implement sliding sessions.
func (s *SessionStore) Refresh(ctx context.Context, userID string) error {
	key := s.sessionStoreKey(userID)

	// Check existence first.
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("session: failed to check existence for refresh")
		return fmt.Errorf("session: refresh check for user %s: %w", userID, err)
	}
	if exists == 0 {
		return redis.Nil
	}

	newExpiry := time.Now().Add(s.ttl)
	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, "expires_at", newExpiry.Unix())
	pipe.ExpireAt(ctx, key, newExpiry)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("session: failed to refresh")
		return fmt.Errorf("session: refresh for user %s: %w", userID, err)
	}
	return nil
}
