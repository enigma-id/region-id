package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enigma-id/region-id"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"

	"github.com/logistics-id/engine/transport/rest"
)

func main() {
	// Load environment variables
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	serverAddr := getEnv("SERVER_ADDR", ":8080")

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Setup database connection
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(databaseURL),
	))

	db := bun.NewDB(sqldb, pgdialect.New())

	// Add query logger for debugging (remove in production)
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	// Setup Redis (optional - if not available, caching will be disabled)
	var rdb *redis.Client
	if redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		// Test Redis connection
		ctx := context.Background()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis connection failed, caching will be disabled", zap.Error(err))
			rdb = nil
		} else {
			logger.Info("Redis connected successfully - caching enabled")
		}
	}

	// Initialize region-id library with auto-migration
	logger.Info("Initializing region-id library...")
	regionHandler, err := regionid.Initialize(regionid.Config{
		DB:          db,
		Redis:       rdb,
		AutoMigrate: true, // Automatically run migrations on startup
	})
	if err != nil {
		logger.Fatal("Failed to initialize region-id library", zap.Error(err))
	}
	logger.Info("region-id library initialized successfully")

	// Create REST server configuration
	cfg := &rest.Config{
		Server: serverAddr,
		IsDev:  true,
	}

	// Create REST server with region routes
	server := rest.NewServer(cfg, logger, func(s *rest.RestServer) {
		// Register region-id routes
		regionHandler.RegisterRoutes(s)
	})

	// Start the server
	logger.Info("Starting REST server", zap.String("addr", serverAddr))
	server.Start(context.Background())

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	logger.Info("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
	logger.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
