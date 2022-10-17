package mcache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

type MemCacheRepo interface {
	Get(string) (interface{}, bool)
	Set(string, interface{}, time.Duration) bool
}

type RistrettoCache struct {
	cache *ristretto.Cache
}

func NewMemCachedRepo() MemCacheRepo {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	return &RistrettoCache{cache: cache}
}

func (c *RistrettoCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *RistrettoCache) Set(key string, value interface{}, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, value, 1, ttl)
}
