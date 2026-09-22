package router

import (
	"time"

	"go-api-starter/internal/transport"

	"go-api-starter/internal/container"
	"go-api-starter/internal/middleware"
)

func registerFileRoutes(api *transport.RouterGroup, c *container.Container, authMw *middleware.AuthMiddleware) {
	h := c.StorageHandler()

	file := api.Group("/files")

	// 公开上传（无需鉴权）
	file.POST("/public/upload", h.PublicUpload)

	// 可选认证：带 token 可查看/筛选私密文件，不带 token 只返回公开文件
	file.GET("", authMw.OptionalAuth(), h.ListFiles)
	file.GET("/:uid", authMw.OptionalAuth(), h.GetFile)

	// 需要认证
	file.Use(authMw.RequireAuth())
	{
		uploads := api.Group("/uploads")
		{
			uploads.Use(authMw.RequireAuth())
			cfg := c.Config()
			limit := 120
			if cfg != nil && cfg.RateLimit.UploadPerMinute > 0 {
				limit = cfg.RateLimit.UploadPerMinute
			}

			// 仅当 Redis 启用时开启针对上传动作端点的用户级限流。
			var uploadActionMw transport.HandlerFunc
			if cfg != nil && cfg.Redis.Enabled {
				uploadActionMw = middleware.NewRedisRateLimiter(c.CacheBackend(), limit, time.Minute).RateLimitByUser()
			}

			withLimit := func(handlers ...transport.HandlerFunc) []transport.HandlerFunc {
				if uploadActionMw == nil {
					return handlers
				}
				return append([]transport.HandlerFunc{uploadActionMw}, handlers...)
			}

			uploads.POST("", withLimit(h.UploadInit)...)
			uploads.POST("/:upload_id/complete", withLimit(h.UploadComplete)...)
			uploads.DELETE("/:upload_id", h.AbortUpload)
		}

		file.PUT("/:uid", h.UpdateFile)
		file.DELETE("/:uid", h.DeleteFile)
	}
}
