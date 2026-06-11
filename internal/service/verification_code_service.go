package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/cache"
	"go-api-starter/pkg/i18n"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/mail"
)

const (
	VerificationCodeLength          = 6
	VerificationCodeExpiry          = 60 * time.Second
	VerificationCodePrefix          = "verify_code:"
	VerificationCodeRateLimit       = "verify_rate:"
	VerificationCodeRateLimitExpiry = 60 * time.Second
	VerifiedEmailPrefix             = "verified_email:"
	VerifiedEmailExpiry             = 10 * time.Minute
)

// VerificationCodeService handles verification code operations
type VerificationCodeService struct {
	cache      cache.CacheBackend
	mailClient *mail.Client
	appName    string
}

// NewVerificationCodeService creates a new verification code service
func NewVerificationCodeService(cache cache.CacheBackend, mailClient *mail.Client, appName string) *VerificationCodeService {
	if appName == "" {
		appName = "Go API Starter"
	}
	return &VerificationCodeService{
		cache:      cache,
		mailClient: mailClient,
		appName:    appName,
	}
}

// GenerateCode generates a random verification code
func (s *VerificationCodeService) GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < VerificationCodeLength; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

// SendVerificationCode sends a verification code to the email
func (s *VerificationCodeService) SendVerificationCode(ctx context.Context, email, purpose string) error {
	rateLimitKey := VerificationCodeRateLimit + email
	_, err := s.cache.Get(ctx, rateLimitKey)
	if err == nil {
		return apperrors.BadRequestCode(i18n.ErrCodeRateLimit)
	}

	code := s.GenerateCode()

	logger.Infof("\033[35m[验证码]\033[0m %s -> %s (%s)", email, code, purpose)

	codeKey := VerificationCodePrefix + purpose + ":" + email
	if err := s.cache.Set(ctx, codeKey, []byte(code), VerificationCodeExpiry); err != nil {
		return apperrors.WrapCode(err, i18n.ErrCodeStoreFailed)
	}

	if err := s.cache.Set(ctx, rateLimitKey, []byte("1"), VerificationCodeRateLimitExpiry); err != nil {
		// Log but don't fail
		_ = err
	}

	subject := fmt.Sprintf("【%s】验证码", s.appName)
	body := s.buildEmailBody(code, purpose)

	if err := s.mailClient.SendHTMLMail([]string{email}, subject, body, true); err != nil {
		s.cache.Delete(ctx, codeKey)
		return apperrors.WrapCode(err, i18n.ErrMailSendFailed)
	}

	return nil
}

// VerifyCode verifies the verification code
func (s *VerificationCodeService) VerifyCode(ctx context.Context, email, purpose, code string) (bool, error) {
	codeKey := VerificationCodePrefix + purpose + ":" + email
	storedCode, err := s.cache.Get(ctx, codeKey)
	if err != nil {
		return false, apperrors.BadRequestCode(i18n.ErrCodeExpired)
	}

	if string(storedCode) != code {
		return false, apperrors.BadRequestCode(i18n.ErrCodeInvalid)
	}

	s.cache.Delete(ctx, codeKey)

	if purpose == "register" {
		verifiedKey := VerifiedEmailPrefix + email
		s.cache.Set(ctx, verifiedKey, []byte("1"), VerifiedEmailExpiry)
	}

	return true, nil
}

// IsEmailVerified checks if the email has been verified for registration
func (s *VerificationCodeService) IsEmailVerified(ctx context.Context, email string) bool {
	verifiedKey := VerifiedEmailPrefix + email
	_, err := s.cache.Get(ctx, verifiedKey)
	return err == nil
}

// ClearEmailVerified clears the verified status after successful registration
func (s *VerificationCodeService) ClearEmailVerified(ctx context.Context, email string) {
	verifiedKey := VerifiedEmailPrefix + email
	s.cache.Delete(ctx, verifiedKey)
}

