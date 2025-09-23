package cache

import (
    "sync"
    "time"
)

// ValkeyCache is a lightweight in-memory stand-in for Valkey until integrated.
type ValkeyCache struct {
    data map[string]item
    mu   sync.RWMutex
}

type item struct {
    value      interface{}
    expiresAt  time.Time
}

// NewValkeyCache creates an in-memory cache simulating Valkey behaviour.
func NewValkeyCache() *ValkeyCache {
    return &ValkeyCache{data: make(map[string]item)}
}

// Get retrieves a cached item if present and not expired.
func (c *ValkeyCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    it, ok := c.data[key]
    if !ok {
        return nil, false
    }
    if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
        delete(c.data, key)
        return nil, false
    }
    return it.value, true
}

// Set stores a value with optional TTL.
func (c *ValkeyCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    var expires time.Time
    if ttl > 0 {
        expires = time.Now().Add(ttl)
    }
    c.data[key] = item{value: value, expiresAt: expires}
}

// Delete removes an entry.
func (c *ValkeyCache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.data, key)
}
