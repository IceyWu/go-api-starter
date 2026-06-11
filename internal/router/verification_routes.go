package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
)

func registerVerificationRoutes(api *gin.RouterGroup, c *container.Container) {
	h := c.VerificationHandler()

	verification := api.Group("/verification")
	{
		verification.POST("/send", h.SendCode)
		verification.POST("/verify", h.VerifyCode)
	}
}
