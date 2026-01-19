// Package handler provides HTTP request handlers for region-related operations.
// It includes request DTOs for searching, retrieving, and managing region data.
package handler

import (
	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/logistics-id/engine/transport/rest"
)

// Handler manages HTTP requests for region operations
type Handler struct {
	repo repository.RegionRepository
}

// NewHandler creates a new region handler
func NewHandler(repo repository.RegionRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// RegisterRoutes registers all region routes to the given RestServer
func (h *Handler) RegisterRoutes(s *rest.RestServer) {
	s.GET("/regions/search", h.Search, nil)
	s.GET("/regions/{id}", h.GetRegion, nil)
	s.GET("/regions/{id}/children", h.GetChildren, nil)
	s.GET("/regions/{id}/path", h.GetPath, nil)
}

// Search handles region search requests
// @Summary Search regions
// @Tags regions
// @Accept json
// @Produce json
// @Param q query string false "Search query (region name)"
// @Param type query string false "Filter by type (province, regency, district, village)"
// @Param parent_id query string false "Filter by parent UUID"
// @Param limit query int false "Max results (default: 10)"
// @Param page query int false "Page number (default: 1)"
// @Success 200 {object} rest.ResponseBody
// @Failure 400 {object} rest.HTTPError
// @Failure 500 {object} rest.HTTPError
// @Router /regions/search [get]
func (h *Handler) Search(ctx *rest.Context) error {
	var req SearchRequest

	if err := ctx.Bind(&req); err != nil {
		return err
	}

	req = *req.with(ctx, h.repo)
	res, err := req.Execute()

	return ctx.Respond(res, err)
}

// GetRegion retrieves a region by ID
// @Summary Get region by ID
// @Tags regions
// @Accept json
// @Produce json
// @Param id path string true "Region UUID"
// @Success 200 {object} rest.ResponseBody
// @Failure 404 {object} rest.HTTPError
// @Failure 500 {object} rest.HTTPError
// @Router /regions/{id} [get]
func (h *Handler) GetRegion(ctx *rest.Context) error {
	var req GetRegionRequest

	if err := ctx.Bind(&req); err != nil {
		return err
	}

	req = *req.with(ctx, h.repo)
	res, err := req.Execute()

	return ctx.Respond(res, err)
}

// GetChildren retrieves children of a region
// @Summary Get region children
// @Tags regions
// @Accept json
// @Produce json
// @Param id path string true "Parent Region UUID"
// @Success 200 {object} rest.ResponseBody
// @Failure 404 {object} rest.HTTPError
// @Failure 500 {object} rest.HTTPError
// @Router /regions/{id}/children [get]
func (h *Handler) GetChildren(ctx *rest.Context) error {
	var req GetChildrenRequest

	if err := ctx.Bind(&req); err != nil {
		return err
	}

	req = *req.with(ctx, h.repo)
	res, err := req.Execute()

	return ctx.Respond(res, err)
}

// GetPath retrieves the hierarchy path for a region
// @Summary Get region hierarchy path
// @Tags regions
// @Accept json
// @Produce json
// @Param id path string true "Region UUID"
// @Success 200 {object} rest.ResponseBody
// @Failure 404 {object} rest.HTTPError
// @Failure 500 {object} rest.HTTPError
// @Router /regions/{id}/path [get]
func (h *Handler) GetPath(ctx *rest.Context) error {
	var req GetPathRequest

	if err := ctx.Bind(&req); err != nil {
		return err
	}

	req = *req.with(ctx, h.repo)
	res, err := req.Execute()

	return ctx.Respond(res, err)
}
