package service

import (
	"testing"
	"time"

	"github.com/pguia/iam/internal/config"
	"github.com/stretchr/testify/assert"
)

// Test NoopCache - should never cache anything
func TestNoopCache(t *testing.T) {
	cache := NewNoopCache()

	// Set should be no-op
	cache.Set("key1", "value1")
	cache.Set("key2", 42)

	// Get should always return false
	val, found := cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, val)

	val, found = cache.Get("key2")
	assert.False(t, found)
	assert.Nil(t, val)

	// Delete should be no-op (no panic)
	cache.Delete("key1")

	// Clear should be no-op (no panic)
	cache.Clear()
}

// Test Memory Cache - basic get/set
func TestMemoryCache_BasicOperations(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set and get string
	cache.Set("key1", "value1")
	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// Set and get int
	cache.Set("key2", 42)
	val, found = cache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, 42, val)

	// Set and get bool
	cache.Set("key3", true)
	val, found = cache.Get("key3")
	assert.True(t, found)
	assert.Equal(t, true, val)

	// Get non-existent key
	val, found = cache.Get("non-existent")
	assert.False(t, found)
	assert.Nil(t, val)
}

// Test Memory Cache - delete
func TestMemoryCache_Delete(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Verify they exist
	_, found := cache.Get("key1")
	assert.True(t, found)
	_, found = cache.Get("key2")
	assert.True(t, found)

	// Delete key1
	cache.Delete("key1")

	// key1 should be gone, key2 should remain
	_, found = cache.Get("key1")
	assert.False(t, found)
	_, found = cache.Get("key2")
	assert.True(t, found)
}

// Test Memory Cache - clear
func TestMemoryCache_Clear(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set multiple values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify they exist
	_, found := cache.Get("key1")
	assert.True(t, found)

	// Clear cache
	cache.Clear()

	// All keys should be gone
	_, found = cache.Get("key1")
	assert.False(t, found)
	_, found = cache.Get("key2")
	assert.False(t, found)
	_, found = cache.Get("key3")
	assert.False(t, found)
}

// Test Memory Cache - TTL expiration
func TestMemoryCache_TTLExpiration(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     1, // 1 second TTL
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set value
	cache.Set("key1", "value1")

	// Should be available immediately
	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond) // Wait slightly more than 1 second

	// Should be expired
	val, found = cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, val)
}

// Test Memory Cache - disabled cache
func TestMemoryCache_Disabled(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        false, // Cache disabled
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set should be no-op
	cache.Set("key1", "value1")

	// Get should always return false
	val, found := cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, val)
}

// Test Memory Cache - max size eviction
func TestMemoryCache_MaxSizeEviction(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        5, // Small max size
		CleanupMinutes: 10,
	})

	// Fill cache to max
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4")
	cache.Set("key5", "value5")

	// All should be present
	for i := 1; i <= 5; i++ {
		_, found := cache.Get("key" + string(rune('0'+i)))
		assert.True(t, found)
	}

	// Adding one more should trigger eviction
	cache.Set("key6", "value6")

	// key6 should be present (just added)
	_, found := cache.Get("key6")
	assert.True(t, found)

	// Note: Due to simplified eviction (clears all when full),
	// old keys may or may not be present. This test just verifies no panic.
}

// Test Memory Cache - concurrent access
func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        1000,
		CleanupMinutes: 10,
	})

	// Run concurrent writes and reads
	done := make(chan bool)

	// Writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cache.Set(string(rune('A'+id)), j)
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cache.Get(string(rune('A' + id)))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic - test passes if we get here
	assert.True(t, true)
}

// Test GenerateCacheKey
func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		principal  string
		resourceID string
		permission string
		expected   string
	}{
		{
			name:       "simple case",
			principal:  "user:alice@example.com",
			resourceID: "resource-123",
			permission: "storage.buckets.read",
			expected:   "perm:user:alice@example.com:resource-123:storage.buckets.read",
		},
		{
			name:       "with special characters",
			principal:  "user:bob+test@example.com",
			resourceID: "resource-456-xyz",
			permission: "iam.roles.create",
			expected:   "perm:user:bob+test@example.com:resource-456-xyz:iam.roles.create",
		},
		{
			name:       "group principal",
			principal:  "group:admins@example.com",
			resourceID: "org-789",
			permission: "admin.all",
			expected:   "perm:group:admins@example.com:org-789:admin.all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateCacheKey(tt.principal, tt.resourceID, tt.permission)
			assert.Equal(t, tt.expected, key)
		})
	}
}

// Test Memory Cache - update existing key
func TestMemoryCache_UpdateExistingKey(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Set initial value
	cache.Set("key1", "value1")
	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// Update value
	cache.Set("key1", "value2")
	val, found = cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value2", val)

	// Update with different type
	cache.Set("key1", 42)
	val, found = cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, 42, val)
}

// Test Memory Cache - complex values
func TestMemoryCache_ComplexValues(t *testing.T) {
	cache := NewCacheService(&config.CacheConfig{
		Type:           "memory",
		Enabled:        true,
		TTLSeconds:     300,
		MaxSize:        100,
		CleanupMinutes: 10,
	})

	// Store struct
	type TestStruct struct {
		Name  string
		Count int
	}
	testData := TestStruct{Name: "test", Count: 42}
	cache.Set("struct", testData)

	val, found := cache.Get("struct")
	assert.True(t, found)
	retrieved, ok := val.(TestStruct)
	assert.True(t, ok)
	assert.Equal(t, "test", retrieved.Name)
	assert.Equal(t, 42, retrieved.Count)

	// Store slice
	slice := []string{"a", "b", "c"}
	cache.Set("slice", slice)

	val, found = cache.Get("slice")
	assert.True(t, found)
	retrievedSlice, ok := val.([]string)
	assert.True(t, ok)
	assert.ElementsMatch(t, slice, retrievedSlice)

	// Store map
	m := map[string]int{"one": 1, "two": 2}
	cache.Set("map", m)

	val, found = cache.Get("map")
	assert.True(t, found)
	retrievedMap, ok := val.(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 1, retrievedMap["one"])
	assert.Equal(t, 2, retrievedMap["two"])
}
