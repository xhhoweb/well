package pool

import (
	"sync"
)

// Cache L1 Cache interface
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Remove(key K)
	Flush()
}

// SimpleCache 简单缓存实现
type SimpleCache[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]*V
	cap  int
}

// NewCache 创建新的SimpleCache
func NewCache[K comparable, V any](capacity int) *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		data: make(map[K]*V, capacity),
		cap:  capacity,
	}
}

// Get 获取缓存
func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	return *v, true
}

// Set 设置缓存
func (c *SimpleCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		c.data[key] = &value
		return
	}

	if len(c.data) >= c.cap {
		// 简单策略：清除第一个
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}

	c.data[key] = &value
}

// Remove 移除缓存
func (c *SimpleCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Flush 清空缓存
func (c *SimpleCache[K, V]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[K]*V)
}
