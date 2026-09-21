package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite" // CGO-free SQLite driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/platform/logger"
)

// Config holds database configuration
type Config struct {
	Driver          string
	Path            string
	Host            string
	Port            int
	Username        string
	Password        string
	DBName          string
	Charset         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// Init initializes database connection based on driver type
func Init(cfg *Config) (*sqlx.DB, error) {
	var driver, dsn string

	switch cfg.Driver {
	case "mysql":
		if err := createMySQLDatabase(cfg); err != nil {
			return nil, err
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&multiStatements=true",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			cfg.Charset,
		)
		driver = "mysql"
		logger.Log.Info("connecting to MySQL", "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
	case "sqlite":
		fallthrough
	default:
		driver = "sqlite"
		dsn = cfg.Path
		logger.Log.Info("connecting to SQLite", "path", cfg.Path)
	}

	stdDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 20
	}
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 100
	}
	maxLifetime := cfg.ConnMaxLifetime
	if maxLifetime <= 0 {
		maxLifetime = time.Hour
	}

	stdDB.SetMaxIdleConns(maxIdle)
	stdDB.SetMaxOpenConns(maxOpen)
	stdDB.SetConnMaxLifetime(maxLifetime)
	if err := stdDB.Ping(); err != nil {
		_ = stdDB.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	db := sqlx.NewDb(stdDB, driver)
	db.MapperFunc(toSnake)

	logger.Log.Info("database connected successfully", "max_idle_conns", maxIdle, "max_open_conns", maxOpen)
	return db, nil
}

var initialismBoundary = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])|([a-z0-9])([A-Z])`)

func toSnake(name string) string {
	return strings.ToLower(initialismBoundary.ReplaceAllString(name, `${1}${3}_${2}${4}`))
}

// createMySQLDatabase creates the database if it doesn't exist
func createMySQLDatabase(cfg *Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)
	_, err = db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	logger.Log.Info("database is ready", "database", cfg.DBName)
	return nil
}
