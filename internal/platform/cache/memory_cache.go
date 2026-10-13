package cache

import (
	"bytes"
	"context"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

func (i *cacheItem) isExpired(now time.Time) bool {
	return !i.expiresAt.IsZero() && !now.Before(i.expiresAt)
}

// MemoryCache keeps values and counters in the same TTL-aware keyspace.
type MemoryCache struct {
	mu        sync.Mutex
	data      map[string]*cacheItem
	now       func() time.Time
	available atomic.Bool
	stopCh    chan struct{}
	closeOnce sync.Once
}

func NewMemoryCache() *MemoryCache {
	return newMemoryCache(time.Now)
}

func newMemoryCache(now func() time.Time) *MemoryCache {
	m := &MemoryCache{data: make(map[string]*cacheItem), now: now, stopCh: make(chan struct{})}
	m.available.Store(true)
	go m.cleanupLoop()
	return m
}

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

func (m *MemoryCache) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	for key, item := range m.data {
		if item.isExpired(now) {
			delete(m.data, key)
		}
	}
}

func (m *MemoryCache) itemLocked(key string) (*cacheItem, bool) {
	item, ok := m.data[key]
	if ok && item.isExpired(m.now()) {
		delete(m.data, key)
		return nil, false
	}
	return item, ok
}

func (m *MemoryCache) newItem(value []byte, ttl time.Duration) *cacheItem {
	item := &cacheItem{value: bytes.Clone(value)}
	if ttl > 0 {
		item.expiresAt = m.now().Add(ttl)
	}
	return item
}

func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.itemLocked(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return bytes.Clone(item.value), nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = m.newItem(value, ttl)
	return nil
}

func (m *MemoryCache) SetIfAbsent(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.itemLocked(key); ok {
		return false, nil
	}
	m.data[key] = m.newItem(value, ttl)
	return true, nil
}

func (m *MemoryCache) CompareAndDelete(ctx context.Context, key string, expected []byte) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.itemLocked(key)
	if !ok {
		return false, ErrKeyNotFound
	}
	if !bytes.Equal(item.value, expected) {
		return false, nil
	}
	delete(m.data, key)
	return true, nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MemoryCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			delete(m.data, key)
		}
	}
	return nil
}

func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.itemLocked(key)
	return ok, nil
}

func (m *MemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return m.IncrWithExpire(ctx, key, 0)
}

func (m *MemoryCache) IncrWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	item, exists := m.itemLocked(key)
	var value int64
	if exists {
		var err error
		value, err = strconv.ParseInt(string(item.value), 10, 64)
		if err != nil {
			return 0, err
		}
	} else {
		item = m.newItem(nil, ttl)
	}
	if value == math.MaxInt64 {
		return 0, strconv.ErrRange
	}
	value++
	item.value = []byte(strconv.FormatInt(value, 10))
	m.data[key] = item
	return value, nil
}

func (m *MemoryCache) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !m.available.Load() {
		return ErrCacheUnavailable
	}
	return nil
}

func (m *MemoryCache) IsAvailable() bool { return m.available.Load() }

func (m *MemoryCache) Close() error {
	m.closeOnce.Do(func() { m.available.Store(false); close(m.stopCh) })
	return nil
}
