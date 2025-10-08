-- =============================================================================
-- 回滚 RBAC 系统初始数据
-- =============================================================================

-- 删除关联数据 (按依赖关系逆序)
DELETE FROM auth.role_permissions;
DELETE FROM auth.permission_group_members;
DELETE FROM auth.permissions;
DELETE FROM auth.permission_groups;
DELETE FROM auth.roles;
