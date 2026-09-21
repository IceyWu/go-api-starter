package main

import (
	"log"

	"go-api-starter/internal/app"
)

// @title Go API Starter
// @version 1.0
// @description 🚀 基于 Go + Chi + Huma + sqlc + Atlas 构建的高性能 RESTful API 脚手架
// @x-logo {"url": "/logo.svg", "altText": "Go API Starter"}
// @description
// @description ---
// @description
// @description ## 🔌 WebSocket
// @description
// @description | 项目 | 说明 |
// @description |------|------|
// @description | 入口 | `GET /ws` — 长连接入口，query 参数 `key` 或 header `X-API-Key` 认证 |
// @description | 消息格式 | `{"type":"...", "id":"...", "data":{}}` |
// @description | ⬇️ 下行指令 | `send_text_msg`、`get_group_list`、`ping` |
// @description | ⬆️ 上行消息 | `ack`、`review`、`pong` |
// @description
// @description ---
// @description
// @description ## 🤖 LLMs 入口
// @description
// @description | 文件 | 说明 |
// @description |------|------|
// @description | 📄 [llms.txt](/llms.txt) | AI 可读接口概览 |
// @description | 📚 [llms-full.txt](/llms-full.txt) | AI 可读完整文档 |
// @host localhost:9527
// @BasePath /

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
