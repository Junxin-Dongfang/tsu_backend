-- 000020_add_missing_attribute_types.down.sql
-- 回滚：删除添加的属性类型

DELETE FROM game_config.hero_attribute_type WHERE attribute_code IN (
    'INITIATIVE',
    'BODY_RESIST',
    'MAGIC_RESIST',
    'MENTAL_RESIST',
    'ENVIRONMENT_RESIST'
);

