package handler

import (
	"github.com/gin-gonic/gin"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/response"
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
// @Description 创建一个新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "用户数据"
// @Success 201 {object} response.Response{data=model.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
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

	response.Created(c, user.ToResponse())
}

// List godoc
// @Summary 获取用户列表
// @Description 获取分页的用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码（默认：1）"
// @Param page_size query int false "每页数量（默认：10）"
// @Param sort query string false "排序，例如 created_at,desc"
// @Success 200 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	p, ok := BindPagination(c)
	if !ok {
		return
	}

	users, total, err := h.service.List(ctx, p.GetOffset(), p.GetPageSize(), p.GetSort())
	if err != nil {
		c.Error(err)
		return
	}
	response.SuccessWithPage(c, model.ToUserResponseList(users), total, p)
}

// Get godoc
// @Summary 获取用户详情
// @Description 根据 SecUID 获取用户信息
// @Tags 用户管理
// @Produce json
// @Param sec_uid path string true "用户 SecUID"
// @Success 200 {object} response.Response{data=model.UserResponse}
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{sec_uid} [get]
func (h *UserHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}

	user, err := h.service.GetBySecUID(ctx, secUID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, user.ToResponse())
}

// Update godoc
// @Summary 更新用户
// @Description 根据 SecUID 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param sec_uid path string true "用户 SecUID"
// @Param user body model.UpdateUserRequest true "用户数据"
// @Success 200 {object} response.Response{data=model.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{sec_uid} [put]
func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}

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
	response.Success(c, updatedUser.ToResponse())
}

// Delete godoc
// @Summary 删除用户
// @Description 软删除用户
// @Tags 用户管理
// @Param sec_uid path string true "用户 SecUID"
// @Success 204 "删除成功"
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{sec_uid} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}

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

// GetMe godoc
// @Summary 获取当前用户信息
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	user, err := h.service.GetByID(ctx, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, user.ToResponse())
}

// UpdateMe godoc
// @Summary 更新当前用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body model.UpdateUserRequest true "用户数据"
// @Success 200 {object} response.Response{data=model.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("validation error: " + err.Error()))
		return
	}

	updatedUser, err := h.service.Update(ctx, userID, &req)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, updatedUser.ToResponse())
}
