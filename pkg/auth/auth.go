package auth

import (
	"time"
)

// TokenResponse represents the authentication token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// NewTokenResponse creates a token response with access and refresh tokens
func NewTokenResponse(accessToken, refreshToken string, expiresIn time.Duration) *TokenResponse {
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expiresIn.Seconds()),
	}
}

// AuthService provides authentication operations
type AuthService struct {
	jwtManager     *JWTManager
	passwordHasher *PasswordHasher
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, accessDays, refreshDays int) *AuthService {
	return &AuthService{
		jwtManager:     NewJWTManager(jwtSecret, accessDays, refreshDays),
		passwordHasher: NewPasswordHasher(),
	}
}

// NewAuthServiceWithConfig creates a new authentication service with custom config
func NewAuthServiceWithConfig(jwtConfig TokenConfig, passwordParams *Argon2Params) *AuthService {
	return &AuthService{
		jwtManager:     NewJWTManagerWithConfig(jwtConfig),
		passwordHasher: NewPasswordHasherWithParams(passwordParams),
	}
}

// GenerateTokens generates access and refresh tokens for a user
func (s *AuthService) GenerateTokens(userID uint) (*TokenResponse, error) {
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	return NewTokenResponse(accessToken, refreshToken, s.jwtManager.config.AccessTokenDuration), nil
}

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	return s.jwtManager.RefreshAccessToken(refreshToken)
}

// ValidateAccessToken validates an access token and returns the user ID
func (s *AuthService) ValidateAccessToken(token string) (uint, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// HashPassword hashes a password using Argon2
func (s *AuthService) HashPassword(password string) (string, error) {
	return s.passwordHasher.HashPassword(password)
}

// VerifyPassword verifies a password against a hash
func (s *AuthService) VerifyPassword(password, hash string) (bool, error) {
	return s.passwordHasher.VerifyPassword(password, hash)
}

// GetJWTManager returns the JWT manager (for middleware integration)
func (s *AuthService) GetJWTManager() *JWTManager {
	return s.jwtManager
}

// GetPasswordHasher returns the password hasher
func (s *AuthService) GetPasswordHasher() *PasswordHasher {
	return s.passwordHasher
}
