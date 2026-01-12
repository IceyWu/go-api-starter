package service

import (
	"context"
	"errors"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	user := req.ToUser()
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, apperrors.Wrap(err, "failed to create user")
	}
	return user, nil
}

// List returns users with pagination and sorting
func (s *UserService) List(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error) {
	users, total, err := s.repo.FindAll(ctx, offset, limit, sort)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list users")
	}
	return users, total, nil
}

// GetByID returns a user by ID
func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Wrap(err, "failed to get user")
	}
	return user, nil
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Wrap(err, "failed to find user")
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Age != nil {
		user.Age = *req.Age
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, apperrors.Wrap(err, "failed to update user")
	}
	return user, nil
}

// Delete deletes a user by ID
func (s *UserService) Delete(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return apperrors.NotFound("user not found")
		}
		return apperrors.Wrap(err, "failed to delete user")
	}
	return nil
}
