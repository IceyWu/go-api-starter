package migration

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrator handles database migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// AutoMigrate runs auto migration for given models
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	log.Println("Running database migrations...")

	if m.db.Dialector.Name() == "mysql" {
		m.db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		defer m.db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	}

	if err := m.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// DropTables drops all tables for given models (use with caution)
func (m *Migrator) DropTables(models ...interface{}) error {
	log.Println("Dropping database tables...")

	migrator := m.db.Migrator()
	for _, model := range models {
		if migrator.HasTable(model) {
			if err := migrator.DropTable(model); err != nil {
				return fmt.Errorf("failed to drop table: %w", err)
			}
		}
	}

	log.Println("Database tables dropped successfully")
	return nil
}

// HasTable checks if a table exists
func (m *Migrator) HasTable(model interface{}) bool {
	return m.db.Migrator().HasTable(model)
}

// CreateTable creates a table for the given model
func (m *Migrator) CreateTable(model interface{}) error {
	if m.HasTable(model) {
		return nil
	}

	if err := m.db.Migrator().CreateTable(model); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}
