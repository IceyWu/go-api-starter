package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// File represents a file in the system
type File struct {
	ID     uint   `json:"-" gorm:"primaryKey"`
	SecUID string `json:"sec_uid" gorm:"size:64;uniqueIndex;not null"`
	UserID uint   `json:"-" gorm:"index:idx_files_user_created;uniqueIndex:idx_md5_user;not null"`

	Name    string  `json:"name" gorm:"size:255;not null"`
	Path    *string `json:"path" gorm:"size:500"`
	Type    string  `json:"type" gorm:"size:50;index;not null"` // MIME type
	FileMd5 string  `json:"file_md5" gorm:"size:32;uniqueIndex:idx_md5_user;not null"`
	Size    uint    `json:"size" gorm:"not null"`

	Key       string `json:"key" gorm:"size:500;uniqueIndex;not null"` // OSS 存储路径（相对路径）
	Extension string `json:"extension" gorm:"size:20"`

	// URL 由 Key + BaseURL 动态生成，不存数据库
	URL string `json:"url" gorm:"-"`

	Width     *uint `json:"width"`
	Height    *uint `json:"height"`
	IsPrivate bool  `json:"is_private" gorm:"default:false;index"`

	CreatedAt time.Time `json:"created_at" gorm:"index:idx_files_user_created"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// FileSimpleResponse 简化的文件响应(用于列表等场景)
type FileSimpleResponse struct {
	SecUID    string            `json:"sec_uid"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	FileMd5   string            `json:"file_md5"`
	URL       string            `json:"url"`
	Size      uint              `json:"size"`
	Width     *uint             `json:"width"`
	Height    *uint             `json:"height"`
	IsPrivate bool              `json:"is_private"`
	User      *FileUserResponse `json:"user,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// FileUserResponse 文件列表中的用户精简信息
type FileUserResponse struct {
	SecUID   string  `json:"sec_uid"`
	Username *string `json:"username"`
}

// ToSimpleResponse converts File to FileSimpleResponse
func (f *File) ToSimpleResponse() *FileSimpleResponse {
	resp := &FileSimpleResponse{
		SecUID:    f.SecUID,
		Name:      f.Name,
		Type:      f.Type,
		FileMd5:   f.FileMd5,
		URL:       f.URL,
		Size:      f.Size,
		Width:     f.Width,
		Height:    f.Height,
		IsPrivate: f.IsPrivate,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
	if f.User != nil {
		resp.User = &FileUserResponse{
			SecUID:   f.User.SecUID,
			Username: f.User.Username,
		}
	}
	return resp
}

// BeforeCreate 创建前自动生成 SecUID
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.SecUID == "" {
		f.SecUID = GenerateSecUID()
	}
	return nil
}

// OSSBaseURL is the base URL for constructing full URLs from keys.
// Set once at startup from config (e.g. "https://bucket.oss-cn-hangzhou.aliyuncs.com").
var OSSBaseURL string

// SetOSSBaseURL sets the global base URL for OSS file access.
func SetOSSBaseURL(baseURL string) {
	OSSBaseURL = strings.TrimRight(baseURL, "/")
}

// BuildURL constructs the full URL from the key using the global base URL.
func BuildURL(key string) string {
	if key == "" || OSSBaseURL == "" {
		return key
	}
	return OSSBaseURL + "/" + key
}

// AfterFind GORM hook: auto-populate URL from Key after loading from DB.
func (f *File) AfterFind(tx *gorm.DB) error {
	f.URL = BuildURL(f.Key)
	return nil
}

// CreateFileRequest represents the request body for creating a file
type CreateFileRequest struct {
	Name      string  `json:"name" binding:"required,min=1,max=255"`
	Path      *string `json:"path" binding:"omitempty,max=500"`
	Type      string  `json:"type" binding:"required,max=50"`
	FileMd5   string  `json:"file_md5" binding:"required,len=32"`
	Size      uint    `json:"size" binding:"required"`
	Width     *uint   `json:"width" binding:"omitempty"`
	Height    *uint   `json:"height" binding:"omitempty"`
	IsPrivate bool    `json:"is_private"`
}

// UpdateFileRequest represents the request body for updating a file
type UpdateFileRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=255"`
	IsPrivate *bool   `json:"is_private" binding:"omitempty"`
}

// FileFilter represents filter options for querying files
type FileFilter struct {
	UserID    *uint
	Type      *string
	IsPrivate *bool
}

// ToFile converts CreateFileRequest to File model
func (r *CreateFileRequest) ToFile(userID uint) *File {
	return &File{
		UserID:    userID,
		Name:      r.Name,
		Path:      r.Path,
		Type:      r.Type,
		FileMd5:   r.FileMd5,
		Size:      r.Size,
		Width:     r.Width,
		Height:    r.Height,
		IsPrivate: r.IsPrivate,
	}
}
