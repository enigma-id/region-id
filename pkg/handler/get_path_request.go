// Package handler provides HTTP request handlers for region-related operations.
// It includes request DTOs for searching, retrieving, and managing region data.
package handler

import (
	"context"

	"github.com/enigma-id/region-id/pkg/entity"
	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/google/uuid"
	"github.com/logistics-id/engine/transport/rest"
)

// GetPathRequest handles requests to get the hierarchy path of a region.
//
// Path parameter:
// - id: Region UUID
//
// Returns the full path from root to the specified region, including all ancestors.
// For example: [Province, Regency, District, Village]
type GetPathRequest struct {
	ID string `json:"id" param:"id"`

	repo repository.RegionRepository
	ctx  context.Context
}

// With sets the context and repository for the request.
func (r *GetPathRequest) With(ctx context.Context, repo repository.RegionRepository) *GetPathRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// with sets context and repository (internal use).
func (r *GetPathRequest) with(ctx context.Context, repo repository.RegionRepository) *GetPathRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// Execute retrieves the hierarchy path from root to the region.
//
// Returns a ResponseBody containing the path as an array of regions,
// ordered from root (province) to the target region (village).
func (r *GetPathRequest) Execute() (*rest.ResponseBody, error) {
	// Parse ID
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	// Get the target region
	region, err := r.repo.FindByID(r.ctx, id)
	if err != nil {
		return nil, err
	}

	// Build path by traversing up the hierarchy
	var path []*entity.Region
	current := region

	for current != nil {
		// Add current region to path
		path = append([]*entity.Region{current}, path...)

		// If no parent, we're done
		if current.ParentID == nil {
			break
		}

		// Get parent
		current, err = r.repo.FindByID(r.ctx, *current.ParentID)
		if err != nil {
			// If we can't find parent, stop here
			break
		}
	}

	// Return path
	return rest.NewResponseBody(path), nil
}
