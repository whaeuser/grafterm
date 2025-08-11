package metric

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/slok/grafterm/internal/model"
)

// MetricCacheKey represents a unique cache key for metric queries
type MetricCacheKey struct {
	DatasourceID string
	Query        string
	Range        model.Range
}

// NewCacheKey creates a cache key from query parameters
func NewCacheKey(datasourceID, query string, tr model.TimeRange) MetricCacheKey {
	h := sha256.New()
	h.Write([]byte(datasourceID))
	h.Write([]byte(query))
	h.Write([]byte(fmt.Sprintf("%v:%v", tr.Start, tr.End)))
	
	return MetricCacheKey{
		DatasourceID: datasourceID,
		Query:        query,
		Range:        tr.Range,
	}
}

// cacheEntry holds cached metric data with expiration
type cacheEntry struct {
	data    []model.MetricSeries
	created time.Time
	expires time.Time
	hits    int64
}

// MetricCache provides thread-safe caching for metric data
type MetricCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxSize int64
	maxAge  time.Duration
	hits    int64
	misses  int64
}

// NewMetricCache creates a new metric cache with default settings
func NewMetricCache(maxSize int64, maxAge time.Duration) *MetricCache {
	cache := &MetricCache{
		entries: make(map[string]*cacheEntry),
		maxSize: maxSize,
		maxAge:  maxAge,
	}
	
	// Start cache cleanup routine
	go cache.cleanupRoutine()
	
	return cache
}

// Get retrieves metrics from cache if available
func (mc *MetricCache) Get(key MetricCacheKey) ([]model.MetricSeries, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s", key.DatasourceID, key.Query))))
	
	if entry, exists := mc.entries[cacheKey]; exists {
		if time.Now().Before(entry.expires) {
			entry.hits++
			mc.hits++
			return entry.data, true
		}
		delete(mc.entries, cacheKey)
	}
	
	mc.misses++
	return nil, false
}

// Set stores metrics in cache
func (mc *MetricCache) Set(key MetricCacheKey, data []model.MetricSeries) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if int64(len(mc.entries))*2 > mc.maxSize {
		mc.evictOldEntries()
	}
	
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s", key.DatasourceID, key.Query))))
	
	mc.entries[cacheKey] = &cacheEntry{
		data:    data,
		created: time.Now(),
		expires: time.Now().Add(mc.maxAge),
	}
}

// Stats returns cache statistics
func (mc *MetricCache) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	total := mc.hits + mc.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(mc.hits) / float64(total) * 100
	}
	
	return CacheStats{
		Hits:    mc.hits,
		Misses:  mc.misses,
		HitRate: hitRate,
		Size:    int64(len(mc.entries)),
	}
}

// Clear removes all entries from cache
func (mc *MetricCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.entries = make(map[string]*cacheEntry)
	mc.hits = 0
	mc.misses = 0
}

// evictOldEntries removes expired entries
func (mc *MetricCache) evictOldEntries() {
	now := time.Now()
	for key, entry := range mc.entries {
		if now.After(entry.expires) {
			delete(mc.entries, key)
		}
	}
}

// cleanupRoutine periodically removes expired entries
func (mc *MetricCache) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.evictOldEntries()
		}
	}
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	Hits    int64
	Misses  int64
	HitRate float64
	Size    int64
}