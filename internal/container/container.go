package container

import (
	"context"
	"github.com/jmoiron/sqlx"
	"io"
	"log/slog"
	"sync"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"
	"go-api-starter/internal/platform/auth"
	"go-api-starter/internal/platform/cache"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/mail"
	"go-api-starter/internal/platform/storage"
	"go-api-starter/internal/repository"
	"go-api-starter/internal/service"
	"go-api-starter/internal/ws"
)

// Container is the dependency injection container for the slim starter.
type Container struct {
	db     *sqlx.DB
	config *config.Config

	// Repositories
	userRepo          repository.UserRepositoryInterface
	userRepoOnce      sync.Once
	permRepo          repository.PermissionRepositoryInterface
	permRepoOnce      sync.Once
	roleRepo          repository.RoleRepositoryInterface
	roleRepoOnce      sync.Once
	spaceRepo         repository.PermissionSpaceRepositoryInterface
	spaceRepoOnce     sync.Once
	userRoleRepo      repository.UserRoleRepositoryInterface
	userRoleRepoOnce  sync.Once
	rolePermRepo      repository.RolePermissionRepositoryInterface
	rolePermRepoOnce  sync.Once
	cacheRepo         repository.UserPermissionCacheRepositoryInterface
	cacheRepoOnce     sync.Once
	multipartRepo     repository.MultipartRepositoryInterface
	multipartRepoOnce sync.Once
	fileRepo          repository.FileRepositoryInterface
	fileRepoOnce      sync.Once

	// Services
	authService         service.AuthServiceInterface
	authServiceOnce     sync.Once
	userService         service.UserServiceInterface
	userServiceOnce     sync.Once
	permService         service.PermissionServiceInterface
	permServiceOnce     sync.Once
	storageService      service.StorageServiceInterface
	storageServiceOnce  sync.Once
	storageProvider     storage.ObjectStorage
	storageProviderOnce sync.Once
	taskManager         *service.TaskManager
	taskManagerOnce     sync.Once
	tokenBlacklist      service.TokenBlacklist
	tokenBlacklistOnce  sync.Once
	wechatService       *service.WechatService
	wechatServiceOnce   sync.Once

	// Permission components
	permManager     *service.BitPermissionManager
	permManagerOnce sync.Once
	permChecker     *service.PermissionChecker
	permCheckerOnce sync.Once
	permCache       *service.PermissionCache
	permCacheOnce   sync.Once

	// Cache components
	redisCache       *cache.RedisCache
	redisCacheOnce   sync.Once
	memoryCache      *cache.MemoryCache
	memoryCacheOnce  sync.Once
	cacheBackend     cache.CacheBackend
	cacheBackendOnce sync.Once

	// Rate limiter
	rateLimiter     *middleware.RedisRateLimiter
	rateLimiterOnce sync.Once

	// Handlers
	authHandler        *handler.AuthHandler
	authHandlerOnce    sync.Once
	userHandler        *handler.UserHandler
	userHandlerOnce    sync.Once
	permHandler        *handler.PermissionHandler
	permHandlerOnce    sync.Once
	storageHandler     *handler.StorageHandler
	storageHandlerOnce sync.Once
	healthHandler      *handler.HealthHandler
	healthHandlerOnce  sync.Once
	verifyHandler      *handler.VerificationHandler
	verifyHandlerOnce  sync.Once

	// Mail
	mailClient     *mail.Client
	mailClientOnce sync.Once

	// Services (additional)
	verifyService     *service.VerificationCodeService
	verifyServiceOnce sync.Once

	// JWT manager
	jwtManager     *auth.JWTManager
	jwtManagerOnce sync.Once

	// WebSocket hub
	wsHub     *ws.Hub
	wsHubOnce sync.Once
}

// NewContainer creates a new dependency injection container
func NewContainer(db *sqlx.DB, cfg *config.Config) *Container {
	return &Container{
		db:     db,
		config: cfg,
	}
}

