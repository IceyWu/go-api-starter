package middleware

import (
	"time"

	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestID returns a request ID middleware
func RequestID() gin.HandlerFunc {
	return requestid.New()
}

// Logger returns a logging middleware using zap
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		requestID := requestid.Get(c)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if query != "" {
			path = path + "?" + query
		}

		logger.Log.Infow("HTTP Request",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"latency", latency,
			"client_ip", clientIP,
			"error", errorMessage,
		)
	}
}

// Recovery returns a recovery middleware that handles panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := requestid.Get(c)
		logger.Log.Errorw("Panic recovered",
			"request_id", requestID,
			"error", recovered,
			"path", c.Request.URL.Path,
		)
		response.InternalError(c, "internal server error")
	})
}
