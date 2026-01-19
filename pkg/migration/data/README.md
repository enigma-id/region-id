# Indonesian Regions Data

This directory contains the Indonesian regions data for the region-id library.

## Data File

**File:** `regions_data.csv.gz` (13 MB compressed)

**Contents:**
- 91,603 regions
- 39 provinces
- 515 regencies/cities
- 7,286 districts
- 83,763 villages

**Data Source:** Exported from regions table
- Export date: January 19, 2025
- Structure: Matches the regions table schema exactly

## Data Structure

The CSV file has 12 columns (no header):

1. **id** - UUID (primary key)
2. **parent_id** - UUID (foreign key to parent region, can be NULL)
3. **name** - Region name
4. **type** - Type (province, regency, district, village)
5. **level** - Hierarchy level (1-4)
6. **code** - BPS code
7. **postal_code** - Postal code (can be empty)
8. **latitude** - Latitude coordinate (can be NULL)
9. **longitude** - Longitude coordinate (can be NULL)
10. **administrative_area** - JSONB hierarchy data
11. **created_at** - Timestamp
12. **updated_at** - Timestamp

## Import Process

The migration file `004_import_regions_data.up.sql`:

1. Disables triggers for faster import
2. Loads data directly into regions table using COPY with gunzip
3. No staging table needed - direct import
4. No transformation needed - data already matches schema
5. Re-enables triggers
6. Analyzes table for query optimization
7. Reports import statistics

## Usage

The data is automatically imported when you run migration `004_import_regions_data.up.sql`:

```bash
# After creating the regions table
psql $DATABASE_URL -f pkg/migration/001_create_regions_table.up.sql
psql $DATABASE_URL -f pkg/migration/002_create_search_function.up.sql
psql $DATABASE_URL -f pkg/migration/003_create_triggers.up.sql
psql $DATABASE_URL -f pkg/migration/004_import_regions_data.up.sql
```

Or with docker-compose:

```bash
make migrate  # Runs all migrations including data import
```

## Notes

- Data already matches the regions table structure exactly
- Includes `level` column for direct import
- Provinces have parent_id = NULL (no country level)
- Administrative area JSONB contains full hierarchy path
- Geographic coordinates included where available
- BPS codes follow official Indonesian standards
- No staging table needed - direct COPY into regions table
- No transformation needed during import

## Regenerating Data

If you need to regenerate this data from a running database:

```bash
# Export from regions table
docker exec region-id-postgres psql -U postgres -d regiondb -c "\copy ( \
  SELECT id, parent_id, name, type, code, postal_code, latitude, longitude, \
  administrative_area::text, created_at, updated_at \
  FROM regions WHERE is_deleted = false \
  ORDER BY level, type, id \
) TO '/tmp/regions_export.csv' WITH CSV HEADER;"

# Remove header and compress
docker exec region-id-postgres tail -n +2 /tmp/regions_export.csv | gzip > /tmp/regions_data.csv.gz

# Copy to project
docker cp region-id-postgres:/tmp/regions_data.csv.gz pkg/migration/data/
```
