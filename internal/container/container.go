package container

import (
	"log"
	"sync"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/handler"
	"go-api-starter/internal/middleware"
	"go-api-starter/internal/repository"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/cache"
	"go-api-starter/pkg/mail"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Container is the dependency injection container
type Container struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger

	// Repositories (lazy initialized)
	userRepo      repository.UserRepositoryInterface
	permRepo      repository.PermissionRepositoryInterface
	roleRepo      repository.RoleRepositoryInterface
	spaceRepo     repository.PermissionSpaceRepositoryInterface
	userRoleRepo  repository.UserRoleRepositoryInterface
	rolePermRepo  repository.RolePermissionRepositoryInterface
	cacheRepo     repository.UserPermissionCacheRepositoryInterface
	ossRepo       repository.OSSRepositoryInterface
	multipartRepo repository.MultipartRepositoryInterface

	// Services (lazy initialized)
	authService    service.AuthServiceInterface
	userService    service.UserServiceInterface
	permService    service.PermissionServiceInterface
	ossService     service.OSSServiceInterface
	tokenBlacklist service.TokenBlacklist

	// Permission components
	permManager *service.BitPermissionManager
	permChecker *service.PermissionChecker
	permCache   *service.PermissionCache

	// Cache components
	redisCache   *cache.RedisCache
	memoryCache  *cache.MemoryCache
	cacheBackend cache.CacheBackend

	// Rate limiter
	rateLimiter *middleware.RedisRateLimiter

	// Handlers (lazy initialized)
	authHandler   *handler.AuthHandler
	userHandler   *handler.UserHandler
	permHandler   *handler.PermissionHandler
	ossHandler    *handler.OSSHandler
	healthHandler *handler.HealthHandler
	verifyHandler *handler.VerificationHandler

	// Mail client
	mailClient *mail.Client

	// Verification service
	verifyService *service.VerificationCodeService

	// Mutex for thread-safe lazy initialization
	mu sync.Mutex
}

// NewContainer creates a new dependency injection container
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	return &Container{
		db:     db,
		config: cfg,
	}
}

// NewContainerWithLogger creates a new container with logger
func NewContainerWithLogger(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Container {
	return &Container{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// DB returns the database connection
func (c *Container) DB() *gorm.DB {
	return c.db
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// ========== Repository Getters ==========

// UserRepository returns the user repository singleton
func (c *Container) UserRepository() repository.UserRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.userRepo == nil {
		c.userRepo = repository.NewUserRepository(c.db)
	}
	return c.userRepo
}

// PermissionRepository returns the permission repository singleton
func (c *Container) PermissionRepository() repository.PermissionRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permRepo == nil {
		c.permRepo = repository.NewPermissionRepository(c.db)
	}
	return c.permRepo
}

// RoleRepository returns the role repository singleton
func (c *Container) RoleRepository() repository.RoleRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.roleRepo == nil {
		c.roleRepo = repository.NewRoleRepository(c.db)
	}
	return c.roleRepo
}

// PermissionSpaceRepository returns the permission space repository singleton
func (c *Container) PermissionSpaceRepository() repository.PermissionSpaceRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.spaceRepo == nil {
		c.spaceRepo = repository.NewPermissionSpaceRepository(c.db)
	}
	return c.spaceRepo
}

// UserRoleRepository returns the user role repository singleton
func (c *Container) UserRoleRepository() repository.UserRoleRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.userRoleRepo == nil {
		c.userRoleRepo = repository.NewUserRoleRepository(c.db)
	}
	return c.userRoleRepo
}

// RolePermissionRepository returns the role permission repository singleton
func (c *Container) RolePermissionRepository() repository.RolePermissionRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rolePermRepo == nil {
		c.rolePermRepo = repository.NewRolePermissionRepository(c.db)
	}
	return c.rolePermRepo
}

// UserPermissionCacheRepository returns the user permission cache repository singleton
func (c *Container) UserPermissionCacheRepository() repository.UserPermissionCacheRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cacheRepo == nil {
		c.cacheRepo = repository.NewUserPermissionCacheRepository(c.db)
	}
	return c.cacheRepo
}

// OSSRepository returns the OSS repository singleton
func (c *Container) OSSRepository() repository.OSSRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ossRepo == nil {
		c.ossRepo = repository.NewOSSRepository(c.db)
	}
	return c.ossRepo
}

// MultipartRepository returns the multipart repository singleton
func (c *Container) MultipartRepository() repository.MultipartRepositoryInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.multipartRepo == nil {
		c.multipartRepo = repository.NewMultipartRepository(c.db)
	}
	return c.multipartRepo
}


