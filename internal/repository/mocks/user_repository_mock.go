package mocks

import (
	"context"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
)

// Compile-time interface check
var _ repository.UserRepositoryInterface = (*MockUserRepository)(nil)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	users     map[uint]*model.User
	byEmail   map[string]*model.User
	nextID    uint
	CreateErr error
	FindErr   error
	UpdateErr error
	DeleteErr error
}

// NewMockUserRepository creates a new MockUserRepository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[uint]*model.User),
		byEmail: make(map[string]*model.User),
		nextID:  1,
	}
}

// Create creates a new user
func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

// FindAll returns all users with pagination
func (m *MockUserRepository) FindAll(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error) {
	if m.FindErr != nil {
		return nil, 0, m.FindErr
	}
	var result []model.User
	for _, u := range m.users {
		result = append(result, *u)
	}
	total := int64(len(result))
	
	// Apply pagination
	if offset >= len(result) {
		return []model.User{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

// FindByID finds a user by ID
func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

// FindByEmail finds a user by email
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, nil
}

// Update updates a user
func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return repository.ErrUserNotFound
	}
	// Update email index if changed
	for email, u := range m.byEmail {
		if u.ID == user.ID && email != user.Email {
			delete(m.byEmail, email)
			break
		}
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

// Delete deletes a user
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	if u, ok := m.users[id]; ok {
		delete(m.byEmail, u.Email)
		delete(m.users, id)
		return nil
	}
	return repository.ErrUserNotFound
}

// AddUser adds a user directly (for test setup)
func (m *MockUserRepository) AddUser(user *model.User) {
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
}

// Reset clears all data
func (m *MockUserRepository) Reset() {
	m.users = make(map[uint]*model.User)
	m.byEmail = make(map[string]*model.User)
	m.nextID = 1
	m.CreateErr = nil
	m.FindErr = nil
	m.UpdateErr = nil
	m.DeleteErr = nil
}
