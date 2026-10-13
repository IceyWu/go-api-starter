//go:build integration

package service

import (
	"crypto/rand"
	"net"
	"os"
	"strconv"
	"testing"

	"go-api-starter/internal/config"
	"go-api-starter/internal/platform/cache"
)

func TestVerificationCodeWithRealRedis(t *testing.T) {
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
	backend, err := cache.NewRedisCache(&config.RedisConfig{Host: host, Port: port})
	if backend != nil {
		t.Cleanup(func() { _ = backend.Close() })
	}
	if err != nil {
		t.Fatal(err)
	}
	runConcurrentVerificationCodeContract(t, backend, "go-api-starter-test-"+rand.Text()+"@example.com")
}
