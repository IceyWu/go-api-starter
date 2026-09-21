package router

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	humachi "github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/jmoiron/sqlx"
	"net/http"
	"net/http/pprof"
	"time"

	httpx "go-api-starter/internal/transport/httpx"
	"golang.org/x/time/rate"

	"go-api-starter/docs"
	"go-api-starter/internal/config"
	"go-api-starter/internal/container"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"
	"go-api-starter/internal/platform/llmstxt"
	"go-api-starter/internal/platform/metrics"
)

// Setup configures and returns the router, permission middleware, and DI container.
func Setup(db *sqlx.DB) (*httpx.Engine, *middleware.PermissionMiddleware, *container.Container) {
	r := httpx.New()
	humaAPI := humachi.New(r.Chi(), huma.DefaultConfig("go-api-starter", "2.0.0"))
	huma.Register(humaAPI, huma.Operation{
		OperationID: "system-ping",
		Method:      http.MethodGet,
		Path:        "/api/v1/system/ping",
		Summary:     "System ping",
		Tags:        []string{"System"},
	}, func(context.Context, *struct{}) (*struct {
		Body struct {
			Status string `json:"status"`
		}
	}, error) {
		return &struct {
			Body struct {
				Status string `json:"status"`
			}
		}{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})
	httpMetrics := metrics.New()

	cfg := config.GetConfig()
	c := container.NewContainer(db, cfg)

	// Core middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.BodyLimit(10 << 20)) // 10MB default body limit
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(httpMetrics.Middleware())

	// Compression and CORS are implemented by the internal transport adapter.
	r.Use(httpx.Gzip())
	r.Use(httpx.CORS(cfg.CORS.AllowOrigins, cfg.CORS.AllowMethods, cfg.CORS.AllowHeaders))

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
		r.GET("/debug/pprof/", httpx.WrapH(http.HandlerFunc(pprof.Index)))
		r.GET("/debug/pprof/cmdline", httpx.WrapH(http.HandlerFunc(pprof.Cmdline)))
		r.GET("/debug/pprof/profile", httpx.WrapH(http.HandlerFunc(pprof.Profile)))
		r.GET("/debug/pprof/symbol", httpx.WrapH(http.HandlerFunc(pprof.Symbol)))
		r.GET("/debug/pprof/trace", httpx.WrapH(http.HandlerFunc(pprof.Trace)))
	}

	// Build shared middleware
	authMw := middleware.NewAuthMiddleware(c.JWTSecret(), c.AuthService(), c.UserRepository())
	permMw := middleware.NewPermissionMiddleware(c.PermissionService())

	// Global base path group (configurable via BASE_PATH env var)
	base := r.Group(cfg.Server.BasePath)

	// Health check routes (no auth)
	base.GET("/health", c.HealthHandler().Health)
	base.GET("/health/ready", c.HealthHandler().Ready)
	base.GET("/metrics", httpx.WrapH(httpMetrics.Handler()))

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
	docsAuth := httpx.BasicAuth(httpx.Accounts{
		cfg.App.DocsUser: cfg.App.DocsPassword,
	})
	base.GET("/swagger/doc.json", docsAuth, func(c *httpx.Context) {
		c.Data(200, "application/json; charset=utf-8", []byte(docs.ReadDoc()))
	})
	base.GET("/docs", docsAuth, handler.DocsHandler)

	// LLMs.txt routes (public, for AI consumption)
	llmsHandler := llmstxt.NewHandler(docs.ReadDoc(), llmstxt.Config{
		BaseURL: "", // 空值表示使用请求时的 Host 动态生成
	})
	llmsHandler.RegisterRoutes(base)

	return r, permMw, c
}
