package router

import (
	httpx "go-api-starter/internal/transport/httpx"

	"go-api-starter/internal/container"
	"go-api-starter/internal/ws"
)

// registerWsRoutes 注册 WebSocket 路由
func registerWsRoutes(base *httpx.RouterGroup, c *container.Container) {
	apiKey := c.Config().WS.Key
	base.GET("/ws", ws.Handler(c.WsHub(), apiKey))

	// Bot 状态查询
	base.GET("/api/v1/ws/status", botStatus(c))
}

// botStatus godoc
// @Summary 查询 WebSocket 连接状态
// @Description 返回当前 WebSocket 客户端是否已连接
// @Tags WebSocket
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ws/status [get]
func botStatus(c *container.Container) httpx.HandlerFunc {
	return func(ctx *httpx.Context) {
		connected := c.WsHub().IsConnected()
		ctx.JSON(200, httpx.H{
			"code": 200,
			"data": httpx.H{
				"connected": connected,
			},
		})
	}
}
