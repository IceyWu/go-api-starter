package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func runAtomicCacheContract(t *testing.T, backend CacheBackend, prefix string) {
	t.Helper()
	ctx := context.Background()
	t.Cleanup(func() { _ = backend.DeleteByPrefix(ctx, prefix) })
	var reserved atomic.Int64
	var group sync.WaitGroup
	for range 32 {
		group.Go(func() {
			ok, err := backend.SetIfAbsent(ctx, prefix+"reservation", []byte("owner"), time.Minute)
			if err != nil {
				t.Error(err)
			}
			if ok {
				reserved.Add(1)
			}
		})
	}
	group.Wait()
	if reserved.Load() != 1 {
		t.Fatalf("expected one reservation, got %d", reserved.Load())
	}
	if ok, err := backend.CompareAndDelete(ctx, prefix+"reservation", []byte("wrong")); err != nil || ok {
		t.Fatalf("mismatched value consumed: %v, %v", ok, err)
	}
	var consumed atomic.Int64
	for range 32 {
		group.Go(func() {
			ok, err := backend.CompareAndDelete(ctx, prefix+"reservation", []byte("owner"))
			if err != nil && !errors.Is(err, ErrKeyNotFound) {
				t.Error(err)
			}
			if ok {
				consumed.Add(1)
			}
		})
	}
	group.Wait()
	if consumed.Load() != 1 {
		t.Fatalf("expected one consumption, got %d", consumed.Load())
	}
	for range 32 {
		group.Go(func() {
			if _, err := backend.IncrWithExpire(ctx, prefix+"counter", time.Minute); err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()
	value, err := backend.Get(ctx, prefix+"counter")
	if err != nil || string(value) != "32" {
		t.Fatalf("lost increments: %q, %v", value, err)
	}
	if err := backend.Delete(ctx, prefix+"counter"); err != nil {
		t.Fatal(err)
	}
	if value, err := backend.Incr(ctx, prefix+"counter"); err != nil || value != 1 {
		t.Fatalf("counter survived deletion: %d, %v", value, err)
	}
}

func runCacheExpiryContract(t *testing.T, backend CacheBackend, advance func(time.Duration)) {
	t.Helper()
	ctx := context.Background()
	if value, err := backend.IncrWithExpire(ctx, "ttl-counter", time.Second); err != nil || value != 1 {
		t.Fatalf("%d: %v", value, err)
	}
	advance(500 * time.Millisecond)
	if value, err := backend.IncrWithExpire(ctx, "ttl-counter", time.Second); err != nil || value != 2 {
		t.Fatalf("%d: %v", value, err)
	}
	advance(500 * time.Millisecond)
	if _, err := backend.Get(ctx, "ttl-counter"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("counter TTL was extended: %v", err)
	}
	if value, err := backend.IncrWithExpire(ctx, "ttl-counter", time.Second); err != nil || value != 1 {
		t.Fatalf("expired counter did not reset: %d, %v", value, err)
	}
	if ok, err := backend.SetIfAbsent(ctx, "ttl-reservation", []byte("old"), time.Second); err != nil || !ok {
		t.Fatalf("%v: %v", ok, err)
	}
	advance(time.Second)
	if ok, err := backend.SetIfAbsent(ctx, "ttl-reservation", []byte("new"), time.Second); err != nil || !ok {
		t.Fatalf("expired reservation blocked reuse: %v, %v", ok, err)
	}
	if ok, err := backend.CompareAndDelete(ctx, "ttl-reservation", []byte("old")); err != nil || ok {
		t.Fatalf("old owner removed replacement: %v, %v", ok, err)
	}
}
