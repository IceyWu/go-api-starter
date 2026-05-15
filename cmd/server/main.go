package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/router"
	"go-api-starter/internal/seed"
	"go-api-starter/pkg/banner"
	"go-api-starter/pkg/database"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/migration"
	"go-api-starter/pkg/netutil"
	"go-api-starter/pkg/oss"
)

// @title Go API Starter
// @version 1.0
// @description A RESTful API starter with Go, Gin, and GORM
// @description
// @description **LLMs 入口：**
// @description - [llms.txt](/llms.txt) - AI 可读接口概览
// @description - [llms-full.txt](/llms-full.txt) - AI 可读完整文档
// @host localhost:9527
// @BasePath /

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath)

	// Initialize database
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	migrator := migration.NewMigrator(db)
	if err := migrator.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize OSS client (optional)
	if cfg.OSS.AccessKeyID != "" && cfg.OSS.AccessKeySecret != "" {
		if err := oss.InitOSS(&cfg.OSS); err != nil {
			logger.Log.Warnf("Failed to initialize OSS: %v", err)
		} else {
			logger.Log.Info("OSS client initialized successfully")
		}
	} else {
		logger.Log.Warn("OSS credentials not configured, OSS features will be disabled")
	}

	// Set global OSS base URL for dynamic URL generation from keys
	if cfg.OSS.Domain != "" {
		model.SetOSSBaseURL(cfg.OSS.Domain)
	} else {
		bucket := cfg.OSS.Bucket
		if bucket == "" {
			bucket = cfg.OSS.BucketName
		}
		if bucket != "" && cfg.OSS.Endpoint != "" {
			model.SetOSSBaseURL("https://" + bucket + "." + cfg.OSS.Endpoint)
		}
	}

	// Setup router
	r, permMw, _ := router.Setup(db)

	// Seed permissions defined in route registrations
	seed.SyncPermissions(db, permMw.CollectedCodes())

	// Seed default admin user and role if configured
	if cfg.App.AdminEmail != "" {
		seed.SyncAdminUser(db, cfg.App.AdminEmail, cfg.App.AdminPassword)
		seed.SyncAdminRole(db, cfg.App.AdminEmail)
	}

	// Print banner (empty tools status since there are no external tool dependencies)
	localIP := netutil.GetLocalIP()
	banner.PrintBanner(cfg.App.Name, cfg.App.Env, cfg.Server.Port, localIP, nil)

	// Create HTTP server
	addr := ":" + cfg.Server.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Log.Infof("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Log.Info("Server exited")
}
