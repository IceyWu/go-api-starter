package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/platform/database"
)

func main() {
	cfg := config.Load()
	dir := "migrations/mysql"
	switch cfg.Database.Driver {
	case "sqlite":
		dir = "migrations/sqlite"
	case "mysql":
		if err := database.EnsureMySQLDatabase(&database.Config{
			Driver:         cfg.Database.Driver,
			Host:           cfg.Database.Host,
			Port:           cfg.Database.Port,
			Username:       cfg.Database.Username,
			Password:       cfg.Database.Password,
			DBName:         cfg.Database.DBName,
			Charset:        cfg.Database.Charset,
			ConnectTimeout: time.Duration(cfg.Database.ConnectTimeout) * time.Second,
			ReadTimeout:    time.Duration(cfg.Database.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.Database.WriteTimeout) * time.Second,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize MySQL database: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unsupported database driver %q\n", cfg.Database.Driver)
		os.Exit(1)
	}
	cmd := exec.Command(
		"atlas", "migrate", "apply",
		"--dir", "file://"+dir,
		"--url", databaseURL(cfg),
	)
	if cfg.Database.AllowDirtyMigrations {
		cmd.Args = append(cmd.Args, "--allow-dirty")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply Atlas migrations: %v\n", err)
		os.Exit(1)
	}
}

func databaseURL(cfg *config.Config) string {
	if cfg.Database.Driver == "sqlite" {
		return "sqlite://" + cfg.Database.Path
	}

	return fmt.Sprintf(
		"mysql://%s:%s@%s:%d/%s?charset=%s",
		url.QueryEscape(cfg.Database.Username),
		url.QueryEscape(cfg.Database.Password),
		cfg.Database.Host,
		cfg.Database.Port,
		url.PathEscape(cfg.Database.DBName),
		url.QueryEscape(cfg.Database.Charset),
	)
}
