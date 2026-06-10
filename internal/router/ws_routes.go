package router

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/ws"
)

// registerWsRoutes 注册 WebSocket 路由
func registerWsRoutes(r *gin.Engine, c *container.Container) {
	apiKey := c.Config().WS.Key
	r.GET("/ws", ws.Handler(c.WsHub(), apiKey))

	// Bot 状态查询
	r.GET("/api/v1/ws/status", botStatus(c))
}

// botStatus godoc
// @Summary 查询 WebSocket 连接状态
// @Description 返回当前 WebSocket 客户端是否已连接
// @Tags WebSocket
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ws/status [get]
func botStatus(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		connected := c.WsHub().IsConnected()
		ctx.JSON(200, gin.H{
			"code": 200,
			"data": gin.H{
				"connected": connected,
			},
		})
	}
}
