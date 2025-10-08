#!/bin/bash

################################################################################
# 测试套件 01: 系统健康检查
################################################################################

test_system_health() {
    start_test_suite "系统健康检查"
    
    # 测试 1: 健康检查接口
    test_case "健康检查接口"
    http_request "GET" "/health" "" false
    assert_status "200" "健康检查接口应返回 200"
    
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        # 验证响应包含 status 字段
        assert_field_exists ".status"
        local status=$(extract_field ".status")
        log_info "系统状态: $status"
    fi
    
    # 测试 2: Swagger 文档可访问性
    test_case "Swagger 文档访问"
    http_request "GET" "/swagger/index.html" "" false
    
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "301" ] || [ "$LAST_HTTP_CODE" = "302" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] Swagger 文档可访问 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] Swagger 文档不可访问 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 3: API 基础路径测试
    test_case "API v1 路径测试"
    http_request "GET" "/api/v1/admin/users" "" false
    
    # 预期返回 401 或 403（因为没有认证）
    if [ "$LAST_HTTP_CODE" = "401" ] || [ "$LAST_HTTP_CODE" = "403" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] API 路径正确（返回认证错误）- HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] API 路径响应异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    end_test_suite
}
