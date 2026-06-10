package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/ws"
)

// registerWsRoutes 注册 WebSocket 路由
func registerWsRoutes(r *gin.Engine, c *container.Container) {
	apiKey := c.Config().WS.Key
	r.GET("/ws/bot", ws.Handler(c.WsHub(), apiKey))

	// Bot 状态查询
	r.GET("/api/v1/bot/status", func(ctx *gin.Context) {
		connected := c.WsHub().IsConnected()
		ctx.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{
				"connected": connected,
			},
		})
	})
}
