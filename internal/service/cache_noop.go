package service

// noopCache is a stateless cache implementation that doesn't cache anything
// Use this for stateless deployments with multiple replicas
type noopCache struct{}

// NewNoopCache creates a no-op cache that doesn't store anything
// This ensures the service is completely stateless
func NewNoopCache() CacheService {
	return &noopCache{}
}

func (c *noopCache) Get(key string) (interface{}, bool) {
	return nil, false
}

func (c *noopCache) Set(key string, value interface{}) {
	// No-op
}

func (c *noopCache) Delete(key string) {
	// No-op
}

func (c *noopCache) Clear() {
	// No-op
}
