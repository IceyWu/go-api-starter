package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite" // CGO-free SQLite driver
	mysqldriver "github.com/go-sql-driver/mysql"
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
	ConnectTimeout  time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// Init initializes database connection based on driver type
func Init(cfg *Config) (*sqlx.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is required")
	}
	var driver, dsn string

	switch cfg.Driver {
	case "mysql":
		if err := validateMySQLConfig(cfg); err != nil {
			return nil, err
		}
		if err := EnsureMySQLDatabase(cfg); err != nil {
			return nil, err
		}
		dsn = mysqlDSN(cfg, cfg.DBName)
		driver = "mysql"
		logger.Log.Info("connecting to MySQL", "host", cfg.Host, "port", cfg.Port, "database", cfg.DBName)
	case "sqlite":
		driver = "sqlite"
		dsn = cfg.Path
		logger.Log.Info("connecting to SQLite", "path", cfg.Path)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
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
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout(cfg))
	defer cancel()
	if err := stdDB.PingContext(ctx); err != nil {
		_ = stdDB.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	db := sqlx.NewDb(stdDB, driver)
	db.MapperFunc(toSnake)

	logger.Log.Info("database connected successfully", "max_idle_conns", maxIdle, "max_open_conns", maxOpen)
	return db, nil
}

var mysqlIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func validateMySQLConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("database config is required")
	}
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("MySQL host is required")
	}
	if strings.TrimSpace(cfg.Username) == "" {
		return fmt.Errorf("MySQL username is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("MySQL port must be between 1 and 65535")
	}
	if len(cfg.DBName) > 64 || !mysqlIdentifierPattern.MatchString(cfg.DBName) {
		return fmt.Errorf("invalid MySQL database name %q: use 1-64 letters, digits, or underscores", cfg.DBName)
	}
	for _, reserved := range []string{"information_schema", "mysql", "performance_schema", "sys"} {
		if strings.EqualFold(cfg.DBName, reserved) {
			return fmt.Errorf("MySQL database name %q is reserved", cfg.DBName)
		}
	}
	if cfg.Charset != "" && !mysqlIdentifierPattern.MatchString(cfg.Charset) {
		return fmt.Errorf("invalid MySQL charset %q", cfg.Charset)
	}
	return nil
}

func mysqlDSN(cfg *Config, databaseName string) string {
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	dsnConfig := mysqldriver.NewConfig()
	dsnConfig.User = cfg.Username
	dsnConfig.Passwd = cfg.Password
	dsnConfig.Net = "tcp"
	dsnConfig.Addr = net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dsnConfig.DBName = databaseName
	dsnConfig.ParseTime = true
	dsnConfig.Loc = time.Local
	dsnConfig.MultiStatements = true
	dsnConfig.Timeout = connectTimeout(cfg)
	dsnConfig.ReadTimeout = cfg.ReadTimeout
	if dsnConfig.ReadTimeout <= 0 {
		dsnConfig.ReadTimeout = 30 * time.Second
	}
	dsnConfig.WriteTimeout = cfg.WriteTimeout
	if dsnConfig.WriteTimeout <= 0 {
		dsnConfig.WriteTimeout = 30 * time.Second
	}
	dsnConfig.Params = map[string]string{"charset": charset}
	return dsnConfig.FormatDSN()
}

func connectTimeout(cfg *Config) time.Duration {
	if cfg.ConnectTimeout > 0 {
		return cfg.ConnectTimeout
	}
	return 10 * time.Second
}

// EnsureMySQLDatabase creates the configured schema only when it is absent.
// It never drops, truncates, or changes an existing database.
func EnsureMySQLDatabase(cfg *Config) error {
	if err := validateMySQLConfig(cfg); err != nil {
		return err
	}
	db, err := sql.Open("mysql", mysqlDSN(cfg, ""))
	if err != nil {
		return fmt.Errorf("open MySQL connection for database initialization: %w", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout(cfg))
	defer cancel()
	var existing string
	err = db.QueryRowContext(ctx,
		"SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?",
		cfg.DBName,
	).Scan(&existing)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("check MySQL database %q: %w", cfg.DBName, err)
	}

	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", cfg.DBName, charset)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create MySQL database %q: %w", cfg.DBName, err)
	}
	return nil
}

var initialismBoundary = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])|([a-z0-9])([A-Z])`)

func toSnake(name string) string {
	return strings.ToLower(initialismBoundary.ReplaceAllString(name, `${1}${3}_${2}${4}`))
}
