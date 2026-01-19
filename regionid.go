// Package regionid provides a comprehensive Indonesian regions library with search,
// caching, and auto-migration support.
//
// This library offers:
// - Complete database of Indonesian regions (provinces, regencies, districts, villages)
// - Fast search with filters by type, name, and parent
// - Optional Redis caching for performance
// - Automatic migration support
// - REST API handlers using logistics-id/engine
//
// Basic usage:
//
//	import "github.com/enigma-id/region-id"
//
//	handler, err := regionid.Initialize(regionid.Config{
//	    DB:          db,
//	    Redis:       redis,
//	    AutoMigrate: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	handler.RegisterRoutes(server)
package regionid

import (
	"context"
	"fmt"
	"log"

	"github.com/enigma-id/region-id/pkg/handler"
	"github.com/enigma-id/region-id/pkg/migration"
	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
)

// Config holds configuration for the region-id library.
type Config struct {
	// Database connection (required)
	DB *bun.DB

	// Redis client (optional - if nil, caching is disabled)
	Redis *redis.Client

	// AutoMigrate runs migrations on initialization (optional, default: false)
	// When true, all pending migrations will be run automatically
	AutoMigrate bool
}

// Initialize initializes the region-id library with the given configuration.
//
// It performs the following steps:
// 1. Validates the configuration (DB is required)
// 2. Runs migrations if AutoMigrate is enabled
// 3. Initializes cache manager if Redis is provided
// 4. Creates the repository and handler
//
// Returns an error if configuration is invalid or migrations fail.
func Initialize(cfg Config) (*handler.Handler, error) {
	// Validate config
	if cfg.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	// Run migrations if AutoMigrate is enabled
	if cfg.AutoMigrate {
		log.Println("Running auto-migration...")
		migrator := migration.NewMigrator(cfg.DB)
		if err := migrator.Up(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
		log.Println("Migrations completed successfully")
	}

	// Initialize cache manager (if Redis provided)
	var cache *repository.CacheManager
	if cfg.Redis != nil {
		cache = repository.NewCacheManager(cfg.Redis)
	}

	// Initialize repository
	repo := repository.NewRegionRepository(cfg.DB, cache)

	// Initialize handler
	h := handler.NewHandler(repo)

	return h, nil
}
