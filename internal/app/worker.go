package app

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/container"
	"go-api-starter/internal/platform/database"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/storage"
	"go-api-starter/internal/service"
)

// RunWorker runs the MPS polling worker independently from the HTTP server.
// Deploy exactly one worker process per environment, or use the database lock
// below when multiple worker replicas are intentionally deployed.
func RunWorker() error {
	cfg := config.Load()
	logger.Init(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.FilePath)
	if !cfg.Transcoding.Enabled {
		logger.Log.Info("Alibaba Cloud MPS worker is disabled")
		return nil
	}

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
	defer closeDB(db)

	if cfg.Storage.AccessKeyID != "" && cfg.Storage.AccessKeySecret != "" {
		if _, err := storage.NewS3Provider(context.Background(), &cfg.Storage); err != nil {
			logger.Log.Warn("failed to initialize object storage", "error", err)
		}
	}

	releaseLock, err := acquireWorkerLock(db, cfg.Database.Driver)
	if err != nil {
		return err
	}
	defer releaseLock()

	workerContainer := container.NewContainer(db, cfg)
	defer workerContainer.Close()
	manager := workerContainer.TranscodingTaskManager()
	transcoder := manager.CloudTranscoder()
	if transcoder == nil {
		return fmt.Errorf("Alibaba Cloud MPS is not configured")
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	poller := service.NewMPSTaskPoller(
		manager,
		transcoder,
		service.NewWebhookNotifier(),
		time.Duration(cfg.Transcoding.MPSPollIntervalSec)*time.Second,
	)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		poller.Start(workerCtx)
	}()
	logger.Log.Info("Alibaba Cloud MPS worker started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	<-quit
	cancelWorker()
	<-workerDone
	logger.Log.Info("Alibaba Cloud MPS worker stopped")
	return nil
}

func acquireWorkerLock(db *sqlx.DB, driver string) (func(), error) {
	if driver != "mysql" {
		logger.Log.Warn("MPS worker database lock is only available for MySQL; run a single worker for this driver")
		return func() {}, nil
	}

	const lockName = "go_api_starter:mps_worker"
	var acquired int
	if err := db.QueryRow("SELECT GET_LOCK(?, 0)", lockName).Scan(&acquired); err != nil {
		return nil, fmt.Errorf("failed to acquire MPS worker lock: %w", err)
	}
	if acquired != 1 {
		return nil, fmt.Errorf("another MPS worker already owns the worker lock")
	}

	return func() {
		if _, err := db.Exec("SELECT RELEASE_LOCK(?)", lockName); err != nil {
			logger.Log.Warn("failed to release MPS worker lock", "error", err)
		}
	}, nil
}
