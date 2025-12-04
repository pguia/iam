package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindingRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create dependencies
	resource := &domain.Resource{Type: "project", Name: "test"}
	require.NoError(t, resourceRepo.Create(resource))

	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	role := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, roleRepo.Create(role))

	// Create binding
	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:alice@example.com", "user:bob@example.com"]`),
	}

	err := bindingRepo.Create(binding)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, binding.ID)
}

func TestBindingRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)
	permRepo := NewPermissionRepository(db)

	// Create dependencies
	resource := &domain.Resource{Type: "bucket", Name: "data"}
	require.NoError(t, resourceRepo.Create(resource))

	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	perm := &domain.Permission{Name: "storage.read", Service: "storage"}
	require.NoError(t, permRepo.Create(perm))

	role := &domain.Role{Name: "roles/reader", Title: "Reader"}
	require.NoError(t, roleRepo.Create(role))
	require.NoError(t, roleRepo.AddPermissions(role.ID, []uuid.UUID{perm.ID}))

	// Create binding
	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:charlie@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding))

	// Get by ID
	retrieved, err := bindingRepo.GetByID(binding.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, binding.ID, retrieved.ID)
	assert.NotNil(t, retrieved.Role)
	assert.Equal(t, "roles/reader", retrieved.Role.Name)
	assert.Len(t, retrieved.Role.Permissions, 1)
	assert.Equal(t, "storage.read", retrieved.Role.Permissions[0].Name)
}

func TestBindingRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBindingRepository(db)

	// Try to get non-existent binding
	retrieved, err := repo.GetByID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestBindingRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create dependencies
	resource := &domain.Resource{Type: "instance", Name: "vm"}
	require.NoError(t, resourceRepo.Create(resource))

	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	role := &domain.Role{Name: "roles/admin", Title: "Admin"}
	require.NoError(t, roleRepo.Create(role))

	// Create binding
	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:admin@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding))

	// Delete the binding
	err := bindingRepo.Delete(binding.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var count int64
	db.Unscoped().Model(&domain.Binding{}).Where("id = ?", binding.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify not found with normal query
	retrieved, err := bindingRepo.GetByID(binding.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestBindingRepository_ListByResourceID(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resources
	resource1 := &domain.Resource{Type: "project", Name: "proj1"}
	resource2 := &domain.Resource{Type: "project", Name: "proj2"}
	require.NoError(t, resourceRepo.Create(resource1))
	require.NoError(t, resourceRepo.Create(resource2))

	// Create policies
	policy1 := &domain.Policy{ResourceID: resource1.ID}
	policy2 := &domain.Policy{ResourceID: resource2.ID}
	require.NoError(t, policyRepo.Create(policy1))
	require.NoError(t, policyRepo.Create(policy2))

	// Create roles
	role1 := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	role2 := &domain.Role{Name: "roles/editor", Title: "Editor"}
	require.NoError(t, roleRepo.Create(role1))
	require.NoError(t, roleRepo.Create(role2))

	// Create bindings for resource1
	binding1 := &domain.Binding{
		PolicyID: policy1.ID,
		RoleID:   role1.ID,
		Members:  []byte(`["user:user1@example.com"]`),
	}
	binding2 := &domain.Binding{
		PolicyID: policy1.ID,
		RoleID:   role2.ID,
		Members:  []byte(`["user:user2@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding1))
	require.NoError(t, bindingRepo.Create(binding2))

	// Create binding for resource2
	binding3 := &domain.Binding{
		PolicyID: policy2.ID,
		RoleID:   role1.ID,
		Members:  []byte(`["user:user3@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding3))

	// List bindings for resource1
	retrieved, err := bindingRepo.ListByResourceID(resource1.ID, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestBindingRepository_ListByResourceID_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resource
	resource := &domain.Resource{Type: "bucket", Name: "data"}
	require.NoError(t, resourceRepo.Create(resource))

	// Create policy
	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	// Create role
	role := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, roleRepo.Create(role))

	// Create 10 bindings
	for i := 0; i < 10; i++ {
		binding := &domain.Binding{
			PolicyID: policy.ID,
			RoleID:   role.ID,
			Members:  []byte(`["user:test@example.com"]`),
		}
		require.NoError(t, bindingRepo.Create(binding))
	}

	// Test limit
	retrieved, err := bindingRepo.ListByResourceID(resource.ID, 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = bindingRepo.ListByResourceID(resource.ID, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)
}

func TestBindingRepository_ListByPrincipal(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resources
	resource := &domain.Resource{Type: "project", Name: "proj"}
	require.NoError(t, resourceRepo.Create(resource))

	// Create policy
	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	// Create roles
	role1 := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	role2 := &domain.Role{Name: "roles/editor", Title: "Editor"}
	require.NoError(t, roleRepo.Create(role1))
	require.NoError(t, roleRepo.Create(role2))

	// Create bindings with different members
	binding1 := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role1.ID,
		Members:  []byte(`["user:alice@example.com", "user:bob@example.com"]`),
	}
	binding2 := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role2.ID,
		Members:  []byte(`["user:alice@example.com", "user:charlie@example.com"]`),
	}
	binding3 := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role1.ID,
		Members:  []byte(`["user:david@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding1))
	require.NoError(t, bindingRepo.Create(binding2))
	require.NoError(t, bindingRepo.Create(binding3))

	// List bindings for alice (should get 2)
	retrieved, err := bindingRepo.ListByPrincipal("user:alice@example.com", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// List bindings for bob (should get 1)
	retrieved, err = bindingRepo.ListByPrincipal("user:bob@example.com", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)

	// List bindings for non-existent user
	retrieved, err = bindingRepo.ListByPrincipal("user:nobody@example.com", 0, 0)
	assert.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestBindingRepository_ListByPrincipal_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resource
	resource := &domain.Resource{Type: "project", Name: "proj"}
	require.NoError(t, resourceRepo.Create(resource))

	// Create policy
	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	// Create role
	role := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, roleRepo.Create(role))

	// Create 10 bindings all containing the same user
	for i := 0; i < 10; i++ {
		binding := &domain.Binding{
			PolicyID: policy.ID,
			RoleID:   role.ID,
			Members:  []byte(`["user:alice@example.com"]`),
		}
		require.NoError(t, bindingRepo.Create(binding))
	}

	// Test limit
	retrieved, err := bindingRepo.ListByPrincipal("user:alice@example.com", 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = bindingRepo.ListByPrincipal("user:alice@example.com", 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)
}

func TestBindingRepository_GetByPolicyAndPrincipal(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resource
	resource := &domain.Resource{Type: "bucket", Name: "data"}
	require.NoError(t, resourceRepo.Create(resource))

	// Create policy
	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	// Create roles
	role1 := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	role2 := &domain.Role{Name: "roles/editor", Title: "Editor"}
	require.NoError(t, roleRepo.Create(role1))
	require.NoError(t, roleRepo.Create(role2))

	// Create bindings
	binding1 := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role1.ID,
		Members:  []byte(`["user:alice@example.com"]`),
	}
	binding2 := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role2.ID,
		Members:  []byte(`["user:alice@example.com", "user:bob@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding1))
	require.NoError(t, bindingRepo.Create(binding2))

	// Get bindings for alice on this policy
	retrieved, err := bindingRepo.GetByPolicyAndPrincipal(policy.ID, "user:alice@example.com")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Get bindings for bob on this policy
	retrieved, err = bindingRepo.GetByPolicyAndPrincipal(policy.ID, "user:bob@example.com")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, role2.ID, retrieved[0].RoleID)
}

func TestBindingRepository_GetByPolicyAndPrincipal_NotFound(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resource
	resource := &domain.Resource{Type: "bucket", Name: "data"}
	require.NoError(t, resourceRepo.Create(resource))

	// Create policy
	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	// Try to get bindings for non-existent principal
	retrieved, err := bindingRepo.GetByPolicyAndPrincipal(policy.ID, "user:nobody@example.com")
	assert.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestBindingRepository_Create_WithCondition(t *testing.T) {
	db := setupTestDB(t)
	bindingRepo := NewBindingRepository(db)
	policyRepo := NewPolicyRepository(db)
	roleRepo := NewRoleRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create dependencies
	resource := &domain.Resource{Type: "project", Name: "proj"}
	require.NoError(t, resourceRepo.Create(resource))

	policy := &domain.Policy{ResourceID: resource.ID}
	require.NoError(t, policyRepo.Create(policy))

	role := &domain.Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, roleRepo.Create(role))

	// Create binding
	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:alice@example.com"]`),
	}
	require.NoError(t, bindingRepo.Create(binding))

	// Create condition for the binding
	condition := &domain.Condition{
		BindingID:   binding.ID,
		Title:       "Test Condition",
		Description: "Only during business hours",
		Expression:  "request.time.hour >= 9 && request.time.hour < 17",
	}
	require.NoError(t, db.Create(condition).Error)

	// Get binding with condition
	retrieved, err := bindingRepo.GetByID(binding.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.NotNil(t, retrieved.Condition)
	assert.Equal(t, "Test Condition", retrieved.Condition.Title)
	assert.Equal(t, "Only during business hours", retrieved.Condition.Description)
}
