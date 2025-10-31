-- Add set_id column to items table for equipment set association
-- Migration: 000018_add_set_id_to_items

-- Add set_id column to game_config.items table
ALTER TABLE game_config.items 
ADD COLUMN set_id UUID REFERENCES game_config.equipment_set_configs(id);

-- Create index for efficient set-based queries
CREATE INDEX idx_items_set_id ON game_config.items(set_id);

-- Add comment to explain the column
COMMENT ON COLUMN game_config.items.set_id IS '装备所属套装ID，关联到equipment_set_configs表';

