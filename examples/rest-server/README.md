# region-id REST Server Example

This example shows how to integrate the region-id library with the **logistics-id/engine** REST framework.

## Features

- ✅ Simple integration with logistics-id/engine
- ✅ Auto-migration support (run migrations on startup)
- ✅ Redis caching with automatic invalidation
- ✅ Graceful degradation (works without Redis)
- ✅ Structured logging with Zap
- ✅ Graceful shutdown handling
- ✅ Full-text search with filtering
- ✅ Parent-child hierarchy navigation

## Prerequisites

1. **PostgreSQL** database running
2. **Redis** (optional, for caching)
3. **Go** 1.21+ installed

## Quick Start

### 1. Set Environment Variables

```bash
# Database
export DATABASE_URL="postgres://user:password@localhost:5432/regiondb?sslmode=disable"

# Redis (optional - caching will be disabled if unavailable)
export REDIS_ADDR="localhost:6379"

# Server address
export SERVER_ADDR=":8080"
```

### 2. Create Database

```bash
createdb regiondb
```

### 3. Run the Server

The server will automatically run migrations on startup:

```bash
cd examples/rest-server
go mod tidy
go run main.go
```

Or run with auto-migration enabled (default):

```bash
# Auto-migration is enabled by default in main.go
# Set AutoMigrate: true in regionid.Config
```

The server will start on `:8080` (or the port specified in `SERVER_ADDR`).

## API Endpoints

### Search Regions

```bash
# Search by name
curl "http://localhost:8080/regions/search?q=Jakarta&limit=10"

# Filter by type
curl "http://localhost:8080/regions/search?q=Jakarta&type=province"

# Filter by parent
curl "http://localhost:8080/regions/search?parent_id={uuid}&limit=50"
```

### Get Region by ID

```bash
curl "http://localhost:8080/regions/{uuid}"
```

### Get Children of a Region

```bash
curl "http://localhost:8080/regions/{uuid}/children"
```

### Get Hierarchy Path

```bash
curl "http://localhost:8080/regions/{uuid}/path"
```

Returns the full path from root to the region:
```json
[
  {"id": "...", "name": "DKI Jakarta", "type": "province", "level": 1},
  {"id": "...", "name": "Jakarta Barat", "type": "regency", "level": 2},
  {"id": "...", "name": "Kebon Jeruk", "type": "district", "level": 3}
]
```

## Response Format

All endpoints return JSON responses in the engine format:

**Success Response:**
```json
{
  "success": true,
  "message": "Success",
  "data": {
    "id": "uuid",
    "name": "DKI Jakarta",
    "code": "31",
    "type": "province",
    "level": 1,
    "administrative_area": {
      "province": "DKI Jakarta",
      "country": "Indonesia"
    }
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Error message",
  "errors": "Detailed error"
}
```

## Code Structure

```go
// main.go
func main() {
    // 1. Setup database connection
    sqldb := sql.OpenDB(pgdriver.NewConnector(...))
    db := bun.NewDB(sqldb, pgdialect.New())

    // 2. Setup Redis (optional)
    rdb := redis.NewClient(&redis.Options{...})

    // 3. Setup logger
    logger, _ := zap.NewDevelopment()

    // 4. Initialize region-id with auto-migration
    regionHandler, err := regionid.Initialize(regionid.Config{
        DB:          db,
        Redis:       rdb,
        AutoMigrate: true,  // Runs migrations automatically
    })

    // 5. Create REST server with engine
    cfg := &rest.Config{
        Server: ":8080",
        IsDev:  true,
    }

    server := rest.NewServer(cfg, logger, func(s *rest.RestServer) {
        // Register region routes
        regionHandler.RegisterRoutes(s)
    })

    // 6. Start server
    server.Start(context.Background())
}
```

## Auto-Migration

When `AutoMigrate: true` is set in the config, the library will:

1. Create the `schema_migrations` table (if it doesn't exist)
2. Check for pending migrations
3. Run migrations in order
4. Track which migrations have been applied

This ensures your database schema is always up to date.

### Manual Migration

If you prefer to run migrations manually:

```bash
psql $DATABASE_URL -f ../../pkg/migration/001_create_regions_table.up.sql
psql $DATABASE_URL -f ../../pkg/migration/002_create_search_function.up.sql
psql $DATABASE_URL -f ../../pkg/migration/003_create_triggers.up.sql
psql $DATABASE_URL -f ../../pkg/migration/004_import_regions_data.up.sql
```

Then set `AutoMigrate: false` in your config.

## Caching

The example uses Redis for caching with the following behavior:

- **Cache enabled**: Redis connection succeeds
- **Cache disabled**: Redis connection fails or not configured
- **Cache TTL**: 24 hours for all cached data
- **Invalidation**: Automatic based on data version

### Caching Benefits

- Faster response times for repeated queries
- Reduced database load
- Automatic cache invalidation when data changes

## Docker Support

### Development with Hot Reload

```bash
# Install air first
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Production Docker Build

```bash
docker build -t region-id-server .
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e REDIS_ADDR="redis:6379" \
  region-id-server
```

## Troubleshooting

### Database Connection Failed

Ensure PostgreSQL is running and `DATABASE_URL` is correct:

```bash
# Check PostgreSQL
pg_isready

# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

### Redis Connection Failed

Redis is optional. The server will log a warning and continue without caching:

```
WARN Redis connection failed, caching will be disabled
```

### No Regions Found

Check if migrations ran successfully:

```bash
# Check if table exists
psql $DATABASE_URL -c "\dt regions"

# Check row count
psql $DATABASE_URL -c "SELECT COUNT(*) FROM regions;"
```

Should return `91603` (all Indonesian regions).

### Import Path Issues

If working locally with the library:

```bash
go mod edit -replace github.com/enigma-id/region-id=../
go mod tidy
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `REDIS_ADDR` | Redis address (host:port) | `localhost:6379` | No |
| `SERVER_ADDR` | Server bind address | `:8080` | No |
| `BUNDEBUG` | Enable query logging | - | No |

## Production Considerations

When deploying to production:

1. **Disable development mode**: Set `IsDev: false` in rest.Config
2. **Remove query logging**: Remove `bundebug` query hook
3. **Use proper logger**: Switch to production Zap config
4. **Set timeouts**: Configure database and Redis timeouts
5. **Enable HTTPS**: Use a reverse proxy (nginx, Traefik)
6. **Monitor**: Add metrics and monitoring

Example production config:

```go
cfg := &rest.Config{
    Server: ":8080",
    IsDev:  false,  // Disable dev mode
}

logger, _ := zap.NewProduction()  // Use production logger
```

## Additional Resources

- [logistics-id/engine Documentation](https://github.com/logistics-id/engine)
- [Main README](../../README.md) - Complete library documentation
- [Testing Guide](../../TESTING.md) - Test coverage information

## License

This example is licensed under the same MIT License as the main library.
