-- Drop indexes
DROP INDEX IF EXISTS ux_regions_code;
DROP INDEX IF EXISTS ux_regions_type_code;
DROP INDEX IF EXISTS ix_regions_administrative_area;
DROP INDEX IF EXISTS ix_regions_type_is_deleted;
DROP INDEX IF EXISTS ix_regions_type_level;
DROP INDEX IF EXISTS ix_regions_parent_id;

-- Drop regions table
DROP TABLE IF EXISTS regions;

-- Drop UUID extension (optional, only if no other tables use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
