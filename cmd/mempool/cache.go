package main

import (
	"context"
	"sync"
	"time"

	"github.com/dipdup-io/workerpool"
)

// Cache -
type Cache struct {
	mux    sync.RWMutex
	lookup map[string]int64
	ticker *time.Ticker
	ttl    time.Duration

	g workerpool.Group
}

// NewCache -
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		lookup: make(map[string]int64),
		ttl:    ttl,
		ticker: time.NewTicker(time.Minute),
		g:      workerpool.NewGroup(),
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
	c.g.GoCtx(ctx, c.checkExpiration)
}

func (c *Cache) Close() error {
	c.g.Wait()
	c.ticker.Stop()
	return nil
}

func (c *Cache) checkExpiration(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-c.ticker.C:
			c.mux.Lock()
			for key, expiration := range c.lookup {
				if time.Now().UnixNano() <= expiration {
					continue
				}
				delete(c.lookup, key)
			}
			c.mux.Unlock()
		}
	}
}
