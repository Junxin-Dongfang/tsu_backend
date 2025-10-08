-- =============================================================================
-- RBAC 系统初始数据
-- 创建基础角色、权限分组、权限和关联关系
-- =============================================================================

-- =============================================================================
-- 1. 插入基础角色
-- =============================================================================

INSERT INTO auth.roles (code, name, description, is_system, is_default) VALUES
('admin', '系统管理员', '拥有系统所有权限,可以管理用户、角色、权限', true, false),
('normal_user', '普通用户', '基础用户权限,可以使用游戏核心功能', true, true);

-- =============================================================================
-- 2. 插入权限分组
-- =============================================================================

INSERT INTO auth.permission_groups (code, name, description, icon, color, sort_order, level) VALUES
('user_management', '用户管理', '用户相关的权限', 'user', '#1890FF', 1, 1),
('role_management', '角色管理', '角色和权限相关的权限', 'safety', '#52C41A', 2, 1),
('permission_management', '权限管理', '权限配置相关的权限', 'lock', '#FA8C16', 3, 1),
('system_management', '系统管理', '系统配置和监控相关的权限', 'setting', '#722ED1', 4, 1),
('game_config', '游戏配置', '游戏策划配置相关的权限', 'appstore', '#13C2C2', 5, 1);

-- =============================================================================
-- 3. 插入权限定义
-- =============================================================================

-- 用户管理相关权限
INSERT INTO auth.permissions (code, name, description, resource, action, is_system) VALUES
('user:read', '查看用户', '查看用户列表和详情', 'user', 'read', true),
('user:create', '创建用户', '创建新用户账号', 'user', 'create', true),
('user:update', '更新用户', '修改用户信息', 'user', 'update', true),
('user:delete', '删除用户', '删除用户账号', 'user', 'delete', true),
('user:ban', '封禁用户', '封禁/解封用户', 'user', 'ban', true);

-- 角色管理相关权限
INSERT INTO auth.permissions (code, name, description, resource, action, is_system) VALUES
('role:read', '查看角色', '查看角色列表和详情', 'role', 'read', true),
('role:create', '创建角色', '创建新角色', 'role', 'create', true),
('role:update', '更新角色', '修改角色信息', 'role', 'update', true),
('role:delete', '删除角色', '删除角色', 'role', 'delete', true),
('role:assign', '分配角色', '为用户分配/撤销角色', 'role', 'assign', true);

-- 权限管理相关权限
INSERT INTO auth.permissions (code, name, description, resource, action, is_system) VALUES
('permission:read', '查看权限', '查看权限列表和详情', 'permission', 'read', true),
('permission:assign', '分配权限', '为角色分配/撤销权限', 'permission', 'assign', true),
('permission:grant_user', '直接授权用户', '直接授予/撤销用户权限(绕过角色)', 'permission', 'grant_user', true);

-- 系统管理相关权限
INSERT INTO auth.permissions (code, name, description, resource, action, is_system) VALUES
('system:config', '系统配置', '修改系统配置', 'system', 'config', true),
('system:monitor', '系统监控', '查看系统运行状态和日志', 'system', 'monitor', true);

-- 游戏配置相关权限
INSERT INTO auth.permissions (code, name, description, resource, action, is_system) VALUES
('hero:manage', '管理英雄配置', '管理游戏英雄配置数据', 'hero', 'manage', true),
('skill:manage', '管理技能配置', '管理游戏技能配置数据', 'skill', 'manage', true),
('class:manage', '管理职业配置', '管理游戏职业配置数据', 'class', 'manage', true);

-- =============================================================================
-- 4. 关联权限到分组
-- =============================================================================

-- 用户管理分组的权限
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT
    (SELECT id FROM auth.permission_groups WHERE code = 'user_management'),
    p.id,
    ROW_NUMBER() OVER (ORDER BY p.code) - 1
FROM auth.permissions p
WHERE p.resource = 'user';

-- 角色管理分组的权限
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT
    (SELECT id FROM auth.permission_groups WHERE code = 'role_management'),
    p.id,
    ROW_NUMBER() OVER (ORDER BY p.code) - 1
FROM auth.permissions p
WHERE p.resource = 'role';

-- 权限管理分组的权限
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT
    (SELECT id FROM auth.permission_groups WHERE code = 'permission_management'),
    p.id,
    ROW_NUMBER() OVER (ORDER BY p.code) - 1
FROM auth.permissions p
WHERE p.resource = 'permission';

-- 系统管理分组的权限
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT
    (SELECT id FROM auth.permission_groups WHERE code = 'system_management'),
    p.id,
    ROW_NUMBER() OVER (ORDER BY p.code) - 1
FROM auth.permissions p
WHERE p.resource = 'system';

-- 游戏配置分组的权限
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT
    (SELECT id FROM auth.permission_groups WHERE code = 'game_config'),
    p.id,
    ROW_NUMBER() OVER (ORDER BY p.code) - 1
FROM auth.permissions p
WHERE p.resource IN ('hero', 'skill', 'class');

-- =============================================================================
-- 5. 为角色分配权限
-- =============================================================================

-- admin 角色拥有所有权限
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM auth.roles WHERE code = 'admin'),
    p.id
FROM auth.permissions p;

-- normal_user 角色只有基础查看权限
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM auth.roles WHERE code = 'normal_user'),
    p.id
FROM auth.permissions p
WHERE p.code IN ('user:read');  -- 普通用户只能查看用户列表(如果需要)

-- =============================================================================
-- 数据验证
-- =============================================================================

-- 验证角色数量
DO $$
DECLARE
    role_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO role_count FROM auth.roles;
    IF role_count != 2 THEN
        RAISE EXCEPTION '角色插入失败: 期望2个角色,实际%个', role_count;
    END IF;
END $$;

-- 验证权限数量
DO $$
DECLARE
    permission_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO permission_count FROM auth.permissions;
    IF permission_count != 18 THEN
        RAISE EXCEPTION '权限插入失败: 期望18个权限,实际%个', permission_count;
    END IF;
END $$;

-- 验证 admin 角色权限数量
DO $$
DECLARE
    admin_permission_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO admin_permission_count
    FROM auth.role_permissions
    WHERE role_id = (SELECT id FROM auth.roles WHERE code = 'admin');

    IF admin_permission_count != 18 THEN
        RAISE EXCEPTION 'admin角色权限分配失败: 期望18个权限,实际%个', admin_permission_count;
    END IF;
END $$;
