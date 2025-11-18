-- 移除团队后台权限
DELETE FROM auth.permission_group_members
WHERE permission_id IN (
    SELECT id FROM auth.permissions WHERE code IN ('team:read', 'team:moderate')
);

DELETE FROM auth.role_permissions
WHERE permission_id IN (
    SELECT id FROM auth.permissions WHERE code IN ('team:read', 'team:moderate')
);

DELETE FROM auth.permissions
WHERE code IN ('team:read', 'team:moderate');
