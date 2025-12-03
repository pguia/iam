package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"gorm.io/gorm"
)

// BindingRepository handles binding data operations
type BindingRepository interface {
	Create(binding *domain.Binding) error
	GetByID(id uuid.UUID) (*domain.Binding, error)
	Delete(id uuid.UUID) error
	ListByResourceID(resourceID uuid.UUID, limit, offset int) ([]domain.Binding, error)
	ListByPrincipal(principal string, limit, offset int) ([]domain.Binding, error)
	GetByPolicyAndPrincipal(policyID uuid.UUID, principal string) ([]domain.Binding, error)
}

type bindingRepository struct {
	db *gorm.DB
}

// NewBindingRepository creates a new binding repository
func NewBindingRepository(db *gorm.DB) BindingRepository {
	return &bindingRepository{db: db}
}

func (r *bindingRepository) Create(binding *domain.Binding) error {
	return r.db.Create(binding).Error
}

func (r *bindingRepository) GetByID(id uuid.UUID) (*domain.Binding, error) {
	var binding domain.Binding
	err := r.db.Preload("Role").Preload("Role.Permissions").Preload("Condition").
		First(&binding, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

func (r *bindingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Binding{}, id).Error
}

func (r *bindingRepository) ListByResourceID(resourceID uuid.UUID, limit, offset int) ([]domain.Binding, error) {
	var bindings []domain.Binding
	query := r.db.Model(&domain.Binding{}).
		Preload("Role").Preload("Role.Permissions").Preload("Condition").
		Joins("JOIN policies ON policies.id = bindings.policy_id").
		Where("policies.resource_id = ?", resourceID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&bindings).Error
	return bindings, err
}

func (r *bindingRepository) ListByPrincipal(principal string, limit, offset int) ([]domain.Binding, error) {
	var bindings []domain.Binding
	query := r.db.Model(&domain.Binding{}).
		Preload("Role").Preload("Role.Permissions").Preload("Condition").
		Where("members @> ?", `["`+principal+`"]`)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&bindings).Error
	return bindings, err
}

func (r *bindingRepository) GetByPolicyAndPrincipal(policyID uuid.UUID, principal string) ([]domain.Binding, error) {
	var bindings []domain.Binding
	err := r.db.Where("policy_id = ? AND members @> ?", policyID, `["`+principal+`"]`).
		Preload("Role").Preload("Role.Permissions").Preload("Condition").
		Find(&bindings).Error
	return bindings, err
}
