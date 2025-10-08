-- =============================================================================
-- 回滚 RBAC 权限系统元数据表
-- =============================================================================

-- 删除触发器
DROP TRIGGER IF EXISTS update_permission_groups_updated_at ON auth.permission_groups;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON auth.permissions;
DROP TRIGGER IF EXISTS update_roles_updated_at ON auth.roles;

-- 删除表 (按依赖关系逆序)
DROP TABLE IF EXISTS auth.role_permissions;
DROP TABLE IF EXISTS auth.permission_group_members;
DROP TABLE IF EXISTS auth.permissions;
DROP TABLE IF EXISTS auth.permission_groups;
DROP TABLE IF EXISTS auth.roles;
