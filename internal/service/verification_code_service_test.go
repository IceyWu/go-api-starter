package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-api-starter/internal/platform/cache"
	"go-api-starter/internal/platform/mail"
)

type failingVerificationCache struct {
	cache.CacheBackend
	deleteErr error
	setErr    error
}

func TestGenerateVerificationCodeUsesSixDigits(t *testing.T) {
	service := NewVerificationCodeService(nil, nil, "test")
	for range 100 {
		code, err := service.GenerateCode()
		if err != nil || len(code) != VerificationCodeLength || strings.Trim(code, "0123456789") != "" {
			t.Fatalf("invalid code format or generation error: %v", err)
		}
	}
}

func TestConcurrentVerificationCodeSendAndConsume(t *testing.T) {
	backend := cache.NewMemoryCache()
	t.Cleanup(func() { _ = backend.Close() })
	runConcurrentVerificationCodeContract(t, backend, "concurrent@example.com")
}

func runConcurrentVerificationCodeContract(t *testing.T, backend cache.CacheBackend, email string) {
	t.Helper()
	t.Cleanup(func() {
		_ = backend.Delete(context.Background(), VerificationCodePrefix+"login:"+email)
		_ = backend.Delete(context.Background(), VerificationCodeRateLimit+email)
	})
	service := NewVerificationCodeService(backend, mail.NewClient(&mail.Config{MockSend: true}), "test")
	ctx := context.Background()
	var sent atomic.Int64
	var group sync.WaitGroup
	for range 32 {
		group.Go(func() {
			if err := service.SendVerificationCode(ctx, email, "login"); err == nil {
				sent.Add(1)
			}
		})
	}
	group.Wait()
	if sent.Load() != 1 {
		t.Fatalf("expected one send, got %d", sent.Load())
	}
	code, err := backend.Get(ctx, VerificationCodePrefix+"login:"+email)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := service.VerifyCode(ctx, email, "login", "wrong"); err == nil || valid {
		t.Fatal("wrong code accepted")
	}
	var consumed atomic.Int64
	for range 32 {
		group.Go(func() {
			if valid, err := service.VerifyCode(ctx, email, "login", string(code)); err == nil && valid {
				consumed.Add(1)
			}
		})
	}
	group.Wait()
	if consumed.Load() != 1 {
		t.Fatalf("code reused: %d successes", consumed.Load())
	}
}

func TestFailedCodeStorageReleasesSendReservation(t *testing.T) {
	backend := cache.NewMemoryCache()
	t.Cleanup(func() { _ = backend.Close() })
	const email = "retry@example.com"
	service := NewVerificationCodeService(failingVerificationCache{CacheBackend: backend, setErr: errors.New("storage failed")}, nil, "test")
	if err := service.SendVerificationCode(context.Background(), email, "login"); err == nil {
		t.Fatal("storage failure was ignored")
	}
	if ok, err := backend.Exists(context.Background(), VerificationCodeRateLimit+email); err != nil || ok {
		t.Fatal("failed send kept its reservation")
	}
	retry := NewVerificationCodeService(backend, mail.NewClient(&mail.Config{MockSend: true}), "test")
	if err := retry.SendVerificationCode(context.Background(), email, "login"); err != nil {
		t.Fatalf("retry blocked: %v", err)
	}
}

func (c failingVerificationCache) CompareAndDelete(ctx context.Context, key string, expected []byte) (bool, error) {
	if c.deleteErr != nil {
		return false, c.deleteErr
	}
	return c.CacheBackend.CompareAndDelete(ctx, key, expected)
}

func (c failingVerificationCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if c.setErr != nil {
		return c.setErr
	}
	return c.CacheBackend.Set(ctx, key, value, ttl)
}

func TestVerifyCodeDoesNotSucceedWhenCacheWritesFail(t *testing.T) {
	for _, operation := range []string{"delete", "set"} {
		t.Run(operation, func(t *testing.T) {
			backend := cache.NewMemoryCache()
			t.Cleanup(func() { _ = backend.Close() })
			const email = "user@example.com"
			if err := backend.Set(context.Background(), VerificationCodePrefix+"register:"+email, []byte("123456"), time.Minute); err != nil {
				t.Fatal(err)
			}
			failing := failingVerificationCache{CacheBackend: backend}
			if operation == "delete" {
				failing.deleteErr = errors.New("cache delete failed")
			} else {
				failing.setErr = errors.New("cache set failed")
			}
			service := NewVerificationCodeService(failing, nil, "test")
			verified, err := service.VerifyCode(context.Background(), email, "register", "123456")
			if err == nil || verified || service.IsEmailVerified(context.Background(), email) {
				t.Fatalf("verification succeeded despite %s failure: %v", operation, err)
			}
		})
	}
}
