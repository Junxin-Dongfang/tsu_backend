#!/bin/bash

# =============================================================================
# 一次性初始化脚本: 从数据库同步到 Keto
# 用途: 数据库迁移后首次同步,或 Keto 数据丢失后重建
# =============================================================================

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo "🔄 初始化 Keto 数据(从数据库)..."
echo ""

# =============================================================================
# 1. 检查数据库配置
# =============================================================================

# 直接从 Docker Compose 容器获取配置
DB_HOST="tsu_postgres"
DB_PORT="5432"
DB_NAME="tsu_db"
DB_USER="tsu_admin_user"
DB_PASSWORD="tsu_admin_password"
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# =============================================================================
# 2. 检查服务状态
# =============================================================================

echo ""
echo "🔍 检查服务状态..."

if ! docker ps --format '{{.Names}}' | grep -q "^tsu_keto_service$"; then
    echo "   ❌ Keto 服务未运行"
    exit 1
fi
echo "   ✅ Keto 服务运行中"

if ! docker exec tsu_postgres psql "${DB_URL}" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "   ❌ 数据库连接失败"
    exit 1
fi
echo "   ✅ 数据库连接正常"

# =============================================================================
# 3. 清空 Keto 现有数据(可选)
# =============================================================================

echo ""
read -p "⚠️  是否清空 Keto 现有数据? (y/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  清空 Keto 数据..."

    # 删除 roles namespace 的所有关系
    docker exec tsu_keto_service keto relation-tuple delete-all \
        --insecure-disable-transport-security \
        --namespace roles > /dev/null 2>&1 || true

    # 删除 permissions namespace 的所有关系
    docker exec tsu_keto_service keto relation-tuple delete-all \
        --insecure-disable-transport-security \
        --namespace permissions > /dev/null 2>&1 || true

    echo "   ✅ Keto 数据已清空"
fi

# =============================================================================
# 4. 同步角色-权限关系
# =============================================================================

echo ""
echo "📋 同步角色-权限关系..."

ROLE_PERMS=$(docker exec tsu_postgres psql "${DB_URL}" -t -A -c "
SELECT
    r.code,
    p.code
FROM auth.role_permissions rp
JOIN auth.roles r ON rp.role_id = r.id
JOIN auth.permissions p ON rp.permission_id = p.id
ORDER BY r.code, p.code;
")

if [ -z "$ROLE_PERMS" ]; then
    echo "   ⚠️  无数据"
else
    COUNT=0
    while IFS='|' read -r role_code perm_code; do
        if [ -n "$role_code" ] && [ -n "$perm_code" ]; then
            # Keto 关系: permissions:user:read#granted@(roles:admin#member)
            docker exec tsu_keto_service keto relation-tuple create \
                --insecure-disable-transport-security \
                --namespace permissions \
                --object "$perm_code" \
                --relation granted \
                --subject-set "roles:$role_code#member" > /dev/null 2>&1 || true

            COUNT=$((COUNT + 1))
            echo "   ✅ $role_code -> $perm_code"
        fi
    done <<< "$ROLE_PERMS"

    echo "   📊 完成: $COUNT 条关系"
fi

echo ""
echo "✅ 初始化完成!"
echo ""
echo "💡 后续使用 Admin API 操作会自动同步到 Keto"
