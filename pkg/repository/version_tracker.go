package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// DataVersionTracker tracks data version for cache invalidation.
//
// The version is based on the maximum updated_at timestamp in the regions table.
// When data changes, IncrementVersion should be called to update the version,
// causing all cached data to be regenerated on next access.
type DataVersionTracker struct {
	db    *bun.DB
	cache *CacheManager
}

// NewDataVersionTracker creates a new data version tracker.
func NewDataVersionTracker(db *bun.DB, cache *CacheManager) *DataVersionTracker {
	return &DataVersionTracker{
		db:    db,
		cache: cache,
	}
}

// GetDataVersion retrieves the current data version from cache or database.
//
// The version is derived from the maximum updated_at timestamp in the regions table.
// If not cached, it queries the database and caches the result.
//
// Returns empty string and an error if the database query fails.
func (dvt *DataVersionTracker) GetDataVersion(ctx context.Context) (string, error) {
	cacheKey := "regions:max_updated_at:v1"

	// Try to get from cache first
	var cachedVersion string
	found, _ := dvt.cache.Get(ctx, cacheKey, &cachedVersion)
	if found && cachedVersion != "" {
		return cachedVersion, nil
	}

	// Cache miss - query database
	var maxUpdatedAt time.Time
	err := dvt.db.NewRaw(
		"SELECT COALESCE(MAX(updated_at), 'epoch'::timestamptz) FROM regions WHERE is_deleted = false",
	).Scan(ctx, &maxUpdatedAt)
	if err != nil {
		return "", fmt.Errorf("failed to get data version from DB: %w", err)
	}

	version := maxUpdatedAt.UTC().Format(time.RFC3339Nano)

	// Cache the version (with no expiry - we update it when data changes)
	_ = dvt.cache.Set(ctx, cacheKey, version, 0)

	return version, nil
}

// IncrementVersion updates the data version to invalidate all caches.
//
// Call this method after any data modifications (insert, update, delete)
// to ensure cached data is regenerated with fresh values on next access.
func (dvt *DataVersionTracker) IncrementVersion(ctx context.Context) error {
	newVersion := time.Now().UTC().Format(time.RFC3339Nano)
	cacheKey := "regions:max_updated_at:v1"

	return dvt.cache.Set(ctx, cacheKey, newVersion, 0)
}
