package repository

import (
	"context"

	"go-api-starter/internal/model"
)

// UserRepositoryInterface defines the interface for user data operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	FindAll(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindBySecUID(ctx context.Context, secUID string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
}

// PermissionRepositoryInterface defines the interface for permission data operations
type PermissionRepositoryInterface interface {
	Create(ctx context.Context, permission *model.Permission) error
	FindByCode(ctx context.Context, code string) (*model.Permission, error)
	FindByID(ctx context.Context, id uint) (*model.Permission, error)
	FindAll(ctx context.Context) ([]model.Permission, error)
	FindBySpaceID(ctx context.Context, spaceID uint) ([]model.Permission, error)
	GetMaxPositionInSpace(ctx context.Context, spaceID uint) (int, error)
	Update(ctx context.Context, permission *model.Permission) error
	SoftDelete(ctx context.Context, id uint) error
	Exists(ctx context.Context, code string) (bool, error)
	FindByCodes(ctx context.Context, codes []string) ([]model.Permission, error)
	CountBySpaceID(ctx context.Context, spaceID uint) (int64, error)
}

// RoleRepositoryInterface defines the interface for role data operations
type RoleRepositoryInterface interface {
	Create(ctx context.Context, role *model.Role) error
	FindByName(ctx context.Context, name string) (*model.Role, error)
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	FindByIDWithPermissions(ctx context.Context, id uint) (*model.Role, error)
	FindAll(ctx context.Context) ([]model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	Exists(ctx context.Context, name string) (bool, error)
}

// OSSRepositoryInterface defines the interface for OSS file data operations
type OSSRepositoryInterface interface {
	Create(file *model.OSSFile) error
	GetByKey(key string) (*model.OSSFile, error)
	GetByMD5(md5 string, userID uint) (*model.OSSFile, error)
	GetByID(id uint) (*model.OSSFile, error)
	List(userID uint, page, pageSize int) ([]model.OSSFile, int64, error)
	Delete(id uint) error
	DeleteByKey(key string) error
}

// PermissionSpaceRepositoryInterface defines the interface for permission space data operations
type PermissionSpaceRepositoryInterface interface {
	Create(ctx context.Context, space *model.PermissionSpace) error
	FindByName(ctx context.Context, name string) (*model.PermissionSpace, error)
	FindByID(ctx context.Context, id uint) (*model.PermissionSpace, error)
	FindAll(ctx context.Context) ([]model.PermissionSpace, error)
	FindAllWithCount(ctx context.Context) ([]model.SpaceWithCount, error)
	Exists(ctx context.Context, name string) (bool, error)
	Update(ctx context.Context, space *model.PermissionSpace) error
}

// UserRoleRepositoryInterface defines the interface for user role data operations
type UserRoleRepositoryInterface interface {
	Create(ctx context.Context, userRole *model.UserRole) error
	Delete(ctx context.Context, userID, roleID uint) error
	FindByUserID(ctx context.Context, userID uint) ([]model.UserRole, error)
	FindByRoleID(ctx context.Context, roleID uint) ([]model.UserRole, error)
	Exists(ctx context.Context, userID, roleID uint) (bool, error)
	GetUserIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error)
}

// RolePermissionRepositoryInterface defines the interface for role permission data operations
type RolePermissionRepositoryInterface interface {
	Create(ctx context.Context, rp *model.RolePermission) error
	Update(ctx context.Context, rp *model.RolePermission) error
	Delete(ctx context.Context, roleID, permissionID uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
	FindByRoleID(ctx context.Context, roleID uint) ([]model.RolePermission, error)
	FindByRoleAndSpace(ctx context.Context, roleID, spaceID uint) (*model.RolePermission, error)
	FindByRoleAndPermission(ctx context.Context, roleID, permissionID uint) (*model.RolePermission, error)
	Exists(ctx context.Context, roleID, permissionID uint) (bool, error)
	GetSpaceValuesByRoleID(ctx context.Context, roleID uint) (map[uint]uint64, error)
}

// UserPermissionCacheRepositoryInterface defines the interface for user permission cache data operations
type UserPermissionCacheRepositoryInterface interface {
	Upsert(ctx context.Context, cache *model.UserPermissionCache) error
	FindByUserAndSpace(ctx context.Context, userID, spaceID uint) (*model.UserPermissionCache, error)
	FindByUserID(ctx context.Context, userID uint) ([]model.UserPermissionCache, error)
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteByUserIDs(ctx context.Context, userIDs []uint) error
	GetUserSpaceValues(ctx context.Context, userID uint) (map[uint]uint64, error)
}

// MultipartRepositoryInterface defines the interface for multipart upload data operations
type MultipartRepositoryInterface interface {
	CreateUpload(upload *model.MultipartUpload) error
	GetUploadByMD5(md5 string, userID uint) (*model.MultipartUpload, error)
	GetUploadByID(uploadID string) (*model.MultipartUpload, error)
	UpdateUploadStatus(uploadID string, status model.MultipartUploadStatus) error
	DeleteUpload(uploadID string) error
	SavePart(part *model.UploadedPart) error
	GetUploadedParts(uploadID string) ([]model.UploadedPart, error)
	DeleteParts(uploadID string) error
	GetUploadedPartNumbers(uploadID string) ([]int, error)
}
