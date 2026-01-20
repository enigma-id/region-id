package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/enigma-id/region-id/pkg/entity"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/sync/singleflight"
)

type regionRepositoryImpl struct {
	db             *bun.DB
	cache          *CacheManager
	versionTracker *DataVersionTracker
	sf             singleflight.Group
}

// NewRegionRepository creates a new region repository with the given database and cache.
//
// If cache is nil, the repository will query the database directly without caching.
// The repository uses singleflight to prevent cache stampede during high load.
func NewRegionRepository(db *bun.DB, cache *CacheManager) RegionRepository {
	return &regionRepositoryImpl{
		db:             db,
		cache:          cache,
		versionTracker: NewDataVersionTracker(db, cache),
		sf:             singleflight.Group{},
	}
}

// Search searches regions with filters and caching
func (r *regionRepositoryImpl) Search(ctx context.Context, query string, opts SearchOptions) ([]*entity.Region, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// If cache is not configured, query database directly
	if r.cache == nil {
		return r.searchDB(ctx, query, opts)
	}

	// 1) Get data version for cache key
	version, err := r.versionTracker.GetDataVersion(ctx)
	if err != nil {
		// If version lookup fails, bypass cache and query DB directly
		log.Printf("[WARN] Failed to get data version, bypassing cache: %v", err)
		return r.searchDB(ctx, query, opts)
	}

	// 2) Generate cache key
	cacheKey := r.cache.GenerateCacheKey(version, "search", map[string]interface{}{
		"q":        query,
		"type":     opts.Type,
		"parentID": opts.ParentID,
		"limit":    opts.Limit,
	})

	// 3) Try to get from cache
	var cachedResults []*entity.Region
	found, err := r.cache.Get(ctx, cacheKey, &cachedResults)
	if err == nil && found && len(cachedResults) > 0 {
		return cachedResults, nil
	}

	// 4) Use singleflight to prevent cache stampede
	result, err, _ := r.sf.Do(cacheKey, func() (interface{}, error) {
		// Check cache again after waiting for singleflight
		var doubleCheck []*entity.Region
		if found, err := r.cache.Get(ctx, cacheKey, &doubleCheck); err == nil && found && len(doubleCheck) > 0 {
			return doubleCheck, nil
		}

		// Query database
		regions, err := r.searchDB(ctx, query, opts)
		if err != nil {
			return nil, err
		}

		// Cache results (24 hour TTL)
		if len(regions) > 0 {
			if err := r.cache.Set(ctx, cacheKey, regions, 24*time.Hour); err != nil {
				log.Printf("[WARN] Failed to cache search results: %v", err)
				// Don't fail - just return the results without caching
			}
		}

		return regions, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entity.Region), nil
}

// searchDB performs the actual database search
func (r *regionRepositoryImpl) searchDB(ctx context.Context, query string, opts SearchOptions) ([]*entity.Region, error) {
	var regions []*entity.Region

	// Use search_regions() SQL function when query is provided and no parent filter.
	// The SQL function searches in: name, code, postal_code, AND administrative_area JSONB.
	// It also provides ranking based on match quality.
	//
	// Fallback to query builder when parent filter is used (SQL function doesn't support parent_id).
	if query != "" && opts.ParentID == nil {
		limit := opts.Limit
		if limit <= 0 {
			limit = 10
		}

		var err error
		if opts.Type != "" {
			// With type filter - use raw SQL query to avoid CROSS JOIN
			err = r.db.NewRaw(
				"SELECT id, name, code, type, postal_code, parent_id, level, administrative_area FROM search_regions(?, ARRAY[?::TEXT], ?)",
				query, opts.Type, limit,
			).Scan(ctx, &regions)
		} else {
			// Without type filter - use raw SQL query to avoid CROSS JOIN
			err = r.db.NewRaw(
				"SELECT id, name, code, type, postal_code, parent_id, level, administrative_area FROM search_regions(?, NULL, ?)",
				query, limit,
			).Scan(ctx, &regions)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search regions: %w", err)
		}
		return regions, nil
	}

	// Fallback: Original query builder for parent filtering or empty query
	queryBuilder := r.db.NewSelect().
		Model(&regions).
		Where("is_deleted = false")

	// Add text search if query provided (limited to name field in this path)
	if query != "" {
		queryBuilder = queryBuilder.Where("name ILIKE ?", "%"+query+"%")
	}

	// Filter by type
	if opts.Type != "" {
		queryBuilder = queryBuilder.Where("type = ?", opts.Type)
	}

	// Filter by parent
	if opts.ParentID != nil {
		queryBuilder = queryBuilder.Where("parent_id = ?", *opts.ParentID)
	}

	// Apply limit
	queryBuilder = queryBuilder.Limit(opts.Limit)

	err := queryBuilder.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search regions: %w", err)
	}

	return regions, nil
}

// FindByID finds a region by ID with caching
func (r *regionRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Region, error) {
	// 1) Get data version
	version, err := r.versionTracker.GetDataVersion(ctx)
	if err != nil {
		log.Printf("[WARN] Failed to get data version, bypassing cache: %v", err)
		return r.findByIDDB(ctx, id)
	}

	// 2) Generate cache key
	cacheKey := r.cache.GenerateCacheKey(version, "id", map[string]interface{}{
		"id": id.String(),
	})

	// 3) Try cache
	var cachedRegion *entity.Region
	found, err := r.cache.Get(ctx, cacheKey, &cachedRegion)
	if err == nil && found && cachedRegion != nil {
		return cachedRegion, nil
	}

	// 4) Use singleflight
	result, err, _ := r.sf.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache
		var doubleCheck *entity.Region
		if found, err := r.cache.Get(ctx, cacheKey, &doubleCheck); err == nil && found && doubleCheck != nil {
			return doubleCheck, nil
		}

		// Query DB
		region, err := r.findByIDDB(ctx, id)
		if err != nil {
			return nil, err
		}

		// Cache result
		if region != nil {
			if err := r.cache.Set(ctx, cacheKey, region, 24*time.Hour); err != nil {
				log.Printf("[WARN] Failed to cache region by ID: %v", err)
			}
		}

		return region, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entity.Region), nil
}

// findByIDDB performs the actual database query for FindByID
func (r *regionRepositoryImpl) findByIDDB(ctx context.Context, id uuid.UUID) (*entity.Region, error) {
	var region entity.Region
	err := r.db.NewSelect().
		Model(&region).
		Where("id = ?", id).
		Where("is_deleted = false").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find region by ID: %w", err)
	}

	return &region, nil
}

// FindByCode finds a region by code with caching
func (r *regionRepositoryImpl) FindByCode(ctx context.Context, code string) (*entity.Region, error) {
	// 1) Get data version
	version, err := r.versionTracker.GetDataVersion(ctx)
	if err != nil {
		log.Printf("[WARN] Failed to get data version, bypassing cache: %v", err)
		return r.findByCodeDB(ctx, code)
	}

	// 2) Generate cache key
	cacheKey := r.cache.GenerateCacheKey(version, "code", map[string]interface{}{
		"code": code,
	})

	// 3) Try cache
	var cachedRegion *entity.Region
	found, err := r.cache.Get(ctx, cacheKey, &cachedRegion)
	if err == nil && found && cachedRegion != nil {
		return cachedRegion, nil
	}

	// 4) Use singleflight
	result, err, _ := r.sf.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache
		var doubleCheck *entity.Region
		if found, err := r.cache.Get(ctx, cacheKey, &doubleCheck); err == nil && found && doubleCheck != nil {
			return doubleCheck, nil
		}

		// Query DB
		region, err := r.findByCodeDB(ctx, code)
		if err != nil {
			return nil, err
		}

		// Cache result
		if region != nil {
			if err := r.cache.Set(ctx, cacheKey, region, 24*time.Hour); err != nil {
				log.Printf("[WARN] Failed to cache region by code: %v", err)
			}
		}

		return region, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entity.Region), nil
}

// findByCodeDB performs the actual database query for FindByCode
func (r *regionRepositoryImpl) findByCodeDB(ctx context.Context, code string) (*entity.Region, error) {
	var region entity.Region
	err := r.db.NewSelect().
		Model(&region).
		Where("code = ?", code).
		Where("is_deleted = false").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find region by code: %w", err)
	}

	return &region, nil
}

// FindByParent finds all child regions with caching
func (r *regionRepositoryImpl) FindByParent(ctx context.Context, parentID uuid.UUID) ([]*entity.Region, error) {
	// 1) Get data version
	version, err := r.versionTracker.GetDataVersion(ctx)
	if err != nil {
		log.Printf("[WARN] Failed to get data version, bypassing cache: %v", err)
		return r.findByParentDB(ctx, parentID)
	}

	// 2) Generate cache key
	cacheKey := r.cache.GenerateCacheKey(version, "children", map[string]interface{}{
		"parentID": parentID.String(),
	})

	// 3) Try cache
	var cachedRegions []*entity.Region
	found, err := r.cache.Get(ctx, cacheKey, &cachedRegions)
	if err == nil && found && len(cachedRegions) > 0 {
		return cachedRegions, nil
	}

	// 4) Use singleflight
	result, err, _ := r.sf.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache
		var doubleCheck []*entity.Region
		if found, err := r.cache.Get(ctx, cacheKey, &doubleCheck); err == nil && found && len(doubleCheck) > 0 {
			return doubleCheck, nil
		}

		// Query DB
		regions, err := r.findByParentDB(ctx, parentID)
		if err != nil {
			return nil, err
		}

		// Cache results
		if len(regions) > 0 {
			if err := r.cache.Set(ctx, cacheKey, regions, 24*time.Hour); err != nil {
				log.Printf("[WARN] Failed to cache children: %v", err)
			}
		}

		return regions, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entity.Region), nil
}

// findByParentDB performs the actual database query for FindByParent
func (r *regionRepositoryImpl) findByParentDB(ctx context.Context, parentID uuid.UUID) ([]*entity.Region, error) {
	var regions []*entity.Region
	err := r.db.NewSelect().
		Model(&regions).
		Where("parent_id = ?", parentID).
		Where("is_deleted = false").
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find regions by parent: %w", err)
	}

	return regions, nil
}

// WithContext returns a new repository instance with the given context
func (r *regionRepositoryImpl) WithContext(ctx context.Context) RegionRepository {
	// For this implementation, context is passed to each method
	// so we just return the same instance
	// If you need context-specific behavior, you can create a new instance
	return r
}
