package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents a specific action that can be performed
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"` // e.g., "storage.buckets.create"
	Description string         `gorm:"type:text" json:"description"`
	Service     string         `gorm:"type:varchar(100);index" json:"service"` // e.g., "storage", "compute"
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate hook to generate UUID if not set
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
