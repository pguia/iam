package domain

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
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
		&Resource{},
		&Permission{},
		&Role{},
		&Policy{},
		&Binding{},
		&Condition{},
	)
	require.NoError(t, err)

	// Cleanup after test
	t.Cleanup(func() {
		db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	})

	return db
}

// Test Role domain model
func TestRole_TableName(t *testing.T) {
	role := Role{}
	assert.Equal(t, "roles", role.TableName())
}

func TestRole_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	// Test UUID generation on create
	role := &Role{
		Name:  "roles/test.role",
		Title: "Test Role",
	}

	err := db.Create(role).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
}

func TestRole_BeforeCreate_PresetID(t *testing.T) {
	db := setupTestDB(t)

	// Test that preset UUID is preserved
	presetID := uuid.New()
	role := &Role{
		ID:    presetID,
		Name:  "roles/preset.role",
		Title: "Preset Role",
	}

	err := db.Create(role).Error
	assert.NoError(t, err)
	assert.Equal(t, presetID, role.ID)
}

func TestRole_HasPermission(t *testing.T) {
	// Create role with permissions
	role := &Role{
		Name:  "roles/test.role",
		Title: "Test Role",
		Permissions: []Permission{
			{Name: "storage.buckets.create"},
			{Name: "storage.buckets.delete"},
			{Name: "compute.instances.create"},
		},
	}

	// Test HasPermission
	assert.True(t, role.HasPermission("storage.buckets.create"))
	assert.True(t, role.HasPermission("storage.buckets.delete"))
	assert.True(t, role.HasPermission("compute.instances.create"))
	assert.False(t, role.HasPermission("database.tables.read"))
	assert.False(t, role.HasPermission(""))
}

func TestRole_HasPermission_EmptyPermissions(t *testing.T) {
	role := &Role{
		Name:        "roles/empty.role",
		Title:       "Empty Role",
		Permissions: []Permission{},
	}

	assert.False(t, role.HasPermission("any.permission"))
}

// Test Permission domain model
func TestPermission_TableName(t *testing.T) {
	perm := Permission{}
	assert.Equal(t, "permissions", perm.TableName())
}

func TestPermission_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	perm := &Permission{
		Name:    "storage.buckets.create",
		Service: "storage",
	}

	err := db.Create(perm).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, perm.ID)
}

func TestPermission_BeforeCreate_PresetID(t *testing.T) {
	db := setupTestDB(t)

	presetID := uuid.New()
	perm := &Permission{
		ID:      presetID,
		Name:    "storage.buckets.create",
		Service: "storage",
	}

	err := db.Create(perm).Error
	assert.NoError(t, err)
	assert.Equal(t, presetID, perm.ID)
}

// Test Policy domain model
func TestPolicy_TableName(t *testing.T) {
	policy := Policy{}
	assert.Equal(t, "policies", policy.TableName())
}

func TestPolicy_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	// Create resource first
	resource := &Resource{Type: "project", Name: "test"}
	require.NoError(t, db.Create(resource).Error)

	policy := &Policy{
		ResourceID: resource.ID,
	}

	err := db.Create(policy).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, policy.ID)
	assert.NotEmpty(t, policy.ETag)
}

func TestPolicy_BeforeCreate_PresetValues(t *testing.T) {
	db := setupTestDB(t)

	// Create resource first
	resource := &Resource{Type: "project", Name: "test"}
	require.NoError(t, db.Create(resource).Error)

	presetID := uuid.New()
	presetETag := "preset-etag"
	policy := &Policy{
		ID:         presetID,
		ResourceID: resource.ID,
		ETag:       presetETag,
	}

	err := db.Create(policy).Error
	assert.NoError(t, err)
	assert.Equal(t, presetID, policy.ID)
	assert.Equal(t, presetETag, policy.ETag)
}

func TestPolicy_BeforeUpdate(t *testing.T) {
	db := setupTestDB(t)

	// Create resource
	resource := &Resource{Type: "bucket", Name: "data"}
	require.NoError(t, db.Create(resource).Error)

	// Create policy
	policy := &Policy{
		ResourceID: resource.ID,
		Version:    1,
	}
	require.NoError(t, db.Create(policy).Error)

	originalETag := policy.ETag
	originalVersion := policy.Version

	// Update policy
	policy.ResourceID = resource.ID // Trigger update
	err := db.Save(policy).Error
	assert.NoError(t, err)

	// Verify ETag changed and version incremented
	assert.NotEqual(t, originalETag, policy.ETag)
	assert.Equal(t, originalVersion+1, policy.Version)
}

// Test Resource domain model
func TestResource_TableName(t *testing.T) {
	resource := Resource{}
	assert.Equal(t, "resources", resource.TableName())
}

