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
// @description 🚀 基于 Go + Gin + GORM 构建的高性能 RESTful API 脚手架
// @x-logo {"url": "/logo.svg", "altText": "Go API Starter"}
// @description
// @description ---
// @description
// @description ## 🔌 WebSocket
// @description
// @description | 项目 | 说明 |
// @description |------|------|
// @description | 入口 | `GET /ws` — 长连接入口，query 参数 `key` 或 header `X-API-Key` 认证 |
// @description | 消息格式 | `{"type":"...", "id":"...", "data":{...}}` |
// @description | ⬇️ 下行指令 | `send_text_msg`、`get_group_list`、`ping` |
// @description | ⬆️ 上行消息 | `ack`、`review`、`pong` |
// @description
// @description ---
// @description
// @description ## 🤖 LLMs 入口
// @description
// @description | 文件 | 说明 |
// @description |------|------|
// @description | 📄 [llms.txt](/llms.txt) | AI 可读接口概览 |
// @description | 📚 [llms-full.txt](/llms-full.txt) | AI 可读完整文档 |
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
	r, permMw, container := router.Setup(db)

	// Seed permissions defined in route registrations
	seed.SyncPermissions(db, permMw.CollectedCodes())

	// Seed default admin user and role if configured
	if cfg.App.AdminEmail != "" {
		seed.SyncAdminUser(db, cfg.App.AdminEmail, cfg.App.AdminPassword)
		seed.SyncAdminRole(db, cfg.App.AdminEmail)
	}

	// Print banner with service status
	localIPs := netutil.GetAllLocalIPs()
	var tools []banner.ToolInfo
	if cfg.Redis.Enabled {
		redisAddr := cfg.Redis.Addr()
		rc := container.RedisCache()
		if rc != nil {
			tools = append(tools, banner.ToolInfo{Name: "Redis", Version: redisAddr, OK: true})
		} else {
			tools = append(tools, banner.ToolInfo{Name: "Redis", Version: redisAddr + " (连接失败, 降级内存)", OK: false})
		}
	} else {
		tools = append(tools, banner.ToolInfo{Name: "Redis", Version: "disabled (使用内存缓存)", OK: false})
	}
	banner.PrintBanner(cfg.App.Name, cfg.App.Env, cfg.Server.Port, localIPs, tools)

	// Create HTTP server with timeouts to prevent slow-loris attacks
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Shutdown HTTP server (stop accepting new requests, wait for in-flight)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	// 2. Close DI container (Redis, memory cache, etc.)
	if err := container.Close(); err != nil {
		log.Printf("Container close error: %v", err)
	}

	// 3. Close database connection pool
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Database close error: %v", err)
		}
	}

	logger.Log.Info("Server exited gracefully")
}
