-- Rollback: Remove set_id column from items table
-- Migration: 000018_add_set_id_to_items

-- Drop index
DROP INDEX IF EXISTS game_config.idx_items_set_id;

-- Drop set_id column
ALTER TABLE game_config.items 
DROP COLUMN IF EXISTS set_id;

