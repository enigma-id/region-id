package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/enigma-id/region-id"
	"github.com/joho/godotenv"
	"github.com/logistics-id/engine"
	"github.com/logistics-id/engine/ds/redis"
	"github.com/logistics-id/engine/transport/rest"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"
)

var regionHandler *regionid.Handler

func init() {
	// Load .env file if exists
	godotenv.Load()

	// Initialize engine with service name
	engine.Init("region-id-server")
}

func main() {
	engine.OnStart(func(ctx context.Context) error {
		logger := engine.Logger

		// Setup database connection
		databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable")
		sqldb := sql.OpenDB(pgdriver.NewConnector(
			pgdriver.WithDSN(databaseURL),
		))

		db := bun.NewDB(sqldb, pgdialect.New())

		// Add query logger for debugging (remove in production)
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))

		// Setup Redis using engine's redis package (optional)
		redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
		if redisAddr != "" {
			cfg := &redis.Config{
				Prefix:   "regions",
				Server:   redisAddr,
				Password: getEnv("REDIS_PASSWORD", ""),
			}

			// Initialize Redis connection via engine
			// The region-id library will use the global singleton automatically
			if err := redis.NewConnection(cfg, logger); err != nil {
				logger.Warn("Redis connection failed, caching will be disabled", zap.Error(err))
			} else {
				logger.Info("Redis connected successfully - caching enabled")
			}
		}

		// Initialize region-id library with auto-migration
		logger.Info("Initializing region-id library...")
		handler, err := regionid.Initialize(regionid.Config{
			DB:          db,
			AutoMigrate: true,
		})
		if err != nil {
			return err
		}

		regionHandler = handler
		logger.Info("region-id library initialized successfully")

		return nil
	})

	engine.OnStop(func(ctx context.Context) {
		logger := engine.Logger
		logger.Info("Shutting down region-id server...")
		// Redis pool cleanup is handled by the connection pool
	})

	engine.Run(func(ctx context.Context) {
		logger := engine.Logger
		serverAddr := getEnv("SERVER_ADDR", ":8080")

		// Create REST server configuration
		cfg := &rest.Config{
			Server: serverAddr,
			IsDev:  engine.Config.IsDev,
		}

		// Create REST server with region routes
		server := rest.NewServer(cfg, logger, func(s *rest.RestServer) {
			// Register region-id routes
			regionHandler.RegisterRoutes(s)
		})

		// Start the server
		logger.Info("Starting REST server", zap.String("addr", serverAddr))
		server.Start(ctx)

		// Wait for context to be done (shutdown signal)
		<-ctx.Done()

		// Graceful shutdown
		logger.Info("Server stopped")
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
