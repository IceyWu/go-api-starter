package middleware

import (
	"net/http"

	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultBodyLimit is the default max request body size (10MB)
	DefaultBodyLimit int64 = 10 << 20
)

// BodyLimit returns a middleware that limits the request body size.
// If the body exceeds maxBytes, the request is rejected with 413 Payload Too Large.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = DefaultBodyLimit
	}

	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.JSON(http.StatusRequestEntityTooLarge, response.ErrorResponse{
				Code:    http.StatusRequestEntityTooLarge,
				Message: "request body too large",
			})
			c.Abort()
			return
		}

		// Wrap the body with http.MaxBytesReader to enforce the limit
		// even when Content-Length is missing or spoofed
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
