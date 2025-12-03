package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/domain"
	"gorm.io/gorm"
)

// RoleRepository handles role data operations
type RoleRepository interface {
	Create(role *domain.Role) error
	GetByID(id uuid.UUID) (*domain.Role, error)
	GetByName(name string) (*domain.Role, error)
	Update(role *domain.Role) error
	Delete(id uuid.UUID) error
	List(includeCustom bool, limit, offset int) ([]domain.Role, error)
	AddPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	GetPermissions(roleID uuid.UUID) ([]domain.Permission, error)
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Role{}, id).Error
}

func (r *roleRepository) List(includeCustom bool, limit, offset int) ([]domain.Role, error) {
	var roles []domain.Role
	query := r.db.Model(&domain.Role{}).Preload("Permissions")

	if !includeCustom {
		query = query.Where("is_custom = ?", false)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AddPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	var role domain.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var permissions []domain.Permission
	if err := r.db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}

	return r.db.Model(&role).Association("Permissions").Append(&permissions)
}

func (r *roleRepository) RemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	var role domain.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var permissions []domain.Permission
	if err := r.db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return err
	}

	return r.db.Model(&role).Association("Permissions").Delete(&permissions)
}

func (r *roleRepository) GetPermissions(roleID uuid.UUID) ([]domain.Permission, error) {
	var role domain.Role
	err := r.db.Preload("Permissions").First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}
