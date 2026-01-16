package handler

import (
	"strconv"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct {
	service service.OperationLogServiceInterface
}

// NewOperationLogHandler 创建操作日志处理器
func NewOperationLogHandler(svc service.OperationLogServiceInterface) *OperationLogHandler {
	return &OperationLogHandler{service: svc}
}

// OperationLogQuery 操作日志查询参数
type OperationLogQuery struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	UserID    *uint  `form:"user_id" json:"user_id"`
	Module    string `form:"module" json:"module"`
	Action    string `form:"action" json:"action"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
}

// OperationLogPageResult Swagger 文档用
type OperationLogPageResult struct {
	List       []model.OperationLog `json:"list"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Total      int64                `json:"total"`
	TotalPages int                  `json:"total_pages"`
}

// List godoc
// @Summary 获取操作日志列表
// @Description 分页查询操作日志，支持按用户、模块、操作类型、时间范围过滤
// @Tags 操作日志
// @Produce json
// @Param page query int false "页码（默认：1）"
// @Param page_size query int false "每页数量（默认：10，最大：100）"
// @Param user_id query int false "用户ID"
// @Param module query string false "模块名称"
// @Param action query string false "操作类型"
// @Param start_time query string false "开始时间（格式：2006-01-02 15:04:05）"
// @Param end_time query string false "结束时间（格式：2006-01-02 15:04:05）"
// @Success 200 {object} response.Response{data=OperationLogPageResult}
// @Failure 400 {object} response.Response
// @Router /api/v1/operation-logs [get]
func (h *OperationLogHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	var query OperationLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(apperrors.BadRequest("参数错误: " + err.Error()))
		return
	}

	// 默认分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if query.UserID != nil {
		filters["user_id"] = *query.UserID
	}
	if query.Module != "" {
		filters["module"] = query.Module
	}
	if query.Action != "" {
		filters["action"] = query.Action
	}
	if query.StartTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", query.StartTime)
		if err == nil {
			filters["start_time"] = t
		}
	}
	if query.EndTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", query.EndTime)
		if err == nil {
			filters["end_time"] = t
		}
	}

	offset := (query.Page - 1) * query.PageSize
	logs, total, err := h.service.List(ctx, offset, query.PageSize, filters)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, response.NewPageResult(logs, total, query.Page, query.PageSize))
}

// Get godoc
// @Summary 获取操作日志详情
// @Description 根据ID获取操作日志详情
// @Tags 操作日志
// @Produce json
// @Param id path int true "日志ID"
// @Success 200 {object} response.Response{data=model.OperationLog}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/operation-logs/{id} [get]
func (h *OperationLogHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.BadRequest("无效的日志ID"))
		return
	}

	log, err := h.service.GetByID(ctx, uint(id))
	if err != nil {
		c.Error(apperrors.NotFound("日志不存在"))
		return
	}

	response.Success(c, log)
}
