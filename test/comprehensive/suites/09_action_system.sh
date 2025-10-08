#!/bin/bash

################################################################################
# 测试套件 09: 动作系统
################################################################################

test_action_system() {
    start_test_suite "动作系统"
    
    # 确保有动作分类
    local action_category_id=$(get_test_id "action_category")
    if [ -z "$action_category_id" ] || [ "$action_category_id" = "null" ]; then
        log_info "未找到已创建的动作分类，创建新的"
        action_category_id=$(create_test_action_category)
    fi
    
    if [ -z "$action_category_id" ] || [ "$action_category_id" = "null" ]; then
        log_error "无法获取动作分类 ID，跳过动作系统测试"
        end_test_suite
        return 1
    fi
    
    # ===== 动作管理 =====
    
    test_case "获取动作列表"
    http_request "GET" "/api/v1/admin/actions?page=1&page_size=10" "" true
    if assert_success "获取动作列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试动作"
    local action_code="TEST_ACT_$(date +%s)"
    local action_id=$(create_test_action "$action_category_id" "$action_code")
    if [ -n "$action_id" ] && [ "$action_id" != "null" ]; then
        log_info "创建的动作 ID: $action_id"
        
        test_case "获取动作详情"
        http_request "GET" "/api/v1/admin/actions/$action_id" "" true
        if assert_success "获取动作详情成功"; then
            assert_field_exists ".data.id"
            assert_field_exists ".data.action_code"
            assert_field_exists ".data.action_name"
        fi
        
        test_case "更新动作"
        local update_data="{
            \"action_name\": \"${TEST_DATA_PREFIX} 动作 Updated\",
            \"description\": \"已更新的动作\",
            \"action_point_cost\": 2
        }"
        http_request "PUT" "/api/v1/admin/actions/$action_id" "$update_data" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新动作成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新动作失败"
        fi
        
        # ===== Action-Effects 关联 =====
        
        # 确保有效果可用
        local effect_id=$(get_test_id "effect")
        if [ -z "$effect_id" ] || [ "$effect_id" = "null" ]; then
            log_info "未找到已创建的效果，创建新的"
            effect_id=$(create_test_effect)
        fi
        
        if [ -n "$effect_id" ] && [ "$effect_id" != "null" ]; then
            test_case "获取动作关联的效果列表"
            http_request "GET" "/api/v1/admin/actions/$action_id/effects" "" true
            if [ "$LAST_HTTP_CODE" = "200" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 获取动作关联效果列表成功"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 获取动作关联效果列表失败"
            fi
            
            test_case "添加效果到动作"
            local add_effect_data="{
                \"effect_id\": \"$effect_id\",
                \"execution_order\": 1,
                \"is_conditional\": false
            }"
            http_request "POST" "/api/v1/admin/actions/$action_id/effects" "$add_effect_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 添加效果到动作成功"
                
                # 验证关联已建立
                test_case "验证 Action-Effect 关联"
                http_request "GET" "/api/v1/admin/actions/$action_id/effects" "" true
                if [ "$LAST_HTTP_CODE" = "200" ]; then
                    local effect_count=$(get_array_length ".data")
                    log_info "动作关联的效果数量: $effect_count"
                    if [ "$effect_count" -gt 0 ]; then
                        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                        log_success "[$TEST_CASE_NUMBER] Action-Effect 关联已建立"
                    else
                        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                        log_error "[$TEST_CASE_NUMBER] Action-Effect 关联未找到"
                    fi
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 验证 Action-Effect 关联失败"
                fi
                
                # 测试批量设置
                test_case "批量设置动作效果"
                local batch_data="{
                    \"effects\": [
                        {
                            \"effect_id\": \"$effect_id\",
                            \"execution_order\": 1,
                            \"is_conditional\": false
                        }
                    ]
                }"
                http_request "POST" "/api/v1/admin/actions/$action_id/effects/batch" "$batch_data" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 批量设置动作效果成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_warning "[$TEST_CASE_NUMBER] 批量设置动作效果失败"
                fi
                
                # 测试移除关联
                test_case "移除 Action-Effect 关联"
                http_request "DELETE" "/api/v1/admin/actions/$action_id/effects/$effect_id" "" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 移除 Action-Effect 关联成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 移除 Action-Effect 关联失败"
                fi
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 添加效果到动作失败"
            fi
        else
            log_warning "未找到效果 ID，跳过 Action-Effect 关联测试"
        fi
        
        # ===== Skill-Unlock-Actions 关联 =====
        
        # 确保有技能可用
        local skill_id=$(get_test_id "skill")
        if [ -n "$skill_id" ] && [ "$skill_id" != "null" ]; then
            test_case "获取技能解锁动作列表"
            http_request "GET" "/api/v1/admin/skills/$skill_id/unlock-actions" "" true
            if [ "$LAST_HTTP_CODE" = "200" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 获取技能解锁动作列表成功"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 获取技能解锁动作列表失败"
            fi
            
            test_case "添加解锁动作到技能"
            local add_unlock_data="{
                \"action_id\": \"$action_id\",
                \"unlock_level\": 1,
                \"is_default\": true
            }"
            http_request "POST" "/api/v1/admin/skills/$skill_id/unlock-actions" "$add_unlock_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 添加解锁动作到技能成功"
                
                # 测试批量设置
                test_case "批量设置技能解锁动作"
                local batch_unlock_data="{
                    \"actions\": [
                        {
                            \"action_id\": \"$action_id\",
                            \"unlock_level\": 1,
                            \"is_default\": true
                        }
                    ]
                }"
                http_request "POST" "/api/v1/admin/skills/$skill_id/unlock-actions/batch" "$batch_unlock_data" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 批量设置技能解锁动作成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_warning "[$TEST_CASE_NUMBER] 批量设置技能解锁动作失败"
                fi
                
                # 测试移除关联
                test_case "移除技能解锁动作"
                http_request "DELETE" "/api/v1/admin/skills/$skill_id/unlock-actions/$action_id" "" true
                
                if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 移除技能解锁动作成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 移除技能解锁动作失败"
                fi
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 添加解锁动作到技能失败"
            fi
        else
            log_warning "未找到技能 ID，跳过 Skill-Unlock-Actions 关联测试"
        fi
    fi
    
    # ===== 动作搜索和过滤 =====
    
    test_case "按分类搜索动作"
    http_request "GET" "/api/v1/admin/actions?page=1&page_size=10&category_id=$action_category_id" "" true
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 按分类搜索动作成功"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 按分类搜索动作失败"
    fi
    
    end_test_suite
}
