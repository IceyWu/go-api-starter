package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisCacheAtomicAndExpiryContract(t *testing.T) {
	server := miniredis.RunT(t)
	backend := &RedisCache{client: redis.NewClient(&redis.Options{Addr: server.Addr()})}
	t.Cleanup(func() { _ = backend.Close() })
	runAtomicCacheContract(t, backend, "redis:")
	runCacheExpiryContract(t, backend, server.FastForward)
}
