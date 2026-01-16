package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MultipartUpload represents an ongoing multipart upload task
type MultipartUpload struct {
	ID          uint                  `gorm:"primarykey" json:"id"`
	UploadID    string                `gorm:"type:varchar(100);not null;uniqueIndex" json:"upload_id"`
	Key         string                `gorm:"type:varchar(500);not null" json:"key"`
	MD5         string                `gorm:"type:varchar(32);not null;index" json:"md5"`
	FileName    string                `gorm:"type:varchar(255);not null" json:"file_name"`
	FileSize    int64                 `gorm:"not null" json:"file_size"`
	ContentType string                `gorm:"type:varchar(100)" json:"content_type"`
	TotalParts  int                   `gorm:"not null" json:"total_parts"`
	ChunkSize   int64                 `gorm:"not null" json:"chunk_size"`
	UserID      uint                  `gorm:"index" json:"user_id"`
	Status      MultipartUploadStatus `gorm:"type:varchar(20);default:initiated;index" json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (MultipartUpload) TableName() string {
	return "multipart_uploads"
}

// UpdateStatus updates the status with transition validation
func (m *MultipartUpload) UpdateStatus(newStatus MultipartUploadStatus) error {
	if !m.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", m.Status, newStatus)
	}
	m.Status = newStatus
	return nil
}

// UploadedPart represents a successfully uploaded part
type UploadedPart struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UploadID   string    `gorm:"type:varchar(100);not null;index:idx_upload_part,unique" json:"upload_id"`
	PartNumber int       `gorm:"not null;index:idx_upload_part,unique" json:"part_number"`
	ETag       string    `gorm:"type:varchar(100);not null" json:"etag"`
	Size       int64     `gorm:"not null" json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UploadedPart) TableName() string {
	return "uploaded_parts"
}
