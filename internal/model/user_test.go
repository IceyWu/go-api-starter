package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenerateSecUID tests SecUID generation
func TestGenerateSecUID(t *testing.T) {
	secUID1 := GenerateSecUID()
	secUID2 := GenerateSecUID()

	// SecUID should be 22 characters (base64 URL encoding of 16 bytes)
	assert.Equal(t, 22, len(secUID1), "SecUID should be 22 characters")
	assert.Equal(t, 22, len(secUID2), "SecUID should be 22 characters")

	// SecUIDs should be unique
	assert.NotEqual(t, secUID1, secUID2, "SecUIDs should be unique")

	// SecUID should be URL-safe (no +, /, or =)
	assert.NotContains(t, secUID1, "+")
	assert.NotContains(t, secUID1, "/")
	assert.NotContains(t, secUID1, "=")
}

// TestGenerateUsername tests username generation
func TestGenerateUsername(t *testing.T) {
	username1 := GenerateUsername()
	username2 := GenerateUsername()

	// Username should start with prefix
	assert.Contains(t, username1, UsernamePrefix+"_", "Username should contain prefix")
	assert.Contains(t, username2, UsernamePrefix+"_", "Username should contain prefix")

	// Usernames should be unique
	assert.NotEqual(t, username1, username2, "Usernames should be unique")

	// Username should be in format: prefix_xxxxxxxx (8 digits)
	assert.Regexp(t, `^go_\d{8}$`, username1, "Username should match pattern")
}

// TestCreateUserRequest tests the CreateUserRequest model
func TestCreateUserRequest(t *testing.T) {
	mobile := "13800138000"
	email := "test@example.com"

	req := &CreateUserRequest{
		Mobile: &mobile,
		Email:  &email,
	}

	user := req.ToUser()

	assert.NotNil(t, user.Mobile, "Mobile should not be nil")
	assert.Equal(t, mobile, *user.Mobile, "Mobile should match")
	assert.NotNil(t, user.Email, "Email should not be nil")
	assert.Equal(t, email, *user.Email, "Email should match")
}

// TestSecUIDUniqueness tests that multiple SecUID generations produce unique values
func TestSecUIDUniqueness(t *testing.T) {
	secUIDs := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		secUID := GenerateSecUID()
		assert.False(t, secUIDs[secUID], "SecUID should be unique")
		secUIDs[secUID] = true
	}

	assert.Len(t, secUIDs, iterations, "All SecUIDs should be unique")
}

// TestUsernameUniqueness tests that multiple username generations produce unique values
func TestUsernameUniqueness(t *testing.T) {
	usernames := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		username := GenerateUsername()
		// Note: There's a small chance of collision with random numbers
		// but with 100M possible values, it's extremely unlikely in 100 iterations
		usernames[username] = true
	}

	// We expect most usernames to be unique (allowing for rare collisions)
	assert.GreaterOrEqual(t, len(usernames), 95, "Most usernames should be unique")
}

