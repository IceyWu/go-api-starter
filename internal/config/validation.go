package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Validate validates the configuration and returns any validation errors
// In production environment, certain security requirements must be met
func (c *Config) Validate() ValidationErrors {
	var errors ValidationErrors

	// Only enforce strict validation in production
	if c.App.Env == "production" || c.App.Env == "prod" {
		// JWT Secret validation
		if c.App.JWTSecret == "" {
			errors = append(errors, ValidationError{
				Field:   "app.jwt_secret",
				Message: "JWT secret must be set in production",
			})
		} else if c.App.JWTSecret == "your-secret-key-change-in-production" {
			errors = append(errors, ValidationError{
				Field:   "app.jwt_secret",
				Message: "JWT secret must be changed from default value in production",
			})
		} else if len(c.App.JWTSecret) < 32 {
			errors = append(errors, ValidationError{
				Field:   "app.jwt_secret",
				Message: "JWT secret should be at least 32 characters in production",
			})
		}

		// Database validation for non-SQLite databases
		if c.Database.Driver != "sqlite" {
			if c.Database.Password == "" {
				errors = append(errors, ValidationError{
					Field:   "database.password",
					Message: "Database password must be set in production",
				})
			} else if c.Database.Password == "123456" || c.Database.Password == "password" {
				errors = append(errors, ValidationError{
					Field:   "database.password",
					Message: "Database password must be changed from default value in production",
				})
			}

			if c.Database.Host == "" || c.Database.Host == "localhost" {
				errors = append(errors, ValidationError{
					Field:   "database.host",
					Message: "Database host should be explicitly configured in production",
				})
			}
		}

		// OSS validation if endpoint is configured
		if c.OSS.Endpoint != "" {
			if c.OSS.AccessKeyID == "" {
				errors = append(errors, ValidationError{
					Field:   "oss.access_key_id",
					Message: "OSS access key ID must be configured when OSS is enabled in production",
				})
			}
			if c.OSS.AccessKeySecret == "" {
				errors = append(errors, ValidationError{
					Field:   "oss.access_key_secret",
					Message: "OSS access key secret must be configured when OSS is enabled in production",
				})
			}
			if c.OSS.BucketName == "" {
				errors = append(errors, ValidationError{
					Field:   "oss.bucket_name",
					Message: "OSS bucket name must be configured when OSS is enabled in production",
				})
			}
		}
	}

	// General validation (all environments)
	if c.App.Port <= 0 || c.App.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "app.port",
			Message: "Port must be between 1 and 65535",
		})
	}

	return errors
}

// MustValidate validates the configuration and panics if there are errors
// This should be called during application startup
func (c *Config) MustValidate() {
	errors := c.Validate()
	if errors.HasErrors() {
		panic(fmt.Sprintf("Configuration validation failed: %s", errors.Error()))
	}
}
