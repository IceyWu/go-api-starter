package container

import (
	"go-api-starter/internal/handler"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/mail"
)

// ========== Handler Getters ==========

func (c *Container) AuthHandler() *handler.AuthHandler {
	c.authHandlerOnce.Do(func() {
		c.authHandler = handler.NewAuthHandler(c.AuthService(), c.UserService())
		// Wire up optional WechatService
		c.authHandler.SetWechatService(c.WechatLoginService())
		// Wire up optional VerificationCodeService
		c.authHandler.SetVerifyService(c.VerificationCodeService())
	})
	return c.authHandler
}

func (c *Container) UserHandler() *handler.UserHandler {
	c.userHandlerOnce.Do(func() {
		c.userHandler = handler.NewUserHandler(c.UserService())
	})
	return c.userHandler
}

func (c *Container) PermissionHandler() *handler.PermissionHandler {
	c.permHandlerOnce.Do(func() {
		c.permHandler = handler.NewPermissionHandler(
			c.PermissionService(), c.UserService(),
		)
	})
	return c.permHandler
}

func (c *Container) OSSHandler() *handler.OSSHandler {
	c.ossHandlerOnce.Do(func() {
		c.ossHandler = handler.NewOSSHandler(c.OSSService(), c.UserService())
	})
	return c.ossHandler
}

func (c *Container) HealthHandler() *handler.HealthHandler {
	c.healthHandlerOnce.Do(func() {
		c.healthHandler = handler.NewHealthHandler(c.db, "1.0.0", c.CacheBackend())
	})
	return c.healthHandler
}

func (c *Container) VerificationHandler() *handler.VerificationHandler {
	c.verifyHandlerOnce.Do(func() {
		c.verifyHandler = handler.NewVerificationHandler(c.VerificationCodeService())
	})
	return c.verifyHandler
}

// ========== Mail ==========

func (c *Container) MailClient() *mail.Client {
	c.mailClientOnce.Do(func() {
		if c.config.Mail.Enabled {
			c.mailClient = mail.NewClient(&mail.Config{
				Host:     c.config.Mail.Host,
				Port:     c.config.Mail.Port,
				User:     c.config.Mail.User,
				Password: c.config.Mail.Password,
				From:     c.config.Mail.From,
				UseTLS:   c.config.Mail.UseTLS,
				MockSend: c.config.Mail.MockSend,
			})
		}
	})
	return c.mailClient
}

func (c *Container) VerificationCodeService() *service.VerificationCodeService {
	c.verifyServiceOnce.Do(func() {
		c.verifyService = service.NewVerificationCodeService(
			c.CacheBackend(), c.MailClient(), c.config.App.Name,
		)
	})
	return c.verifyService
}

// WechatLoginService returns the WeChat mini-program login service (singleton).
func (c *Container) WechatLoginService() *service.WechatService {
	c.wechatServiceOnce.Do(func() {
		c.wechatService = service.NewWechatService(&c.config.Wechat, c.config.App.DefaultUserPassword, c.UserRepository(), c.JWTManager())
	})
	return c.wechatService
}
