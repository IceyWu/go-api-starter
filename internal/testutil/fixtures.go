package testutil

import (
	"time"

	"go-api-starter/internal/model"
)

// NewTestUser creates a test user with default values
func NewTestUser() *model.User {
	return &model.User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestUserWithID creates a test user with specified ID
func NewTestUserWithID(id uint) *model.User {
	user := NewTestUser()
	user.ID = id
	user.Email = "test" + string(rune('0'+id)) + "@example.com"
	return user
}

// NewTestRole creates a test role with default values
func NewTestRole() *model.Role {
	return &model.Role{
		ID:          1,
		Name:        "admin",
		Description: "Administrator role",
		IsActive:    true,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestRoleWithID creates a test role with specified ID
func NewTestRoleWithID(id uint, name string) *model.Role {
	return &model.Role{
		ID:          id,
		Name:        name,
		Description: name + " role",
		IsActive:    true,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestPermissionSpace creates a test permission space
func NewTestPermissionSpace() *model.PermissionSpace {
	return &model.PermissionSpace{
		ID:          1,
		Name:        "user",
		Description: "User management permissions",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestPermission creates a test permission
func NewTestPermission() *model.Permission {
	return &model.Permission{
		ID:          1,
		Code:        "user.create",
		Name:        "Create User",
		Description: "Permission to create users",
		SpaceID:     1,
		Position:    0,
		Value:       1,
		Module:      "user",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestPermissionWithPosition creates a test permission with specified position
func NewTestPermissionWithPosition(id uint, code string, spaceID uint, position uint8) *model.Permission {
	return &model.Permission{
		ID:          id,
		Code:        code,
		Name:        code,
		Description: "Permission " + code,
		SpaceID:     spaceID,
		Position:    position,
		Value:       uint64(1) << position,
		Module:      "test",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewTestUserRole creates a test user-role association
func NewTestUserRole(userID, roleID uint) *model.UserRole {
	return &model.UserRole{
		ID:        1,
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestRolePermission creates a test role-permission association
func NewTestRolePermission(roleID, permissionID, spaceID uint, value uint64) *model.RolePermission {
	return &model.RolePermission{
		ID:           1,
		RoleID:       roleID,
		PermissionID: permissionID,
		SpaceID:      spaceID,
		Value:        value,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewTestRegisterRequest creates a test registration request
func NewTestRegisterRequest() *model.RegisterRequest {
	return &model.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}
}

// NewTestLoginRequest creates a test login request
func NewTestLoginRequest() *model.LoginRequest {
	return &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
}

// NewTestCreateUserRequest creates a test create user request
func NewTestCreateUserRequest() *model.CreateUserRequest {
	return &model.CreateUserRequest{
		Name:  "New User",
		Email: "newuser@example.com",
		Age:   30,
	}
}

// NewTestUpdateUserRequest creates a test update user request
func NewTestUpdateUserRequest() *model.UpdateUserRequest {
	age := 35
	return &model.UpdateUserRequest{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   &age,
	}
}
