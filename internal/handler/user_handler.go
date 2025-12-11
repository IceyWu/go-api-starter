package handler

import (
	"errors"
	"strconv"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// Create godoc
// @Summary Create a new user
// @Description Create a new user with name, email and age
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "User data: name(required,2-100), email(required), age(0-150)"
// @Success 201 {object} response.Response{data=model.User} "Created successfully"
// @Failure 400 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal error"
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "validation error: "+err.Error())
		return
	}

	user, err := h.service.Create(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, user)
}


// List godoc
// @Summary List users with pagination
// @Description Get a paginated list of users. Supports sorting by: id, name, email, age, createdAt, updatedAt
// @Tags users
// @Produce json
// @Param page query int false "Page number (default: 1)" Example(1)
// @Param page_size query int false "Items per page, max 100 (default: 10)" Example(10)
// @Param sort query string false "Sort: field,order. Example: createdAt,desc or name,asc (default: id,desc)" Example(createdAt,desc)
// @Success 200 {object} response.Response{data=response.UserPageResult}
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var pagination response.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.BadRequest(c, "invalid pagination params")
		return
	}

	users, total, err := h.service.List(pagination.GetOffset(), pagination.GetPageSize(), pagination.GetSort())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, users, total, &pagination)
}

// Get godoc
// @Summary Get a user by ID
// @Description Get a single user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID" Example(1)
// @Success 200 {object} response.Response{data=model.User} "Success"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 404 {object} response.Response "User not found"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, err := h.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

// Update godoc
// @Summary Update a user
// @Description Update an existing user by ID. All fields are optional.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID" Example(1)
// @Param user body model.UpdateUserRequest true "User data: name(2-100), email, age(0-150). All optional."
// @Success 200 {object} response.Response{data=model.User} "Updated successfully"
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "User not found"
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "validation error: "+err.Error())
		return
	}

	user, err := h.service.Update(uint(id), &req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, user)
}

// Delete godoc
// @Summary Delete a user
// @Description Delete a user by ID (soft delete)
// @Tags users
// @Param id path int true "User ID" Example(1)
// @Success 204 "Deleted successfully"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 404 {object} response.Response "User not found"
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.NoContent(c)
}
