package service

import (
	"testing"

	"github.com/pguia/iam/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test that our toJSON helper works correctly
func TestToJSON_Helper(t *testing.T) {
	members := []string{"user:alice@example.com", "user:bob@example.com"}
	json := toJSON(members)

	binding := domain.Binding{
		Members: json,
	}

	// Test GetMembers now that it uses json.Unmarshal
	retrieved, err := binding.GetMembers()
	assert.NoError(t, err)
	assert.ElementsMatch(t, members, retrieved)

	// Test HasMember
	assert.True(t, binding.HasMember("user:alice@example.com"))
	assert.True(t, binding.HasMember("user:bob@example.com"))
	assert.False(t, binding.HasMember("user:charlie@example.com"))
}
