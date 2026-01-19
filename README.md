<div align="center">

# region-id

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**A comprehensive, searchable database of Indonesian administrative regions for Go**

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Examples](#examples)

</div>

---

## Overview

**region-id** is a production-ready Go library that provides complete access to Indonesian administrative regions (provinces, regencies, districts, and villages). Built with performance and simplicity in mind, it offers full-text search, Redis caching, and seamless integration with any Go web framework.

### What's Included

- **91,603 regions** - Complete Indonesian administrative hierarchy
- **39 provinces** • **515 regencies** • **7,286 districts** • **83,763 villages**
- **Full-text search** with fuzzy matching and ranking
- **Redis caching** with automatic invalidation (3x faster queries)
- **Auto-migration** - One-command database setup
- **Clean architecture** - Repository pattern, domain entities, handlers
- **Framework-agnostic** - Works with Echo, Gin, Chi, or logistics-id/engine

---

## Features

### 🔍 Full-Text Search
- Fuzzy matching using PostgreSQL trigrams
- Smart ranking based on match quality
- Filter by type, parent, or name
- Search across name, code, and administrative area

### ⚡ Performance
- **Redis caching** with 24-hour TTL
- **Singleflight** pattern to prevent cache stampede
- **Optimized indexes** for fast queries
- **Version-based cache** invalidation
- **3x faster** cached responses (~10ms vs ~30ms)

### 🛠️ Developer Experience
- **Auto-migration** - Run migrations programmatically
- **Embedded data** - No external data files needed
- **Graceful degradation** - Works without Redis
- **Type-safe** entities and repository interfaces
- **Comprehensive tests** - Unit tests with high coverage

### 🏗️ Architecture
- **Clean architecture** with separation of concerns
- **Repository pattern** for data access
- **Domain entities** with validation
- **HTTP handlers** with logistics-id/engine integration
- **Migration system** with version tracking

---

## Quick Start

### Installation

```bash
go get github.com/enigma-id/region-id
```

### 1. Start Dependencies

```bash
# Using Docker Compose
docker-compose up -d postgres redis

# Or use your own PostgreSQL + Redis instances
```

### 2. Initialize the Library

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/enigma-id/region-id"
    "github.com/redis/go-redis/v9"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/pgdialect"
    "github.com/uptrace/bun/driver/pgdriver"
    "github.com/logistics-id/engine/transport/rest"
    "go.uber.org/zap"
)

func main() {
    // Setup database
    sqldb := sql.OpenDB(pgdriver.NewConnector(
        pgdriver.WithDSN("postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable"),
    ))
    db := bun.NewDB(sqldb, pgdialect.New())

    // Setup Redis (optional - for caching)
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Setup logger
    logger, _ := zap.NewDevelopment()

    // Initialize with auto-migration
    regionHandler, err := regionid.Initialize(regionid.Config{
        DB:          db,
        Redis:       rdb,
        AutoMigrate: true, // Automatically runs migrations!
    })
    if err != nil {
        logger.Fatal("Failed to initialize", zap.Error(err))
    }

    // Create REST server
    cfg := &rest.Config{
        Server: ":8080",
        IsDev:  true,
    }

    server := rest.NewServer(cfg, logger, func(s *rest.RestServer) {
        regionHandler.RegisterRoutes(s) // Register region routes
    })

    server.Start(context.Background())
}
```

### 3. Run and Test

```bash
# Run your server
go run main.go

# Test the API
curl "http://localhost:8080/regions/search?q=Jakarta&limit=5"
```

**That's it!** Your database now has 91,603 Indonesian regions ready to search.

---

## Documentation

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/regions/search` | Search regions by name with filters |
| `GET` | `/regions/:id` | Get region by ID |
| `GET` | `/regions/:id/children` | Get direct children of a region |
| `GET` | `/regions/:id/path` | Get full hierarchy path |

#### Search Regions

```bash
GET /regions/search?q=Jakarta&type=province&limit=10
```

**Query Parameters:**
- `q` - Search query (region name, case-insensitive)
- `type` - Filter by type (`province`, `regency`, `district`, `village`)
- `parent_id` - Filter by parent UUID
- `limit` - Max results (default: 10)

**Response:**

