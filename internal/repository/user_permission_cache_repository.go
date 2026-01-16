package repository

import (
	"context"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Compile-time interface check
var _ UserPermissionCacheRepositoryInterface = (*UserPermissionCacheRepository)(nil)

// UserPermissionCacheRepository handles user permission cache data operations
type UserPermissionCacheRepository struct {
	db *gorm.DB
}

// NewUserPermissionCacheRepository creates a new UserPermissionCacheRepository
func NewUserPermissionCacheRepository(db *gorm.DB) *UserPermissionCacheRepository {
	return &UserPermissionCacheRepository{db: db}
}

// Upsert creates or updates a user permission cache
func (r *UserPermissionCacheRepository) Upsert(ctx context.Context, cache *model.UserPermissionCache) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "space_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(cache).Error
}

// FindByUserAndSpace finds a cache entry by user and space
func (r *UserPermissionCacheRepository) FindByUserAndSpace(ctx context.Context, userID, spaceID uint) (*model.UserPermissionCache, error) {
	var cache model.UserPermissionCache
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND space_id = ?", userID, spaceID).
		First(&cache).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &cache, err
}

// FindByUserID finds all cache entries for a user
func (r *UserPermissionCacheRepository) FindByUserID(ctx context.Context, userID uint) ([]model.UserPermissionCache, error) {
	var caches []model.UserPermissionCache
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&caches).Error
	return caches, err
}

// DeleteByUserID deletes all cache entries for a user
func (r *UserPermissionCacheRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserPermissionCache{}).Error
}

// DeleteByUserIDs deletes all cache entries for multiple users
func (r *UserPermissionCacheRepository) DeleteByUserIDs(ctx context.Context, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Delete(&model.UserPermissionCache{}).Error
}

// GetUserSpaceValues returns all space values for a user
func (r *UserPermissionCacheRepository) GetUserSpaceValues(ctx context.Context, userID uint) (map[uint]uint64, error) {
	var caches []model.UserPermissionCache
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&caches).Error
	if err != nil {
		return nil, err
	}

	spaceValues := make(map[uint]uint64)
	for _, c := range caches {
		spaceValues[c.SpaceID] = c.Value
	}
	return spaceValues, nil
}
