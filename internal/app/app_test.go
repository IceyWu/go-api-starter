package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-api-starter/internal/config"
)

func TestNewHTTPServerUsesSafeTimeoutDefaults(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{Port: "9527"}}
	srv := newHTTPServer(cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	require.Equal(t, ":9527", srv.Addr)
	require.Equal(t, defaultReadHeaderTimeout, srv.ReadHeaderTimeout)
	require.Equal(t, defaultReadTimeout, srv.ReadTimeout)
	require.Equal(t, defaultWriteTimeout, srv.WriteTimeout)
	require.Equal(t, defaultIdleTimeout, srv.IdleTimeout)
	require.Equal(t, defaultMaxHeaderBytes, srv.MaxHeaderBytes)
}

func TestNewHTTPServerHonorsConfiguredTimeouts(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{
		Port:              "9527",
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      4 * time.Second,
		IdleTimeout:       5 * time.Second,
		MaxHeaderBytes:    4096,
	}}
	srv := newHTTPServer(cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	require.Equal(t, 2*time.Second, srv.ReadHeaderTimeout)
	require.Equal(t, 3*time.Second, srv.ReadTimeout)
	require.Equal(t, 4*time.Second, srv.WriteTimeout)
	require.Equal(t, 5*time.Second, srv.IdleTimeout)
	require.Equal(t, 4096, srv.MaxHeaderBytes)
}

func TestHTTPServerGracefulShutdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, srv.Config.Shutdown(ctx))
}
