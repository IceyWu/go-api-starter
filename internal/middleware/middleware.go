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

		if logger.Log != nil {
			// Structured logging for production-friendly log collection
			logger.Log.Infow("request",
				"method", method,
				"path", path,
				"status", status,
				"latency_ms", latency.Milliseconds(),
				"client_ip", c.ClientIP(),
				"request_id", GetRequestID(c),
			)
		}
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
		c.JSON(500, response.ErrorResponse{
			Code:    500,
			Message: "internal server error",
			Error:   "panic recovered",
		})
	})
}
