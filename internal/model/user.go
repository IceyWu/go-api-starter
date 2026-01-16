package model

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        uint           `json:"-" gorm:"primaryKey"`                          // 内部ID，不对外暴露
	SecUID    string         `json:"sec_uid" gorm:"size:32;uniqueIndex;not null"`  // 安全标识符，对外暴露
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100;uniqueIndex"`
	Password  string         `json:"-" gorm:"size:255;not null"` // Password hash, not exposed in JSON
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 创建前自动生成 SecUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SecUID == "" {
		u.SecUID = GenerateSecUID()
	}
	return nil
}

// GenerateSecUID 生成安全标识符 (22字符，URL安全的base64)
func GenerateSecUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"omitempty,email" example:"john@example.com"`
}

// ToUser converts CreateUserRequest to User model
func (r *CreateUserRequest) ToUser() *User {
	return &User{
		Name:  r.Name,
		Email: r.Email,
	}
}
