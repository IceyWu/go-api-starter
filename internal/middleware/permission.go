package middleware

import (
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

type PermissionMiddleware struct {
	permService *service.PermissionService
}

func NewPermissionMiddleware(permService *service.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permService: permService,
	}
}

// RequirePermission checks if the user has the required permission
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
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

		// Check if user has the permission
		hasPermission, err := m.permService.CheckUserPermission(userID, permissionCode)
		if err != nil {
			response.InternalError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, "没有权限执行此操作")
			c.Abort()
			return
		}

		c.Next()
	}
}
