-- 000019_create_monster_config_tables.down.sql
-- 回滚怪物配置相关表

-- 删除表（按依赖关系倒序删除）

-- 1. 删除怪物掉落配置表（依赖 monsters 和 drop_pools）
DROP TABLE IF EXISTS game_config.monster_drops CASCADE;

-- 2. 删除怪物技能关联表（依赖 monsters 和 skills）
DROP TABLE IF EXISTS game_config.monster_skills CASCADE;

-- 3. 删除怪物配置主表
DROP TABLE IF EXISTS game_config.monsters CASCADE;

