package cache

import (
	"context"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	modePrimary  int32 = 0
	modeFallback int32 = 1
)

// FallbackCache implements CacheBackend with automatic fallback
type FallbackCache struct {
	primary       CacheBackend
	fallback      CacheBackend
	logger        *zap.Logger
	mode          atomic.Int32
	checkInterval time.Duration
	stopCh        chan struct{}
}

// NewFallbackCache creates a new fallback cache
func NewFallbackCache(primary, fallback CacheBackend, logger *zap.Logger) *FallbackCache {
	if logger == nil {
		logger = zap.NewNop()
	}

	fc := &FallbackCache{
		primary:       primary,
		fallback:      fallback,
		logger:        logger,
		checkInterval: 10 * time.Second,
		stopCh:        make(chan struct{}),
	}

	// Check initial availability
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := primary.Ping(ctx); err != nil {
		fc.mode.Store(modeFallback)
		logger.Warn("Primary cache unavailable, starting in fallback mode", zap.Error(err))
	} else {
		fc.mode.Store(modePrimary)
	}

	// Start health check goroutine
	go fc.healthCheckLoop()

	return fc
}

// healthCheckLoop periodically checks primary cache availability
func (f *FallbackCache) healthCheckLoop() {
	ticker := time.NewTicker(f.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.checkAndRecover()
		case <-f.stopCh:
			return
		}
	}
}


// checkAndRecover checks if primary is available and recovers if possible
func (f *FallbackCache) checkAndRecover() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := f.primary.Ping(ctx)
	currentMode := f.mode.Load()

	if err != nil && currentMode == modePrimary {
		// Primary failed, switch to fallback
		f.mode.Store(modeFallback)
		f.logger.Warn("Primary cache became unavailable, switching to fallback", zap.Error(err))
	} else if err == nil && currentMode == modeFallback {
		// Primary recovered, switch back
		f.mode.Store(modePrimary)
		f.logger.Info("Primary cache recovered, switching back from fallback")
	}
}

// current returns the current active cache backend
func (f *FallbackCache) current() CacheBackend {
	if f.mode.Load() == modeFallback {
		return f.fallback
	}
	return f.primary
}

// Get retrieves a value by key
func (f *FallbackCache) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := f.current().Get(ctx, key)
	if err != nil && f.mode.Load() == modePrimary {
		// Try fallback on primary failure
		f.mode.Store(modeFallback)
		f.logger.Warn("Primary cache failed, falling back", zap.Error(err))
		return f.fallback.Get(ctx, key)
	}
	return result, err
}

// Set stores a value with TTL
func (f *FallbackCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := f.current().Set(ctx, key, value, ttl)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		f.logger.Warn("Primary cache failed on set, falling back", zap.Error(err))
		return f.fallback.Set(ctx, key, value, ttl)
	}
	return err
}

// Delete removes a key
func (f *FallbackCache) Delete(ctx context.Context, key string) error {
	err := f.current().Delete(ctx, key)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		return f.fallback.Delete(ctx, key)
	}
	return err
}

// DeleteByPrefix removes all keys matching the prefix
func (f *FallbackCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	err := f.current().DeleteByPrefix(ctx, prefix)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		return f.fallback.DeleteByPrefix(ctx, prefix)
	}
	return err
}

// Exists checks if a key exists
func (f *FallbackCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := f.current().Exists(ctx, key)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		return f.fallback.Exists(ctx, key)
	}
	return result, err
}

// Incr increments a counter and returns the new value
func (f *FallbackCache) Incr(ctx context.Context, key string) (int64, error) {
	result, err := f.current().Incr(ctx, key)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		return f.fallback.Incr(ctx, key)
	}
	return result, err
}

// IncrWithExpire increments a counter with TTL and returns the new value
func (f *FallbackCache) IncrWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	result, err := f.current().IncrWithExpire(ctx, key, ttl)
	if err != nil && f.mode.Load() == modePrimary {
		f.mode.Store(modeFallback)
		return f.fallback.IncrWithExpire(ctx, key, ttl)
	}
	return result, err
}

// Ping checks if the cache is available
func (f *FallbackCache) Ping(ctx context.Context) error {
	return f.current().Ping(ctx)
}

// IsAvailable returns whether the cache backend is currently available
func (f *FallbackCache) IsAvailable() bool {
	return f.current().IsAvailable()
}

// IsDegraded returns true if running in fallback mode
func (f *FallbackCache) IsDegraded() bool {
	return f.mode.Load() == modeFallback
}

// Close closes both cache backends
func (f *FallbackCache) Close() error {
	close(f.stopCh)
	f.primary.Close()
	f.fallback.Close()
	return nil
}