// DB returns the database connection
func (c *Container) DB() *sqlx.DB {
	return c.db
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// ========== Service Getters ==========

func (c *Container) AuthService() service.AuthServiceInterface {
	c.authServiceOnce.Do(func() {
		c.authService = service.NewAuthService(
			c.UserRepository(), c.JWTManager(), c.TokenBlacklist(),
		)
	})
	return c.authService
}

func (c *Container) UserService() service.UserServiceInterface {
	c.userServiceOnce.Do(func() {
		c.userService = service.NewUserService(
			c.UserRepository(), c.FileRepository(),
		)
	})
	return c.userService
}

func (c *Container) PermissionCache() *service.PermissionCache {
	c.permCacheOnce.Do(func() {
		c.permCache = service.NewPermissionCache(
			c.UserPermissionCacheRepository(), time.Hour,
		)
	})
	return c.permCache
}

func (c *Container) PermissionChecker() *service.PermissionChecker {
	c.permCheckerOnce.Do(func() {
		c.permChecker = service.NewPermissionChecker(
			c.PermissionRepository(),
			c.RolePermissionRepository(),
			c.UserRoleRepository(),
			c.PermissionCache(),
		)
	})
	return c.permChecker
}

func (c *Container) BitPermissionManager() *service.BitPermissionManager {
	c.permManagerOnce.Do(func() {
		c.permManager = service.NewBitPermissionManager(
			c.PermissionSpaceRepository(),
			c.PermissionRepository(),
			c.RoleRepository(),
			c.UserRoleRepository(),
			c.RolePermissionRepository(),
			c.UserPermissionCacheRepository(),
		)
	})
	return c.permManager
}

func (c *Container) PermissionService() service.PermissionServiceInterface {
	c.permServiceOnce.Do(func() {
		c.permService = service.NewPermissionService(
			c.BitPermissionManager(), c.PermissionChecker(), c.PermissionCache(),
		)
	})
	return c.permService
}

func (c *Container) StorageService() service.StorageServiceInterface {
	c.storageServiceOnce.Do(func() {
		c.storageService = service.NewStorageService(
			c.db, c.FileRepository(), c.MultipartRepository(),
			c.StorageProvider(), &c.config.Storage, c.config.App.Env,
			c.config.App.AMapAPIKey,
		)
	})
	return c.storageService
}

// StorageProvider returns the configured provider-neutral object storage client.
func (c *Container) StorageProvider() storage.ObjectStorage {
	c.storageProviderOnce.Do(func() {
		if c.config.Storage.Endpoint == "" || c.config.Storage.Bucket == "" || c.config.Storage.AccessKeyID == "" || c.config.Storage.AccessKeySecret == "" {
			return
		}
		provider, err := storage.NewS3Provider(context.Background(), &c.config.Storage)
		if err == nil {
			c.storageProvider = provider
		}
	})
	return c.storageProvider
}

func (c *Container) TokenBlacklist() service.TokenBlacklist {
	c.tokenBlacklistOnce.Do(func() {
		c.tokenBlacklist = service.NewRedisTokenBlacklist(c.CacheBackend())
	})
	return c.tokenBlacklist
}

// ========== Infrastructure ==========

func (c *Container) JWTSecret() string {
	jwtSecret := c.config.App.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}
	return jwtSecret
}

func (c *Container) JWTManager() *auth.JWTManager {
	c.jwtManagerOnce.Do(func() {
		c.jwtManager = auth.NewJWTManager(
			c.JWTSecret(),
			c.config.App.AccessTokenDays,
			c.config.App.RefreshTokenDays,
		)
	})
	return c.jwtManager
}

func (c *Container) RedisCache() *cache.RedisCache {
	c.redisCacheOnce.Do(func() {
		if c.config.Redis.Enabled {
			var err error
			c.redisCache, err = cache.NewRedisCache(&c.config.Redis)
			if err != nil {
				logger.Log.Warn("failed to connect to Redis", "error", err)
			}
		}
	})
	return c.redisCache
}

func (c *Container) MemoryCache() *cache.MemoryCache {
	c.memoryCacheOnce.Do(func() {
		c.memoryCache = cache.NewMemoryCache()
	})
	return c.memoryCache
}

func (c *Container) CacheBackend() cache.CacheBackend {
	c.cacheBackendOnce.Do(func() {
		memCache := c.MemoryCache()
		redisCache := c.RedisCache()

		if redisCache != nil && c.config.Redis.EnableFallback {
			c.cacheBackend = cache.NewFallbackCache(redisCache, memCache, slog.New(slog.NewTextHandler(io.Discard, nil)))
		} else if redisCache != nil {
			c.cacheBackend = redisCache
		} else {
			c.cacheBackend = memCache
		}
	})
	return c.cacheBackend
}

func (c *Container) RateLimiter() *middleware.RedisRateLimiter {
	c.rateLimiterOnce.Do(func() {
		c.rateLimiter = middleware.NewRedisRateLimiter(c.CacheBackend(), 100, time.Minute)
	})
	return c.rateLimiter
}

// WsHub returns the WebSocket hub singleton
func (c *Container) WsHub() *ws.Hub {
	c.wsHubOnce.Do(func() {
		c.wsHub = ws.NewHub()
	})
	return c.wsHub
}

// Close closes all resources
func (c *Container) Close() error {
	if c.redisCache != nil {
		c.redisCache.Close()
	}
	if c.memoryCache != nil {
		c.memoryCache.Close()
	}
	return nil
}
