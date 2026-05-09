package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/middleware"
)

func registerAuthRoutes(api *gin.RouterGroup, c *container.Container, authMw *middleware.AuthMiddleware) {
	h := c.AuthHandler()

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/reset-password/:id", authMw.RequireAuth(), h.ResetPassword)
		auth.POST("/logout", authMw.RequireAuth(), h.Logout)
		auth.POST("/logout-all", authMw.RequireAuth(), h.LogoutAllDevices)
	}
}
