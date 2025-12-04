package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test: Update Resource
func TestIAMService_UpdateResource(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "updated-bucket",
		Attributes: map[string]string{
			"region": "us-west-2",
		},
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(resource, nil)
	resourceRepo.On("Update", mock.AnythingOfType("*domain.Resource")).Return(nil)

	// Update resource
	updatedResource, err := service.UpdateResource(resourceID, "updated-bucket", map[string]string{"region": "us-west-2"})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, updatedResource)
	resourceRepo.AssertExpectations(t)
}

// Test: List Resources
func TestIAMService_ListResources(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	parentID := uuid.New()
	expectedResources := []domain.Resource{
		{ID: uuid.New(), Type: "project", Name: "proj1"},
		{ID: uuid.New(), Type: "project", Name: "proj2"},
	}

	// Mock expectations
	resourceRepo.On("List", &parentID, "project", 10, 0).Return(expectedResources, nil)

	// List resources
	resources, err := service.ListResources(&parentID, "project", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	resourceRepo.AssertExpectations(t)
}

// Test: Get Resource Hierarchy
func TestIAMService_GetResourceHierarchy(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	ancestors := []domain.Resource{
		{ID: uuid.New(), Type: "folder", Name: "eng"},
		{ID: uuid.New(), Type: "org", Name: "acme"},
	}
	descendants := []domain.Resource{
		{ID: uuid.New(), Type: "bucket", Name: "data"},
	}

	// Mock expectations
	resourceRepo.On("GetAncestors", resourceID).Return(ancestors, nil)
	resourceRepo.On("GetDescendants", resourceID).Return(descendants, nil)

	// Get hierarchy
	returnedAncestors, returnedDescendants, err := service.GetResourceHierarchy(resourceID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ancestors, returnedAncestors)
	assert.Equal(t, descendants, returnedDescendants)
	resourceRepo.AssertExpectations(t)
}

// Test: Get Permission
func TestIAMService_GetPermission(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	permID := uuid.New()
	expectedPerm := &domain.Permission{
		ID:      permID,
		Name:    "storage.read",
		Service: "storage",
	}

	// Mock expectations
	permissionRepo.On("GetByID", permID).Return(expectedPerm, nil)

	// Get permission
	perm, err := service.GetPermission(permID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPerm, perm)
	permissionRepo.AssertExpectations(t)
}

// Test: List Permissions
func TestIAMService_ListPermissions(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	expectedPerms := []domain.Permission{
		{ID: uuid.New(), Name: "storage.read", Service: "storage"},
		{ID: uuid.New(), Name: "storage.write", Service: "storage"},
	}

	// Mock expectations
	permissionRepo.On("List", "storage", 10, 0).Return(expectedPerms, nil)

	// List permissions
	perms, err := service.ListPermissions("storage", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, perms, 2)
	permissionRepo.AssertExpectations(t)
}

// Test: Get Role
func TestIAMService_GetRole(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	roleID := uuid.New()
	expectedRole := &domain.Role{
		ID:    roleID,
		Name:  "roles/viewer",
		Title: "Viewer",
	}

	// Mock expectations
	roleRepo.On("GetByID", roleID).Return(expectedRole, nil)

	// Get role
	role, err := service.GetRole(roleID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRole, role)
	roleRepo.AssertExpectations(t)
}

// Test: Update Role
func TestIAMService_UpdateRole(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	roleID := uuid.New()
	role := &domain.Role{
		ID:    roleID,
		Name:  "roles/editor",
		Title: "Editor Updated",
	}
	permIDs := []uuid.UUID{uuid.New(), uuid.New()}
	perms := []domain.Permission{
		{ID: permIDs[0], Name: "storage.read"},
		{ID: permIDs[1], Name: "storage.write"},
	}

	// Mock expectations
	roleRepo.On("GetByID", roleID).Return(role, nil)
	permissionRepo.On("GetByIDs", permIDs).Return(perms, nil)
	roleRepo.On("Update", mock.AnythingOfType("*domain.Role")).Return(nil)

	// Update role
	updatedRole, err := service.UpdateRole(roleID, role.Title, role.Description, permIDs)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, updatedRole)
	roleRepo.AssertExpectations(t)
	permissionRepo.AssertExpectations(t)
}

