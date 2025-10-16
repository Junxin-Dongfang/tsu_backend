-- Insert test class data
INSERT INTO game_config.classes (id, class_code, class_name, tier, description, is_active, is_visible, display_order, created_at, updated_at)
VALUES 
    (uuid_generate_v4(), 'warrior', 'Warrior', 'basic', 'Melee physical class with strong defense', true, true, 1, NOW(), NOW()),
    (uuid_generate_v4(), 'mage', 'Mage', 'basic', 'Ranged magic class with powerful spells', true, true, 2, NOW(), NOW()),
    (uuid_generate_v4(), 'archer', 'Archer', 'basic', 'Ranged physical class with high mobility', true, true, 3, NOW(), NOW()),
    (uuid_generate_v4(), 'priest', 'Priest', 'basic', 'Support class with healing abilities', true, true, 4, NOW(), NOW())
ON CONFLICT (class_code) DO NOTHING;

-- Show inserted classes
SELECT id, class_code, class_name, tier, display_order FROM game_config.classes ORDER BY display_order;
