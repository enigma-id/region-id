// Package repository provides data access interfaces and implementations
// for region data with caching support.
//
// The repository package includes:
// - RegionRepository: Interface for region data access
// - CacheManager: Redis-based caching with automatic invalidation
// - DataVersionTracker: Tracks data changes for cache invalidation
//
// The implementation uses singleflight pattern to prevent cache stampede
// and supports optional Redis caching for performance optimization.
package repository

import (
	"context"

	"github.com/enigma-id/region-id/pkg/entity"
	"github.com/google/uuid"
)

// SearchOptions specifies filtering criteria for region searches.
type SearchOptions struct {
	// Type filters by region type (province, regency, district, village)
	Type string
	// ParentID filters by parent region UUID
	ParentID *uuid.UUID
	// Limit specifies maximum number of results (default: 10)
	Limit int
}

// RegionRepository defines the interface for region data access operations.
//
// The repository provides CRUD operations with optional caching support.
// All methods accept a context for timeout/cancellation control.
type RegionRepository interface {
	// Search searches for regions matching the query with optional filters.
	//
	// The query performs a case-insensitive partial match on region names.
	// Results can be filtered by type and parent ID.
	Search(ctx context.Context, query string, opts SearchOptions) ([]*entity.Region, error)

	// FindByID retrieves a region by its UUID.
	//
	// Returns an error if the region is not found or deleted.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Region, error)

	// FindByCode retrieves a region by its code.
	//
	// Returns an error if the region is not found or deleted.
	FindByCode(ctx context.Context, code string) (*entity.Region, error)

	// FindByParent retrieves all direct children of a region.
	//
	// Results are ordered by name alphabetically.
	FindByParent(ctx context.Context, parentID uuid.UUID) ([]*entity.Region, error)

	// WithContext returns a repository instance bound to the given context.
	//
	// This allows for context-specific behavior like request-scoped transactions.
	WithContext(ctx context.Context) RegionRepository
}
