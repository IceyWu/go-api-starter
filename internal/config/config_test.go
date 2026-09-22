package config

import (
	"testing"
)

func TestLoadAppliesProfileAndEnvironmentOverrides(t *testing.T) {
	t.Setenv("CONFIG_FILE", "../../config/config.yaml")
	t.Setenv("APP_ENV", "development")
	t.Setenv("GO_API_SERVER__PORT", "9123")
	t.Setenv("GO_API_DATABASE__PORT", "3307")
	t.Setenv("GO_API_REDIS__ENABLED", "true")

	cfg := Load()

	if cfg.App.Env != "development" {
		t.Fatalf("expected development profile, got %q", cfg.App.Env)
	}
	if cfg.Server.Port != "9123" {
		t.Fatalf("expected environment override for server.port, got %q", cfg.Server.Port)
	}
	if cfg.Database.Port != 3307 {
		t.Fatalf("expected environment override for database.port, got %d", cfg.Database.Port)
	}
	if !cfg.Redis.Enabled {
		t.Fatal("expected environment override for redis.enabled")
	}
}

func TestProductionValidationRejectsInsecureDefaults(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Env:                 "production",
			JWTSecret:           "your-secret-key-change-in-production",
			DocsUser:            "admin",
			DocsPassword:        "admin123",
			AdminPassword:       "123456",
			DefaultUserPassword: "123456",
		},
		Server:   ServerConfig{Port: "8080"},
		Database: DatabaseConfig{Driver: "sqlite"},
		CORS:     CORSConfig{AllowOrigins: []string{"*"}},
		Transcoding: TranscodingConfig{
			Enabled:       false,
			MPSRegion:     "",
			MPSPipelineID: "",
		},
	}

	errs := cfg.Validate()
	if !errs.HasErrors() {
		t.Fatal("expected insecure production defaults to be rejected")
	}
	for _, field := range []string{"app.jwt_secret", "app.docs_password", "app.admin_password", "app.default_user_password", "cors.allow_origins"} {
		found := false
		for _, err := range errs {
			if err.Field == field {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected validation error for %s, got %v", field, errs)
		}
	}
}

func TestProductionValidationAcceptsSecureOverrides(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Env:                 "production",
			JWTSecret:           "secure-jwt-secret-that-is-at-least-32-chars",
			DocsUser:            "docs-admin",
			DocsPassword:        "secure-docs-password",
			AdminPassword:       "secure-admin-password",
			DefaultUserPassword: "secure-default-password",
		},
		Server:   ServerConfig{Port: "8080"},
		Database: DatabaseConfig{Driver: "sqlite"},
		CORS:     CORSConfig{AllowOrigins: []string{"https://app.example.com"}},
		Transcoding: TranscodingConfig{
			Enabled:       false,
			MPSRegion:     "",
			MPSPipelineID: "",
		},
	}

	if errs := cfg.Validate(); errs.HasErrors() {
		t.Fatalf("expected secure production configuration without MPS to pass, got %v", errs)
	}
}

func TestProductionValidationRequiresMPSWhenEnabled(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Env:                 "production",
			JWTSecret:           "secure-jwt-secret-that-is-at-least-32-chars",
			DocsUser:            "docs-admin",
			DocsPassword:        "secure-docs-password",
			AdminPassword:       "secure-admin-password",
			DefaultUserPassword: "secure-default-password",
		},
		Server:   ServerConfig{Port: "8080"},
		Database: DatabaseConfig{Driver: "sqlite"},
		CORS:     CORSConfig{AllowOrigins: []string{"https://app.example.com"}},
		Transcoding: TranscodingConfig{
			Enabled:       true,
			MPSRegion:     "",
			MPSPipelineID: "pipeline-id",
		},
	}

	errs := cfg.Validate()
	for _, err := range errs {
		if err.Field == "transcoding.mps_region" {
			return
		}
	}
	t.Fatalf("expected MPS region validation error, got %v", errs)
}
