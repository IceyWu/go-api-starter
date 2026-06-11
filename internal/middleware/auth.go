package middleware

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go-api-starter/internal/model"
	"go-api-starter/pkg/response"
)

// TokenBlacklistChecker defines the interface for checking token blacklist
type TokenBlacklistChecker interface {
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// UserRepository defines the interface for user data operations needed by auth middleware
type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

// userFreezeEntry holds cached freeze status for a user
type userFreezeEntry struct {
	freezed   bool
	expiresAt time.Time
}

type AuthMiddleware struct {
	jwtSecret        string
	blacklistChecker TokenBlacklistChecker
	userRepo         UserRepository

	// User freeze status cache (short TTL to reduce DB queries)
	freezeCacheMu  sync.RWMutex
	freezeCache    map[uint]*userFreezeEntry
	freezeCacheTTL time.Duration
}

// NewAuthMiddleware creates an auth middleware with all features
func NewAuthMiddleware(jwtSecret string, checker TokenBlacklistChecker, userRepo UserRepository) *AuthMiddleware {
	m := &AuthMiddleware{
		jwtSecret:        jwtSecret,
		blacklistChecker: checker,
		userRepo:         userRepo,
		freezeCache:      make(map[uint]*userFreezeEntry),
		freezeCacheTTL:   5 * time.Minute,
	}

	// Start background cleanup goroutine
	go m.cleanupFreezeCache()

	return m
}

// cleanupFreezeCache periodically removes expired entries from the freeze cache
func (m *AuthMiddleware) cleanupFreezeCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.freezeCacheMu.Lock()
		for uid, entry := range m.freezeCache {
			if now.After(entry.expiresAt) {
				delete(m.freezeCache, uid)
			}
		}
		m.freezeCacheMu.Unlock()
	}
}

// isUserFrozen checks if a user is frozen, using cache to reduce DB hits
func (m *AuthMiddleware) isUserFrozen(ctx context.Context, userID uint) (frozen bool, exists bool, err error) {
	// Check cache first
	m.freezeCacheMu.RLock()
	entry, ok := m.freezeCache[userID]
	m.freezeCacheMu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.freezed, true, nil
	}

	// Cache miss or expired, query DB
	user, err := m.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, false, err
	}

	// Update cache
	m.freezeCacheMu.Lock()
	m.freezeCache[userID] = &userFreezeEntry{
		freezed:   user.Freezed,
		expiresAt: time.Now().Add(m.freezeCacheTTL),
	}
	m.freezeCacheMu.Unlock()

	return user.Freezed, true, nil
}

// InvalidateUserFreezeCache removes a user from the freeze status cache.
// Call this when a user's freeze status changes.
func (m *AuthMiddleware) InvalidateUserFreezeCache(userID uint) {
	m.freezeCacheMu.Lock()
	delete(m.freezeCache, userID)
	m.freezeCacheMu.Unlock()
}

// RequireAuth validates JWT token and sets userID in context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Check if token is blacklisted
		if m.blacklistChecker != nil {
			blacklisted, err := m.blacklistChecker.IsTokenBlacklisted(c.Request.Context(), tokenString)
			if err == nil && blacklisted {
				response.Unauthorized(c, "认证令牌已失效")
				c.Abort()
				return
			}
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "认证令牌无效")
			c.Abort()
			return
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "认证令牌解析失败")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			response.Unauthorized(c, "用户ID无效")
			c.Abort()
			return
		}

		userIDUint := uint(userID)

		// Check if user is frozen (using cache to avoid hitting DB every request)
		if m.userRepo != nil {
			frozen, exists, err := m.isUserFrozen(c.Request.Context(), userIDUint)
			if err != nil || !exists {
				response.Unauthorized(c, "用户不存在")
				c.Abort()
				return
			}
			if frozen {
				response.Forbidden(c, "用户已被冻结")
				c.Abort()
				return
			}
		}

		// Set user ID and token in context
		c.Set("userID", userIDUint)
		c.Set("token", tokenString)
		c.Next()
	}
}

// OptionalAuth tries to parse JWT token and set userID if present, but does not block the request
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.Next()
			return
		}

		c.Set("userID", uint(userID))
		c.Set("token", tokenString)
		c.Next()
	}
}
