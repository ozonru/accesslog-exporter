package cache

// Cache is an interface for cache that stores parsed user agents
type Cache interface {
	Set(key string, value map[string]string)
	Get(key string) (map[string]string, bool)
	Len() int
}
