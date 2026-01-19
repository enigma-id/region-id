-- Drop triggers
DROP TRIGGER IF EXISTS trg_after_sync_descendants ON regions;
DROP TRIGGER IF EXISTS trg_before_fill_admin_area ON regions;

-- Drop PostGIS location trigger if it was created
DROP TRIGGER IF EXISTS trigger_set_location ON regions;

-- Drop trigger functions
DROP FUNCTION IF EXISTS after_sync_descendants();
DROP FUNCTION IF EXISTS before_fill_administrative_area();

-- Drop PostGIS set_location function if it was created
DROP FUNCTION IF EXISTS set_location();

-- Note: PostGIS extension removal is optional
-- DROP EXTENSION IF EXISTS postgis;
