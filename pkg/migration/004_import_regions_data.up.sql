-- Import Indonesian Regions Data
-- This migration imports 91,603 Indonesian regions
-- Data exported from regions table with our schema structure
--
-- CSV Columns (no header):
-- 1. id (UUID)
-- 2. parent_id (UUID, can be NULL)
-- 3. name (TEXT)
-- 4. type (TEXT) - province, regency, district, village
-- 5. level (INTEGER) - 1, 2, 3, or 4
-- 6. code (TEXT)
-- 7. postal_code (TEXT, can be empty)
-- 8. latitude (DOUBLE PRECISION, can be NULL)
-- 9. longitude (DOUBLE PRECISION, can be NULL)
-- 10. administrative_area (JSONB as text)
-- 11. created_at (TIMESTAMPTZ)
-- 12. updated_at (TIMESTAMPTZ)

-- Disable triggers for faster import
ALTER TABLE regions DISABLE TRIGGER ALL;

-- Import data directly - no staging table needed!
-- is_deleted will use its DEFAULT value of false
COPY regions (id, parent_id, name, type, level, code, postal_code, latitude, longitude, administrative_area, created_at, updated_at)
FROM PROGRAM 'gunzip -c /migration/data/regions_data.csv.gz'
WITH (FORMAT csv, DELIMITER ',', NULL '');

-- Re-enable triggers
ALTER TABLE regions ENABLE TRIGGER ALL;

-- Analyze table for query optimization
ANALYZE regions;

-- Report results
DO $$
DECLARE
    v_total INTEGER;
    v_province INTEGER;
    v_regency INTEGER;
    v_district INTEGER;
    v_village INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_total FROM regions WHERE is_deleted = false;
    SELECT COUNT(*) INTO v_province FROM regions WHERE type = 'province' AND is_deleted = false;
    SELECT COUNT(*) INTO v_regency FROM regions WHERE type = 'regency' AND is_deleted = false;
    SELECT COUNT(*) INTO v_district FROM regions WHERE type = 'district' AND is_deleted = false;
    SELECT COUNT(*) INTO v_village FROM regions WHERE type = 'village' AND is_deleted = false;

    RAISE NOTICE 'Regions imported successfully! Total: %, Provinces: %, Regencies: %, Districts: %, Villages: %',
        v_total, v_province, v_regency, v_district, v_village;
END $$;