func TestResource_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	resource := &Resource{
		Type: "project",
		Name: "my-project",
	}

	err := db.Create(resource).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, resource.ID)
}

func TestResource_BeforeCreate_PresetID(t *testing.T) {
	db := setupTestDB(t)

	presetID := uuid.New()
	resource := &Resource{
		ID:   presetID,
		Type: "project",
		Name: "my-project",
	}

	err := db.Create(resource).Error
	assert.NoError(t, err)
	assert.Equal(t, presetID, resource.ID)
}

func TestResource_GetAncestors(t *testing.T) {
	db := setupTestDB(t)

	// Create hierarchy: org -> folder -> project
	org := &Resource{Type: "organization", Name: "my-org"}
	require.NoError(t, db.Create(org).Error)

	folder := &Resource{Type: "folder", Name: "engineering", ParentID: &org.ID}
	require.NoError(t, db.Create(folder).Error)

	project := &Resource{Type: "project", Name: "backend", ParentID: &folder.ID}
	require.NoError(t, db.Create(project).Error)

	// Get ancestors of project
	ancestors, err := project.GetAncestors(db)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 2)

	// Verify ancestors (should be folder and org in order)
	assert.Equal(t, folder.ID, ancestors[0].ID)
	assert.Equal(t, org.ID, ancestors[1].ID)
}

func TestResource_GetAncestors_NoParent(t *testing.T) {
	db := setupTestDB(t)

	// Create root resource
	root := &Resource{Type: "organization", Name: "root"}
	require.NoError(t, db.Create(root).Error)

	// Get ancestors (should be empty)
	ancestors, err := root.GetAncestors(db)
	assert.NoError(t, err)
	assert.Empty(t, ancestors)
}

func TestResource_GetAncestors_SingleParent(t *testing.T) {
	db := setupTestDB(t)

	// Create parent and child
	parent := &Resource{Type: "folder", Name: "parent"}
	require.NoError(t, db.Create(parent).Error)

	child := &Resource{Type: "project", Name: "child", ParentID: &parent.ID}
	require.NoError(t, db.Create(child).Error)

	// Get ancestors
	ancestors, err := child.GetAncestors(db)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 1)
	assert.Equal(t, parent.ID, ancestors[0].ID)
}

// Test Binding domain model
func TestBinding_TableName(t *testing.T) {
	binding := Binding{}
	assert.Equal(t, "bindings", binding.TableName())
}

func TestBinding_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	// Create dependencies
	resource := &Resource{Type: "project", Name: "test"}
	require.NoError(t, db.Create(resource).Error)

	policy := &Policy{ResourceID: resource.ID}
	require.NoError(t, db.Create(policy).Error)

	role := &Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, db.Create(role).Error)

	binding := &Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:alice@example.com"]`),
	}

	err := db.Create(binding).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, binding.ID)
}

func TestBinding_GetMembers(t *testing.T) {
	binding := &Binding{
		Members: []byte(`["user:alice@example.com", "user:bob@example.com", "group:admins"]`),
	}

	members, err := binding.GetMembers()
	assert.NoError(t, err)
	assert.Len(t, members, 3)
	assert.Contains(t, members, "user:alice@example.com")
	assert.Contains(t, members, "user:bob@example.com")
	assert.Contains(t, members, "group:admins")
}

func TestBinding_GetMembers_EmptyArray(t *testing.T) {
	binding := &Binding{
		Members: []byte(`[]`),
	}

	members, err := binding.GetMembers()
	assert.NoError(t, err)
	assert.Empty(t, members)
}

func TestBinding_GetMembers_InvalidJSON(t *testing.T) {
	binding := &Binding{
		Members: []byte(`invalid json`),
	}

	members, err := binding.GetMembers()
	assert.Error(t, err)
	assert.Nil(t, members)
}

func TestBinding_HasMember(t *testing.T) {
	binding := &Binding{
		Members: []byte(`["user:alice@example.com", "user:bob@example.com", "group:admins"]`),
	}

	assert.True(t, binding.HasMember("user:alice@example.com"))
	assert.True(t, binding.HasMember("user:bob@example.com"))
	assert.True(t, binding.HasMember("group:admins"))
	assert.False(t, binding.HasMember("user:charlie@example.com"))
	assert.False(t, binding.HasMember(""))
}

func TestBinding_HasMember_EmptyMembers(t *testing.T) {
	binding := &Binding{
		Members: []byte(`[]`),
	}

	assert.False(t, binding.HasMember("user:alice@example.com"))
}