// ========== Service Getters ==========

// AuthService returns the auth service singleton
func (c *Container) AuthService() service.AuthServiceInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.authService == nil {
		jwtSecret := c.config.App.JWTSecret
		if jwtSecret == "" {
			jwtSecret = "your-secret-key-change-in-production"
		}
		// Need to get dependencies without lock to avoid deadlock
		c.mu.Unlock()
		userRepo := c.UserRepository()
		blacklist := c.TokenBlacklist()
		c.mu.Lock()
		c.authService = service.NewAuthServiceWithBlacklist(userRepo, jwtSecret, blacklist)
	}
	return c.authService
}

// UserService returns the user service singleton
func (c *Container) UserService() service.UserServiceInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.userService == nil {
		c.mu.Unlock()
		userRepo := c.UserRepository()
		c.mu.Lock()
		c.userService = service.NewUserService(userRepo)
	}
	return c.userService
}

// PermissionCache returns the permission cache singleton
func (c *Container) PermissionCache() *service.PermissionCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permCache == nil {
		c.mu.Unlock()
		cacheRepo := c.UserPermissionCacheRepository()
		c.mu.Lock()
		// Default TTL of 1 hour
		c.permCache = service.NewPermissionCache(cacheRepo, time.Hour)
	}
	return c.permCache
}

// PermissionChecker returns the permission checker singleton
func (c *Container) PermissionChecker() *service.PermissionChecker {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permChecker == nil {
		c.mu.Unlock()
		permRepo := c.PermissionRepository()
		rolePermRepo := c.RolePermissionRepository()
		userRoleRepo := c.UserRoleRepository()
		cache := c.PermissionCache()
		c.mu.Lock()
		c.permChecker = service.NewPermissionChecker(permRepo, rolePermRepo, userRoleRepo, cache)
	}
	return c.permChecker
}

// BitPermissionManager returns the bit permission manager singleton
func (c *Container) BitPermissionManager() *service.BitPermissionManager {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permManager == nil {
		c.mu.Unlock()
		spaceRepo := c.PermissionSpaceRepository().(*repository.PermissionSpaceRepository)
		permRepo := c.PermissionRepository().(*repository.PermissionRepository)
		roleRepo := c.RoleRepository().(*repository.RoleRepository)
		userRoleRepo := c.UserRoleRepository().(*repository.UserRoleRepository)
		rolePermRepo := c.RolePermissionRepository().(*repository.RolePermissionRepository)
		cacheRepo := c.UserPermissionCacheRepository().(*repository.UserPermissionCacheRepository)
		c.mu.Lock()
		c.permManager = service.NewBitPermissionManager(spaceRepo, permRepo, roleRepo, userRoleRepo, rolePermRepo, cacheRepo)
	}
	return c.permManager
}

// PermissionService returns the permission service singleton
func (c *Container) PermissionService() service.PermissionServiceInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permService == nil {
		c.mu.Unlock()
		manager := c.BitPermissionManager()
		checker := c.PermissionChecker()
		cache := c.PermissionCache()
		c.mu.Lock()
		c.permService = service.NewPermissionServiceWithComponents(manager, checker, cache)
	}
	return c.permService
}

// OSSService returns the OSS service singleton
func (c *Container) OSSService() service.OSSServiceInterface {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ossService == nil {
		c.mu.Unlock()
		ossRepo := c.OSSRepository()
		multipartRepo := c.MultipartRepository()
		c.mu.Lock()
		c.ossService = service.NewOSSService(ossRepo, multipartRepo, &c.config.OSS)
	}
	return c.ossService
}

// ========== Handler Getters ==========

// AuthHandler returns the auth handler singleton
func (c *Container) AuthHandler() *handler.AuthHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.authHandler == nil {
		c.mu.Unlock()
		authService := c.AuthService()
		verifyService := c.VerificationCodeService()
		c.mu.Lock()
		c.authHandler = handler.NewAuthHandlerWithVerification(authService, verifyService)
	}
	return c.authHandler
}

// UserHandler returns the user handler singleton
func (c *Container) UserHandler() *handler.UserHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.userHandler == nil {
		c.mu.Unlock()
		userService := c.UserService()
		c.mu.Lock()
		c.userHandler = handler.NewUserHandler(userService)
	}
	return c.userHandler
}

// PermissionHandler returns the permission handler singleton
func (c *Container) PermissionHandler() *handler.PermissionHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permHandler == nil {
		c.mu.Unlock()
		permService := c.PermissionService()
		c.mu.Lock()
		c.permHandler = handler.NewPermissionHandler(permService)
	}
	return c.permHandler
}

