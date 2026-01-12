package service

import (
	"context"
	"errors"
	"time"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
)

var (
	ErrPermissionSpaceNotFound   = errors.New("permission space not found")
	ErrPermissionSpaceNameExists = errors.New("permission space name already exists")
	ErrPermissionSpaceFull       = errors.New("permission space has reached maximum capacity (64)")
	ErrPermissionNotFound        = errors.New("permission not found")
	ErrPermissionCodeExists      = errors.New("permission code already exists")
	ErrRoleNotFound              = errors.New("role not found")
	ErrRoleNameExists            = errors.New("role name already exists")
	ErrSystemRoleCannotBeDeleted = errors.New("system role cannot be deleted")
	ErrUserRoleNotFound          = errors.New("user role not found")
	ErrUserRoleAlreadyExists     = errors.New("user already has this role")
)

type BitPermissionManager struct {
	spaceRepo    *repository.PermissionSpaceRepository
	permRepo     *repository.PermissionRepository
	roleRepo     *repository.RoleRepository
	userRoleRepo *repository.UserRoleRepository
	rolePermRepo *repository.RolePermissionRepository
	cacheRepo    *repository.UserPermissionCacheRepository
}

func NewBitPermissionManager(spaceRepo *repository.PermissionSpaceRepository, permRepo *repository.PermissionRepository, roleRepo *repository.RoleRepository, userRoleRepo *repository.UserRoleRepository, rolePermRepo *repository.RolePermissionRepository, cacheRepo *repository.UserPermissionCacheRepository) *BitPermissionManager {
	return &BitPermissionManager{spaceRepo: spaceRepo, permRepo: permRepo, roleRepo: roleRepo, userRoleRepo: userRoleRepo, rolePermRepo: rolePermRepo, cacheRepo: cacheRepo}
}

func (m *BitPermissionManager) CreateSpace(ctx context.Context, name, description string) (*model.PermissionSpace, error) {
	if exists, _ := m.spaceRepo.Exists(ctx, name); exists {
		return nil, ErrPermissionSpaceNameExists
	}
	space := &model.PermissionSpace{Name: name, Description: description, IsActive: true}
	return space, m.spaceRepo.Create(ctx, space)
}

func (m *BitPermissionManager) GetAllSpaces(ctx context.Context) ([]model.SpaceWithCount, error) {
	return m.spaceRepo.FindAllWithCount(ctx)
}

func (m *BitPermissionManager) GetSpaceByID(ctx context.Context, id uint) (*model.PermissionSpace, error) {
	space, err := m.spaceRepo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrPermissionSpaceNotFound) {
		return nil, ErrPermissionSpaceNotFound
	}
	return space, err
}


func (m *BitPermissionManager) CreatePermission(ctx context.Context, code, name, description string, spaceID uint, module string) (*model.Permission, error) {
	if _, err := m.spaceRepo.FindByID(ctx, spaceID); errors.Is(err, repository.ErrPermissionSpaceNotFound) {
		return nil, ErrPermissionSpaceNotFound
	}
	if exists, _ := m.permRepo.Exists(ctx, code); exists {
		return nil, ErrPermissionCodeExists
	}
	maxPos, _ := m.permRepo.GetMaxPositionInSpace(ctx, spaceID)
	nextPos := maxPos + 1
	if nextPos >= 64 {
		return nil, ErrPermissionSpaceFull
	}
	p := &model.Permission{Code: code, Name: name, Description: description, SpaceID: spaceID, Position: uint8(nextPos), Value: uint64(1) << uint(nextPos), Module: module, IsActive: true}
	return p, m.permRepo.Create(ctx, p)
}

func (m *BitPermissionManager) UpdatePermission(ctx context.Context, id uint, name, description string, isActive *bool) (*model.Permission, error) {
	p, err := m.permRepo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	if isActive != nil {
		p.IsActive = *isActive
	}
	return p, m.permRepo.Update(ctx, p)
}

func (m *BitPermissionManager) DeletePermission(ctx context.Context, id uint) error {
	err := m.permRepo.SoftDelete(ctx, id)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return ErrPermissionNotFound
	}
	return err
}


func (m *BitPermissionManager) GetAllPermissions(ctx context.Context) ([]model.PermissionDetail, error) {
	perms, err := m.permRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	details := make([]model.PermissionDetail, len(perms))
	for i, p := range perms {
		sn := ""
		if p.Space != nil {
			sn = p.Space.Name
		}
		details[i] = model.PermissionDetail{ID: p.ID, Code: p.Code, Name: p.Name, Description: p.Description, SpaceID: p.SpaceID, SpaceName: sn, Position: p.Position, Value: p.Value, Module: p.Module, IsActive: p.IsActive}
	}
	return details, nil
}

