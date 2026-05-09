package service

import (
	"context"
	"mime/multipart"

	"go-api-starter/internal/model"
	"go-api-starter/pkg/oss"
)

// AuthServiceInterface defines the interface for authentication service operations
type AuthServiceInterface interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	AccessTokenExpiresIn() int64
	GetCurrentUser(ctx context.Context, userID uint) (*model.User, error)
	ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error
	Logout(ctx context.Context, token string) error
	LogoutAllDevices(ctx context.Context, userID uint) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetBySecUID(ctx context.Context, secUID string) (*model.User, error)
	List(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error)
	Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id uint) error
}

// PermissionServiceInterface defines the interface for permission service operations
type PermissionServiceInterface interface {
	// Space operations
	CreateSpace(ctx context.Context, req *model.CreateSpaceRequest) (*model.PermissionSpace, error)
	GetAllSpaces(ctx context.Context) ([]model.SpaceWithCount, error)

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
	CheckUserPermission(userID uint, permissionCode string) (bool, error)
}

// OSSServiceInterface defines the interface for OSS service operations
type OSSServiceInterface interface {
	// Simple upload operations
	GetUploadToken(userID uint) (*oss.UploadToken, error)
	GetUploadTokenWithFileName(userID uint, fileName string) (*oss.UploadToken, error)
	CheckFileExists(md5 string, userID uint) (*model.File, bool)
	SaveFileRecord(key, md5, fileName string, fileSize int64, userID uint) (*model.File, error)

	// File operations (all use sec_uid)
	GetFileBySecUID(secUID string) (*model.File, error)
	UpdateFile(secUID string, req *model.UpdateFileRequest) error
	ListFiles(userID uint, isPrivate *bool, offset, limit int, sort string) ([]model.File, int64, error)
	DeleteFile(secUID string) error

	// Multipart upload operations
	InitMultipartUpload(fileName string, md5 string, fileSize int64, chunkSize int64, userID uint) (*MultipartInitResult, error)
	GetPartUploadURL(key, uploadID string, partNumber int) (*PartUploadInfo, error)
	GetPartUploadURLs(key, uploadID string, partNumbers []int) ([]PartUploadInfo, error)
	CompleteMultipartUpload(key, uploadID, md5, fileName string, fileSize int64, parts []CompletePart, userID uint) (*model.File, error)
	AbortMultipartUpload(key, uploadID string) error
	ListUploadedParts(key, uploadID string) ([]CompletePart, error)

	// Resumable upload support
	SaveUploadedPart(uploadID string, partNumber int, etag string, size int64) error
	GetUploadedPartsFromDB(uploadID string) ([]CompletePart, error)
}

// FileServiceInterface defines the interface for file service operations
type FileServiceInterface interface {
	Upload(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (*model.File, error)
	GetByID(ctx context.Context, id uint, userID *uint) (*model.File, error)
	List(ctx context.Context, filter model.FileFilter, offset, limit int, sort string, requestUserID *uint) ([]model.File, int64, error)
	Update(ctx context.Context, id uint, userID uint, req *model.UpdateFileRequest) (*model.File, error)
	Delete(ctx context.Context, id uint, userID uint) error
}
