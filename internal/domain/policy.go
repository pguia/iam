package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Policy represents an IAM policy attached to a resource
type Policy struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ResourceID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"resource_id"`
	Resource   *Resource      `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
	Bindings   []Binding      `gorm:"foreignKey:PolicyID" json:"bindings,omitempty"`
	ETag       string         `gorm:"type:varchar(64)" json:"etag"` // For optimistic concurrency control
	Version    int            `gorm:"default:1;not null" json:"version"`
	CreatedAt  time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return "policies"
}

// BeforeCreate hook to generate UUID and ETag if not set
func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.ETag == "" {
		p.ETag = uuid.New().String()
	}
	return nil
}

// BeforeUpdate hook to update ETag on changes
func (p *Policy) BeforeUpdate(tx *gorm.DB) error {
	p.ETag = uuid.New().String()
	p.Version++
	return nil
}
