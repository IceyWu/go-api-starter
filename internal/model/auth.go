package model

import "strings"

// LoginRequest represents the login request
type LoginRequest struct {
	Account   string  `json:"account" binding:"required" example:"admin@example.com"` // 账号：邮箱或手机号，自动识别
	Password  string  `json:"password" binding:"omitempty,min=6" example:"password123"`
	Code      string  `json:"code" binding:"omitempty,len=6" example:"123456"`         // 验证码登录时使用
	LoginType string  `json:"login_type" binding:"omitempty" example:"code"`           // "code" 表示验证码登录，空或其他表示密码登录
	Mobile    *string `json:"-"` // 内部使用，由 ResolveAccount 填充
	Email     *string `json:"-"` // 内部使用，由 ResolveAccount 填充
}

// ResolveAccount 将 account 字段解析到 email 或 mobile
func (r *LoginRequest) ResolveAccount() {
	if r.Account != "" {
		if strings.Contains(r.Account, "@") {
			r.Email = &r.Account
		} else {
			r.Mobile = &r.Account
		}
	}
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Mobile   *string `json:"mobile" binding:"omitempty,len=11" example:"13800138000"`
	Email    *string `json:"email" binding:"omitempty,email" example:"admin@example.com"`
	Password string  `json:"password" binding:"required,min=6" example:"password123"`
	Code     string  `json:"code" binding:"required,len=6" example:"123456"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string        `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string        `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64         `json:"expires_in" example:"86400"`    // access_token 过期时间（秒）
	User         *UserResponse `json:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn   int64  `json:"expires_in" example:"86400"` // access_token 过期时间（秒）
}

// ResetPasswordRequest represents the reset password request (admin only)
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// SelfResetPasswordRequest represents the self-service password reset request
type SelfResetPasswordRequest struct {
	Account     string `json:"account" binding:"required" example:"john@example.com"`       // 账号：邮箱或手机号
	Code        string `json:"code" binding:"required,len=6" example:"123456"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}


