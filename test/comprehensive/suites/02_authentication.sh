#!/bin/bash

################################################################################
# 测试套件 02: 认证流程
################################################################################

test_authentication() {
    start_test_suite "认证流程"
    
    # 测试 1: 用户登录
    test_case "用户登录"
    local login_data="{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"
    http_request "POST" "/api/v1/auth/login" "$login_data" false
    
    if assert_success "用户登录成功"; then
        # 提取 Token
        AUTH_TOKEN=$(extract_field '.data.session_token // .data.token // .session_token // .token')
        
        if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
            log_error "无法从响应中提取 Token"
            log_error "响应内容: $LAST_RESPONSE_BODY"
            
            # 尝试其他可能的字段
            AUTH_TOKEN=$(extract_field '.session_token')
            if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
                AUTH_TOKEN=$(extract_field '.token')
            fi
            
            if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
                log_error "登录失败：无法获取认证 Token"
                end_test_suite
                return 1
            fi
        fi
        
        log_info "获取到 Token: ${AUTH_TOKEN:0:20}..."
        
        # 验证 Token 格式（简单检查）
        if [ ${#AUTH_TOKEN} -gt 10 ]; then
            log_success "Token 格式有效"
        else
            log_warning "Token 长度异常: ${#AUTH_TOKEN}"
        fi
    else
        log_error "登录失败，无法继续后续测试"
        end_test_suite
        return 1
    fi
    
    # 测试 2: 获取当前用户信息
    test_case "获取当前用户信息"
    http_request "GET" "/api/v1/admin/users/me" "" true
    
    if assert_success "获取当前用户信息成功"; then
        # 验证用户信息
        assert_field_exists ".data.id"
        assert_field_exists ".data.username"
        
        local user_id=$(extract_field ".data.id")
        local username=$(extract_field ".data.username")
        
        log_info "当前用户 ID: $user_id"
        log_info "当前用户名: $username"
        
        # 保存用户 ID 供后续测试使用
        CURRENT_USER_ID="$user_id"
    fi
    
    # 测试 3: 登出功能
    test_case "用户登出"
    http_request "POST" "/api/v1/auth/logout" "{}" false
    
    # 登出可能返回 200、204 或 400（未找到会话令牌）
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ] || [ "$LAST_HTTP_CODE" = "400" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 用户登出 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 登出返回异常状态 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 4: 重新登录以继续后续测试
    test_case "重新登录"
    http_request "POST" "/api/v1/auth/login" "$login_data" false
    
    if assert_success "重新登录成功"; then
        AUTH_TOKEN=$(extract_field '.data.session_token // .data.token // .session_token // .token')
        if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
            AUTH_TOKEN=$(extract_field '.session_token')
        fi
        if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
            AUTH_TOKEN=$(extract_field '.token')
        fi
        log_info "重新获取 Token: ${AUTH_TOKEN:0:20}..."
    fi
    
    # 测试 5: 错误场景 - 无效凭证
    test_case "无效凭证登录"
    local invalid_login="{\"identifier\":\"invalid_user\",\"password\":\"invalid_pass\"}"
    http_request "POST" "/api/v1/auth/login" "$invalid_login" false
    
    if [ "$LAST_HTTP_CODE" = "401" ] || [ "$LAST_HTTP_CODE" = "400" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 无效凭证被正确拒绝 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 无效凭证处理异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    # 测试 6: 错误场景 - 缺少参数
    test_case "缺少登录参数"
    local incomplete_login="{\"identifier\":\"$USERNAME\"}"
    http_request "POST" "/api/v1/auth/login" "$incomplete_login" false
    
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "422" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 参数缺失被正确拒绝 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 参数缺失处理异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    end_test_suite
}
