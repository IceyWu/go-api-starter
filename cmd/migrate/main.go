package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"go-api-starter/internal/config"
)

func main() {
	cfg := config.Load()
	dir := "migrations/mysql"
	if cfg.Database.Driver == "sqlite" {
		dir = "migrations/sqlite"
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
