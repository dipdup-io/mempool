package main

import (
	"context"
	"sync"
	"time"
)

// Cache -
type Cache struct {
	mux    sync.RWMutex
	lookup map[string]int64
	ticker *time.Ticker
	ttl    time.Duration

	wg sync.WaitGroup
}

// NewCache -
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		lookup: make(map[string]int64),
		ttl:    ttl,
		ticker: time.NewTicker(time.Minute),
	}
}

// Has -
func (c *Cache) Has(key string) bool {
	c.mux.RLock()
	_, ok := c.lookup[key]
	c.mux.RUnlock()
	return ok
}

// Set -
func (c *Cache) Set(key string) {
	expires := time.Now().Add(c.ttl).UnixNano()
	c.mux.Lock()
	c.lookup[key] = expires
	c.mux.Unlock()
}

// Start -
func (c *Cache) Start(ctx context.Context) {
	c.wg.Add(1)
	go c.checkExpiration(ctx)
}

func (c *Cache) checkExpiration(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			c.ticker.Stop()
			return
		case <-c.ticker.C:
			c.mux.RLock()
			for key, expiration := range c.lookup {
				if time.Now().UnixNano() <= expiration {
					continue
				}
				c.mux.Lock()
				delete(c.lookup, key)
				c.mux.Unlock()
			}
			c.mux.RUnlock()
		}
	}
}
