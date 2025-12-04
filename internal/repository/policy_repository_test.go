package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a resource first
	resource := &domain.Resource{
		Type: "project",
		Name: "test-project",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a policy
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}

	err = policyRepo.Create(policy)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, policy.ID)
	assert.NotEmpty(t, policy.ETag)
}

func TestPolicyRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "bucket",
		Name: "my-bucket",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a policy
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	err = policyRepo.Create(policy)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := policyRepo.GetByID(policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, policy.ID, retrieved.ID)
	assert.Equal(t, resource.ID, retrieved.ResourceID)
	assert.NotNil(t, retrieved.Resource)
	assert.Equal(t, "my-bucket", retrieved.Resource.Name)
}

func TestPolicyRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPolicyRepository(db)

	// Try to get non-existent policy
	retrieved, err := repo.GetByID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPolicyRepository_GetByResourceID(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "dataset",
		Name: "analytics-data",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a policy
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	err = policyRepo.Create(policy)
	require.NoError(t, err)

	// Get by resource ID
	retrieved, err := policyRepo.GetByResourceID(resource.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, policy.ID, retrieved.ID)
	assert.Equal(t, resource.ID, retrieved.ResourceID)
}

func TestPolicyRepository_GetByResourceID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPolicyRepository(db)

	// Try to get policy for non-existent resource
	retrieved, err := repo.GetByResourceID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPolicyRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "instance",
		Name: "compute-1",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a policy
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	err = policyRepo.Create(policy)
	require.NoError(t, err)

	originalETag := policy.ETag
	originalVersion := policy.Version

	// Update the policy (this should trigger BeforeUpdate hook)
	err = policyRepo.Update(policy)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := policyRepo.GetByID(policy.ID)
	assert.NoError(t, err)
	assert.NotEqual(t, originalETag, retrieved.ETag) // ETag should change
	assert.Equal(t, originalVersion+1, retrieved.Version) // Version should increment
}

func TestPolicyRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "table",
		Name: "users-table",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a policy
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	err = policyRepo.Create(policy)
	require.NoError(t, err)

	// Delete the policy
	err = policyRepo.Delete(policy.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var count int64
	db.Unscoped().Model(&domain.Policy{}).Where("id = ?", policy.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify not found with normal query
	retrieved, err := policyRepo.GetByID(policy.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPolicyRepository_List(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create resources
	resources := []*domain.Resource{
		{Type: "project", Name: "project-1"},
		{Type: "project", Name: "project-2"},
		{Type: "project", Name: "project-3"},
	}

	for _, resource := range resources {
		err := resourceRepo.Create(resource)
		require.NoError(t, err)

		// Create policy for each resource
		policy := &domain.Policy{
			ResourceID: resource.ID,
			Version:    1,
		}
		err = policyRepo.Create(policy)
		require.NoError(t, err)
	}

	// List all policies
	retrieved, err := policyRepo.List(nil, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestPolicyRepository_List_WithParentResourceID(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create parent resource
	parent := &domain.Resource{
		Type: "organization",
		Name: "my-org",
	}
	err := resourceRepo.Create(parent)
	require.NoError(t, err)

	// Create child resources
	child1 := &domain.Resource{
		Type:     "project",
		Name:     "project-a",
		ParentID: &parent.ID,
	}
	err = resourceRepo.Create(child1)
	require.NoError(t, err)

	child2 := &domain.Resource{
		Type:     "project",
		Name:     "project-b",
		ParentID: &parent.ID,
	}
	err = resourceRepo.Create(child2)
	require.NoError(t, err)

	// Create another resource without parent
	orphan := &domain.Resource{
		Type: "project",
		Name: "orphan-project",
	}
	err = resourceRepo.Create(orphan)
	require.NoError(t, err)

	// Create policies for all resources
	for _, resource := range []*domain.Resource{child1, child2, orphan} {
		policy := &domain.Policy{
			ResourceID: resource.ID,
			Version:    1,
		}
		err = policyRepo.Create(policy)
		require.NoError(t, err)
	}

	// List policies for children of parent only
	retrieved, err := policyRepo.List(&parent.ID, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2) // Should only get child1 and child2 policies
}

func TestPolicyRepository_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)

	// Create 10 resources and policies
	for i := 1; i <= 10; i++ {
		resource := &domain.Resource{
			Type: "resource",
			Name: "resource",
		}
		err := resourceRepo.Create(resource)
		require.NoError(t, err)

		policy := &domain.Policy{
			ResourceID: resource.ID,
			Version:    1,
		}
		err = policyRepo.Create(policy)
		require.NoError(t, err)
	}

	// Test limit
	retrieved, err := policyRepo.List(nil, 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = policyRepo.List(nil, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test limit and offset
	retrieved, err = policyRepo.List(nil, 3, 7)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestPolicyRepository_Create_WithBindings(t *testing.T) {
	db := setupTestDB(t)
	policyRepo := NewPolicyRepository(db)
	resourceRepo := NewResourceRepository(db)
	roleRepo := NewRoleRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "bucket",
		Name: "data-bucket",
	}
	err := resourceRepo.Create(resource)
	require.NoError(t, err)

	// Create a role
	role := &domain.Role{
		Name:  "roles/viewer",
		Title: "Viewer",
	}
	err = roleRepo.Create(role)
	require.NoError(t, err)

	// Create a policy with bindings
	policy := &domain.Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	err = policyRepo.Create(policy)
	require.NoError(t, err)

	// Create a binding
	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:alice@example.com"]`),
	}
	err = db.Create(binding).Error
	require.NoError(t, err)

	// Get policy with bindings
	retrieved, err := policyRepo.GetByID(policy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Len(t, retrieved.Bindings, 1)
	assert.Equal(t, role.ID, retrieved.Bindings[0].RoleID)
}
