package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/logger"

	"github.com/gin-gonic/gin"
)

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	// 需要记录的路径模式
	IncludePaths []string
	// 排除的路径模式
	ExcludePaths []string
	// 最大请求体大小(字节)
	MaxReqBodySize int
	// 最大响应体大小(字节)
	MaxRespBodySize int
	// 敏感字段(会被脱敏)
	SensitiveFields []string
}

// DefaultOperationLogConfig 默认配置
func DefaultOperationLogConfig() *OperationLogConfig {
	return &OperationLogConfig{
		IncludePaths: []string{
			"/api/v1/users",
			"/api/v1/permissions",
			"/api/v1/auth/register",
			"/api/v1/auth/reset-password",
			"/api/v1/oss",
		},
		ExcludePaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/me",
			"/health",
			"/swagger",
		},
		MaxReqBodySize:  2048,
		MaxRespBodySize: 1024,
		SensitiveFields: []string{"password", "token", "secret", "code"},
	}
}

// OperationLogMiddleware 操作日志中间件
type OperationLogMiddleware struct {
	service     service.OperationLogServiceInterface
	userRepo    repository.UserRepositoryInterface
	config      *OperationLogConfig
	moduleMap   map[string]string
	actionMap   map[string]string
}

// NewOperationLogMiddleware 创建操作日志中间件
func NewOperationLogMiddleware(
	svc service.OperationLogServiceInterface,
	userRepo repository.UserRepositoryInterface,
	config *OperationLogConfig,
) *OperationLogMiddleware {
	if config == nil {
		config = DefaultOperationLogConfig()
	}
	return &OperationLogMiddleware{
		service:  svc,
		userRepo: userRepo,
		config:   config,
		moduleMap: map[string]string{
			"/api/v1/users":       "用户管理",
			"/api/v1/permissions": "权限管理",
			"/api/v1/auth":        "认证",
			"/api/v1/oss":         "文件存储",
		},
		actionMap: map[string]string{
			"POST":   "创建",
			"PUT":    "更新",
			"DELETE": "删除",
			"GET":    "查询",
		},
	}
}

// responseWriter 包装响应写入器以捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Log 返回操作日志中间件
func (m *OperationLogMiddleware) Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 检查是否需要记录
		if !m.shouldLog(path, method) {
			c.Next()
			return
		}

		start := time.Now()

		// 读取请求体
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			reqBody = m.sanitizeBody(string(bodyBytes))
			if len(reqBody) > m.config.MaxReqBodySize {
				reqBody = reqBody[:m.config.MaxReqBodySize] + "...[truncated]"
			}
		}

		// 包装响应写入器
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = rw

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start).Milliseconds()

		// 获取响应体
		respBody := rw.body.String()
		if len(respBody) > m.config.MaxRespBodySize {
			respBody = respBody[:m.config.MaxRespBodySize] + "...[truncated]"
		}

		// 获取错误信息
		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		// 构建日志记录
		log := &model.OperationLog{
			Module:     m.getModule(path),
			Action:     m.getAction(method, path),
			Method:     method,
			Path:       path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			RequestID:  GetRequestID(c),
			StatusCode: c.Writer.Status(),
			Latency:    latency,
			ReqBody:    reqBody,
			RespBody:   respBody,
			Error:      errMsg,
		}

		// 获取用户信息（如果已认证）
		if userID, exists := c.Get("userID"); exists {
			uid := userID.(uint)
			log.UserID = &uid
			// 异步获取用户详情
			go m.fillUserInfo(log, uid)
		} else {
			// 同步保存（无用户信息）
			go m.saveLog(log)
		}
	}
}

// shouldLog 判断是否需要记录
func (m *OperationLogMiddleware) shouldLog(path, method string) bool {
	// GET 请求默认不记录（除非明确包含）
	if method == "GET" {
		return false
	}

	// 检查排除路径
	for _, p := range m.config.ExcludePaths {
		if strings.HasPrefix(path, p) {
			return false
		}
	}

	// 检查包含路径
	for _, p := range m.config.IncludePaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// getModule 获取模块名称
func (m *OperationLogMiddleware) getModule(path string) string {
	for prefix, name := range m.moduleMap {
		if strings.HasPrefix(path, prefix) {
			return name
		}
	}
	return "其他"
}

// getAction 获取操作类型
func (m *OperationLogMiddleware) getAction(method, path string) string {
	// 特殊路径处理
	if strings.Contains(path, "reset-password") {
		return "重置密码"
	}
	if strings.Contains(path, "register") {
		return "注册"
	}
	if strings.Contains(path, "roles") && method == "POST" && strings.Contains(path, "/permissions") {
		return "分配权限"
	}

	if action, ok := m.actionMap[method]; ok {
		return action
	}
	return method
}

// sanitizeBody 脱敏请求体
func (m *OperationLogMiddleware) sanitizeBody(body string) string {
	for _, field := range m.config.SensitiveFields {
		// 匹配 "field": "value" 或 "field":"value"
		pattern := regexp.MustCompile(`"` + field + `"\s*:\s*"[^"]*"`)
		body = pattern.ReplaceAllString(body, `"`+field+`":"***"`)
	}
	return body
}

// fillUserInfo 填充用户信息并保存
func (m *OperationLogMiddleware) fillUserInfo(log *model.OperationLog, userID uint) {
	if m.userRepo != nil {
		user, err := m.userRepo.FindByID(nil, userID)
		if err == nil && user != nil {
			log.UserName = user.Name
			log.UserEmail = user.Email
		}
	}
	m.saveLog(log)
}

// saveLog 保存日志
func (m *OperationLogMiddleware) saveLog(log *model.OperationLog) {
	if err := m.service.Create(nil, log); err != nil {
		logger.Errorf("保存操作日志失败: %v", err)
	}
}
