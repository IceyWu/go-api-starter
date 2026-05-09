package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-api-starter/internal/container"
	"go-api-starter/internal/middleware"
)

func registerFileRoutes(api *gin.RouterGroup, c *container.Container, authMw *middleware.AuthMiddleware) {
	h := c.OSSHandler()

	file := api.Group("/file")

	// 公开上传（无需鉴权）
	file.POST("/public/upload", h.PublicUpload)

	// 可选认证：带 token 可查看/筛选私密文件，不带 token 只返回公开文件
	file.GET("", authMw.OptionalAuth(), h.ListFiles)
	file.GET("/:sec_uid", authMw.OptionalAuth(), h.GetFile)

	// 需要认证
	file.Use(authMw.RequireAuth())
	{
		upload := file.Group("/upload")
		{
			cfg := c.Config()
			limit := 120
			if cfg != nil && cfg.RateLimit.UploadPerMinute > 0 {
				limit = cfg.RateLimit.UploadPerMinute
			}

			// 仅当 Redis 启用时开启针对上传动作端点的用户级限流。
			var uploadActionMw gin.HandlerFunc
			if cfg != nil && cfg.Redis.Enabled {
				uploadActionMw = middleware.NewRedisRateLimiter(c.CacheBackend(), limit, time.Minute).RateLimitByUser()
			}

			withLimit := func(handlers ...gin.HandlerFunc) []gin.HandlerFunc {
				if uploadActionMw == nil {
					return handlers
				}
				return append([]gin.HandlerFunc{uploadActionMw}, handlers...)
			}

			upload.POST("/init", withLimit(h.UploadInit)...)
			upload.POST("/complete", withLimit(h.UploadComplete)...)
			// 辅助接口：总在 /init 之后调用，不再限流
			upload.POST("/urls", h.GetPartUploadURLs)
			upload.POST("/abort", h.AbortMultipart)
		}

		file.PUT("/:sec_uid", h.UpdateFile)
		file.DELETE("/:sec_uid", h.DeleteFile)
	}
}
