package pool

import (
	"testing"
	"time"
)

// ThreadDTO 测试用DTO
type ThreadDTO struct {
	TID      int64  `json:"tid"`
	Subject  string `json:"subject"`
	Views    int    `json:"views"`
	Replies  int    `json:"replies"`
	Dateline int64  `json:"dateline"`
}

func BenchmarkBigCache_Set(b *testing.B) {
	cache, err := NewBigCache[string, ThreadDTO](64, 10*time.Minute)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := formatKey(i)
		value := ThreadDTO{
			TID:      int64(i),
			Subject:  "Test Thread Subject",
			Views:    1000 + i,
			Replies:  10 + i%100,
			Dateline: time.Now().Unix(),
		}
		cache.Set(key, value)
	}
}

func BenchmarkBigCache_Get(b *testing.B) {
	cache, err := NewBigCache[string, ThreadDTO](64, 10*time.Minute)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	// 预热数据
	for i := 0; i < 10000; i++ {
		key := formatKey(i)
		value := ThreadDTO{
			TID:      int64(i),
			Subject:  "Test Thread Subject",
			Views:    1000 + i,
			Replies:  10 + i%100,
			Dateline: time.Now().Unix(),
		}
		cache.Set(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 100% 缓存命中
	for i := 0; i < b.N; i++ {
		key := formatKey(i % 10000)
		cache.Get(key)
	}
}

func BenchmarkBigCache_Get_L2Pattern(b *testing.B) {
	cache, err := NewBigCache[string, ThreadDTO](64, 10*time.Minute)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	// 模拟 L2 模式：10% 热点数据
	hotData := make([]ThreadDTO, 1000)
	for i := 0; i < 1000; i++ {
		key := formatKey(i)
		value := ThreadDTO{
			TID:      int64(i),
			Subject:  "Hot Thread",
			Views:    100000 + i,
			Replies:  1000 + i,
			Dateline: time.Now().Unix(),
		}
		cache.Set(key, value)
		hotData[i] = value
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 90% L1命中，10% L2命中
	for i := 0; i < b.N; i++ {
		if i%10 == 0 {
			// 10% L2 miss -> write L1
			key := formatKey(10000 + i)
			cache.Set(key, hotData[i%1000])
		} else {
			// 90% L1 hit
			key := formatKey(i % 1000)
			cache.Get(key)
		}
	}
}

func BenchmarkSimpleCache_Set(b *testing.B) {
	cache := NewCache[string, ThreadDTO](10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := formatKey(i)
		value := ThreadDTO{
			TID:      int64(i),
			Subject:  "Test Thread Subject",
			Views:    1000 + i,
			Replies:  10 + i%100,
			Dateline: time.Now().Unix(),
		}
		cache.Set(key, value)
	}
}

func BenchmarkSimpleCache_Get(b *testing.B) {
	cache := NewCache[string, ThreadDTO](10000)

	// 预热数据
	for i := 0; i < 10000; i++ {
		key := formatKey(i)
		value := ThreadDTO{
			TID:      int64(i),
			Subject:  "Test Thread Subject",
			Views:    1000 + i,
			Replies:  10 + i%100,
			Dateline: time.Now().Unix(),
		}
		cache.Set(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// 100% 缓存命中
	for i := 0; i < b.N; i++ {
		key := formatKey(i % 10000)
		cache.Get(key)
	}
}

func formatKey(i int) string {
	return "thread:" + string(rune('a'+i%26)) + itoa(i)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
