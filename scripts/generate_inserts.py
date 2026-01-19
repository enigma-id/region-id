#!/usr/bin/env python3
"""Generate INSERT statements from CSV data for regions table."""

import csv
import sys
import os

def escape_sql_string(s):
    """Escape string for SQL."""
    if s is None or s == '':
        return 'NULL'
    # Replace single quotes with two single quotes
    return "'" + s.replace("'", "''") + "'"


def escape_json(json_str):
    """Escape JSON string for SQL."""
    if json_str is None or json_str == '':
        return 'NULL'
    # Replace single quotes with two single quotes
    return "'" + json_str.replace("'", "''") + "'"


def generate_inserts():
    input_file = '/tmp/regions_data.csv'
    output_file = '/tmp/regions_inserts.sql'

    print(f"Reading {input_file}...")
    regions = []
    with open(input_file, 'r') as f:
        reader = csv.reader(f, delimiter='|')
        for row in reader:
            regions.append(row)

    print(f"Found {len(regions)} regions")

    # Write output
    with open(output_file, 'w') as out:
        # Header
        out.write("-- Import Indonesian Regions Data\n")
        out.write("-- This migration imports 91,603 Indonesian regions\n")
        out.write("-- Generated from existing database\n")
        out.write("\n")
        out.write("-- Disable triggers for faster import\n")
        out.write("ALTER TABLE regions DISABLE TRIGGER ALL;\n")
        out.write("\n")

        # Generate INSERT statements with 100 rows each (reduced for compatibility)
        batch_size = 100
        current_batch = 0
        total_batches = (len(regions) + batch_size - 1) // batch_size

        for i, row in enumerate(regions):
            if i % batch_size == 0:
                if current_batch > 0:
                    out.write(";\n")
                current_batch += 1
                print(f"Generating batch {current_batch}/{total_batches}...")
                out.write("INSERT INTO regions (id, parent_id, name, type, level, code, postal_code, latitude, longitude, administrative_area, created_at, updated_at) VALUES\n")

            if i % batch_size != 0:
                out.write(",\n")

            # Parse CSV row
            (id_val, parent_id, name, type_val, level, code, postal_code,
             latitude, longitude, admin_area, created_at, updated_at) = row

            # Build VALUES clause
            values = "("
            values += f"'{id_val}', "
            values += f"{escape_sql_string(parent_id)}, "
            values += f"{escape_sql_string(name)}, "
            values += f"'{type_val}', "
            values += f"{level}, "
            values += f"{escape_sql_string(code)}, "
            values += f"{escape_sql_string(postal_code)}, "
            values += f"{latitude if latitude else 'NULL'}, "
            values += f"{longitude if longitude else 'NULL'}, "
            values += f"{escape_json(admin_area)}, "
            values += f"'{created_at}', "
            values += f"'{updated_at}'"
            values += ")"

            out.write(values)

        out.write(";\n")
        out.write("\n")

        # Footer
        out.write("-- Re-enable triggers\n")
        out.write("ALTER TABLE regions ENABLE TRIGGER ALL;\n")
        out.write("\n")
        out.write("-- Analyze table for query optimization\n")
        out.write("ANALYZE regions;\n")
        out.write("\n")

        # Report
        out.write("-- Report results\n")
        out.write("DO $$\n")
        out.write("DECLARE\n")
        out.write("    v_total INTEGER;\n")
        out.write("    v_province INTEGER;\n")
        out.write("    v_regency INTEGER;\n")
        out.write("    v_district INTEGER;\n")
        out.write("    v_village INTEGER;\n")
        out.write("BEGIN\n")
        out.write("    SELECT COUNT(*) INTO v_total FROM regions WHERE is_deleted = false;\n")
        out.write("    SELECT COUNT(*) INTO v_province FROM regions WHERE type = 'province' AND is_deleted = false;\n")
        out.write("    SELECT COUNT(*) INTO v_regency FROM regions WHERE type = 'regency' AND is_deleted = false;\n")
        out.write("    SELECT COUNT(*) INTO v_district FROM regions WHERE type = 'district' AND is_deleted = false;\n")
        out.write("    SELECT COUNT(*) INTO v_village FROM regions WHERE type = 'village' AND is_deleted = false;\n")
        out.write("\n")
        out.write("    RAISE NOTICE 'Regions imported successfully! Total: %, Provinces: %, Regencies: %, Districts: %, Villages: %',\n")
        out.write("        v_total, v_province, v_regency, v_district, v_village;\n")
        out.write("END $$;\n")

    print(f"\nGenerated INSERT statements for {len(regions)} regions")
    print(f"Output written to: {output_file}")

    # Show file size
    size = os.path.getsize(output_file)
    print(f"File size: {size / 1024 / 1024:.2f} MB")


if __name__ == '__main__':
    generate_inserts()
