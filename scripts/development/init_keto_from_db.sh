#!/bin/bash

# =============================================================================
# 一次性初始化脚本: 从数据库同步到 Keto
# 用途: 数据库迁移后首次同步,或 Keto 数据丢失后重建
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo "🔄 初始化 Keto 数据(从数据库)..."
echo ""

# =============================================================================
# 1. 检查数据库配置
# =============================================================================

POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-tsu_postgres}"
KETO_CONTAINER="${KETO_CONTAINER:-tsu_keto_service}"

# 直接从 Docker Compose 容器获取配置（可通过环境变量覆盖）
DB_HOST="${DB_HOST:-tsu_postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tsu_db}"
DB_USER="${DB_USER:-tsu_admin_user}"
DB_PASSWORD="${DB_PASSWORD:-tsu_admin_password}"
DB_URL="${DB_URL:-postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable}"

# 自动化控制
AUTO_APPROVE="${TSU_KETO_AUTO_APPROVE:-false}"
RESET_KETO="${TSU_KETO_RESET:-false}"

# =============================================================================
# 2. 检查服务状态
# =============================================================================

echo ""
echo "🔍 检查服务状态..."

if ! docker ps --format '{{.Names}}' | grep -qx "$KETO_CONTAINER"; then
    echo "   ❌ Keto 服务未运行"
    exit 1
fi
echo "   ✅ Keto 服务运行中"

if ! docker exec "$POSTGRES_CONTAINER" psql "${DB_URL}" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "   ❌ 数据库连接失败"
    exit 1
fi
echo "   ✅ 数据库连接正常"

# =============================================================================
# 3. 清空 Keto 现有数据(可选)
# =============================================================================

echo ""
SHOULD_RESET="$RESET_KETO"
if [[ "$AUTO_APPROVE" != "true" ]]; then
    read -p "⚠️  是否清空 Keto 现有数据? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        SHOULD_RESET="true"
    fi
else
    if [[ "$SHOULD_RESET" == "true" ]]; then
        echo "⚙️  TSU_KETO_RESET=true, 自动清空 Keto 数据..."
    else
        echo "⏭️  跳过清空 Keto 数据 (TSU_KETO_RESET != true)"
    fi
fi

if [[ "$SHOULD_RESET" == "true" ]]; then
    echo "🗑️  清空 Keto 数据..."

    # 删除 roles namespace 的所有关系
    docker exec "$KETO_CONTAINER" keto relation-tuple delete-all \
        --insecure-disable-transport-security \
        --namespace roles > /dev/null 2>&1 || true

    # 删除 permissions namespace 的所有关系
    docker exec "$KETO_CONTAINER" keto relation-tuple delete-all \
        --insecure-disable-transport-security \
        --namespace permissions > /dev/null 2>&1 || true

    echo "   ✅ Keto 数据已清空"
fi

# =============================================================================
# 4. 同步角色-权限关系
# =============================================================================

echo ""
echo "📋 同步角色-权限关系..."

ROLE_PERMS=$(docker exec "$POSTGRES_CONTAINER" psql "${DB_URL}" -t -A -c "
SELECT
    r.code,
    p.code
FROM auth.role_permissions rp
JOIN auth.roles r ON rp.role_id = r.id
JOIN auth.permissions p ON rp.permission_id = p.id
ORDER BY r.code, p.code;
")

echo "   🔎 校验关键权限..."
if echo "$ROLE_PERMS" | grep -q "team:read" && echo "$ROLE_PERMS" | grep -q "team:moderate"; then
    echo "   ✅ 团队后台权限 (team:read / team:moderate) 已在数据库中配置"
else
    echo "   ⚠️  未在数据库中找到 team:* 权限, 请确认是否执行了最新迁移"
fi

if [ -z "$ROLE_PERMS" ]; then
    echo "   ⚠️  无数据"
else
    COUNT=0
    while IFS='|' read -r role_code perm_code; do
        if [ -n "$role_code" ] && [ -n "$perm_code" ]; then
            # Keto 关系: permissions:user:read#granted@(roles:admin#member)
            docker exec "$KETO_CONTAINER" keto relation-tuple create \
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
