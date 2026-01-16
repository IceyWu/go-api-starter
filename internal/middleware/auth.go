package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"go-api-starter/pkg/response"
)

// TokenBlacklistChecker defines the interface for checking token blacklist
type TokenBlacklistChecker interface {
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

type AuthMiddleware struct {
	jwtSecret        string
	blacklistChecker TokenBlacklistChecker
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

// NewAuthMiddlewareWithBlacklist creates an auth middleware with blacklist support
func NewAuthMiddlewareWithBlacklist(jwtSecret string, checker TokenBlacklistChecker) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:        jwtSecret,
		blacklistChecker: checker,
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

		// Set user ID and token in context
		c.Set("userID", uint(userID))
		c.Set("token", tokenString)
		c.Next()
	}
}
