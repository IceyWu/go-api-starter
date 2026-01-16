package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var (
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrPermissionCodeExists = errors.New("permission code already exists")
)

// Compile-time interface check
var _ PermissionRepositoryInterface = (*PermissionRepository)(nil)

// PermissionRepository handles permission data operations
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new PermissionRepository
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// FindByCode finds a permission by code
func (r *PermissionRepository) FindByCode(ctx context.Context, code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPermissionNotFound
	}
	return &permission, err
}

// FindByID finds a permission by ID
func (r *PermissionRepository) FindByID(ctx context.Context, id uint) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).Preload("Space").First(&permission, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPermissionNotFound
	}
	return &permission, err
}

// FindAll returns all permissions with space info
func (r *PermissionRepository) FindAll(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.WithContext(ctx).Preload("Space").Order("space_id ASC, position ASC").Find(&permissions).Error
	return permissions, err
}

// FindBySpaceID returns all permissions in a space
func (r *PermissionRepository) FindBySpaceID(ctx context.Context, spaceID uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Order("position ASC").Find(&permissions).Error
	return permissions, err
}

// GetMaxPositionInSpace returns the max position in a space
func (r *PermissionRepository) GetMaxPositionInSpace(ctx context.Context, spaceID uint) (int, error) {
	var maxPosition *int
	err := r.db.WithContext(ctx).
		Model(&model.Permission{}).
		Where("space_id = ?", spaceID).
		Select("MAX(position)").
		Scan(&maxPosition).Error
	if err != nil {
		return -1, err
	}
	if maxPosition == nil {
		return -1, nil
	}
	return *maxPosition, nil
}

// Update updates a permission
func (r *PermissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// SoftDelete soft deletes a permission
func (r *PermissionRepository) SoftDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Permission{}, id)
	if result.RowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return result.Error
}

// Exists checks if a permission with the given code exists
func (r *PermissionRepository) Exists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// FindByCodes finds permissions by codes
func (r *PermissionRepository) FindByCodes(ctx context.Context, codes []string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

// CountBySpaceID returns the count of permissions in a space
func (r *PermissionRepository) CountBySpaceID(ctx context.Context, spaceID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("space_id = ?", spaceID).Count(&count).Error
	return count, err
}
