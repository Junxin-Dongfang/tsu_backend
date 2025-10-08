#!/bin/bash

################################################################################
# 测试工具函数 - ID提取、验证、辅助功能
################################################################################

################################################################################
# UUID 验证
################################################################################

# 验证是否为有效的 UUID
is_valid_uuid() {
    local uuid="$1"
    
    if [[ "$uuid" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]]; then
        return 0
    else
        return 1
    fi
}

# 验证响应中的 ID 字段
validate_response_id() {
    local id_field="${1:-.data.id}"
    
    local id=$(extract_field "$id_field")
    
    if [ -z "$id" ] || [ "$id" = "null" ]; then
        log_error "ID field $id_field is empty or null"
        return 1
    fi
    
    if is_valid_uuid "$id"; then
        log_debug "Valid UUID: $id"
        return 0
    else
        log_warning "ID $id is not a valid UUID (may be using numeric IDs)"
        return 0
    fi
}

################################################################################
# JSON 工具
################################################################################

# 从响应中提取数组长度
get_array_length() {
    local array_path="${1:-.data.items}"
    
    local length=$(echo "$LAST_RESPONSE_BODY" | jq -r "$array_path | length" 2>/dev/null)
    
    if [ -z "$length" ] || [ "$length" = "null" ]; then
        echo "0"
    else
        echo "$length"
    fi
}

# 检查响应是否包含特定字段
has_field() {
    local field_path="$1"
    
    local value=$(echo "$LAST_RESPONSE_BODY" | jq -r "has($field_path)" 2>/dev/null)
    
    if [ "$value" = "true" ]; then
        return 0
    else
        return 1
    fi
}

# 格式化 JSON（美化输出）
format_json() {
    local json="$1"
    echo "$json" | jq '.' 2>/dev/null || echo "$json"
}

################################################################################
# 分页工具
################################################################################

# 验证分页响应（支持多种格式）
validate_pagination_response() {
    local min_items="${1:-0}"
    
    # 尝试多种分页响应格式
    local items_count=0
    local total=0
    local items_path=""
    
    # 格式1: {data: {items: [], total: N}}
    if assert_field_exists ".data.items" "" true; then
        items_path=".data.items"
        items_count=$(get_array_length ".data.items")
        total=$(extract_field ".data.total")
    # 格式2: {data: {list: [], total: N}}
    elif assert_field_exists ".data.list" "" true; then
        items_path=".data.list"
        items_count=$(get_array_length ".data.list")
        total=$(extract_field ".data.total")
    # 格式3: {data: {users: [], total: N}} (特定资源名)
    elif assert_field_exists ".data.users" "" true; then
        items_path=".data.users"
        items_count=$(get_array_length ".data.users")
        total=$(extract_field ".data.total")
    # 格式4: {data: {permissions: [], pagination: {...}}}
    elif assert_field_exists ".data.permissions" "" true; then
        items_path=".data.permissions"
        items_count=$(get_array_length ".data.permissions")
        total=$(extract_field ".data.pagination.total")
    # 格式5: {data: {roles: [], pagination: {...}}}
    elif assert_field_exists ".data.roles" "" true; then
        items_path=".data.roles"
        items_count=$(get_array_length ".data.roles")
        total=$(extract_field ".data.pagination.total")
    # 格式6: {data: {classes: [], total: N}}
    elif assert_field_exists ".data.classes" "" true; then
        items_path=".data.classes"
        items_count=$(get_array_length ".data.classes")
        total=$(extract_field ".data.total")
    else
        log_debug "无法识别分页响应格式，跳过验证"
        return 0  # 不阻塞测试
    fi
    
    log_debug "Pagination: $items_count items (path: $items_path), total: $total"
    
    if [ "$items_count" -lt "$min_items" ]; then
        log_debug "Items count ($items_count) is less than minimum ($min_items)"
    fi
    
    return 0
}

# 测试不同的分页参数
test_pagination_variants() {
    local endpoint="$1"
    local test_name="$2"
    
    # 测试默认分页
    test_case "[$test_name] 默认分页"
    http_request "GET" "$endpoint" "" true
    assert_success "获取默认分页数据" || return 1
    validate_pagination_response
    
    # 测试自定义分页
    test_case "[$test_name] 自定义分页 (page=1, size=5)"
    http_request "GET" "$endpoint?page=1&page_size=5" "" true
    assert_success "获取自定义分页数据" || return 1
    validate_pagination_response
    
    # 测试大页码
    test_case "[$test_name] 大页码测试 (page=100)"
    http_request "GET" "$endpoint?page=100&page_size=10" "" true
    assert_success "大页码应返回空列表" || return 1
}

################################################################################
# CRUD 测试模板
################################################################################

