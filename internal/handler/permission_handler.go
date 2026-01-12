package handler

import (
	"strconv"

	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// PermissionHandler handles permission HTTP requests
type PermissionHandler struct {
	service *service.PermissionService
}

// NewPermissionHandler creates a new PermissionHandler
func NewPermissionHandler(svc *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{service: svc}
}

// CreateSpace godoc
// @Summary Create a permission space
// @Description Create a new permission space
// @Tags permissions
// @Accept json
// @Produce json
// @Param space body model.CreateSpaceRequest true "Space data"
// @Success 201 {object} response.Response{data=model.PermissionSpace}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/permissions/spaces [post]
func (h *PermissionHandler) CreateSpace(c *gin.Context) {
	var req model.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	space, err := h.service.CreateSpace(c.Request.Context(), &req)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Created(c, space)
}

// GetAllSpaces godoc
// @Summary Get all permission spaces
// @Description Get all permission spaces with permission count
// @Tags permissions
// @Produce json
// @Success 200 {object} response.Response{data=[]model.SpaceWithCount}
// @Router /api/v1/permissions/spaces [get]
func (h *PermissionHandler) GetAllSpaces(c *gin.Context) {
	spaces, err := h.service.GetAllSpaces(c.Request.Context())
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, spaces)
}


// CreatePermission godoc
// @Summary Create a permission
// @Description Create a new permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body model.CreatePermissionRequest true "Permission data"
// @Success 201 {object} response.Response{data=model.Permission}
// @Failure 400 {object} response.Response
// @Router /api/v1/permissions/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req model.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	perm, err := h.service.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Created(c, perm)
}

// GetAllPermissions godoc
// @Summary Get all permissions
// @Description Get all permissions with details
// @Tags permissions
// @Produce json
// @Success 200 {object} response.Response{data=[]model.PermissionDetail}
// @Router /api/v1/permissions/permissions [get]
func (h *PermissionHandler) GetAllPermissions(c *gin.Context) {
	perms, err := h.service.GetAllPermissions(c.Request.Context())
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, perms)
}

// GetPermission godoc
// @Summary Get a permission
// @Description Get a permission by ID
// @Tags permissions
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} response.Response{data=model.PermissionDetail}
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}
	perm, err := h.service.GetPermissionByID(c.Request.Context(), uint(id))
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, perm)
}

// UpdatePermission godoc
// @Summary Update a permission
// @Description Update a permission by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Param permission body model.UpdatePermissionRequest true "Permission data"
// @Success 200 {object} response.Response{data=model.Permission}
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}
	var req model.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	perm, err := h.service.UpdatePermission(c.Request.Context(), uint(id), &req)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, perm)
}

// DeletePermission godoc
// @Summary Delete a permission
// @Description Delete a permission by ID
// @Tags permissions
// @Param id path int true "Permission ID"
// @Success 204
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}
	if err := h.service.DeletePermission(c.Request.Context(), uint(id)); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.NoContent(c)
}


// CreateRole godoc
// @Summary Create a role
// @Description Create a new role with optional permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Param role body model.CreateRoleRequest true "Role data"
// @Success 201 {object} response.Response{data=model.Role}
// @Failure 400 {object} response.Response
// @Router /api/v1/permissions/roles [post]
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	role, err := h.service.CreateRole(c.Request.Context(), &req)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Created(c, role)
}

// GetAllRoles godoc
// @Summary Get all roles
// @Description Get all roles
// @Tags permissions
// @Produce json
// @Success 200 {object} response.Response{data=[]model.Role}
// @Router /api/v1/permissions/roles [get]
func (h *PermissionHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.service.GetAllRoles(c.Request.Context())
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, roles)
}

