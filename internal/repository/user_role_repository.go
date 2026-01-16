package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var (
	ErrUserRoleNotFound      = errors.New("user role not found")
	ErrUserRoleAlreadyExists = errors.New("user already has this role")
)

// Compile-time interface check
var _ UserRoleRepositoryInterface = (*UserRoleRepository)(nil)

// UserRoleRepository handles user role data operations
type UserRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new UserRoleRepository
func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

// Create creates a new user role association
func (r *UserRoleRepository) Create(ctx context.Context, userRole *model.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// Delete deletes a user role association
func (r *UserRoleRepository) Delete(ctx context.Context, userID, roleID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{})
	if result.RowsAffected == 0 {
		return ErrUserRoleNotFound
	}
	return result.Error
}

// FindByUserID finds all roles for a user
func (r *UserRoleRepository) FindByUserID(ctx context.Context, userID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ?", userID).
		Find(&userRoles).Error
	return userRoles, err
}

// FindByRoleID finds all users with a role
func (r *UserRoleRepository) FindByRoleID(ctx context.Context, roleID uint) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&userRoles).Error
	return userRoles, err
}

// Exists checks if a user role association exists
func (r *UserRoleRepository) Exists(ctx context.Context, userID, roleID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	return count > 0, err
}

// GetUserIDsByRoleID returns all user IDs with a specific role
func (r *UserRoleRepository) GetUserIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.UserRole{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
