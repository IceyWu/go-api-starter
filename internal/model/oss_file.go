package model

import (
	"time"

	"gorm.io/gorm"
)

// OSSFile represents a file uploaded to OSS
type OSSFile struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Key         string         `gorm:"type:varchar(500);not null;uniqueIndex" json:"key" example:"go_oss/uploads/2026-01-15/uuid.jpg"` // OSS file path (relative)
	MD5         string         `gorm:"type:varchar(32);not null;uniqueIndex:idx_md5_user" json:"md5" example:"5d41402abc4b2a76b9719d911017c592"` // File MD5 hash
	FileName    string         `gorm:"type:varchar(255);not null" json:"file_name" example:"example.jpg"`                        // Original file name
	FileSize    int64          `gorm:"not null" json:"file_size" example:"102400"`                                               // File size in bytes
	ContentType string         `gorm:"type:varchar(100)" json:"content_type" example:"image/jpeg"`                               // MIME type
	Extension   string         `gorm:"type:varchar(20)" json:"extension" example:".jpg"`                                         // File extension
	URL         string         `gorm:"-" json:"url" example:"https://cdn.example.com/go_oss/uploads/2026-01-15/uuid.jpg"`         // Access URL (dynamically generated, not stored)
	UserID      uint           `gorm:"index:idx_md5_user" json:"user_id" example:"1"`                                            // Uploader user ID
	Status      int            `gorm:"default:1;index" json:"status" example:"1"`                                                // Status: 1=active, 0=deleted
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (OSSFile) TableName() string {
	return "oss_files"
}
