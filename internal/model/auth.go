package model

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100" example:"Admin User"`
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Age      int    `json:"age" binding:"gte=0,lte=150" example:"25"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  *User  `json:"user"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}
