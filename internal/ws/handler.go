package ws

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"go-api-starter/internal/transport"
)

// Handler WebSocket 握手入口
func Handler(hub *Hub, apiKey string) transport.HandlerFunc {
	return func(c *transport.Context) {
		// API Key 验证（使用常量时间比较防止时序攻击）
		if apiKey != "" {
			key := c.Query("key")
			if key == "" {
				key = c.GetHeader("X-API-Key")
			}
			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				c.JSON(http.StatusUnauthorized, transport.H{"error": "invalid api key"})
				return
			}
		}

		conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // wechat_hook 内部通信，允许所有来源
		})
		if err != nil {
			log.Printf("[WS Handler] 升级失败: %v", err)
			return
		}

		hub.Register(conn)
	}
}
