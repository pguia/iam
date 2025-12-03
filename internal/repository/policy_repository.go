package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"gorm.io/gorm"
)

// PolicyRepository handles policy data operations
type PolicyRepository interface {
	Create(policy *domain.Policy) error
	GetByID(id uuid.UUID) (*domain.Policy, error)
	GetByResourceID(resourceID uuid.UUID) (*domain.Policy, error)
	Update(policy *domain.Policy) error
	Delete(id uuid.UUID) error
	List(parentResourceID *uuid.UUID, limit, offset int) ([]domain.Policy, error)
}

type policyRepository struct {
	db *gorm.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *gorm.DB) PolicyRepository {
	return &policyRepository{db: db}
}

func (r *policyRepository) Create(policy *domain.Policy) error {
	return r.db.Create(policy).Error
}

func (r *policyRepository) GetByID(id uuid.UUID) (*domain.Policy, error) {
	var policy domain.Policy
	err := r.db.Preload("Resource").Preload("Bindings").Preload("Bindings.Role").
		Preload("Bindings.Role.Permissions").Preload("Bindings.Condition").
		First(&policy, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *policyRepository) GetByResourceID(resourceID uuid.UUID) (*domain.Policy, error) {
	var policy domain.Policy
	err := r.db.Preload("Resource").Preload("Bindings").Preload("Bindings.Role").
		Preload("Bindings.Role.Permissions").Preload("Bindings.Condition").
		Where("resource_id = ?", resourceID).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *policyRepository) Update(policy *domain.Policy) error {
	return r.db.Save(policy).Error
}

func (r *policyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Policy{}, id).Error
}

func (r *policyRepository) List(parentResourceID *uuid.UUID, limit, offset int) ([]domain.Policy, error) {
	var policies []domain.Policy
	query := r.db.Model(&domain.Policy{}).Preload("Resource").Preload("Bindings")

	if parentResourceID != nil {
		// Get all policies for resources under the parent
		query = query.Joins("JOIN resources ON resources.id = policies.resource_id").
			Where("resources.parent_id = ?", parentResourceID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&policies).Error
	return policies, err
}
