package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// UsernamePrefix 用户账号前缀，默认 "go"，可通过配置修改
var UsernamePrefix = "go"

// User represents the user model
type User struct {
	ID               uint    `json:"-" gorm:"primaryKey"`                         // 内部ID，不对外暴露
	SecUID           string  `json:"sec_uid" gorm:"size:64;uniqueIndex;not null"` // 安全标识符，对外暴露
	LPID             string  `json:"lp_id" gorm:"size:20;uniqueIndex;not null"`   // LP号，类似抖音号
	Username         *string `json:"username,omitempty" gorm:"size:100;index"`    // 用户账号
	Mobile           *string `json:"mobile,omitempty" gorm:"size:20;uniqueIndex"` // 手机号，可选
	Email            *string `json:"email,omitempty" gorm:"size:50;uniqueIndex"`  // 邮箱，可选
	Password         *string `json:"-" gorm:"size:255"`                           // Password hash, not exposed in JSON
	AvatarFileID     *uint   `json:"-" gorm:"index"`                              // 头像文件ID
	BackgroundFileID *uint   `json:"-" gorm:"index"`                              // 背景文件ID

	// Relations
	AvatarFile     *File  `json:"avatar_file,omitempty" gorm:"foreignKey:AvatarFileID"`                      // 头像文件对象
	BackgroundFile *File  `json:"background_file,omitempty" gorm:"foreignKey:BackgroundFileID"`              // 背景文件对象
	Roles          []Role `json:"-" gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID"` // RBAC 角色

	Sex       int            `json:"sex" gorm:"default:0"` // 性别: 0-未知, 1-男, 2-女
	Birthday  *time.Time     `json:"birthday,omitempty"`
	City      *string        `json:"city,omitempty"`
	Job       *string        `json:"job,omitempty"`
	Company   *string        `json:"company,omitempty"`
	Signature *string        `json:"signature,omitempty"`
	Website   *string        `json:"website,omitempty"`
	Freezed   bool           `json:"freezed" gorm:"default:false;index"` // 是否冻结
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate 创建前自动生成 SecUID、Username 和 LPID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SecUID == "" {
		u.SecUID = GenerateSecUID()
	}
	if u.Username == nil || *u.Username == "" {
		username := GenerateUsername()
		u.Username = &username
	}
	if u.LPID == "" {
		// 最多重试 5 次确保 LPID 唯一
		for i := 0; i < 5; i++ {
			lpID := GenerateLPID()
			var count int64
			tx.Model(&User{}).Where("lp_id = ?", lpID).Count(&count)
			if count == 0 {
				u.LPID = lpID
				break
			}
		}
		if u.LPID == "" {
			u.LPID = GenerateLPID() // fallback，依赖 uniqueIndex 兜底
		}
	}
	return nil
}

// GenerateSecUID 生成安全标识符 (22字符，URL安全的base64)
func GenerateSecUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// GenerateUsername 生成默认用户账号 (prefix_xxxxxxxx，8位随机数字)
func GenerateUsername() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000000))
	return fmt.Sprintf("%s_%08d", UsernamePrefix, n.Int64())
}

// GenerateLPID 生成 LP 号 (LP_XXXXXXXXXX，10位随机数字)
func GenerateLPID() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000000000))
	return fmt.Sprintf("LP_%010d", n.Int64())
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,min=1,max=50" example:"john_doe"`
	Mobile   *string `json:"mobile" binding:"omitempty,len=11" example:"13800138000"`
	Email    *string `json:"email" binding:"omitempty,email" example:"john@example.com"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username         *string    `json:"username" binding:"omitempty,min=1,max=50" example:"john_doe"`
	LPID             *string    `json:"lp_id" binding:"omitempty,min=3,max=20" example:"LP_8806386288"`
	Mobile           *string    `json:"mobile" binding:"omitempty,len=11" example:"13800138000"`
	Email            *string    `json:"email" binding:"omitempty,email" example:"john@example.com"`
	AvatarSecUID     *string    `json:"avatar_sec_uid" binding:"omitempty" example:"abc123"`
	BackgroundSecUID *string    `json:"background_sec_uid" binding:"omitempty" example:"def456"`
	Sex              *int       `json:"sex" binding:"omitempty,min=0,max=2" example:"1"`
	Birthday         *time.Time `json:"birthday" example:"1990-01-01T00:00:00Z"`
	City             *string    `json:"city" binding:"omitempty,max=100" example:"Beijing"`
	Job              *string    `json:"job" binding:"omitempty,max=100" example:"Software Engineer"`
	Company          *string    `json:"company" binding:"omitempty,max=100" example:"Tech Corp"`
	Signature        *string    `json:"signature" binding:"omitempty,max=500" example:"Hello World"`
	Website          *string    `json:"website" binding:"omitempty,url" example:"https://example.com"`
}

// ToUser converts CreateUserRequest to User model
func (r *CreateUserRequest) ToUser() *User {
	return &User{
		Username: r.Username,
		Mobile:   r.Mobile,
		Email:    r.Email,
	}
}

// AvatarFileResponse 头像/背景文件的精简响应
type AvatarFileResponse struct {
	SecUID string `json:"sec_uid"`
	URL    string `json:"url"`
}

// UserResponse 用户 API 响应 DTO
type UserResponse struct {
	SecUID         string              `json:"sec_uid"`
	LPID           string              `json:"lp_id"`
	Username       *string             `json:"username"`
	Mobile         *string             `json:"mobile,omitempty"`
	Email          *string             `json:"email"`
	AvatarFile     *AvatarFileResponse `json:"avatar_file"`
	BackgroundFile *AvatarFileResponse `json:"background_file"`
	Roles          []string            `json:"roles"`
	Sex            int                 `json:"sex"`
	Birthday       *time.Time          `json:"birthday"`
	City           *string             `json:"city"`
	Job            *string             `json:"job"`
	Company        *string             `json:"company"`
	Signature      *string             `json:"signature"`
	Website        *string             `json:"website"`
	Freezed        bool                `json:"freezed"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// ToResponse 将 User model 转换为 API 响应
func (u *User) ToResponse() *UserResponse {
	roleNames := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleNames = append(roleNames, r.Name)
	}

	resp := &UserResponse{
		SecUID:    u.SecUID,
		LPID:      u.LPID,
		Username:  u.Username,
		Email:     u.Email,
		Roles:     roleNames,
		Sex:       u.Sex,
		Birthday:  u.Birthday,
		City:      u.City,
		Job:       u.Job,
		Company:   u.Company,
		Signature: u.Signature,
		Website:   u.Website,
		Freezed:   u.Freezed,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.AvatarFile != nil {
		resp.AvatarFile = &AvatarFileResponse{
			SecUID: u.AvatarFile.SecUID,
			URL:    u.AvatarFile.URL,
		}
	}
	if u.BackgroundFile != nil {
		resp.BackgroundFile = &AvatarFileResponse{
			SecUID: u.BackgroundFile.SecUID,
			URL:    u.BackgroundFile.URL,
		}
	}
	return resp
}

// ToUserResponseList 批量转换
func ToUserResponseList(users []User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i := range users {
		result[i] = users[i].ToResponse()
	}
	return result
}
