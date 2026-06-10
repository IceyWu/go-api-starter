package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/middleware"
)

func registerPermissionRoutes(api *gin.RouterGroup, c *container.Container, authMw *middleware.AuthMiddleware, permMw *middleware.PermissionMiddleware) {
	h := c.PermissionHandler()

	permissions := api.Group("/permissions")
	permissions.Use(authMw.RequireAuth())
	{
		// Permission spaces
		permissions.POST("/spaces", permMw.RequirePermission("role.manage"), h.CreateSpace)
		permissions.GET("/spaces", h.GetAllSpaces)

		// Permissions
		permissions.POST("/permissions", permMw.RequirePermission("role.manage"), h.CreatePermission)
		permissions.GET("/permissions", h.GetAllPermissions)
		permissions.GET("/permissions/:id", h.GetPermission)
		permissions.PUT("/permissions/:id", permMw.RequirePermission("role.manage"), h.UpdatePermission)
		permissions.DELETE("/permissions/:id", permMw.RequirePermission("role.manage"), h.DeletePermission)

		// Roles
		permissions.POST("/roles", permMw.RequirePermission("role.manage"), h.CreateRole)
		permissions.GET("/roles", h.GetAllRoles)
		permissions.GET("/roles/:id", h.GetRole)
		permissions.PUT("/roles/:id", permMw.RequirePermission("role.manage"), h.UpdateRole)
		permissions.DELETE("/roles/:id", permMw.RequirePermission("role.manage"), h.DeleteRole)
		permissions.GET("/roles/:id/permissions", h.GetRolePermissions)
		permissions.POST("/roles/:id/permissions", permMw.RequirePermission("role.manage"), h.AddRolePermissions)
		permissions.DELETE("/roles/:id/permissions", permMw.RequirePermission("role.manage"), h.RemoveRolePermissions)

		// User roles
		permissions.GET("/users/:uid/roles", h.GetUserRolesByUID)
		permissions.POST("/users/:uid/roles", permMw.RequirePermission("role.manage"), h.AssignUserRoleByUID)
		permissions.DELETE("/users/:uid/roles/:roleId", permMw.RequirePermission("role.manage"), h.RemoveUserRoleByUID)
		permissions.GET("/me/permissions", h.GetMyPermissions)
	}
}
