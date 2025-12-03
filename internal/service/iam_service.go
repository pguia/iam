package service

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"github.com/pguia/iam/internal/repository"
	"gorm.io/datatypes"
)

// IAMService provides IAM functionality
type IAMService struct {
	resourceRepo   repository.ResourceRepository
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	policyRepo     repository.PolicyRepository
	bindingRepo    repository.BindingRepository
	evaluator      PermissionEvaluator
	cache          CacheService
}

// NewIAMService creates a new IAM service
func NewIAMService(
	resourceRepo repository.ResourceRepository,
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	policyRepo repository.PolicyRepository,
	bindingRepo repository.BindingRepository,
	evaluator PermissionEvaluator,
	cache CacheService,
) *IAMService {
	return &IAMService{
		resourceRepo:   resourceRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		policyRepo:     policyRepo,
		bindingRepo:    bindingRepo,
		evaluator:      evaluator,
		cache:          cache,
	}
}

// =============== Permission Checking ===============

// CheckPermission checks if a principal has a permission on a resource
func (s *IAMService) CheckPermission(
	principal string,
	resourceID uuid.UUID,
	permission string,
	context map[string]string,
) (bool, string, error) {
	return s.evaluator.CheckPermission(principal, resourceID, permission, context)
}

// GetEffectivePermissions gets all effective permissions for a principal on a resource
func (s *IAMService) GetEffectivePermissions(
	principal string,
	resourceID uuid.UUID,
) ([]string, []string, error) {
	return s.evaluator.GetEffectivePermissions(principal, resourceID)
}

// =============== Resource Management ===============

// CreateResource creates a new resource
func (s *IAMService) CreateResource(
	resourceType, name string,
	parentID *uuid.UUID,
	attributes map[string]string,
) (*domain.Resource, error) {
	resource := &domain.Resource{
		Type:       resourceType,
		Name:       name,
		ParentID:   parentID,
		Attributes: attributes,
	}

	if err := s.resourceRepo.Create(resource); err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return resource, nil
}

// GetResource gets a resource by ID
func (s *IAMService) GetResource(id uuid.UUID) (*domain.Resource, error) {
	return s.resourceRepo.GetByID(id)
}

// UpdateResource updates a resource
func (s *IAMService) UpdateResource(
	id uuid.UUID,
	name string,
	attributes map[string]string,
) (*domain.Resource, error) {
	resource, err := s.resourceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, fmt.Errorf("resource not found")
	}

	resource.Name = name
	resource.Attributes = attributes

	if err := s.resourceRepo.Update(resource); err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	return resource, nil
}

// DeleteResource deletes a resource
func (s *IAMService) DeleteResource(id uuid.UUID) error {
	return s.resourceRepo.Delete(id)
}

// ListResources lists resources
func (s *IAMService) ListResources(
	parentID *uuid.UUID,
	resourceType string,
	pageSize, offset int,
) ([]domain.Resource, error) {
	return s.resourceRepo.List(parentID, resourceType, pageSize, offset)
}

// GetResourceHierarchy gets ancestors and descendants of a resource
func (s *IAMService) GetResourceHierarchy(id uuid.UUID) ([]domain.Resource, []domain.Resource, error) {
	ancestors, err := s.resourceRepo.GetAncestors(id)
	if err != nil {
		return nil, nil, err
	}

	descendants, err := s.resourceRepo.GetDescendants(id)
	if err != nil {
		return nil, nil, err
	}

	return ancestors, descendants, nil
}

// =============== Permission Management ===============

