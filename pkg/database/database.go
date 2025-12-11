package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config holds database configuration
type Config struct {
	Driver   string
	Path     string
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
	Charset  string
}

// Init initializes database connection based on driver type
func Init(cfg *Config) (*gorm.DB, error) {
	var err error
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		// Auto create database if not exists
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

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Database connected successfully!")
	return DB, nil
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

	// Create database if not exists
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)
	_, err = db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	log.Printf("Database '%s' is ready", cfg.DBName)
	return nil
}

// AutoMigrate runs auto migration for given models
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		log.Fatal("Database not initialized")
	}
	return DB.AutoMigrate(models...)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
