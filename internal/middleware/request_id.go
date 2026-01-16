package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestIDConfig holds configuration for RequestID middleware
type RequestIDConfig struct {
	// Generator is a function that generates request IDs
	// If nil, UUID v4 will be used
	Generator func() string
	// HeaderName is the header to read/write request ID
	// Default is "X-Request-ID"
	HeaderName string
}

// DefaultRequestIDConfig returns default configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Generator:  generateUUID,
		HeaderName: RequestIDHeader,
	}
}

// generateUUID generates a new UUID v4
func generateUUID() string {
	return uuid.New().String()
}

// RequestIDWithConfig returns a RequestID middleware with custom config
func RequestIDWithConfig(config RequestIDConfig) gin.HandlerFunc {
	if config.Generator == nil {
		config.Generator = generateUUID
	}
	if config.HeaderName == "" {
		config.HeaderName = RequestIDHeader
	}

	return func(c *gin.Context) {
		// Try to get request ID from header
		requestID := c.GetHeader(config.HeaderName)

		// If not present, generate a new one
		if requestID == "" {
			requestID = config.Generator()
		}

		// Set request ID in context
		c.Set(RequestIDKey, requestID)

		// Set request ID in response header
		c.Header(config.HeaderName, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

// GetRequestIDFromHeader retrieves the request ID from request header
func GetRequestIDFromHeader(c *gin.Context) string {
	return c.GetHeader(RequestIDHeader)
}
