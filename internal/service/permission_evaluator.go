package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/pguia/iam/internal/repository"
)

// PermissionEvaluator evaluates permission checks
type PermissionEvaluator interface {
	CheckPermission(principal string, resourceID uuid.UUID, permission string, context map[string]string) (bool, string, error)
	GetEffectivePermissions(principal string, resourceID uuid.UUID) ([]string, []string, error)
}

type permissionEvaluator struct {
	resourceRepo   repository.ResourceRepository
	policyRepo     repository.PolicyRepository
	permissionRepo repository.PermissionRepository
	cache          CacheService
}

// NewPermissionEvaluator creates a new permission evaluator
func NewPermissionEvaluator(
	resourceRepo repository.ResourceRepository,
	policyRepo repository.PolicyRepository,
	permissionRepo repository.PermissionRepository,
	cache CacheService,
) PermissionEvaluator {
	return &permissionEvaluator{
		resourceRepo:   resourceRepo,
		policyRepo:     policyRepo,
		permissionRepo: permissionRepo,
		cache:          cache,
	}
}

// CheckPermission checks if a principal has a specific permission on a resource
func (pe *permissionEvaluator) CheckPermission(
	principal string,
	resourceID uuid.UUID,
	permission string,
	context map[string]string,
) (bool, string, error) {
	// Check cache first
	cacheKey := GenerateCacheKey(principal, resourceID.String(), permission)
	if cached, found := pe.cache.Get(cacheKey); found {
		result := cached.(bool)
		if result {
			return true, "Permission granted (cached)", nil
		}
	}

	// Get the resource
	resource, err := pe.resourceRepo.GetByID(resourceID)
	if err != nil {
		return false, "Error fetching resource", err
	}
	if resource == nil {
		return false, "Resource not found", nil
	}

	// Check permission on this resource and all ancestors (hierarchical inheritance)
	resources := []uuid.UUID{resourceID}

	// Get ancestors
	ancestors, err := pe.resourceRepo.GetAncestors(resourceID)
	if err != nil {
		return false, "Error fetching resource ancestors", err
	}
	for _, ancestor := range ancestors {
		resources = append(resources, ancestor.ID)
	}

	// Check each resource in the hierarchy
	for _, resID := range resources {
		allowed, reason, err := pe.checkResourcePermission(principal, resID, permission, context)
		if err != nil {
			return false, reason, err
		}
		if allowed {
			// Cache the positive result
			pe.cache.Set(cacheKey, true)
			return true, reason, nil
		}
	}

	return false, "Permission denied: no matching policy found", nil
}

// checkResourcePermission checks permission on a specific resource (no hierarchy)
func (pe *permissionEvaluator) checkResourcePermission(
	principal string,
	resourceID uuid.UUID,
	permission string,
	context map[string]string,
) (bool, string, error) {
	// Get policy for this resource
	policy, err := pe.policyRepo.GetByResourceID(resourceID)
	if err != nil {
		return false, "Error fetching policy", err
	}
	if policy == nil {
		return false, "No policy found for resource", nil
	}

	// Check each binding in the policy
	for _, binding := range policy.Bindings {
		// Check if principal is in members
		if !binding.HasMember(principal) {
			continue
		}

		// Check if binding has a condition
		if binding.Condition != nil {
			// Evaluate condition (simplified - in production use CEL)
			allowed := pe.evaluateCondition(binding.Condition, context)
			if !allowed {
				continue
			}
		}

		// Check if role has the required permission
		if binding.Role != nil {
			if binding.Role.HasPermission(permission) {
				return true, fmt.Sprintf("Permission granted via role '%s' on resource '%s'",
					binding.Role.Name, resourceID), nil
			}
		}
	}

	return false, "No matching binding found", nil
}

// evaluateCondition evaluates a condition expression (simplified)
// In production, use CEL (Common Expression Language) for this
func (pe *permissionEvaluator) evaluateCondition(condition *domain.Condition, context map[string]string) bool {
	if condition == nil || condition.Expression == "" {
		return true
	}

	// Simplified condition evaluation
	// In production, integrate with CEL library
	// For now, just return true to allow testing
	// TODO: Implement CEL integration
	return true
}

// GetEffectivePermissions returns all effective permissions for a principal on a resource
func (pe *permissionEvaluator) GetEffectivePermissions(
	principal string,
	resourceID uuid.UUID,
) ([]string, []string, error) {
	permissions := make(map[string]bool)
	roles := make(map[string]bool)

	// Get the resource
	resource, err := pe.resourceRepo.GetByID(resourceID)
	if err != nil {
		return nil, nil, err
	}
	if resource == nil {
		return nil, nil, fmt.Errorf("resource not found")
	}

	// Collect from this resource and all ancestors
	resources := []uuid.UUID{resourceID}
	ancestors, err := pe.resourceRepo.GetAncestors(resourceID)
	if err != nil {
		return nil, nil, err
	}
	for _, ancestor := range ancestors {
		resources = append(resources, ancestor.ID)
	}

	// Check each resource
	for _, resID := range resources {
		policy, err := pe.policyRepo.GetByResourceID(resID)
		if err != nil {
			continue
		}
		if policy == nil {
			continue
		}

		// Check each binding
		for _, binding := range policy.Bindings {
			if !binding.HasMember(principal) {
				continue
			}

			if binding.Role != nil {
				roles[binding.Role.Name] = true

				// Add all permissions from this role
				for _, perm := range binding.Role.Permissions {
					permissions[perm.Name] = true
				}
			}
		}
	}

	// Convert maps to slices
	permList := make([]string, 0, len(permissions))
	for perm := range permissions {
		permList = append(permList, perm)
	}

	roleList := make([]string, 0, len(roles))
	for role := range roles {
		roleList = append(roleList, role)
	}

	return permList, roleList, nil
}
