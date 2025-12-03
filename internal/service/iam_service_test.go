package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock PermissionEvaluator
type MockPermissionEvaluator struct {
	mock.Mock
}

func (m *MockPermissionEvaluator) CheckPermission(principal string, resourceID uuid.UUID, permission string, context map[string]string) (bool, string, error) {
	args := m.Called(principal, resourceID, permission, context)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockPermissionEvaluator) GetEffectivePermissions(principal string, resourceID uuid.UUID) ([]string, []string, error) {
	args := m.Called(principal, resourceID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]string), args.Get(1).([]string), args.Error(2)
}

// Test: Create Resource
func TestIAMService_CreateResource(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	// Mock expectations
	resourceRepo.On("Create", mock.AnythingOfType("*domain.Resource")).Return(nil).Run(func(args mock.Arguments) {
		res := args.Get(0).(*domain.Resource)
		res.ID = uuid.New() // Simulate DB assigning ID
	})

	// Create resource
	parentID := uuid.New()
	resource, err := service.CreateResource(
		"bucket",
		"test-bucket",
		&parentID,
		map[string]string{"region": "us-east-1"},
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resource)
	assert.Equal(t, "bucket", resource.Type)
	assert.Equal(t, "test-bucket", resource.Name)
	assert.Equal(t, parentID, *resource.ParentID)
	assert.Equal(t, "us-east-1", resource.Attributes["region"])

	resourceRepo.AssertExpectations(t)
}

// Test: Get Resource
func TestIAMService_GetResource(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	expectedResource := &domain.Resource{
		ID:   resourceID,
		Type: "project",
		Name: "my-project",
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(expectedResource, nil)

	// Get resource
	resource, err := service.GetResource(resourceID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResource, resource)

	resourceRepo.AssertExpectations(t)
}

// Test: Delete Resource
func TestIAMService_DeleteResource(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()

	// Mock expectations
	resourceRepo.On("Delete", resourceID).Return(nil)
	cache.Clear() // Clear cache after delete

	// Delete resource
	err := service.DeleteResource(resourceID)

	// Assert
	assert.NoError(t, err)

	resourceRepo.AssertExpectations(t)
}

// Test: Create Permission
func TestIAMService_CreatePermission(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	// Mock expectations
	permissionRepo.On("Create", mock.AnythingOfType("*domain.Permission")).Return(nil).Run(func(args mock.Arguments) {
		perm := args.Get(0).(*domain.Permission)
		perm.ID = uuid.New()
	})

	// Create permission
	permission, err := service.CreatePermission(
		"storage.buckets.read",
		"Read buckets",
		"storage",
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, permission)
	assert.Equal(t, "storage.buckets.read", permission.Name)
	assert.Equal(t, "Read buckets", permission.Description)
	assert.Equal(t, "storage", permission.Service)

	permissionRepo.AssertExpectations(t)
}

// Test: Create Role
func TestIAMService_CreateRole(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	permID1 := uuid.New()
	permID2 := uuid.New()
	permissionIDs := []uuid.UUID{permID1, permID2}

	permissions := []domain.Permission{
		{ID: permID1, Name: "storage.buckets.read"},
		{ID: permID2, Name: "storage.buckets.write"},
	}

	// Mock expectations
	permissionRepo.On("GetByIDs", permissionIDs).Return(permissions, nil)
	roleRepo.On("Create", mock.AnythingOfType("*domain.Role")).Return(nil).Run(func(args mock.Arguments) {
		role := args.Get(0).(*domain.Role)
		role.ID = uuid.New()
	})

	// Create role
	role, err := service.CreateRole(
		"roles/storage.editor",
		"Storage Editor",
		"Can read and write buckets",
		permissionIDs,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "roles/storage.editor", role.Name)
	assert.Equal(t, "Storage Editor", role.Title)
	assert.Equal(t, "Can read and write buckets", role.Description)
	assert.Len(t, role.Permissions, 2)

	permissionRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

// Test: Create Policy
func TestIAMService_CreatePolicy(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	roleID := uuid.New()

	bindings := []domain.Binding{
		{
			ID:      uuid.New(),
			RoleID:  roleID,
			Members: toJSON([]string{"user:alice@example.com"}),
		},
	}

	// Mock expectations
	createdPolicyID := uuid.New()
	policyRepo.On("Create", mock.AnythingOfType("*domain.Policy")).Return(nil).Run(func(args mock.Arguments) {
		policy := args.Get(0).(*domain.Policy)
		policy.ID = createdPolicyID
		policy.ETag = "etag-123"
	})

	// Binding creation
	bindingRepo.On("Create", mock.AnythingOfType("*domain.Binding")).Return(nil)

	// GetByID is called at the end - return the policy with bindings
	finalPolicy := &domain.Policy{
		ID:         createdPolicyID,
		ResourceID: resourceID,
		Bindings:   bindings,
		ETag:       "etag-123",
	}
	policyRepo.On("GetByID", createdPolicyID).Return(finalPolicy, nil)

	// Create policy
	policy, err := service.CreatePolicy(resourceID, bindings)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, resourceID, policy.ResourceID)
	assert.Len(t, policy.Bindings, 1)
	assert.Equal(t, "etag-123", policy.ETag)

	policyRepo.AssertExpectations(t)
}

// Test: Get Policy
func TestIAMService_GetPolicy(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	expectedPolicy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		ETag:       "etag-456",
	}

	// Mock expectations
	policyRepo.On("GetByResourceID", resourceID).Return(expectedPolicy, nil)

	// Get policy
	policy, err := service.GetPolicy(resourceID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicy, policy)

	policyRepo.AssertExpectations(t)
}

