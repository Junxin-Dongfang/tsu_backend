-- =============================================================================
-- Rollback Dungeon System
-- 回滚地城系统的所有变更
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 删除运行时表 (game_runtime schema)
-- --------------------------------------------------------------------------------

DROP TABLE IF EXISTS game_runtime.team_dungeon_progress CASCADE;
DROP TABLE IF EXISTS game_runtime.team_dungeon_records CASCADE;

-- --------------------------------------------------------------------------------
-- 删除配置表 (game_config schema)
-- --------------------------------------------------------------------------------

DROP TABLE IF EXISTS game_config.dungeon_events CASCADE;
DROP TABLE IF EXISTS game_config.dungeon_battles CASCADE;
DROP TABLE IF EXISTS game_config.dungeon_rooms CASCADE;
DROP TABLE IF EXISTS game_config.dungeons CASCADE;

