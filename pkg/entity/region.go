// Package entity defines domain entities for the region-id library.
//
// The entity package contains the core data models used throughout the library:
// - Region: Represents an Indonesian administrative region (province, regency, district, village)
// - AdministrativeArea: Hierarchical area information with full names
//
// All entities include validation methods to ensure data integrity.
package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Region represents an Indonesian administrative region.
//
// Region types and levels:
//   - Level 1: Province (type="province")
//   - Level 2: Regency/City (type="regency")
//   - Level 3: District (type="district")
//   - Level 4: Village (type="village")
//
// Each region (except provinces) has a parent region, forming a hierarchy.
type Region struct {
	ID                 uuid.UUID              `bun:"type:uuid,default:uuid_generate_v4(),pk" json:"id"`
	ParentID           *uuid.UUID             `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	Name               string                 `bun:"name,notnull" json:"name"`
	Code               string                 `bun:"code" json:"code"`
	Type               string                 `bun:"type,notnull" json:"type"` // province, regency, district, village
	Level              int                    `bun:"level,notnull" json:"level"` // 1-4
	PostalCode         string                 `bun:"postal_code" json:"postal_code,omitempty"`
	AdministrativeArea AdministrativeArea     `bun:"administrative_area,type:jsonb,notnull" json:"administrative_area"`
	Latitude           *float64               `bun:"latitude" json:"latitude,omitempty"`
	Longitude          *float64               `bun:"longitude" json:"longitude,omitempty"`
	CreatedAt          time.Time              `bun:"created_at,nullzero" json:"created_at,omitempty"`
	UpdatedAt          time.Time              `bun:"updated_at,nullzero" json:"updated_at,omitempty"`
	IsDeleted          bool                   `bun:"is_deleted,notnull" json:"is_deleted"`
}

func (Region) TableName() string {
	return "regions"
}

// Validate checks if the region data is valid.
//
// It validates:
// - Name is not empty
// - Type is one of: province, regency, district, village
// - Level is between 1-4
// - Type matches the level (e.g., level 1 must be type "province")
//
// Returns an error describing the validation failure, or nil if valid.
func (r *Region) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if !IsValidType(r.Type) {
		return fmt.Errorf("invalid type: %s", r.Type)
	}

	if !IsValidLevel(r.Level) {
		return fmt.Errorf("invalid level: %d", r.Level)
	}

	// Validate type and level match
	expectedType := GetTypeFromLevel(r.Level)
	if expectedType != "" && r.Type != expectedType {
		return fmt.Errorf("type %s doesn't match level %d (expected %s)", r.Type, r.Level, expectedType)
	}

	return nil
}

// HasParent returns true if the region has a parent region.
//
// Provinces (level 1) do not have parents. All other region types
// should have a parent ID pointing to their containing region.
func (r *Region) HasParent() bool {
	return r.ParentID != nil && *r.ParentID != uuid.Nil
}

// IsValidType checks if the given region type is valid.
//
// Valid types are: province, regency, district, village
func IsValidType(typ string) bool {
	switch typ {
	case "province", "regency", "district", "village":
		return true
	default:
		return false
	}
}

// IsValidLevel checks if the given level is valid.
//
// Valid levels are 1-4, corresponding to:
// - 1: Province
// - 2: Regency
// - 3: District
// - 4: Village
func IsValidLevel(level int) bool {
	return level >= 1 && level <= 4
}

// GetTypeFromLevel returns the expected region type for a given level.
//
// Returns empty string for invalid levels.
// Level to type mapping:
// - 1 → "province"
// - 2 → "regency"
// - 3 → "district"
// - 4 → "village"
func GetTypeFromLevel(level int) string {
	switch level {
	case 1:
		return "province"
	case 2:
		return "regency"
	case 3:
		return "district"
	case 4:
		return "village"
	default:
		return ""
	}
}