// Test: Delete Role
func TestIAMService_DeleteRole(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	roleID := uuid.New()

	// Mock expectations
	roleRepo.On("Delete", roleID).Return(nil)

	// Delete role
	err := service.DeleteRole(roleID)

	// Assert
	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

// Test: List Roles
func TestIAMService_ListRoles(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	expectedRoles := []domain.Role{
		{ID: uuid.New(), Name: "roles/viewer", Title: "Viewer"},
		{ID: uuid.New(), Name: "roles/editor", Title: "Editor"},
	}

	// Mock expectations
	roleRepo.On("List", true, 10, 0).Return(expectedRoles, nil)

	// List roles
	roles, err := service.ListRoles(true, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	roleRepo.AssertExpectations(t)
}

// Test: Update Policy
func TestIAMService_UpdatePolicy(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	policyID := uuid.New()
	resourceID := uuid.New()
	roleID := uuid.New()

	existingPolicy := &domain.Policy{
		ID:         policyID,
		ResourceID: resourceID,
		ETag:       "old-etag",
		Bindings: []domain.Binding{
			{ID: uuid.New(), RoleID: roleID},
		},
	}

	newBindings := []domain.Binding{
		{
			RoleID:  roleID,
			Members: toJSON([]string{"user:alice@example.com"}),
		},
	}

	// Mock expectations
	policyRepo.On("GetByResourceID", resourceID).Return(existingPolicy, nil)
	bindingRepo.On("Delete", mock.AnythingOfType("uuid.UUID")).Return(nil)
	bindingRepo.On("Create", mock.AnythingOfType("*domain.Binding")).Return(nil)
	policyRepo.On("Update", mock.AnythingOfType("*domain.Policy")).Return(nil)

	updatedPolicy := &domain.Policy{
		ID:         policyID,
		ResourceID: resourceID,
		ETag:       "new-etag",
		Bindings:   newBindings,
	}
	policyRepo.On("GetByID", policyID).Return(updatedPolicy, nil)

	// Update policy
	policy, err := service.UpdatePolicy(resourceID, newBindings, "old-etag")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, policy)
	policyRepo.AssertExpectations(t)
}

// Test: List Policies
func TestIAMService_ListPolicies(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	parentID := uuid.New()
	expectedPolicies := []domain.Policy{
		{ID: uuid.New(), ResourceID: uuid.New()},
		{ID: uuid.New(), ResourceID: uuid.New()},
	}

	// Mock expectations
	policyRepo.On("List", &parentID, 10, 0).Return(expectedPolicies, nil)

	// List policies
	policies, err := service.ListPolicies(&parentID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
	policyRepo.AssertExpectations(t)
}

// Test: Create Binding
func TestIAMService_CreateBinding(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	policyID := uuid.New()
	roleID := uuid.New()
	members := []string{"user:alice@example.com", "user:bob@example.com"}

	existingPolicy := &domain.Policy{
		ID:         policyID,
		ResourceID: resourceID,
		Version:    1,
	}

	// Mock expectations
	policyRepo.On("GetByResourceID", resourceID).Return(existingPolicy, nil)
	bindingRepo.On("Create", mock.AnythingOfType("*domain.Binding")).Return(nil).Run(func(args mock.Arguments) {
		binding := args.Get(0).(*domain.Binding)
		binding.ID = uuid.New()
	})

	createdBinding := &domain.Binding{
		ID:       uuid.New(),
		PolicyID: policyID,
		RoleID:   roleID,
		Members:  toJSON(members),
	}
	bindingRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(createdBinding, nil)

	// Create binding
	binding, err := service.CreateBinding(resourceID, roleID, members, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, binding)
	bindingRepo.AssertExpectations(t)
}

// Test: Delete Binding
func TestIAMService_DeleteBinding(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	bindingID := uuid.New()

	// Mock expectations
	bindingRepo.On("Delete", bindingID).Return(nil)

	// Delete binding
	err := service.DeleteBinding(bindingID)

	// Assert
	assert.NoError(t, err)
	bindingRepo.AssertExpectations(t)
}

// Test: List Bindings
func TestIAMService_ListBindings(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	expectedBindings := []domain.Binding{
		{ID: uuid.New(), PolicyID: uuid.New(), RoleID: uuid.New()},
		{ID: uuid.New(), PolicyID: uuid.New(), RoleID: uuid.New()},
	}

	// Mock expectations
	bindingRepo.On("ListByResourceID", resourceID, 10, 0).Return(expectedBindings, nil)

	// List bindings
	bindings, err := service.ListBindings(resourceID, "", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, bindings, 2)
	bindingRepo.AssertExpectations(t)
}
