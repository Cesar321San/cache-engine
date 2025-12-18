package cache

import (
	"fmt"
	"testing"
)

// BenchmarkSet mide el rendimiento de escrituras
func BenchmarkSet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}
}

// BenchmarkGet mide el rendimiento de lecturas
func BenchmarkGet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	// Pre-poblar el cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Get(key)
	}
}

// BenchmarkSetGet mide lecturas y escrituras mixtas
func BenchmarkSetGet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			key := fmt.Sprintf("key%d", i)
			cache.Set(key, i)
		} else {
			key := fmt.Sprintf("key%d", i-1)
			cache.Get(key)
		}
	}
}

// BenchmarkConcurrentSet mide escrituras concurrentes
func BenchmarkConcurrentSet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i)
			cache.Set(key, i)
			i++
		}
	})
}

// BenchmarkConcurrentGet mide lecturas concurrentes
func BenchmarkConcurrentGet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	// Pre-poblar el cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			cache.Get(key)
			i++
		}
	})
}

// BenchmarkConcurrentSetGet mide lecturas y escrituras concurrentes
func BenchmarkConcurrentSetGet(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				key := fmt.Sprintf("key%d", i)
				cache.Set(key, i)
			} else {
				key := fmt.Sprintf("key%d", i-1)
				cache.Get(key)
			}
			i++
		}
	})
}

// BenchmarkDelete mide el rendimiento de eliminaciones
func BenchmarkDelete(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	// Pre-poblar el cache
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Delete(key)
	}
}

// BenchmarkExpire mide el rendimiento de establecer expiraciones
func BenchmarkExpire(b *testing.B) {
	cache := NewCacheEngine(10000)
	defer cache.Close()

	// Pre-poblar el cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Expire(key, 60)
	}
}
