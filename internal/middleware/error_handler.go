package middleware

import (
	"errors"

	"go-api-starter/internal/config"
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
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if logger.Log != nil && appErr.HTTPStatus >= 500 {
			logger.Log.Errorf("%s %s - %s", c.Request.Method, c.Request.URL.Path, appErr.Message)
		}

		resp := response.ErrorResponse{
			Code:      appErr.HTTPStatus,
			ErrorCode: appErr.Code,
			Message:   appErr.Message,
			Details:   appErr.Details,
		}
		// Only expose internal error details in non-production environments
		if !isProduction() {
			resp.Error = appErr.Error()
		}
		c.JSON(appErr.HTTPStatus, resp)
		return
	}

	// Unknown error - log and return generic 500
	if logger.Log != nil {
		logger.Log.Errorf("%s %s - %v", c.Request.Method, c.Request.URL.Path, err)
	}
	resp := response.ErrorResponse{
		Code:    500,
		Message: "internal server error",
	}
	if !isProduction() {
		resp.Error = err.Error()
	}
	c.JSON(500, resp)
}

func isProduction() bool {
	cfg := config.GetConfig()
	if cfg == nil {
		return false
	}
	return cfg.App.Env == "production" || cfg.App.Env == "prod"
}
