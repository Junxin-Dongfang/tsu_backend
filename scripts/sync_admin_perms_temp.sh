#!/bin/bash
set -e

echo "同步 admin 角色权限到 Keto..."

docker exec tsu_postgres psql "postgres://tsu_admin_user:tsu_admin_password@tsu_postgres:5432/tsu_db?sslmode=disable" -t -A -c "
SELECT p.code
FROM auth.role_permissions rp
JOIN auth.roles r ON rp.role_id = r.id
JOIN auth.permissions p ON rp.permission_id = p.id
WHERE r.code = 'admin'
ORDER BY p.code;
" | while read perm_code; do
  if [ -n "$perm_code" ]; then
    echo "{\"namespace\":\"permissions\",\"object\":\"$perm_code\",\"relation\":\"granted\",\"subject_set\":{\"namespace\":\"roles\",\"object\":\"admin\",\"relation\":\"member\"}}" | \
    docker exec -i tsu_keto_service keto relation-tuple create --insecure-disable-transport-security - > /dev/null 2>&1 && \
    echo "✅ admin -> $perm_code"
  fi
done

echo ""
echo "同步 normal_user 角色权限到 Keto..."

docker exec tsu_postgres psql "postgres://tsu_admin_user:tsu_admin_password@tsu_postgres:5432/tsu_db?sslmode=disable" -t -A -c "
SELECT p.code
FROM auth.role_permissions rp
JOIN auth.roles r ON rp.role_id = r.id
JOIN auth.permissions p ON rp.permission_id = p.id
WHERE r.code = 'normal_user'
ORDER BY p.code;
" | while read perm_code; do
  if [ -n "$perm_code" ]; then
    echo "{\"namespace\":\"permissions\",\"object\":\"$perm_code\",\"relation\":\"granted\",\"subject_set\":{\"namespace\":\"roles\",\"object\":\"normal_user\",\"relation\":\"member\"}}" | \
    docker exec -i tsu_keto_service keto relation-tuple create --insecure-disable-transport-security - > /dev/null 2>&1 && \
    echo "✅ normal_user -> $perm_code"
  fi
done

echo ""
echo "✅ 完成!"
