//go:build integration

package cache

import (
	"context"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"go-api-starter/internal/config"
)

func TestRealRedisAtomicCacheContract(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		if os.Getenv("REQUIRE_INTEGRATION") == "true" {
			t.Fatal("REDIS_TEST_ADDR is required")
		}
		t.Skip("REDIS_TEST_ADDR is not configured")
	}
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	backend, err := NewRedisCache(&config.RedisConfig{Host: host, Port: port})
	if backend != nil {
		t.Cleanup(func() { _ = backend.Close() })
	}
	if err != nil {
		t.Fatal(err)
	}
	prefix := "go-api-starter-integration:" + rand.Text() + ":"
	runAtomicCacheContract(t, backend, prefix)
	ctx := context.Background()
	key := prefix + "expiry"
	if _, err := backend.IncrWithExpire(ctx, key, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	before, err := backend.Client().PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	if _, err := backend.IncrWithExpire(ctx, key, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	after, err := backend.Client().PTTL(ctx, key).Result()
	if err != nil || after <= 0 || after >= before {
		t.Fatalf("increment extended TTL: before %v, after %v, error %v", before, after, err)
	}
	time.Sleep(2 * time.Second)
	if _, err := backend.Get(ctx, key); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("real Redis counter did not expire: %v", err)
	}
	if value, err := backend.IncrWithExpire(ctx, key, time.Minute); err != nil || value != 1 {
		t.Fatalf("expired counter did not reset: %d, %v", value, err)
	}
}
