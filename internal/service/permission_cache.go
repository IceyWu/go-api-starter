package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
)

// CacheStats holds cache statistics
type CacheStats struct {
	Hits   int64
	Misses int64
}

// PermissionCache manages user permission caching with TTL support
type PermissionCache struct {
	cacheRepo repository.UserPermissionCacheRepositoryInterface
	ttl       time.Duration
	mu        sync.RWMutex
	stats     CacheStats
}

// NewPermissionCache creates a new PermissionCache
func NewPermissionCache(cacheRepo repository.UserPermissionCacheRepositoryInterface, ttl time.Duration) *PermissionCache {
	return &PermissionCache{
		cacheRepo: cacheRepo,
		ttl:       ttl,
	}
}

// Get retrieves cached permission value for a user and space
// Returns nil if cache miss or expired
func (c *PermissionCache) Get(ctx context.Context, userID, spaceID uint) (*uint64, error) {
	cache, err := c.cacheRepo.FindByUserAndSpace(ctx, userID, spaceID)
	if err != nil {
		return nil, err
	}

	if cache == nil {
		atomic.AddInt64(&c.stats.Misses, 1)
		return nil, nil
	}

	// Check TTL expiration
	if c.ttl > 0 && cache.ExpiresAt.Before(time.Now()) {
		atomic.AddInt64(&c.stats.Misses, 1)
		return nil, nil
	}

	atomic.AddInt64(&c.stats.Hits, 1)
	return &cache.Value, nil
}

// Set stores permission values for a user with TTL
func (c *PermissionCache) Set(ctx context.Context, userID uint, permissions map[uint]uint64) error {
	now := time.Now()
	expiresAt := now.Add(c.ttl)

	for spaceID, value := range permissions {
		cache := &model.UserPermissionCache{
			UserID:    userID,
			SpaceID:   spaceID,
			Value:     value,
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := c.cacheRepo.Upsert(ctx, cache); err != nil {
			return err
		}
	}
	return nil
}

// InvalidateUser removes all cached permissions for a user
func (c *PermissionCache) InvalidateUser(ctx context.Context, userID uint) error {
	return c.cacheRepo.DeleteByUserID(ctx, userID)
}

// InvalidateByRole removes cached permissions for all users with a specific role
func (c *PermissionCache) InvalidateByRole(ctx context.Context, roleID uint, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	return c.cacheRepo.DeleteByUserIDs(ctx, userIDs)
}

// GetStats returns cache statistics
func (c *PermissionCache) GetStats() CacheStats {
	return CacheStats{
		Hits:   atomic.LoadInt64(&c.stats.Hits),
		Misses: atomic.LoadInt64(&c.stats.Misses),
	}
}

// GetTTL returns the cache TTL duration
func (c *PermissionCache) GetTTL() time.Duration {
	return c.ttl
}

// GetAllForUser retrieves all cached permission values for a user
func (c *PermissionCache) GetAllForUser(ctx context.Context, userID uint) (map[uint]uint64, error) {
	return c.cacheRepo.GetUserSpaceValues(ctx, userID)
}
