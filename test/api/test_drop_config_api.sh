#!/bin/bash

# 掉落配置管理API测试脚本
# 日期: 2025-10-30

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 基础URL
BASE_URL="http://localhost/api/v1/admin"

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试函数
test_api() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=${5:-100000}
    
    log_info "测试: $test_name"
    
    if [ -z "$data" ]; then
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Cookie: ory_kratos_session=$TOKEN")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Cookie: ory_kratos_session=$TOKEN" \
            -d "$data")
    fi
    
    code=$(echo $response | grep -o '"code":[0-9]*' | head -1 | cut -d':' -f2)
    
    if [ "$code" = "$expected_code" ]; then
        log_success "$test_name - 通过 (code: $code)"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "$test_name - 失败 (expected: $expected_code, got: $code)"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    fi
    
    echo ""
}

# 提取字段值
extract_field() {
    echo $1 | grep -o "\"$2\":\"[^\"]*\"" | cut -d'"' -f4
}

# 主测试流程
main() {
    echo "=========================================="
    echo "掉落配置管理API测试"
    echo "=========================================="
    echo ""
    
    # 1. 登录获取Token
    log_info "步骤1: 登录获取Token"
    login_response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"identifier": "root", "password": "password"}')
    
    TOKEN=$(extract_field "$login_response" "session_token")
    
    if [ -z "$TOKEN" ]; then
        log_error "登录失败，无法获取Token"
        exit 1
    fi
    
    log_success "登录成功，Token: ${TOKEN:0:20}..."
    echo ""
    
    # 2. 创建测试物品（如果不存在）
    log_info "步骤2: 准备测试数据"
    ITEM_CODE="API_TEST_ITEM_$(date +%s)"
    
    create_item_data='{
        "item_code": "'$ITEM_CODE'",
        "item_name": "API测试物品",
        "item_type": "equipment",
        "item_quality": "normal",
        "item_level": 1,
        "description": "用于API测试的物品"
    }'
    
    item_response=$(curl -s -X POST "$BASE_URL/items" \
        -H "Content-Type: application/json" \
        -H "Cookie: ory_kratos_session=$TOKEN" \
        -d "$create_item_data")
    
    ITEM_ID=$(extract_field "$item_response" "id")
    
    if [ -z "$ITEM_ID" ]; then
        log_error "创建测试物品失败"
        exit 1
    fi
    
    log_success "测试物品创建成功，ID: $ITEM_ID"
    echo ""
    
    # ==================== 掉落池配置管理测试 ====================
    echo "=========================================="
    echo "掉落池配置管理API测试"
    echo "=========================================="
    echo ""
    
    # 3. 创建掉落池
    POOL_CODE="API_TEST_POOL_$(date +%s)"
    create_pool_data='{
        "pool_code": "'$POOL_CODE'",
        "pool_name": "API测试掉落池",
        "pool_type": "monster",
        "description": "这是一个API测试掉落池",
        "min_drops": 1,
        "max_drops": 3,
        "guaranteed_drops": 1
    }'
    
    test_api "创建掉落池" "POST" "/drop-pools" "$create_pool_data"
    
    # 提取pool_id
    pool_response=$(curl -s -X GET "$BASE_URL/drop-pools?keyword=$POOL_CODE" \
        -H "Cookie: ory_kratos_session=$TOKEN")
    POOL_ID=$(echo $pool_response | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -z "$POOL_ID" ]; then
        log_error "无法获取掉落池ID"
        exit 1
    fi
    
    log_info "掉落池ID: $POOL_ID"
    echo ""
    
    # 4. 查询掉落池列表
    test_api "查询掉落池列表" "GET" "/drop-pools?page=1&page_size=10"
    
    # 5. 按类型筛选
    test_api "按类型筛选掉落池" "GET" "/drop-pools?pool_type=monster"
    
    # 6. 关键词搜索
    test_api "关键词搜索掉落池" "GET" "/drop-pools?keyword=API测试"
    
    # 7. 获取掉落池详情
    test_api "获取掉落池详情" "GET" "/drop-pools/$POOL_ID"
    
    # 8. 更新掉落池
    update_pool_data='{
        "pool_name": "更新后的API测试掉落池",
        "max_drops": 5,
        "description": "这是更新后的描述"
    }'
    
    test_api "更新掉落池" "PUT" "/drop-pools/$POOL_ID" "$update_pool_data"
    
    # ==================== 掉落池物品管理测试 ====================
    echo "=========================================="
    echo "掉落池物品管理API测试"
    echo "=========================================="
    echo ""
    
    # 9. 添加掉落物品
    add_item_data='{
        "item_id": "'$ITEM_ID'",
        "drop_weight": 100,
        "drop_rate": 0.15,
        "min_quantity": 1,
        "max_quantity": 5,
        "quality_weights": {"normal": 50, "fine": 30, "excellent": 15, "epic": 5}
    }'
    
    test_api "添加掉落物品" "POST" "/drop-pools/$POOL_ID/items" "$add_item_data"
    
    # 10. 查询掉落池物品列表
    test_api "查询掉落池物品列表" "GET" "/drop-pools/$POOL_ID/items"
    
    # 11. 获取掉落物品详情
    test_api "获取掉落物品详情" "GET" "/drop-pools/$POOL_ID/items/$ITEM_ID"
    
    # 12. 更新掉落物品
    update_item_data='{
        "drop_weight": 200,
        "min_quantity": 2,
        "max_quantity": 10
    }'
    
    test_api "更新掉落物品" "PUT" "/drop-pools/$POOL_ID/items/$ITEM_ID" "$update_item_data"
    
    # 13. 删除掉落物品
    test_api "删除掉落物品" "DELETE" "/drop-pools/$POOL_ID/items/$ITEM_ID"
    
    # ==================== 世界掉落配置管理测试 ====================
    echo "=========================================="
    echo "世界掉落配置管理API测试"
    echo "=========================================="
    echo ""
    
    # 14. 创建世界掉落配置
    create_world_drop_data='{
        "item_id": "'$ITEM_ID'",
        "base_drop_rate": 0.05,
        "total_drop_limit": 1000,
        "daily_drop_limit": 100,
        "hourly_drop_limit": 10,
        "min_drop_interval": 300,
        "max_drop_interval": 600,
        "trigger_conditions": {"type": "level_range", "min_level": 10, "max_level": 50},
        "drop_rate_modifiers": {"vip_bonus": 0.05, "event_bonus": 0.1}
    }'
    
    test_api "创建世界掉落配置" "POST" "/world-drops" "$create_world_drop_data"
    
    # 提取world_drop_id
    world_drop_response=$(curl -s -X GET "$BASE_URL/world-drops?item_id=$ITEM_ID" \
        -H "Cookie: ory_kratos_session=$TOKEN")
    WORLD_DROP_ID=$(echo $world_drop_response | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -z "$WORLD_DROP_ID" ]; then
        log_error "无法获取世界掉落配置ID"
    else
        log_info "世界掉落配置ID: $WORLD_DROP_ID"
        echo ""
        
        # 15. 查询世界掉落配置列表
        test_api "查询世界掉落配置列表" "GET" "/world-drops?page=1&page_size=10"
        
        # 16. 按物品ID筛选
        test_api "按物品ID筛选世界掉落" "GET" "/world-drops?item_id=$ITEM_ID"
        
        # 17. 获取世界掉落配置详情
        test_api "获取世界掉落配置详情" "GET" "/world-drops/$WORLD_DROP_ID"
        
        # 18. 更新世界掉落配置
        update_world_drop_data='{
            "base_drop_rate": 0.1,
            "total_drop_limit": 2000,
            "daily_drop_limit": 200
        }'
        
        test_api "更新世界掉落配置" "PUT" "/world-drops/$WORLD_DROP_ID" "$update_world_drop_data"
        
        # 19. 删除世界掉落配置
        test_api "删除世界掉落配置" "DELETE" "/world-drops/$WORLD_DROP_ID"
    fi
    
    # ==================== 清理测试数据 ====================
    echo "=========================================="
    echo "清理测试数据"
    echo "=========================================="
    echo ""
    
    # 20. 删除掉落池
    test_api "删除掉落池" "DELETE" "/drop-pools/$POOL_ID"
    
    # 21. 删除测试物品
    delete_item_response=$(curl -s -X DELETE "$BASE_URL/items/$ITEM_ID" \
        -H "Cookie: ory_kratos_session=$TOKEN")
    log_info "删除测试物品: $ITEM_ID"
    
    # ==================== 测试总结 ====================
    echo ""
    echo "=========================================="
    echo "测试总结"
    echo "=========================================="
    echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通过: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}有 $FAILED_TESTS 个测试失败${NC}"
        exit 1
    fi
}

# 执行主函数
main

