-- 回滚技能升级系统重新设计

-- 1. 删除 skill_unlock_actions 表新添加的字段
ALTER TABLE game_config.skill_unlock_actions 
DROP COLUMN IF EXISTS level_scaling_config;
