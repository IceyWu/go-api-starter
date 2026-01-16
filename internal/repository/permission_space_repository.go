package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var (
	ErrPermissionSpaceNotFound   = errors.New("permission space not found")
	ErrPermissionSpaceNameExists = errors.New("permission space name already exists")
)

// Compile-time interface check
var _ PermissionSpaceRepositoryInterface = (*PermissionSpaceRepository)(nil)

// PermissionSpaceRepository handles permission space data operations
type PermissionSpaceRepository struct {
	db *gorm.DB
}

// NewPermissionSpaceRepository creates a new PermissionSpaceRepository
func NewPermissionSpaceRepository(db *gorm.DB) *PermissionSpaceRepository {
	return &PermissionSpaceRepository{db: db}
}

// Create creates a new permission space
func (r *PermissionSpaceRepository) Create(ctx context.Context, space *model.PermissionSpace) error {
	return r.db.WithContext(ctx).Create(space).Error
}

// FindByName finds a permission space by name
func (r *PermissionSpaceRepository) FindByName(ctx context.Context, name string) (*model.PermissionSpace, error) {
	var space model.PermissionSpace
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&space).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPermissionSpaceNotFound
	}
	return &space, err
}

// FindByID finds a permission space by ID
func (r *PermissionSpaceRepository) FindByID(ctx context.Context, id uint) (*model.PermissionSpace, error) {
	var space model.PermissionSpace
	err := r.db.WithContext(ctx).First(&space, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPermissionSpaceNotFound
	}
	return &space, err
}

// FindAll returns all permission spaces
func (r *PermissionSpaceRepository) FindAll(ctx context.Context) ([]model.PermissionSpace, error) {
	var spaces []model.PermissionSpace
	err := r.db.WithContext(ctx).Order("id ASC").Find(&spaces).Error
	return spaces, err
}

// FindAllWithCount returns all permission spaces with permission count
func (r *PermissionSpaceRepository) FindAllWithCount(ctx context.Context) ([]model.SpaceWithCount, error) {
	var results []model.SpaceWithCount
	err := r.db.WithContext(ctx).
		Model(&model.PermissionSpace{}).
		Select("permission_spaces.id, permission_spaces.name, permission_spaces.description, permission_spaces.is_active, COUNT(permissions.id) as permission_count").
		Joins("LEFT JOIN permissions ON permissions.space_id = permission_spaces.id AND permissions.deleted_at IS NULL").
		Group("permission_spaces.id").
		Order("permission_spaces.id ASC").
		Scan(&results).Error
	return results, err
}

// Exists checks if a permission space with the given name exists
func (r *PermissionSpaceRepository) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PermissionSpace{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// Update updates a permission space
func (r *PermissionSpaceRepository) Update(ctx context.Context, space *model.PermissionSpace) error {
	return r.db.WithContext(ctx).Save(space).Error
}
