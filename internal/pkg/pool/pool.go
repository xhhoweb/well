package pool

import (
	"time"

	"github.com/allegro/bigcache/v3"
)

// Cache L1 Cache interface（零GC设计）
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Remove(key K)
	Flush()
}

// BigCache bigcache包装器
// 设计原则：
// 1. 底层直接使用bigcache的[]byte接口
// 2. 序列化/反序列化在Service层处理
// 3. Cache层只负责存储，无额外GC分配
type BigCache struct {
	cache *bigcache.BigCache
}

// NewBigCache 创建bigcache实例
// capacityMB: 缓存容量（MB）
// expiration: 过期时间
func NewBigCache(capacityMB int, expiration time.Duration) (*BigCache, error) {
	config := bigcache.DefaultConfig(expiration)
	config.HardMaxCacheSize = capacityMB
	config.MaxEntrySize = 512 * 1024 // 512KB max entry

	cache, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}

	return &BigCache{cache: cache}, nil
}

// Get 直接返回[]byte，由上层反序列化
// 零GC：直接返回底层数据，无拷贝
func (c *BigCache) Get(key string) ([]byte, bool) {
	data, err := c.cache.Get(key)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Set 直接存储[]byte，由上层序列化
// 零GC：直接传入底层存储，无转换
func (c *BigCache) Set(key string, value []byte) error {
	return c.cache.Set(key, value)
}

// Remove 删除键
func (c *BigCache) Remove(key string) error {
	return c.cache.Delete(key)
}

// Flush 清空所有缓存
func (c *BigCache) Flush() error {
	return c.cache.Reset()
}

// Close 关闭缓存
func (c *BigCache) Close() error {
	return c.cache.Close()
}

// SimpleCache 简单缓存（map实现，作为备选）
// 注意：map实现有GC压力，不适合高频场景
type SimpleCache[K comparable, V any] struct {
	data map[K]*V
}

// NewSimpleCache 创建简单缓存
func NewSimpleCache[K comparable, V any]() *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		data: make(map[K]*V),
	}
}

func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	v, ok := c.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	return *v, true
}

func (c *SimpleCache[K, V]) Set(key K, value V) {
	c.data[key] = &value
}

func (c *SimpleCache[K, V]) Remove(key K) {
	delete(c.data, key)
}

func (c *SimpleCache[K, V]) Flush() {
	c.data = make(map[K]*V)
}
