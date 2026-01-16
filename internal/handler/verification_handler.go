package handler

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"
)

// SendCodeRequest is the request for sending verification code
type SendCodeRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Purpose string `json:"purpose" binding:"required,oneof=register reset_password bind_email"`
}

// VerifyCodeRequest is the request for verifying code
type VerifyCodeRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Code    string `json:"code" binding:"required,len=6"`
	Purpose string `json:"purpose" binding:"required,oneof=register reset_password bind_email"`
}

// VerificationHandler handles verification code requests
type VerificationHandler struct {
	verifyService *service.VerificationCodeService
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verifyService *service.VerificationCodeService) *VerificationHandler {
	return &VerificationHandler{
		verifyService: verifyService,
	}
}

// SendCode godoc
// @Summary 发送验证码
// @Description 发送邮箱验证码
// @Tags 验证码
// @Accept json
// @Produce json
// @Param request body SendCodeRequest true "发送验证码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 429 {object} response.Response
// @Router /api/v1/verification/send [post]
func (h *VerificationHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()
	if err := h.verifyService.SendVerificationCode(ctx, req.Email, req.Purpose); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	response.Success(c, gin.H{"message": "验证码已发送"})
}

// VerifyCode godoc
// @Summary 验证验证码
// @Description 验证邮箱验证码
// @Tags 验证码
// @Accept json
// @Produce json
// @Param request body VerifyCodeRequest true "验证验证码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/verification/verify [post]
func (h *VerificationHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	ctx := c.Request.Context()
	valid, err := h.verifyService.VerifyCode(ctx, req.Email, req.Purpose, req.Code)
	if err != nil {
		c.Error(apperrors.BadRequest(err.Error()))
		return
	}

	if !valid {
		c.Error(apperrors.BadRequest("验证码错误"))
		return
	}

	response.Success(c, gin.H{"message": "验证成功", "valid": true})
}
