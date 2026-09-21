package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TranscodingTask represents a video transcoding task
type TranscodingTask struct {
	ID             uint           `json:"-"`
	TaskID         string         `json:"task_id"`
	FileID         *uint          `json:"file_id,omitempty"` // 关联的文件 ID,用于保存转码结果到 video_variants
	SourceURL      string         `json:"source_url"`
	WebhookURL     string         `json:"webhook_url"` // Webhook callback URL
	Provider       string         `json:"provider"`    // aliyun_mps
	Status         string         `json:"status"`      // processing, success, partial_success, failed
	Resolutions    JSONStringList `json:"resolutions"` // ["1080p", "720p", "480p"]
	ExternalJobIDs JSONStringList `json:"-"`           // Cloud provider job IDs, aligned with Resolutions
	Results        JSONResults    `json:"results"`     // Array of TranscodingResult
	ErrorMsg       string         `json:"error_msg"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	DeletedAt      *time.Time     `json:"-"`
}

// TableName specifies the table name
func (TranscodingTask) TableName() string {
	return "transcoding_tasks"
}

// TranscodingResult represents the result of transcoding a single resolution
type TranscodingResult struct {
	Resolution string `json:"resolution"` // 1080p, 720p, 480p
	Status     string `json:"status"`     // success, failed
	URL        string `json:"url,omitempty"`
	Size       int64  `json:"size,omitempty"` // 文件大小（字节）
	Error      string `json:"error,omitempty"`
}

// JSONStringList is a custom type for storing string arrays as JSON
type JSONStringList []string

// Scan implements the sql.Scanner interface
func (j *JSONStringList) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(data, j)
}

// Value implements the driver.Valuer interface
func (j JSONStringList) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}

// JSONResults is a custom type for storing TranscodingResult arrays as JSON
type JSONResults []TranscodingResult

// Scan implements the sql.Scanner interface
func (j *JSONResults) Scan(value interface{}) error {
	if value == nil {
		*j = []TranscodingResult{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(data, j)
}

// Value implements the driver.Valuer interface
func (j JSONResults) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}

// CreateTaskRequest represents the request body for creating a transcoding task
type CreateTaskRequest struct {
	SourceURL   string   `json:"source_url" binding:"required,url" example:"https://example.com/video.mov"`
	Resolutions []string `json:"resolutions" binding:"omitempty" example:"original,1080p,720p,480p"`                                  // 可选，不传则使用默认值(original为原画质格式转换)
	WebhookURL  string   `json:"webhook_url" binding:"omitempty,url" example:"https://example.com/api/webhooks/transcoding/callback"` // 可选，转码完成后回调
	FileID      *uint    `json:"file_id,omitempty"`                                                                                   // 内部使用,关联的文件 ID
}

// CreateTaskResponse represents the response for creating a transcoding task
type CreateTaskResponse struct {
	TaskID string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status string `json:"status" example:"processing"`
}

// TaskStatusResponse represents the response for querying task status
type TaskStatusResponse struct {
	TaskID      string              `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status      string              `json:"status" example:"success"`
	Provider    string              `json:"provider" example:"aliyun_mps"`
	SourceURL   string              `json:"source_url" example:"https://example.com/video.mov"`
	Results     []TranscodingResult `json:"results"`
	ErrorMsg    string              `json:"error_msg,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
}

// ToTaskStatusResponse converts TranscodingTask to TaskStatusResponse
func (t *TranscodingTask) ToTaskStatusResponse() *TaskStatusResponse {
	return &TaskStatusResponse{
		TaskID:      t.TaskID,
		Status:      t.Status,
		Provider:    t.Provider,
		SourceURL:   t.SourceURL,
		Results:     t.Results,
		ErrorMsg:    t.ErrorMsg,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}
