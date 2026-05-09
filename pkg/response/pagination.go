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
	Preset   string `form:"preset" json:"preset"`       // 返回字段预设: mini|simple|full
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
	List       []any `json:"list"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
	AllPageSize     = -1 // page_size=-1 表示获取全部数据
	DefaultSort     = "id desc"

	// Preset 预设模式
	PresetMini   = "mini"   // 最精简，适用于地图打点、缩略图等
	PresetSimple = "simple" // 默认，适用于列表页
	PresetFull   = "full"   // 完整数据，适用于详情/管理后台
)

// AllowedSortFields 默认允许排序的字段映射，可通过 RegisterSortFields 扩展
var AllowedSortFields = map[string]string{
	"id":         "id",
	"createdAt":  "created_at",
	"created_at": "created_at",
	"updatedAt":  "updated_at",
	"updated_at": "updated_at",
	"name":       "name",
	"email":      "email",
}

// RegisterSortFields 注册额外的排序字段
func RegisterSortFields(fields map[string]string) {
	for k, v := range fields {
		AllowedSortFields[k] = v
	}
}

// IsAll 判断是否请求全部数据（page_size=-1）
func (p *Pagination) IsAll() bool {
	return p.PageSize == AllPageSize
}

func (p *Pagination) GetPage() int {
	if p.IsAll() {
		return 1
	}
	if p.Page <= 0 {
		return DefaultPage
	}
	return p.Page
}

func (p *Pagination) GetPageSize() int {
	if p.IsAll() {
		return AllPageSize
	}
	if p.PageSize <= 0 {
		return DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return p.PageSize
}

func (p *Pagination) GetOffset() int {
	if p.IsAll() {
		return 0
	}
	return (p.GetPage() - 1) * p.GetPageSize()
}

// GetPreset 获取返回字段预设模式，默认 simple
func (p *Pagination) GetPreset() string {
	switch p.Preset {
	case PresetMini, PresetSimple, PresetFull:
		return p.Preset
	default:
		return PresetSimple
	}
}

// IsMini 是否为精简模式
func (p *Pagination) IsMini() bool {
	return p.GetPreset() == PresetMini
}

// IsFull 是否为完整模式
func (p *Pagination) IsFull() bool {
	return p.GetPreset() == PresetFull
}

// GetSort 解析排序参数
// 支持格式: "created_at,desc" 或 "desc,createdAt"
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
	// 全量模式：pageSize 用实际总数表示
	if pageSize == AllPageSize {
		return &PageResult[T]{
			List:       list,
			Page:       1,
			PageSize:   int(total),
			Total:      total,
			TotalPages: 1,
		}
	}
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
		if p.IsAll() {
			return db.Order(p.GetSort())
		}
		return db.Order(p.GetSort()).Offset(p.GetOffset()).Limit(p.GetPageSize())
	}
}

// SuccessWithPage 返回分页成功响应
func SuccessWithPage[T any](c *gin.Context, list []T, total int64, p *Pagination) {
	Success(c, NewPageResult(list, total, p.GetPage(), p.GetPageSize()))
}
