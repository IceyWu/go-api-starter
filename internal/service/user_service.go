package service

import (
	"context"
	"errors"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/i18n"
)

// UserService handles user business logic
type UserService struct {
	repo     repository.UserRepositoryInterface
	fileRepo repository.FileRepositoryInterface
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepositoryInterface, fileRepo repository.FileRepositoryInterface) *UserService {
	return &UserService{
		repo:     repo,
		fileRepo: fileRepo,
	}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// Check uniqueness of email
	if req.Email != nil && *req.Email != "" {
		existing, err := s.repo.FindByEmail(ctx, *req.Email)
		if err == nil && existing != nil {
			return nil, apperrors.ConflictCode(i18n.ErrEmailTaken)
		}
	}
	// Check uniqueness of mobile
	if req.Mobile != nil && *req.Mobile != "" {
		existing, err := s.repo.FindByMobile(ctx, *req.Mobile)
		if err == nil && existing != nil {
			return nil, apperrors.ConflictCode(i18n.ErrMobileTaken)
		}
	}

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

// GetBySecUID returns a user by SecUID
func (s *UserService) GetBySecUID(ctx context.Context, secUID string) (*model.User, error) {
	user, err := s.repo.FindBySecUID(ctx, secUID)
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

	if req.LPID != nil && *req.LPID != "" {
		if *req.LPID != user.LPID {
			existing, err := s.repo.FindByLPID(ctx, *req.LPID)
			if err == nil && existing != nil && existing.ID != user.ID {
				return nil, apperrors.ConflictCode(i18n.ErrLPIDTaken)
			}
			user.LPID = *req.LPID
		}
	}
	if req.Username != nil && *req.Username != "" {
		user.Username = req.Username
	}
	if req.Mobile != nil {
		if *req.Mobile != "" && (user.Mobile == nil || *req.Mobile != *user.Mobile) {
			existing, err := s.repo.FindByMobile(ctx, *req.Mobile)
			if err == nil && existing != nil && existing.ID != user.ID {
				return nil, apperrors.ConflictCode(i18n.ErrMobileTaken)
			}
		}
		user.Mobile = req.Mobile
	}
	if req.Email != nil {
		if *req.Email != "" && (user.Email == nil || *req.Email != *user.Email) {
			existing, err := s.repo.FindByEmail(ctx, *req.Email)
			if err == nil && existing != nil && existing.ID != user.ID {
				return nil, apperrors.ConflictCode(i18n.ErrEmailTaken)
			}
		}
		user.Email = req.Email
	}
	if req.AvatarSecUID != nil {
		if *req.AvatarSecUID == "" {
			user.AvatarFileID = nil
			user.AvatarFile = nil
		} else if s.fileRepo != nil {
			file, err := s.fileRepo.FindBySecUID(ctx, *req.AvatarSecUID)
			if err != nil {
				return nil, apperrors.BadRequest("avatar file not found: " + *req.AvatarSecUID)
			}
			user.AvatarFileID = &file.ID
			user.AvatarFile = file
		}
	}
	if req.BackgroundSecUID != nil {
		if *req.BackgroundSecUID == "" {
			user.BackgroundFileID = nil
			user.BackgroundFile = nil
		} else if s.fileRepo != nil {
			file, err := s.fileRepo.FindBySecUID(ctx, *req.BackgroundSecUID)
			if err != nil {
				return nil, apperrors.BadRequest("background file not found: " + *req.BackgroundSecUID)
			}
			user.BackgroundFileID = &file.ID
			user.BackgroundFile = file
		}
	}
	if req.Sex != nil {
		user.Sex = *req.Sex
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.City != nil {
		user.City = req.City
	}
	if req.Job != nil {
		user.Job = req.Job
	}
	if req.Company != nil {
		user.Company = req.Company
	}
	if req.Signature != nil {
		user.Signature = req.Signature
	}
	if req.Website != nil {
		user.Website = req.Website
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
