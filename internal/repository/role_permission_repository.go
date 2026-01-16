package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var ErrRolePermissionNotFound = errors.New("role permission not found")

// Compile-time interface check
var _ RolePermissionRepositoryInterface = (*RolePermissionRepository)(nil)

// RolePermissionRepository handles role permission data operations
type RolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository creates a new RolePermissionRepository
func NewRolePermissionRepository(db *gorm.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

// Create creates a new role permission association
func (r *RolePermissionRepository) Create(ctx context.Context, rp *model.RolePermission) error {
	return r.db.WithContext(ctx).Create(rp).Error
}

// Update updates a role permission
func (r *RolePermissionRepository) Update(ctx context.Context, rp *model.RolePermission) error {
	return r.db.WithContext(ctx).Save(rp).Error
}

// Delete deletes a role permission association
func (r *RolePermissionRepository) Delete(ctx context.Context, roleID, permissionID uint) error {
	result := r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{})
	if result.RowsAffected == 0 {
		return ErrRolePermissionNotFound
	}
	return result.Error
}

// DeleteByRoleID deletes all permissions for a role
func (r *RolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

// FindByRoleID finds all permissions for a role
func (r *RolePermissionRepository) FindByRoleID(ctx context.Context, roleID uint) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission
	err := r.db.WithContext(ctx).
		Preload("Permission").
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error
	return rolePermissions, err
}

// FindByRoleAndSpace finds role permission by role and space
func (r *RolePermissionRepository) FindByRoleAndSpace(ctx context.Context, roleID, spaceID uint) (*model.RolePermission, error) {
	var rp model.RolePermission
	err := r.db.WithContext(ctx).
		Where("role_id = ? AND space_id = ?", roleID, spaceID).
		First(&rp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rp, err
}

// FindByRoleAndPermission finds role permission by role and permission
func (r *RolePermissionRepository) FindByRoleAndPermission(ctx context.Context, roleID, permissionID uint) (*model.RolePermission, error) {
	var rp model.RolePermission
	err := r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		First(&rp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rp, err
}

// Exists checks if a role permission association exists
func (r *RolePermissionRepository) Exists(ctx context.Context, roleID, permissionID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

// GetSpaceValuesByRoleID returns space values for a role (for permission calculation)
func (r *RolePermissionRepository) GetSpaceValuesByRoleID(ctx context.Context, roleID uint) (map[uint]uint64, error) {
	var results []struct {
		SpaceID uint
		Value   uint64
	}
	err := r.db.WithContext(ctx).
		Model(&model.RolePermission{}).
		Select("space_id, BIT_OR(value) as value").
		Where("role_id = ?", roleID).
		Group("space_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	spaceValues := make(map[uint]uint64)
	for _, r := range results {
		spaceValues[r.SpaceID] = r.Value
	}
	return spaceValues, nil
}
