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

// ====================
// 权限空间 (Permission Space)
// ====================

// CreateSpace godoc
// @Summary 创建权限空间
// @Description 创建一个新的权限空间
// @Tags 权限空间
// @Accept json
// @Produce json
// @Param space body model.CreateSpaceRequest true "权限空间数据"
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
// @Summary 获取所有权限空间
// @Description 获取所有权限空间及其权限数量
// @Tags 权限空间
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

// ====================
// 权限管理 (Permission Management)
// ====================

// CreatePermission godoc
// @Summary 创建权限
// @Description 创建一个新的权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param permission body model.CreatePermissionRequest true "权限数据"
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
// @Summary 获取所有权限
// @Description 获取所有权限详情
// @Tags 权限管理
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
// @Summary 获取权限详情
// @Description 根据ID获取权限详情
// @Tags 权限管理
// @Produce json
// @Param id path int true "权限ID"
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
// @Summary 更新权限
// @Description 根据ID更新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Param permission body model.UpdatePermissionRequest true "权限数据"
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
// @Summary 删除权限
// @Description 根据ID删除权限
// @Tags 权限管理
// @Param id path int true "权限ID"
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

// ====================
// 角色管理 (Role Management)
// ====================

// CreateRole godoc
// @Summary 创建角色
// @Description 创建一个新角色，可选择性地分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param role body model.CreateRoleRequest true "角色数据"
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
// @Summary 获取所有角色
// @Description 获取所有角色列表
// @Tags 角色管理
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
// @Summary 获取角色详情
// @Description 根据ID获取角色详情及其权限
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
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
// @Summary 更新角色
// @Description 根据ID更新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param role body model.UpdateRoleRequest true "角色数据"
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
// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags 角色管理
// @Param id path int true "角色ID"
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
// @Summary 获取角色权限
// @Description 获取角色的所有权限代码
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
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
// @Summary 为角色添加权限
// @Description 为角色添加一个或多个权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param permissions body model.RolePermissionsRequest true "权限代码列表"
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
// @Summary 移除角色权限
// @Description 从角色中移除一个或多个权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param permissions body model.RolePermissionsRequest true "权限代码列表"
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

// ====================
// 用户权限 (User Permissions)
// ====================

// GetUserRoles godoc
// @Summary 获取用户角色
// @Description 获取用户的所有角色
// @Tags 用户权限
// @Produce json
// @Param id path int true "用户ID"
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
// @Summary 为用户分配角色
// @Description 为用户分配一个角色
// @Tags 用户权限
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param role body model.AssignRoleRequest true "角色ID"
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
// @Summary 移除用户角色
// @Description 从用户中移除一个角色
// @Tags 用户权限
// @Param id path int true "用户ID"
// @Param roleId path int true "角色ID"
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
// @Summary 获取用户权限
// @Description 获取用户的所有权限代码
// @Tags 用户权限
// @Produce json
// @Param id path int true "用户ID"
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
// @Summary 获取当前用户权限
// @Description 获取当前登录用户的所有权限代码
// @Tags 用户权限
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
