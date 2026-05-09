package handler

import (
	"strconv"

	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetUserID extracts the authenticated user ID from gin context.
// Returns 0 and sets an error if not authenticated.
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.Unauthorized("user not authenticated"))
		return 0, false
	}
	return userID.(uint), true
}

// GetOptionalUserID extracts user ID from context if present, returns 0 if not.
// Does not set any error — use for endpoints with optional auth.
func GetOptionalUserID(c *gin.Context) uint {
	if uid, exists := c.Get("userID"); exists {
		return uid.(uint)
	}
	return 0
}

// GetSecUID extracts sec_uid path parameter.
// Returns empty string and sets an error if missing.
func GetSecUID(c *gin.Context) (string, bool) {
	secUID := c.Param("sec_uid")
	if secUID == "" {
		c.Error(apperrors.BadRequest("invalid sec_uid"))
		return "", false
	}
	return secUID, true
}

// GetIDParam extracts and parses a uint path parameter by name.
// Returns 0 and sets an error if invalid.
func GetIDParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		c.Error(apperrors.BadRequest("invalid " + name))
		return 0, false
	}
	return uint(id), true
}

// BindPagination binds pagination query parameters.
// Returns nil and sets an error if binding fails.
func BindPagination(c *gin.Context) (*response.Pagination, bool) {
	var p response.Pagination
	if err := c.ShouldBindQuery(&p); err != nil {
		c.Error(apperrors.BadRequest("invalid pagination params"))
		return nil, false
	}
	return &p, true
}
