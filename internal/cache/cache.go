package cache

import (
	"sync"
	"time"
)

func New[K comparable, V any](lifeTime time.Duration, retrieveFunc func(K) V, saveFunc func(K, V)) *Cache[K, V] {
	cache := &Cache[K, V]{
		lifeTime:     lifeTime,
		cache:        map[K]*entry[V]{},
		retrieveFunc: retrieveFunc,
		saveFunc:     saveFunc,
	}
	cache.StartCleanup()
	return cache
}

type Cache[K comparable, V any] struct {
	lifeTime     time.Duration
	mu           sync.Mutex
	cache        map[K]*entry[V]
	retrieveFunc func(K) V
	saveFunc     func(K, V)
}

type entry[V any] struct {
	v          V
	lastAccess time.Time
}

func (m *Cache[K, V]) StartCleanup() {
	go func() {
		for range time.After(time.Minute * 1) {
			m.mu.Lock()
			for k, v := range m.cache {
				if time.Since(v.lastAccess) > m.lifeTime {
					delete(m.cache, k)
				}
			}
			m.mu.Unlock()
		}
	}()
}

func (m *Cache[K, V]) Get(k K) V {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.cache[k]; ok {
		v.lastAccess = time.Now()
		return v.v
	}
	v := m.retrieveFunc(k)
	m.cache[k] = &entry[V]{
		v:          v,
		lastAccess: time.Now(),
	}
	return v
}

func (m *Cache[K, V]) Set(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[k] = &entry[V]{
		v:          v,
		lastAccess: time.Now(),
	}
	m.saveFunc(k, v)
}
