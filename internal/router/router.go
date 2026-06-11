package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"go-api-starter/docs"
	"go-api-starter/internal/config"
	"go-api-starter/internal/container"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"
	"go-api-starter/pkg/llmstxt"
)

// Setup configures and returns the router, permission middleware, and DI container.
func Setup(db *gorm.DB) (*gin.Engine, *middleware.PermissionMiddleware, *container.Container) {
	r := gin.New()

	cfg := config.GetConfig()
	c := container.NewContainer(db, cfg)

	// Core middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.BodyLimit(10 << 20)) // 10MB default body limit
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	// Gzip compression middleware
	r.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithExcludedExtensions([]string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip"}),
	))

	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowOrigins
	corsConfig.AllowMethods = cfg.CORS.AllowMethods
	corsConfig.AllowHeaders = cfg.CORS.AllowHeaders
	corsConfig.ExposeHeaders = []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"}
	r.Use(cors.New(corsConfig))

	// Rate limiting
	if cfg.Redis.Enabled {
		redisRateLimiter := middleware.NewMultiLevelRateLimiter(c.CacheBackend()).
			SetGlobalLimit(cfg.RateLimit.GlobalPerMinute, time.Minute).
			SetUserLimit(cfg.RateLimit.UserPerMinute, time.Minute).
			SetEndpointLimit("/api/v1/auth/login", cfg.RateLimit.LoginPerMinute, time.Minute)
		r.Use(redisRateLimiter.RateLimit())
	} else {
		rateLimiter := middleware.NewRateLimiter(rate.Limit(cfg.RateLimit.FallbackRPS), cfg.RateLimit.FallbackBurst)
		r.Use(rateLimiter.RateLimit())
	}

	// pprof in development
	if cfg != nil && cfg.App.Env == "development" {
		pprof.Register(r)
	}

	// Build shared middleware
	authMw := middleware.NewAuthMiddleware(c.JWTSecret(), c.AuthService(), c.UserRepository())
	permMw := middleware.NewPermissionMiddleware(c.PermissionService())

	// Global base path group (configurable via BASE_PATH env var)
	base := r.Group(cfg.Server.BasePath)

	// Health check routes (no auth)
	base.GET("/health", c.HealthHandler().Health)
	base.GET("/health/ready", c.HealthHandler().Ready)

	// Static files (logo, favicon)
	base.StaticFile("/logo.svg", "./public/logo.svg")
	base.StaticFile("/favicon.ico", "./public/favicon.ico")

	// API routes
	api := base.Group("/api/v1")

	// Register module routes
	registerAuthRoutes(api, c, authMw)
	registerUserRoutes(api, c, authMw, permMw)
	registerFileRoutes(api, c, authMw)
	registerPermissionRoutes(api, c, authMw, permMw)
	registerVerificationRoutes(api, c)

	// WebSocket route
	registerWsRoutes(base, c)

	// Documentation routes (protected by Basic Auth)
	docs.SwaggerInfo.BasePath = cfg.Server.BasePath + "/"
	docs.SwaggerInfo.Host = ""
	docsAuth := gin.BasicAuth(gin.Accounts{
		cfg.App.DocsUser: cfg.App.DocsPassword,
	})
	base.GET("/swagger/*any", docsAuth, ginSwagger.WrapHandler(swaggerFiles.Handler))
	base.GET("/docs", docsAuth, handler.DocsHandler)

	// LLMs.txt routes (public, for AI consumption)
	llmsHandler := llmstxt.NewHandler(docs.SwaggerInfo.ReadDoc(), llmstxt.Config{
		BaseURL: "", // 空值表示使用请求时的 Host 动态生成
	})
	llmsHandler.RegisterRoutes(base)

	return r, permMw, c
}
