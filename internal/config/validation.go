package config

import (
	"fmt"
	"strconv"
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
// In development environment, critical misconfigurations still raise warnings
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

		if c.App.DocsUser == "" || c.App.DocsPassword == "" {
			errors = append(errors, ValidationError{
				Field:   "app.docs_user/app.docs_password",
				Message: "documentation credentials must be configured in production",
			})
		} else if c.App.DocsPassword == "admin123" || c.App.DocsPassword == "password" {
			errors = append(errors, ValidationError{
				Field:   "app.docs_password",
				Message: "documentation password must be changed from default value in production",
			})
		}

		if c.App.AdminPassword == "" || c.App.AdminPassword == "123456" || c.App.AdminPassword == "password" {
			errors = append(errors, ValidationError{
				Field:   "app.admin_password",
				Message: "admin password must be changed from default value in production",
			})
		}

		if c.App.DefaultUserPassword == "" || c.App.DefaultUserPassword == "123456" || c.App.DefaultUserPassword == "password" {
			errors = append(errors, ValidationError{
				Field:   "app.default_user_password",
				Message: "default user password must be changed from default value in production",
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
			if c.OSS.BucketName == "" && c.OSS.Bucket == "" {
				errors = append(errors, ValidationError{
					Field:   "oss.bucket_name",
					Message: "OSS bucket name must be configured when OSS is enabled in production",
				})
			}
		}

		if c.Transcoding.MPSRegion == "" {
			errors = append(errors, ValidationError{
				Field:   "transcoding.mps_region",
				Message: "MPS region must be configured in production",
			})
		}
		if c.Transcoding.MPSPipelineID == "" {
			errors = append(errors, ValidationError{
				Field:   "transcoding.mps_pipeline_id",
				Message: "MPS pipeline ID must be configured in production",
			})
		}

		if len(c.CORS.AllowOrigins) == 0 || containsWildcard(c.CORS.AllowOrigins) {
			errors = append(errors, ValidationError{
				Field:   "cors.allow_origins",
				Message: "wildcard CORS origins are not allowed in production",
			})
		}
	}

	// Development environment: warn about critical security misconfigurations
	if c.App.Env == "development" || c.App.Env == "dev" {
		if c.App.JWTSecret == "" || c.App.JWTSecret == "your-secret-key-change-in-production" {
			errors = append(errors, ValidationError{
				Field:   "app.jwt_secret",
				Message: "JWT secret is using default value (set GO_API_APP__JWT_SECRET env var to suppress this warning)",
			})
		}
	}

	// General validation (all environments)
	if port, convErr := strconv.Atoi(c.Server.Port); convErr != nil || port <= 0 || port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "server.port",
			Message: "Port must be a number between 1 and 65535",
		})
	}

	return errors
}

func containsWildcard(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "*" {
			return true
		}
	}
	return false
}

// MustValidate validates the configuration and panics if there are errors
// This should be called during application startup
func (c *Config) MustValidate() {
	errors := c.Validate()
	if errors.HasErrors() {
		panic(fmt.Sprintf("Configuration validation failed: %s", errors.Error()))
	}
}