// GetRole godoc
// @Summary Get a role
// @Description Get a role by ID with permissions
// @Tags permissions
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.Response{data=model.RoleDetail}
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/roles/{id} [get]
func (h *PermissionHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	role, err := h.service.GetRoleByID(c.Request.Context(), uint(id))
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, role)
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update a role by ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param role body model.UpdateRoleRequest true "Role data"
// @Success 200 {object} response.Response{data=model.Role}
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/roles/{id} [put]
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	role, err := h.service.UpdateRole(c.Request.Context(), uint(id), &req)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, role)
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role by ID
// @Tags permissions
// @Param id path int true "Role ID"
// @Success 204
// @Failure 404 {object} response.Response
// @Router /api/v1/permissions/roles/{id} [delete]
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	if err := h.service.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.NoContent(c)
}


// GetRolePermissions godoc
// @Summary Get role permissions
// @Description Get all permission codes for a role
// @Tags permissions
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/v1/permissions/roles/{id}/permissions [get]
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	codes, err := h.service.GetRolePermissions(c.Request.Context(), uint(id))
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, codes)
}

// AddRolePermissions godoc
// @Summary Add permissions to role
// @Description Add permissions to a role
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permissions body model.RolePermissionsRequest true "Permission codes"
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/roles/{id}/permissions [post]
func (h *PermissionHandler) AddRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	var req model.RolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	if err := h.service.AddRolePermissions(c.Request.Context(), uint(id), req.PermissionCodes); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, nil)
}

// RemoveRolePermissions godoc
// @Summary Remove permissions from role
// @Description Remove permissions from a role
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permissions body model.RolePermissionsRequest true "Permission codes"
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/roles/{id}/permissions [delete]
func (h *PermissionHandler) RemoveRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	var req model.RolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	if err := h.service.RemoveRolePermissions(c.Request.Context(), uint(id), req.PermissionCodes); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, nil)
}


// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles for a user
// @Tags permissions
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=[]model.Role}
// @Router /api/v1/permissions/users/{id}/roles [get]
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	roles, err := h.service.GetUserRoles(c.Request.Context(), uint(id))
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, roles)
}

// AssignUserRole godoc
// @Summary Assign role to user
// @Description Assign a role to a user
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param role body model.AssignRoleRequest true "Role ID"
// @Success 200 {object} response.Response
// @Router /api/v1/permissions/users/{id}/roles [post]
func (h *PermissionHandler) AssignUserRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}
	if err := h.service.AssignUserRole(c.Request.Context(), uint(id), req.RoleID); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, nil)
}

// RemoveUserRole godoc
// @Summary Remove role from user
// @Description Remove a role from a user
// @Tags permissions
// @Param id path int true "User ID"
// @Param roleId path int true "Role ID"
// @Success 204
// @Router /api/v1/permissions/users/{id}/roles/{roleId} [delete]
func (h *PermissionHandler) RemoveUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	if err := h.service.RemoveUserRole(c.Request.Context(), uint(userID), uint(roleID)); err != nil {
		handlePermissionError(c, err)
		return
	}
	response.NoContent(c)
}

// GetUserPermissions godoc
// @Summary Get user permissions
// @Description Get all permission codes for a user
// @Tags permissions
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/v1/permissions/users/{id}/permissions [get]
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	codes, err := h.service.GetUserPermissions(c.Request.Context(), uint(id))
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, codes)
}

// GetMyPermissions godoc
// @Summary Get current user permissions
// @Description Get all permission codes for the current user
// @Tags permissions
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/v1/permissions/me/permissions [get]
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		response.Unauthorized(c, "用户ID无效")
		return
	}
	codes, err := h.service.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		handlePermissionError(c, err)
		return
	}
	response.Success(c, codes)
}


func handlePermissionError(c *gin.Context, err error) {
	switch err {
	case service.ErrPermissionSpaceNotFound, service.ErrPermissionNotFound, service.ErrRoleNotFound, service.ErrUserRoleNotFound:
		response.NotFound(c, err.Error())
	case service.ErrPermissionSpaceNameExists, service.ErrPermissionCodeExists, service.ErrRoleNameExists, service.ErrUserRoleAlreadyExists:
		response.Conflict(c, err.Error())
	case service.ErrSystemRoleCannotBeDeleted, service.ErrPermissionSpaceFull:
		response.UnprocessableEntity(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
}
