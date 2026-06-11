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

// AddUserToken associates a token with a user for batch invalidation.
// Uses a timestamp-based approach: stores each token hash with its own key
// and maintains a counter for the user's invalidation generation.
func (b *RedisTokenBlacklist) AddUserToken(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	tokenHash := hashToken(token)

	// Store individual token entry with user association
	tokenUserKey := fmt.Sprintf("%s%d:%s", userTokensPrefix, userID, tokenHash)
	return b.cache.Set(ctx, tokenUserKey, []byte("1"), expiration)
}

// InvalidateUserTokens invalidates all tokens for a user by setting a
// "invalidate_before" timestamp. Any token issued before this time is invalid.
func (b *RedisTokenBlacklist) InvalidateUserTokens(ctx context.Context, userID uint) error {
	// Set user invalidation timestamp — all tokens before this are invalid
	invalidateKey := b.buildUserInvalidateKey(userID)
	nowBytes := []byte(fmt.Sprintf("%d", time.Now().Unix()))
	return b.cache.Set(ctx, invalidateKey, nowBytes, userTokensTTL)
}

// IsUserTokenInvalidated checks if user's tokens have been bulk-invalidated
// after the given issued-at time. Returns true if the token should be rejected.
func (b *RedisTokenBlacklist) IsUserTokenInvalidated(ctx context.Context, userID uint, issuedAt int64) (bool, error) {
	invalidateKey := b.buildUserInvalidateKey(userID)
	data, err := b.cache.Get(ctx, invalidateKey)
	if err == cache.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var invalidateBefore int64
	fmt.Sscanf(string(data), "%d", &invalidateBefore)
	return issuedAt <= invalidateBefore, nil
}

// buildUserInvalidateKey builds the cache key for user token invalidation timestamp
func (b *RedisTokenBlacklist) buildUserInvalidateKey(userID uint) string {
	return fmt.Sprintf("%sinvalidate:%d", userTokensPrefix, userID)
}
