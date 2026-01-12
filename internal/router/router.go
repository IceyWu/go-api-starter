package router

import (
	"go-api-starter/docs"
	"go-api-starter/internal/config"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"
	"go-api-starter/internal/repository"
	"go-api-starter/internal/service"

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

	// Core middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	// Gzip compression middleware
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip"})))

	// CORS middleware - using gin-contrib/cors for better control
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"}
	corsConfig.ExposeHeaders = []string{"X-Request-ID"}
	r.Use(cors.New(corsConfig))

	// Rate limiting - 100 requests per second with burst of 200
	// Adjust these values based on your needs
	rateLimiter := middleware.NewRateLimiter(rate.Limit(100), 200)
	r.Use(rateLimiter.RateLimit())

	// Enable pprof in development environment
	cfg := config.GetConfig()
	if cfg != nil && cfg.App.Env == "development" {
		pprof.Register(r)
	}

	// Initialize handlers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	
	ossRepo := repository.NewOSSRepository(db)
	ossService := service.NewOSSService(ossRepo, &cfg.OSS)
	ossHandler := handler.NewOSSHandler(ossService)
	
	healthHandler := handler.NewHealthHandler(db, "1.0.0")

	// Permission system
	spaceRepo := repository.NewPermissionSpaceRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	rolePermRepo := repository.NewRolePermissionRepository(db)
	cacheRepo := repository.NewUserPermissionCacheRepository(db)
	permManager := service.NewBitPermissionManager(spaceRepo, permRepo, roleRepo, userRoleRepo, rolePermRepo, cacheRepo)
	permService := service.NewPermissionService(permManager)
	permHandler := handler.NewPermissionHandler(permService)

	// Health check routes (no rate limiting)
	r.GET("/health", healthHandler.Health)
	r.GET("/health/ready", healthHandler.Ready)

	// API routes
	api := r.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		oss := api.Group("/oss")
		{
			oss.GET("/token", ossHandler.GetUploadToken)
			oss.POST("/callback", ossHandler.Callback)
			oss.GET("/files", ossHandler.ListFiles)
			oss.DELETE("/files/:id", ossHandler.DeleteFile)
		}

		// Permission management routes
		permissions := api.Group("/permissions")
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
	}

	// Documentation routes
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", handler.DocsHandler)

	return r
}
