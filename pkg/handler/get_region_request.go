// Package handler provides HTTP request handlers for region-related operations.
// It includes request DTOs for searching, retrieving, and managing region data.
package handler

import (
	"context"

	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/google/uuid"
	"github.com/logistics-id/engine/transport/rest"
)

// GetRegionRequest handles requests to get a region by ID.
//
// Path parameter:
// - id: Region UUID
type GetRegionRequest struct {
	ID string `json:"id" param:"id"`

	repo repository.RegionRepository
	ctx  context.Context
}

// With sets the context and repository for the request.
func (r *GetRegionRequest) With(ctx context.Context, repo repository.RegionRepository) *GetRegionRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// with sets context and repository (internal use).
func (r *GetRegionRequest) with(ctx context.Context, repo repository.RegionRepository) *GetRegionRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// Execute retrieves the region by ID.
//
// Returns a ResponseBody containing the region data.
func (r *GetRegionRequest) Execute() (*rest.ResponseBody, error) {
	// Parse ID
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	// Call repository
	result, err := r.repo.FindByID(r.ctx, id)
	if err != nil {
		return nil, err
	}

	// Return entity directly
	return rest.NewResponseBody(result), nil
}
