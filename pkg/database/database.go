package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	_ "github.com/glebarez/go-sqlite" // CGO-free SQLite driver
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
func Init(cfg *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		if err := createMySQLDatabase(cfg); err != nil {
			return nil, err
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			cfg.Charset,
		)
		dialector = mysql.Open(dsn)
		log.Printf("Connecting to MySQL: %s@%s:%d/%s", cfg.Username, cfg.Host, cfg.Port, cfg.DBName)
	case "sqlite":
		fallthrough
	default:
		dialector = sqlite.Open(cfg.Path)
		log.Printf("Connecting to SQLite: %s", cfg.Path)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

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

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(maxLifetime)

	log.Printf("Database connected successfully!")
	log.Printf("Connection pool: MaxIdle=%d, MaxOpen=%d", maxIdle, maxOpen)
	return db, nil
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

	log.Printf("Database '%s' is ready", cfg.DBName)
	return nil
}
