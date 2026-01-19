-- Drop search function
DROP FUNCTION IF EXISTS search_regions(TEXT, TEXT[], INTEGER);

-- Drop trigram indexes
DROP INDEX IF EXISTS ix_regions_code_trgm;
DROP INDEX IF EXISTS ix_regions_name_trgm;

-- Drop pg_trgm extension (optional, only if no other tables use it)
-- DROP EXTENSION IF EXISTS pg_trgm;
