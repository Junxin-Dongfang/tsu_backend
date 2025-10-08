-- 回滚技能升级系统重新设计

-- 1. 删除新添加的字段
ALTER TABLE game_config.skills 
DROP COLUMN IF EXISTS level_scaling_type,
DROP COLUMN IF EXISTS level_scaling_config;

-- 2. 删除全局升级消耗表
DROP TABLE IF EXISTS game_config.skill_upgrade_costs CASCADE;

-- 3. 如果需要，可以恢复旧的 skill_level_configs 表
-- （从备份恢复或重新运行原始迁移脚本）
