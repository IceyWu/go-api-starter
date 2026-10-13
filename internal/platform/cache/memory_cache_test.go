package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryCacheConcurrentIncrements(t *testing.T) {
	backend := NewMemoryCache()
	t.Cleanup(func() { _ = backend.Close() })
	const workers, increments = 32, 100
	start := make(chan struct{})
	var group sync.WaitGroup
	for range workers {
		group.Go(func() {
			<-start
			for range increments {
				if _, err := backend.Incr(context.Background(), "shared"); err != nil {
					t.Error(err)
				}
			}
		})
	}
	close(start)
	group.Wait()
	got, err := backend.Incr(context.Background(), "shared")
	if err != nil || got != workers*increments+1 {
		t.Fatalf("lost concurrent increments: got %d, error %v", got, err)
	}
}

func TestMemoryCacheAtomicAndExpiryContract(t *testing.T) {
	var now atomic.Int64
	now.Store(time.Now().UnixNano())
	backend := newMemoryCache(func() time.Time { return time.Unix(0, now.Load()) })
	t.Cleanup(func() { _ = backend.Close() })
	t.Run("atomic", func(t *testing.T) { runAtomicCacheContract(t, backend, "memory:") })
	runCacheExpiryContract(t, backend, func(delta time.Duration) { now.Add(int64(delta)) })
	now.Add(int64(time.Minute))
	backend.cleanup()
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if len(backend.data) != 0 {
		t.Fatalf("expired entries were retained: %d", len(backend.data))
	}
}
