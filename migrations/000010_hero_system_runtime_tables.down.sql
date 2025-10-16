-- ============================================
-- 回滚英雄系统运行时表
-- ============================================

-- 1. 删除属性计算视图
DROP VIEW IF EXISTS game_runtime.hero_computed_attributes;

-- 2. 删除英雄已分配属性表
DROP TABLE IF EXISTS game_runtime.hero_allocated_attributes CASCADE;

-- 3. 移除 hero_skills 表的新增字段
ALTER TABLE IF EXISTS game_runtime.hero_skills
DROP COLUMN IF EXISTS first_learned_at;

-- 4. 删除技能操作历史表
DROP TABLE IF EXISTS game_runtime.hero_skill_operations CASCADE;

-- 5. 删除属性操作历史表
DROP TABLE IF EXISTS game_runtime.hero_attribute_operations CASCADE;

-- 6. 删除英雄职业历史表
DROP TABLE IF EXISTS game_runtime.hero_class_history CASCADE;

