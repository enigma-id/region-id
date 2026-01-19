// Package handler provides HTTP request handlers for region-related operations.
// It includes request DTOs for searching, retrieving, and managing region data.
package handler

import (
	"context"

	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/google/uuid"
	"github.com/logistics-id/engine/transport/rest"
)

// GetChildrenRequest handles requests to get children of a region.
//
// Path parameter:
// - id: Parent Region UUID
//
// Returns all direct children (one level down) ordered by name.
type GetChildrenRequest struct {
	ID string `json:"id" param:"id"`

	repo repository.RegionRepository
	ctx  context.Context
}

// With sets the context and repository for the request.
func (r *GetChildrenRequest) With(ctx context.Context, repo repository.RegionRepository) *GetChildrenRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// with sets context and repository (internal use).
func (r *GetChildrenRequest) with(ctx context.Context, repo repository.RegionRepository) *GetChildrenRequest {
	r.ctx = ctx
	r.repo = repo.WithContext(ctx)
	return r
}

// Execute retrieves the children of the region.
//
// Returns a ResponseBody containing all direct children, ordered by name.
func (r *GetChildrenRequest) Execute() (*rest.ResponseBody, error) {
	// Parse ID
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	// Call repository
	results, err := r.repo.FindByParent(r.ctx, id)
	if err != nil {
		return nil, err
	}

	// Return entities directly
	return rest.NewResponseBody(results), nil
}
