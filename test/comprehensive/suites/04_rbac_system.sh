#!/bin/bash

################################################################################
# 测试套件 04: RBAC 权限系统
################################################################################

test_rbac_system() {
    start_test_suite "RBAC 权限系统"
    
    # ===== 权限管理 =====
    
    # 测试 1: 获取权限列表
    test_case "获取权限列表"
    http_request "GET" "/api/v1/admin/permissions?page=1&page_size=20" "" true
    
    if assert_success "获取权限列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "系统权限总数: $total"
    fi
    
    # 测试 2: 获取权限组列表
    test_case "获取权限组列表"
    http_request "GET" "/api/v1/admin/permission-groups" "" true
    assert_success "获取权限组列表成功"
    
    # ===== 角色管理 =====
    
    # 测试 3: 获取角色列表
    test_case "获取角色列表"
    http_request "GET" "/api/v1/admin/roles?page=1&page_size=10" "" true
    
    if assert_success "获取角色列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "系统角色总数: $total"
    fi
    
    # 测试 4: 创建测试角色
    test_case "创建测试角色"
    local role_name="${TEST_DATA_PREFIX} 测试角色"
    local role_data="{
        \"name\": \"$role_name\",
        \"code\": \"test_role_$(date +%s)\",
        \"description\": \"自动化测试创建的角色\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/roles" "$role_data" true
    
    if assert_success "创建角色成功"; then
        local role_id=$(extract_field '.data.id')
        save_test_id "role" "$role_id"
        log_info "创建的角色 ID: $role_id"
        
        # 测试 5: 获取角色详情
        test_case "获取角色详情"
        http_request "GET" "/api/v1/admin/roles/$role_id" "" true
        
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取角色详情成功 - HTTP $LAST_HTTP_CODE"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取角色详情失败 - HTTP $LAST_HTTP_CODE"
        fi
        
        # 测试 6: 更新角色
        test_case "更新角色"
        local update_role_data="{
            \"name\": \"$role_name (Updated)\",
            \"description\": \"已更新的测试角色\"
        }"
        http_request "PUT" "/api/v1/admin/roles/$role_id" "$update_role_data" true
        
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新角色成功 - HTTP $LAST_HTTP_CODE"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新角色失败 - HTTP $LAST_HTTP_CODE"
        fi
        
        # 测试 7: 获取角色权限
        test_case "获取角色权限"
        http_request "GET" "/api/v1/admin/roles/$role_id/permissions" "" true
        assert_success "获取角色权限成功"
        
        # 测试 8: 为角色分配权限
        test_case "为角色分配权限"
        
        # 先获取一个可用的权限 ID
        http_request "GET" "/api/v1/admin/permissions?page=1&page_size=1" "" true
        local permission_id=$(extract_field '.data.items[0].id')
        
        if [ -n "$permission_id" ] && [ "$permission_id" != "null" ]; then
            local assign_data="{\"permission_ids\": [\"$permission_id\"]}"
            http_request "POST" "/api/v1/admin/roles/$role_id/permissions" "$assign_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 分配权限成功 - HTTP $LAST_HTTP_CODE"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 分配权限失败 - HTTP $LAST_HTTP_CODE"
            fi
        else
            log_warning "未找到可用的权限ID，跳过权限分配测试"
            SUITE_TESTS_TOTAL=$((SUITE_TESTS_TOTAL + 1))
        fi
        
        # ===== 用户-角色关联 =====
        
        if [ -n "$CURRENT_USER_ID" ] && [ "$CURRENT_USER_ID" != "null" ]; then
            # 测试 9: 获取用户角色
            test_case "获取用户角色"
            http_request "GET" "/api/v1/admin/users/$CURRENT_USER_ID/roles" "" true
            assert_success "获取用户角色成功"
            
            # 测试 10: 为用户分配角色
            test_case "为用户分配角色"
            local assign_role_data="{\"role_ids\": [\"$role_id\"]}"
            http_request "POST" "/api/v1/admin/users/$CURRENT_USER_ID/roles" "$assign_role_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 为用户分配角色成功 - HTTP $LAST_HTTP_CODE"
                
                # 测试 11: 验证角色已分配
                test_case "验证角色已分配"
                http_request "GET" "/api/v1/admin/users/$CURRENT_USER_ID/roles" "" true
                assert_success "验证角色分配成功"
                
                # 测试 12: 撤销用户角色
                test_case "撤销用户角色"
                local revoke_data="{\"role_ids\": [\"$role_id\"]}"
                http_request "DELETE" "/api/v1/admin/users/$CURRENT_USER_ID/roles" "$revoke_data" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 撤销用户角色成功 - HTTP $LAST_HTTP_CODE"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 撤销用户角色失败 - HTTP $LAST_HTTP_CODE"
                fi
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 为用户分配角色失败 - HTTP $LAST_HTTP_CODE"
            fi
            
            # 测试 13: 获取用户权限
            test_case "获取用户权限"
            http_request "GET" "/api/v1/admin/users/$CURRENT_USER_ID/permissions" "" true
            assert_success "获取用户权限成功"
        else
            log_warning "跳过用户-角色关联测试：未获取到用户 ID"
        fi
        
        # 测试 14: 删除测试角色
        test_case "删除测试角色"
        http_request "DELETE" "/api/v1/admin/roles/$role_id" "" true
        
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 删除角色成功 - HTTP $LAST_HTTP_CODE"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 删除角色失败 - HTTP $LAST_HTTP_CODE"
        fi
    else
        log_error "创建角色失败，跳过后续角色相关测试"
    fi
    
    end_test_suite
}
