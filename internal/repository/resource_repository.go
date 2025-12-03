package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"gorm.io/gorm"
)

// ResourceRepository handles resource data operations
type ResourceRepository interface {
	Create(resource *domain.Resource) error
	GetByID(id uuid.UUID) (*domain.Resource, error)
	Update(resource *domain.Resource) error
	Delete(id uuid.UUID) error
	List(parentID *uuid.UUID, resourceType string, limit, offset int) ([]domain.Resource, error)
	GetChildren(id uuid.UUID) ([]domain.Resource, error)
	GetAncestors(id uuid.UUID) ([]domain.Resource, error)
	GetDescendants(id uuid.UUID) ([]domain.Resource, error)
}

type resourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository creates a new resource repository
func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(resource *domain.Resource) error {
	return r.db.Create(resource).Error
}

func (r *resourceRepository) GetByID(id uuid.UUID) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.Preload("Parent").First(&resource, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resource, nil
}

func (r *resourceRepository) Update(resource *domain.Resource) error {
	return r.db.Save(resource).Error
}

func (r *resourceRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Resource{}, id).Error
}

func (r *resourceRepository) List(parentID *uuid.UUID, resourceType string, limit, offset int) ([]domain.Resource, error) {
	var resources []domain.Resource
	query := r.db.Model(&domain.Resource{})

	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	}

	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&resources).Error
	return resources, err
}

func (r *resourceRepository) GetChildren(id uuid.UUID) ([]domain.Resource, error) {
	var children []domain.Resource
	err := r.db.Where("parent_id = ?", id).Find(&children).Error
	return children, err
}

func (r *resourceRepository) GetAncestors(id uuid.UUID) ([]domain.Resource, error) {
	var ancestors []domain.Resource

	// Use recursive CTE to get all ancestors
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, type, name, parent_id, attributes, created_at, updated_at, deleted_at
			FROM resources
			WHERE id = ?
			UNION ALL
			SELECT r.id, r.type, r.name, r.parent_id, r.attributes, r.created_at, r.updated_at, r.deleted_at
			FROM resources r
			INNER JOIN ancestors a ON r.id = a.parent_id
			WHERE r.deleted_at IS NULL
		)
		SELECT * FROM ancestors WHERE id != ?
	`

	err := r.db.Raw(query, id, id).Scan(&ancestors).Error
	return ancestors, err
}

func (r *resourceRepository) GetDescendants(id uuid.UUID) ([]domain.Resource, error) {
	var descendants []domain.Resource

	// Use recursive CTE to get all descendants
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id, type, name, parent_id, attributes, created_at, updated_at, deleted_at
			FROM resources
			WHERE id = ?
			UNION ALL
			SELECT r.id, r.type, r.name, r.parent_id, r.attributes, r.created_at, r.updated_at, r.deleted_at
			FROM resources r
			INNER JOIN descendants d ON r.parent_id = d.id
			WHERE r.deleted_at IS NULL
		)
		SELECT * FROM descendants WHERE id != ?
	`

	err := r.db.Raw(query, id, id).Scan(&descendants).Error
	return descendants, err
}
