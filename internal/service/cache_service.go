package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/pguia/iam/internal/config"
)

// CacheService provides in-memory caching for permission checks
type CacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

type cacheService struct {
	cfg     *config.CacheConfig
	data    map[string]cacheEntry
	mu      sync.RWMutex
	enabled bool
	ttl     time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(cfg *config.CacheConfig) CacheService {
	cs := &cacheService{
		cfg:     cfg,
		data:    make(map[string]cacheEntry),
		enabled: cfg.Enabled,
		ttl:     time.Duration(cfg.TTLSeconds) * time.Second,
	}

	// Start cleanup goroutine
	if cs.enabled {
		go cs.cleanup()
	}

	return cs
}

func (c *cacheService) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (c *cacheService) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check max size
	if len(c.data) >= c.cfg.MaxSize {
		// Simple eviction: remove expired entries
		c.evictExpired()

		// If still at max, clear oldest entries (simplified)
		if len(c.data) >= c.cfg.MaxSize {
			// In production, use LRU eviction
			c.data = make(map[string]cacheEntry)
		}
	}

	c.data[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

func (c *cacheService) Delete(key string) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *cacheService) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}

func (c *cacheService) cleanup() {
	ticker := time.NewTicker(time.Duration(c.cfg.CleanupMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		c.evictExpired()
		c.mu.Unlock()
	}
}

func (c *cacheService) evictExpired() {
	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiration) {
			delete(c.data, key)
		}
	}
}

// GenerateCacheKey generates a cache key for permission checks
func GenerateCacheKey(principal, resourceID, permission string) string {
	return fmt.Sprintf("perm:%s:%s:%s", principal, resourceID, permission)
}
