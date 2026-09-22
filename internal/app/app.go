package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/banner"
	"go-api-starter/internal/platform/database"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/netutil"
	"go-api-starter/internal/platform/storage"
	"go-api-starter/internal/router"
	"go-api-starter/internal/seed"
)

// Run initializes the application, serves HTTP traffic, and performs a
// graceful shutdown when the process receives SIGINT or SIGTERM.
func Run() error {
	cfg := config.Load()
	logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath)

	db, err := database.Init(&database.Config{
		Driver:          cfg.Database.Driver,
		Path:            cfg.Database.Path,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		Charset:         cfg.Database.Charset,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if cfg.Storage.AccessKeyID != "" && cfg.Storage.AccessKeySecret != "" {
		if _, err := storage.NewS3Provider(context.Background(), &cfg.Storage); err != nil {
			logger.Log.Warnf("Failed to initialize object storage: %v", err)
		} else {
			logger.Log.Info("object storage client configured successfully")
		}
	} else {
		logger.Log.Warn("object storage credentials not configured, upload features will be disabled")
	}
	setStorageBaseURL(cfg)

	r, permMw, container := router.Setup(db)
	seed.SyncPermissions(db, permMw.CollectedCodes())
	if cfg.App.AdminEmail != "" {
		seed.SyncAdminUser(db, cfg.App.AdminEmail, cfg.App.AdminPassword)
	}
	seed.SyncAdminRole(db, cfg.App.AdminEmail)

	localIPs := netutil.GetAllLocalIPs()
	var tools []banner.ToolInfo
	if cfg.Redis.Enabled {
		redisAddr := cfg.Redis.Addr()
		if rc := container.RedisCache(); rc != nil {
			tools = append(tools, banner.ToolInfo{Name: "Redis", Version: redisAddr, OK: true})
		} else {
			tools = append(tools, banner.ToolInfo{Name: "Redis", Version: redisAddr + " (连接失败, 降级内存)", OK: false})
		}
	} else {
		tools = append(tools, banner.ToolInfo{Name: "Redis", Version: "disabled (使用内存缓存)", OK: false})
	}
	banner.PrintBanner(cfg.App.Name, cfg.App.Env, cfg.Server.Port, cfg.Server.BasePath, localIPs, tools)

	srv := newHTTPServer(cfg, r)
	serverErr := make(chan error, 1)
	go func() {
		logger.Log.Infof("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverErr:
		_ = container.Close()
		_ = closeDB(db)
		return fmt.Errorf("HTTP server failed: %w", err)
	case <-quit:
	}

	logger.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout(cfg))
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("HTTP server forced to shutdown", "error", err)
	}
	if err := container.Close(); err != nil {
		logger.Log.Error("container close error", "error", err)
	}
	if err := closeDB(db); err != nil {
		logger.Log.Error("database close error", "error", err)
	}
	logger.Log.Info("Server exited gracefully")
	return nil
}

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultReadTimeout       = 30 * time.Second
	defaultWriteTimeout      = 60 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultMaxHeaderBytes    = 1 << 20
)

func newHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: durationOrDefault(cfg.Server.ReadHeaderTimeout, defaultReadHeaderTimeout),
		ReadTimeout:       durationOrDefault(cfg.Server.ReadTimeout, defaultReadTimeout),
		WriteTimeout:      durationOrDefault(cfg.Server.WriteTimeout, defaultWriteTimeout),
		IdleTimeout:       durationOrDefault(cfg.Server.IdleTimeout, defaultIdleTimeout),
		MaxHeaderBytes:    intOrDefault(cfg.Server.MaxHeaderBytes, defaultMaxHeaderBytes),
	}
}

func serverShutdownTimeout(cfg *config.Config) time.Duration {
	return durationOrDefault(cfg.Server.ShutdownTimeout, defaultShutdownTimeout)
}

func durationOrDefault(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func intOrDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func closeDB(db interface{ Close() error }) error {
	return db.Close()
}

func setStorageBaseURL(cfg *config.Config) {
	if cfg.Storage.PublicBaseURL != "" {
		model.SetStorageBaseURL(cfg.Storage.PublicBaseURL)
		return
	}
	if cfg.Storage.Bucket != "" && cfg.Storage.Endpoint != "" {
		endpoint := strings.TrimRight(cfg.Storage.Endpoint, "/")
		if parsed, err := url.Parse(endpoint); err == nil && parsed.Host != "" {
			if cfg.Storage.ForcePathStyle {
				parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + cfg.Storage.Bucket
			} else if !strings.HasPrefix(parsed.Host, cfg.Storage.Bucket+".") {
				parsed.Host = cfg.Storage.Bucket + "." + parsed.Host
			}
			model.SetStorageBaseURL(parsed.String())
			return
		}
		model.SetStorageBaseURL(endpoint)
	}
}
