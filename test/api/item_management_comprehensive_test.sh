#!/bin/bash

# 物品管理端接口全面测试脚本

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
MISSING_APIS=()
ERRORS=()

# 全局变量
AUTH_TOKEN=""
CREATED_ITEM_ID=""
CREATED_SET_ID=""

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          物品管理端接口全面测试                                        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 登录函数
login() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}步骤 0: 用户登录${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "200" ]; then
        AUTH_TOKEN=$(echo "$body" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$AUTH_TOKEN" ]; then
            echo -e "${GREEN}✓ 登录成功${NC}"
            echo ""
            return 0
        fi
    fi
    
    echo -e "${RED}✗ 登录失败${NC}"
    exit 1
}

# 测试函数
test_api() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5
    local is_critical=${6:-false}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}测试 $TOTAL_TESTS: $test_name${NC}"
    
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
    
    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ 通过 (HTTP $http_code)${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "$body"
        echo ""
        return 0
    else
        echo -e "${RED}✗ 失败 (期望: $expected_code, 实际: $http_code)${NC}"
        echo -e "响应: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        
        if [ "$is_critical" = true ]; then
            ERRORS+=("$test_name: HTTP $http_code (期望 $expected_code)")
        fi
        
        if [ "$http_code" -eq "404" ]; then
            MISSING_APIS+=("$method $endpoint")
        fi
        
        echo ""
        return 1
    fi
}

# 执行登录
login

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}第一部分: 物品配置CRUD测试${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 1. 创建物品
response=$(test_api "创建物品" "POST" "/items" \
    "{
        \"item_code\": \"test_item_$(date +%s)\",
        \"item_name\": \"测试物品\",
        \"description\": \"全面测试用物品\",
        \"item_type\": \"equipment\",
        \"item_quality\": \"fine\",
        \"item_level\": 10,
        \"equip_slot\": \"mainhand\",
        \"icon_url\": \"https://example.com/icon.png\"
    }" \
    200 true)

if [ $? -eq 0 ]; then
    CREATED_ITEM_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}创建的物品ID: $CREATED_ITEM_ID${NC}"
    echo ""
fi

# 2. 查询物品详情
if [ -n "$CREATED_ITEM_ID" ]; then
    test_api "查询物品详情" "GET" "/items/$CREATED_ITEM_ID" "" 200
fi

# 3. 更新物品
if [ -n "$CREATED_ITEM_ID" ]; then
    test_api "更新物品" "PUT" "/items/$CREATED_ITEM_ID" \
        "{\"item_name\": \"更新后的测试物品\"}" \
        200
fi

# 4. 查询物品列表
test_api "查询物品列表" "GET" "/items?page=1&page_size=10" "" 200

# 5. 按类型筛选物品
test_api "按类型筛选物品" "GET" "/items?item_type=equipment&page_size=5" "" 200

# 6. 按品质筛选物品
test_api "按品质筛选物品" "GET" "/items?item_quality=fine&page_size=5" "" 200

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}第二部分: 物品职业限制测试${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 7. 查询职业列表（用于获取职业ID）
response=$(test_api "查询职业列表" "GET" "/classes?page_size=2" "" 200)
CLASS_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

# 8. 更新物品职业限制
if [ -n "$CREATED_ITEM_ID" ] && [ -n "$CLASS_ID" ]; then
    test_api "更新物品职业限制" "PUT" "/items/$CREATED_ITEM_ID/classes" \
        "{\"class_ids\": [\"$CLASS_ID\"]}" \
        200
fi

# 9. 查询物品职业限制
if [ -n "$CREATED_ITEM_ID" ]; then
    test_api "查询物品职业限制" "GET" "/items/$CREATED_ITEM_ID" "" 200
fi

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}第三部分: 装备套装测试${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 10. 创建装备套装
response=$(test_api "创建装备套装" "POST" "/equipment-sets" \
    "{
        \"set_code\": \"test_set_$(date +%s)\",
        \"set_name\": \"测试套装\",
        \"description\": \"全面测试用套装\",
        \"set_effects\": [
            {
                \"piece_count\": 2,
                \"effect_description\": \"2件套效果\",
                \"out_of_combat_effects\": [],
                \"in_combat_effects\": null
            }
        ],
        \"is_active\": true
    }" \
    200)

