-- Enable pg_trgm extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ==============================================
-- Search Regions Function
-- ==============================================
-- Purpose: Full-text search with ranking based on match quality
-- Parameters:
--   query: search query text
--   type_filter: array of types to filter (e.g., ARRAY['province', 'regency'])
--   limit: maximum number of results
-- Returns: Ordered regions with rank based on match quality
-- ==============================================

CREATE OR REPLACE FUNCTION search_regions(
    query TEXT,
    type_filter TEXT[] DEFAULT NULL,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE (
    id UUID,
    parent_id UUID,
    name TEXT,
    code TEXT,
    type TEXT,
    level INTEGER,
    postal_code TEXT,
    administrative_area JSONB,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    is_deleted BOOLEAN,
    rank DOUBLE PRECISION
) AS $$
DECLARE
    normalized_query TEXT;
BEGIN
    -- Normalize query: lowercase and trim
    normalized_query := LOWER(TRIM(query));

    -- Return empty result if query is empty
    IF normalized_query IS NULL OR normalized_query = '' THEN
        RETURN;
    END IF;

    RETURN QUERY
    WITH search_results AS (
        SELECT
            r.id,
            r.parent_id,
            r.name,
            r.code,
            r.type,
            r.level,
            r.postal_code,
            r.administrative_area,
            r.latitude,
            r.longitude,
            r.created_at,
            r.updated_at,
            r.is_deleted,
            -- Calculate rank based on match quality
            CASE
                -- Exact match (highest priority)
                WHEN LOWER(r.name) = normalized_query THEN 1000
                -- Starts with query (high priority)
                WHEN LOWER(r.name) LIKE normalized_query || '%' THEN 900
                -- Contains query word by word (medium-high priority)
                WHEN LOWER(r.name) LIKE '%' || normalized_query || '%' THEN 700
                -- Fuzzy match using similarity (medium priority)
                WHEN SIMILARITY(LOWER(r.name), normalized_query) > 0.5 THEN 600 + (SIMILARITY(LOWER(r.name), normalized_query) * 100)
                -- Code match (if code is provided)
                WHEN LOWER(r.code) = normalized_query THEN 950
                WHEN LOWER(r.code) LIKE normalized_query || '%' THEN 850
                -- Low priority for partial matches
                ELSE 100
            END +
            -- Boost based on administrative level (higher level = higher boost)
            CASE
                WHEN r.level = 1 THEN 50  -- Province
                WHEN r.level = 2 THEN 40  -- Regency
                WHEN r.level = 3 THEN 30  -- District
                WHEN r.level = 4 THEN 20  -- Village
                ELSE 0
            END AS rank
        FROM regions r
        WHERE
            r.is_deleted = false
            AND (
                -- Type filter: if provided, only match specified types
                type_filter IS NULL
                OR r.type = ANY(type_filter)
            )
            AND (
                -- Search in name
                LOWER(r.name) LIKE '%' || normalized_query || '%'
                -- Search in code
                OR LOWER(r.code) LIKE '%' || normalized_query || '%'
                -- Search in postal code
                OR LOWER(r.postal_code) LIKE '%' || normalized_query || '%'
                -- Search in administrative area JSONB
                OR EXISTS (
                    SELECT 1
                    FROM jsonb_each_text(r.administrative_area) AS t(key, value)
                    WHERE LOWER(t.value) LIKE '%' || normalized_query || '%'
                )
                -- Fuzzy match using trigram similarity
                OR SIMILARITY(LOWER(r.name), normalized_query) > 0.3
            )
    )
    SELECT
        sr.id,
        sr.parent_id,
        sr.name,
        sr.code,
        sr.type,
        sr.level,
        sr.postal_code,
        sr.administrative_area,
        sr.latitude,
        sr.longitude,
        sr.created_at,
        sr.updated_at,
        sr.is_deleted,
        sr.rank
    FROM search_results sr
    ORDER BY sr.rank DESC, sr.name ASC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Create index for better search performance
CREATE INDEX IF NOT EXISTS ix_regions_name_trgm ON regions USING GIN (LOWER(name) gin_trgm_ops);

CREATE INDEX IF NOT EXISTS ix_regions_code_trgm ON regions USING GIN (LOWER(code) gin_trgm_ops) WHERE code IS NOT NULL;

-- Comment on function
COMMENT ON FUNCTION search_regions(TEXT, TEXT[], INTEGER) IS '
Full-text search for regions with ranking based on match quality.
Parameters:
  - query: search query text (searches in name, code, postal_code, administrative_area)
  - type_filter: optional array of types to filter (NULL = all types)
  - limit_count: maximum number of results (default: 10)

Returns regions ordered by:
  1. Match quality (exact > prefix > contains > fuzzy)
  2. Administrative level (province > regency > district > village)
  3. Name (alphabetical)
';
