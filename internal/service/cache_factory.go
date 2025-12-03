package service

import (
	"fmt"
	"strings"

	"github.com/pguia/iam/internal/config"
)

// NewCache creates the appropriate cache implementation based on configuration
// Returns a stateless cache by default (type="none")
func NewCache(cfg *config.CacheConfig) (CacheService, error) {
	// If explicitly disabled, use no-op cache
	if !cfg.Enabled {
		return NewNoopCache(), nil
	}

	// Determine cache type
	cacheType := strings.ToLower(cfg.Type)

	switch cacheType {
	case "none", "":
		// Stateless - no caching
		return NewNoopCache(), nil

	case "memory":
		// In-memory cache (NOT stateless - use only for single instance)
		return NewCacheService(cfg), nil

	case "redis":
		// Redis distributed cache (stateless)
		cache, err := NewRedisCache(&cfg.Redis)
		if err != nil {
			return nil, fmt.Errorf("failed to create redis cache: %w", err)
		}
		return cache, nil

	default:
		return nil, fmt.Errorf("unknown cache type: %s (valid: none, memory, redis)", cfg.Type)
	}
}
