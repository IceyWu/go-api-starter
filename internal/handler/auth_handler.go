package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/i18n"
	"go-api-starter/pkg/response"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary 注册新用户
// @Description 使用邮箱或手机号 + 密码注册一个新用户，注册成功后自动返回登录令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册请求数据"
// @Success 201 {object} response.Response{data=model.LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	if req.Mobile == nil && req.Email == nil {
		c.Error(apperrors.BadRequestCode(i18n.ErrMobileOrEmailRequired))
		return
	}

	ctx := c.Request.Context()
	loginResp, err := h.authService.Register(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Created(c, loginResp)
}

// Login godoc
// @Summary 用户登录
// @Description 使用手机号或邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求数据"
// @Success 200 {object} response.Response{data=model.LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	// 将 account 字段解析到 email 或 mobile
	req.ResolveAccount()

	if req.Password == "" {
		c.Error(apperrors.BadRequestCode(i18n.ErrPasswordRequired))
		return
	}

	ctx := c.Request.Context()
	loginResp, err := h.authService.Login(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, loginResp)
}

// RefreshToken godoc
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "刷新令牌请求数据"
// @Success 200 {object} response.Response{data=model.RefreshTokenResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()
	accessToken, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, model.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   h.authService.AccessTokenExpiresIn(),
	})
}

// ResetPassword godoc
// @Summary 重置用户密码（管理员）
// @Description 重置指定用户的密码（仅管理员）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body model.ResetPasswordRequest true "重置密码请求数据"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/auth/reset-password/{id} [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.Error(apperrors.BadRequestCode(i18n.ErrInvalidUserID))
		return
	}

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.ResetPassword(ctx, uint(userID), &req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// Logout godoc
// @Summary 用户登出
// @Description 使当前令牌失效（需要 Redis 支持）
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.Error(apperrors.UnauthorizedCode(i18n.ErrUnauthenticated))
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.Logout(ctx, token.(string)); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"message": "登出成功"})
}

// LogoutAllDevices godoc
// @Summary 登出所有设备
// @Description 使当前用户的所有令牌失效
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAllDevices(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.LogoutAllDevices(ctx, userID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"message": "已登出所有设备"})
}
