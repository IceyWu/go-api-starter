package handler

import (
	"github.com/gin-gonic/gin"
	"go-api-starter/pkg/response"
)

type TestHandler struct{}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// TestUserCreate godoc
// @Summary 测试 user.create 权限
// @Description 测试需要 user.create 权限的端点
// @Tags 权限测试
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/test/user-create [get]
func (h *TestHandler) TestUserCreate(c *gin.Context) {
	response.Success(c, gin.H{
		"message":    "你有 user.create 权限！",
		"permission": "user.create",
	})
}

// TestUserRead godoc
// @Summary 测试 user.read 权限
// @Description 测试需要 user.read 权限的端点
// @Tags 权限测试
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/v1/test/user-read [get]
func (h *TestHandler) TestUserRead(c *gin.Context) {
	response.Success(c, gin.H{
		"message":    "你有 user.read 权限！",
		"permission": "user.read",
	})
}

// TestNoPermission godoc
// @Summary 测试无权限要求的端点
// @Description 测试不需要任何特定权限的端点，只需要登录
// @Tags 权限测试
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/v1/test/no-permission [get]
func (h *TestHandler) TestNoPermission(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	response.Success(c, gin.H{
		"message": "这个接口不需要特定权限，只需要登录",
		"user_id": userIDVal,
	})
}
