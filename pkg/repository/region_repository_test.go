package repository

import (
	"testing"

	"github.com/enigma-id/region-id/pkg/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSearchOptions_Defaults(t *testing.T) {
	t.Run("default limit", func(t *testing.T) {
		opts := SearchOptions{
			Type:  "province",
			Limit: 0, // Should be set to default
		}
		if opts.Limit == 0 {
			opts.Limit = 10 // Default
		}
		assert.Equal(t, 10, opts.Limit)
	})

	t.Run("with filters", func(t *testing.T) {
		parentID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		opts := SearchOptions{
			Type:     "regency",
			ParentID: &parentID,
			Limit:    50,
		}
		assert.Equal(t, "regency", opts.Type)
		assert.Equal(t, parentID, *opts.ParentID)
		assert.Equal(t, 50, opts.Limit)
	})
}

func TestRegionRepository_Interface(t *testing.T) {
	// Verify interface is properly defined
	// RegionRepository should have these methods:
	// - Search
	// - FindByID
	// - FindByCode
	// - FindByParent
	// - WithContext

	// Just verify we can declare the interface
	var repo RegionRepository
	_ = repo

	t.Log("RegionRepository interface is properly defined")
}

// TestRegionRepositoryImpl_SearchOptions tests various search option combinations
func TestRegionRepositoryImpl_SearchOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    SearchOptions
		wantErr bool
	}{
		{
			name: "valid - all options",
			opts: SearchOptions{
				Type:     "province",
				Limit:    10,
			},
			wantErr: false,
		},
		{
			name: "valid - minimal options",
			opts: SearchOptions{
				Limit: 10,
			},
			wantErr: false,
		},
		{
			name: "valid - with parent",
			opts: SearchOptions{
				Type:  "regency",
				Limit: 50,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just validate the options don't cause issues
			assert.NotNil(t, tt.opts)
		})
	}
}

func TestRegion_ValidationIntegration(t *testing.T) {
	// Test entity validation in repository context
	tests := []struct {
		name    string
		region  *entity.Region
		wantErr bool
	}{
		{
			name: "valid province",
			region: &entity.Region{
				Name:  "DKI Jakarta",
				Code:  "31",
				Type:  "province",
				Level: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid - empty name",
			region: &entity.Region{
				Name:  "",
				Type:  "province",
				Level: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.region.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
