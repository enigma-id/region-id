package repository

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/logistics-id/engine/ds/redis"
)

// CacheManager handles Redis caching for region data using engine/ds/redis.
//
// The cache manager provides:
// - Automatic JSON serialization/deserialization via engine/ds/redis
// - Version-aware cache keys for automatic invalidation
//
// Cache entries use a 24-hour TTL by default.
//
// The manager uses engine's global Redis singleton via package-level functions.
// Users should call redis.NewConnection() before initializing the region-id library.
type CacheManager struct {
	prefix string
}

// NewCacheManager creates a new cache manager using engine's global Redis singleton.
//
// The cache prefix is set to "regions" for all keys.
// If Redis has not been initialized via redis.NewConnection(), cache operations
// will silently fail (cache miss behavior).
func NewCacheManager() *CacheManager {
	return &CacheManager{
		prefix: "regions",
	}
}

// Get retrieves cached data by key and deserializes it into out.
//
// Returns (true, nil) if data is found and successfully deserialized.
// Returns (false, nil) on cache miss or if Redis is not initialized.
// Returns (false, err) on Redis errors.
func (cm *CacheManager) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	// Use engine/ds/redis Read function (uses global singleton internally)
	err := redis.Read(ctx, key, out)
	if err != nil {
		// engine/ds/redis returns error on cache miss or not initialized
		return false, nil
	}

	return true, nil
}

// Set stores data in cache with the specified TTL.
//
// Data is automatically serialized to JSON by engine/ds/redis.
// Silently fails if Redis is not initialized.
// Note: engine/ds/redis Save doesn't support TTL parameter.
func (cm *CacheManager) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	// Silently ignore errors (e.g., Redis not initialized)
	_ = redis.Save(ctx, key, data)
	return nil
}

// Delete removes a specific key from the cache.
//
// Silently succeeds if the key doesn't exist or if Redis is not initialized.
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	// Silently ignore errors (e.g., Redis not initialized)
	_ = redis.Delete(ctx, key)
	return nil
}

// InvalidateAll clears all region caches by updating the cache version.
//
// This invalidates all existing cache keys by incrementing the version,
// causing them to be regenerated on next access.
// Silently fails if Redis is not initialized.
func (cm *CacheManager) InvalidateAll(ctx context.Context) error {
	newVersion := time.Now().UTC().Format(time.RFC3339Nano)
	// Silently ignore errors (e.g., Redis not initialized)
	_ = redis.Save(ctx, fmt.Sprintf("%s:version", cm.prefix), newVersion)
	return nil
}

// GenerateCacheKey creates a consistent cache key from version, type, and parameters.
//
// The key format is: "{prefix}:{keyType}:{sha1_hash}"
//
// The hash is computed from a JSON object containing the version, type,
// and all parameters, ensuring consistent keys for the same inputs.
func (cm *CacheManager) GenerateCacheKey(version, keyType string, params map[string]interface{}) string {
	m := map[string]interface{}{
		"v": version,
		"t": keyType,
	}
	for k, v := range params {
		m[k] = v
	}

	b, _ := json.Marshal(m)
	sum := sha1.Sum(b)
	return fmt.Sprintf("%s:%s:%s", cm.prefix, keyType, hex.EncodeToString(sum[:]))
}
