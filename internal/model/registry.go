package model

// AllModels returns all database models for auto-migration.
// Add new models here instead of modifying main.go.
func AllModels() []interface{} {
	return []interface{}{
		// User & Auth
		&User{},
		&PermissionSpace{},
		&Permission{},
		&Role{},
		&UserRole{},
		&RolePermission{},
		&UserPermissionCache{},

		// File & Upload
		&File{},
		&MultipartUpload{},
		&UploadedPart{},
	}
}
