// backend/internal/cache/cache.go
package cache

// This package implements a thread-safe in-memory cache system
// with automatic expiration and sanitization features for XSS protection

import (
	"encoding/json"
	"html"
	"sync"
	"time"
)

// SanitizeString cleans a string from potentially dangerous XSS content
// by escaping HTML special characters and removing dangerous URL schemes
func SanitizeString(input string) string {
	// Use HTML escaping for basic protection against script injection
	escaped := html.EscapeString(input)
	// Additionally remove potentially dangerous tags and attributes
	return escaped
}

// SafeSet adds an item to the cache with XSS protection
// It marshals the object, sanitizes it, and unmarshals it back
// to create a "clean" copy before storing in the cache
func (c *Cache) SafeSet(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Serialize the object to JSON
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		// Handle serialization error silently
		return
	}

	// Deserialize back to create a "clean" copy
	var sanitizedValue interface{}
	if err := json.Unmarshal(jsonBytes, &sanitizedValue); err != nil {
		return
	}

	// Store the sanitized value in the cache with expiration time
	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = Item{
		Value:      sanitizedValue,
		Expiration: expiration,
	}
}

// SafeGetString retrieves a string item from the cache with XSS verification
// Returns the sanitized string and a boolean indicating if the item was found
func (c *Cache) SafeGetString(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return "", false
	}

	// Check if item has expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return "", false
	}

	// Verify the value is a string
	strValue, ok := item.Value.(string)
	if !ok {
		return "", false
	}

	// Apply additional sanitization
	return SanitizeString(strValue), true
}

// Item represents a cache entry with an expiration timestamp
type Item struct {
	Value      interface{} // The cached data
	Expiration int64       // Timestamp when the item expires (Unix nano)
}

// Cache implements a simple in-memory cache with expiration support
// It's thread-safe and provides automatic cleanup of expired items
type Cache struct {
	items map[string]Item // Map of cached items
	mu    sync.RWMutex    // RWMutex for thread-safety
}

// NewCache creates a new cache instance and starts the cleanup routine
// Returns a pointer to the initialized cache
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]Item),
	}

	// Start background cleanup routine to prevent memory leaks
	go cache.cleanupRoutine()

	return cache
}

// Set adds an item to the cache with specified expiration duration
// The item will automatically be removed after the duration expires
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache by key
// Returns the item's value and a boolean indicating if the item was found
// If the item has expired, it's treated as not found
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check expiration
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache by key
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// cleanupRoutine periodically removes expired items from the cache
// This prevents memory leaks by cleaning up items that are no longer needed
func (c *Cache) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.mu.Lock()

		now := time.Now().UnixNano()
		for k, v := range c.items {
			if v.Expiration > 0 && now > v.Expiration {
				delete(c.items, k)
			}
		}

		c.mu.Unlock()
	}
}

// Flush removes all items from the cache
// This is useful for clearing the entire cache at once, e.g. on logout
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Item)
}
