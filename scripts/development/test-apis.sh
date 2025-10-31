#!/bin/bash

# API接口测试脚本
# 使用方式: ./scripts/development/test-apis.sh

set -e

BASE_URL="http://localhost"
ADMIN_API="${BASE_URL}/api/v1/admin"
GAME_API="${BASE_URL}/api/v1/game"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  TSU API 接口测试"
echo "=========================================="
echo ""

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    local token=$5
    
    echo -e "${YELLOW}测试: ${name}${NC}"
    echo "URL: ${method} ${url}"
    
    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -X ${method} "${url}" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer ${token}" \
                -d "${data}")
        else
            response=$(curl -s -X ${method} "${url}" \
                -H "Authorization: Bearer ${token}")
        fi
    else
        if [ -n "$data" ]; then
            response=$(curl -s -X ${method} "${url}" \
                -H "Content-Type: application/json" \
                -d "${data}")
        else
            response=$(curl -s -X ${method} "${url}")
        fi
    fi
    
    echo "响应: ${response}" | jq '.' 2>/dev/null || echo "${response}"
    
    # 检查响应码
    code=$(echo "${response}" | jq -r '.code' 2>/dev/null || echo "")
    if [ "$code" = "100000" ] || [ "$code" = "200" ]; then
        echo -e "${GREEN}✅ 成功${NC}"
    else
        echo -e "${RED}❌ 失败${NC}"
    fi
    echo ""
}

# ==================== Admin Server 测试 ====================

echo "=========================================="
echo "  Admin Server API 测试"
echo "=========================================="
echo ""

# 1. 注册管理员账号
echo "1. 注册管理员账号"
test_api "注册管理员" "POST" "${ADMIN_API}/auth/register" \
'{
  "email": "admin@test.com",
  "username": "admin_test",
  "password": "admin123456"
}'

# 2. 登录获取token
echo "2. 登录管理员账号"
login_response=$(curl -s -X POST "${ADMIN_API}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
      "identifier": "admin@test.com",
      "password": "admin123456"
    }')

echo "登录响应: ${login_response}" | jq '.'
ADMIN_TOKEN=$(echo "${login_response}" | jq -r '.data.session_token' 2>/dev/null || echo "")

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo -e "${RED}❌ 登录失败，无法获取token${NC}"
    echo "尝试使用已有账号登录..."
    # 可能账号已存在，继续测试
else
    echo -e "${GREEN}✅ 登录成功，Token: ${ADMIN_TOKEN:0:20}...${NC}"
fi
echo ""

# 3. 查询物品列表（无需token也可以测试权限）
echo "3. 查询物品列表"
test_api "查询物品列表" "GET" "${ADMIN_API}/items?page=1&page_size=10" "" "${ADMIN_TOKEN}"

# 4. 创建物品配置
echo "4. 创建物品配置"
test_api "创建测试物品" "POST" "${ADMIN_API}/items" \
'{
  "item_code": "test_sword_001",
  "item_name": "测试长剑",
  "item_type": "equipment",
  "item_quality": "fine",
  "item_level": 10,
  "description": "一把用于测试的长剑",
  "equip_slot": "mainhand",
  "base_price": 100,
  "max_stack": 1,
  "is_tradable": true,
  "is_droppable": true
}' "${ADMIN_TOKEN}"

# ==================== Game Server 测试 ====================

echo "=========================================="
echo "  Game Server API 测试"
echo "=========================================="
echo ""

# 1. 注册玩家账号
echo "1. 注册玩家账号"
test_api "注册玩家" "POST" "${GAME_API}/auth/register" \
'{
  "email": "player@test.com",
  "username": "player_test",
  "password": "player123456"
}'

# 2. 登录获取token
echo "2. 登录玩家账号"
game_login_response=$(curl -s -X POST "${GAME_API}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
      "identifier": "player@test.com",
      "password": "player123456"
    }')

echo "登录响应: ${game_login_response}" | jq '.'
GAME_TOKEN=$(echo "${game_login_response}" | jq -r '.data.session_token' 2>/dev/null || echo "")

if [ -z "$GAME_TOKEN" ] || [ "$GAME_TOKEN" = "null" ]; then
    echo -e "${RED}❌ 登录失败，无法获取token${NC}"
else
    echo -e "${GREEN}✅ 登录成功，Token: ${GAME_TOKEN:0:20}...${NC}"
fi
echo ""

# 3. 创建英雄
echo "3. 创建英雄"
hero_response=$(curl -s -X POST "${GAME_API}/heroes" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${GAME_TOKEN}" \
    -d '{
      "hero_name": "测试战士",
      "class_id": "warrior"
    }')

echo "创建英雄响应: ${hero_response}" | jq '.'
HERO_ID=$(echo "${hero_response}" | jq -r '.data.hero_id' 2>/dev/null || echo "")

if [ -z "$HERO_ID" ] || [ "$HERO_ID" = "null" ]; then
    echo -e "${RED}❌ 创建英雄失败${NC}"
else
    echo -e "${GREEN}✅ 创建英雄成功，Hero ID: ${HERO_ID}${NC}"
fi
echo ""

# 4. 查询装备槽位
if [ -n "$HERO_ID" ] && [ "$HERO_ID" != "null" ]; then
    echo "4. 查询装备槽位"
    test_api "查询装备槽位" "GET" "${GAME_API}/equipment/slots/${HERO_ID}" "" "${GAME_TOKEN}"
fi

# 5. 查询背包
echo "5. 查询背包"
USER_ID=$(echo "${game_login_response}" | jq -r '.data.user_id' 2>/dev/null || echo "")
if [ -n "$USER_ID" ] && [ "$USER_ID" != "null" ]; then
    test_api "查询背包" "GET" "${GAME_API}/inventory?owner_id=${USER_ID}&item_location=backpack&page=1&page_size=20" "" "${GAME_TOKEN}"
fi

echo "=========================================="
echo "  测试完成"
echo "=========================================="
echo ""
echo "访问Swagger文档:"
echo "  Admin: http://localhost/admin/swagger/index.html"
echo "  Game:  http://localhost/game/swagger/index.html"

