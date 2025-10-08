#!/bin/bash

################################################################################
# 测试套件 03: 用户管理
################################################################################

test_user_management() {
    start_test_suite "用户管理"
    
    # 测试 1: 获取用户列表
    test_case "获取用户列表（分页）"
    http_request "GET" "/api/v1/admin/users?page=1&page_size=10" "" true
    
    if assert_success "获取用户列表成功"; then
        validate_pagination_response 1
        
        local total=$(extract_field ".data.total")
        log_info "系统用户总数: $total"
    fi
    
    # 测试 2: 获取用户详情
    if [ -n "$CURRENT_USER_ID" ] && [ "$CURRENT_USER_ID" != "null" ]; then
        test_case "获取用户详情"
        http_request "GET" "/api/v1/admin/users/$CURRENT_USER_ID" "" true
        
        if assert_success "获取用户详情成功"; then
            assert_field_exists ".data.id"
            assert_field_exists ".data.username"
            assert_field_exists ".data.email"
            
            local username=$(extract_field ".data.username")
            local email=$(extract_field ".data.email")
            log_info "用户名: $username, 邮箱: $email"
        fi
    else
        log_warning "跳过用户详情测试：未获取到用户 ID"
        SUITE_TESTS_TOTAL=$((SUITE_TESTS_TOTAL + 1))
    fi
    
    # 测试 3: 更新用户信息
    if [ -n "$CURRENT_USER_ID" ] && [ "$CURRENT_USER_ID" != "null" ]; then
        test_case "更新用户信息"
        local update_data="{\"display_name\":\"Test Admin Updated\"}"
        http_request "PUT" "/api/v1/admin/users/$CURRENT_USER_ID" "$update_data" true
        
        # 更新可能返回 200 或 204
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新用户信息成功 - HTTP $LAST_HTTP_CODE"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新用户信息失败 - HTTP $LAST_HTTP_CODE"
        fi
    else
        log_warning "跳过用户更新测试：未获取到用户 ID"
        SUITE_TESTS_TOTAL=$((SUITE_TESTS_TOTAL + 1))
    fi
    
    # 测试 4: 用户列表搜索/过滤
    test_case "用户列表搜索（按用户名）"
    http_request "GET" "/api/v1/admin/users?page=1&page_size=10&username=$USERNAME" "" true
    
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 用户搜索成功 - HTTP $LAST_HTTP_CODE"
        
        local items_count=$(get_array_length ".data.items")
        log_info "搜索结果数量: $items_count"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 用户搜索失败 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 5: 错误场景 - 访问不存在的用户
    test_case "访问不存在的用户"
    local fake_uuid="00000000-0000-0000-0000-000000000000"
    http_request "GET" "/api/v1/admin/users/$fake_uuid" "" true
    
    if [ "$LAST_HTTP_CODE" = "404" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 不存在用户返回 404 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 404 错误处理异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 6: 错误场景 - 无效的分页参数
    test_case "无效的分页参数"
    http_request "GET" "/api/v1/admin/users?page=-1&page_size=0" "" true
    
    # 可能返回 400 或者忽略无效参数返回 200
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 无效分页参数处理 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 分页参数验证异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 7: 封禁用户功能（如果有测试用户）
    # 注意：不要封禁当前登录用户
    test_case "封禁/解封用户功能检查"
    log_info "跳过封禁测试（避免封禁当前登录用户）"
    # 这里可以创建一个测试用户来测试封禁功能
    
    end_test_suite
}
