#!/bin/bash

################################################################################
# 测试套件 05: 基础游戏配置
################################################################################

test_game_config_basic() {
    start_test_suite "基础游戏配置"
    
    # ===== 职业管理 =====
    
    test_case "获取职业列表"
    http_request "GET" "/api/v1/admin/classes?page=1&page_size=10" "" true
    if assert_success "获取职业列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试职业"
    local class_id=$(create_test_class)
    if [ -n "$class_id" ] && [ "$class_id" != "null" ]; then
        log_info "创建的职业 ID: $class_id"
        
        test_case "获取职业详情"
        http_request "GET" "/api/v1/admin/classes/$class_id" "" true
        assert_success "获取职业详情成功"
        
        test_case "更新职业"
        local update_data='{"name":"测试职业 Updated","description":"已更新"}'
        http_request "PUT" "/api/v1/admin/classes/$class_id" "$update_data" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 更新职业成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 更新职业失败"
        fi
    fi
    
    # ===== 技能分类管理 =====
    
    test_case "获取技能分类列表"
    http_request "GET" "/api/v1/admin/skill-categories?page=1&page_size=10" "" true
    if assert_success "获取技能分类列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试技能分类"
    local skill_category_id=$(create_test_skill_category)
    if [ -n "$skill_category_id" ] && [ "$skill_category_id" != "null" ]; then
        log_info "创建的技能分类 ID: $skill_category_id"
        
        test_case "获取技能分类详情"
        http_request "GET" "/api/v1/admin/skill-categories/$skill_category_id" "" true
        assert_success "获取技能分类详情成功"
    fi
    
    # ===== 动作分类管理 =====
    
    test_case "获取动作分类列表"
    http_request "GET" "/api/v1/admin/action-categories?page=1&page_size=10" "" true
    if assert_success "获取动作分类列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试动作分类"
    local action_category_id=$(create_test_action_category)
    if [ -n "$action_category_id" ] && [ "$action_category_id" != "null" ]; then
        log_info "创建的动作分类 ID: $action_category_id"
        
        test_case "获取动作分类详情"
        http_request "GET" "/api/v1/admin/action-categories/$action_category_id" "" true
        assert_success "获取动作分类详情成功"
    fi
    
    # ===== 伤害类型管理 =====
    
    test_case "获取伤害类型列表"
    http_request "GET" "/api/v1/admin/damage-types?page=1&page_size=10" "" true
    if assert_success "获取伤害类型列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试伤害类型"
    local damage_type_id=$(create_test_damage_type)
    if [ -n "$damage_type_id" ] && [ "$damage_type_id" != "null" ]; then
        log_info "创建的伤害类型 ID: $damage_type_id"
        
        test_case "获取伤害类型详情"
        http_request "GET" "/api/v1/admin/damage-types/$damage_type_id" "" true
        assert_success "获取伤害类型详情成功"
    fi
    
    # ===== 英雄属性类型管理 =====
    
    test_case "获取英雄属性类型列表"
    http_request "GET" "/api/v1/admin/hero-attribute-types?page=1&page_size=10" "" true
    if assert_success "获取英雄属性类型列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试英雄属性类型"
    local attr_code="TEST_ATTR_$(date +%s)"
    local hero_attr_id=$(create_test_hero_attribute_type "$attr_code")
    if [ -n "$hero_attr_id" ] && [ "$hero_attr_id" != "null" ]; then
        log_info "创建的英雄属性类型 ID: $hero_attr_id"
        
        test_case "获取英雄属性类型详情"
        http_request "GET" "/api/v1/admin/hero-attribute-types/$hero_attr_id" "" true
        assert_success "获取英雄属性类型详情成功"
    fi
    
    # ===== 标签管理 =====
    
    test_case "获取标签列表"
    http_request "GET" "/api/v1/admin/tags?page=1&page_size=10" "" true
    if assert_success "获取标签列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试标签"
    local tag_id=$(create_test_tag)
    if [ -n "$tag_id" ] && [ "$tag_id" != "null" ]; then
        log_info "创建的标签 ID: $tag_id"
        
        test_case "获取标签详情"
        http_request "GET" "/api/v1/admin/tags/$tag_id" "" true
        assert_success "获取标签详情成功"
    fi
    
    # ===== 动作标记管理 =====
    
    test_case "获取动作标记列表"
    http_request "GET" "/api/v1/admin/action-flags?page=1&page_size=10" "" true
    if assert_success "获取动作标记列表成功"; then
        validate_pagination_response
    fi
    
    test_case "创建测试动作标记"
    local flag_id=$(create_test_action_flag)
    if [ -n "$flag_id" ] && [ "$flag_id" != "null" ]; then
        log_info "创建的动作标记 ID: $flag_id"
        
        test_case "获取动作标记详情"
        http_request "GET" "/api/v1/admin/action-flags/$flag_id" "" true
        assert_success "获取动作标记详情成功"
    fi
    
    end_test_suite
}
