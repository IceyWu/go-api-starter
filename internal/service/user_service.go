package service

import (
	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
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
func (s *UserService) Create(req *model.CreateUserRequest) (*model.User, error) {
	user := req.ToUser()
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// List returns all users
func (s *UserService) List() ([]model.User, error) {
	return s.repo.FindAll()
}

// GetByID returns a user by ID
func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

// Update updates a user
func (s *UserService) Update(id uint, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
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

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Delete deletes a user by ID
func (s *UserService) Delete(id uint) error {
	return s.repo.Delete(id)
}
