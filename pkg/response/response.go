package response

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents unified API response
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse represents unified API error response
type ErrorResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Details   any    `json:"details,omitempty"`
}

// jsonWithNull 自定义 JSON 编码，确保 nil 指针也会被序列化为 null
func jsonWithNull(c *gin.Context, code int, obj any) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(obj); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to encode response",
			Error:   err.Error(),
		})
		return
	}

	c.Data(code, "application/json; charset=utf-8", buf.Bytes())
}

// Success sends a success response with data
func Success(c *gin.Context, data any) {
	jsonWithNull(c, http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Created sends a created response with data
func Created(c *gin.Context, data any) {
	jsonWithNull(c, http.StatusCreated, Response{
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
	jsonWithNull(c, code, ErrorResponse{
		Code:    code,
		Message: message,
		Error:   message,
	})
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, code int, message string, errorMsg string, details any) {
	jsonWithNull(c, code, ErrorResponse{
		Code:    code,
		Message: message,
		Error:   errorMsg,
		Details: details,
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
