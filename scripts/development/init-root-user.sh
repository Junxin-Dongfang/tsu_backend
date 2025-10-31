#!/bin/bash

# ==========================================
# Root 用户初始化脚本
# ==========================================
# 功能：
#   1. 在 Kratos 中创建 root identity
#   2. 在 auth.users 表中同步用户信息
# 
# 默认账号：
#   用户名: root
#   密码: password
#   邮箱: root@tsu-game.com

set -e

# 配置
KRATOS_ADMIN_URL="http://localhost:4434"
DB_HOST="localhost"
DB_PORT="5432"

# 从环境变量读取数据库配置
DB_USER="${DB_USER:-tsu_user}"
DB_PASSWORD="${DB_PASSWORD}"
DB_NAME="${DB_NAME:-tsu_db}"

# Root 用户信息
ROOT_USERNAME="root"
ROOT_EMAIL="root@tsu-game.com"
ROOT_PASSWORD="password"

echo "=========================================="
echo "  Root 用户初始化"
echo "=========================================="
echo ""

# ==========================================
# 1. 检查 Kratos 是否就绪
# ==========================================
echo "检查 Kratos 服务..."
if ! curl -sf "$KRATOS_ADMIN_URL/health/ready" > /dev/null; then
    echo "❌ Kratos 服务未就绪，请先启动 Kratos"
    exit 1
fi
echo "✅ Kratos 服务正常"

# ==========================================
# 2. 检查 root identity 是否已存在
# ==========================================
echo ""
echo "检查 root 用户是否已在 Kratos 中存在..."

# 查询 Kratos 中是否存在该邮箱的 identity
EXISTING_IDENTITY=$(curl -sf "$KRATOS_ADMIN_URL/admin/identities" 2>/dev/null | grep -o "\"$ROOT_EMAIL\"" || true)

if [ -n "$EXISTING_IDENTITY" ]; then
    echo "ℹ️  Root 用户已在 Kratos 中存在"
    
    # 获取 identity ID
    IDENTITY_ID=$(curl -sf "$KRATOS_ADMIN_URL/admin/identities" 2>/dev/null | \
        jq -r ".[] | select(.traits.email == \"$ROOT_EMAIL\") | .id" | head -1)
    
    echo "   Identity ID: $IDENTITY_ID"
else
    # ==========================================
    # 3. 在 Kratos 中创建 root identity
    # ==========================================
    echo ""
    echo "在 Kratos 中创建 root identity..."
    
    # 构建请求 JSON
    REQUEST_JSON=$(cat <<EOF
{
    "schema_id": "default",
    "traits": {
        "email": "$ROOT_EMAIL",
        "username": "$ROOT_USERNAME"
    },
    "credentials": {
        "password": {
            "config": {
                "password": "$ROOT_PASSWORD"
            }
        }
    }
}
EOF
)
    
    # 调用 Kratos Admin API 创建 identity
    RESPONSE=$(curl -sf -X POST "$KRATOS_ADMIN_URL/admin/identities" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_JSON" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        IDENTITY_ID=$(echo "$RESPONSE" | jq -r '.id')
        echo "✅ Kratos identity 创建成功"
        echo "   Identity ID: $IDENTITY_ID"
    else
        echo "❌ 创建 Kratos identity 失败"
        exit 1
    fi
fi

# ==========================================
# 4. 在 auth.users 表中同步用户信息
# ==========================================
echo ""
echo "在 auth.users 表中同步用户信息..."

# 使用 docker exec 执行 SQL（避免需要安装 psql 客户端）
docker exec tsu_postgres_main psql -U "$DB_USER" -d "$DB_NAME" <<SQL_EOF
-- 检查并插入 root 用户
DO \$\$
DECLARE
    v_user_exists BOOLEAN;
BEGIN
    -- 检查 auth.users 表是否存在
    IF NOT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'auth' 
        AND table_name = 'users'
    ) THEN
        RAISE WARNING '⚠️  auth.users 表不存在，跳过用户初始化';
        RAISE NOTICE '请先执行数据库迁移';
        RETURN;
    END IF;
    
    -- 检查 root 用户是否已存在
    SELECT EXISTS (
        SELECT 1 FROM auth.users 
        WHERE username = '$ROOT_USERNAME' 
        AND deleted_at IS NULL
    ) INTO v_user_exists;
    
    IF v_user_exists THEN
        RAISE NOTICE 'ℹ️  Root 用户已在数据库中存在';
    ELSE
        -- 插入 root 用户
        INSERT INTO auth.users (
            id,
            username,
            email,
            nickname,
            is_banned,
            login_count,
            created_at,
            updated_at
        ) VALUES (
            '$IDENTITY_ID'::UUID,
            '$ROOT_USERNAME',
            '$ROOT_EMAIL',
            'Administrator',
            false,
            0,
            NOW(),
            NOW()
        );
        
        RAISE NOTICE '✅ Root 用户创建成功';
        RAISE NOTICE '   用户名: $ROOT_USERNAME';
        RAISE NOTICE '   邮箱: $ROOT_EMAIL';
        RAISE NOTICE '   密码: $ROOT_PASSWORD';
    END IF;
END\$\$;
SQL_EOF

# ==========================================
# 完成
# ==========================================
echo ""
echo "=========================================="
echo "  Root 用户初始化完成！"
echo "=========================================="
echo ""
echo "管理员账号信息："
echo "  用户名: $ROOT_USERNAME"
echo "  密码: $ROOT_PASSWORD"
echo "  邮箱: $ROOT_EMAIL"
echo ""
echo "⚠️  重要提示："
echo "  请登录后立即修改密码！"
echo ""
