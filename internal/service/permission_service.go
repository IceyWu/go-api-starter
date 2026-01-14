package service

import (
	"context"

	"go-api-starter/internal/model"
)

// PermissionService wraps BitPermissionManager with additional business logic
type PermissionService struct {
	manager *BitPermissionManager
}

// NewPermissionService creates a new PermissionService
func NewPermissionService(manager *BitPermissionManager) *PermissionService {
	return &PermissionService{manager: manager}
}

// CreateSpace creates a new permission space
func (s *PermissionService) CreateSpace(ctx context.Context, req *model.CreateSpaceRequest) (*model.PermissionSpace, error) {
	return s.manager.CreateSpace(ctx, req.Name, req.Description)
}

// GetAllSpaces returns all permission spaces
func (s *PermissionService) GetAllSpaces(ctx context.Context) ([]model.SpaceWithCount, error) {
	return s.manager.GetAllSpaces(ctx)
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, req *model.CreatePermissionRequest) (*model.Permission, error) {
	return s.manager.CreatePermission(ctx, req.Code, req.Name, req.Description, req.SpaceID, req.Module)
}

// GetAllPermissions returns all permissions
func (s *PermissionService) GetAllPermissions(ctx context.Context) ([]model.PermissionDetail, error) {
	return s.manager.GetAllPermissions(ctx)
}

// GetPermissionByID returns a permission by ID
func (s *PermissionService) GetPermissionByID(ctx context.Context, id uint) (*model.PermissionDetail, error) {
	return s.manager.GetPermissionByID(ctx, id)
}

// UpdatePermission updates a permission
func (s *PermissionService) UpdatePermission(ctx context.Context, id uint, req *model.UpdatePermissionRequest) (*model.Permission, error) {
	return s.manager.UpdatePermission(ctx, id, req.Name, req.Description, req.IsActive)
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(ctx context.Context, id uint) error {
	return s.manager.DeletePermission(ctx, id)
}


// CreateRole creates a new role
func (s *PermissionService) CreateRole(ctx context.Context, req *model.CreateRoleRequest) (*model.Role, error) {
	return s.manager.CreateRoleWithPermissions(ctx, req.Name, req.Description, req.PermissionCodes)
}

// GetAllRoles returns all roles
func (s *PermissionService) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	return s.manager.GetAllRoles(ctx)
}

// GetRoleByID returns a role by ID
func (s *PermissionService) GetRoleByID(ctx context.Context, id uint) (*model.RoleDetail, error) {
	return s.manager.GetRoleByID(ctx, id)
}

// UpdateRole updates a role
func (s *PermissionService) UpdateRole(ctx context.Context, id uint, req *model.UpdateRoleRequest) (*model.Role, error) {
	return s.manager.UpdateRole(ctx, id, req.Name, req.Description, req.IsActive)
}

// DeleteRole deletes a role
func (s *PermissionService) DeleteRole(ctx context.Context, id uint) error {
	return s.manager.DeleteRole(ctx, id)
}

// GetRolePermissions returns all permission codes for a role
func (s *PermissionService) GetRolePermissions(ctx context.Context, roleID uint) ([]string, error) {
	return s.manager.GetRolePermissions(ctx, roleID)
}

// AddRolePermissions adds permissions to a role
func (s *PermissionService) AddRolePermissions(ctx context.Context, roleID uint, codes []string) error {
	return s.manager.AddPermissionsToRole(ctx, roleID, codes)
}

// RemoveRolePermissions removes permissions from a role
func (s *PermissionService) RemoveRolePermissions(ctx context.Context, roleID uint, codes []string) error {
	return s.manager.RemovePermissionsFromRole(ctx, roleID, codes)
}

// GetUserRoles returns all roles for a user
func (s *PermissionService) GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	return s.manager.GetUserRoles(ctx, userID)
}

// AssignUserRole assigns a role to a user
func (s *PermissionService) AssignUserRole(ctx context.Context, userID, roleID uint) error {
	return s.manager.AssignRoleToUser(ctx, userID, roleID)
}

// RemoveUserRole removes a role from a user
func (s *PermissionService) RemoveUserRole(ctx context.Context, userID, roleID uint) error {
	return s.manager.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserPermissions returns all permission codes for a user
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	return s.manager.GetUserPermissions(ctx, userID)
}

// HasPermission checks if a user has a permission
func (s *PermissionService) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
	return s.manager.HasPermission(ctx, userID, code)
}

// CheckUserPermission checks if a user has a specific permission
func (s *PermissionService) CheckUserPermission(userID uint, permissionCode string) (bool, error) {
	ctx := context.Background()
	return s.manager.HasPermission(ctx, userID, permissionCode)
}
