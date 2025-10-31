#!/bin/bash

# 掉落配置管理API简化测试脚本
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost/api/v1/admin"
PASSED=0
FAILED=0

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED=$((PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED=$((FAILED + 1))
}

# 1. 登录
echo "=========================================="
echo "步骤1: 登录"
echo "=========================================="
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier": "root", "password": "password"}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    log_fail "登录失败"
    exit 1
fi
log_pass "登录成功"
echo ""

# 2. 获取已存在的物品ID
echo "=========================================="
echo "步骤2: 获取测试物品"
echo "=========================================="
ITEM_ID=$(curl -s -X GET "$BASE_URL/items?page=1&page_size=1" \
  -H "Cookie: ory_kratos_session=$TOKEN" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$ITEM_ID" ]; then
    log_fail "获取物品ID失败"
    exit 1
fi
log_pass "获取物品ID: $ITEM_ID"
echo ""

# 3. 创建掉落池
echo "=========================================="
echo "掉落池配置管理API测试"
echo "=========================================="

log_test "创建掉落池"
POOL_CODE="API_TEST_$(date +%s)"
response=$(curl -s -X POST "$BASE_URL/drop-pools" \
  -H "Content-Type: application/json" \
  -H "Cookie: ory_kratos_session=$TOKEN" \
  -d '{
    "pool_code": "'$POOL_CODE'",
    "pool_name": "API测试掉落池",
    "pool_type": "monster",
    "description": "API测试",
    "min_drops": 1,
    "max_drops": 3,
    "guaranteed_drops": 1
  }')

code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "创建掉落池成功"
    echo $response | jq '.'
else
    log_fail "创建掉落池失败 (code: $code)"
    echo $response | jq '.'
fi
echo ""

# 获取pool_id
POOL_ID=$(curl -s -X GET "$BASE_URL/drop-pools?keyword=$POOL_CODE" \
  -H "Cookie: ory_kratos_session=$TOKEN" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$POOL_ID" ]; then
    log_fail "无法获取掉落池ID"
    exit 1
fi
echo "掉落池ID: $POOL_ID"
echo ""

# 4. 查询掉落池列表
log_test "查询掉落池列表"
response=$(curl -s -X GET "$BASE_URL/drop-pools?page=1&page_size=10" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "查询掉落池列表成功"
else
    log_fail "查询掉落池列表失败"
fi
echo ""

# 5. 获取掉落池详情
log_test "获取掉落池详情"
response=$(curl -s -X GET "$BASE_URL/drop-pools/$POOL_ID" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "获取掉落池详情成功"
    echo $response | jq '.data'
else
    log_fail "获取掉落池详情失败"
fi
echo ""

# 6. 更新掉落池
log_test "更新掉落池"
response=$(curl -s -X PUT "$BASE_URL/drop-pools/$POOL_ID" \
  -H "Content-Type: application/json" \
  -H "Cookie: ory_kratos_session=$TOKEN" \
  -d '{
    "pool_name": "更新后的API测试掉落池",
    "max_drops": 5
  }')
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "更新掉落池成功"
else
    log_fail "更新掉落池失败"
fi
echo ""

# 7. 添加掉落物品
echo "=========================================="
echo "掉落池物品管理API测试"
echo "=========================================="

log_test "添加掉落物品"
response=$(curl -s -X POST "$BASE_URL/drop-pools/$POOL_ID/items" \
  -H "Content-Type: application/json" \
  -H "Cookie: ory_kratos_session=$TOKEN" \
  -d '{
    "item_id": "'$ITEM_ID'",
    "drop_weight": 100,
    "drop_rate": 0.15,
    "min_quantity": 1,
    "max_quantity": 5,
    "quality_weights": {"normal": 50, "fine": 30, "excellent": 15, "epic": 5}
  }')
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "添加掉落物品成功"
    echo $response | jq '.data'
else
    log_fail "添加掉落物品失败"
    echo $response | jq '.'
fi
echo ""

# 8. 查询掉落池物品列表
log_test "查询掉落池物品列表"
response=$(curl -s -X GET "$BASE_URL/drop-pools/$POOL_ID/items" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "查询掉落池物品列表成功"
else
    log_fail "查询掉落池物品列表失败"
fi
echo ""

# 9. 获取掉落物品详情
log_test "获取掉落物品详情"
response=$(curl -s -X GET "$BASE_URL/drop-pools/$POOL_ID/items/$ITEM_ID" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "获取掉落物品详情成功"
    echo $response | jq '.data'
else
    log_fail "获取掉落物品详情失败"
fi
echo ""

# 10. 更新掉落物品
log_test "更新掉落物品"
response=$(curl -s -X PUT "$BASE_URL/drop-pools/$POOL_ID/items/$ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "Cookie: ory_kratos_session=$TOKEN" \
  -d '{
    "drop_weight": 200,
    "min_quantity": 2,
    "max_quantity": 10
  }')
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "更新掉落物品成功"
else
    log_fail "更新掉落物品失败"
fi
echo ""

# 11. 创建世界掉落配置
echo "=========================================="
echo "世界掉落配置管理API测试"
echo "=========================================="

log_test "创建世界掉落配置"
response=$(curl -s -X POST "$BASE_URL/world-drops" \
  -H "Content-Type: application/json" \
  -H "Cookie: ory_kratos_session=$TOKEN" \
  -d '{
    "item_id": "'$ITEM_ID'",
    "base_drop_rate": 0.05,
    "total_drop_limit": 1000,
    "daily_drop_limit": 100,
    "hourly_drop_limit": 10,
    "min_drop_interval": 300,
    "max_drop_interval": 600,
    "trigger_conditions": {"type": "level_range", "min_level": 10, "max_level": 50},
    "drop_rate_modifiers": {"vip_bonus": 0.05, "event_bonus": 0.1}
  }')
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "创建世界掉落配置成功"
    echo $response | jq '.data'
else
    log_fail "创建世界掉落配置失败"
    echo $response | jq '.'
fi
echo ""

# 获取world_drop_id
WORLD_DROP_ID=$(curl -s -X GET "$BASE_URL/world-drops?item_id=$ITEM_ID" \
  -H "Cookie: ory_kratos_session=$TOKEN" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ ! -z "$WORLD_DROP_ID" ]; then
    echo "世界掉落配置ID: $WORLD_DROP_ID"
    echo ""
    
    # 12. 查询世界掉落配置列表
    log_test "查询世界掉落配置列表"
    response=$(curl -s -X GET "$BASE_URL/world-drops?page=1&page_size=10" \
      -H "Cookie: ory_kratos_session=$TOKEN")
    code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
    if [ "$code" = "100000" ]; then
        log_pass "查询世界掉落配置列表成功"
    else
        log_fail "查询世界掉落配置列表失败"
    fi
    echo ""
    
    # 13. 获取世界掉落配置详情
    log_test "获取世界掉落配置详情"
    response=$(curl -s -X GET "$BASE_URL/world-drops/$WORLD_DROP_ID" \
      -H "Cookie: ory_kratos_session=$TOKEN")
    code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
    if [ "$code" = "100000" ]; then
        log_pass "获取世界掉落配置详情成功"
        echo $response | jq '.data'
    else
        log_fail "获取世界掉落配置详情失败"
    fi
    echo ""
    
    # 14. 更新世界掉落配置
    log_test "更新世界掉落配置"
    response=$(curl -s -X PUT "$BASE_URL/world-drops/$WORLD_DROP_ID" \
      -H "Content-Type: application/json" \
      -H "Cookie: ory_kratos_session=$TOKEN" \
      -d '{
        "base_drop_rate": 0.1,
        "total_drop_limit": 2000
      }')
    code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
    if [ "$code" = "100000" ]; then
        log_pass "更新世界掉落配置成功"
    else
        log_fail "更新世界掉落配置失败"
    fi
    echo ""
    
    # 15. 删除世界掉落配置
    log_test "删除世界掉落配置"
    response=$(curl -s -X DELETE "$BASE_URL/world-drops/$WORLD_DROP_ID" \
      -H "Cookie: ory_kratos_session=$TOKEN")
    code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
    if [ "$code" = "100000" ]; then
        log_pass "删除世界掉落配置成功"
    else
        log_fail "删除世界掉落配置失败"
    fi
    echo ""
fi

# 16. 删除掉落物品
log_test "删除掉落物品"
response=$(curl -s -X DELETE "$BASE_URL/drop-pools/$POOL_ID/items/$ITEM_ID" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "删除掉落物品成功"
else
    log_fail "删除掉落物品失败"
fi
echo ""

# 17. 删除掉落池
log_test "删除掉落池"
response=$(curl -s -X DELETE "$BASE_URL/drop-pools/$POOL_ID" \
  -H "Cookie: ory_kratos_session=$TOKEN")
code=$(echo $response | grep -o '"code":[0-9]*' | cut -d':' -f2)
if [ "$code" = "100000" ]; then
    log_pass "删除掉落池成功"
else
    log_fail "删除掉落池失败"
fi
echo ""

# 测试总结
echo "=========================================="
echo "测试总结"
echo "=========================================="
echo -e "通过: ${GREEN}$PASSED${NC}"
echo -e "失败: ${RED}$FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}有 $FAILED 个测试失败${NC}"
    exit 1
fi