if [ $? -eq 0 ]; then
    CREATED_SET_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "${GREEN}创建的套装ID: $CREATED_SET_ID${NC}"
    echo ""
fi

# 11. 查询套装详情
if [ -n "$CREATED_SET_ID" ]; then
    test_api "查询套装详情" "GET" "/equipment-sets/$CREATED_SET_ID" "" 200
fi

# 12. 更新套装
if [ -n "$CREATED_SET_ID" ]; then
    test_api "更新套装" "PUT" "/equipment-sets/$CREATED_SET_ID" \
        "{\"set_name\": \"更新后的测试套装\"}" \
        200
fi

# 13. 查询套装列表
test_api "查询套装列表" "GET" "/equipment-sets?page=1&page_size=10" "" 200

# 14. 分配物品到套装
if [ -n "$CREATED_ITEM_ID" ] && [ -n "$CREATED_SET_ID" ]; then
    test_api "分配物品到套装" "PUT" "/items/$CREATED_ITEM_ID" \
        "{\"set_id\": \"$CREATED_SET_ID\"}" \
        200
fi

# 15. 查询套装包含的物品
if [ -n "$CREATED_SET_ID" ]; then
    test_api "查询套装包含的物品" "GET" "/equipment-sets/$CREATED_SET_ID/items" "" 200
fi

# 16. 批量分配物品到套装
if [ -n "$CREATED_ITEM_ID" ] && [ -n "$CREATED_SET_ID" ]; then
    test_api "批量分配物品到套装" "POST" "/equipment-sets/$CREATED_SET_ID/items/batch-assign" \
        "{\"item_ids\": [\"$CREATED_ITEM_ID\"]}" \
        200
fi

# 17. 从套装移除物品
if [ -n "$CREATED_ITEM_ID" ] && [ -n "$CREATED_SET_ID" ]; then
    test_api "从套装移除物品" "DELETE" "/equipment-sets/$CREATED_SET_ID/items/$CREATED_ITEM_ID" \
        "" \
        200
fi

# 18. 查询未分配套装的物品
test_api "查询未分配套装的物品" "GET" "/equipment-sets/unassigned-items?page_size=5" "" 200

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}第四部分: 掉落配置测试${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 19. 创建掉落池
test_api "创建掉落池" "POST" "/drop-pools" \
    "{
        \"pool_code\": \"test_pool_$(date +%s)\",
        \"pool_name\": \"测试掉落池\",
        \"description\": \"全面测试用掉落池\"
    }" \
    200

# 20. 查询掉落池列表
test_api "查询掉落池列表" "GET" "/drop-pools?page=1&page_size=10" "" 200

# 21. 创建世界掉落配置
test_api "创建世界掉落配置" "POST" "/world-drops" \
    "{
        \"drop_code\": \"test_world_drop_$(date +%s)\",
        \"drop_name\": \"测试世界掉落\",
        \"description\": \"全面测试用世界掉落\"
    }" \
    200

# 22. 查询世界掉落列表
test_api "查询世界掉落列表" "GET" "/world-drops?page=1&page_size=10" "" 200

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}第五部分: 删除操作测试${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 23. 删除套装
if [ -n "$CREATED_SET_ID" ]; then
    test_api "删除套装" "DELETE" "/equipment-sets/$CREATED_SET_ID" "" 200
fi

# 24. 删除物品
if [ -n "$CREATED_ITEM_ID" ]; then
    test_api "删除物品" "DELETE" "/items/$CREATED_ITEM_ID" "" 200
fi

# 测试结果汇总
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试结果汇总${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

if [ ${#MISSING_APIS[@]} -gt 0 ]; then
    echo -e "${YELLOW}缺失的API端点:${NC}"
    for api in "${MISSING_APIS[@]}"; do
        echo -e "  - $api"
    done
    echo ""
fi

if [ ${#ERRORS[@]} -gt 0 ]; then
    echo -e "${RED}关键错误:${NC}"
    for error in "${ERRORS[@]}"; do
        echo -e "  - $error"
    done
    echo ""
fi

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

