#!/bin/bash

################################################################################
# 物品管理端接口全面测试脚本
# 用途: 测试物品配置管理的所有接口,确保接口闭环和文档完整性
# 使用: ./item_admin_endpoints_test.sh
################################################################################

# 不使用 set -e,因为我们要捕获所有错误并继续测试

# ==================== 配置区 ====================
BASE_URL="${BASE_URL:-http://localhost:80}"
API_BASE="/api/v1/admin"
USERNAME="${TEST_USERNAME:-root}"
PASSWORD="${TEST_PASSWORD:-password}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# 全局变量
AUTH_TOKEN=""
CREATED_ITEM_ID=""
CREATED_TAG_IDS=()
CREATED_CLASS_IDS=()
TEST_TIMESTAMP=$(date +%s)

# 测试报告文件
TEST_REPORT_DIR="test_results_${TEST_TIMESTAMP}"
TEST_LOG_FILE="${TEST_REPORT_DIR}/test_log.txt"
FAILED_TESTS_FILE="${TEST_REPORT_DIR}/failed_tests.txt"

# ==================== 工具函数 ====================

# 打印彩色信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1" | tee -a "$TEST_LOG_FILE"
    ((PASSED_TESTS++))
}

print_error() {
    echo -e "${RED}[✗ FAIL]${NC} $1" | tee -a "$TEST_LOG_FILE"
    echo "$1" >> "$FAILED_TESTS_FILE"
    ((FAILED_TESTS++))
}

print_skip() {
    echo -e "${YELLOW}[⊘ SKIP]${NC} $1" | tee -a "$TEST_LOG_FILE"
    ((SKIPPED_TESTS++))
}

print_header() {
    echo -e "\n${CYAN}========================================${NC}" | tee -a "$TEST_LOG_FILE"
    echo -e "${CYAN}$1${NC}" | tee -a "$TEST_LOG_FILE"
    echo -e "${CYAN}========================================${NC}\n" | tee -a "$TEST_LOG_FILE"
}

# 初始化测试环境
init_test_env() {
    mkdir -p "$TEST_REPORT_DIR"
    print_header "物品管理端接口全面测试"
    print_info "测试开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    print_info "API 地址: $BASE_URL"
    print_info "测试账号: $USERNAME"
    print_info "测试报告目录: $TEST_REPORT_DIR"
    print_info ""
}

# HTTP 请求封装 - GET
http_get() {
    local url="$1"
    local desc="$2"
    local expected_code="${3:-200}"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        print_success "$desc (HTTP $http_code)"
        echo "$body"
        return 0
    else
        print_error "$desc - 期望: $expected_code, 实际: $http_code"
        echo "响应: $body" | tee -a "$TEST_LOG_FILE"
        return 1
    fi
}

# HTTP 请求封装 - POST
http_post() {
    local url="$1"
    local data="$2"
    local desc="$3"
    local expected_code="${4:-200}"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        print_success "$desc (HTTP $http_code)"
        echo "$body"
        return 0
    else
        print_error "$desc - 期望: $expected_code, 实际: $http_code"
        echo "响应: $body" | tee -a "$TEST_LOG_FILE"
        return 1
    fi
}

# HTTP 请求封装 - PUT
http_put() {
    local url="$1"
    local data="$2"
    local desc="$3"
    local expected_code="${4:-200}"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        print_success "$desc (HTTP $http_code)"
        echo "$body"
        return 0
    else
        print_error "$desc - 期望: $expected_code, 实际: $http_code"
        echo "响应: $body" | tee -a "$TEST_LOG_FILE"
        return 1
    fi
}

# HTTP 请求封装 - DELETE
http_delete() {
    local url="$1"
    local desc="$2"
    local expected_code="${3:-200}"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        print_success "$desc (HTTP $http_code)"
        echo "$body"
        return 0
    else
        print_error "$desc - 期望: $expected_code, 实际: $http_code"
        echo "响应: $body" | tee -a "$TEST_LOG_FILE"
        return 1
    fi
}

