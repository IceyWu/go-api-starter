package handler

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/i18n"
	"go-api-starter/pkg/response"
)

type AuthHandler struct {
	authService   service.AuthServiceInterface
	userService   service.UserServiceInterface
	verifyService *service.VerificationCodeService
	wechatService *service.WechatService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService service.AuthServiceInterface, userService service.UserServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// SetWechatService sets the wechat service (optional dependency)
func (h *AuthHandler) SetWechatService(ws *service.WechatService) {
	h.wechatService = ws
}

// SetVerifyService sets the verification code service (optional dependency)
func (h *AuthHandler) SetVerifyService(vs *service.VerificationCodeService) {
	h.verifyService = vs
}

// Register godoc
// @Summary 注册新用户
// @Description 注册一个新的用户账号（需要邮箱或手机号验证码），注册成功后自动返回登录令牌
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

	// Verify the code
	if h.verifyService != nil {
		var identifier string
		if req.Mobile != nil {
			identifier = *req.Mobile
		} else if req.Email != nil {
			identifier = *req.Email
		}

		valid, err := h.verifyService.VerifyCode(ctx, identifier, "register", req.Code)
		if err != nil || !valid {
			c.Error(apperrors.BadRequestCode(i18n.ErrCodeExpired))
			return
		}
	}

	loginResp, err := h.authService.Register(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Created(c, loginResp)
}

// Login godoc
// @Summary 用户登录
// @Description 使用手机号或邮箱和密码登录，或使用验证码登录（login_type=code）
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

	ctx := c.Request.Context()

	// 验证码登录
	if req.LoginType == "code" {
		if req.Code == "" {
			c.Error(apperrors.BadRequestCode(i18n.ErrCodeRequired))
			return
		}
		// 校验验证码
		if h.verifyService != nil {
			valid, err := h.verifyService.VerifyCode(ctx, req.Account, "login", req.Code)
			if err != nil || !valid {
				c.Error(apperrors.BadRequestCode(i18n.ErrCodeExpired))
				return
			}
		}
		// 通过验证码登录（不需要密码）
		loginResp, err := h.authService.LoginByCode(ctx, &req)
		if err != nil {
			c.Error(err)
			return
		}
		response.Success(c, loginResp)
		return
	}

	// 密码登录
	if req.Password == "" {
		c.Error(apperrors.BadRequestCode(i18n.ErrPasswordRequired))
		return
	}

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
// @Param uid path string true "用户 UID"
// @Param request body model.ResetPasswordRequest true "重置密码请求数据"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/auth/reset-password/{uid} [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		c.Error(apperrors.BadRequestCode(i18n.ErrInvalidUserID))
		return
	}

	ctx := c.Request.Context()
	user, err := h.userService.GetByUID(ctx, uid)
	if err != nil {
		c.Error(err)
		return
	}

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	if err := h.authService.ResetPassword(ctx, user.ID, &req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}

// SelfResetPassword godoc
// @Summary 用户自助重置密码
// @Description 用户通过邮箱验证码重置自己的密码（无需登录）
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.SelfResetPasswordRequest true "自助重置密码请求数据"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/auth/self-reset-password [post]
func (h *AuthHandler) SelfResetPassword(c *gin.Context) {
	var req model.SelfResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// 校验验证码
	if h.verifyService != nil {
		valid, err := h.verifyService.VerifyCode(ctx, req.Account, "reset_password", req.Code)
		if err != nil || !valid {
			c.Error(apperrors.BadRequestCode(i18n.ErrCodeExpired))
			return
		}
	}

	if err := h.authService.SelfResetPassword(ctx, &req); err != nil {
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

// WxLogin godoc
// @Summary 微信小程序登录
// @Description 微信一键登录/注册接口。传入 js_code + phone_code，自动完成 openid 获取、手机号解析、用户查找或创建、签发 JWT，一步到位
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.WxLoginRequest true "微信登录请求"
// @Success 200 {object} response.Response{data=model.WxLoginResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response "手机号已绑定其他微信号"
// @Failure 500 {object} response.Response
// @Router /api/v1/auth/wx-login [post]
func (h *AuthHandler) WxLogin(c *gin.Context) {
	if h.wechatService == nil {
		c.Error(apperrors.InternalCode(nil, i18n.ErrWechatNotConfigured))
		return
	}

	var req model.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()
	loginResp, err := h.wechatService.Login(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, loginResp)
}
