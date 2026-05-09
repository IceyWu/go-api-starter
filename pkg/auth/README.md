# Auth Package

JWT authentication and password hashing utilities for the lp-api-go project.

## Features

- **JWT Token Management**
  - Access token generation (configurable, default 7 days)
  - Refresh token generation (configurable, default 30 days)
  - Token validation and verification
  - Token refresh mechanism
  
- **Password Hashing**
  - Argon2id algorithm (recommended by OWASP)
  - Secure password hashing and verification
  - Configurable parameters

## Usage

### Basic Authentication Service

```go
import "go-api-starter/pkg/auth"

// Create auth service (secret, accessDays, refreshDays)
authService := auth.NewAuthService("your-jwt-secret", 7, 30)

// Hash password during registration
hashedPassword, err := authService.HashPassword("userPassword123")

// Verify password during login
valid, err := authService.VerifyPassword("userPassword123", hashedPassword)

// Generate tokens after successful login
tokens, err := authService.GenerateTokens(userID)
// Returns: AccessToken, RefreshToken, TokenType, ExpiresIn

// Validate access token in middleware
userID, err := authService.ValidateAccessToken(accessToken)

// Refresh access token when expired
newAccessToken, err := authService.RefreshToken(refreshToken)
```

### JWT Manager (Standalone)

```go
// Create JWT manager (secret, accessDays, refreshDays)
jwtManager := auth.NewJWTManager("your-secret", 7, 30)

// Generate tokens
accessToken, err := jwtManager.GenerateAccessToken(userID)
refreshToken, err := jwtManager.GenerateRefreshToken(userID)

// Or generate both at once
accessToken, refreshToken, err := jwtManager.GenerateTokenPair(userID)

// Validate tokens
claims, err := jwtManager.ValidateAccessToken(accessToken)
claims, err := jwtManager.ValidateRefreshToken(refreshToken)

// Refresh access token
newAccessToken, err := jwtManager.RefreshAccessToken(refreshToken)
```

### Password Hasher (Standalone)

```go
// Create password hasher
hasher := auth.NewPasswordHasher()

// Hash password
hash, err := hasher.HashPassword("password123")

// Verify password
valid, err := hasher.VerifyPassword("password123", hash)

// Or use convenience functions
hash, err := auth.HashPassword("password123")
valid, err := auth.VerifyPassword("password123", hash)
```

### Custom Configuration

```go
// Custom JWT configuration
jwtConfig := auth.TokenConfig{
    Secret:               "custom-secret",
    AccessTokenDuration:  30 * time.Minute,
    RefreshTokenDuration: 14 * 24 * time.Hour,
}

// Custom Argon2 parameters
passwordParams := &auth.Argon2Params{
    Memory:      64 * 1024, // 64 MB
    Iterations:  3,
    Parallelism: 2,
    SaltLength:  16,
    KeyLength:   32,
}

// Create service with custom config
authService := auth.NewAuthServiceWithConfig(jwtConfig, passwordParams)
```

## Token Format

### Access Token Claims
```json
{
  "user_id": 123,
  "token_type": "access",
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

### Refresh Token Claims
```json
{
  "user_id": 123,
  "token_type": "refresh",
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

## Password Hash Format

Argon2id hash format:
```
$argon2id$v=19$m=65536,t=3,p=2$<base64-salt>$<base64-hash>
```

## Security Considerations

1. **JWT Secret**: Use a strong, random secret key (at least 32 characters)
2. **Token Storage**: Store refresh tokens securely (HTTP-only cookies recommended)
3. **Token Expiration**: Configurable via `ACCESS_TOKEN_DAYS` / `REFRESH_TOKEN_DAYS` env vars (defaults 7d / 30d)
4. **Password Hashing**: Uses Argon2id with secure default parameters
5. **Constant-Time Comparison**: Password verification uses constant-time comparison to prevent timing attacks

## Integration with Middleware

```go
// In your middleware setup
authService := auth.NewAuthService(config.JWTSecret, 7, 30)
jwtManager := authService.GetJWTManager()

// Use with existing auth middleware
authMiddleware := middleware.NewAuthMiddleware(config.JWTSecret, authSvc, userRepo)
```

## Testing

Run tests:
```bash
go test ./pkg/auth/ -v
```

Run tests with coverage:
```bash
go test ./pkg/auth/ -cover
```

## Requirements Validation

This package implements:
- **Requirement 1.1**: JWT token generation for valid credentials
- **Requirement 1.2**: Email/mobile authentication support
- **Requirement 1.4**: Refresh token mechanism
- **Requirement 1.8**: Argon2 password hashing