```json
{
  "success": true,
  "message": "success",
  "data": [
    {
      "id": "551131ea-5323-4b5e-a8fb-a085430f01ca",
      "parent_id": null,
      "name": "DKI Jakarta",
      "code": "31",
      "type": "province",
      "level": 1,
      "administrative_area": {
        "country": "Indonesia",
        "province": "DKI Jakarta",
        "country_id": "d92a0534-6dd6-47d0-a865-04d9efaacbc3",
        "province_id": "551131ea-5323-4b5e-a8fb-a085430f01ca"
      },
      "latitude": -6.1990,
      "longitude": 106.8343,
      "created_at": "2025-08-19T09:05:59.85979Z",
      "is_deleted": false
    }
  ]
}
```

#### Get Region by ID

```bash
GET /regions/:id
```

Returns detailed information for a specific region.

#### Get Children

```bash
GET /regions/:id/children
```

Returns all direct children (one level down) ordered by name.

#### Get Hierarchy Path

```bash
GET /regions/:id/path
```

Returns the full path from root (province) to the specified region:

```json
{
  "success": true,
  "message": "success",
  "data": [
    {"name": "DKI Jakarta", "type": "province", "level": 1},
    {"name": "Jakarta Selatan", "type": "regency", "level": 2},
    {"name": "Tebet", "type": "district", "level": 3}
  ]
}
```

### Configuration

```go
config := regionid.Config{
    DB:          db,           // Required: *bun.DB database connection
    Redis:       rdb,          // Optional: *redis.Client (nil = no caching)
    AutoMigrate: true,         // Optional: Run migrations on startup (default: false)
}
```

#### Without Redis (No Caching)

```go
regionHandler, err := regionid.Initialize(regionid.Config{
    DB:    db,
    Redis: nil,  // Caching disabled - queries go directly to database
})
```

#### Manual Migration

If you prefer to run migrations manually:

```bash
# Run migrations before starting your app
psql $DATABASE_URL -f pkg/migration/001_create_regions_table.up.sql
psql $DATABASE_URL -f pkg/migration/002_create_search_function.up.sql
psql $DATABASE_URL -f pkg/migration/003_create_triggers.up.sql
psql $DATABASE_URL -f pkg/migration/004_import_regions_data.up.sql
```

Then initialize without auto-migration:

```go
regionHandler, err := regionid.Initialize(regionid.Config{
    DB:          db,
    Redis:       rdb,
    AutoMigrate: false,  // Migrations already run manually
})
```

### Entity Structure

```go
type Region struct {
    ID                 uuid.UUID              // Unique identifier
    ParentID           *uuid.UUID             // Parent region ID
    Name               string                 // Region name
    Code               string                 // BPS code
    Type               string                 // province, regency, district, village
    Level              int                    // Administrative level (1-4)
    PostalCode         string                 // Postal code
    AdministrativeArea AdministrativeArea     // Hierarchy information
    Latitude           *float64               // Geographic coordinates
    Longitude          *float64
    CreatedAt          time.Time
    UpdatedAt          time.Time
    IsDeleted          bool                   // Soft delete flag
}
```

**Administrative Levels:**
- **Level 1**: Province (provinsi)
- **Level 2**: Regency/City (kabupaten/kota)
- **Level 3**: District (kecamatan)
- **Level 4**: Village (desa/kelurahan)

---

## Examples

A complete REST API server example using **logistics-id/engine** is available in the [examples/rest-server](./examples/rest-server/) directory.

### Running the Example

```bash
cd examples/rest-server

# Set environment variables
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export SERVER_ADDR=":8080"

# Run with auto-migration
go run main.go
```

The example demonstrates:
- ✅ Auto-migration on startup
- ✅ Redis caching configuration
- ✅ Structured logging with Zap
- ✅ Graceful shutdown
- ✅ Health check endpoint

### Testing the Example

```bash
# Search regions
curl "http://localhost:8080/regions/search?q=Bandung&limit=3"

# Filter by type
curl "http://localhost:8080/regions/search?q=Jawa&type=province"

# Get by ID (use ID from search results)
curl "http://localhost:8080/regions/{uuid}"
```

---

## Performance

Benchmarks from production usage:

