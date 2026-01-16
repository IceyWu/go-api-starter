package mocks

import (
	"context"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
)

// Compile-time interface check
var _ service.AuthServiceInterface = (*MockAuthService)(nil)

// MockAuthService is a mock implementation of AuthServiceInterface
type MockAuthService struct {
	users            map[string]*model.User
	tokens           map[uint]string
	blacklistedTokens map[string]bool
	nextID           uint
	RegisterErr      error
	LoginErr         error
	GetUserErr       error
	ResetPassErr     error
	LogoutErr        error
}

// NewMockAuthService creates a new MockAuthService
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		users:            make(map[string]*model.User),
		tokens:           make(map[uint]string),
		blacklistedTokens: make(map[string]bool),
		nextID:           1,
	}
}

// Register creates a new user
func (m *MockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	if m.RegisterErr != nil {
		return nil, m.RegisterErr
	}
	user := &model.User{
		ID:    m.nextID,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	m.nextID++
	m.users[req.Email] = user
	return user, nil
}

// Login authenticates a user
func (m *MockAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	if m.LoginErr != nil {
		return nil, m.LoginErr
	}
	user, ok := m.users[req.Email]
	if !ok {
		return nil, nil
	}
	token := "mock-token-" + req.Email
	m.tokens[user.ID] = token
	return &model.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetCurrentUser retrieves user by ID
func (m *MockAuthService) GetCurrentUser(ctx context.Context, userID uint) (*model.User, error) {
	if m.GetUserErr != nil {
		return nil, m.GetUserErr
	}
	for _, user := range m.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, nil
}

// ResetPassword resets user password
func (m *MockAuthService) ResetPassword(ctx context.Context, userID uint, req *model.ResetPasswordRequest) error {
	if m.ResetPassErr != nil {
		return m.ResetPassErr
	}
	return nil
}

// Logout invalidates a token
func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	if m.LogoutErr != nil {
		return m.LogoutErr
	}
	m.blacklistedTokens[token] = true
	return nil
}

// LogoutAllDevices invalidates all tokens for a user
func (m *MockAuthService) LogoutAllDevices(ctx context.Context, userID uint) error {
	if m.LogoutErr != nil {
		return m.LogoutErr
	}
	// Clear all tokens for the user
	delete(m.tokens, userID)
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (m *MockAuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return m.blacklistedTokens[token], nil
}

// AddUser adds a user directly (for test setup)
func (m *MockAuthService) AddUser(user *model.User) {
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	m.users[user.Email] = user
}

// Reset clears all data
func (m *MockAuthService) Reset() {
	m.users = make(map[string]*model.User)
	m.tokens = make(map[uint]string)
	m.blacklistedTokens = make(map[string]bool)
	m.nextID = 1
	m.RegisterErr = nil
	m.LoginErr = nil
	m.GetUserErr = nil
	m.ResetPassErr = nil
	m.LogoutErr = nil
}
