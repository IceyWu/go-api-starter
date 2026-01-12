package handler

import (
	"errors"

	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// handleError handles application errors and sends appropriate HTTP responses
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	requestID := requestid.Get(c)
	
	// Check if it's an AppError
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		// Log the error with context
		logger.Log.Errorw("Application error",
			"request_id", requestID,
			"code", appErr.Code,
			"message", appErr.Message,
			"http_status", appErr.HTTPStatus,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"error", appErr.Err,
		)
		
		// Send error response
		if appErr.Details != nil {
			c.JSON(appErr.HTTPStatus, response.Response{
				Code:    appErr.HTTPStatus,
				Message: appErr.Message,
				Data:    appErr.Details,
			})
		} else {
			response.Error(c, appErr.HTTPStatus, appErr.Message)
		}
		return
	}
	
	// Unknown error - log and return 500
	logger.Log.Errorw("Unexpected error",
		"request_id", requestID,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"error", err,
	)
	response.InternalError(c, "internal server error")
}