# 提取JSON字段值
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4 | head -1
}

# ==================== 测试函数 ====================

# 1. 认证测试
test_authentication() {
    print_header "1. 认证测试"
    
    local login_data="{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"
    local response
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    ((TOTAL_TESTS++))
    
    if [ "$http_code" -eq "200" ]; then
        AUTH_TOKEN=$(extract_json_field "$body" "session_token")
        
        if [ -n "$AUTH_TOKEN" ]; then
            print_success "用户登录成功"
            print_info "Token: ${AUTH_TOKEN:0:30}..."
            return 0
        else
            print_error "登录成功但未获取到 Token"
            echo "$body" | tee -a "$TEST_LOG_FILE"
            exit 1
        fi
    else
        print_error "用户登录失败 (HTTP $http_code)"
        echo "$body" | tee -a "$TEST_LOG_FILE"
        exit 1
    fi
}

# 2. 物品配置CRUD测试
test_item_crud() {
    print_header "2. 物品配置CRUD测试"
    
    # 2.1 创建物品
    print_info "2.1 创建物品测试"
    
    local item_code="TEST_ITEM_${TEST_TIMESTAMP}"
    local create_data=$(cat <<EOF
{
    "item_code": "$item_code",
    "item_name": "测试物品_${TEST_TIMESTAMP}",
    "item_type": "equipment",
    "item_quality": "fine",
    "item_level": 10,
    "description": "这是一个测试物品",
    "equip_slot": "mainhand",
    "max_durability": 100,
    "base_value": 1000,
    "is_tradable": true,
    "is_droppable": true
}
EOF
)
    
    local response
    response=$(http_post "$API_BASE/items" "$create_data" "创建物品" 200)
    
    if [ $? -eq 0 ]; then
        CREATED_ITEM_ID=$(extract_json_field "$response" "id")
        print_info "创建的物品ID: $CREATED_ITEM_ID"
    else
        print_error "创建物品失败,后续测试可能受影响"
        return 1
    fi
    
    # 2.2 创建重复item_code的物品(应该失败)
    print_info "2.2 创建重复item_code的物品(应该失败)"
    http_post "$API_BASE/items" "$create_data" "创建重复item_code的物品" 409 || true
    
    # 2.3 获取物品详情
    print_info "2.3 获取物品详情"
    http_get "$API_BASE/items/$CREATED_ITEM_ID" "获取物品详情" 200
    
    # 2.4 获取不存在的物品
    print_info "2.4 获取不存在的物品(应该失败)"
    http_get "$API_BASE/items/00000000-0000-0000-0000-000000000000" "获取不存在的物品" 404 || true
    
    # 2.5 更新物品
    print_info "2.5 更新物品"
    local update_data='{"item_name":"更新后的测试物品","item_level":20}'
    http_put "$API_BASE/items/$CREATED_ITEM_ID" "$update_data" "更新物品" 200
}

# 3. 物品列表查询测试
test_item_list() {
    print_header "3. 物品列表查询测试"

    # 3.1 基础分页查询
    print_info "3.1 基础分页查询"
    http_get "$API_BASE/items?page=1&page_size=10" "查询物品列表(分页)" 200

    # 3.2 按类型筛选
    print_info "3.2 按类型筛选"
    http_get "$API_BASE/items?item_type=equipment&page_size=5" "按类型筛选物品" 200

    # 3.3 按品质筛选
    print_info "3.3 按品质筛选"
    http_get "$API_BASE/items?item_quality=fine&page_size=5" "按品质筛选物品" 200

    # 3.4 按装备槽位筛选
    print_info "3.4 按装备槽位筛选"
    http_get "$API_BASE/items?equip_slot=mainhand&page_size=5" "按装备槽位筛选物品" 200

    # 3.5 按等级范围筛选
    print_info "3.5 按等级范围筛选"
    http_get "$API_BASE/items?min_level=10&max_level=20&page_size=5" "按等级范围筛选物品" 200

    # 3.6 按启用状态筛选
    print_info "3.6 按启用状态筛选"
    http_get "$API_BASE/items?is_active=true&page_size=5" "按启用状态筛选物品" 200

    # 3.7 关键词搜索
    print_info "3.7 关键词搜索"
    http_get "$API_BASE/items?keyword=测试&page_size=5" "关键词搜索物品" 200
}