// Test: Delete Policy
func TestIAMService_DeletePolicy(t *testing.T) {
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

	etag := "etag-123"
	policy := &domain.Policy{
		ID:         policyID,
		ResourceID: resourceID,
		ETag:       etag,
	}

	// Mock expectations
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil)
	policyRepo.On("Delete", policyID).Return(nil)

	// Delete policy
	err := service.DeletePolicy(resourceID, etag)

	// Assert
	assert.NoError(t, err)

	policyRepo.AssertExpectations(t)
}

// Test: CheckPermission delegates to evaluator
func TestIAMService_CheckPermission(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()

	// Mock expectations
	evaluator.On("CheckPermission", "user:alice@example.com", resourceID, "storage.buckets.read", mock.Anything).
		Return(true, "Permission granted", nil)

	// Check permission
	allowed, reason, err := service.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.buckets.read",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, "Permission granted", reason)

	evaluator.AssertExpectations(t)
}

// Test: GetEffectivePermissions delegates to evaluator
func TestIAMService_GetEffectivePermissions(t *testing.T) {
	resourceRepo := new(MockResourceRepository)
	permissionRepo := new(MockPermissionRepository)
	roleRepo := new(MockRoleRepository)
	policyRepo := new(MockPolicyRepository)
	bindingRepo := new(MockBindingRepository)
	evaluator := new(MockPermissionEvaluator)
	cache := NewNoopCache()

	service := NewIAMService(resourceRepo, permissionRepo, roleRepo, policyRepo, bindingRepo, evaluator, cache)

	resourceID := uuid.New()
	expectedPerms := []string{"storage.buckets.read", "storage.buckets.write"}
	expectedRoles := []string{"roles/storage.editor"}

	// Mock expectations
	evaluator.On("GetEffectivePermissions", "user:alice@example.com", resourceID).
		Return(expectedPerms, expectedRoles, nil)

	// Get effective permissions
	perms, roles, err := service.GetEffectivePermissions(
		"user:alice@example.com",
		resourceID,
	)

	// Assert
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedPerms, perms)
	assert.ElementsMatch(t, expectedRoles, roles)

	evaluator.AssertExpectations(t)
}

// Mock RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(id uuid.UUID) (*domain.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleRepository) List(includeCustom bool, limit, offset int) ([]domain.Role, error) {
	args := m.Called(includeCustom, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockRoleRepository) AddPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	args := m.Called(roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) RemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	args := m.Called(roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) GetPermissions(roleID uuid.UUID) ([]domain.Permission, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

// Mock BindingRepository
type MockBindingRepository struct {
	mock.Mock
}

func (m *MockBindingRepository) Create(binding *domain.Binding) error {
	args := m.Called(binding)
	return args.Error(0)
}

func (m *MockBindingRepository) GetByID(id uuid.UUID) (*domain.Binding, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Binding), args.Error(1)
}

func (m *MockBindingRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBindingRepository) ListByResourceID(resourceID uuid.UUID, limit, offset int) ([]domain.Binding, error) {
	args := m.Called(resourceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Binding), args.Error(1)
}

func (m *MockBindingRepository) ListByPrincipal(principal string, limit, offset int) ([]domain.Binding, error) {
	args := m.Called(principal, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Binding), args.Error(1)
}

func (m *MockBindingRepository) GetByPolicyAndPrincipal(policyID uuid.UUID, principal string) ([]domain.Binding, error) {
	args := m.Called(policyID, principal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Binding), args.Error(1)
}
