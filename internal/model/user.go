package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100;uniqueIndex"`
	Age       int            `json:"age" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	Age   int    `json:"age" binding:"gte=0,lte=150" example:"25"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"omitempty,email" example:"john@example.com"`
	Age   *int   `json:"age" binding:"omitempty,gte=0,lte=150" example:"30"`
}

// ToUser converts CreateUserRequest to User model
func (r *CreateUserRequest) ToUser() *User {
	return &User{
		Name:  r.Name,
		Email: r.Email,
		Age:   r.Age,
	}
}
