package middleware

import httpx "go-api-starter/internal/transport/httpx"

// SecurityHeaders adds common security headers to all responses.
func SecurityHeaders() httpx.HandlerFunc {
	return func(c *httpx.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
