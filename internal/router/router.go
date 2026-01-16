package router

import (
	"time"

	"go-api-starter/docs"
	"go-api-starter/internal/config"
	"go-api-starter/internal/container"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// Setup configures and returns the router
func Setup(db *gorm.DB) *gin.Engine {
	r := gin.New()

	// Get config
	cfg := config.GetConfig()

	// Create DI container
	c := container.NewContainer(db, cfg)

	// Core middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler()) // Unified error handling

	// Gzip compression middleware
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip"})))

	// CORS middleware - using gin-contrib/cors for better control
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"}
	corsConfig.ExposeHeaders = []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"}
	r.Use(cors.New(corsConfig))

	// Rate limiting - use Redis-backed rate limiter if Redis is enabled
	if cfg.Redis.Enabled {
		// Use Redis-backed distributed rate limiter
		redisRateLimiter := middleware.NewMultiLevelRateLimiter(c.CacheBackend()).
			SetGlobalLimit(100, time.Minute).           // 100 requests per minute globally
			SetUserLimit(60, time.Minute).              // 60 requests per minute per user
			SetEndpointLimit("/api/v1/auth/login", 10, time.Minute) // 10 login attempts per minute
		r.Use(redisRateLimiter.RateLimit())
	} else {
		// Fallback to memory-based rate limiter
		rateLimiter := middleware.NewRateLimiter(rate.Limit(100), 200)
		r.Use(rateLimiter.RateLimit())
	}

	// Enable pprof in development environment
	if cfg != nil && cfg.App.Env == "development" {
		pprof.Register(r)
	}

	// Get handlers from container
	authHandler := c.AuthHandler()
	userHandler := c.UserHandler()
	permHandler := c.PermissionHandler()
	ossHandler := c.OSSHandler()
	healthHandler := c.HealthHandler()
	verifyHandler := c.VerificationHandler()
	operationLogHandler := c.OperationLogHandler()

	// Get middleware dependencies
	authMiddleware := middleware.NewAuthMiddlewareWithBlacklist(c.JWTSecret(), c.AuthService())
	permMiddleware := middleware.NewPermissionMiddleware(c.PermissionService())
	operationLogMiddleware := c.OperationLogMiddleware()

	// Test handler
	testHandler := handler.NewTestHandler()

	// Health check routes (no rate limiting)
	r.GET("/health", healthHandler.Health)
	r.GET("/health/ready", healthHandler.Ready)

	// API routes
	api := r.Group("/api/v1")
	api.Use(operationLogMiddleware.Log()) // 操作日志中间件
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetCurrentUser)
			auth.POST("/reset-password/:id", authMiddleware.RequireAuth(), authHandler.ResetPassword)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
			auth.POST("/logout-all", authMiddleware.RequireAuth(), authHandler.LogoutAllDevices)
		}

		// Verification routes (public)
		verification := api.Group("/verification")
		{
			verification.POST("/send", verifyHandler.SendCode)
			verification.POST("/verify", verifyHandler.VerifyCode)
		}

		// Protected routes
		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:sec_uid", userHandler.Get)
			users.PUT("/:sec_uid", userHandler.Update)
			users.DELETE("/:sec_uid", userHandler.Delete)
		}

		oss := api.Group("/oss")
		oss.Use(authMiddleware.RequireAuth())
		{
			// Simple upload
			oss.GET("/token", ossHandler.GetUploadToken)
			oss.POST("/callback", ossHandler.Callback)
			
			// Multipart upload (分片上传)
			oss.POST("/multipart/init", ossHandler.InitMultipart)
			oss.POST("/multipart/urls", ossHandler.GetPartUploadURLs)
			oss.POST("/multipart/complete", ossHandler.CompleteMultipart)
			oss.POST("/multipart/abort", ossHandler.AbortMultipart)
			oss.GET("/multipart/parts", ossHandler.ListParts)
			oss.POST("/multipart/part", ossHandler.SavePart)
			oss.GET("/multipart/db-parts", ossHandler.GetUploadedPartsFromDB)
			
			// File management
			oss.GET("/files", ossHandler.ListFiles)
			oss.DELETE("/files/:id", ossHandler.DeleteFile)
		}

		// Permission management routes
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware.RequireAuth())
		{
			// Permission spaces
			permissions.POST("/spaces", permHandler.CreateSpace)
			permissions.GET("/spaces", permHandler.GetAllSpaces)

			// Permissions
			permissions.POST("/permissions", permHandler.CreatePermission)
			permissions.GET("/permissions", permHandler.GetAllPermissions)
			permissions.GET("/permissions/:id", permHandler.GetPermission)
			permissions.PUT("/permissions/:id", permHandler.UpdatePermission)
			permissions.DELETE("/permissions/:id", permHandler.DeletePermission)

			// Roles
			permissions.POST("/roles", permHandler.CreateRole)
			permissions.GET("/roles", permHandler.GetAllRoles)
			permissions.GET("/roles/:id", permHandler.GetRole)
			permissions.PUT("/roles/:id", permHandler.UpdateRole)
			permissions.DELETE("/roles/:id", permHandler.DeleteRole)
			permissions.GET("/roles/:id/permissions", permHandler.GetRolePermissions)
			permissions.POST("/roles/:id/permissions", permHandler.AddRolePermissions)
			permissions.DELETE("/roles/:id/permissions", permHandler.RemoveRolePermissions)

			// User roles and permissions
			permissions.GET("/users/:id/roles", permHandler.GetUserRoles)
			permissions.POST("/users/:id/roles", permHandler.AssignUserRole)
			permissions.DELETE("/users/:id/roles/:roleId", permHandler.RemoveUserRole)
			permissions.GET("/users/:id/permissions", permHandler.GetUserPermissions)
			permissions.GET("/me/permissions", permHandler.GetMyPermissions)
		}

		// Test routes for permission validation
		test := api.Group("/test")
		test.Use(authMiddleware.RequireAuth())
		{
			test.GET("/no-permission", testHandler.TestNoPermission)
			test.GET("/user-create", permMiddleware.RequirePermission("user.create"), testHandler.TestUserCreate)
			test.GET("/user-read", permMiddleware.RequirePermission("user.read"), testHandler.TestUserRead)
		}

		// Operation log routes
		operationLogs := api.Group("/operation-logs")
		operationLogs.Use(authMiddleware.RequireAuth())
		{
			operationLogs.GET("", operationLogHandler.List)
			operationLogs.GET("/:id", operationLogHandler.Get)
		}
	}

	// Documentation routes
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", handler.DocsHandler)

	return r
}
