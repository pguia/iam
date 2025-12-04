package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	// Get test database connection string from env or use default
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dsn := fmt.Sprintf("host=%s port=5432 user=postgres password=postgres dbname=iam_db sslmode=disable",
		dbHost)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Create a unique schema for this test to avoid conflicts
	schemaName := fmt.Sprintf("test_%s", uuid.New().String()[:8])
	db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))

	// Auto-migrate all tables
	err = db.AutoMigrate(
		&domain.Resource{},
		&domain.Permission{},
		&domain.Role{},
		&domain.Policy{},
		&domain.Binding{},
		&domain.Condition{},
	)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	})

	return db
}

func TestRoleRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	role := &domain.Role{
		Name:        "roles/storage.admin",
		Title:       "Storage Admin",
		Description: "Full access to storage resources",
		IsCustom:    false,
	}

	err := repo.Create(role)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
}

func TestRoleRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create a role
	role := &domain.Role{
		Name:        "roles/compute.admin",
		Title:       "Compute Admin",
		Description: "Full access to compute resources",
		IsCustom:    false,
	}
	err := repo.Create(role)
	require.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(role.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, role.Name, retrieved.Name)
	assert.Equal(t, role.Title, retrieved.Title)
}

func TestRoleRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Try to get non-existent role
	retrieved, err := repo.GetByID(uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestRoleRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create a role
	role := &domain.Role{
		Name:        "roles/database.admin",
		Title:       "Database Admin",
		Description: "Full access to database resources",
		IsCustom:    false,
	}
	err := repo.Create(role)
	require.NoError(t, err)

	// Get by name
	retrieved, err := repo.GetByName("roles/database.admin")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, role.ID, retrieved.ID)
	assert.Equal(t, role.Title, retrieved.Title)
}

func TestRoleRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Try to get non-existent role
	retrieved, err := repo.GetByName("roles/nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestRoleRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create a role
	role := &domain.Role{
		Name:        "roles/network.admin",
		Title:       "Network Admin",
		Description: "Manage network resources",
		IsCustom:    true,
	}
	err := repo.Create(role)
	require.NoError(t, err)

	// Update the role
	role.Title = "Network Administrator"
	role.Description = "Full access to network resources"
	err = repo.Update(role)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(role.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Network Administrator", retrieved.Title)
	assert.Equal(t, "Full access to network resources", retrieved.Description)
}

func TestRoleRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create a role
	role := &domain.Role{
		Name:        "roles/temp.role",
		Title:       "Temporary Role",
		Description: "A temporary role",
		IsCustom:    true,
	}
	err := repo.Create(role)
	require.NoError(t, err)

	// Delete the role
	err = repo.Delete(role.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var count int64
	db.Unscoped().Model(&domain.Role{}).Where("id = ?", role.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify not found with normal query
	retrieved, err := repo.GetByID(role.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestRoleRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create multiple roles
	roles := []*domain.Role{
		{Name: "roles/role1", Title: "Role 1", IsCustom: false},
		{Name: "roles/role2", Title: "Role 2", IsCustom: false},
		{Name: "roles/role3", Title: "Role 3", IsCustom: true},
		{Name: "roles/role4", Title: "Role 4", IsCustom: true},
	}

	for _, role := range roles {
		err := repo.Create(role)
		require.NoError(t, err)
	}

	// List all roles
	retrieved, err := repo.List(true, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 4)

	// List only predefined roles
	retrieved, err = repo.List(false, 0, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestRoleRepository_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create multiple roles
	for i := 1; i <= 10; i++ {
		role := &domain.Role{
			Name:     "roles/role" + string(rune(i)),
			Title:    "Role",
			IsCustom: false,
		}
		err := repo.Create(role)
		require.NoError(t, err)
	}

	// Test limit
	retrieved, err := repo.List(true, 5, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test offset
	retrieved, err = repo.List(true, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 5)

	// Test limit and offset
	retrieved, err = repo.List(true, 3, 7)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestRoleRepository_AddPermissions(t *testing.T) {
	db := setupTestDB(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)

	// Create a role
	role := &domain.Role{
		Name:     "roles/test.role",
		Title:    "Test Role",
		IsCustom: true,
	}
	err := roleRepo.Create(role)
	require.NoError(t, err)

	// Create permissions
	perm1 := &domain.Permission{
		Name:        "storage.buckets.create",
		Description: "Create storage buckets",
		Service:     "storage",
	}
	perm2 := &domain.Permission{
		Name:        "storage.buckets.delete",
		Description: "Delete storage buckets",
		Service:     "storage",
	}
	err = permRepo.Create(perm1)
	require.NoError(t, err)
	err = permRepo.Create(perm2)
	require.NoError(t, err)

	// Add permissions to role
	err = roleRepo.AddPermissions(role.ID, []uuid.UUID{perm1.ID, perm2.ID})
	assert.NoError(t, err)

	// Verify permissions were added
	permissions, err := roleRepo.GetPermissions(role.ID)
	assert.NoError(t, err)
	assert.Len(t, permissions, 2)
}

func TestRoleRepository_AddPermissions_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Try to add permissions to non-existent role
	err := repo.AddPermissions(uuid.New(), []uuid.UUID{uuid.New()})
	assert.Error(t, err)
}

func TestRoleRepository_RemovePermissions(t *testing.T) {
	db := setupTestDB(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)

	// Create a role
	role := &domain.Role{
		Name:     "roles/test.role",
		Title:    "Test Role",
		IsCustom: true,
	}
	err := roleRepo.Create(role)
	require.NoError(t, err)

	// Create and add permissions
	perm1 := &domain.Permission{
		Name:        "compute.instances.create",
		Description: "Create compute instances",
		Service:     "compute",
	}
	perm2 := &domain.Permission{
		Name:        "compute.instances.delete",
		Description: "Delete compute instances",
		Service:     "compute",
	}
	err = permRepo.Create(perm1)
	require.NoError(t, err)
	err = permRepo.Create(perm2)
	require.NoError(t, err)

	err = roleRepo.AddPermissions(role.ID, []uuid.UUID{perm1.ID, perm2.ID})
	require.NoError(t, err)

	// Remove one permission
	err = roleRepo.RemovePermissions(role.ID, []uuid.UUID{perm1.ID})
	assert.NoError(t, err)

	// Verify only one permission remains
	permissions, err := roleRepo.GetPermissions(role.ID)
	assert.NoError(t, err)
	assert.Len(t, permissions, 1)
	assert.Equal(t, perm2.ID, permissions[0].ID)
}

func TestRoleRepository_GetPermissions(t *testing.T) {
	db := setupTestDB(t)
	roleRepo := NewRoleRepository(db)
	permRepo := NewPermissionRepository(db)

	// Create a role
	role := &domain.Role{
		Name:     "roles/viewer",
		Title:    "Viewer",
		IsCustom: false,
	}
	err := roleRepo.Create(role)
	require.NoError(t, err)

	// Create permissions
	perms := []*domain.Permission{
		{Name: "resource.get", Service: "iam"},
		{Name: "resource.list", Service: "iam"},
	}
	for _, perm := range perms {
		err := permRepo.Create(perm)
		require.NoError(t, err)
	}

	// Add permissions
	permIDs := []uuid.UUID{perms[0].ID, perms[1].ID}
	err = roleRepo.AddPermissions(role.ID, permIDs)
	require.NoError(t, err)

	// Get permissions
	retrieved, err := roleRepo.GetPermissions(role.ID)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestRoleRepository_GetPermissions_NoPermissions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Create a role without permissions
	role := &domain.Role{
		Name:     "roles/empty",
		Title:    "Empty Role",
		IsCustom: true,
	}
	err := repo.Create(role)
	require.NoError(t, err)

	// Get permissions
	permissions, err := repo.GetPermissions(role.ID)
	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

func TestRoleRepository_GetPermissions_RoleNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRoleRepository(db)

	// Try to get permissions for non-existent role
	_, err := repo.GetPermissions(uuid.New())
	assert.Error(t, err)
}
