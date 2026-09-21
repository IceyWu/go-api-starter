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
