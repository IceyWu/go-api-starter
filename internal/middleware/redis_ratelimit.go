package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go-api-starter/pkg/cache"
)

const (
	rateLimitKeyPrefix = "ratelimit:"
)

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	cache     cache.CacheBackend
	rate      int           // requests per window
	window    time.Duration // sliding window size
	keyPrefix string
}

// NewRedisRateLimiter creates a new Redis-backed rate limiter
func NewRedisRateLimiter(cacheBackend cache.CacheBackend, rate int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		cache:     cacheBackend,
		rate:      rate,
		window:    window,
		keyPrefix: rateLimitKeyPrefix,
	}
}

// buildKey builds a rate limit key for the given identifier
func (r *RedisRateLimiter) buildKey(identifier string) string {
	// Use window-aligned time bucket for sliding window
	bucket := time.Now().Unix() / int64(r.window.Seconds())
	return fmt.Sprintf("%s%s:%d", r.keyPrefix, identifier, bucket)
}

// Allow checks if a request is allowed and returns rate limit info
func (r *RedisRateLimiter) Allow(ctx context.Context, identifier string) (bool, RateLimitInfo, error) {
	key := r.buildKey(identifier)

	// Increment counter with expiration
	count, err := r.cache.IncrWithExpire(ctx, key, r.window)
	if err != nil {
		// On error, allow the request (fail-open)
		return true, RateLimitInfo{Limit: r.rate, Remaining: r.rate - 1}, nil
	}

	remaining := r.rate - int(count)
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time
	bucket := time.Now().Unix() / int64(r.window.Seconds())
	resetAt := time.Unix((bucket+1)*int64(r.window.Seconds()), 0)

	info := RateLimitInfo{
		Limit:     r.rate,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	return count <= int64(r.rate), info, nil
}


// RateLimit returns a Gin middleware for rate limiting
func (r *RedisRateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as identifier
		identifier := c.ClientIP()

		allowed, info, _ := r.Allow(c.Request.Context(), identifier)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetAt.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(info.ResetAt).Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

// RateLimitByUser returns a middleware that rate limits by user ID
func (r *RedisRateLimiter) RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("userID")
		var identifier string
		if exists {
			identifier = fmt.Sprintf("user:%d", userID.(uint))
		} else {
			identifier = c.ClientIP()
		}

		allowed, info, _ := r.Allow(c.Request.Context(), identifier)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetAt.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(info.ResetAt).Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

// RateLimitByEndpoint returns a middleware that rate limits by endpoint
func (r *RedisRateLimiter) RateLimitByEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Combine IP and endpoint for identifier
		identifier := fmt.Sprintf("%s:%s:%s", c.ClientIP(), c.Request.Method, c.FullPath())

		allowed, info, _ := r.Allow(c.Request.Context(), identifier)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetAt.Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(info.ResetAt).Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Rate   int           // requests per window
	Window time.Duration // window duration
}

// MultiLevelRateLimiter supports multiple rate limit levels
type MultiLevelRateLimiter struct {
	cache    cache.CacheBackend
	global   *RateLimitConfig
	endpoint map[string]*RateLimitConfig
	user     *RateLimitConfig
}

// NewMultiLevelRateLimiter creates a new multi-level rate limiter
func NewMultiLevelRateLimiter(cacheBackend cache.CacheBackend) *MultiLevelRateLimiter {
	return &MultiLevelRateLimiter{
		cache:    cacheBackend,
		endpoint: make(map[string]*RateLimitConfig),
	}
}

// SetGlobalLimit sets the global rate limit
func (m *MultiLevelRateLimiter) SetGlobalLimit(rate int, window time.Duration) *MultiLevelRateLimiter {
	m.global = &RateLimitConfig{Rate: rate, Window: window}
	return m
}

// SetEndpointLimit sets rate limit for a specific endpoint
func (m *MultiLevelRateLimiter) SetEndpointLimit(endpoint string, rate int, window time.Duration) *MultiLevelRateLimiter {
	m.endpoint[endpoint] = &RateLimitConfig{Rate: rate, Window: window}
	return m
}

// SetUserLimit sets the per-user rate limit
func (m *MultiLevelRateLimiter) SetUserLimit(rate int, window time.Duration) *MultiLevelRateLimiter {
	m.user = &RateLimitConfig{Rate: rate, Window: window}
	return m
}

// RateLimit returns a middleware that applies multi-level rate limiting
func (m *MultiLevelRateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check endpoint-specific limit first
		endpoint := c.FullPath()
		if cfg, ok := m.endpoint[endpoint]; ok {
			limiter := NewRedisRateLimiter(m.cache, cfg.Rate, cfg.Window)
			identifier := fmt.Sprintf("endpoint:%s:%s", c.ClientIP(), endpoint)
			if allowed, info, _ := limiter.Allow(c.Request.Context(), identifier); !allowed {
				setRateLimitHeaders(c, info)
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"code":    http.StatusTooManyRequests,
					"message": "请求过于频繁，请稍后再试",
				})
				return
			}
		}

		// Check user limit if authenticated
		if m.user != nil {
			if userID, exists := c.Get("userID"); exists {
				limiter := NewRedisRateLimiter(m.cache, m.user.Rate, m.user.Window)
				identifier := fmt.Sprintf("user:%d", userID.(uint))
				if allowed, info, _ := limiter.Allow(c.Request.Context(), identifier); !allowed {
					setRateLimitHeaders(c, info)
					c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"code":    http.StatusTooManyRequests,
						"message": "请求过于频繁，请稍后再试",
					})
					return
				}
			}
		}

		// Check global limit
		if m.global != nil {
			limiter := NewRedisRateLimiter(m.cache, m.global.Rate, m.global.Window)
			identifier := c.ClientIP()
			allowed, info, _ := limiter.Allow(c.Request.Context(), identifier)
			setRateLimitHeaders(c, info)
			if !allowed {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"code":    http.StatusTooManyRequests,
					"message": "请求过于频繁，请稍后再试",
				})
				return
			}
		}

		c.Next()
	}
}

func setRateLimitHeaders(c *gin.Context, info RateLimitInfo) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetAt.Unix(), 10))
}
