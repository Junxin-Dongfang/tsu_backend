#!/bin/bash

################################################################################
# 测试套件 07: 技能系统
################################################################################

test_skill_system() {
    start_test_suite "技能系统"
    
    # 确保有技能分类
    local skill_category_id=$(get_test_id "skill_category")
    if [ -z "$skill_category_id" ] || [ "$skill_category_id" = "null" ]; then
        log_info "未找到已创建的技能分类，创建新的"
        skill_category_id=$(create_test_skill_category)
    fi
    
    if [ -z "$skill_category_id" ] || [ "$skill_category_id" = "null" ]; then
        log_error "无法获取技能分类 ID，跳过技能系统测试"
        end_test_suite
        return 1
    fi
    
    # ===== 技能管理 =====
    
    test_case "获取技能列表"
    http_request "GET" "/api/v1/admin/skills?page=1&page_size=10" "" true
    if assert_success "获取技能列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试技能"
    local skill_id=$(create_test_skill "$skill_category_id")
    if [ -n "$skill_id" ] && [ "$skill_id" != "null" ]; then
        log_info "创建的技能 ID: $skill_id"
        
        test_case "获取技能详情"
        http_request "GET" "/api/v1/admin/skills/$skill_id" "" true
        if assert_success "获取技能详情成功"; then
            assert_field_exists ".data.id"
            assert_field_exists ".data.name"
            assert_field_exists ".data.skill_category_id"
        fi
        
        test_case "更新技能"
        local update_data="{
            \"name\": \"${TEST_DATA_PREFIX} 技能 Updated\",
            \"description\": \"已更新的技能\",
            \"max_level\": 10
        }"
        http_request "PUT" "/api/v1/admin/skills/$skill_id" "$update_data" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新技能成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新技能失败"
        fi
        
        # ===== 技能升级消耗 =====
        
        test_case "获取技能升级消耗列表"
        http_request "GET" "/api/v1/admin/skill-upgrade-costs?page=1&page_size=10" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取技能升级消耗列表成功"
            validate_pagination_response
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取技能升级消耗列表失败"
        fi
        
        test_case "创建技能升级消耗"
        local cost_data="{
            \"level\": 1,
            \"gold_cost\": 100,
            \"experience_cost\": 50,
            \"description\": \"测试升级消耗\"
        }"
        http_request "POST" "/api/v1/admin/skill-upgrade-costs" "$cost_data" true
        
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 创建技能升级消耗成功"
            
            local cost_id=$(extract_field '.data.id')
            if [ -n "$cost_id" ] && [ "$cost_id" != "null" ]; then
                save_test_id "skill_upgrade_cost" "$cost_id"
                
                test_case "获取技能升级消耗详情"
                http_request "GET" "/api/v1/admin/skill-upgrade-costs/$cost_id" "" true
                assert_success "获取技能升级消耗详情成功"
                
                test_case "按等级查询升级消耗"
                http_request "GET" "/api/v1/admin/skill-upgrade-costs/level/1" "" true
                if [ "$LAST_HTTP_CODE" = "200" ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 按等级查询升级消耗成功"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 按等级查询升级消耗失败"
                fi
            fi
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_warning "[$TEST_CASE_NUMBER] 创建技能升级消耗失败（可能已存在）"
        fi
        
        # ===== 技能解锁动作（稍后测试，需要先有动作）=====
        
        test_case "获取技能解锁动作列表"
        http_request "GET" "/api/v1/admin/skills/$skill_id/unlock-actions" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取技能解锁动作列表成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取技能解锁动作列表失败"
        fi
        
    else
        log_error "创建技能失败，跳过后续测试"
    fi
    
    # ===== 技能搜索和过滤 =====
    
    test_case "按分类搜索技能"
    http_request "GET" "/api/v1/admin/skills?page=1&page_size=10&category_id=$skill_category_id" "" true
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 按分类搜索技能成功"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 按分类搜索技能失败"
    fi
    
    end_test_suite
}
