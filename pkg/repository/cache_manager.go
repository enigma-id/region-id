package repository

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager handles Redis caching for region data.
//
// The cache manager provides:
// - Automatic JSON serialization/deserialization
// - Base64 encoding to avoid Redis special character issues
// - Version-aware cache keys for automatic invalidation
// - Fallback handling for old cached data
//
// Cache entries use a 24-hour TTL by default.
type CacheManager struct {
	client *redis.Client
	prefix string
}

// NewCacheManager creates a new cache manager with the given Redis client.
//
// The cache prefix is set to "regions" for all keys.
func NewCacheManager(client *redis.Client) *CacheManager {
	return &CacheManager{
		client: client,
		prefix: "regions",
	}
}

// Get retrieves cached data by key and deserializes it into out.
//
// Returns (true, nil) if data is found and successfully deserialized.
// Returns (false, nil) on cache miss.
// Returns (false, err) on Redis errors (excluding redis.Nil).
//
// Handles both base64-encoded and plain JSON for backward compatibility.
func (cm *CacheManager) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	var encoded string
	if err := cm.client.Get(ctx, key).Scan(&encoded); err != nil {
		if err == redis.Nil {
			return false, nil // Cache miss
		}
		return false, err
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Fallback: might be plain JSON from old version
		if err := json.Unmarshal([]byte(encoded), out); err != nil {
			return false, nil
		}
		return true, nil
	}

	// Decode JSON
	if err := json.Unmarshal(decoded, out); err != nil {
		return false, nil
	}

	return true, nil
}

// Set stores data in cache with the specified TTL.
//
// Data is automatically serialized to JSON and base64-encoded.
// Use ttl=0 for no expiration (not recommended for most cases).
func (cm *CacheManager) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return cm.client.Set(ctx, key, encoded, ttl).Err()
}

// Delete removes a specific key from the cache.
//
// Silently succeeds if the key doesn't exist.
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.client.Del(ctx, key).Err()
}

// InvalidateAll clears all region caches by updating the cache version.
//
// This invalidates all existing cache keys by incrementing the version,
// causing them to be regenerated on next access.
func (cm *CacheManager) InvalidateAll(ctx context.Context) error {
	newVersion := time.Now().UTC().Format(time.RFC3339Nano)
	return cm.client.Set(ctx, fmt.Sprintf("%s:version", cm.prefix), newVersion, 0).Err()
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
