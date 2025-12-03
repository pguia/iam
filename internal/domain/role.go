package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a collection of permissions
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"` // e.g., "roles/storage.admin"
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	IsCustom    bool           `gorm:"default:false;not null" json:"is_custom"` // true for custom roles, false for predefined
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate hook to generate UUID if not set
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permissionName string) bool {
	for _, perm := range r.Permissions {
		if perm.Name == permissionName {
			return true
		}
	}
	return false
}
