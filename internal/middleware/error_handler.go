package middleware

import (
	"errors"

	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a middleware that handles errors set in context
// It should be placed early in the middleware chain to catch all errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors set in context
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError processes the error and sends appropriate HTTP response
func handleError(c *gin.Context, err error) {
	reqID := GetRequestID(c)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		// Log application error with details (if logger is initialized)
		if logger.Log != nil {
			logger.Log.Errorw("Application error",
				"request_id", reqID,
				"code", appErr.Code,
				"message", appErr.Message,
				"http_status", appErr.HTTPStatus,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"error", appErr.Err,
			)
		}

		// Send structured error response
		c.JSON(appErr.HTTPStatus, response.Response{
			Code:    appErr.HTTPStatus,
			Message: appErr.Message,
			Data:    appErr.Details,
		})
		return
	}

	// Unknown error - log and return generic 500
	if logger.Log != nil {
		logger.Log.Errorw("Unexpected error",
			"request_id", reqID,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"error", err,
		)
	}
	response.InternalError(c, "internal server error")
}
