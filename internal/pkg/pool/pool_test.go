package pool

import (
	"encoding/json"
	"fmt"
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
	cache, err := NewBigCache(64, 10*time.Minute)
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
		data, _ := json.Marshal(value)
		_ = cache.Set(key, data)
	}
}

func BenchmarkBigCache_Get(b *testing.B) {
	cache, err := NewBigCache(64, 10*time.Minute)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	for i := 0; i < 10000; i++ {
		key := formatKey(i)
		value := ThreadDTO{TID: int64(i), Subject: "Test Thread Subject", Views: 1000 + i, Replies: 10 + i%100, Dateline: time.Now().Unix()}
		data, _ := json.Marshal(value)
		_ = cache.Set(key, data)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := formatKey(i % 10000)
		if data, ok := cache.Get(key); ok {
			var dto ThreadDTO
			_ = json.Unmarshal(data, &dto)
		}
	}
}

func BenchmarkSimpleCache_Set(b *testing.B) {
	cache := NewSimpleCache[string, ThreadDTO]()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := formatKey(i)
		value := ThreadDTO{TID: int64(i), Subject: "Test Thread Subject", Views: 1000 + i, Replies: 10 + i%100, Dateline: time.Now().Unix()}
		cache.Set(key, value)
	}
}

func BenchmarkSimpleCache_Get(b *testing.B) {
	cache := NewSimpleCache[string, ThreadDTO]()

	for i := 0; i < 10000; i++ {
		key := formatKey(i)
		value := ThreadDTO{TID: int64(i), Subject: "Test Thread Subject", Views: 1000 + i, Replies: 10 + i%100, Dateline: time.Now().Unix()}
		cache.Set(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := formatKey(i % 10000)
		cache.Get(key)
	}
}

func formatKey(i int) string { return fmt.Sprintf("thread:%d", i) }
