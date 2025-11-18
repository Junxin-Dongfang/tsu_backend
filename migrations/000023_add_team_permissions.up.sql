-- 新增团队后台权限
WITH new_permissions AS (
    INSERT INTO auth.permissions (code, name, description, resource, action, is_system)
    VALUES
        ('team:read', '查看团队(后台)', '允许后台查询团队信息', 'team', 'read', true),
        ('team:moderate', '管理团队(后台)', '允许后台执行解散、调整等操作', 'team', 'moderate', true)
    ON CONFLICT (code) DO NOTHING
    RETURNING id, code
)
INSERT INTO auth.role_permissions (role_id, permission_id)
SELECT r.id, np.id
FROM auth.roles r
JOIN new_permissions np ON 1=1
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 将新权限归入 system_management 分组,便于后台展示
INSERT INTO auth.permission_group_members (group_id, permission_id, sort_order)
SELECT pg.id, p.id, 0
FROM auth.permission_groups pg
JOIN auth.permissions p ON p.code IN ('team:read', 'team:moderate')
WHERE pg.code = 'system_management'
ON CONFLICT (group_id, permission_id) DO NOTHING;
