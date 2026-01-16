package service

import (
	"context"
	"time"

	"go-api-starter/internal/repository"
)

// PermissionChecker handles permission checking with caching support
type PermissionChecker struct {
	permRepo     repository.PermissionRepositoryInterface
	rolePermRepo repository.RolePermissionRepositoryInterface
	userRoleRepo repository.UserRoleRepositoryInterface
	cache        *PermissionCache
}

// NewPermissionChecker creates a new PermissionChecker
func NewPermissionChecker(
	permRepo repository.PermissionRepositoryInterface,
	rolePermRepo repository.RolePermissionRepositoryInterface,
	userRoleRepo repository.UserRoleRepositoryInterface,
	cache *PermissionCache,
) *PermissionChecker {
	return &PermissionChecker{
		permRepo:     permRepo,
		rolePermRepo: rolePermRepo,
		userRoleRepo: userRoleRepo,
		cache:        cache,
	}
}

// HasPermission checks if a user has a specific permission
func (c *PermissionChecker) HasPermission(ctx context.Context, userID uint, code string) (bool, error) {
	// Get permission by code
	perm, err := c.permRepo.FindByCode(ctx, code)
	if err != nil {
		return false, nil // Permission not found means no access
	}

	// Try to get from cache
	cachedValue, err := c.cache.Get(ctx, userID, perm.SpaceID)
	if err != nil {
		return false, err
	}

	// Cache miss - calculate and cache
	if cachedValue == nil {
		permissions, err := c.CalculateUserPermissions(ctx, userID)
		if err != nil {
			return false, err
		}

		// Cache the calculated permissions
		if err := c.cache.Set(ctx, userID, permissions); err != nil {
			return false, err
		}

		// Get the value for this space
		value, ok := permissions[perm.SpaceID]
		if !ok {
			return false, nil
		}
		return (value & perm.Value) == perm.Value, nil
	}

	// Cache hit - check permission
	return (*cachedValue & perm.Value) == perm.Value, nil
}

// GetUserPermissions returns all permission codes for a user
func (c *PermissionChecker) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	// Get user roles
	userRoles, err := c.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect unique permission codes
	codeSet := make(map[string]struct{})
	for _, ur := range userRoles {
		rolePerms, err := c.rolePermRepo.FindByRoleID(ctx, ur.RoleID)
		if err != nil {
			continue
		}
		for _, rp := range rolePerms {
			if rp.Permission != nil {
				codeSet[rp.Permission.Code] = struct{}{}
			}
		}
	}

	// Convert to slice
	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	return codes, nil
}

// CalculateUserPermissions calculates all permission values for a user by space
func (c *PermissionChecker) CalculateUserPermissions(ctx context.Context, userID uint) (map[uint]uint64, error) {
	// Get user roles
	userRoles, err := c.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Aggregate permissions by space using bitwise OR
	spaceValues := make(map[uint]uint64)
	for _, ur := range userRoles {
		rolePerms, err := c.rolePermRepo.FindByRoleID(ctx, ur.RoleID)
		if err != nil {
			continue
		}
		for _, rp := range rolePerms {
			spaceValues[rp.SpaceID] |= rp.Value
		}
	}

	return spaceValues, nil
}

// RefreshUserCache recalculates and caches user permissions
func (c *PermissionChecker) RefreshUserCache(ctx context.Context, userID uint) error {
	// Invalidate existing cache
	if err := c.cache.InvalidateUser(ctx, userID); err != nil {
		return err
	}

	// Calculate new permissions
	permissions, err := c.CalculateUserPermissions(ctx, userID)
	if err != nil {
		return err
	}

	// Cache the new permissions
	return c.cache.Set(ctx, userID, permissions)
}

// GetCacheStats returns cache statistics
func (c *PermissionChecker) GetCacheStats() CacheStats {
	return c.cache.GetStats()
}

// GetCachedPermissions returns cached permissions for a user (for debugging/testing)
func (c *PermissionChecker) GetCachedPermissions(ctx context.Context, userID uint) (map[uint]uint64, error) {
	return c.cache.GetAllForUser(ctx, userID)
}

// CheckPermissionWithCache checks permission and returns whether cache was used
func (c *PermissionChecker) CheckPermissionWithCache(ctx context.Context, userID uint, code string) (hasPermission bool, cacheHit bool, err error) {
	// Get permission by code
	perm, err := c.permRepo.FindByCode(ctx, code)
	if err != nil {
		return false, false, nil
	}

	// Try to get from cache
	cachedValue, err := c.cache.Get(ctx, userID, perm.SpaceID)
	if err != nil {
		return false, false, err
	}

	// Cache miss
	if cachedValue == nil {
		permissions, err := c.CalculateUserPermissions(ctx, userID)
		if err != nil {
			return false, false, err
		}

		if err := c.cache.Set(ctx, userID, permissions); err != nil {
			return false, false, err
		}

		value, ok := permissions[perm.SpaceID]
		if !ok {
			return false, false, nil
		}
		return (value & perm.Value) == perm.Value, false, nil
	}

	// Cache hit
	return (*cachedValue & perm.Value) == perm.Value, true, nil
}

// InvalidateUserCache invalidates cache for a specific user
func (c *PermissionChecker) InvalidateUserCache(ctx context.Context, userID uint) error {
	return c.cache.InvalidateUser(ctx, userID)
}

// InvalidateRoleCache invalidates cache for all users with a specific role
func (c *PermissionChecker) InvalidateRoleCache(ctx context.Context, roleID uint) error {
	userIDs, err := c.userRoleRepo.GetUserIDsByRoleID(ctx, roleID)
	if err != nil {
		return err
	}
	return c.cache.InvalidateByRole(ctx, roleID, userIDs)
}

// CacheEntry represents a cached permission entry with metadata
type CacheEntry struct {
	Value     map[uint]uint64
	ExpiresAt time.Time
}
