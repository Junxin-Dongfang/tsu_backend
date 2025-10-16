-- 检查职业表
SELECT COUNT(*) as class_count FROM game_config.classes;
SELECT id, class_code, class_name, tier FROM game_config.classes LIMIT 10;
