package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go-api-starter/pkg/cache"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/mail"
)

const (
	// VerificationCodeLength is the length of verification code
	VerificationCodeLength = 6
	// VerificationCodeExpiry is the expiry time of verification code (60 seconds)
	VerificationCodeExpiry = 60 * time.Second
	// VerificationCodePrefix is the cache key prefix
	VerificationCodePrefix = "verify_code:"
	// VerificationCodeRateLimit is the rate limit key prefix
	VerificationCodeRateLimit = "verify_rate:"
	// VerificationCodeRateLimitExpiry is the rate limit expiry
	VerificationCodeRateLimitExpiry = 60 * time.Second
	// VerifiedEmailPrefix is the prefix for verified email status
	VerifiedEmailPrefix = "verified_email:"
	// VerifiedEmailExpiry is the expiry time for verified email status (10 minutes)
	VerifiedEmailExpiry = 10 * time.Minute
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
	// Check rate limit
	rateLimitKey := VerificationCodeRateLimit + email
	_, err := s.cache.Get(ctx, rateLimitKey)
	if err == nil {
		return fmt.Errorf("请等待60秒后再发送验证码")
	}

	// Generate code
	code := s.GenerateCode()

	// Log email and code for debugging
	logger.Infof("\033[35m[验证码]\033[0m %s -> %s (%s)", email, code, purpose)

	// Store code in cache
	codeKey := VerificationCodePrefix + purpose + ":" + email
	if err := s.cache.Set(ctx, codeKey, []byte(code), VerificationCodeExpiry); err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// Set rate limit
	if err := s.cache.Set(ctx, rateLimitKey, []byte("1"), VerificationCodeRateLimitExpiry); err != nil {
		// Log but don't fail
	}

	// Send email
	subject := fmt.Sprintf("【%s】验证码", s.appName)
	body := s.buildEmailBody(code, purpose)

	if err := s.mailClient.SendHTMLMail([]string{email}, subject, body, true); err != nil {
		// Delete the stored code if email fails
		s.cache.Delete(ctx, codeKey)
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// VerifyCode verifies the verification code
func (s *VerificationCodeService) VerifyCode(ctx context.Context, email, purpose, code string) (bool, error) {
	codeKey := VerificationCodePrefix + purpose + ":" + email
	storedCode, err := s.cache.Get(ctx, codeKey)
	if err != nil {
		return false, fmt.Errorf("验证码已过期或不存在")
	}

	if string(storedCode) != code {
		return false, fmt.Errorf("验证码错误")
	}

	// Delete the code after successful verification
	s.cache.Delete(ctx, codeKey)

	// Mark email as verified for registration
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
	case "reset_password":
		purposeText = "重置密码"
	case "bind_email":
		purposeText = "绑定邮箱"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .header h1 { color: #333; font-size: 24px; margin: 0; }
        .code { background: #f8f9fa; border-radius: 8px; padding: 20px; text-align: center; margin: 20px 0; }
        .code span { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #007bff; }
        .info { color: #666; font-size: 14px; line-height: 1.6; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; color: #999; font-size: 12px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <p class="info">您正在进行<strong>%s</strong>操作，验证码为：</p>
        <div class="code">
            <span>%s</span>
        </div>
        <p class="info">验证码有效期为 <strong>5 分钟</strong>，请勿将验证码泄露给他人。</p>
        <p class="info">如果这不是您本人的操作，请忽略此邮件。</p>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
        </div>
    </div>
</body>
</html>
`, s.appName, purposeText, code)
}
