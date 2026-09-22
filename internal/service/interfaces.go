package service

import (
	"context"
	"io"

	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/storage"
)

// AuthServiceInterface defines the interface for authentication service operations
type AuthServiceInterface interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	LoginByCode(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	AccessTokenExpiresIn() int64
	GetCurrentUser(ctx context.Context, userID uint) (*model.User, error)
	ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error
	SelfResetPassword(ctx context.Context, req *model.SelfResetPasswordRequest) error
	Logout(ctx context.Context, token string) error
	LogoutAllDevices(ctx context.Context, userID uint) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUID(ctx context.Context, uid string) (*model.User, error)
	List(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error)
	Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id uint) error
}

// PermissionServiceInterface defines the interface for permission service operations
type PermissionServiceInterface interface {
	// Space operations
	CreateSpace(ctx context.Context, req *model.CreateSpaceRequest) (*model.PermissionSpace, error)
	GetAllSpaces(ctx context.Context) ([]model.SpaceWithCount, error)
	UpdateSpace(ctx context.Context, id uint, req *model.UpdateSpaceRequest) (*model.PermissionSpace, error)
	DeleteSpace(ctx context.Context, id uint) error

	// Permission operations
	CreatePermission(ctx context.Context, req *model.CreatePermissionRequest) (*model.Permission, error)
	GetAllPermissions(ctx context.Context) ([]model.PermissionDetail, error)
	GetPermissionByID(ctx context.Context, id uint) (*model.PermissionDetail, error)
	UpdatePermission(ctx context.Context, id uint, req *model.UpdatePermissionRequest) (*model.Permission, error)
	DeletePermission(ctx context.Context, id uint) error

	// Role operations
	CreateRole(ctx context.Context, req *model.CreateRoleRequest) (*model.Role, error)
	GetAllRoles(ctx context.Context) ([]model.Role, error)
	GetRoleByID(ctx context.Context, id uint) (*model.RoleDetail, error)
	UpdateRole(ctx context.Context, id uint, req *model.UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, id uint) error

	// Role permission operations
	GetRolePermissions(ctx context.Context, roleID uint) ([]string, error)
	AddRolePermissions(ctx context.Context, roleID uint, codes []string) error
	RemoveRolePermissions(ctx context.Context, roleID uint, codes []string) error

	// User role operations
	GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error)
	AssignUserRole(ctx context.Context, userID, roleID uint) error
	RemoveUserRole(ctx context.Context, userID, roleID uint) error

	// Permission check operations
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
	HasPermission(ctx context.Context, userID uint, code string) (bool, error)
	CheckUserPermission(ctx context.Context, userID uint, permissionCode string) (bool, error)
}

// StorageServiceInterface defines file and object-storage operations.
type StorageServiceInterface interface {
	// Upload session operations
	CreateUpload(fileName, contentType, checksum string, fileSize, partSize int64, userID uint) (*UploadSession, error)
	CompleteUpload(uploadID string, userID uint, parts []CompletePart) (*model.File, error)
	AbortUpload(uploadID string, userID uint) error
	UploadPublic(ctx context.Context, key string, body io.Reader, contentType string) (*storage.ObjectInfo, error)
	CheckFileExists(md5 string, userID uint) (*model.File, bool)

	// File operations (all use uid)
	GetFileByUID(uid string) (*model.File, error)
	UpdateFile(uid string, req *model.UpdateFileRequest) error
	ListFiles(userID uint, isPrivate *bool, offset, limit int, sort string) ([]model.File, int64, error)
	DeleteFile(uid string) error
	UpdateFileTranscodingTask(uid, taskID string) error
}
