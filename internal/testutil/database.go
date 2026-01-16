package testutil

import (
	"go-api-starter/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB creates an in-memory SQLite database for testing
func NewTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&model.User{},
		&model.PermissionSpace{},
		&model.Permission{},
		&model.Role{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.UserPermissionCache{},
		&model.OSSFile{},
		&model.MultipartUpload{},
		&model.UploadedPart{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewTestDBWithData creates a test database with pre-populated data
func NewTestDBWithData() (*gorm.DB, error) {
	db, err := NewTestDB()
	if err != nil {
		return nil, err
	}

	// Create test user
	user := NewTestUser()
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	// Create test permission space
	space := NewTestPermissionSpace()
	if err := db.Create(space).Error; err != nil {
		return nil, err
	}

	// Create test permission
	perm := NewTestPermission()
	if err := db.Create(perm).Error; err != nil {
		return nil, err
	}

	// Create test role
	role := NewTestRole()
	if err := db.Create(role).Error; err != nil {
		return nil, err
	}

	return db, nil
}

// CleanupTestDB cleans up all data in the test database
func CleanupTestDB(db *gorm.DB) error {
	// Delete in reverse order of dependencies
	tables := []interface{}{
		&model.UserPermissionCache{},
		&model.RolePermission{},
		&model.UserRole{},
		&model.Permission{},
		&model.PermissionSpace{},
		&model.Role{},
		&model.UploadedPart{},
		&model.MultipartUpload{},
		&model.OSSFile{},
		&model.User{},
	}

	for _, table := range tables {
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// MustNewTestDB creates a test database and panics on error
func MustNewTestDB() *gorm.DB {
	db, err := NewTestDB()
	if err != nil {
		panic(err)
	}
	return db
}
