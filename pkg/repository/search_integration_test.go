package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/enigma-id/region-id/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// TestSearchIntegration tests the Search method with real database connection
func TestSearchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get database URL from env or use default
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable"
	}

	// Setup database connection
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(databaseURL),
	))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Create repository without cache for testing
	repo := NewRegionRepository(db, nil)

	ctx := context.Background()

	t.Run("search karang tengah kota tangerang", func(t *testing.T) {
		results, err := repo.Search(ctx, "Karang Tengah Kota Tangerang", SearchOptions{
			Limit: 5,
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 5, "Should not return more than limit results (possible CROSS JOIN bug)")
		assert.NotEmpty(t, results, "Should find results")

		// Check that we found Karang Tengah district in Kota Tangerang
		var karangTengahDistrict *entity.Region
		for _, r := range results {
			if r.Name == "Karang Tengah" && r.Type == "district" {
				karangTengahDistrict = r
				break
			}
		}
		assert.NotNil(t, karangTengahDistrict, "Should find Karang Tengah district")

		if karangTengahDistrict != nil {
			// Verify administrative_area contains Kota Tangerang
			adminArea := karangTengahDistrict.AdministrativeArea
			assert.Equal(t, "Banten", adminArea.Province, "Province should be Banten")
			assert.Equal(t, "Kota Tangerang", adminArea.Regency, "Regency should be Kota Tangerang")

			t.Logf("✅ Found: %s (district) in %s, %s",
				karangTengahDistrict.Name,
				adminArea.Regency,
				adminArea.Province)
		}
	})

	t.Run("search banten", func(t *testing.T) {
		results, err := repo.Search(ctx, "banten", SearchOptions{
			Limit: 3,
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 3, "Should not return more than limit results (possible CROSS JOIN bug)")

		// Should find Banten province
		var bantenProvince *entity.Region
		for _, r := range results {
			if r.Name == "Banten" && r.Type == "province" {
				bantenProvince = r
				break
			}
		}
		assert.NotNil(t, bantenProvince, "Should find Banten province")

		if bantenProvince != nil {
			t.Logf("✅ Found: %s (%s)", bantenProvince.Name, bantenProvince.Type)
		}
	})

	t.Run("search with parent filter (fallback to query builder)", func(t *testing.T) {
		// Find a district first to get its parent (regency)
		districts, err := repo.Search(ctx, "Karang Tengah", SearchOptions{
			Type: "district",
			Limit: 1,
		})
		require.NoError(t, err)
		require.NotEmpty(t, districts, "Should find Karang Tengah district")

		parentID := districts[0].ParentID

		// Search siblings within the same parent using parent filter
		siblings, err := repo.Search(ctx, "", SearchOptions{
			ParentID: parentID,
			Limit:    10,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, siblings, "Should find sibling regions")

		t.Logf("✅ Found %d sibling regions using parent filter", len(siblings))
	})
}
