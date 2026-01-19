-- Rollback: Remove all imported regions data
-- WARNING: This will delete all regions data!

DELETE FROM regions
WHERE is_deleted = false;
