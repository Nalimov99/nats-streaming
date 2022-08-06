package cache

import (
	"errors"
	"sync"
)

var ErrItemNotFound = errors.New("item not found")

type Cache struct {
	mutex sync.RWMutex
	items map[string][]byte
}

// Set knows how to write value into Cache. It locks Cache when writing.
func (c *Cache) Set(key string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[key] = value
}

// NewCache knows how to initialize Cache
func NewCache(initialCacheItems map[string][]byte) *Cache {
	var cache Cache
	cache.items = initialCacheItems
	return &cache
}

// Get knows how to retrieve item from Cache by key. It will error if the
// specified key does not reference the existing item.
func (c *Cache) Get(key string) (*[]byte, error) {
	item, ok := c.items[key]

	if !ok {
		return nil, ErrItemNotFound
	}

	return &item, nil
}

// GetItems returns all known Orders
func (c *Cache) GetItems() map[string][]byte {
	return c.items
}
