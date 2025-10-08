-- =============================================================================
-- RBAC 权限系统元数据表 (配合 Ory Keto 使用)
-- 设计理念: 数据库存储元数据(用于管理界面), Keto 存储关系(用于权限检查)
-- =============================================================================

-- =============================================================================
-- 1. 角色元数据表
-- =============================================================================

CREATE TABLE auth.roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 角色标识
    code        VARCHAR(30) NOT NULL UNIQUE,  -- 如: 'admin', 'normal_user'
    name        VARCHAR(50) NOT NULL,         -- 如: '系统管理员', '普通用户'
    description TEXT,

    -- 角色属性
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,  -- 系统角色,不可删除
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,  -- 默认角色(新用户自动分配)

    -- 时间戳
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 角色表索引
CREATE INDEX idx_roles_code ON auth.roles(code);
CREATE INDEX idx_roles_is_system ON auth.roles(is_system);
CREATE INDEX idx_roles_is_default ON auth.roles(is_default);

-- 角色表注释
COMMENT ON TABLE auth.roles IS 'RBAC角色元数据表,存储角色的业务信息,实际用户-角色关系存储在Keto中';
COMMENT ON COLUMN auth.roles.code IS '角色代码,用于Keto relation tuples';
COMMENT ON COLUMN auth.roles.is_system IS '系统角色不可通过API删除';
COMMENT ON COLUMN auth.roles.is_default IS '新用户注册时自动分配此角色';

-- =============================================================================
-- 2. 权限分组表 (用于管理界面组织权限)
-- =============================================================================

CREATE TABLE auth.permission_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 分组标识
    code        VARCHAR(50) NOT NULL UNIQUE,   -- 如: 'user_management', 'system_settings'
    name        VARCHAR(100) NOT NULL,         -- 如: '用户管理', '系统设置'
    description TEXT,

    -- 显示属性
    icon        VARCHAR(100),                   -- 图标名称
    color       VARCHAR(7),                     -- 十六进制颜色值 (#FFFFFF)
    sort_order  INTEGER NOT NULL DEFAULT 0,    -- 排序权重

    -- 层级结构 (可选)
    parent_id   UUID REFERENCES auth.permission_groups(id),
    level       INTEGER NOT NULL DEFAULT 1,

    -- 时间戳
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 权限分组表索引
CREATE INDEX idx_permission_groups_code ON auth.permission_groups(code);
CREATE INDEX idx_permission_groups_parent_id ON auth.permission_groups(parent_id);
CREATE INDEX idx_permission_groups_level ON auth.permission_groups(level);
CREATE INDEX idx_permission_groups_sort_order ON auth.permission_groups(sort_order);

-- 权限分组表注释
COMMENT ON TABLE auth.permission_groups IS '权限分组表,用于管理界面组织和展示权限';
COMMENT ON COLUMN auth.permission_groups.sort_order IS '排序权重,数值越小越靠前';
COMMENT ON COLUMN auth.permission_groups.level IS '分组层级,用于树形结构展示';

-- =============================================================================
-- 3. 权限元数据表
-- =============================================================================

CREATE TABLE auth.permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 权限标识
    code        VARCHAR(100) NOT NULL UNIQUE,  -- 如: 'user:create', 'role:manage'
    name        VARCHAR(100) NOT NULL,         -- 如: '创建用户', '管理角色'
    description TEXT,

    -- 权限分类
    resource    VARCHAR(50) NOT NULL,          -- 资源类型: 'user', 'role', 'hero'
    action      VARCHAR(50) NOT NULL,          -- 操作类型: 'create', 'read', 'update', 'delete', 'manage'

    -- 权限属性
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,  -- 系统权限,不可删除

    -- 时间戳
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(resource, action)  -- 确保资源和操作的组合唯一
);

-- 权限表索引
CREATE INDEX idx_permissions_code ON auth.permissions(code);
CREATE INDEX idx_permissions_resource ON auth.permissions(resource);
CREATE INDEX idx_permissions_action ON auth.permissions(action);
CREATE INDEX idx_permissions_resource_action ON auth.permissions(resource, action);
CREATE INDEX idx_permissions_is_system ON auth.permissions(is_system);

-- 权限表注释
COMMENT ON TABLE auth.permissions IS 'RBAC权限元数据表,存储权限的业务信息,实际权限关系存储在Keto中';
COMMENT ON COLUMN auth.permissions.code IS '权限代码,用于Keto relation tuples,格式: resource:action';
COMMENT ON COLUMN auth.permissions.resource IS '权限所属资源类型';
COMMENT ON COLUMN auth.permissions.action IS '权限允许的操作类型';

-- =============================================================================
-- 4. 权限-分组关联表
-- =============================================================================

CREATE TABLE auth.permission_group_members (
    id             BIGSERIAL PRIMARY KEY,
    group_id       UUID NOT NULL REFERENCES auth.permission_groups(id) ON DELETE CASCADE,
    permission_id  UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    sort_order     INTEGER NOT NULL DEFAULT 0,  -- 权限在分组内的排序

    UNIQUE(group_id, permission_id)
);

-- 权限分组关联索引
CREATE INDEX idx_permission_group_members_group_id ON auth.permission_group_members(group_id);
CREATE INDEX idx_permission_group_members_permission_id ON auth.permission_group_members(permission_id);
CREATE INDEX idx_permission_group_members_sort_order ON auth.permission_group_members(sort_order);

-- 权限分组关联注释
COMMENT ON TABLE auth.permission_group_members IS '权限与分组的关联表,一个权限可以属于多个分组';

-- =============================================================================
-- 5. 角色-权限关联表 (用于管理界面快速查询)
-- =============================================================================

CREATE TABLE auth.role_permissions (
    role_id       UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,

    -- 审计信息
    granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by    UUID REFERENCES auth.users(id),

    PRIMARY KEY (role_id, permission_id)
);

-- 角色权限关联索引
CREATE INDEX idx_role_permissions_role_id ON auth.role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON auth.role_permissions(permission_id);

-- 角色权限关联注释
COMMENT ON TABLE auth.role_permissions IS '角色-权限关联表,用于管理界面展示,实际权限检查使用Keto';
COMMENT ON COLUMN auth.role_permissions.granted_by IS '授权操作者,用于审计';

-- =============================================================================
-- 触发器 - 自动更新 updated_at
-- =============================================================================

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON auth.roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE ON auth.permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permission_groups_updated_at
    BEFORE UPDATE ON auth.permission_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
