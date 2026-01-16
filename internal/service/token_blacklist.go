package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go-api-starter/pkg/cache"
)

const (
	tokenBlacklistPrefix = "blacklist:token:"
	userTokensPrefix     = "blacklist:user:"
	userTokensTTL        = 7 * 24 * time.Hour // 7 days
)

// TokenBlacklist defines the interface for token blacklist operations
type TokenBlacklist interface {
	// Add adds a token to the blacklist with the given expiration
	Add(ctx context.Context, token string, expiration time.Duration) error

	// IsBlacklisted checks if a token is in the blacklist
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// InvalidateUserTokens invalidates all tokens for a user
	InvalidateUserTokens(ctx context.Context, userID uint) error

	// AddUserToken associates a token with a user for batch invalidation
	AddUserToken(ctx context.Context, userID uint, token string, expiration time.Duration) error
}

// RedisTokenBlacklist implements TokenBlacklist using Redis
type RedisTokenBlacklist struct {
	cache cache.CacheBackend
}

// NewRedisTokenBlacklist creates a new Redis-backed token blacklist
func NewRedisTokenBlacklist(cacheBackend cache.CacheBackend) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{
		cache: cacheBackend,
	}
}

// hashToken creates a SHA256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// buildTokenKey builds a cache key for a token
func (b *RedisTokenBlacklist) buildTokenKey(tokenHash string) string {
	return tokenBlacklistPrefix + tokenHash
}

// buildUserTokensKey builds a cache key for user's tokens
func (b *RedisTokenBlacklist) buildUserTokensKey(userID uint) string {
	return fmt.Sprintf("%s%d", userTokensPrefix, userID)
}


// Add adds a token to the blacklist with the given expiration
func (b *RedisTokenBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	tokenHash := hashToken(token)
	key := b.buildTokenKey(tokenHash)
	return b.cache.Set(ctx, key, []byte("1"), expiration)
}

// IsBlacklisted checks if a token is in the blacklist
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := hashToken(token)
	key := b.buildTokenKey(tokenHash)
	return b.cache.Exists(ctx, key)
}

// AddUserToken associates a token with a user for batch invalidation
func (b *RedisTokenBlacklist) AddUserToken(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	tokenHash := hashToken(token)

	// Store the token hash in user's token set
	userKey := b.buildUserTokensKey(userID)
	data, err := b.cache.Get(ctx, userKey)
	if err != nil && err != cache.ErrKeyNotFound {
		return err
	}

	// Append token hash to existing list (simple approach using comma-separated values)
	var tokens string
	if len(data) > 0 {
		tokens = string(data) + "," + tokenHash
	} else {
		tokens = tokenHash
	}

	return b.cache.Set(ctx, userKey, []byte(tokens), userTokensTTL)
}

// InvalidateUserTokens invalidates all tokens for a user
func (b *RedisTokenBlacklist) InvalidateUserTokens(ctx context.Context, userID uint) error {
	userKey := b.buildUserTokensKey(userID)
	data, err := b.cache.Get(ctx, userKey)
	if err == cache.ErrKeyNotFound {
		return nil // No tokens to invalidate
	}
	if err != nil {
		return err
	}

	// Parse token hashes and blacklist each one
	tokens := string(data)
	if tokens == "" {
		return nil
	}

	// Split by comma and blacklist each token
	for _, tokenHash := range splitTokens(tokens) {
		if tokenHash == "" {
			continue
		}
		key := b.buildTokenKey(tokenHash)
		// Use a long TTL since we don't know the original expiration
		if err := b.cache.Set(ctx, key, []byte("1"), userTokensTTL); err != nil {
			return err
		}
	}

	// Clear the user's token list
	return b.cache.Delete(ctx, userKey)
}

// splitTokens splits a comma-separated string of tokens
func splitTokens(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// MemoryTokenBlacklist implements TokenBlacklist using memory (for fallback)
type MemoryTokenBlacklist struct {
	cache cache.CacheBackend
}

// NewMemoryTokenBlacklist creates a new memory-backed token blacklist
func NewMemoryTokenBlacklist(cacheBackend cache.CacheBackend) *MemoryTokenBlacklist {
	return &MemoryTokenBlacklist{
		cache: cacheBackend,
	}
}

// Add adds a token to the blacklist
func (b *MemoryTokenBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	tokenHash := hashToken(token)
	key := tokenBlacklistPrefix + tokenHash
	return b.cache.Set(ctx, key, []byte("1"), expiration)
}

// IsBlacklisted checks if a token is in the blacklist
func (b *MemoryTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := hashToken(token)
	key := tokenBlacklistPrefix + tokenHash
	return b.cache.Exists(ctx, key)
}

// InvalidateUserTokens is a no-op for memory blacklist (tokens expire naturally)
func (b *MemoryTokenBlacklist) InvalidateUserTokens(ctx context.Context, userID uint) error {
	return nil
}

// AddUserToken is a no-op for memory blacklist
func (b *MemoryTokenBlacklist) AddUserToken(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	return nil
}
