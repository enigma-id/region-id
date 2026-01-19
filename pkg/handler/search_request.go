// Package handler provides HTTP request handlers for region-related operations.
// It includes request DTOs for searching, retrieving, and managing region data.
package handler

import (
	"context"

	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/google/uuid"
	"github.com/logistics-id/engine/transport/rest"
)

// SearchRequest handles region search requests.
//
// Query parameters:
// - q: Search query (region name, case-insensitive partial match)
// - type: Filter by region type (province, regency, district, village)
// - parent_id: Filter by parent UUID
// - limit: Maximum results (default: 10)
// - page: Page number for pagination (default: 1)
type SearchRequest struct {
	Query    string `json:"q" query:"q"`
	Type     string `json:"type" query:"type"`
	ParentID string `json:"parent_id" query:"parent_id"`
	Limit    int64  `json:"limit" query:"limit"`
	Page     int64  `json:"page" query:"page"`

	repo repository.RegionRepository
	ctx  context.Context
}

// With sets the context and repository for the request.
//
// This is a fluent method that allows chaining before Execute.
// With is exported for external use, while with is for internal handler use.
func (r *SearchRequest) With(ctx context.Context, repo repository.RegionRepository) *SearchRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// with sets context and repository (internal use, lowercase).
func (r *SearchRequest) with(ctx context.Context, repo repository.RegionRepository) *SearchRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// Execute performs the search operation using the configured parameters.
//
// Returns a ResponseBody containing the matching regions.
func (r *SearchRequest) Execute() (*rest.ResponseBody, error) {
	// Set defaults
	if r.Limit <= 0 {
		r.Limit = 10
	}

	// Parse parent_id if provided
	var parentID *uuid.UUID
	if r.ParentID != "" {
		pid, err := uuid.Parse(r.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	// Call repository
	results, err := r.repo.Search(r.ctx, r.Query, repository.SearchOptions{
		Type:     r.Type,
		ParentID: parentID,
		Limit:    int(r.Limit),
	})

	if err != nil {
		return nil, err
	}

	// Return entity directly - no separate response DTO
	return rest.NewResponseBody(results), nil
}
