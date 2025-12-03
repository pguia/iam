package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Binding represents a binding between members and a role on a policy
type Binding struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PolicyID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"policy_id"`
	Policy    *Policy        `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
	RoleID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"role_id"`
	Role      *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Members   datatypes.JSON `gorm:"type:jsonb;not null" json:"members"` // Array of strings: ["user:alice@example.com", "group:admins"]
	Condition *Condition     `gorm:"foreignKey:BindingID" json:"condition,omitempty"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Binding
func (Binding) TableName() string {
	return "bindings"
}

// BeforeCreate hook to generate UUID if not set
func (b *Binding) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// GetMembers unmarshals the Members JSON to a string slice
func (b *Binding) GetMembers() ([]string, error) {
	var members []string
	// datatypes.JSON is just []byte, so we can unmarshal directly
	if err := json.Unmarshal(b.Members, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// HasMember checks if a principal is in the members list
func (b *Binding) HasMember(principal string) bool {
	members, err := b.GetMembers()
	if err != nil {
		return false
	}
	for _, member := range members {
		if member == principal {
			return true
		}
	}
	return false
}
