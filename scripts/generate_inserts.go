package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Region struct {
	ID                 string  `bun:"id,type:uuid"`
	ParentID           *string `bun:"parent_id,type:uuid"`
	Name               string  `bun:"name"`
	Type               string  `bun:"type"`
	Level              int     `bun:"level"`
	Code               string  `bun:"code"`
	PostalCode         string  `bun:"postal_code"`
	Latitude           *float64 `bun:"latitude"`
	Longitude          *float64 `bun:"longitude"`
	AdministrativeArea string  `bun:"administrative_area,type:jsonb"`
	CreatedAt          string  `bun:"created_at"`
	UpdatedAt          string  `bun:"updated_at"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/regiondb?sslmode=disable"
	}

	db := bun.NewDB(
		pgdriver.NewConnector(pgdriver.WithDSN(dbURL)),
		pgdialect.New(),
	)

	var regions []Region
	err := db.NewSelect().Model(&regions).
		Where("is_deleted = false").
		Order("level ASC", "name ASC").
		Scan(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying regions: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Found %d regions\n", len(regions))

	// Create output file
	outFile, err := os.Create("/tmp/regions_inserts.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write header
	fmt.Fprintln(outFile, "-- Import Indonesian Regions Data")
	fmt.Fprintln(outFile, "-- This migration imports 91,603 Indonesian regions")
	fmt.Fprintln(outFile, "-- Generated from existing database")
	fmt.Fprintln(outFile, "")
	fmt.Fprintln(outFile, "-- Disable triggers for faster import")
	fmt.Fprintln(outFile, "ALTER TABLE regions DISABLE TRIGGER ALL;")
	fmt.Fprintln(outFile, "")

	// Generate INSERT statements with 1000 rows each
	batchSize := 1000
	currentBatch := 0
	totalBatches := (len(regions) + batchSize - 1) / batchSize

	for i, region := range regions {
		if i%batchSize == 0 {
			if currentBatch > 0 {
				fmt.Fprintln(outFile, ";")
			}
			currentBatch++
			fmt.Fprintf(os.Stderr, "Generating batch %d/%d...\n", currentBatch, totalBatches)
			fmt.Fprintf(outFile, "INSERT INTO regions (id, parent_id, name, type, level, code, postal_code, latitude, longitude, administrative_area, created_at, updated_at) VALUES\n")
		} else {
			fmt.Fprintln(outFile, ",")
		}

		// Build VALUES clause
		values := buildValuesClause(region)
		fmt.Fprint(outFile, values)
	}

	fmt.Fprintln(outFile, ";")
	fmt.Fprintln(outFile, "")
	fmt.Fprintln(outFile, "-- Re-enable triggers")
	fmt.Fprintln(outFile, "ALTER TABLE regions ENABLE TRIGGER ALL;")
	fmt.Fprintln(outFile, "")
	fmt.Fprintln(outFile, "-- Analyze table for query optimization")
	fmt.Fprintln(outFile, "ANALYZE regions;")
	fmt.Fprintln(outFile, "")

	// Write report
	fmt.Fprintln(outFile, "-- Report results")
	fmt.Fprintln(outFile, "DO $$")
	fmt.Fprintln(outFile, "DECLARE")
	fmt.Fprintln(outFile, "    v_total INTEGER;")
	fmt.Fprintln(outFile, "    v_province INTEGER;")
	fmt.Fprintln(outFile, "    v_regency INTEGER;")
	fmt.Fprintln(outFile, "    v_district INTEGER;")
	fmt.Fprintln(outFile, "    v_village INTEGER;")
	fmt.Fprintln(outFile, "BEGIN")
	fmt.Fprintln(outFile, "    SELECT COUNT(*) INTO v_total FROM regions WHERE is_deleted = false;")
	fmt.Fprintln(outFile, "    SELECT COUNT(*) INTO v_province FROM regions WHERE type = 'province' AND is_deleted = false;")
	fmt.Fprintln(outFile, "    SELECT COUNT(*) INTO v_regency FROM regions WHERE type = 'regency' AND is_deleted = false;")
	fmt.Fprintln(outFile, "    SELECT COUNT(*) INTO v_district FROM regions WHERE type = 'district' AND is_deleted = false;")
	fmt.Fprintln(outFile, "    SELECT COUNT(*) INTO v_village FROM regions WHERE type = 'village' AND is_deleted = false;")
	fmt.Fprintln(outFile, "")
	fmt.Fprintln(outFile, "    RAISE NOTICE 'Regions imported successfully! Total: %, Provinces: %, Regencies: %, Districts: %, Villages: %',")
	fmt.Fprintln(outFile, "        v_total, v_province, v_regency, v_district, v_village;")
	fmt.Fprintln(outFile, "END $$;")

	fmt.Fprintf(os.Stderr, "\nGenerated INSERT statements for %d regions\n", len(regions))
	fmt.Fprintf(os.Stderr, "Output written to: /tmp/regions_inserts.sql\n")
}

func buildValuesClause(r Region) string {
	var sb strings.Builder
	sb.WriteString("(")

	// id
	sb.WriteString("'")
	sb.WriteString(r.ID)
	sb.WriteString("', ")

	// parent_id
	if r.ParentID != nil {
		sb.WriteString("'")
		sb.WriteString(*r.ParentID)
		sb.WriteString("', ")
	} else {
		sb.WriteString("NULL, ")
	}

	// name - escape single quotes
	sb.WriteString("'")
	sb.WriteString(strings.ReplaceAll(r.Name, "'", "''"))
	sb.WriteString("', ")

	// type
	sb.WriteString("'")
	sb.WriteString(r.Type)
	sb.WriteString("', ")

	// level
	sb.WriteString(fmt.Sprintf("%d, ", r.Level))

	// code
	sb.WriteString("'")
	sb.WriteString(strings.ReplaceAll(r.Code, "'", "''"))
	sb.WriteString("', ")

	// postal_code
	if r.PostalCode != "" {
		sb.WriteString("'")
		sb.WriteString(strings.ReplaceAll(r.PostalCode, "'", "''"))
		sb.WriteString("', ")
	} else {
		sb.WriteString("NULL, ")
	}

	// latitude
	if r.Latitude != nil {
		sb.WriteString(fmt.Sprintf("%f, ", *r.Latitude))
	} else {
		sb.WriteString("NULL, ")
	}

	// longitude
	if r.Longitude != nil {
		sb.WriteString(fmt.Sprintf("%f, ", *r.Longitude))
	} else {
		sb.WriteString("NULL, ")
	}

	// administrative_area - need to escape properly for JSON
	adminArea := strings.ReplaceAll(r.AdministrativeArea, "'", "''")
	sb.WriteString("'")
	sb.WriteString(adminArea)
	sb.WriteString("', ")

	// created_at
	sb.WriteString("'")
	sb.WriteString(r.CreatedAt)
	sb.WriteString("', ")

	// updated_at
	if r.UpdatedAt != "" {
		sb.WriteString("'")
		sb.WriteString(r.UpdatedAt)
		sb.WriteString("'")
	} else {
		sb.WriteString("NULL")
	}

	sb.WriteString(")")

	return sb.String()
}
