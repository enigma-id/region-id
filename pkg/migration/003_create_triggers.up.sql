-- Note: PostGIS extension is optional and should be enabled if geolocation features are needed
-- CREATE EXTENSION IF NOT EXISTS postgis;

-- ==============================================
-- Trigger Function: set_location()
-- Purpose: Populate location column from latitude/longitude (requires PostGIS)
-- Note: This is optional and only needed if using PostGIS features
-- ==============================================

-- Example implementation (uncomment if using PostGIS):
/*
CREATE OR REPLACE FUNCTION set_location()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.latitude IS NOT NULL AND NEW.longitude IS NOT NULL THEN
        NEW.location := ST_AsGeoJSON(ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326))::json;
    ELSE
        NEW.location := NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_location
BEFORE INSERT OR UPDATE OF latitude, longitude ON regions
FOR EACH ROW
EXECUTE FUNCTION set_location();
*/

-- ==============================================
-- Trigger Function: before_fill_administrative_area()
-- Purpose: Auto-populate administrative_area JSONB before INSERT/UPDATE
-- Builds hierarchy by walking up the parent chain
-- ==============================================

CREATE OR REPLACE FUNCTION before_fill_administrative_area()
RETURNS TRIGGER AS $$
DECLARE
    cur_parent RECORD;
    adm_area JSONB := '{}'::jsonb;
BEGIN
    cur_parent := NEW;

    -- Walk up the parent chain to build administrative area
    WHILE cur_parent.parent_id IS NOT NULL LOOP
        SELECT id, parent_id, name, type
        INTO cur_parent
        FROM regions
        WHERE id = cur_parent.parent_id
        AND is_deleted = false;

        EXIT WHEN cur_parent IS NULL;

        -- Build JSONB object with parent info
        adm_area := jsonb_build_object(
            cur_parent.type, cur_parent.name,
            cur_parent.type || '_id', cur_parent.id
        ) || adm_area;
    END LOOP;

    -- Add self to administrative area
    IF NEW.type IN ('village', 'district', 'regency', 'province') THEN
        adm_area := jsonb_build_object(
            NEW.type, NEW.name,
            NEW.type || '_id', NEW.id
        ) || adm_area;
    END IF;

    NEW.administrative_area := COALESCE(adm_area, '{}'::jsonb);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- Trigger Function: after_sync_descendants()
-- Purpose: Sync administrative_area for all descendants when parent changes
-- Uses recursive CTE to update all children, grandchildren, etc.
-- ==============================================

CREATE OR REPLACE FUNCTION after_sync_descendants()
RETURNS TRIGGER AS $$
BEGIN
    WITH RECURSIVE descendants AS (
        -- Direct children
        SELECT id FROM regions WHERE parent_id = NEW.id AND is_deleted = false
        UNION ALL
        -- All descendants through recursive join
        SELECT r.id
        FROM regions r
        INNER JOIN descendants d ON r.parent_id = d.id
        WHERE r.is_deleted = false
    )
    UPDATE regions r
    SET
        administrative_area = (
            -- Rebuild administrative_area for each descendant
            WITH RECURSIVE chain AS (
                -- Start with current descendant
                SELECT id, parent_id, name, type FROM regions WHERE id = r.id
                UNION ALL
                -- Walk up the chain
                SELECT p.id, p.parent_id, p.name, p.type
                FROM regions p
                INNER JOIN chain c ON p.id = c.parent_id
                WHERE p.is_deleted = false
            )
            SELECT COALESCE(
                jsonb_object_agg(key, val),
                '{}'::jsonb
            )
            FROM (
                -- Select name fields
                SELECT type, name FROM chain
                UNION ALL
                -- Select ID fields
                SELECT type || '_id', id::text FROM chain
            ) s(key, val)
        ),
        updated_at = NOW()
    WHERE id IN (SELECT id FROM descendants);

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- Create Triggers
-- ==============================================

-- BEFORE trigger: Fill administrative_area before insert/update
DROP TRIGGER IF EXISTS trg_before_fill_admin_area ON regions;

CREATE TRIGGER trg_before_fill_admin_area
BEFORE INSERT OR UPDATE OF name, type, parent_id
ON regions
FOR EACH ROW
EXECUTE FUNCTION before_fill_administrative_area();

-- AFTER trigger: Sync administrative_area in descendants when parent changes
DROP TRIGGER IF EXISTS trg_after_sync_descendants ON regions;

CREATE TRIGGER trg_after_sync_descendants
AFTER INSERT OR UPDATE OF name, type, parent_id
ON regions
FOR EACH ROW
WHEN (NEW.is_deleted = false)
EXECUTE FUNCTION after_sync_descendants();

-- ==============================================
-- Comments
-- ==============================================

COMMENT ON FUNCTION before_fill_administrative_area() IS '
Auto-populates administrative_area JSONB by walking up the parent chain.
Runs BEFORE INSERT or UPDATE OF name, type, parent_id.
Builds hierarchy: village -> district -> regency -> province';

COMMENT ON FUNCTION after_sync_descendants() IS '
Syncs administrative_area for all descendants when parent changes.
Uses recursive CTE to update all children recursively.
Runs AFTER INSERT or UPDATE OF name, type, parent_id, only when is_deleted = false';

COMMENT ON TRIGGER trg_before_fill_admin_area ON regions IS 'Fills administrative_area before insert/update';

COMMENT ON TRIGGER trg_after_sync_descendants ON regions IS 'Syncs administrative_area in all descendants after parent change';
