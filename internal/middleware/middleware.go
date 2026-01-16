package middleware

import (
	"time"

	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequestID returns a request ID middleware
// It reads X-Request-ID from header if present, otherwise generates a new UUID
func RequestID() gin.HandlerFunc {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

// Logger returns a logging middleware using zap
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Format: METHOD PATH STATUS LATENCY
		statusColor := "\033[32m" // green
		if status >= 400 {
			statusColor = "\033[33m" // yellow
		}
		if status >= 500 {
			statusColor = "\033[31m" // red
		}

		logger.Log.Infof("%s%-7s\033[0m %s %s%d\033[0m %v",
			"\033[36m", method, path, statusColor, status, latency)
	}
}

// Recovery returns a recovery middleware that handles panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)
		logger.Log.Errorw("Panic recovered",
			"request_id", requestID,
			"error", recovered,
			"path", c.Request.URL.Path,
		)
		response.InternalError(c, "internal server error")
	})
}
