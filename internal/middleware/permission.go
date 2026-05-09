package middleware

import (
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"
	"sync"

	"github.com/gin-gonic/gin"
)

type PermissionMiddleware struct {
	permService    service.PermissionServiceInterface
	collectedCodes map[string]struct{}
	mu             sync.Mutex
}

func NewPermissionMiddleware(permService service.PermissionServiceInterface) *PermissionMiddleware {
	return &PermissionMiddleware{
		permService:    permService,
		collectedCodes: make(map[string]struct{}),
	}
}

// RequirePermission checks if the user has the required permission.
// It also collects the permission code for auto-seeding.
func (m *PermissionMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	// 路由注册阶段自动收集 code
	m.mu.Lock()
	m.collectedCodes[permissionCode] = struct{}{}
	m.mu.Unlock()

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

// CollectedCodes returns all permission codes that were registered via RequirePermission.
func (m *PermissionMiddleware) CollectedCodes() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	codes := make([]string, 0, len(m.collectedCodes))
	for code := range m.collectedCodes {
		codes = append(codes, code)
	}
	return codes
}
