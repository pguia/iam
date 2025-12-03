package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Condition represents a conditional expression for attribute-based access control
type Condition struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BindingID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"binding_id"`
	Title       string         `gorm:"type:varchar(255)" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Expression  string         `gorm:"type:text;not null" json:"expression"` // CEL expression
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Condition
func (Condition) TableName() string {
	return "conditions"
}

// BeforeCreate hook to generate UUID if not set
func (c *Condition) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
