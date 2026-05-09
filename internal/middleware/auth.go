package middleware

import (
	"context"
	"strings"

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

type AuthMiddleware struct {
	jwtSecret        string
	blacklistChecker TokenBlacklistChecker
	userRepo         UserRepository
}

// NewAuthMiddleware creates an auth middleware with all features
func NewAuthMiddleware(jwtSecret string, checker TokenBlacklistChecker, userRepo UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:        jwtSecret,
		blacklistChecker: checker,
		userRepo:         userRepo,
	}
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

		// Check if user is frozen (if userRepo is available)
		if m.userRepo != nil {
			user, err := m.userRepo.FindByID(c.Request.Context(), userIDUint)
			if err != nil {
				response.Unauthorized(c, "用户不存在")
				c.Abort()
				return
			}

			if user.Freezed {
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
