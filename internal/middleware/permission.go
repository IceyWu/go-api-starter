package middleware

import (
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware holds the permission service
type PermissionMiddleware struct {
	permService *service.PermissionService
}

// NewPermissionMiddleware creates a new PermissionMiddleware
func NewPermissionMiddleware(permService *service.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{permService: permService}
}

// RequirePermissions returns a middleware that checks if the user has all required permissions
func (m *PermissionMiddleware) RequirePermissions(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDVal, exists := c.Get("userID")
		if !exists {
			response.Unauthorized(c, "用户未认证")
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uint)
		if !ok {
			response.Unauthorized(c, "用户ID无效")
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		// Check all required permissions
		for _, code := range codes {
			hasPermission, err := m.permService.HasPermission(ctx, userID, code)
			if err != nil {
				response.InternalError(c, "权限检查失败")
				c.Abort()
				return
			}
			if !hasPermission {
				response.Forbidden(c, "缺少权限: "+code)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermission returns a middleware that checks if the user has any of the required permissions
func (m *PermissionMiddleware) RequireAnyPermission(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			response.Unauthorized(c, "用户未认证")
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uint)
		if !ok {
			response.Unauthorized(c, "用户ID无效")
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		for _, code := range codes {
			hasPermission, err := m.permService.HasPermission(ctx, userID, code)
			if err != nil {
				continue
			}
			if hasPermission {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "缺少所需权限")
		c.Abort()
	}
}
