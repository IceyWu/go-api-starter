package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"
)

type AuthHandler struct {
	authService   service.AuthServiceInterface
	verifyService *service.VerificationCodeService
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// NewAuthHandlerWithVerification creates a new AuthHandler with verification service
func NewAuthHandlerWithVerification(authService service.AuthServiceInterface, verifyService *service.VerificationCodeService) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		verifyService: verifyService,
	}
}

// Register godoc
// @Summary 注册新用户
// @Description 注册一个新的用户账号（需要邮箱验证码）
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册请求数据"
// @Success 201 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Verify the code
	if h.verifyService != nil {
		valid, err := h.verifyService.VerifyCode(ctx, req.Email, "register", req.Code)
		if err != nil || !valid {
			c.Error(apperrors.BadRequest("验证码错误或已过期"))
			return
		}
	}

	user, err := h.authService.Register(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Created(c, user)
}

// Login godoc
// @Summary 用户登录
// @Description 使用邮箱和密码登录
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

	ctx := c.Request.Context()
	loginResp, err := h.authService.Login(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, loginResp)
}

// GetCurrentUser godoc
// @Summary 获取当前用户信息
// @Description 获取当前已认证用户的详细信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.User}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.Unauthorized("用户未认证"))
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.Error(apperrors.Unauthorized("用户ID无效"))
		return
	}

	ctx := c.Request.Context()
	user, err := h.authService.GetCurrentUser(ctx, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, user)
}

// ResetPassword godoc
// @Summary 重置用户密码
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
		c.Error(apperrors.BadRequest("无效的用户ID"))
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
// @Description 使当前令牌失效
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.Error(apperrors.Unauthorized("用户未认证"))
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
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.Error(apperrors.Unauthorized("用户未认证"))
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.Error(apperrors.Unauthorized("用户ID无效"))
		return
	}

	ctx := c.Request.Context()
	if err := h.authService.LogoutAllDevices(ctx, userID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"message": "已登出所有设备"})
}
