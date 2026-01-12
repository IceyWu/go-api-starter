package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents unified API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a success response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Created sends a created response with data
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "created",
		Data:    data,
	})
}

// NoContent sends a no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest sends a 400 error response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound sends a 404 error response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 error response
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}


// Unauthorized sends a 401 error response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 error response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// Conflict sends a 409 error response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// UnprocessableEntity sends a 422 error response
func UnprocessableEntity(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message)
}
