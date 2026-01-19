-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create regions table
CREATE TABLE regions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    parent_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    code TEXT,
    type TEXT NOT NULL CHECK(
        type IN ('province', 'regency', 'district', 'village')
    ),
    level INTEGER NOT NULL CHECK(
        level IN (1, 2, 3, 4)
    ),
    postal_code TEXT,
    administrative_area JSONB NOT NULL DEFAULT '{}'::jsonb,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT FALSE
);

-- Create indexes
CREATE INDEX ix_regions_parent_id ON regions(parent_id);

CREATE INDEX ix_regions_type_level ON regions(type, level);

CREATE INDEX ix_regions_type_is_deleted ON regions(type, is_deleted);

CREATE INDEX ix_regions_administrative_area ON regions USING GIN (administrative_area);

CREATE UNIQUE INDEX ux_regions_type_code ON regions(type, code) WHERE code IS NOT NULL;

CREATE UNIQUE INDEX ux_regions_code ON regions(code) WHERE code IS NOT NULL;
