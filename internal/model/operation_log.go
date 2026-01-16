package model

import (
	"time"
)

// OperationLog 操作日志
type OperationLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     *uint     `json:"user_id" gorm:"index"`                    // 操作人ID，匿名操作为nil
	UserName   string    `json:"user_name" gorm:"size:100"`               // 操作人名称
	UserEmail  string    `json:"user_email" gorm:"size:100"`              // 操作人邮箱
	Module     string    `json:"module" gorm:"size:50;index"`             // 模块名称
	Action     string    `json:"action" gorm:"size:50;index"`             // 操作类型
	Method     string    `json:"method" gorm:"size:10"`                   // HTTP方法
	Path       string    `json:"path" gorm:"size:255"`                    // 请求路径
	IP         string    `json:"ip" gorm:"size:50"`                       // 客户端IP
	UserAgent  string    `json:"user_agent" gorm:"size:500"`              // User-Agent
	RequestID  string    `json:"request_id" gorm:"size:50;index"`         // 请求ID
	StatusCode int       `json:"status_code"`                             // 响应状态码
	Latency    int64     `json:"latency"`                                 // 耗时(毫秒)
	ReqBody    string    `json:"req_body" gorm:"type:text"`               // 请求体(脱敏)
	RespBody   string    `json:"resp_body" gorm:"type:text"`              // 响应体(截断)
	Error      string    `json:"error" gorm:"type:text"`                  // 错误信息
	CreatedAt  time.Time `json:"created_at" gorm:"index"`                 // 创建时间
}

// TableName 表名
func (OperationLog) TableName() string {
	return "operation_logs"
}
