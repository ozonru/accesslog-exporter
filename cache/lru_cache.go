package cache

import (
	"github.com/hashicorp/golang-lru"
)

// LRUCache is the wrapper of lru cache implementation from hashicorp for storing parsed user agents
type LRUCache struct {
	*lru.Cache
}

// NewLRUCache creates new wrapper for storing parsed user agents
func NewLRUCache(size int) (*LRUCache, error) {
	cc, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	return &LRUCache{cc}, nil
}

// Set sets value under key to cache
func (c *LRUCache) Set(key string, value map[string]string) {
	c.Cache.Add(key, value)
}

// Get returns value by key if such key exists in cache
func (c *LRUCache) Get(key string) (map[string]string, bool) {
	value, ok := c.Cache.Get(key)
	if !ok {
		return nil, false
	}

	return value.(map[string]string), true
}

// Len returns cache length
func (c *LRUCache) Len() int {
	return c.Cache.Len()
}