// buildEmailBody builds the HTML email body
func (s *VerificationCodeService) buildEmailBody(code, purpose string) string {
	purposeText := "操作"
	switch purpose {
	case "register":
		purposeText = "注册账号"
	case "login":
		purposeText = "登录账号"
	case "reset_password":
		purposeText = "重置密码"
	case "bind_email":
		purposeText = "绑定邮箱"
	}

	const tpl = `<!doctype html>
<html lang="zh-CN">
<head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width,initial-scale=1.0"/><title>{{APP_NAME}} - 验证码</title></head>
<body style="margin:0;padding:0;background:#fff;background-image:radial-gradient(#d5d5d5 1px,transparent 1px);background-size:20px 20px;font-family:'Comic Sans MS','Chalkboard SE','Marker Felt','PingFang SC',sans-serif;color:#000;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0">
<tr><td align="center" style="padding:60px 16px;">

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:-15px;">
<tr><td align="center">
<div style="width:140px;height:35px;background:#fff;border:2px solid #000;border-left:4px dashed #000;border-right:4px dashed #000;transform:rotate(-3deg);position:relative;z-index:10;box-shadow:4px 4px 0px #000;"></div>
</td></tr></table>

<div style="display:inline-block;position:relative;">
<div style="position:absolute;top:2px;left:2px;width:100%;height:100%;background:#000;border-radius:20px 4px 25px 2px/4px 25px 2px 20px;transform:rotate(1deg);"></div>
<table role="presentation" width="440" cellpadding="0" cellspacing="0" style="position:relative;z-index:2;background:#fff;background-image:repeating-linear-gradient(transparent,transparent 31px,#e0e0e0 31px,#e0e0e0 32px);border:3px solid #000;border-radius:2px 25px 4px 20px/25px 4px 20px 2px;text-align:left;transform:rotate(-0.5deg);">

<tr><td style="padding:50px 40px 10px;">

<div style="display:inline-block;margin-bottom:24px;font-size:15px;font-weight:900;letter-spacing:2px;background:#000;color:#fff;padding:4px 16px;transform:rotate(-3deg) skew(-5deg);border-radius:255px 15px 225px 15px/15px 225px 15px 255px;border:2px solid #000;">URGENT</div>

<p style="margin:0 0 20px;font-size:16px;line-height:2.1;font-weight:bold;">
你好，你正在进行 <strong style="border-bottom:3px dotted #000;">{{PURPOSE}}</strong> 操作。<br/>
请收好这张写着暗号的小纸条：
</p>

<div style="font-family:'Comic Sans MS',monospace;font-weight:bold;font-size:22px;text-align:center;margin-bottom:-10px;margin-top:15px;transform:rotate(8deg);color:#000;letter-spacing:2px;">
/ /<br/>\ /<br/>V
</div>

<div style="background:#000;border:4px solid #111;border-radius:10px 255px 10px 25px/255px 15px 225px 10px;padding:24px 0;text-align:center;margin-bottom:30px;transform:rotate(-2deg);position:relative;">
<span style="font-size:42px;font-weight:900;letter-spacing:12px;color:#fff;font-family:Courier,'Courier New',monospace;text-shadow:1px 1px 0px #aaa,-1px -1px 0px #666;">~{{CODE}}~</span>
</div>

<ul style="margin:0;padding:0;list-style:none;font-size:15px;line-height:2.1;font-weight:bold;">
<li style="margin-bottom:15px;transform:rotate(0.5deg);">
<span style="font-family:Courier,monospace;color:#000;font-weight:900;border:2px solid #000;border-radius:50%;padding:0 4px;display:inline-block;transform:rotate(-10deg);">!</span>
保鲜期：<span style="border-bottom:3px double #000;padding-bottom:2px;">只限 <strong>{{EXPIRY}} 秒</strong> 内食用</span>
</li>
<li style="transform:rotate(-0.5deg);">
<span style="font-family:Courier,monospace;opacity:0.7;border:2px solid #555;border-radius:50%;padding:0 4px;display:inline-block;transform:rotate(10deg);">?</span>
非本人：<span style="text-decoration:line-through;opacity:0.6;">赶紧跑</span> 把它<strong style="text-decoration:underline solid #000;">当废纸扔了</strong>吧。
</li>
</ul>

</td></tr>

<tr><td style="padding:10px 40px 40px;text-align:right;">
<div style="display:inline-block;font-size:15px;font-weight:bold;transform:rotate(-12deg);border:3px solid #000;padding:8px 16px;background-color:transparent;text-align:center;border-radius:2px 25px 4px 10px/25px 2px 20px 2px;position:relative;">
<div style="font-family:'Courier New',monospace;font-size:16px;margin-bottom:4px;font-weight:900;color:#000;text-transform:uppercase;">[ {{APP_NAME}} ]</div>
<div style="text-decoration:line-through;font-size:13px;color:#555;transform:rotate(2deg);">Do not reply</div>
</div>
</td></tr>

</table>
</div>

</td></tr></table>
</body></html>`

	expiry := fmt.Sprintf("%d", int(VerificationCodeExpiry.Seconds()))

	r := strings.NewReplacer(
		"{{PURPOSE}}", purposeText,
		"{{CODE}}", code,
		"{{APP_NAME}}", s.appName,
		"{{EXPIRY}}", expiry,
	)
	return r.Replace(tpl)
}
