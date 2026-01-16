package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

// Compile-time interface check
var _ UserRepositoryInterface = (*UserRepository)(nil)

// UserRepository handles user data operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindAll returns all users with pagination and sorting
func (r *UserRepository) FindAll(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated and sorted data
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order(sort).Find(&users).Error
	return users, total, err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return result.Error
}
