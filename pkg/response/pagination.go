package response

import (
	"math"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Pagination 分页请求参数
type Pagination struct {
	Page     int    `form:"page" json:"page"`           // 当前页码，从1开始
	PageSize int    `form:"page_size" json:"page_size"` // 每页数量
	Sort     string `form:"sort" json:"sort"`           // 排序: field,asc 或 asc,field
}

// PageResult 通用分页响应结构
type PageResult[T any] struct {
	List       []T   `json:"list"`        // 数据列表
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页数量
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// UserPageResult Swagger 文档用
type UserPageResult struct {
	List       []interface{} `json:"list"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Total      int64         `json:"total"`
	TotalPages int           `json:"total_pages"`
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	DefaultSort     = "id desc"
)

// AllowedSortFields 允许排序的字段映射
var AllowedSortFields = map[string]string{
	"id":         "id",
	"createdAt":  "created_at",
	"created_at": "created_at",
	"updatedAt":  "updated_at",
	"updated_at": "updated_at",
	"name":       "name",
	"email":      "email",
}

func (p *Pagination) GetPage() int {
	if p.Page <= 0 {
		return DefaultPage
	}
	return p.Page
}

func (p *Pagination) GetPageSize() int {
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetPageSize()
}

// GetSort 解析排序参数
// 支持格式: "createdAt,desc" 或 "desc,createdAt"
func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		return DefaultSort
	}

	parts := strings.Split(p.Sort, ",")
	if len(parts) != 2 {
		return DefaultSort
	}

	p0 := strings.TrimSpace(strings.ToLower(parts[0]))
	p1 := strings.TrimSpace(strings.ToLower(parts[1]))

	var field, order string
	if p0 == "asc" || p0 == "desc" {
		order, field = p0, p1
	} else if p1 == "asc" || p1 == "desc" {
		field, order = p0, p1
	} else {
		return DefaultSort
	}

	if dbField, ok := AllowedSortFields[field]; ok {
		return dbField + " " + order
	}
	return DefaultSort
}

// NewPageResult 创建分页结果
func NewPageResult[T any](list []T, total int64, page, pageSize int) *PageResult[T] {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}
	return &PageResult[T]{
		List:       list,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Paginate GORM 分页 scope
func Paginate(p *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(p.GetSort()).Offset(p.GetOffset()).Limit(p.GetPageSize())
	}
}

// SuccessWithPage 返回分页成功响应
func SuccessWithPage[T any](c *gin.Context, list []T, total int64, p *Pagination) {
	Success(c, NewPageResult(list, total, p.GetPage(), p.GetPageSize()))
}
