package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"gorm.io/gorm"
)

// PermissionRepository handles permission data operations
type PermissionRepository interface {
	Create(permission *domain.Permission) error
	GetByID(id uuid.UUID) (*domain.Permission, error)
	GetByName(name string) (*domain.Permission, error)
	Delete(id uuid.UUID) error
	List(service string, limit, offset int) ([]domain.Permission, error)
	GetByIDs(ids []uuid.UUID) ([]domain.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(permission *domain.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) GetByID(id uuid.UUID) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.First(&permission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByName(name string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Permission{}, id).Error
}

func (r *permissionRepository) List(service string, limit, offset int) ([]domain.Permission, error) {
	var permissions []domain.Permission
	query := r.db.Model(&domain.Permission{})

	if service != "" {
		query = query.Where("service = ?", service)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByIDs(ids []uuid.UUID) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}
