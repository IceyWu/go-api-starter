package model

import (
	"time"

	"gorm.io/gorm"
)

// PermissionSpace 权限空间
type PermissionSpace struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Permissions []Permission `json:"permissions,omitempty" gorm:"foreignKey:SpaceID"`
}

// Permission 权限定义
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"size:50;uniqueIndex;not null"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	SpaceID     uint           `json:"space_id" gorm:"not null;index"`
	Position    uint8          `json:"position" gorm:"not null"`          // 0-63
	Value       uint64         `json:"value" gorm:"not null"`             // 2^position
	Module      string         `json:"module" gorm:"size:100;index"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Space *PermissionSpace `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
}

// Role 角色
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index"`
	IsSystem    bool           `json:"is_system" gorm:"default:false;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:RoleID"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:uk_user_role"`
	RoleID    uint      `json:"role_id" gorm:"not null;uniqueIndex:uk_user_role;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}


// RolePermission 角色权限关联（存储位运算值）
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RoleID       uint      `json:"role_id" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	PermissionID uint      `json:"permission_id" gorm:"not null;uniqueIndex:uk_role_permission;index"`
	SpaceID      uint      `json:"space_id" gorm:"not null;index"`
	Value        uint64    `json:"value" gorm:"not null"` // 该空间下的位运算值
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Role       *Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission *Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	Space      *PermissionSpace `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
}

// UserPermissionCache 用户权限缓存
type UserPermissionCache struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:uk_user_space"`
	SpaceID   uint      `json:"space_id" gorm:"not null;uniqueIndex:uk_user_space;index"`
	Value     uint64    `json:"value" gorm:"not null"` // 该空间下用户的权限位运算值
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Space *PermissionSpace `json:"space,omitempty" gorm:"foreignKey:SpaceID"`
}

// IsExpired checks if the cache entry has expired
func (c *UserPermissionCache) IsExpired() bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// TableName returns the table name for PermissionSpace
func (PermissionSpace) TableName() string {
	return "permission_spaces"
}

// TableName returns the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// TableName returns the table name for Role
func (Role) TableName() string {
	return "roles"
}

// TableName returns the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// TableName returns the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}

// TableName returns the table name for UserPermissionCache
func (UserPermissionCache) TableName() string {
	return "user_permission_caches"
}


// ==================== Request DTOs ====================

// CreateSpaceRequest 创建权限空间请求
type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100" example:"user"`
	Description string `json:"description" binding:"max=500" example:"用户管理权限空间"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=50" example:"USER_CREATE"`
	Name        string `json:"name" binding:"required,min=2,max=100" example:"创建用户"`
	Description string `json:"description" binding:"max=500" example:"允许创建新用户"`
	SpaceID     uint   `json:"space_id" binding:"required" example:"1"`
	Module      string `json:"module" binding:"max=100" example:"user"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100" example:"创建用户"`
	Description string `json:"description" binding:"max=500" example:"允许创建新用户"`
	IsActive    *bool  `json:"is_active" example:"true"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name            string   `json:"name" binding:"required,min=2,max=100" example:"admin"`
	Description     string   `json:"description" binding:"max=500" example:"系统管理员"`
	PermissionCodes []string `json:"permission_codes" example:"USER_CREATE,USER_READ"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100" example:"admin"`
	Description string `json:"description" binding:"max=500" example:"系统管理员"`
	IsActive    *bool  `json:"is_active" example:"true"`
}

// RolePermissionsRequest 角色权限操作请求
type RolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes" binding:"required,min=1" example:"USER_CREATE,USER_READ"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	RoleID uint `json:"role_id" binding:"required" example:"1"`
}

// ==================== Response DTOs ====================

// SpaceWithCount 带权限数量的权限空间
type SpaceWithCount struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsActive        bool   `json:"is_active"`
	PermissionCount int64  `json:"permission_count"`
}

// PermissionDetail 权限详情
type PermissionDetail struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SpaceID     uint   `json:"space_id"`
	SpaceName   string `json:"space_name"`
	Position    uint8  `json:"position"`
	Value       uint64 `json:"value"`
	Module      string `json:"module"`
	IsActive    bool   `json:"is_active"`
}

// RoleDetail 角色详情
type RoleDetail struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	IsActive        bool       `json:"is_active"`
	IsSystem        bool       `json:"is_system"`
	PermissionCodes []string   `json:"permission_codes"`
	Permissions     []PermissionDetail `json:"permissions,omitempty"`
}

// UserPermissionInfo 用户权限信息
type UserPermissionInfo struct {
	UserID          uint     `json:"user_id"`
	Roles           []Role   `json:"roles"`
	PermissionCodes []string `json:"permission_codes"`
}
