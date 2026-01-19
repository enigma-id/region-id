package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegion_Validation(t *testing.T) {
	tests := []struct {
		name    string
		region  *Region
		wantErr bool
	}{
		{
			name: "valid region",
			region: &Region{
				ID:                 uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Name:               "DKI Jakarta",
				Code:               "31",
				Type:               "province",
				Level:              1,
				PostalCode:         "10000",
				AdministrativeArea: AdministrativeArea{Country: "Indonesia", Province: "DKI Jakarta"},
				IsDeleted:          false,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			region: &Region{
				ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Name:      "",
				Code:      "31",
				Type:      "province",
				Level:     1,
				IsDeleted: false,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			region: &Region{
				ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Name:      "DKI Jakarta",
				Code:      "31",
				Type:      "invalid",
				Level:     1,
				IsDeleted: false,
			},
			wantErr: true,
		},
		{
			name: "invalid level",
			region: &Region{
				ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Name:      "DKI Jakarta",
				Code:      "31",
				Type:      "province",
				Level:     5,
				IsDeleted: false,
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

func TestRegion_GetFullName(t *testing.T) {
	tests := []struct {
		name   string
		area   AdministrativeArea
		want   string
	}{
		{
			name: "province only",
			area: AdministrativeArea{
				Country:  "Indonesia",
				Province: "DKI Jakarta",
			},
			want: "DKI Jakarta, Indonesia",
		},
		{
			name: "with regency",
			area: AdministrativeArea{
				Country:  "Indonesia",
				Province: "DKI Jakarta",
				Regency:  "Jakarta Selatan",
			},
			want: "Jakarta Selatan, DKI Jakarta, Indonesia",
		},
		{
			name: "full hierarchy",
			area: AdministrativeArea{
				Country:  "Indonesia",
				Province: "Jawa Barat",
				Regency:  "Kota Bandung",
				District: "Cicendo",
				Village:  "Lebakgede",
			},
			want: "Lebakgede, Cicendo, Kota Bandung, Jawa Barat, Indonesia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.area.GetFullName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_IsValidType(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		want bool
	}{
		{name: "province", typ: "province", want: true},
		{name: "regency", typ: "regency", want: true},
		{name: "district", typ: "district", want: true},
		{name: "village", typ: "village", want: true},
		{name: "invalid - empty", typ: "", want: false},
		{name: "invalid - random", typ: "city", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidType(tt.typ)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_IsValidLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  bool
	}{
		{name: "level 1 - province", level: 1, want: true},
		{name: "level 2 - regency", level: 2, want: true},
		{name: "level 3 - district", level: 3, want: true},
		{name: "level 4 - village", level: 4, want: true},
		{name: "invalid - 0", level: 0, want: false},
		{name: "invalid - 5", level: 5, want: false},
		{name: "invalid - negative", level: -1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_GetTypeFromLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  string
	}{
		{name: "level 1", level: 1, want: "province"},
		{name: "level 2", level: 2, want: "regency"},
		{name: "level 3", level: 3, want: "district"},
		{name: "level 4", level: 4, want: "village"},
		{name: "invalid", level: 0, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTypeFromLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_HasParent(t *testing.T) {
	parentID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name     string
		region   *Region
		wantHas  bool
		wantUUID uuid.UUID
	}{
		{
			name: "with parent",
			region: &Region{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				ParentID: &parentID,
				Name:     "Jakarta Selatan",
				Type:     "regency",
				Level:    2,
			},
			wantHas:  true,
			wantUUID: parentID,
		},
		{
			name: "without parent (province)",
			region: &Region{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				ParentID: nil,
				Name:     "DKI Jakarta",
				Type:     "province",
				Level:    1,
			},
			wantHas:  false,
			wantUUID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasParent := tt.region.HasParent()
			assert.Equal(t, tt.wantHas, hasParent)

			if hasParent {
				require.NotNil(t, tt.region.ParentID)
				assert.Equal(t, tt.wantUUID, *tt.region.ParentID)
			}
		})
	}
}
