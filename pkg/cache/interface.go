package cache

import (
	"context"
	"errors"
	"time"
)

// Common cache errors
var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrCacheUnavailable = errors.New("cache unavailable")
)

// CacheBackend defines the interface for cache operations
type CacheBackend interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key
	Delete(ctx context.Context, key string) error

	// DeleteByPrefix removes all keys matching the prefix
	DeleteByPrefix(ctx context.Context, prefix string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Incr increments a counter and returns the new value
	Incr(ctx context.Context, key string) (int64, error)

	// IncrWithExpire increments a counter with TTL and returns the new value
	IncrWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error)

	// Ping checks if the cache is available
	Ping(ctx context.Context) error

	// IsAvailable returns whether the cache backend is currently available
	IsAvailable() bool

	// Close closes the cache connection
	Close() error
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Sets       int64 `json:"sets"`
	Deletes    int64 `json:"deletes"`
	Errors     int64 `json:"errors"`
	HitRate    float64 `json:"hit_rate"`
}

// CalculateHitRate calculates the hit rate
func (s *CacheStats) CalculateHitRate() {
	total := s.Hits + s.Misses
	if total > 0 {
		s.HitRate = float64(s.Hits) / float64(total)
	}
}
