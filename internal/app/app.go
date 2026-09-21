package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/banner"
	"go-api-starter/internal/platform/database"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/netutil"
	"go-api-starter/internal/platform/oss"
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

	if cfg.OSS.AccessKeyID != "" && cfg.OSS.AccessKeySecret != "" {
		if err := oss.InitOSS(&cfg.OSS); err != nil {
			logger.Log.Warnf("Failed to initialize OSS: %v", err)
		} else {
			logger.Log.Info("OSS client initialized successfully")
		}
	} else {
		logger.Log.Warn("OSS credentials not configured, OSS features will be disabled")
	}
	setOSSBaseURL(cfg)

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

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	if err := container.Close(); err != nil {
		log.Printf("Container close error: %v", err)
	}
	if err := closeDB(db); err != nil {
		log.Printf("Database close error: %v", err)
	}
	logger.Log.Info("Server exited gracefully")
	return nil
}

func closeDB(db interface{ Close() error }) error {
	return db.Close()
}

func setOSSBaseURL(cfg *config.Config) {
	if cfg.OSS.Domain != "" {
		model.SetOSSBaseURL(cfg.OSS.Domain)
		return
	}
	bucket := cfg.OSS.Bucket
	if bucket == "" {
		bucket = cfg.OSS.BucketName
	}
	if bucket != "" && cfg.OSS.Endpoint != "" {
		model.SetOSSBaseURL("https://" + bucket + "." + cfg.OSS.Endpoint)
	}
}
