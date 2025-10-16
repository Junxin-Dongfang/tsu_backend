-- ============================================
-- 回滚英雄系统配置表
-- ============================================

-- 1. 移除职业技能池的初始技能字段
ALTER TABLE game_config.class_skill_pools 
DROP COLUMN IF EXISTS is_initial_skill;

-- 2. 删除英雄升级经验需求表
DROP TABLE IF EXISTS game_config.hero_level_requirements CASCADE;

-- 3. 删除属性加点消耗表
DROP TABLE IF EXISTS game_config.attribute_upgrade_costs CASCADE;

