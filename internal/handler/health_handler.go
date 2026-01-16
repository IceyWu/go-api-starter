package handler

import (
	"context"
	"net/http"
	"time"

	"go-api-starter/pkg/cache"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db           *gorm.DB
	startTime    time.Time
	version      string
	cacheChecker *cache.HealthChecker
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// NewHealthHandlerWithCache creates a new HealthHandler with cache health checking
func NewHealthHandlerWithCache(db *gorm.DB, version string, cacheBackend cache.CacheBackend) *HealthHandler {
	var checker *cache.HealthChecker
	if cacheBackend != nil {
		checker = cache.NewHealthChecker(cacheBackend)
	}
	return &HealthHandler{
		db:           db,
		startTime:    time.Now(),
		version:      version,
		cacheChecker: checker,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// ReadinessResponse represents the readiness check response
type ReadinessResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// Health godoc
// @Summary 健康检查
// @Description 获取服务健康状态
// @Tags 健康检查
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, HealthResponse{
		Status:    "ok",
		Version:   h.version,
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
	})
}

// Ready godoc
// @Summary 就绪检查
// @Description 检查服务是否准备好接受请求
// @Tags 健康检查
// @Produce json
// @Success 200 {object} ReadinessResponse "服务已就绪"
// @Failure 503 {object} response.Response "服务未就绪"
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]string)
	allHealthy := true
	
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else if err := sqlDB.Ping(); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "healthy"
	}
	
	// Check Redis/cache connection
	if h.cacheChecker != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		
		result := h.cacheChecker.Check(ctx)
		switch result.Status {
		case cache.HealthStatusHealthy:
			checks["redis"] = "healthy"
		case cache.HealthStatusDegraded:
			checks["redis"] = "degraded: " + result.Message
			// Degraded is still considered ready
		case cache.HealthStatusUnhealthy:
			checks["redis"] = "unhealthy: " + result.Message
			// Redis unhealthy doesn't make service not ready if fallback is enabled
		}
	}
	
	if !allHealthy {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Code:    http.StatusServiceUnavailable,
			Message: "service not ready",
			Data: ReadinessResponse{
				Status: "not_ready",
				Checks: checks,
			},
		})
		return
	}
	
	response.Success(c, ReadinessResponse{
		Status: "ready",
		Checks: checks,
	})
}
