package model

import (
	"fmt"
	"time"
)

// MultipartUpload represents an ongoing multipart upload task
type MultipartUpload struct {
	ID          uint                  ` json:"id"`
	UploadID    string                ` json:"upload_id"`
	Key         string                ` json:"key"`
	MD5         string                ` json:"md5"`
	FileName    string                ` json:"file_name"`
	FileSize    int64                 ` json:"file_size"`
	ContentType string                ` json:"content_type"`
	TotalParts  int                   ` json:"total_parts"`
	ChunkSize   int64                 ` json:"chunk_size"`
	UserID      uint                  ` json:"user_id"`
	Status      MultipartUploadStatus ` json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
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
	ID         uint      ` json:"id"`
	UploadID   string    ` json:"upload_id"`
	PartNumber int       ` json:"part_number"`
	ETag       string    ` json:"etag"`
	Size       int64     ` json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UploadedPart) TableName() string {
	return "uploaded_parts"
}
