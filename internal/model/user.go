package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// UsernamePrefix 用户账号前缀，默认 "go"，可通过配置修改
var UsernamePrefix = "go"

// User represents the user model
type User struct {
	ID        uint           `json:"-" gorm:"primaryKey"`                          // 内部ID，不对外暴露
	SecUID    string         `json:"sec_uid" gorm:"size:32;uniqueIndex;not null"`  // 安全标识符，对外暴露
	Username  string         `json:"username" gorm:"size:50;uniqueIndex;not null"` // 用户账号，唯一
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100;uniqueIndex"`
	Password  string         `json:"-" gorm:"size:255;not null"` // Password hash, not exposed in JSON
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 创建前自动生成 SecUID 和 Username
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SecUID == "" {
		u.SecUID = GenerateSecUID()
	}
	if u.Username == "" {
		u.Username = GenerateUsername()
	}
	return nil
}

// GenerateSecUID 生成安全标识符 (22字符，URL安全的base64)
func GenerateSecUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// GenerateUsername 生成默认用户账号 (prefix_xxxxxxxx，8位随机数字)
func GenerateUsername() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000000))
	return fmt.Sprintf("%s_%08d", UsernamePrefix, n.Int64())
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=4,max=50,alphanum" example:"john_doe"` // 用户账号
	Name     string `json:"name" binding:"omitempty,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" binding:"omitempty,email" example:"john@example.com"`
}

// ToUser converts CreateUserRequest to User model
func (r *CreateUserRequest) ToUser() *User {
	return &User{
		Name:  r.Name,
		Email: r.Email,
	}
}
