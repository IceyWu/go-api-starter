package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"go-api-starter/pkg/cache"
)

const (
	permCacheKeyPrefix = "perm:user:"
	permCacheAllSuffix = ":all"
)

// RedisPermissionCache implements permission caching using Redis
type RedisPermissionCache struct {
	cache     cache.CacheBackend
	ttl       time.Duration
	keyPrefix string
	stats     struct {
		hits   int64
		misses int64
	}
}

// NewRedisPermissionCache creates a new Redis-backed permission cache
func NewRedisPermissionCache(cacheBackend cache.CacheBackend, ttl time.Duration) *RedisPermissionCache {
	return &RedisPermissionCache{
		cache:     cacheBackend,
		ttl:       ttl,
		keyPrefix: permCacheKeyPrefix,
	}
}

// buildKey builds a cache key for a specific user and space
func (c *RedisPermissionCache) buildKey(userID, spaceID uint) string {
	return fmt.Sprintf("%s%d:space:%d", c.keyPrefix, userID, spaceID)
}

// buildAllKey builds a cache key for all permissions of a user
func (c *RedisPermissionCache) buildAllKey(userID uint) string {
	return fmt.Sprintf("%s%d%s", c.keyPrefix, userID, permCacheAllSuffix)
}

// buildUserPrefix builds a key prefix for all keys of a user
func (c *RedisPermissionCache) buildUserPrefix(userID uint) string {
	return fmt.Sprintf("%s%d:", c.keyPrefix, userID)
}


// Get retrieves cached permission value for a user and space
func (c *RedisPermissionCache) Get(ctx context.Context, userID, spaceID uint) (*uint64, error) {
	key := c.buildKey(userID, spaceID)
	data, err := c.cache.Get(ctx, key)
	if err == cache.ErrKeyNotFound {
		atomic.AddInt64(&c.stats.misses, 1)
		return nil, nil
	}
	if err != nil {
		atomic.AddInt64(&c.stats.misses, 1)
		return nil, err
	}

	value, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		atomic.AddInt64(&c.stats.misses, 1)
		return nil, err
	}

	atomic.AddInt64(&c.stats.hits, 1)
	return &value, nil
}

// Set stores permission values for a user with TTL
func (c *RedisPermissionCache) Set(ctx context.Context, userID uint, permissions map[uint]uint64) error {
	// Store individual space permissions
	for spaceID, value := range permissions {
		key := c.buildKey(userID, spaceID)
		data := []byte(strconv.FormatUint(value, 10))
		if err := c.cache.Set(ctx, key, data, c.ttl); err != nil {
			return err
		}
	}

	// Also store all permissions as JSON for bulk retrieval
	allKey := c.buildAllKey(userID)
	allData, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, allKey, allData, c.ttl)
}

// InvalidateUser removes all cached permissions for a user
func (c *RedisPermissionCache) InvalidateUser(ctx context.Context, userID uint) error {
	prefix := c.buildUserPrefix(userID)
	return c.cache.DeleteByPrefix(ctx, prefix)
}

// InvalidateByRole removes cached permissions for all users with a specific role
func (c *RedisPermissionCache) InvalidateByRole(ctx context.Context, roleID uint, userIDs []uint) error {
	for _, userID := range userIDs {
		if err := c.InvalidateUser(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

// GetStats returns cache statistics
func (c *RedisPermissionCache) GetStats() CacheStats {
	return CacheStats{
		Hits:   atomic.LoadInt64(&c.stats.hits),
		Misses: atomic.LoadInt64(&c.stats.misses),
	}
}

// GetTTL returns the cache TTL duration
func (c *RedisPermissionCache) GetTTL() time.Duration {
	return c.ttl
}

// GetAllForUser retrieves all cached permission values for a user
func (c *RedisPermissionCache) GetAllForUser(ctx context.Context, userID uint) (map[uint]uint64, error) {
	allKey := c.buildAllKey(userID)
	data, err := c.cache.Get(ctx, allKey)
	if err == cache.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var permissions map[uint]uint64
	if err := json.Unmarshal(data, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}
