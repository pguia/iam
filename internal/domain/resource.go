package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Resource represents a resource in the system (hierarchical)
type Resource struct {
	ID         uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type       string            `gorm:"type:varchar(100);not null;index" json:"type"` // e.g., "project", "organization", "bucket"
	Name       string            `gorm:"type:varchar(255);not null" json:"name"`
	ParentID   *uuid.UUID        `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent     *Resource         `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children   []Resource        `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Attributes map[string]string `gorm:"type:jsonb" json:"attributes"`
	Policies   []Policy          `gorm:"foreignKey:ResourceID" json:"policies,omitempty"`
	CreatedAt  time.Time         `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"not null" json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Resource
func (Resource) TableName() string {
	return "resources"
}

// BeforeCreate hook to generate UUID if not set
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// GetAncestors returns all ancestors of the resource (parent, grandparent, etc.)
func (r *Resource) GetAncestors(db *gorm.DB) ([]Resource, error) {
	var ancestors []Resource
	current := r

	for current.ParentID != nil {
		var parent Resource
		if err := db.First(&parent, current.ParentID).Error; err != nil {
			return ancestors, err
		}
		ancestors = append(ancestors, parent)
		current = &parent
	}

	return ancestors, nil
}
