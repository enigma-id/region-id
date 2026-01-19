package entity

import "strings"

// AdministrativeArea contains hierarchical region information for display purposes.
//
// This struct stores the full names of each administrative level in the hierarchy.
// It's useful for displaying complete location information (e.g., "Kebon Jeruk, Jakarta Barat, DKI Jakarta, Indonesia").
//
// Both names and IDs are stored for flexibility in different use cases.
type AdministrativeArea struct {
	Country    string `json:"country,omitempty"`
	Province   string `json:"province,omitempty"`
	Regency    string `json:"regency,omitempty"`
	District   string `json:"district,omitempty"`
	Village    string `json:"village,omitempty"`
	CountryID  string `json:"country_id,omitempty"`
	ProvinceID string `json:"province_id,omitempty"`
	RegencyID  string `json:"regency_id,omitempty"`
	DistrictID string `json:"district_id,omitempty"`
	VillageID  string `json:"village_id,omitempty"`
}

// GetFullName returns the full hierarchical name as a comma-separated string.
//
// The returned string includes all non-empty levels from village to country,
// formatted as: "Village, District, Regency, Province, Country"
//
// Example output: "Kebon Jeruk, Jakarta Barat, DKI Jakarta, Indonesia"
//
// Empty levels are automatically excluded from the output.
func (a *AdministrativeArea) GetFullName() string {
	parts := []string{}

	if a.Village != "" {
		parts = append(parts, a.Village)
	}
	if a.District != "" {
		parts = append(parts, a.District)
	}
	if a.Regency != "" {
		parts = append(parts, a.Regency)
	}
	if a.Province != "" {
		parts = append(parts, a.Province)
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}

	return strings.Join(parts, ", ")
}
