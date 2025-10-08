#!/bin/bash

# 验证游戏配置导入结果

echo "================================"
echo "🎮 游戏配置导入结果验证"
echo "================================"
echo ""

# 数据库连接信息
DB_USER="${DB_USER:-tsu_user}"
DB_NAME="${DB_NAME:-tsu_db}"
CONTAINER="${CONTAINER:-tsu_postgres}"

# 定义要检查的表
TABLES=(
    "hero_attribute_type"
    "damage_types"
    "action_categories"
    "tags"
    "buffs"
    "skills"
    "actions"
)

echo "📊 配置表记录统计："
echo "----------------------------"

TOTAL=0

for table in "${TABLES[@]}"; do
    COUNT=$(docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM game_config.$table" 2>/dev/null | tr -d ' ')
    
    if [ -z "$COUNT" ]; then
        echo "❌ $table: 查询失败"
    elif [ "$COUNT" -eq 0 ]; then
        echo "⚠️  $table: $COUNT 条记录"
    else
        echo "✅ $table: $COUNT 条记录"
        TOTAL=$((TOTAL + COUNT))
    fi
done

echo "----------------------------"
echo "总计: $TOTAL 条记录"
echo ""

# 详细检查每个表的示例数据
echo "📋 示例数据预览："
echo "================================"

echo ""
echo "1️⃣  角色数据类型 (hero_attribute_type):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT code, name, value_type FROM game_config.hero_attribute_type LIMIT 5" 2>/dev/null

echo ""
echo "2️⃣  伤害类型 (damage_types):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT code, name, category FROM game_config.damage_types LIMIT 5" 2>/dev/null

echo ""
echo "3️⃣  特征标签 (tags):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT tag_code, tag_name, category FROM game_config.tags LIMIT 5" 2>/dev/null

echo ""
echo "4️⃣  技能配置 (skills):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT skill_code, skill_name, skill_type FROM game_config.skills LIMIT 5" 2>/dev/null

echo ""
echo "5️⃣  Buff配置 (buffs):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT buff_code, buff_name, buff_type FROM game_config.buffs LIMIT 5" 2>/dev/null

echo ""
echo "6️⃣  动作配置 (actions):"
docker exec $CONTAINER psql -U $DB_USER -d $DB_NAME -c "SELECT action_code, action_name, action_type FROM game_config.actions LIMIT 5" 2>/dev/null

echo ""
echo "================================"
echo "✅ 验证完成"
echo "================================"
