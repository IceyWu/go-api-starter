package handler

import (
	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	service service.UserServiceInterface
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(svc service.UserServiceInterface) *UserHandler {
	return &UserHandler{service: svc}
}

// Create godoc
// @Summary 创建用户
// @Description 创建一个新用户，需要提供姓名和邮箱
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "用户数据：name(必填,2-100字符), email(必填)"
// @Success 201 {object} response.Response{data=model.User} "创建成功"
// @Failure 400 {object} response.Response "参数验证错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("validation error: " + err.Error()))
		return
	}

	user, err := h.service.Create(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Created(c, user)
}

// List godoc
// @Summary 获取用户列表
// @Description 获取分页的用户列表，支持按以下字段排序：id, name, email, createdAt, updatedAt
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码（默认：1）" Example(1)
// @Param page_size query int false "每页数量，最大100（默认：10）" Example(10)
// @Param sort query string false "排序：字段,顺序。示例：createdAt,desc 或 name,asc（默认：id,desc）" Example(createdAt,desc)
// @Success 200 {object} response.Response{data=response.UserPageResult}
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	
	var pagination response.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.Error(apperrors.BadRequest("invalid pagination params"))
		return
	}

	users, total, err := h.service.List(ctx, pagination.GetOffset(), pagination.GetPageSize(), pagination.GetSort())
	if err != nil {
		c.Error(err)
		return
	}
	response.SuccessWithPage(c, users, total, &pagination)
}

// Get godoc
// @Summary 获取用户详情
// @Description 根据用户SecUID获取单个用户的详细信息
// @Tags 用户管理
// @Produce json
// @Param sec_uid path string true "用户SecUID"
// @Success 200 {object} response.Response{data=model.User} "获取成功"
// @Failure 400 {object} response.Response "无效的用户SecUID"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{sec_uid} [get]
func (h *UserHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	
	secUID := c.Param("sec_uid")
	if secUID == "" {
		c.Error(apperrors.BadRequest("invalid user SecUID"))
		return
	}

	user, err := h.service.GetBySecUID(ctx, secUID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, user)
}

// Update godoc
// @Summary 更新用户
// @Description 根据SecUID更新现有用户，所有字段都是可选的
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param sec_uid path string true "用户SecUID"
// @Param user body model.UpdateUserRequest true "用户数据：name(2-100字符), email，所有字段可选"
// @Success 200 {object} response.Response{data=model.User} "更新成功"
// @Failure 400 {object} response.Response "参数验证错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{sec_uid} [put]
func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	
	secUID := c.Param("sec_uid")
	if secUID == "" {
		c.Error(apperrors.BadRequest("invalid user SecUID"))
		return
	}

	// 先通过 SecUID 获取用户
	user, err := h.service.GetBySecUID(ctx, secUID)
	if err != nil {
		c.Error(err)
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("validation error: " + err.Error()))
		return
	}

	updatedUser, err := h.service.Update(ctx, user.ID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, updatedUser)
}

// Delete godoc
// @Summary 删除用户
// @Description 根据SecUID删除用户（软删除）
// @Tags 用户管理
// @Param sec_uid path string true "用户SecUID"
// @Success 204 "删除成功"
// @Failure 400 {object} response.Response "无效的用户SecUID"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/v1/users/{sec_uid} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	
	secUID := c.Param("sec_uid")
	if secUID == "" {
		c.Error(apperrors.BadRequest("invalid user SecUID"))
		return
	}

	// 先通过 SecUID 获取用户
	user, err := h.service.GetBySecUID(ctx, secUID)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.service.Delete(ctx, user.ID); err != nil {
		c.Error(err)
		return
	}
	response.NoContent(c)
}
