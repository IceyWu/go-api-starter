package cache

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// isExpired checks if the item has expired
func (i *cacheItem) isExpired() bool {
	if i.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.expiresAt)
}

// MemoryCache implements CacheBackend using in-memory storage
type MemoryCache struct {
	data      sync.Map
	counters  sync.Map
	available atomic.Bool
	stopCh    chan struct{}
}

// NewMemoryCache creates a new memory cache backend
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		stopCh: make(chan struct{}),
	}
	mc.available.Store(true)

	// Start cleanup goroutine
	go mc.cleanupLoop()

	return mc
}

// cleanupLoop periodically removes expired items
func (m *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired items
func (m *MemoryCache) cleanup() {
	m.data.Range(func(key, value interface{}) bool {
		if item, ok := value.(*cacheItem); ok && item.isExpired() {
			m.data.Delete(key)
		}
		return true
	})
}

// Get retrieves a value by key
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, ok := m.data.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}

	item, ok := value.(*cacheItem)
	if !ok || item.isExpired() {
		m.data.Delete(key)
		return nil, ErrKeyNotFound
	}

	return item.value, nil
}


// Set stores a value with TTL
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	item := &cacheItem{
		value: value,
	}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	m.data.Store(key, item)
	return nil
}

// Delete removes a key
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.data.Delete(key)
	return nil
}

// DeleteByPrefix removes all keys matching the prefix
func (m *MemoryCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	m.data.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok && strings.HasPrefix(k, prefix) {
			m.data.Delete(key)
		}
		return true
	})
	return nil
}

// Exists checks if a key exists
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	value, ok := m.data.Load(key)
	if !ok {
		return false, nil
	}

	item, ok := value.(*cacheItem)
	if !ok || item.isExpired() {
		m.data.Delete(key)
		return false, nil
	}

	return true, nil
}

// Incr increments a counter and returns the new value
func (m *MemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	for {
		value, loaded := m.counters.LoadOrStore(key, &atomic.Int64{})
		counter := value.(*atomic.Int64)
		if loaded {
			return counter.Add(1), nil
		}
		counter.Store(1)
		return 1, nil
	}
}

// IncrWithExpire increments a counter with TTL and returns the new value
func (m *MemoryCache) IncrWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	result, err := m.Incr(ctx, key)
	if err != nil {
		return 0, err
	}

	// Store expiration info
	item := &cacheItem{
		value:     nil,
		expiresAt: time.Now().Add(ttl),
	}
	m.data.Store(key+"_exp", item)

	return result, nil
}

// Ping checks if the cache is available
func (m *MemoryCache) Ping(ctx context.Context) error {
	if !m.available.Load() {
		return ErrCacheUnavailable
	}
	return nil
}

// IsAvailable returns whether the cache backend is currently available
func (m *MemoryCache) IsAvailable() bool {
	return m.available.Load()
}

// Close stops the cleanup goroutine
func (m *MemoryCache) Close() error {
	close(m.stopCh)
	m.available.Store(false)
	return nil
}
