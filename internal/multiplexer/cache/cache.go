package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Cache is an inetrafce for caching responses. after getting context.Done() cache is no longer available
type Cache interface {
	Set(key string, value string) error
	Get(key string) (string, error)
}

type cache struct {
	data map[string]cachedResponse
	// todo: consider sync.Map
	mu *sync.RWMutex
	// cacheTTL is ttl of cached items in seconds
	cacheTTL int64
	// available is a flag to indicate if cache is available
	available *atomic.Bool
}

type cachedResponse struct {
	timestamp int64
	body      string
}

const (
	cacheTTLDefault = 5
)

func NewCache(ctx context.Context, cacheTTL int64) Cache {
	if cacheTTL == 0 {
		cacheTTL = cacheTTLDefault
	}

	available := &atomic.Bool{}

	available.Store(true)

	c := &cache{
		data:      make(map[string]cachedResponse),
		mu:        &sync.RWMutex{},
		cacheTTL:  cacheTTL,
		available: available,
	}

	go func() {
		t := time.NewTicker(time.Duration(cacheTTL) * time.Second)

		for {
			select {
			case <-ctx.Done():
				c.available.Store(false)
				c.data = nil
				return

			case <-t.C:
				c.mu.Lock()
				for k, v := range c.data {
					if time.Now().Unix()-v.timestamp > c.cacheTTL {
						delete(c.data, k)
					}
				}
				c.mu.Unlock()
			}
		}
	}()

	return c
}

func (c *cache) Set(key string, value string) error {
	if !c.available.Load() {
		return ErrorCacheUnavailable
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cachedResponse{
		timestamp: time.Now().Unix(),
		body:      value,
	}

	return nil
}

func (c *cache) Get(key string) (string, error) {
	if !c.available.Load() {
		return "", ErrorCacheUnavailable
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, ok := c.data[key]
	if !ok {
		return "", ErrorNotFoundCacheItem
	}

	if time.Now().Unix()-resp.timestamp > c.cacheTTL {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.data, key)

		return "", ErrorNotFoundCacheItem
	}

	return resp.body, nil
}
