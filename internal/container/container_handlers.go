package container

import (
	"go-api-starter/internal/handler"
)

// ========== Handler Getters ==========

func (c *Container) AuthHandler() *handler.AuthHandler {
	c.authHandlerOnce.Do(func() {
		c.authHandler = handler.NewAuthHandler(c.AuthService())
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
