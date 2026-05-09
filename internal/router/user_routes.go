package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/middleware"
)

func registerUserRoutes(api *gin.RouterGroup, c *container.Container, authMw *middleware.AuthMiddleware, permMw *middleware.PermissionMiddleware) {
	userH := c.UserHandler()

	users := api.Group("/users")

	// 公开接口（可选认证）— 查看用户公开资料
	users.GET("/:sec_uid", authMw.OptionalAuth(), userH.Get)

	// 需要认证的接口
	users.Use(authMw.RequireAuth())
	{
		// Current user endpoints (self-service)
		users.GET("/me", userH.GetMe)
		users.PUT("/me", userH.UpdateMe)

		// User management endpoints (需要权限)
		users.POST("", permMw.RequirePermission("user.create"), userH.Create)
		users.GET("", permMw.RequirePermission("user.read"), userH.List)
		users.PUT("/:sec_uid", permMw.RequirePermission("user.update"), userH.Update)
		users.DELETE("/:sec_uid", permMw.RequirePermission("user.delete"), userH.Delete)
	}
}
