package service

import (
	"context"
	"errors"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/auth"
	"go-api-starter/pkg/i18n"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       repository.UserRepositoryInterface
	jwtManager     *auth.JWTManager
	passwordHasher *auth.PasswordHasher
	tokenBlacklist TokenBlacklist
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtManager *auth.JWTManager, blacklist TokenBlacklist) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		passwordHasher: auth.NewPasswordHasher(),
		tokenBlacklist: blacklist,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.LoginResponse, error) {
	// Validate that at least one of mobile or email is provided
	if req.Mobile == nil && req.Email == nil {
		return nil, apperrors.BadRequestCode(i18n.ErrMobileOrEmailRequired)
	}

	// Check if mobile already exists
	if req.Mobile != nil {
		existingUser, err := s.userRepo.FindByMobile(ctx, *req.Mobile)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.InternalCode(err, i18n.ErrQueryUserFailed)
		}
		if existingUser != nil {
			return nil, apperrors.ConflictCode(i18n.ErrMobileTaken)
		}
	}

	// Check if email already exists
	if req.Email != nil {
		existingUser, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.InternalCode(err, i18n.ErrQueryUserFailed)
		}
		if existingUser != nil {
			return nil, apperrors.ConflictCode(i18n.ErrEmailTaken)
		}
	}

	// Hash password using Argon2
	hashedPassword, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrHashPasswordFailed)
	}

	// Create user
	user := &model.User{
		Mobile:   req.Mobile,
		Email:    req.Email,
		Password: &hashedPassword,
		Freezed:  false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrCreateUserFailed)
	}

	// Generate JWT tokens for the newly registered user
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrGenerateTokenFailed)
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.AccessTokenExpiresIn(),
		User:         user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// Validate that at least one of mobile or email is provided
	if req.Mobile == nil && req.Email == nil {
		return nil, apperrors.BadRequestCode(i18n.ErrMobileOrEmailRequired)
	}

	var user *model.User
	var err error

	// Find user by mobile or email
	if req.Mobile != nil {
		user, err = s.userRepo.FindByMobile(ctx, *req.Mobile)
	} else if req.Email != nil {
		user, err = s.userRepo.FindByEmail(ctx, *req.Email)
	}

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.UnauthorizedCode(i18n.ErrWrongCredentials)
		}
		return nil, apperrors.InternalCode(err, i18n.ErrQueryUserFailed)
	}

	// Check if user is frozen
	if user.Freezed {
		return nil, apperrors.ForbiddenCode(i18n.ErrAccountFrozen)
	}

	// Verify password using Argon2
	if user.Password == nil {
		return nil, apperrors.UnauthorizedCode(i18n.ErrWrongCredentials)
	}
	valid, err := s.passwordHasher.VerifyPassword(req.Password, *user.Password)
	if err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrVerifyPasswordFailed)
	}
	if !valid {
		return nil, apperrors.UnauthorizedCode(i18n.ErrWrongCredentials)
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrGenerateTokenFailed)
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.AccessTokenExpiresIn(),
		User:         user.ToResponse(),
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return "", apperrors.UnauthorizedCode(i18n.ErrRefreshTokenExpired)
		}
		return "", apperrors.UnauthorizedCode(i18n.ErrInvalidRefreshToken)
	}
	return accessToken, nil
}

// AccessTokenExpiresIn returns the access token duration in seconds
func (s *AuthService) AccessTokenExpiresIn() int64 {
	return s.jwtManager.AccessTokenExpiresIn()
}

// GetCurrentUser retrieves the current authenticated user
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperrors.NotFoundCode(i18n.ErrUserNotFound)
		}
		return nil, apperrors.InternalCode(err, i18n.ErrQueryUserFailed)
	}
	return user, nil
}

// ResetPassword resets a user's password (admin)
func (s *AuthService) ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return apperrors.NotFoundCode(i18n.ErrUserNotFound)
		}
		return apperrors.InternalCode(err, i18n.ErrQueryUserFailed)
	}

	hashedPassword, err := s.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		return apperrors.InternalCode(err, i18n.ErrHashPasswordFailed)
	}

	user.Password = &hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperrors.InternalCode(err, i18n.ErrResetPasswordFailed)
	}
	return nil
}

// Logout invalidates the current token
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if s.tokenBlacklist == nil {
		return nil
	}

	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		// Token already invalid, no need to blacklist
		return nil
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	return s.tokenBlacklist.Add(ctx, token, ttl)
}

// LogoutAllDevices invalidates all tokens for a user
func (s *AuthService) LogoutAllDevices(ctx context.Context, userID uint) error {
	if s.tokenBlacklist == nil {
		return nil
	}
	return s.tokenBlacklist.InvalidateUserTokens(ctx, userID)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if s.tokenBlacklist == nil {
		return false, nil
	}
	return s.tokenBlacklist.IsBlacklisted(ctx, token)
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	claims, err := s.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return 0, apperrors.UnauthorizedCode(i18n.ErrTokenExpired)
		}
		return 0, apperrors.UnauthorizedCode(i18n.ErrInvalidToken)
	}
	return claims.UserID, nil
}
