package main

import (
	"sync"
	"time"
)

// Cache -
type Cache struct {
	mux    sync.RWMutex
	lookup map[string]int64
	ttl    time.Duration
}

// NewCache -
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		lookup: make(map[string]int64),
		ttl:    ttl,
	}
}

// Has -
func (c *Cache) Has(key string) bool {
	c.mux.RLock()
	expires, ok := c.lookup[key]
	c.mux.RUnlock()

	if !ok {
		return false
	}

	if time.Now().UnixNano() > expires {
		c.mux.Lock()
		delete(c.lookup, key)
		c.mux.Unlock()
		return false
	}

	return true
}

// Set -
func (c *Cache) Set(key string) {
	expires := time.Now().Add(c.ttl).UnixNano()
	c.mux.Lock()
	c.lookup[key] = expires
	c.mux.Unlock()
}
