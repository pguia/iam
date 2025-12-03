package service

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/config"
	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// Helper function to convert string slice to datatypes.JSON
// datatypes.JSON is just []byte, so we can directly marshal and cast
func toJSON(members []string) datatypes.JSON {
	data, _ := json.Marshal(members)
	return datatypes.JSON(data)
}

// Mock repositories for testing
type MockResourceRepository struct {
	mock.Mock
}

func (m *MockResourceRepository) Create(resource *domain.Resource) error {
	args := m.Called(resource)
	return args.Error(0)
}

func (m *MockResourceRepository) GetByID(id uuid.UUID) (*domain.Resource, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) Update(resource *domain.Resource) error {
	args := m.Called(resource)
	return args.Error(0)
}

func (m *MockResourceRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockResourceRepository) List(parentID *uuid.UUID, resourceType string, limit, offset int) ([]domain.Resource, error) {
	args := m.Called(parentID, resourceType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) GetAncestors(id uuid.UUID) ([]domain.Resource, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) GetChildren(parentID uuid.UUID) ([]domain.Resource, error) {
	args := m.Called(parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Resource), args.Error(1)
}

func (m *MockResourceRepository) GetDescendants(id uuid.UUID) ([]domain.Resource, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Resource), args.Error(1)
}

type MockPolicyRepository struct {
	mock.Mock
}

func (m *MockPolicyRepository) Create(policy *domain.Policy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockPolicyRepository) GetByID(id uuid.UUID) (*domain.Policy, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Policy), args.Error(1)
}

func (m *MockPolicyRepository) GetByResourceID(resourceID uuid.UUID) (*domain.Policy, error) {
	args := m.Called(resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Policy), args.Error(1)
}

func (m *MockPolicyRepository) Update(policy *domain.Policy) error {
	args := m.Called(policy)
	return args.Error(0)
}

func (m *MockPolicyRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPolicyRepository) List(parentResourceID *uuid.UUID, limit, offset int) ([]domain.Policy, error) {
	args := m.Called(parentResourceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Policy), args.Error(1)
}

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(permission *domain.Permission) error {
	args := m.Called(permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(id uuid.UUID) (*domain.Permission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByName(name string) (*domain.Permission, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) List(service string, limit, offset int) ([]domain.Permission, error) {
	args := m.Called(service, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByIDs(ids []uuid.UUID) ([]domain.Permission, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// Test: Permission check on direct resource
func TestCheckPermission_DirectResource(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	// Create test data
	resourceID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "test-bucket",
	}

	permission := &domain.Permission{
		ID:   permID,
		Name: "storage.objects.read",
	}

	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.viewer",
		Permissions: []domain.Permission{*permission},
	}

	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:alice@example.com"}),
	}

	policy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(resource, nil)
	resourceRepo.On("GetAncestors", resourceID).Return([]domain.Resource{}, nil)
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil)

	// Execute
	allowed, reason, err := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.read",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Contains(t, reason, "Permission granted")
	assert.Contains(t, reason, "roles/storage.viewer")

	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: Permission denied when user not in binding
func TestCheckPermission_UserNotInBinding(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	resourceID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "test-bucket",
	}

	permission := &domain.Permission{
		ID:   permID,
		Name: "storage.objects.read",
	}

	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.viewer",
		Permissions: []domain.Permission{*permission},
	}

	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:bob@example.com"}), // Alice not in members
	}

	policy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(resource, nil)
	resourceRepo.On("GetAncestors", resourceID).Return([]domain.Resource{}, nil)
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil)

	// Execute
	allowed, reason, err := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.read",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Contains(t, reason, "Permission denied")

	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: Hierarchical permission inheritance
