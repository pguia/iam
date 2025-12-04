package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	resource := &domain.Resource{
		Type: "project",
		Name: "my-project",
		Attributes: map[string]string{
			"region": "us-west-1",
		},
	}

	err := repo.Create(resource)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resource.ID)
}

func TestResourceRepository_Create_WithParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create parent
	parent := &domain.Resource{
		Type: "organization",
		Name: "my-org",
	}
	err := repo.Create(parent)
	require.NoError(t, err)

	// Create child
	child := &domain.Resource{
		Type:     "project",
		Name:     "my-project",
		ParentID: &parent.ID,
	}
	err = repo.Create(child)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, child.ID)
	assert.Equal(t, parent.ID, *child.ParentID)
}

func TestResourceRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "bucket",
		Name: "data-bucket",
		Attributes: map[string]string{
			"location": "europe-west1",
		},
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(resource.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, resource.Name, retrieved.Name)
	assert.Equal(t, resource.Type, retrieved.Type)
	assert.Equal(t, "europe-west1", retrieved.Attributes["location"])
}

func TestResourceRepository_GetByID_WithParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create parent
	parent := &domain.Resource{
		Type: "folder",
		Name: "engineering",
	}
	err := repo.Create(parent)
	require.NoError(t, err)

	// Create child
	child := &domain.Resource{
		Type:     "project",
		Name:     "backend-api",
		ParentID: &parent.ID,
	}
	err = repo.Create(child)
	require.NoError(t, err)

	// Get child by ID
	retrieved, err := repo.GetByID(child.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.NotNil(t, retrieved.Parent)
	assert.Equal(t, parent.Name, retrieved.Parent.Name)
}

func TestResourceRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Try to get non-existent resource
	retrieved, err := repo.GetByID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestResourceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "instance",
		Name: "web-server",
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Update the resource
	resource.Name = "web-server-updated"
	resource.Attributes = map[string]string{
		"environment": "production",
	}
	err = repo.Update(resource)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(resource.ID)
	assert.NoError(t, err)
	assert.Equal(t, "web-server-updated", retrieved.Name)
	assert.Equal(t, "production", retrieved.Attributes["environment"])
}

func TestResourceRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create a resource
	resource := &domain.Resource{
		Type: "temp-resource",
		Name: "temporary",
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Delete the resource
	err = repo.Delete(resource.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var count int64
	db.Unscoped().Model(&domain.Resource{}).Where("id = ?", resource.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify not found with normal query
	retrieved, err := repo.GetByID(resource.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestResourceRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create multiple resources
	resources := []*domain.Resource{
		{Type: "project", Name: "project-1"},
		{Type: "project", Name: "project-2"},
		{Type: "bucket", Name: "bucket-1"},
	}

	for _, resource := range resources {
		err := repo.Create(resource)
		require.NoError(t, err)
	}

	// List all resources
	retrieved, err := repo.List(nil, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestResourceRepository_List_FilterByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create resources of different types
	resources := []*domain.Resource{
		{Type: "project", Name: "project-1"},
		{Type: "project", Name: "project-2"},
		{Type: "bucket", Name: "bucket-1"},
		{Type: "bucket", Name: "bucket-2"},
		{Type: "instance", Name: "instance-1"},
	}

	for _, resource := range resources {
		err := repo.Create(resource)
		require.NoError(t, err)
	}

	// List only projects
	retrieved, err := repo.List(nil, "project", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	for _, r := range retrieved {
		assert.Equal(t, "project", r.Type)
	}

	// List only buckets
	retrieved, err = repo.List(nil, "bucket", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	for _, r := range retrieved {
		assert.Equal(t, "bucket", r.Type)
	}
}

func TestResourceRepository_List_FilterByParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create parent
	parent := &domain.Resource{
		Type: "organization",
		Name: "my-org",
	}
	err := repo.Create(parent)
	require.NoError(t, err)

	// Create children
	children := []*domain.Resource{
		{Type: "project", Name: "child-1", ParentID: &parent.ID},
		{Type: "project", Name: "child-2", ParentID: &parent.ID},
	}
	for _, child := range children {
		err := repo.Create(child)
		require.NoError(t, err)
	}

	// Create orphan
	orphan := &domain.Resource{
		Type: "project",
		Name: "orphan",
	}
	err = repo.Create(orphan)
	require.NoError(t, err)

	// List children of parent
	retrieved, err := repo.List(&parent.ID, "", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	for _, r := range retrieved {
		assert.Equal(t, parent.ID, *r.ParentID)
	}
}

func TestResourceRepository_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create 10 resources
	for i := 1; i <= 10; i++ {
		resource := &domain.Resource{
			Type: "resource",
			Name: "resource",
		}
		err := repo.Create(resource)
		require.NoError(t, err)
	}

	// Test limit
	retrieved, err := repo.List(nil, "", 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = repo.List(nil, "", 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test limit and offset
	retrieved, err = repo.List(nil, "", 3, 7)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestResourceRepository_GetChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create parent
	parent := &domain.Resource{
		Type: "folder",
		Name: "parent",
	}
	err := repo.Create(parent)
	require.NoError(t, err)

	// Create children
	child1 := &domain.Resource{
		Type:     "project",
		Name:     "child-1",
		ParentID: &parent.ID,
	}
	child2 := &domain.Resource{
		Type:     "project",
		Name:     "child-2",
		ParentID: &parent.ID,
	}
	err = repo.Create(child1)
	require.NoError(t, err)
	err = repo.Create(child2)
	require.NoError(t, err)

	// Get children
	children, err := repo.GetChildren(parent.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 2)
}

func TestResourceRepository_GetChildren_NoChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create resource without children
	resource := &domain.Resource{
		Type: "project",
		Name: "lonely",
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Get children
	children, err := repo.GetChildren(resource.ID)
	assert.NoError(t, err)
	assert.Empty(t, children)
}

func TestResourceRepository_GetAncestors(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create hierarchy: org -> folder -> project
	org := &domain.Resource{
		Type: "organization",
		Name: "my-org",
	}
	err := repo.Create(org)
	require.NoError(t, err)

	folder := &domain.Resource{
		Type:     "folder",
		Name:     "engineering",
		ParentID: &org.ID,
	}
	err = repo.Create(folder)
	require.NoError(t, err)

	project := &domain.Resource{
		Type:     "project",
		Name:     "backend",
		ParentID: &folder.ID,
	}
	err = repo.Create(project)
	require.NoError(t, err)

	// Get ancestors of project
	ancestors, err := repo.GetAncestors(project.ID)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 2)

	// Verify ancestor IDs (should be folder and org)
	ancestorIDs := make(map[uuid.UUID]bool)
	for _, a := range ancestors {
		ancestorIDs[a.ID] = true
	}
	assert.True(t, ancestorIDs[folder.ID])
	assert.True(t, ancestorIDs[org.ID])
}

func TestResourceRepository_GetAncestors_NoParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create resource without parent
	resource := &domain.Resource{
		Type: "organization",
		Name: "root",
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Get ancestors (should be empty)
	ancestors, err := repo.GetAncestors(resource.ID)
	assert.NoError(t, err)
	assert.Empty(t, ancestors)
}

func TestResourceRepository_GetDescendants(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create hierarchy: org -> folder -> project1, project2
	org := &domain.Resource{
		Type: "organization",
		Name: "my-org",
	}
	err := repo.Create(org)
	require.NoError(t, err)

	folder := &domain.Resource{
		Type:     "folder",
		Name:     "engineering",
		ParentID: &org.ID,
	}
	err = repo.Create(folder)
	require.NoError(t, err)

	project1 := &domain.Resource{
		Type:     "project",
		Name:     "backend",
		ParentID: &folder.ID,
	}
	err = repo.Create(project1)
	require.NoError(t, err)

	project2 := &domain.Resource{
		Type:     "project",
		Name:     "frontend",
		ParentID: &folder.ID,
	}
	err = repo.Create(project2)
	require.NoError(t, err)

	// Get descendants of org
	descendants, err := repo.GetDescendants(org.ID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 3) // folder, project1, project2

	// Verify descendant IDs
	descendantIDs := make(map[uuid.UUID]bool)
	for _, d := range descendants {
		descendantIDs[d.ID] = true
	}
	assert.True(t, descendantIDs[folder.ID])
	assert.True(t, descendantIDs[project1.ID])
	assert.True(t, descendantIDs[project2.ID])
}

func TestResourceRepository_GetDescendants_NoChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create resource without children
	resource := &domain.Resource{
		Type: "project",
		Name: "leaf",
	}
	err := repo.Create(resource)
	require.NoError(t, err)

	// Get descendants (should be empty)
	descendants, err := repo.GetDescendants(resource.ID)
	assert.NoError(t, err)
	assert.Empty(t, descendants)
}

func TestResourceRepository_ComplexHierarchy(t *testing.T) {
	db := setupTestDB(t)
	repo := NewResourceRepository(db)

	// Create complex hierarchy:
	// org
	// ├── folder1
	// │   ├── project1
	// │   └── project2
	// └── folder2
	//     └── project3

	org := &domain.Resource{Type: "organization", Name: "org"}
	require.NoError(t, repo.Create(org))

	folder1 := &domain.Resource{Type: "folder", Name: "folder1", ParentID: &org.ID}
	require.NoError(t, repo.Create(folder1))

	folder2 := &domain.Resource{Type: "folder", Name: "folder2", ParentID: &org.ID}
	require.NoError(t, repo.Create(folder2))

	project1 := &domain.Resource{Type: "project", Name: "project1", ParentID: &folder1.ID}
	require.NoError(t, repo.Create(project1))

	project2 := &domain.Resource{Type: "project", Name: "project2", ParentID: &folder1.ID}
	require.NoError(t, repo.Create(project2))

	project3 := &domain.Resource{Type: "project", Name: "project3", ParentID: &folder2.ID}
	require.NoError(t, repo.Create(project3))

	// Test descendants of org (should include all)
	descendants, err := repo.GetDescendants(org.ID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 5)

	// Test descendants of folder1
	descendants, err = repo.GetDescendants(folder1.ID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 2) // project1 and project2

	// Test ancestors of project1
	ancestors, err := repo.GetAncestors(project1.ID)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 2) // folder1 and org

	// Test children of folder1
	children, err := repo.GetChildren(folder1.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 2) // project1 and project2
}
