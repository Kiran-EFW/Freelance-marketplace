package domain

import "errors"

// Sentinel errors used across the domain and service layers.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidState   = errors.New("invalid state transition")
	ErrConflict       = errors.New("conflict")
	ErrRateLimited    = errors.New("rate limited")
	ErrInternalServer = errors.New("internal server error")
)

// CacheStore abstracts a key-value cache (e.g., Redis).
type CacheStore interface {
	Get(key string) (string, error)
	Set(key, value string, ttlSeconds int) error
	Delete(key string) error
}