func (m *BitPermissionManager) GetPermissionByID(ctx context.Context, id uint) (*model.PermissionDetail, error) {
	p, err := m.permRepo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return nil, ErrPermissionNotFound
	}
	if err != nil {
		return nil, err
	}
	sn := ""
	if p.Space != nil {
		sn = p.Space.Name
	}
	return &model.PermissionDetail{ID: p.ID, Code: p.Code, Name: p.Name, Description: p.Description, SpaceID: p.SpaceID, SpaceName: sn, Position: p.Position, Value: p.Value, Module: p.Module, IsActive: p.IsActive}, nil
}


func (m *BitPermissionManager) CreateRoleWithPermissions(ctx context.Context, name, description string, codes []string) (*model.Role, error) {
	if exists, _ := m.roleRepo.Exists(ctx, name); exists {
		return nil, ErrRoleNameExists
	}
	role := &model.Role{Name: name, Description: description, IsActive: true, IsSystem: false}
	if err := m.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	if len(codes) > 0 {
		if err := m.AddPermissionsToRole(ctx, role.ID, codes); err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (m *BitPermissionManager) UpdateRole(ctx context.Context, id uint, name, description string, isActive *bool) (*model.Role, error) {
	role, err := m.roleRepo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrRoleNotFound) {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	if name != "" && name != role.Name {
		if exists, _ := m.roleRepo.Exists(ctx, name); exists {
			return nil, ErrRoleNameExists
		}
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}
	if isActive != nil {
		role.IsActive = *isActive
	}
	return role, m.roleRepo.Update(ctx, role)
}

func (m *BitPermissionManager) DeleteRole(ctx context.Context, id uint) error {
	role, err := m.roleRepo.FindByID(ctx, id)
	if errors.Is(err, repository.ErrRoleNotFound) {
		return ErrRoleNotFound
	}
	if err != nil {
		return err
	}
	if role.IsSystem {
		return ErrSystemRoleCannotBeDeleted
	}
	if uids, _ := m.userRoleRepo.GetUserIDsByRoleID(ctx, id); len(uids) > 0 {
		m.cacheRepo.DeleteByUserIDs(ctx, uids)
	}
	m.rolePermRepo.DeleteByRoleID(ctx, id)
	return m.roleRepo.Delete(ctx, id)
}

func (m *BitPermissionManager) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	return m.roleRepo.FindAll(ctx)
}


func (m *BitPermissionManager) GetRoleByID(ctx context.Context, id uint) (*model.RoleDetail, error) {
	role, err := m.roleRepo.FindByIDWithPermissions(ctx, id)
	if errors.Is(err, repository.ErrRoleNotFound) {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0)
	perms := make([]model.PermissionDetail, 0)
	for _, rp := range role.RolePermissions {
		if rp.Permission != nil {
			codes = append(codes, rp.Permission.Code)
			perms = append(perms, model.PermissionDetail{ID: rp.Permission.ID, Code: rp.Permission.Code, Name: rp.Permission.Name, SpaceID: rp.Permission.SpaceID, Position: rp.Permission.Position, Value: rp.Permission.Value, Module: rp.Permission.Module, IsActive: rp.Permission.IsActive})
		}
	}
	return &model.RoleDetail{ID: role.ID, Name: role.Name, Description: role.Description, IsActive: role.IsActive, IsSystem: role.IsSystem, PermissionCodes: codes, Permissions: perms}, nil
}

func (m *BitPermissionManager) AddPermissionToRole(ctx context.Context, roleID uint, code string) error {
	if _, err := m.roleRepo.FindByID(ctx, roleID); errors.Is(err, repository.ErrRoleNotFound) {
		return ErrRoleNotFound
	}
	p, err := m.permRepo.FindByCode(ctx, code)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return ErrPermissionNotFound
	}
	if err != nil {
		return err
	}
	if exists, _ := m.rolePermRepo.Exists(ctx, roleID, p.ID); exists {
		return nil
	}
	rp := &model.RolePermission{RoleID: roleID, PermissionID: p.ID, SpaceID: p.SpaceID, Value: p.Value}
	if err := m.rolePermRepo.Create(ctx, rp); err != nil {
		return err
	}
	return m.clearCacheForRole(ctx, roleID)
}

func (m *BitPermissionManager) AddPermissionsToRole(ctx context.Context, roleID uint, codes []string) error {
	for _, c := range codes {
		if err := m.AddPermissionToRole(ctx, roleID, c); err != nil {
			return err
		}
	}
	return nil
}


