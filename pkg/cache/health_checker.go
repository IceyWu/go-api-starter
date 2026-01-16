package cache

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a cache backend
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus `json:"status"`
	Latency   int64        `json:"latency_ms"`
	Message   string       `json:"message,omitempty"`
	IsDegraded bool        `json:"is_degraded,omitempty"`
}

// HealthChecker provides health check functionality for cache backends
type HealthChecker struct {
	cache CacheBackend
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cache CacheBackend) *HealthChecker {
	return &HealthChecker{
		cache: cache,
	}
}

// Check performs a health check on the cache backend
func (h *HealthChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	err := h.cache.Ping(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return HealthCheckResult{
			Status:  HealthStatusUnhealthy,
			Latency: latency,
			Message: err.Error(),
		}
	}

	// Check if running in degraded mode (fallback cache)
	if fc, ok := h.cache.(*FallbackCache); ok && fc.IsDegraded() {
		return HealthCheckResult{
			Status:     HealthStatusDegraded,
			Latency:    latency,
			Message:    "Running in fallback mode (memory cache)",
			IsDegraded: true,
		}
	}

	return HealthCheckResult{
		Status:  HealthStatusHealthy,
		Latency: latency,
	}
}

// IsHealthy returns true if the cache is healthy
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	result := h.Check(ctx)
	return result.Status == HealthStatusHealthy
}

// IsAvailable returns true if the cache is available (healthy or degraded)
func (h *HealthChecker) IsAvailable(ctx context.Context) bool {
	result := h.Check(ctx)
	return result.Status != HealthStatusUnhealthy
}