func TestCheckPermission_HierarchicalInheritance(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	// Create hierarchy: org -> project -> bucket
	orgID := uuid.New()
	projectID := uuid.New()
	bucketID := uuid.New()

	org := &domain.Resource{
		ID:   orgID,
		Type: "organization",
		Name: "Acme Corp",
	}

	project := &domain.Resource{
		ID:       projectID,
		Type:     "project",
		Name:     "Web App",
		ParentID: &orgID,
	}

	bucket := &domain.Resource{
		ID:       bucketID,
		Type:     "bucket",
		Name:     "user-uploads",
		ParentID: &projectID,
	}

	roleID := uuid.New()
	permID := uuid.New()

	permission := &domain.Permission{
		ID:   permID,
		Name: "storage.objects.read",
	}

	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.admin",
		Permissions: []domain.Permission{*permission},
	}

	// Policy is on the ORG level
	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:alice@example.com"}),
	}

	orgPolicy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: orgID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations
	resourceRepo.On("GetByID", bucketID).Return(bucket, nil)
	resourceRepo.On("GetAncestors", bucketID).Return([]domain.Resource{*project, *org}, nil)

	// No policy on bucket
	policyRepo.On("GetByResourceID", bucketID).Return(nil, nil)

	// No policy on project
	policyRepo.On("GetByResourceID", projectID).Return(nil, nil)

	// Policy on org
	policyRepo.On("GetByResourceID", orgID).Return(orgPolicy, nil)

	// Execute - check permission on BUCKET, but policy is on ORG
	allowed, reason, err := evaluator.CheckPermission(
		"user:alice@example.com",
		bucketID,
		"storage.objects.read",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Contains(t, reason, "Permission granted")
	assert.Contains(t, reason, orgID.String())

	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: Permission denied when role lacks permission
func TestCheckPermission_RoleLacksPermission(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	resourceID := uuid.New()
	roleID := uuid.New()
	readPermID := uuid.New()

	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "test-bucket",
	}

	readPermission := &domain.Permission{
		ID:   readPermID,
		Name: "storage.objects.read",
	}

	// Role only has read permission
	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.viewer",
		Permissions: []domain.Permission{*readPermission},
	}

	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:alice@example.com"}),
	}

	policy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(resource, nil)
	resourceRepo.On("GetAncestors", resourceID).Return([]domain.Resource{}, nil)
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil)

	// Execute - try to delete (role doesn't have this permission)
	allowed, reason, err := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.delete",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Contains(t, reason, "Permission denied")

	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: Caching behavior
func TestCheckPermission_Caching(t *testing.T) {
	// Setup with memory cache
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewTestMemoryCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	resourceID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "test-bucket",
	}

	permission := &domain.Permission{
		ID:   permID,
		Name: "storage.objects.read",
	}

	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.viewer",
		Permissions: []domain.Permission{*permission},
	}

	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:alice@example.com"}),
	}

	policy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations - should only be called ONCE due to caching
	// Note: We'll call CheckPermission twice, but second call should use cache
	resourceRepo.On("GetByID", resourceID).Return(resource, nil).Once()
	resourceRepo.On("GetAncestors", resourceID).Return([]domain.Resource{}, nil).Once()
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil).Once()

	// First call - should hit DB
	allowed1, _, err1 := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.read",
		nil,
	)

	assert.NoError(t, err1)
	assert.True(t, allowed1)

	// Second call - should hit cache (mocks won't be called again)
	allowed2, reason2, err2 := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.read",
		nil,
	)

	assert.NoError(t, err2)
	assert.True(t, allowed2)
	assert.Contains(t, reason2, "cached")

	// Verify mocks were only called once
	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: GetEffectivePermissions
func TestGetEffectivePermissions(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	resourceID := uuid.New()
	roleID := uuid.New()

	resource := &domain.Resource{
		ID:   resourceID,
		Type: "bucket",
		Name: "test-bucket",
	}

	permissions := []domain.Permission{
		{ID: uuid.New(), Name: "storage.objects.read"},
		{ID: uuid.New(), Name: "storage.objects.write"},
		{ID: uuid.New(), Name: "storage.objects.delete"},
	}

	role := &domain.Role{
		ID:          roleID,
		Name:        "roles/storage.admin",
		Permissions: permissions,
	}

	binding := domain.Binding{
		ID:      uuid.New(),
		RoleID:  roleID,
		Role:    role,
		Members: toJSON([]string{"user:alice@example.com"}),
	}

	policy := &domain.Policy{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Bindings:   []domain.Binding{binding},
	}

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(resource, nil)
	resourceRepo.On("GetAncestors", resourceID).Return([]domain.Resource{}, nil)
	policyRepo.On("GetByResourceID", resourceID).Return(policy, nil)

	// Execute
	perms, roles, err := evaluator.GetEffectivePermissions(
		"user:alice@example.com",
		resourceID,
	)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, perms, 3)
	assert.Contains(t, perms, "storage.objects.read")
	assert.Contains(t, perms, "storage.objects.write")
	assert.Contains(t, perms, "storage.objects.delete")
	assert.Len(t, roles, 1)
	assert.Contains(t, roles, "roles/storage.admin")

	resourceRepo.AssertExpectations(t)
	policyRepo.AssertExpectations(t)
}

// Test: Resource not found
func TestCheckPermission_ResourceNotFound(t *testing.T) {
	// Setup
	resourceRepo := new(MockResourceRepository)
	policyRepo := new(MockPolicyRepository)
	permissionRepo := new(MockPermissionRepository)
	cache := NewNoopCache()

	evaluator := NewPermissionEvaluator(resourceRepo, policyRepo, permissionRepo, cache)

	resourceID := uuid.New()

	// Mock expectations
	resourceRepo.On("GetByID", resourceID).Return(nil, nil)

	// Execute
	allowed, reason, err := evaluator.CheckPermission(
		"user:alice@example.com",
		resourceID,
		"storage.objects.read",
		nil,
	)

	// Assert
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, "Resource not found", reason)

	resourceRepo.AssertExpectations(t)
}

// Helper to create memory cache for tests
func NewTestMemoryCache() CacheService {
	return NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})
}