| Operation | Uncached | Cached | Improvement |
|-----------|----------|--------|-------------|
| Search query | ~30ms | ~10ms | **3x faster** |
| Get by ID | ~5ms | ~2ms | **2.5x faster** |
| Get children | ~15ms | ~5ms | **3x faster** |

**Cache Configuration:**
- Search results: 24-hour TTL
- Individual regions: 24-hour TTL
- Data version: Persistent (no expiration)
- Automatic invalidation on data changes

**Throughput:**
- Supports 1000+ concurrent requests
- Cache hit rate: 90%+ for common queries
- Database connections managed via Bun connection pool

---

## Data Source

### Coverage

Complete Indonesian regions data exported and validated:

- **Total**: 91,603 regions
- **Provinces**: 39 (level 1)
- **Regencies/Cities**: 515 (level 2)
- **Districts**: 7,286 (level 3)
- **Villages**: 83,763 (level 4)

### Data Quality

- **Source**: Official Indonesian administrative boundaries
- **Export Date**: January 2025
- **BPS Codes**: Official Indonesian statistics codes
- **Validation**: All entities validated before import
- **Hierarchy**: Complete parent-child relationships
- **Geolocation**: Latitude/longitude coordinates included

### Import Method

Data is embedded in migration files and imported using PostgreSQL `COPY` command:
- **No staging table** - Direct import for speed
- **No transformation** - Data matches schema exactly
- **Fast import** - ~4 seconds for 91,603 regions
- **Embedded** - No external files or downloads needed

---

## Architecture

```
region-id/
├── regionid.go              # Main library initialization
├── pkg/
│   ├── entity/              # Domain entities
│   │   ├── region.go        # Region entity with validation
│   │   └── administrative_area.go
│   ├── repository/          # Data access layer
│   │   ├── region_repository.go      # Interface definition
│   │   ├── region_repository_impl.go  # Implementation with caching
│   │   ├── cache_manager.go           # Redis cache management
│   │   └── version_tracker.go         # Data version tracking
│   ├── handler/             # HTTP handlers
│   │   ├── region_handler.go          # Route registration
│   │   ├── search_request.go          # Search DTO
│   │   ├── get_region_request.go      # Get by ID DTO
│   │   ├── get_children_request.go    # Get children DTO
│   │   └── get_path_request.go        # Get path DTO
│   └── migration/           # Database migrations
│       ├── migrator.go              # Migrator implementation
│       ├── 001_create_regions_table.up.sql
│       ├── 002_create_search_function.up.sql
│       ├── 003_create_triggers.up.sql
│       └── 004_import_regions_data.up.sql  # Embedded 91,603 regions
└── examples/
    └── rest-server/         # logistics-id/engine example
        ├── main.go
        ├── README.md
        └── Dockerfile
```

### Design Patterns

- **Repository Pattern** - Clean data access abstraction
- **Dependency Injection** - Pass DB and Redis as dependencies
- **Singleflight** - Prevent cache stampede on concurrent requests
- **Version-based Caching** - Automatic cache invalidation
- **Migration Tracking** - Version-controlled schema changes

---

## Testing

The library includes comprehensive unit tests:

```bash
# Run all tests
GOWORK=off go test ./pkg/... -v

# Run with coverage
GOWORK=off go test ./pkg/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

- ✅ Entity validation tests
- ✅ Repository interface tests
- ✅ Migration system tests
- ✅ Cache manager tests

See [TESTING.md](TESTING.md) for details.

---

## Docker Support

### Docker Compose (Development)

```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: regiondb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - ./pkg/migration:/docker-entrypoint-initdb.d:ro

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
```

```bash
docker-compose up -d
```

### Production Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/enigma-id/region-id.git
cd region-id

# Run tests
GOWORK=off go test ./pkg/...

# Run example
cd examples/rest-server
go mod tidy
go run main.go
```

### Coding Standards

- Follow Go conventions and effective Go guidelines
- Write tests for new features
- Update documentation for API changes
- Use meaningful commit messages

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- **Indonesian Government** - For the official administrative region data
- **BPS (Badan Pusat Statistik)** - For the region codes and standards
- **logistics-id/engine** - For the excellent REST framework

---

<div align="center">

**Built with ❤️ for the Go community**

[⬆ Back to Top](#region-id)

</div>
