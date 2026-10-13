package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

type unavailableAtomicCache struct{ CacheBackend }

func (b unavailableAtomicCache) SetIfAbsent(context.Context, string, []byte, time.Duration) (bool, error) {
	return false, ErrCacheUnavailable
}
func (b unavailableAtomicCache) CompareAndDelete(context.Context, string, []byte) (bool, error) {
	return false, ErrCacheUnavailable
}

func TestFallbackCacheDoesNotRetryAtomicSecurityOperations(t *testing.T) {
	primary := NewMemoryCache()
	secondary := NewMemoryCache()
	backend := NewFallbackCache(unavailableAtomicCache{primary}, secondary, nil)
	t.Cleanup(func() { _ = backend.Close() })
	backend.mode.Store(modeFallback)
	if _, err := backend.SetIfAbsent(context.Background(), "code", []byte("value"), time.Minute); !errors.Is(err, ErrCacheUnavailable) {
		t.Fatal("reservation did not fail closed")
	}
	if _, err := backend.CompareAndDelete(context.Background(), "code", []byte("value")); !errors.Is(err, ErrCacheUnavailable) {
		t.Fatal("consumption did not fail closed")
	}
	if ok, _ := secondary.Exists(context.Background(), "code"); ok {
		t.Fatal("security operation wrote fallback state")
	}
}

func TestCacheMissDoesNotActivateFallback(t *testing.T) {
	backend := NewFallbackCache(NewMemoryCache(), NewMemoryCache(), nil)
	t.Cleanup(func() { _ = backend.Close() })
	if _, err := backend.Get(context.Background(), "missing"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatal(err)
	}
	if backend.IsDegraded() {
		t.Fatal("ordinary cache miss activated fallback")
	}
}
