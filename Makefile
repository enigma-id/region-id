.PHONY: help up down restart logs ps db redis clean status

# Default target
help:
	@echo "region-id Library - Docker Compose Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Infrastructure:"
	@echo "  up              Start PostgreSQL and Redis"
	@echo "  down            Stop all services"
	@echo "  restart         Restart all services"
	@echo "  logs            Show logs from all services"
	@echo "  logs-db         Show logs from PostgreSQL"
	@echo "  logs-redis      Show logs from Redis"
	@echo "  ps              Show running containers"
	@echo "  db              Open psql shell in PostgreSQL"
	@echo "  redis           Open redis-cli shell"
	@echo "  status          Show service status"
	@echo "  clean           Remove all containers and volumes"
	@echo ""
	@echo "Database:"
	@echo "  migrate         Run database migrations manually"
	@echo "  import-data     Import regions data (91,603 regions)"
	@echo ""
	@echo "API Server:"
	@echo "  api             Run API server locally (requires Go)"
	@echo "  api-dev         Run API server with hot reload"

# Start PostgreSQL and Redis
up:
	@echo "Starting PostgreSQL and Redis..."
	docker-compose up -d
	@echo ""
	@echo "Services started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis:      localhost:6380"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run migrations: make migrate"
	@echo "  2. Import data: make import-data"
	@echo "  3. Or use AutoMigrate: true in your code"

# Stop services
down:
	@echo "Stopping all services..."
	docker-compose down

# Restart services
restart:
	@echo "Restarting all services..."
	docker-compose restart
	@echo "Services restarted!"

# Show logs from all services
logs:
	docker-compose logs -f

# Show logs from PostgreSQL
logs-db:
	docker-compose logs -f postgres

# Show logs from Redis
logs-redis:
	docker-compose logs -f redis

# Show running containers
ps:
	docker-compose ps

# Open psql shell in PostgreSQL
db:
	docker-compose exec postgres psql -U postgres -d regiondb

# Open redis-cli shell
redis:
	docker-compose exec redis redis-cli

# Run migrations manually
migrate:
	@echo "Running database migrations..."
	@docker-compose exec -T postgres psql -U postgres -d regiondb < pkg/migration/001_create_regions_table.up.sql
	@docker-compose exec -T postgres psql -U postgres -d regiondb < pkg/migration/002_create_search_function.up.sql
	@docker-compose exec -T postgres psql -U postgres -d regiondb < pkg/migration/003_create_triggers.up.sql
	@echo "Migrations completed!"
	@echo ""
	@echo "Now import data with: make import-data"

# Import regions data
import-data:
	@echo "Importing 91,603 regions..."
	@docker-compose exec -T postgres psql -U postgres -d regiondb < pkg/migration/004_import_regions_data.up.sql
	@echo ""
	@echo "Data import completed!"
	@echo "Check with: make db"
	@echo "Then run: SELECT type, COUNT(*) FROM regions WHERE is_deleted = false GROUP BY type;"

# Run API server locally (requires Go)
api:
	@echo "Starting API server..."
	@export DATABASE_URL="postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable" && \
	export REDIS_ADDR="localhost:6380" && \
	export SERVER_ADDR=":8080" && \
	cd examples/simple-server && \
	go run main.go

# Run API server with hot reload
api-dev:
	@echo "Starting API server with hot reload..."
	@export DATABASE_URL="postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable" && \
	export REDIS_ADDR="localhost:6380" && \
	export SERVER_ADDR=":8080" && \
	cd examples/simple-server && \
	air

# Show service status
status:
	@echo "=== Container Status ==="
	@docker-compose ps
	@echo ""
	@echo "=== PostgreSQL Status ==="
	@docker-compose exec -T postgres pg_isready -U postgres 2>/dev/null || echo "PostgreSQL not ready"
	@echo ""
	@echo "=== Redis Status ==="
	@docker-compose exec -T redis redis-cli ping 2>/dev/null || echo "Redis not ready"
	@echo ""
	@echo "=== Database Info ==="
	@-docker-compose exec -T postgres psql -U postgres -d regiondb -c "SELECT COUNT(*) as region_count FROM regions WHERE is_deleted = false;" 2>/dev/null || echo "No data yet"

# Clean up everything
clean:
	@echo "Removing all containers and volumes..."
	docker-compose down -v
	@echo "Cleanup complete!"

# Full setup (manual migration)
setup: up
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Running migrations..."
	make migrate
	@echo ""
	@echo "Importing data..."
	make import-data
	@echo ""
	@echo "Setup complete!"
	@echo ""
	@echo "✅ Database ready with 91,603 regions!"
	@echo ""
	@echo "Start API server:"
	@echo "  make api"
