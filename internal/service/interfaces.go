package service

import (
	"context"

	"go-api-starter/internal/model"
	"go-api-starter/pkg/oss"
)

// AuthServiceInterface defines the interface for authentication service operations
// Validates: Requirements 7.1, 7.3
type AuthServiceInterface interface {
	// Register creates a new user account
	// Validates input, hashes password, and creates user record
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)

	// Login authenticates a user and returns a JWT token
	// Verifies credentials and generates JWT token
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)

	// GetCurrentUser retrieves user data by ID
	GetCurrentUser(ctx context.Context, userID uint) (*model.User, error)

	// ResetPassword updates a user's password
	// Validates user existence and updates password hash
	ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error

	// Logout invalidates the current token
	Logout(ctx context.Context, token string) error

	// LogoutAllDevices invalidates all tokens for a user
	LogoutAllDevices(ctx context.Context, userID uint) error

	// IsTokenBlacklisted checks if a token is blacklisted
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// UserServiceInterface defines the interface for user service operations
// Validates: Requirements 7.1, 7.2
type UserServiceInterface interface {
	// Create creates a new user
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*model.User, error)

	// GetBySecUID retrieves a user by SecUID
	GetBySecUID(ctx context.Context, secUID string) (*model.User, error)

	// List returns users with pagination and sorting
	List(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error)

	// Update updates a user's information
	Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error)

	// Delete removes a user by ID
	Delete(ctx context.Context, id uint) error
}

// PermissionServiceInterface defines the interface for permission service operations
// Validates: Requirements 7.1, 7.4
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
// Validates: Requirements 7.1, 7.5
type OSSServiceInterface interface {
	// Simple upload operations
	GetUploadToken(fileName string, userID uint) (*oss.UploadToken, error)
	CheckFileExists(md5 string, userID uint) (*model.OSSFile, bool)
	SaveFileRecord(key, md5, fileName string, fileSize int64, contentType string, userID uint) (*model.OSSFile, error)

	// File operations
	GetFileByKey(key string) (*model.OSSFile, error)
	GetFileByID(id uint) (*model.OSSFile, error)
	ListFiles(userID uint, page, pageSize int) ([]model.OSSFile, int64, error)
	DeleteFile(id uint) error

	// Multipart upload operations
	InitMultipartUpload(fileName string, md5 string, fileSize int64, chunkSize int64, userID uint) (*MultipartInitResult, error)
	GetPartUploadURL(key, uploadID string, partNumber int) (*PartUploadInfo, error)
	GetPartUploadURLs(key, uploadID string, partNumbers []int) ([]PartUploadInfo, error)
	CompleteMultipartUpload(key, uploadID, md5, fileName string, fileSize int64, contentType string, parts []CompletePart, userID uint) (*model.OSSFile, error)
	AbortMultipartUpload(key, uploadID string) error
	ListUploadedParts(key, uploadID string) ([]CompletePart, error)

	// Resumable upload support
	SaveUploadedPart(uploadID string, partNumber int, etag string, size int64) error
	GetUploadedPartsFromDB(uploadID string) ([]CompletePart, error)
}
