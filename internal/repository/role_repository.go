package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound   = errors.New("role not found")
	ErrRoleNameExists = errors.New("role name already exists")
)

// RoleRepository handles role data operations
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new RoleRepository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// FindByName finds a role by name
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

// FindByID finds a role by ID
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

// FindByIDWithPermissions finds a role by ID with permissions
func (r *RoleRepository) FindByIDWithPermissions(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

// FindAll returns all roles
func (r *RoleRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Order("id ASC").Find(&roles).Error
	return roles, err
}

// Update updates a role
func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete soft deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Role{}, id)
	if result.RowsAffected == 0 {
		return ErrRoleNotFound
	}
	return result.Error
}

// Exists checks if a role with the given name exists
func (r *RoleRepository) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Role{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
