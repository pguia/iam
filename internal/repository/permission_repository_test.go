package repository

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	permission := &domain.Permission{
		Name:        "storage.buckets.create",
		Description: "Create storage buckets",
		Service:     "storage",
	}

	err := repo.Create(permission)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, permission.ID)
}

func TestPermissionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create a permission
	permission := &domain.Permission{
		Name:        "compute.instances.create",
		Description: "Create compute instances",
		Service:     "compute",
	}
	err := repo.Create(permission)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(permission.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, permission.Name, retrieved.Name)
	assert.Equal(t, permission.Description, retrieved.Description)
	assert.Equal(t, permission.Service, retrieved.Service)
}

func TestPermissionRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Try to get non-existent permission
	retrieved, err := repo.GetByID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPermissionRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create a permission
	permission := &domain.Permission{
		Name:        "database.tables.read",
		Description: "Read database tables",
		Service:     "database",
	}
	err := repo.Create(permission)
	require.NoError(t, err)

	// Get by name
	retrieved, err := repo.GetByName("database.tables.read")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, permission.ID, retrieved.ID)
	assert.Equal(t, permission.Description, retrieved.Description)
}

func TestPermissionRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Try to get non-existent permission
	retrieved, err := repo.GetByName("nonexistent.permission")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPermissionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create a permission
	permission := &domain.Permission{
		Name:        "temp.permission",
		Description: "Temporary permission",
		Service:     "test",
	}
	err := repo.Create(permission)
	require.NoError(t, err)

	// Delete the permission
	err = repo.Delete(permission.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var count int64
	db.Unscoped().Model(&domain.Permission{}).Where("id = ?", permission.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify not found with normal query
	retrieved, err := repo.GetByID(permission.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPermissionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create multiple permissions across different services
	permissions := []*domain.Permission{
		{Name: "storage.buckets.create", Service: "storage"},
		{Name: "storage.buckets.delete", Service: "storage"},
		{Name: "compute.instances.create", Service: "compute"},
		{Name: "compute.instances.delete", Service: "compute"},
		{Name: "database.tables.read", Service: "database"},
	}

	for _, perm := range permissions {
		err := repo.Create(perm)
		require.NoError(t, err)
	}

	// List all permissions
	retrieved, err := repo.List("", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// List permissions for storage service
	retrieved, err = repo.List("storage", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	for _, perm := range retrieved {
		assert.Equal(t, "storage", perm.Service)
	}

	// List permissions for compute service
	retrieved, err = repo.List("compute", 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
	for _, perm := range retrieved {
		assert.Equal(t, "compute", perm.Service)
	}
}

func TestPermissionRepository_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create 10 permissions
	for i := 1; i <= 10; i++ {
		permission := &domain.Permission{
			Name:    fmt.Sprintf("permission.%d", i),
			Service: "test",
		}
		err := repo.Create(permission)
		require.NoError(t, err)
	}

	// Test limit
	retrieved, err := repo.List("", 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = repo.List("", 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test limit and offset
	retrieved, err = repo.List("", 3, 7)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestPermissionRepository_List_ServiceFilter_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create permissions for different services
	for i := 1; i <= 5; i++ {
		perm1 := &domain.Permission{
			Name:    fmt.Sprintf("storage.action.%d", i),
			Service: "storage",
		}
		perm2 := &domain.Permission{
			Name:    fmt.Sprintf("compute.action.%d", i),
			Service: "compute",
		}
		require.NoError(t, repo.Create(perm1))
		require.NoError(t, repo.Create(perm2))
	}

	// List storage permissions with pagination
	retrieved, err := repo.List("storage", 3, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
	for _, perm := range retrieved {
		assert.Equal(t, "storage", perm.Service)
	}

	// List remaining storage permissions
	retrieved, err = repo.List("storage", 3, 3)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestPermissionRepository_GetByIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create multiple permissions
	perms := []*domain.Permission{
		{Name: "perm1", Service: "service1"},
		{Name: "perm2", Service: "service1"},
		{Name: "perm3", Service: "service2"},
		{Name: "perm4", Service: "service2"},
	}

	for _, perm := range perms {
		err := repo.Create(perm)
		require.NoError(t, err)
	}

	// Get specific permissions by IDs
	ids := []uuid.UUID{perms[0].ID, perms[2].ID}
	retrieved, err := repo.GetByIDs(ids)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Verify correct permissions were retrieved
	retrievedIDs := make(map[uuid.UUID]bool)
	for _, perm := range retrieved {
		retrievedIDs[perm.ID] = true
	}
	assert.True(t, retrievedIDs[perms[0].ID])
	assert.True(t, retrievedIDs[perms[2].ID])
}

func TestPermissionRepository_GetByIDs_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Get permissions with empty ID list
	retrieved, err := repo.GetByIDs([]uuid.UUID{})
	assert.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestPermissionRepository_GetByIDs_NonExistent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Try to get non-existent permissions
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	retrieved, err := repo.GetByIDs(ids)
	assert.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestPermissionRepository_GetByIDs_PartialMatch(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create one permission
	perm := &domain.Permission{
		Name:    "existing.permission",
		Service: "test",
	}
	err := repo.Create(perm)
	require.NoError(t, err)

	// Request both existing and non-existing IDs
	ids := []uuid.UUID{perm.ID, uuid.New(), uuid.New()}
	retrieved, err := repo.GetByIDs(ids)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, perm.ID, retrieved[0].ID)
}

func TestPermissionRepository_Create_DuplicateName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPermissionRepository(db)

	// Create first permission
	perm1 := &domain.Permission{
		Name:    "duplicate.permission",
		Service: "test",
	}
	err := repo.Create(perm1)
	require.NoError(t, err)

	// Try to create duplicate
	perm2 := &domain.Permission{
		Name:    "duplicate.permission",
		Service: "test",
	}
	err = repo.Create(perm2)
	assert.Error(t, err) // Should fail due to unique constraint
}