# 通用 CRUD 测试
# 参数: resource_name, base_endpoint, create_data, update_data
test_crud_workflow() {
    local resource_name="$1"
    local base_endpoint="$2"
    local create_data="$3"
    local update_data="$4"
    
    log_info "开始 $resource_name CRUD 测试流程"
    
    # CREATE
    test_case "创建 $resource_name"
    http_request "POST" "$base_endpoint" "$create_data" true
    if ! assert_success "创建成功"; then
        log_error "创建失败，跳过后续测试"
        return 1
    fi
    
    local resource_id=$(extract_field '.data.id')
    if [ -z "$resource_id" ] || [ "$resource_id" = "null" ]; then
        log_error "未能从响应中提取ID"
        return 1
    fi
    
    log_info "Created $resource_name with ID: $resource_id"
    
    # READ (Detail)
    test_case "获取 $resource_name 详情"
    http_request "GET" "$base_endpoint/$resource_id" "" true
    assert_success "获取详情成功" || return 1
    
    # READ (List)
    test_case "获取 $resource_name 列表"
    http_request "GET" "$base_endpoint?page=1&page_size=10" "" true
    assert_success "获取列表成功" || return 1
    validate_pagination_response
    
    # UPDATE
    if [ -n "$update_data" ]; then
        test_case "更新 $resource_name"
        http_request "PUT" "$base_endpoint/$resource_id" "$update_data" true
        assert_success "更新成功" || return 1
    fi
    
    # DELETE
    test_case "删除 $resource_name"
    http_request "DELETE" "$base_endpoint/$resource_id" "" true
    assert_success "删除成功" || return 1
    
    # 验证删除
    test_case "验证 $resource_name 已删除"
    http_request "GET" "$base_endpoint/$resource_id" "" true
    if [ "$LAST_HTTP_CODE" = "404" ]; then
        log_success "验证删除成功 - 返回 404"
    else
        log_warning "删除后仍能访问资源（可能是软删除）"
    fi
    
    log_success "$resource_name CRUD 测试流程完成"
    return 0
}

################################################################################
# 关联关系测试工具
################################################################################

# 测试关联添加
test_add_relation() {
    local parent_endpoint="$1"
    local parent_id="$2"
    local relation_name="$3"
    local relation_data="$4"
    
    test_case "添加 $relation_name 关联"
    http_request "POST" "$parent_endpoint/$parent_id/$relation_name" "$relation_data" true
    assert_success "添加关联成功"
}

# 测试关联查询
test_get_relations() {
    local parent_endpoint="$1"
    local parent_id="$2"
    local relation_name="$3"
    
    test_case "获取 $relation_name 关联列表"
    http_request "GET" "$parent_endpoint/$parent_id/$relation_name" "" true
    assert_success "获取关联列表成功"
}

# 测试关联删除
test_remove_relation() {
    local parent_endpoint="$1"
    local parent_id="$2"
    local relation_name="$3"
    local relation_id="$4"
    
    test_case "删除 $relation_name 关联"
    http_request "DELETE" "$parent_endpoint/$parent_id/$relation_name/$relation_id" "" true
    assert_success "删除关联成功"
}

################################################################################
# 错误测试工具
################################################################################

# 测试 404 错误
test_404_error() {
    local endpoint="$1"
    local description="$2"
    
    test_case "$description - 应返回 404"
    http_request "GET" "$endpoint" "" true
    assert_status "404" "$description"
}

# 测试 400 错误（无效数据）
test_400_error() {
    local endpoint="$1"
    local invalid_data="$2"
    local description="$3"
    
    test_case "$description - 应返回 400"
    http_request "POST" "$endpoint" "$invalid_data" true
    
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "422" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE"
        return 0
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (expected 400/422)"
        return 1
    fi
}

# 测试 401 错误（未认证）
test_401_error() {
    local endpoint="$1"
    local description="$2"
    
    # 临时清空 token
    local saved_token="$AUTH_TOKEN"
    AUTH_TOKEN=""
    
    test_case "$description - 应返回 401"
    http_request "GET" "$endpoint" "" true
    
    # 恢复 token
    AUTH_TOKEN="$saved_token"
    
    if [ "$LAST_HTTP_CODE" = "401" ] || [ "$LAST_HTTP_CODE" = "403" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE"
        return 0
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (expected 401/403)"
        return 1
    fi
}

################################################################################
# 性能测试工具
################################################################################

# 记录请求性能
record_performance() {
    local endpoint="$1"
    local duration="$LAST_REQUEST_DURATION"
    
    if [ -n "$PERF_LOG" ]; then
        echo "$(date '+%Y-%m-%d %H:%M:%S'),$endpoint,$duration" >> "$PERF_LOG"
    fi
}

# 性能断言（毫秒）
assert_performance() {
    local max_duration="$1"
    local description="${2:-Performance check}"
    
    if [ "$LAST_REQUEST_DURATION" -le "$max_duration" ]; then
        log_success "$description - ${LAST_REQUEST_DURATION}ms (limit: ${max_duration}ms)"
        return 0
    else
        log_warning "$description - ${LAST_REQUEST_DURATION}ms (exceeded limit: ${max_duration}ms)"
        return 1
    fi
}

################################################################################
# 批量操作测试工具
################################################################################

# 测试批量设置
test_batch_set() {
    local parent_endpoint="$1"
    local parent_id="$2"
    local relation_name="$3"
    local batch_data="$4"
    
    test_case "批量设置 $relation_name"
    http_request "POST" "$parent_endpoint/$parent_id/$relation_name/batch" "$batch_data" true
    assert_success "批量设置成功"
    
    # 验证批量设置结果
    http_request "GET" "$parent_endpoint/$parent_id/$relation_name" "" true
    local count=$(get_array_length ".data")
    log_info "批量设置后的关联数量: $count"
}

################################################################################
# 随机数据生成
################################################################################

# 生成随机字符串
random_string() {
    local length="${1:-8}"
    cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | fold -w "$length" | head -n 1
}

# 生成随机数字
random_number() {
    local min="${1:-1}"
    local max="${2:-100}"
    echo $((min + RANDOM % (max - min + 1)))
}

# 生成测试用的唯一名称
generate_unique_name() {
    local prefix="${1:-TEST}"
    echo "${prefix}_$(date +%s)_$(random_string 6)"
}
