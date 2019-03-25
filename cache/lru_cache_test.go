package cache

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestLRUCache(t *testing.T) { TestingT(t) }

type LRUCacheSuite struct{}

var _ = Suite(&LRUCacheSuite{})

func (s LRUCacheSuite) TestGet(c *C) {
	cache, err := NewLRUCache(1)
	c.Assert(err, IsNil)
	c.Assert(cache, NotNil)

	values := make(map[string]string)
	values["foo"] = "bar"
	values["zoo"] = "tar"

	cache.Set("test_key1", values)
	res, exists := cache.Get("test_key1")
	c.Assert(exists, Equals, true)
	c.Assert(res, DeepEquals, values)

	cache.Set("test_key2", values)
	res, exists = cache.Get("test_key1")
	c.Assert(exists, Equals, false)
	c.Assert(res, IsNil)

	res, exists = cache.Get("nonexistent")
	c.Assert(exists, Equals, false)
	c.Assert(res, IsNil)
}

func (s LRUCacheSuite) TestLength(c *C) {
	cache, err := NewLRUCache(2)
	c.Assert(err, IsNil)
	c.Assert(cache, NotNil)

	cache.Set("key1", map[string]string{})
	c.Assert(cache.Len(), Equals, 1)
	cache.Set("key2", map[string]string{})
	c.Assert(cache.Len(), Equals, 2)
	cache.Set("key3", map[string]string{})
	c.Assert(cache.Len(), Equals, 2)
}
