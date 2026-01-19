// Package migration provides database migration functionality for the region-id library.
//
// The migration package includes:
// - Embedded SQL migration files (using Go embed)
// - Migrator for running migrations programmatically
// - Automatic migration tracking via schema_migrations table
// - Support for both up and down migrations
//
// Migration files follow the naming pattern: XXX_description.(up|down).sql
// where XXX is a 3-digit sequence number.
package migration

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/uptrace/bun"
)

//go:embed *.sql
var Migrations embed.FS

// Migrator handles database migrations using embedded SQL files.
//
// The migrator tracks applied migrations in the schema_migrations table
// and only runs pending migrations on Up().
type Migrator struct {
	db *bun.DB
}

// NewMigrator creates a new migrator with the given database connection.
func NewMigrator(db *bun.DB) *Migrator {
	return &Migrator{db: db}
}

// Up runs all pending migrations in ascending order.
//
// Migrations are tracked in the schema_migrations table.
// Only .up.sql files that haven't been applied yet will be executed.
func (m *Migrator) Up(ctx context.Context) error {
	// Create migrations tracking table if it doesn't exist
	if err := m.createSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range files {
		// Skip if already applied
		if _, ok := applied[file]; ok {
			fmt.Printf("Skipping already applied migration: %s\n", file)
			continue
		}

		// Only run .up.sql files
		if !strings.HasSuffix(file, ".up.sql") {
			continue
		}

		fmt.Printf("Running migration: %s\n", file)

		if err := m.runMigration(ctx, file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		fmt.Printf("Migration completed: %s\n", file)
	}

	return nil
}

// createSchemaMigrationsTable creates the migrations tracking table
func (m *Migrator) createSchemaMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getMigrationFiles returns sorted list of migration files
func (m *Migrator) getMigrationFiles() ([]string, error) {
	var files []string

	entries, err := Migrations.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}

	// Sort files to ensure correct order
	sort.Strings(files)

	return files, nil
}

// getAppliedMigrations returns map of applied migrations
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations;`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// runMigration executes a single migration file
func (m *Migrator) runMigration(ctx context.Context, filename string) error {
	// Read migration file content
	content, err := Migrations.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	_, err = m.db.ExecContext(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	query := `INSERT INTO schema_migrations (version) VALUES (?);`
	_, err = m.db.ExecContext(ctx, query, filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// Down rolls back the most recently applied migration.
//
// It finds the last applied .up.sql migration and runs the corresponding .down.sql file.
// Returns an error if no migration can be rolled back.
func (m *Migrator) Down(ctx context.Context) error {
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Find last applied migration
	var lastMigration string
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if strings.HasSuffix(file, ".up.sql") && applied[file] {
			lastMigration = file
			break
		}
	}

	if lastMigration == "" {
		return fmt.Errorf("no migration to rollback")
	}

	// Find corresponding .down.sql file
	downFile := strings.Replace(lastMigration, ".up.sql", ".down.sql", 1)

	// Check if down file exists
	var exists bool
	for _, file := range files {
		if file == downFile {
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("no rollback file found for %s", lastMigration)
	}

	fmt.Printf("Rolling back migration: %s\n", lastMigration)

	// Read and execute down migration
	content, err := Migrations.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("failed to read rollback file: %w", err)
	}

	_, err = m.db.ExecContext(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove from migrations table
	query := `DELETE FROM schema_migrations WHERE version = ?;`
	_, err = m.db.ExecContext(ctx, query, lastMigration)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("Rollback completed: %s\n", lastMigration)

	return nil
}

// RunMigrationsFromFilesystem runs migrations from a directory path.
//
// This is useful for development or when not using embedded migration files.
// It reads .up.sql files from the specified directory and runs pending migrations.
//
// The migrationDir should contain SQL files following the naming pattern:
// XXX_description.up.sql where XXX is a 3-digit sequence number.
func RunMigrationsFromFilesystem(ctx context.Context, db *bun.DB, migrationDir string) error {
	// Create migrations tracking table
	m := &Migrator{db: db}
	if err := m.createSchemaMigrationsTable(ctx); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Read migration files from directory
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}

	// Sort files
	sort.Strings(files)

	// Run pending migrations
	for _, file := range files {
		if _, ok := applied[file]; ok {
			fmt.Printf("Skipping: %s\n", file)
			continue
		}

		fmt.Printf("Running: %s\n", file)

		content, err := os.ReadFile(filepath.Join(migrationDir, file))
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute %s: %w", file, err)
		}

		// Record migration
		query := `INSERT INTO schema_migrations (version) VALUES ($1);`
		if _, err := db.ExecContext(ctx, query, file); err != nil {
			return fmt.Errorf("failed to record %s: %w", file, err)
		}
	}

	return nil
}
