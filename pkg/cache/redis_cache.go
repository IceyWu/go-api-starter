package cache

import (
	"context"
	"sync/atomic"
	"time"

	"go-api-starter/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements CacheBackend using Redis
type RedisCache struct {
	client      redis.UniversalClient
	config      *config.RedisConfig
	isAvailable atomic.Bool
}

// NewRedisCache creates a new Redis cache backend
func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	var client redis.UniversalClient

	if cfg.ClusterMode && len(cfg.ClusterAddrs) > 0 {
		// Cluster mode
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.ClusterAddrs,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	} else {
		// Single node mode
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addr(),
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	}

	rc := &RedisCache{
		client: client,
		config: cfg,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		rc.isAvailable.Store(false)
		return rc, err
	}

	rc.isAvailable.Store(true)
	return rc, nil
}

// Get retrieves a value by key
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Set stores a value with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}


// DeleteByPrefix removes all keys matching the prefix using SCAN
func (r *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	var keys []string

	for {
		var err error
		var batch []string
		batch, cursor, err = r.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

// Exists checks if a key exists
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Incr increments a counter and returns the new value
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrWithExpire increments a counter with TTL and returns the new value
func (r *RedisCache) IncrWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// Ping checks if the cache is available
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	r.isAvailable.Store(err == nil)
	return err
}

// IsAvailable returns whether the cache backend is currently available
func (r *RedisCache) IsAvailable() bool {
	return r.isAvailable.Load()
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Client returns the underlying Redis client for advanced operations
func (r *RedisCache) Client() redis.UniversalClient {
	return r.client
}
