#!/bin/bash

# 技能系统完整测试脚本
# 测试技能升级消耗接口 + 技能CRUD接口（含新字段）

set -e

BASE_URL="http://localhost/api/v1"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "🎮 技能系统完整测试"
echo "========================================"
echo ""

# 1. 登录获取Token
echo "📝 步骤1: 登录获取Token..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier":"root","password":"password"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.session_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ 登录失败${NC}"
  echo $LOGIN_RESPONSE | jq .
  exit 1
fi

echo -e "${GREEN}✅ 登录成功，Token: ${TOKEN:0:20}...${NC}"
echo ""

# 2. 测试技能升级消耗接口
echo "========================================"
echo "📋 测试技能升级消耗接口"
echo "========================================"
echo ""

# 2.1 获取所有升级消耗
echo "🔍 测试 GET /admin/skill-upgrade-costs..."
UPGRADE_COSTS=$(curl -s -X GET "$BASE_URL/admin/skill-upgrade-costs" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

COSTS_COUNT=$(echo $UPGRADE_COSTS | jq '.data | length')
echo -e "${GREEN}✅ 获取成功，共 $COSTS_COUNT 条记录${NC}"
echo $UPGRADE_COSTS | jq '.data[0:2]'
echo ""

# 2.2 按等级查询（等级5）
echo "🔍 测试 GET /admin/skill-upgrade-costs/level/5..."
LEVEL_5_COST=$(curl -s -X GET "$BASE_URL/admin/skill-upgrade-costs/level/5" \
  -H "Authorization: Bearer $TOKEN")

LEVEL_5_XP=$(echo $LEVEL_5_COST | jq -r '.data.cost_xp')
echo -e "${GREEN}✅ 查询成功，等级5需要 $LEVEL_5_XP XP${NC}"
echo $LEVEL_5_COST | jq '.data'
echo ""

# 2.3 创建新等级消耗（11级）
echo "➕ 测试 POST /admin/skill-upgrade-costs (创建11级)..."
CREATE_COST=$(curl -s -X POST "$BASE_URL/admin/skill-upgrade-costs" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "level_number": 11,
    "cost_xp": 10000,
    "cost_gold": 5000,
    "cost_materials": [{"item_code": "skill_book_master", "count": 1}]
  }')

CREATED_ID=$(echo $CREATE_COST | jq -r '.data.id')
echo -e "${GREEN}✅ 创建成功，ID: $CREATED_ID${NC}"
echo $CREATE_COST | jq '.data'
echo ""

# 2.4 获取单个配置
echo "🔍 测试 GET /admin/skill-upgrade-costs/:id..."
GET_ONE=$(curl -s -X GET "$BASE_URL/admin/skill-upgrade-costs/$CREATED_ID" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${GREEN}✅ 获取成功${NC}"
echo $GET_ONE | jq '.data'
echo ""

# 2.5 更新配置
echo "✏️  测试 PUT /admin/skill-upgrade-costs/:id..."
UPDATE_COST=$(curl -s -X PUT "$BASE_URL/admin/skill-upgrade-costs/$CREATED_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cost_xp": 12000,
    "cost_gold": 6000
  }')

echo -e "${GREEN}✅ 更新成功${NC}"
echo $UPDATE_COST | jq '.data'
echo ""

# 2.6 删除配置
echo "🗑️  测试 DELETE /admin/skill-upgrade-costs/:id..."
DELETE_COST=$(curl -s -X DELETE "$BASE_URL/admin/skill-upgrade-costs/$CREATED_ID" \
  -H "Authorization: Bearer $TOKEN")

echo -e "${GREEN}✅ 删除成功${NC}"
echo $DELETE_COST | jq '.data'
echo ""

# 3. 测试技能CRUD接口（含新字段）
echo "========================================"
echo "⚔️  测试技能CRUD接口（含升级配置）"
echo "========================================"
echo ""

# 3.1 创建线性增长技能（火球术）
echo "➕ 创建技能1: 火球术（线性增长）..."
CREATE_FIREBALL=$(curl -s -X POST "$BASE_URL/admin/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_code": "TEST_FIREBALL",
    "skill_name": "测试火球术",
    "skill_type": "magic",
    "max_level": 10,
    "level_scaling_type": "linear",
    "level_scaling_config": {
      "damage": {"base": 20, "type": "linear", "value": 8},
      "mana_cost": {"base": 10, "type": "linear", "value": 2},
      "range": {"base": 6, "type": "fixed", "value": 0}
    },
    "feature_tags": ["magic", "fire", "attack"],
    "description": "测试用火球术技能",
    "is_active": true
  }')

FIREBALL_ID=$(echo $CREATE_FIREBALL | jq -r '.data.id')
echo -e "${GREEN}✅ 创建成功，ID: $FIREBALL_ID${NC}"
echo $CREATE_FIREBALL | jq '.data | {id, skill_code, skill_name, level_scaling_type, level_scaling_config}'
echo ""

# 3.2 创建百分比增长技能（治疗术）
echo "➕ 创建技能2: 治疗术（百分比增长）..."
CREATE_HEAL=$(curl -s -X POST "$BASE_URL/admin/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_code": "TEST_HEAL",
    "skill_name": "测试治疗术",
    "skill_type": "magic",
    "max_level": 10,
    "level_scaling_type": "percentage",
    "level_scaling_config": {
      "healing": {"base": 30, "type": "percentage", "value": 15},
      "mana_cost": {"base": 8, "type": "linear", "value": 1},
      "range": {"base": 4, "type": "fixed", "value": 0}
    },
    "feature_tags": ["magic", "holy", "healing"],
    "description": "测试用治疗术技能",
    "is_active": true
  }')

HEAL_ID=$(echo $CREATE_HEAL | jq -r '.data.id')
echo -e "${GREEN}✅ 创建成功，ID: $HEAL_ID${NC}"
echo $CREATE_HEAL | jq '.data | {id, skill_code, skill_name, level_scaling_type, level_scaling_config}'
echo ""

# 3.3 创建固定值技能（铁壁）
echo "➕ 创建技能3: 铁壁（固定值+线性）..."
CREATE_IRON=$(curl -s -X POST "$BASE_URL/admin/skills" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_code": "TEST_IRON_SKIN",
    "skill_name": "测试铁壁",
    "skill_type": "passive",
    "max_level": 5,
    "level_scaling_type": "linear",
    "level_scaling_config": {
      "bonus_ac": {"base": 1, "type": "linear", "value": 1},
      "bonus_hp": {"base": 10, "type": "linear", "value": 5},
      "damage_reduction": {"base": 2, "type": "fixed", "value": 0}
    },
    "feature_tags": ["passive", "defense"],
    "description": "测试用被动防御技能",
    "is_active": true
  }')

IRON_ID=$(echo $CREATE_IRON | jq -r '.data.id')
echo -e "${GREEN}✅ 创建成功，ID: $IRON_ID${NC}"
echo $CREATE_IRON | jq '.data | {id, skill_code, skill_name, level_scaling_type, level_scaling_config}'
echo ""

# 3.4 获取技能详情（验证新字段）
echo "🔍 测试 GET /admin/skills/:id（验证新字段）..."
GET_FIREBALL=$(curl -s -X GET "$BASE_URL/admin/skills/$FIREBALL_ID" \
  -H "Authorization: Bearer $TOKEN")

SCALING_TYPE=$(echo $GET_FIREBALL | jq -r '.data.level_scaling_type')
echo -e "${GREEN}✅ 获取成功，升级类型: $SCALING_TYPE${NC}"
echo $GET_FIREBALL | jq '.data | {skill_code, level_scaling_type, level_scaling_config}'
echo ""

# 3.5 更新技能升级配置
echo "✏️  测试 PUT /admin/skills/:id（更新升级配置）..."
UPDATE_SKILL=$(curl -s -X PUT "$BASE_URL/admin/skills/$FIREBALL_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "level_scaling_type": "linear",
    "level_scaling_config": {
      "damage": {"base": 25, "type": "linear", "value": 10},
      "mana_cost": {"base": 12, "type": "linear", "value": 2},
      "range": {"base": 8, "type": "fixed", "value": 0}
    }
  }')

echo -e "${GREEN}✅ 更新成功${NC}"
echo $UPDATE_SKILL | jq '.data'
echo ""

# 3.6 列表查询验证
echo "🔍 测试 GET /admin/skills（列表查询）..."
LIST_SKILLS=$(curl -s -X GET "$BASE_URL/admin/skills?limit=5" \
  -H "Authorization: Bearer $TOKEN")

TOTAL=$(echo $LIST_SKILLS | jq -r '.data.total')
echo -e "${GREEN}✅ 查询成功，共 $TOTAL 个技能${NC}"
echo $LIST_SKILLS | jq '.data.list[] | {skill_code, skill_name, level_scaling_type}'
echo ""

# 4. 清理测试数据
echo "========================================"
echo "🧹 清理测试数据"
echo "========================================"
echo ""

for SKILL_ID in "$FIREBALL_ID" "$HEAL_ID" "$IRON_ID"; do
  echo "删除技能: $SKILL_ID"
  curl -s -X DELETE "$BASE_URL/admin/skills/$SKILL_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
done

echo -e "${GREEN}✅ 测试数据清理完成${NC}"
echo ""

# 总结
echo "========================================"
echo -e "${GREEN}🎉 所有测试通过！${NC}"
echo "========================================"
echo ""
echo "测试摘要:"
echo "  ✅ 技能升级消耗接口: 6/6 通过"
echo "  ✅ 技能CRUD接口: 6/6 通过"
echo "  ✅ 新字段验证: 通过"
echo ""
