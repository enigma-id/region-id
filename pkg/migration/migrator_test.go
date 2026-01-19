package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMigrator(t *testing.T) {
	t.Run("creates migrator with nil db", func(t *testing.T) {
		migrator := NewMigrator(nil)
		assert.NotNil(t, migrator)
	})
}

func TestMigrator_Up_NoDB(t *testing.T) {
	t.Run("fails gracefully without database", func(t *testing.T) {
		migrator := NewMigrator(nil)
		// Should panic with nil DB - that's expected behavior
		assert.Panics(t, func() {
			migrator.Up(nil)
		})
	})
}

func TestMigrator_Down_NoDB(t *testing.T) {
	t.Run("fails gracefully without database", func(t *testing.T) {
		migrator := NewMigrator(nil)
		// Should panic with nil DB - that's expected behavior
		assert.Panics(t, func() {
			migrator.Down(nil)
		})
	})
}

func TestMigrations_EmbeddedFiles(t *testing.T) {
	t.Run("migration files are embedded", func(t *testing.T) {
		// Check that migrations FS has content
		assert.NotNil(t, Migrations)

		// List files in the embed
		entries, err := Migrations.ReadDir(".")
		assert.NoError(t, err)
		assert.NotEmpty(t, entries)

		// Count SQL files
		sqlCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".sql" {
				sqlCount++
			}
		}

		assert.GreaterOrEqual(t, sqlCount, 8, "Expected at least 8 migration files (4 up, 4 down)")
	})
}

func TestMigrationNaming(t *testing.T) {
	t.Run("migration files follow naming pattern", func(t *testing.T) {
		entries, err := Migrations.ReadDir(".")
		assert.NoError(t, err)

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Check format: XXX_description.(up|down).sql
			assert.Regexp(t, `^\d{3}`, name, "Migration should start with 3-digit number")
			assert.Contains(t, name, ".", "Migration should have extension")
			assert.Regexp(t, `.(up|down).sql$`, name, "Migration should end with .up.sql or .down.sql")

			t.Logf("✓ Migration file: %s", name)
		}
	})
}

func TestMigrationPairs(t *testing.T) {
	t.Run("each up migration has corresponding down", func(t *testing.T) {
		entries, err := Migrations.ReadDir(".")
		assert.NoError(t, err)

		upMigrations := make(map[string]bool)
		downMigrations := make(map[string]bool)

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()

			// Extract base name by removing .up.sql or .down.sql
			var base string
			if len(name) > 7 && name[len(name)-7:] == ".up.sql" {
				base = name[:len(name)-7] // Remove .up.sql
				upMigrations[base] = true
			} else if len(name) > 9 && name[len(name)-9:] == ".down.sql" {
				base = name[:len(name)-9] // Remove .down.sql
				downMigrations[base] = true
			}
		}

		// Check that all up migrations have corresponding down
		for base := range upMigrations {
			assert.True(t, downMigrations[base], "Up migration %s should have corresponding down migration", base)
		}

		// Check that all down migrations have corresponding up
		for base := range downMigrations {
			assert.True(t, upMigrations[base], "Down migration %s should have corresponding up migration", base)
		}

		t.Logf("Found %d migration pairs", len(upMigrations))
	})
}