func (m *BitPermissionManager) RemovePermissionFromRole(ctx context.Context, roleID uint, code string) error {
	p, err := m.permRepo.FindByCode(ctx, code)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return ErrPermissionNotFound
	}
	if err != nil {
		return err
	}
	m.rolePermRepo.Delete(ctx, roleID, p.ID)
	return m.clearCacheForRole(ctx, roleID)
}

func (m *BitPermissionManager) RemovePermissionsFromRole(ctx context.Context, roleID uint, codes []string) error {
	for _, c := range codes {
		if err := m.RemovePermissionFromRole(ctx, roleID, c); err != nil {
			return err
		}
	}
	return nil
}

func (m *BitPermissionManager) GetRolePermissions(ctx context.Context, roleID uint) ([]string, error) {
	rps, err := m.rolePermRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0)
	for _, rp := range rps {
		if rp.Permission != nil {
			codes = append(codes, rp.Permission.Code)
		}
	}
	return codes, nil
}

func (m *BitPermissionManager) clearCacheForRole(ctx context.Context, roleID uint) error {
	if uids, _ := m.userRoleRepo.GetUserIDsByRoleID(ctx, roleID); len(uids) > 0 {
		return m.cacheRepo.DeleteByUserIDs(ctx, uids)
	}
	return nil
}


func (m *BitPermissionManager) AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
	if _, err := m.roleRepo.FindByID(ctx, roleID); errors.Is(err, repository.ErrRoleNotFound) {
		return ErrRoleNotFound
	}
	if exists, _ := m.userRoleRepo.Exists(ctx, userID, roleID); exists {
		return ErrUserRoleAlreadyExists
	}
	if err := m.userRoleRepo.Create(ctx, &model.UserRole{UserID: userID, RoleID: roleID}); err != nil {
		return err
	}
	return m.cacheRepo.DeleteByUserID(ctx, userID)
}

func (m *BitPermissionManager) RemoveRoleFromUser(ctx context.Context, userID, roleID uint) error {
	if err := m.userRoleRepo.Delete(ctx, userID, roleID); errors.Is(err, repository.ErrUserRoleNotFound) {
		return ErrUserRoleNotFound
	} else if err != nil {
		return err
	}
	return m.cacheRepo.DeleteByUserID(ctx, userID)
}

func (m *BitPermissionManager) GetUserRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	urs, err := m.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles := make([]model.Role, 0)
	for _, ur := range urs {
		if ur.Role != nil {
			roles = append(roles, *ur.Role)
		}
	}
	return roles, nil
}


func (m *BitPermissionManager) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
	p, err := m.permRepo.FindByCode(ctx, code)
	if errors.Is(err, repository.ErrPermissionNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	cache, err := m.cacheRepo.FindByUserAndSpace(ctx, userID, p.SpaceID)
	if err != nil {
		return false, err
	}
	if cache == nil {
		if err := m.CalculateUserPermissions(ctx, userID); err != nil {
			return false, err
		}
		cache, _ = m.cacheRepo.FindByUserAndSpace(ctx, userID, p.SpaceID)
		if cache == nil {
			return false, nil
		}
	}
	return (cache.Value & p.Value) == p.Value, nil
}

func (m *BitPermissionManager) CalculateUserPermissions(ctx context.Context, userID uint) error {
	m.cacheRepo.DeleteByUserID(ctx, userID)
	urs, err := m.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	sv := make(map[uint]uint64)
	for _, ur := range urs {
		rps, _ := m.rolePermRepo.FindByRoleID(ctx, ur.RoleID)
		for _, rp := range rps {
			sv[rp.SpaceID] |= rp.Value
		}
	}
	now := time.Now()
	for sid, v := range sv {
		m.cacheRepo.Upsert(ctx, &model.UserPermissionCache{UserID: userID, SpaceID: sid, Value: v, CreatedAt: now, UpdatedAt: now})
	}
	return nil
}


func (m *BitPermissionManager) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	urs, err := m.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	cs := make(map[string]struct{})
	for _, ur := range urs {
		rps, _ := m.rolePermRepo.FindByRoleID(ctx, ur.RoleID)
		for _, rp := range rps {
			if rp.Permission != nil {
				cs[rp.Permission.Code] = struct{}{}
			}
		}
	}
	codes := make([]string, 0, len(cs))
	for c := range cs {
		codes = append(codes, c)
	}
	return codes, nil
}

func (m *BitPermissionManager) ClearUserPermissionCache(ctx context.Context, userID uint) error {
	return m.cacheRepo.DeleteByUserID(ctx, userID)
}