// OSSHandler returns the OSS handler singleton
func (c *Container) OSSHandler() *handler.OSSHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ossHandler == nil {
		c.mu.Unlock()
		ossService := c.OSSService()
		c.mu.Lock()
		c.ossHandler = handler.NewOSSHandler(ossService)
	}
	return c.ossHandler
}

// HealthHandler returns the health handler singleton
func (c *Container) HealthHandler() *handler.HealthHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.healthHandler == nil {
		c.mu.Unlock()
		cacheBackend := c.CacheBackend()
		c.mu.Lock()
		c.healthHandler = handler.NewHealthHandlerWithCache(c.db, "1.0.0", cacheBackend)
	}
	return c.healthHandler
}

// JWTSecret returns the JWT secret from config
func (c *Container) JWTSecret() string {
	jwtSecret := c.config.App.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}
	return jwtSecret
}

// ========== Cache Getters ==========

// RedisCache returns the Redis cache singleton
func (c *Container) RedisCache() *cache.RedisCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.redisCache == nil && c.config.Redis.Enabled {
		var err error
		c.redisCache, err = cache.NewRedisCache(&c.config.Redis)
		if err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
		}
	}
	return c.redisCache
}

// MemoryCache returns the memory cache singleton
func (c *Container) MemoryCache() *cache.MemoryCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.memoryCache == nil {
		c.memoryCache = cache.NewMemoryCache()
	}
	return c.memoryCache
}

// CacheBackend returns the cache backend with fallback support
func (c *Container) CacheBackend() cache.CacheBackend {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cacheBackend == nil {
		c.mu.Unlock()
		memCache := c.MemoryCache()
		redisCache := c.RedisCache()
		c.mu.Lock()

		if redisCache != nil && c.config.Redis.EnableFallback {
			// Use fallback cache with Redis as primary
			logger := c.logger
			if logger == nil {
				logger = zap.NewNop()
			}
			c.cacheBackend = cache.NewFallbackCache(redisCache, memCache, logger)
		} else if redisCache != nil {
			// Use Redis only
			c.cacheBackend = redisCache
		} else {
			// Use memory cache only
			c.cacheBackend = memCache
		}
	}
	return c.cacheBackend
}

// TokenBlacklist returns the token blacklist singleton
func (c *Container) TokenBlacklist() service.TokenBlacklist {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tokenBlacklist == nil {
		c.mu.Unlock()
		cacheBackend := c.CacheBackend()
		c.mu.Lock()
		c.tokenBlacklist = service.NewRedisTokenBlacklist(cacheBackend)
	}
	return c.tokenBlacklist
}

// RateLimiter returns the rate limiter singleton
func (c *Container) RateLimiter() *middleware.RedisRateLimiter {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rateLimiter == nil {
		c.mu.Unlock()
		cacheBackend := c.CacheBackend()
		c.mu.Lock()
		// Default: 100 requests per minute
		c.rateLimiter = middleware.NewRedisRateLimiter(cacheBackend, 100, time.Minute)
	}
	return c.rateLimiter
}

// ========== Mail Getters ==========

// MailClient returns the mail client singleton
func (c *Container) MailClient() *mail.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.mailClient == nil && c.config.Mail.Enabled {
		c.mailClient = mail.NewClient(&mail.Config{
			Host:     c.config.Mail.Host,
			Port:     c.config.Mail.Port,
			User:     c.config.Mail.User,
			Password: c.config.Mail.Password,
			From:     c.config.Mail.From,
			UseTLS:   c.config.Mail.UseTLS,
			MockSend: c.config.Mail.MockSend,
		})
	}
	return c.mailClient
}

// VerificationCodeService returns the verification code service singleton
func (c *Container) VerificationCodeService() *service.VerificationCodeService {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifyService == nil {
		c.mu.Unlock()
		cacheBackend := c.CacheBackend()
		mailClient := c.MailClient()
		c.mu.Lock()
		c.verifyService = service.NewVerificationCodeService(cacheBackend, mailClient, c.config.App.Name)
	}
	return c.verifyService
}

// VerificationHandler returns the verification handler singleton
func (c *Container) VerificationHandler() *handler.VerificationHandler {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifyHandler == nil {
		c.mu.Unlock()
		verifyService := c.VerificationCodeService()
		c.mu.Lock()
		c.verifyHandler = handler.NewVerificationHandler(verifyService)
	}
	return c.verifyHandler
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
