#!/bin/bash

################################################################################
# 测试套件 08: 效果系统
################################################################################

test_effect_system() {
    start_test_suite "效果系统"
    
    # ===== 效果管理 =====
    
    test_case "获取效果列表"
    http_request "GET" "/api/v1/admin/effects?page=1&page_size=10" "" true
    if assert_success "获取效果列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试效果"
    local effect_id=$(create_test_effect)
    if [ -n "$effect_id" ] && [ "$effect_id" != "null" ]; then
        log_info "创建的效果 ID: $effect_id"
        
        test_case "获取效果详情"
        http_request "GET" "/api/v1/admin/effects/$effect_id" "" true
        if assert_success "获取效果详情成功"; then
            assert_field_exists ".data.id"
            assert_field_exists ".data.name"
        fi
        
        test_case "更新效果"
        local update_data="{
            \"name\": \"${TEST_DATA_PREFIX} 效果 Updated\",
            \"description\": \"已更新的效果\"
        }"
        http_request "PUT" "/api/v1/admin/effects/$effect_id" "$update_data" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新效果成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新效果失败"
        fi
    fi
    
    # ===== Buff 管理 =====
    
    test_case "获取 Buff 列表"
    http_request "GET" "/api/v1/admin/buffs?page=1&page_size=10" "" true
    if assert_success "获取 Buff 列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试 Buff"
    local buff_id=$(create_test_buff)
    if [ -n "$buff_id" ] && [ "$buff_id" != "null" ]; then
        log_info "创建的 Buff ID: $buff_id"
        
        test_case "获取 Buff 详情"
        http_request "GET" "/api/v1/admin/buffs/$buff_id" "" true
        if assert_success "获取 Buff 详情成功"; then
            assert_field_exists ".data.id"
            assert_field_exists ".data.name"
            assert_field_exists ".data.buff_type"
        fi
        
        test_case "更新 Buff"
        local update_buff_data="{
            \"name\": \"${TEST_DATA_PREFIX} Buff Updated\",
            \"description\": \"已更新的 Buff\",
            \"buff_type\": \"negative\"
        }"
        http_request "PUT" "/api/v1/admin/buffs/$buff_id" "$update_buff_data" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新 Buff 成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新 Buff 失败"
        fi
        
        # ===== Buff-Effects 关联 =====
        
        if [ -n "$effect_id" ] && [ "$effect_id" != "null" ]; then
            test_case "获取 Buff 关联的效果列表"
            http_request "GET" "/api/v1/admin/buffs/$buff_id/effects" "" true
            if [ "$LAST_HTTP_CODE" = "200" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 获取 Buff 关联效果列表成功"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 获取 Buff 关联效果列表失败"
            fi
            
            test_case "添加效果到 Buff"
            local add_effect_data="{
                \"effect_id\": \"$effect_id\",
                \"execution_order\": 1,
                \"is_conditional\": false
            }"
            http_request "POST" "/api/v1/admin/buffs/$buff_id/effects" "$add_effect_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 添加效果到 Buff 成功"
                
                # 验证关联已建立
                test_case "验证 Buff-Effect 关联"
                http_request "GET" "/api/v1/admin/buffs/$buff_id/effects" "" true
                if [ "$LAST_HTTP_CODE" = "200" ]; then
                    local effect_count=$(get_array_length ".data")
                    log_info "Buff 关联的效果数量: $effect_count"
                    if [ "$effect_count" -gt 0 ]; then
                        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                        log_success "[$TEST_CASE_NUMBER] Buff-Effect 关联已建立"
                    else
                        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                        log_error "[$TEST_CASE_NUMBER] Buff-Effect 关联未找到"
                    fi
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 验证 Buff-Effect 关联失败"
                fi
                
                # 测试批量设置
                test_case "批量设置 Buff 效果"
                local batch_data="{
                    \"effects\": [
                        {
                            \"effect_id\": \"$effect_id\",
                            \"execution_order\": 1,
                            \"is_conditional\": false
                        }
                    ]
                }"
                http_request "POST" "/api/v1/admin/buffs/$buff_id/effects/batch" "$batch_data" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 批量设置 Buff 效果成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_warning "[$TEST_CASE_NUMBER] 批量设置 Buff 效果失败"
                fi
                
                # 测试移除关联
                test_case "移除 Buff-Effect 关联"
                http_request "DELETE" "/api/v1/admin/buffs/$buff_id/effects/$effect_id" "" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 移除 Buff-Effect 关联成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 移除 Buff-Effect 关联失败"
                fi
                
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 添加效果到 Buff 失败"
            fi
        else
            log_warning "未找到效果 ID，跳过 Buff-Effect 关联测试"
        fi
    fi
    
    end_test_suite
}
