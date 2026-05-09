package container

import (
	"go-api-starter/internal/repository"
)

// ========== Repository Getters ==========

func (c *Container) UserRepository() repository.UserRepositoryInterface {
	c.userRepoOnce.Do(func() {
		c.userRepo = repository.NewUserRepository(c.db)
	})
	return c.userRepo
}

func (c *Container) PermissionRepository() repository.PermissionRepositoryInterface {
	c.permRepoOnce.Do(func() {
		c.permRepo = repository.NewPermissionRepository(c.db)
	})
	return c.permRepo
}

func (c *Container) RoleRepository() repository.RoleRepositoryInterface {
	c.roleRepoOnce.Do(func() {
		c.roleRepo = repository.NewRoleRepository(c.db)
	})
	return c.roleRepo
}

func (c *Container) PermissionSpaceRepository() repository.PermissionSpaceRepositoryInterface {
	c.spaceRepoOnce.Do(func() {
		c.spaceRepo = repository.NewPermissionSpaceRepository(c.db)
	})
	return c.spaceRepo
}

func (c *Container) UserRoleRepository() repository.UserRoleRepositoryInterface {
	c.userRoleRepoOnce.Do(func() {
		c.userRoleRepo = repository.NewUserRoleRepository(c.db)
	})
	return c.userRoleRepo
}

func (c *Container) RolePermissionRepository() repository.RolePermissionRepositoryInterface {
	c.rolePermRepoOnce.Do(func() {
		c.rolePermRepo = repository.NewRolePermissionRepository(c.db)
	})
	return c.rolePermRepo
}

func (c *Container) UserPermissionCacheRepository() repository.UserPermissionCacheRepositoryInterface {
	c.cacheRepoOnce.Do(func() {
		c.cacheRepo = repository.NewUserPermissionCacheRepository(c.db)
	})
	return c.cacheRepo
}

func (c *Container) MultipartRepository() repository.MultipartRepositoryInterface {
	c.multipartRepoOnce.Do(func() {
		c.multipartRepo = repository.NewMultipartRepository(c.db)
	})
	return c.multipartRepo
}

func (c *Container) FileRepository() repository.FileRepositoryInterface {
	c.fileRepoOnce.Do(func() {
		c.fileRepo = repository.NewFileRepository(c.db)
	})
	return c.fileRepo
}
