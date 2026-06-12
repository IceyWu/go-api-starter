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
		permissions.POST("/spaces", permMw.RequirePermission("permission.create"), h.CreateSpace)
		permissions.GET("/spaces", h.GetAllSpaces)
		permissions.PUT("/spaces/:id", permMw.RequirePermission("permission.update"), h.UpdateSpace)
		permissions.DELETE("/spaces/:id", permMw.RequirePermission("permission.delete"), h.DeleteSpace)

		// Permissions
		permissions.POST("/permissions", permMw.RequirePermission("permission.create"), h.CreatePermission)
		permissions.GET("/permissions", h.GetAllPermissions)
		permissions.GET("/permissions/:id", h.GetPermission)
		permissions.PUT("/permissions/:id", permMw.RequirePermission("permission.update"), h.UpdatePermission)
		permissions.DELETE("/permissions/:id", permMw.RequirePermission("permission.delete"), h.DeletePermission)

		// Roles
		permissions.POST("/roles", permMw.RequirePermission("role.create"), h.CreateRole)
		permissions.GET("/roles", h.GetAllRoles)
		permissions.GET("/roles/:id", h.GetRole)
		permissions.PUT("/roles/:id", permMw.RequirePermission("role.update"), h.UpdateRole)
		permissions.DELETE("/roles/:id", permMw.RequirePermission("role.delete"), h.DeleteRole)
		permissions.GET("/roles/:id/permissions", h.GetRolePermissions)
		permissions.POST("/roles/:id/permissions", permMw.RequirePermission("role.update"), h.AddRolePermissions)
		permissions.DELETE("/roles/:id/permissions", permMw.RequirePermission("role.update"), h.RemoveRolePermissions)

		// User roles
		permissions.GET("/users/:uid/roles", h.GetUserRolesByUID)
		permissions.POST("/users/:uid/roles", permMw.RequirePermission("role.update"), h.AssignUserRoleByUID)
		permissions.DELETE("/users/:uid/roles/:roleId", permMw.RequirePermission("role.update"), h.RemoveUserRoleByUID)
		permissions.GET("/me/permissions", h.GetMyPermissions)
	}
}
