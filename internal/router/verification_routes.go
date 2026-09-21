package router

import (
	httpx "go-api-starter/internal/transport/httpx"

	"go-api-starter/internal/container"
)

func registerVerificationRoutes(api *httpx.RouterGroup, c *container.Container) {
	h := c.VerificationHandler()

	verification := api.Group("/verification")
	{
		verification.POST("/send", h.SendCode)
		verification.POST("/verify", h.VerifyCode)
	}
}
