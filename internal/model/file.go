package model

import (
	"strings"
	"time"
)

// File represents a file in the system
type File struct {
	ID     uint   `json:"-"`
	UID    string `json:"uid"`
	UserID uint   `json:"-"`

	Name    string  `json:"name"`
	Path    *string `json:"path"`
	Type    string  `json:"type"` // MIME type
	FileMd5 string  `json:"file_md5"`
	Size    uint    `json:"size"`

	Key       string `json:"key"` // OSS 存储路径（相对路径）
	Extension string `json:"extension"`

	// URL 由 Key + BaseURL 动态生成，不存数据库
	URL string `json:"url"`

	Width             *uint      `json:"width"`
	Height            *uint      `json:"height"`
	Blurhash          *string    `json:"blurhash"`
	Arthash           *string    `json:"arthash"`
	ArthashCodec      *string    `json:"arthash_codec"`
	Lng               *float64   `json:"lng"`
	Lat               *float64   `json:"lat"`
	Country           *string    `json:"country"`
	CountryCode       *string    `json:"country_code"`
	Province          *string    `json:"province"`
	City              *string    `json:"city"`
	District          *string    `json:"district"`
	Address           *string    `json:"address"`
	Altitude          *float64   `json:"altitude"`
	TakenAt           *time.Time `json:"taken_at"`
	DeviceMake        *string    `json:"device_make"`
	DeviceModel       *string    `json:"device_model"`
	LensModel         *string    `json:"lens_model"`
	FNumber           *string    `json:"f_number"`
	ExposureTime      *string    `json:"exposure_time"`
	ISO               *int       `json:"iso"`
	FocalLength       *string    `json:"focal_length"`
	ExifRaw           []byte     `json:"exif_raw"`
	Duration          *float64   `json:"duration,omitempty"`
	Codec             *string    `json:"codec,omitempty"`
	Bitrate           *uint      `json:"bitrate,omitempty"`
	FrameRate         *float64   `json:"frame_rate,omitempty"`
	VideoMetadata     []byte     `json:"video_metadata,omitempty"`
	TranscodingTaskID *string    `json:"transcoding_task_id,omitempty"`
	IsPrivate         bool       `json:"is_private"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User            *User            `json:"user,omitempty"`
	Colors          []FileColor      `json:"colors,omitempty"`
	VideoVariants   []VideoVariant   `json:"video_variants,omitempty"`
	TranscodingTask *TranscodingTask `json:"transcoding_task,omitempty"`
}

type Color struct {
	ID        uint      `json:"-"`
	Hex       string    `json:"hex"`
	R         uint8     `json:"r"`
	G         uint8     `json:"g"`
	B         uint8     `json:"b"`
	CreatedAt time.Time `json:"created_at"`
}
type FileColor struct {
	ID         uint      `json:"-"`
	FileID     uint      `json:"-"`
	ColorID    uint      `json:"-"`
	IsPrimary  bool      `json:"is_primary"`
	Rank       uint8     `json:"rank"`
	Percentage *float64  `json:"percentage,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	Color      Color     `json:"color"`
}
type VideoVariant struct {
	ID        uint      `json:"-"`
	FileID    uint      `json:"-"`
	Quality   string    `json:"quality"`
	Key       string    `json:"key"`
	Format    string    `json:"format"`
	Size      *uint     `json:"size,omitempty"`
	Width     *uint     `json:"width,omitempty"`
	Height    *uint     `json:"height,omitempty"`
	Bitrate   *uint     `json:"bitrate,omitempty"`
	FPS       *uint     `json:"fps,omitempty"`
	Duration  *float64  `json:"duration,omitempty"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

func (v *VideoVariant) PrepareForResponse() { v.URL = BuildURL(v.Key) }

// FileSimpleResponse 简化的文件响应(用于列表等场景)
type FileSimpleResponse struct {
	UID       string            `json:"uid"`
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
	UID      string  `json:"uid"`
	Username *string `json:"username"`
}

// ToSimpleResponse converts File to FileSimpleResponse
func (f *File) ToSimpleResponse() *FileSimpleResponse {
	resp := &FileSimpleResponse{
		UID:       f.UID,
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
			UID:      f.User.UID,
			Username: f.User.Username,
		}
	}
	return resp
}

// BeforeCreate 创建前自动生成 UID
func (f *File) PrepareForCreate() {
	if f.UID == "" {
		f.UID = GenerateUID()
	}
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

// AfterFind SQL hook: auto-populate URL from Key after loading from DB.
func (f *File) PrepareForResponse() {
	f.URL = BuildURL(f.Key)
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
