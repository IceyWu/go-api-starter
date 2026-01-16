package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
)

const (
	// TokenExpiration is the default token expiration time
	TokenExpiration = time.Hour * 24 * 7 // 7 days
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       repository.UserRepositoryInterface
	jwtSecret      string
	tokenBlacklist TokenBlacklist
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// NewAuthServiceWithBlacklist creates a new AuthService with token blacklist support
func NewAuthServiceWithBlacklist(userRepo repository.UserRepositoryInterface, jwtSecret string, blacklist TokenBlacklist) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		tokenBlacklist: blacklist,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != repository.ErrUserNotFound {
		return nil, apperrors.Internal(err, "查询用户失败")
	}
	if existingUser != nil {
		return nil, apperrors.Conflict("邮箱已被注册")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal(err, "密码加密失败")
	}

	// Create user
	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.Internal(err, "创建用户失败")
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, apperrors.Unauthorized("邮箱或密码错误")
		}
		return nil, apperrors.Internal(err, "查询用户失败")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("邮箱或密码错误")
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, apperrors.Internal(err, "生成令牌失败")
	}

	return &model.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetCurrentUser retrieves the current authenticated user
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, apperrors.NotFound("用户不存在")
		}
		return nil, apperrors.Internal(err, "查询用户失败")
	}
	return user, nil
}

// ResetPassword resets a user's password
func (s *AuthService) ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error {
	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return apperrors.NotFound("用户不存在")
		}
		return apperrors.Internal(err, "查询用户失败")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.Internal(err, "密码加密失败")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperrors.Internal(err, "密码重置失败")
	}

	return nil
}

// generateToken creates a JWT token for the given user ID
func (s *AuthService) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(TokenExpiration).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// Logout invalidates the current token
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if s.tokenBlacklist == nil {
		// No blacklist configured, logout is a no-op
		return nil
	}

	// Parse token to get expiration time
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		// Token is invalid, no need to blacklist
		return nil
	}

	// Calculate remaining TTL
	var ttl time.Duration
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			ttl = time.Until(expTime)
			if ttl <= 0 {
				// Token already expired
				return nil
			}
		}
	}

	if ttl == 0 {
		ttl = TokenExpiration
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
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.Unauthorized("无效的令牌签名方法")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, apperrors.Unauthorized("无效的令牌")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, apperrors.Unauthorized("无效的令牌内容")
		}
		return uint(userIDFloat), nil
	}

	return 0, apperrors.Unauthorized("无效的令牌")
}
