package router

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"go-api-starter/internal/config"
)

func TestSetupRegistersPublicDocumentationAndSystemRoutes(t *testing.T) {
	previous := config.GlobalConfig
	t.Cleanup(func() { config.GlobalConfig = previous })
	config.GlobalConfig = &config.Config{
		App: config.AppConfig{
			Env:          "development",
			Name:         "go-api-starter",
			DocsUser:     "admin",
			DocsPassword: "admin123",
			JWTSecret:    "test-secret-with-more-than-32-characters",
		},
		Server:   config.ServerConfig{Port: "9527"},
		Database: config.DatabaseConfig{Driver: "sqlite"},
		CORS: config.CORSConfig{
			AllowOrigins: []string{"http://localhost:3000"},
			AllowMethods: []string{"GET", "POST", "OPTIONS"},
			AllowHeaders: []string{"Content-Type", "Authorization"},
		},
		RateLimit: config.RateLimitConfig{FallbackRPS: 100, FallbackBurst: 100},
	}

	stdDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = stdDB.Close() })
	db := sqlx.NewDb(stdDB, "sqlite")
	schema, err := os.ReadFile("../../db/schema.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)
	r, _, _ := Setup(db)

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{name: "system ping", method: http.MethodGet, path: "/api/v1/system/ping", status: http.StatusOK},
		{name: "openapi", method: http.MethodGet, path: "/openapi.json", status: http.StatusOK},
		{name: "llms", method: http.MethodGet, path: "/llms.txt", status: http.StatusOK},
		{name: "health", method: http.MethodGet, path: "/health", status: http.StatusOK},
		{name: "readiness", method: http.MethodGet, path: "/health/ready", status: http.StatusOK},
		{name: "metrics", method: http.MethodGet, path: "/metrics", status: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			require.Equal(t, tt.status, res.Code)
		})
	}

	for _, authorized := range []bool{false, true} {
		t.Run("docs authorization", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/docs", nil)
			if authorized {
				req.SetBasicAuth("admin", "admin123")
			}
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if authorized {
				require.Equal(t, http.StatusOK, res.Code)
				require.Contains(t, res.Header().Get("Content-Type"), "text/html")
			} else {
				require.Equal(t, http.StatusUnauthorized, res.Code)
			}
		})
	}
}
