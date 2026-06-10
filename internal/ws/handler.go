package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // wechat_hook 内部通信，允许所有来源
	},
}

// Handler WebSocket 握手入口
func Handler(hub *Hub, apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// API Key 验证
		if apiKey != "" {
			key := c.Query("key")
			if key == "" {
				key = c.GetHeader("X-API-Key")
			}
			if key != apiKey {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				return
			}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WS Handler] 升级失败: %v", err)
			return
		}

		hub.Register(conn)
	}
}
