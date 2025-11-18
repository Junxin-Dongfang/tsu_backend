-- =============================================================================
-- Rollback Team System Tables
-- 回滚团队系统表
-- =============================================================================

-- 删除表（按依赖关系逆序删除）

-- 8. 删除战利品分配历史表
DROP TABLE IF EXISTS game_runtime.team_loot_distribution_history CASCADE;

-- 7. 删除团队仓库物品表
DROP TABLE IF EXISTS game_runtime.team_warehouse_items CASCADE;

-- 6. 删除团队仓库表
DROP TABLE IF EXISTS game_runtime.team_warehouses CASCADE;

-- 5. 删除团队踢出记录表
DROP TABLE IF EXISTS game_runtime.team_kicked_records CASCADE;

-- 4. 删除团队邀请表
DROP TABLE IF EXISTS game_runtime.team_invitations CASCADE;

-- 3. 删除团队加入申请表
DROP TABLE IF EXISTS game_runtime.team_join_requests CASCADE;

-- 2. 删除团队成员表
DROP TABLE IF EXISTS game_runtime.team_members CASCADE;

-- 1. 删除团队表
DROP TABLE IF EXISTS game_runtime.teams CASCADE;