func TestBinding_HasMember_InvalidJSON(t *testing.T) {
	binding := &Binding{
		Members: []byte(`invalid`),
	}

	assert.False(t, binding.HasMember("user:alice@example.com"))
}

// Test Condition domain model
func TestCondition_TableName(t *testing.T) {
	condition := Condition{}
	assert.Equal(t, "conditions", condition.TableName())
}

func TestCondition_BeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	// Create dependencies
	resource := &Resource{Type: "project", Name: "test"}
	require.NoError(t, db.Create(resource).Error)

	policy := &Policy{ResourceID: resource.ID}
	require.NoError(t, db.Create(policy).Error)

	role := &Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, db.Create(role).Error)

	binding := &Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:test@example.com"]`),
	}
	require.NoError(t, db.Create(binding).Error)

	condition := &Condition{
		BindingID:  binding.ID,
		Title:      "Business Hours",
		Expression: "request.time.hour >= 9 && request.time.hour < 17",
	}

	err := db.Create(condition).Error
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, condition.ID)
}

func TestCondition_BeforeCreate_PresetID(t *testing.T) {
	db := setupTestDB(t)

	// Create dependencies
	resource := &Resource{Type: "project", Name: "test"}
	require.NoError(t, db.Create(resource).Error)

	policy := &Policy{ResourceID: resource.ID}
	require.NoError(t, db.Create(policy).Error)

	role := &Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, db.Create(role).Error)

	binding := &Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:test@example.com"]`),
	}
	require.NoError(t, db.Create(binding).Error)

	presetID := uuid.New()
	condition := &Condition{
		ID:         presetID,
		BindingID:  binding.ID,
		Title:      "Test Condition",
		Expression: "true",
	}

	err := db.Create(condition).Error
	assert.NoError(t, err)
	assert.Equal(t, presetID, condition.ID)
}

// Integration tests for relationships
func TestDomain_RolePermissionRelationship(t *testing.T) {
	db := setupTestDB(t)

	// Create permissions
	perm1 := &Permission{Name: "storage.read", Service: "storage"}
	perm2 := &Permission{Name: "storage.write", Service: "storage"}
	require.NoError(t, db.Create(perm1).Error)
	require.NoError(t, db.Create(perm2).Error)

	// Create role
	role := &Role{Name: "roles/storage.admin", Title: "Storage Admin"}
	require.NoError(t, db.Create(role).Error)

	// Add permissions to role
	err := db.Model(role).Association("Permissions").Append([]Permission{*perm1, *perm2})
	require.NoError(t, err)

	// Reload role with permissions
	var loadedRole Role
	err = db.Preload("Permissions").First(&loadedRole, role.ID).Error
	require.NoError(t, err)

	assert.Len(t, loadedRole.Permissions, 2)
	assert.True(t, loadedRole.HasPermission("storage.read"))
	assert.True(t, loadedRole.HasPermission("storage.write"))
}

func TestDomain_PolicyBindingRelationship(t *testing.T) {
	db := setupTestDB(t)

	// Create resource
	resource := &Resource{Type: "bucket", Name: "data"}
	require.NoError(t, db.Create(resource).Error)

	// Create policy
	policy := &Policy{ResourceID: resource.ID}
	require.NoError(t, db.Create(policy).Error)

	// Create role
	role := &Role{Name: "roles/viewer", Title: "Viewer"}
	require.NoError(t, db.Create(role).Error)

	// Create bindings
	binding1 := &Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:alice@example.com"]`),
	}
	binding2 := &Binding{
		PolicyID: policy.ID,
		RoleID:   role.ID,
		Members:  []byte(`["user:bob@example.com"]`),
	}
	require.NoError(t, db.Create(binding1).Error)
	require.NoError(t, db.Create(binding2).Error)

	// Reload policy with bindings
	var loadedPolicy Policy
	err := db.Preload("Bindings").First(&loadedPolicy, policy.ID).Error
	require.NoError(t, err)

	assert.Len(t, loadedPolicy.Bindings, 2)
}

func TestDomain_ResourceHierarchy(t *testing.T) {
	db := setupTestDB(t)

	// Create parent-child hierarchy
	parent := &Resource{Type: "organization", Name: "parent"}
	require.NoError(t, db.Create(parent).Error)

	child := &Resource{Type: "project", Name: "child", ParentID: &parent.ID}
	require.NoError(t, db.Create(child).Error)

	// Reload child with parent
	var loadedChild Resource
	err := db.Preload("Parent").First(&loadedChild, child.ID).Error
	require.NoError(t, err)

	assert.NotNil(t, loadedChild.Parent)
	assert.Equal(t, parent.ID, loadedChild.Parent.ID)
	assert.Equal(t, "parent", loadedChild.Parent.Name)
}