// CreatePermission creates a new permission
func (s *IAMService) CreatePermission(
	name, description, service string,
) (*domain.Permission, error) {
	permission := &domain.Permission{
		Name:        name,
		Description: description,
		Service:     service,
	}

	if err := s.permissionRepo.Create(permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetPermission gets a permission by ID
func (s *IAMService) GetPermission(id uuid.UUID) (*domain.Permission, error) {
	return s.permissionRepo.GetByID(id)
}

// ListPermissions lists permissions
func (s *IAMService) ListPermissions(service string, pageSize, offset int) ([]domain.Permission, error) {
	return s.permissionRepo.List(service, pageSize, offset)
}

// =============== Role Management ===============

// CreateRole creates a new role
func (s *IAMService) CreateRole(
	name, title, description string,
	permissionIDs []uuid.UUID,
) (*domain.Role, error) {
	// Get permissions
	permissions, err := s.permissionRepo.GetByIDs(permissionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	role := &domain.Role{
		Name:        name,
		Title:       title,
		Description: description,
		Permissions: permissions,
		IsCustom:    true,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole gets a role by ID
func (s *IAMService) GetRole(id uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetByID(id)
}

// UpdateRole updates a role
func (s *IAMService) UpdateRole(
	id uuid.UUID,
	title, description string,
	permissionIDs []uuid.UUID,
) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Get new permissions
	permissions, err := s.permissionRepo.GetByIDs(permissionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	role.Title = title
	role.Description = description
	role.Permissions = permissions

	if err := s.roleRepo.Update(role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *IAMService) DeleteRole(id uuid.UUID) error {
	return s.roleRepo.Delete(id)
}

// ListRoles lists roles
func (s *IAMService) ListRoles(includePredefined bool, pageSize, offset int) ([]domain.Role, error) {
	return s.roleRepo.List(includePredefined, pageSize, offset)
}

// =============== Policy Management ===============

// CreatePolicy creates a new policy for a resource
func (s *IAMService) CreatePolicy(resourceID uuid.UUID, bindings []domain.Binding) (*domain.Policy, error) {
	policy := &domain.Policy{
		ResourceID: resourceID,
		Version:    1,
	}

	if err := s.policyRepo.Create(policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// Create bindings
	for i := range bindings {
		bindings[i].PolicyID = policy.ID
		if err := s.bindingRepo.Create(&bindings[i]); err != nil {
			return nil, fmt.Errorf("failed to create binding: %w", err)
		}
	}

	// Clear cache for this resource
	s.cache.Clear()

	return s.policyRepo.GetByID(policy.ID)
}

// GetPolicy gets a policy for a resource
func (s *IAMService) GetPolicy(resourceID uuid.UUID) (*domain.Policy, error) {
	return s.policyRepo.GetByResourceID(resourceID)
}

// UpdatePolicy updates a policy
func (s *IAMService) UpdatePolicy(
	resourceID uuid.UUID,
	bindings []domain.Binding,
	etag string,
) (*domain.Policy, error) {
	policy, err := s.policyRepo.GetByResourceID(resourceID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("policy not found")
	}

	// Check etag for optimistic concurrency control
	if policy.ETag != etag {
		return nil, fmt.Errorf("policy has been modified, etag mismatch")
	}

	// Delete existing bindings
	for _, binding := range policy.Bindings {
		if err := s.bindingRepo.Delete(binding.ID); err != nil {
			return nil, fmt.Errorf("failed to delete binding: %w", err)
		}
	}

	// Create new bindings
	for i := range bindings {
		bindings[i].PolicyID = policy.ID
		if err := s.bindingRepo.Create(&bindings[i]); err != nil {
			return nil, fmt.Errorf("failed to create binding: %w", err)
		}
	}

	// Update policy (will increment version and generate new etag)
	if err := s.policyRepo.Update(policy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// Clear cache
	s.cache.Clear()

	return s.policyRepo.GetByID(policy.ID)
}

// DeletePolicy deletes a policy
func (s *IAMService) DeletePolicy(resourceID uuid.UUID, etag string) error {
	policy, err := s.policyRepo.GetByResourceID(resourceID)
	if err != nil {
		return err
	}
	if policy == nil {
		return fmt.Errorf("policy not found")
	}

	if policy.ETag != etag {
		return fmt.Errorf("policy has been modified, etag mismatch")
	}

	// Clear cache
	s.cache.Clear()

	return s.policyRepo.Delete(policy.ID)
}

// ListPolicies lists policies
func (s *IAMService) ListPolicies(
	parentResourceID *uuid.UUID,
	pageSize, offset int,
) ([]domain.Policy, error) {
	return s.policyRepo.List(parentResourceID, pageSize, offset)
}

// =============== Binding Management ===============

// CreateBinding creates a new binding
func (s *IAMService) CreateBinding(
	resourceID, roleID uuid.UUID,
	members []string,
	condition *domain.Condition,
) (*domain.Binding, error) {
	// Get or create policy for this resource
	policy, err := s.policyRepo.GetByResourceID(resourceID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		// Create policy
		policy = &domain.Policy{
			ResourceID: resourceID,
			Version:    1,
		}
		if err := s.policyRepo.Create(policy); err != nil {
			return nil, fmt.Errorf("failed to create policy: %w", err)
		}
	}

	// Convert members to JSON
	membersJSON, err := json.Marshal(members)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal members: %w", err)
	}

	binding := &domain.Binding{
		PolicyID: policy.ID,
		RoleID:   roleID,
		Members:  datatypes.JSON(membersJSON),
	}

	if err := s.bindingRepo.Create(binding); err != nil {
		return nil, fmt.Errorf("failed to create binding: %w", err)
	}

	// Create condition if provided
	if condition != nil {
		condition.BindingID = binding.ID
		// Note: You'd need a condition repository to save this
	}

	// Clear cache
	s.cache.Clear()

	return s.bindingRepo.GetByID(binding.ID)
}

// DeleteBinding deletes a binding
func (s *IAMService) DeleteBinding(id uuid.UUID) error {
	// Clear cache
	s.cache.Clear()

	return s.bindingRepo.Delete(id)
}

// ListBindings lists bindings for a resource
func (s *IAMService) ListBindings(
	resourceID uuid.UUID,
	principal string,
	pageSize, offset int,
) ([]domain.Binding, error) {
	if principal != "" {
		return s.bindingRepo.ListByPrincipal(principal, pageSize, offset)
	}
	return s.bindingRepo.ListByResourceID(resourceID, pageSize, offset)
}