# 4. 物品标签管理测试
test_item_tags() {
    print_header "4. 物品标签管理测试"

    if [ -z "$CREATED_ITEM_ID" ]; then
        print_skip "跳过标签测试 - 物品未创建"
        return 1
    fi

    # 4.1 查询可用标签(先获取一些标签ID)
    print_info "4.1 查询可用标签"
    local tags_response
    tags_response=$(http_get "$API_BASE/tags?page_size=5" "查询标签列表" 200)

    # 提取标签ID(简单处理,实际可能需要更复杂的JSON解析)
    local tag_id1=$(echo "$tags_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    local tag_id2=$(echo "$tags_response" | grep -o '"id":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)

    if [ -n "$tag_id1" ]; then
        print_info "找到标签ID: $tag_id1"

        # 4.2 添加标签
        print_info "4.2 添加标签"
        local add_tags_data="{\"tag_ids\":[\"$tag_id1\"]}"
        http_post "$API_BASE/items/$CREATED_ITEM_ID/tags" "$add_tags_data" "添加标签" 200

        # 4.3 查询物品标签
        print_info "4.3 查询物品标签"
        http_get "$API_BASE/items/$CREATED_ITEM_ID/tags" "查询物品标签" 200

        if [ -n "$tag_id2" ]; then
            # 4.4 批量更新标签
            print_info "4.4 批量更新标签"
            local update_tags_data="{\"tag_ids\":[\"$tag_id2\"]}"
            http_put "$API_BASE/items/$CREATED_ITEM_ID/tags" "$update_tags_data" "批量更新标签" 200

            # 4.5 删除单个标签
            print_info "4.5 删除单个标签"
            http_delete "$API_BASE/items/$CREATED_ITEM_ID/tags/$tag_id2" "删除单个标签" 200
        fi
    else
        print_skip "跳过标签测试 - 未找到可用标签"
    fi

    # 4.6 添加不存在的标签(应该失败)
    print_info "4.6 添加不存在的标签(应该失败)"
    local invalid_tag_data='{"tag_ids":["12345678-1234-4234-8234-123456789012"]}'
    http_post "$API_BASE/items/$CREATED_ITEM_ID/tags" "$invalid_tag_data" "添加不存在的标签" 404 || true
}

# 5. 物品职业关联测试
test_item_classes() {
    print_header "5. 物品职业关联测试"

    if [ -z "$CREATED_ITEM_ID" ]; then
        print_skip "跳过职业关联测试 - 物品未创建"
        return 1
    fi

    # 5.1 查询可用职业(先获取一些职业ID)
    print_info "5.1 查询可用职业"
    local classes_response
    classes_response=$(http_get "$API_BASE/classes?page_size=5" "查询职业列表" 200)

    # 提取职业ID
    local class_id1=$(echo "$classes_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    local class_id2=$(echo "$classes_response" | grep -o '"id":"[^"]*"' | head -2 | tail -1 | cut -d'"' -f4)

    if [ -n "$class_id1" ]; then
        print_info "找到职业ID: $class_id1"

        # 5.2 添加职业限制
        print_info "5.2 添加职业限制"
        local add_classes_data="{\"class_ids\":[\"$class_id1\"]}"
        http_post "$API_BASE/items/$CREATED_ITEM_ID/classes" "$add_classes_data" "添加职业限制" 200

        # 5.3 查询物品职业限制
        print_info "5.3 查询物品职业限制"
        http_get "$API_BASE/items/$CREATED_ITEM_ID/classes" "查询物品职业限制" 200

        if [ -n "$class_id2" ]; then
            # 5.4 批量更新职业限制
            print_info "5.4 批量更新职业限制"
            local update_classes_data="{\"class_ids\":[\"$class_id2\"]}"
            http_put "$API_BASE/items/$CREATED_ITEM_ID/classes" "$update_classes_data" "批量更新职业限制" 200

            # 5.5 删除单个职业限制
            print_info "5.5 删除单个职业限制"
            http_delete "$API_BASE/items/$CREATED_ITEM_ID/classes/$class_id2" "删除单个职业限制" 200
        fi

        # 5.6 清空所有职业限制(变为通用装备)
        print_info "5.6 清空所有职业限制"
        local clear_classes_data='{"class_ids":[]}'
        http_put "$API_BASE/items/$CREATED_ITEM_ID/classes" "$clear_classes_data" "清空职业限制" 200
    else
        print_skip "跳过职业关联测试 - 未找到可用职业"
    fi

    # 5.7 添加不存在的职业(应该失败)
    print_info "5.7 添加不存在的职业(应该失败)"
    local invalid_class_data='{"class_ids":["12345678-1234-4234-8234-123456789012"]}'
    http_post "$API_BASE/items/$CREATED_ITEM_ID/classes" "$invalid_class_data" "添加不存在的职业" 404 || true
}

# 6. 业务流程闭环测试
test_complete_workflow() {
    print_header "6. 业务流程闭环测试"

    print_info "执行完整业务流程: 创建 → 添加标签 → 添加职业 → 查询 → 更新 → 删除"

    # 这个测试已经通过前面的测试覆盖了,这里只是验证数据一致性
    if [ -n "$CREATED_ITEM_ID" ]; then
        print_info "验证物品数据一致性"
        local item_detail
        item_detail=$(http_get "$API_BASE/items/$CREATED_ITEM_ID" "获取物品详情验证" 200)

        if [ $? -eq 0 ]; then
            print_success "业务流程闭环验证通过"
        else
            print_error "业务流程闭环验证失败"
        fi
    fi
}

# 7. 清理测试数据
cleanup_test_data() {
    print_header "7. 清理测试数据"

    if [ -n "$CREATED_ITEM_ID" ]; then
        print_info "删除测试物品: $CREATED_ITEM_ID"
        http_delete "$API_BASE/items/$CREATED_ITEM_ID" "删除测试物品" 200
    else
        print_info "无需清理 - 未创建测试物品"
    fi
}

# 8. 生成测试报告
generate_report() {
    print_header "测试报告"

    local pass_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi

    print_info "总测试数: $TOTAL_TESTS"
    print_info "通过: $PASSED_TESTS"
    print_info "失败: $FAILED_TESTS"
    print_info "跳过: $SKIPPED_TESTS"
    print_info "通过率: ${pass_rate}%"

    if [ $FAILED_TESTS -gt 0 ]; then
        print_info ""
        print_info "失败的测试详情请查看: $FAILED_TESTS_FILE"
    fi

    print_info ""
    print_info "完整测试日志: $TEST_LOG_FILE"
    print_info "测试结束时间: $(date '+%Y-%m-%d %H:%M:%S')"

    # 返回退出码
    if [ $FAILED_TESTS -gt 0 ]; then
        return 1
    else
        return 0
    fi
}

# ==================== 主流程 ====================

main() {
    # 初始化
    init_test_env

    # 执行测试
    test_authentication
    test_item_crud
    test_item_list
    test_item_tags
    test_item_classes
    test_complete_workflow

    # 清理
    cleanup_test_data

    # 生成报告
    generate_report
}

# 执行主流程
main
exit $?

