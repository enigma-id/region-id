// Package regionid provides a comprehensive Indonesian regions library with search,
// caching, and auto-migration support.
//
// This library offers:
// - Complete database of Indonesian regions (provinces, regencies, districts, villages)
// - Fast search with filters by type, name, and parent
// - Optional Redis caching for performance via engine/ds/redis
// - Automatic migration support
// - REST API handlers using logistics-id/engine
//
// Basic usage with Redis caching:
//
//	import "github.com/enigma-id/region-id"
//	import "github.com/logistics-id/engine/ds/redis"
//
//	// Initialize Redis using engine's global singleton
//	redis.NewConnection(redis.Config{
//	    Server:   "localhost:6379",
//	    Password: "",
//	    Prefix:   "regions",
//	}, logger)
//
//	handler, err := regionid.Initialize(regionid.Config{
//	    DB:          db,
//	    AutoMigrate: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	handler.RegisterRoutes(server)
//
// Basic usage without caching:
//
//	handler, err := regionid.Initialize(regionid.Config{
//	    DB:          db,
//	    AutoMigrate: true,
//	})
package regionid

import (
	"context"
	"fmt"
	"log"

	"github.com/enigma-id/region-id/pkg/handler"
	"github.com/enigma-id/region-id/pkg/migration"
	"github.com/enigma-id/region-id/pkg/repository"
	"github.com/uptrace/bun"
)

// Handler is an alias for pkg/handler.Handler for easier external access
type Handler = handler.Handler

// Config holds configuration for the region-id library.
type Config struct {
	// Database connection (required)
	DB *bun.DB

	// AutoMigrate runs migrations on initialization (optional, default: false)
	// When true, all pending migrations will be run automatically
	AutoMigrate bool
}

// Initialize initializes the region-id library with the given configuration.
//
// It performs the following steps:
// 1. Validates the configuration (DB is required)
// 2. Runs migrations if AutoMigrate is enabled
// 3. Initializes cache manager using global Redis singleton (if available)
// 4. Creates the repository and handler
//
// Note: If you want to use Redis caching, call redis.NewConnection() before
// initializing this library. The cache manager will automatically use the
// global Redis singleton if it has been initialized.
//
// Returns an error if configuration is invalid or migrations fail.
func Initialize(cfg Config) (*Handler, error) {
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

	// Initialize cache manager (uses global Redis singleton if available)
	cache := repository.NewCacheManager()

	// Initialize repository
	repo := repository.NewRegionRepository(cfg.DB, cache)

	// Initialize handler
	h := handler.NewHandler(repo)

	return h, nil
}
