package service

import (
	"context"
	"testing"
	"time"

	"go-api-starter/internal/platform/cache"
)

func TestTokenInvalidationRejectsMalformedCacheData(t *testing.T) {
	backend := cache.NewMemoryCache()
	t.Cleanup(func() { _ = backend.Close() })
	blacklist := NewRedisTokenBlacklist(backend)
	if err := backend.Set(context.Background(), blacklist.buildUserInvalidateKey(1), []byte("invalid"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := blacklist.IsUserTokenInvalidated(context.Background(), 1, time.Now().Unix()); err == nil {
		t.Fatal("expected malformed token invalidation data to produce an error")
	}
}
