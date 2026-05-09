package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidTokenType     = errors.New("invalid token type")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// TokenConfig holds JWT token configuration
type TokenConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// Claims represents JWT claims
type Claims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	config TokenConfig
}

// NewJWTManager creates a new JWT manager
// accessDays and refreshDays specify token lifetimes in days.
// If either is <= 0, defaults to 7 and 30 days respectively.
func NewJWTManager(secret string, accessDays, refreshDays int) *JWTManager {
	if accessDays <= 0 {
		accessDays = 7
	}
	if refreshDays <= 0 {
		refreshDays = 30
	}
	return &JWTManager{
		config: TokenConfig{
			Secret:               secret,
			AccessTokenDuration:  time.Duration(accessDays) * 24 * time.Hour,
			RefreshTokenDuration: time.Duration(refreshDays) * 24 * time.Hour,
		},
	}
}

// NewJWTManagerWithConfig creates a new JWT manager with custom config
func NewJWTManagerWithConfig(config TokenConfig) *JWTManager {
	return &JWTManager{
		config: config,
	}
}

// GenerateAccessToken generates an access token for a user
func (m *JWTManager) GenerateAccessToken(userID uint) (string, error) {
	return m.generateToken(userID, TokenTypeAccess, m.config.AccessTokenDuration)
}

// GenerateRefreshToken generates a refresh token for a user
func (m *JWTManager) GenerateRefreshToken(userID uint) (string, error) {
	return m.generateToken(userID, TokenTypeRefresh, m.config.RefreshTokenDuration)
}

// GenerateTokenPair generates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error) {
	accessToken, err = m.GenerateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// AccessTokenExpiresIn returns the access token duration in seconds
func (m *JWTManager) AccessTokenExpiresIn() int64 {
	return int64(m.config.AccessTokenDuration.Seconds())
}

// generateToken generates a JWT token with specified type and duration
func (m *JWTManager) generateToken(userID uint, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	return m.GenerateAccessToken(claims.UserID)
}

// ExtractUserID extracts user ID from token without full validation (for logging, etc.)
func (m *JWTManager) ExtractUserID(tokenString string) (uint, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
