#!/bin/bash

# 装备套装物品分配API测试脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost/api/v1/admin"
CONTENT_TYPE="Content-Type: application/json"
USERNAME="root"
PASSWORD="password"

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 全局变量
AUTH_TOKEN=""
CREATED_SET_ID=""
CREATED_ITEM_IDS=()

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          装备套装物品分配API测试                                        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 登录函数
login() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}步骤 0: 用户登录${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${YELLOW}正在登录...${NC}"
    echo -e "  用户名: $USERNAME"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "200" ]; then
        AUTH_TOKEN=$(echo "$body" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$AUTH_TOKEN" ]; then
            echo -e "${GREEN}  ✓ 登录成功${NC}"
            echo -e "  Token: ${AUTH_TOKEN:0:20}..."
            echo ""
            return 0
        else
            echo -e "${RED}  ✗ 登录失败: 无法提取token${NC}"
            echo -e "  响应: $body"
            exit 1
        fi
    else
        echo -e "${RED}  ✗ 登录失败 (HTTP $http_code)${NC}"
        echo -e "  响应: $body"
        exit 1
    fi
}

# 测试函数
test_api() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}测试 $TOTAL_TESTS: $test_name${NC}"
    echo -e "  方法: $method"
    echo -e "  端点: $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $AUTH_TOKEN")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo -e "  响应码: $http_code"
    
    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}  ✓ 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "$body"
        return 0
    else
        echo -e "${RED}  ✗ 失败 (期望: $expected_code, 实际: $http_code)${NC}"
        echo -e "  响应: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 执行登录
login

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}步骤 1: 准备测试数据${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 创建测试套装
echo -e "${YELLOW}创建测试套装...${NC}"
TIMESTAMP=$(date +%s)
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/equipment-sets" \
    -H "$CONTENT_TYPE" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -d "{
        \"set_code\": \"test_assign_set_$TIMESTAMP\",
        \"set_name\": \"测试分配套装\",
        \"description\": \"用于测试物品分配的套装\",
        \"set_effects\": [
            {
                \"piece_count\": 2,
                \"effect_description\": \"2件套: 攻击力+10%\",
                \"out_of_combat_effects\": [
                    {
                        \"Data_type\": \"Status\",
                        \"Data_ID\": \"ATK\",
                        \"Bouns_type\": \"percent\",
                        \"Bouns_Number\": \"10\"
                    }
                ],
                \"in_combat_effects\": null
            }
        ],
        \"is_active\": true
    }")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq "200" ]; then
    CREATED_SET_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}  ✓ 套装创建成功${NC}"
    echo -e "  套装ID: $CREATED_SET_ID"
else
    echo -e "${RED}  ✗ 套装创建失败${NC}"
    echo -e "  响应: $body"
    exit 1
fi

# 创建测试装备物品
echo -e "${YELLOW}创建测试装备物品...${NC}"
for i in {1..3}; do
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/items" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "{
            \"item_code\": \"test_assign_eq_${TIMESTAMP}_$i\",
            \"item_name\": \"测试分配装备$i\",
            \"description\": \"用于测试套装分配的装备\",
            \"item_type\": \"equipment\",
            \"item_quality\": \"fine\",
            \"item_level\": 10,
            \"equip_slot\": \"mainhand\",
            \"icon_url\": \"https://example.com/test.png\"
        }")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "200" ]; then
        item_id=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        CREATED_ITEM_IDS+=("$item_id")
        echo -e "${GREEN}  ✓ 装备$i创建成功 (ID: $item_id)${NC}"
    else
        echo -e "${RED}  ✗ 装备$i创建失败${NC}"
        echo -e "  响应: $body"
        exit 1
    fi
done

echo ""

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}步骤 2: 测试物品分配功能${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试1: 通过更新API分配单个物品到套装
test_api "通过更新API分配物品到套装" \
    "PUT" \
    "/items/${CREATED_ITEM_IDS[0]}" \
    "{\"set_id\":\"$CREATED_SET_ID\"}" \
    200

echo ""

# 测试2: 批量分配物品到套装
test_api "批量分配物品到套装" \
    "POST" \
    "/equipment-sets/$CREATED_SET_ID/items/batch-assign" \
    "{\"item_ids\":[\"${CREATED_ITEM_IDS[1]}\",\"${CREATED_ITEM_IDS[2]}\"]}" \
    200

echo ""

# 测试3: 查询套装包含的装备
test_api "查询套装包含的装备" \
    "GET" \
    "/equipment-sets/$CREATED_SET_ID/items" \
    "" \
    200

echo ""

# 测试4: 查询物品详情（应包含套装信息）
test_api "查询物品详情（包含套装信息）" \
    "GET" \
    "/items/${CREATED_ITEM_IDS[0]}" \
    "" \
    200

echo ""

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}步骤 3: 测试物品移除功能${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试5: 移除单个物品从套装
test_api "移除单个物品从套装" \
    "DELETE" \
    "/equipment-sets/$CREATED_SET_ID/items/${CREATED_ITEM_IDS[0]}" \
    "" \
    200

echo ""

# 测试6: 批量移除物品从套装
test_api "批量移除物品从套装" \
    "POST" \
    "/equipment-sets/$CREATED_SET_ID/items/batch-remove" \
    "{\"item_ids\":[\"${CREATED_ITEM_IDS[1]}\",\"${CREATED_ITEM_IDS[2]}\"]}" \
    200

echo ""

# 测试7: 通过更新API移除物品套装关联
test_api "通过更新API移除套装关联" \
    "PUT" \
    "/items/${CREATED_ITEM_IDS[0]}" \
    "{\"set_id\":null}" \
    200

echo ""

# 测试结果汇总
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试结果汇总${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    所有测试通过！ ✓                                    ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔══════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    部分测试失败！ ✗                                    ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi

