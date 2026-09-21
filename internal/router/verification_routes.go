package router

import (
	"go-api-starter/internal/transport"

	"go-api-starter/internal/container"
)

func registerVerificationRoutes(api *transport.RouterGroup, c *container.Container) {
	h := c.VerificationHandler()

	verification := api.Group("/verification")
	{
		verification.POST("/send", h.SendCode)
		verification.POST("/verify", h.VerifyCode)
	}
}
