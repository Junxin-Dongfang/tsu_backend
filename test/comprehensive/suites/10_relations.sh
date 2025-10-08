#!/bin/bash

################################################################################
# 测试套件 10: 关联关系（标签、职业属性加成、职业进阶）
################################################################################

test_relations() {
    start_test_suite "关联关系系统"
    
    # ===== 标签关联 =====
    
    # 确保有标签和技能可用
    local tag_id=$(get_test_id "tag")
    if [ -z "$tag_id" ] || [ "$tag_id" = "null" ]; then
        log_info "创建测试标签用于关联测试"
        tag_id=$(create_test_tag)
    fi
    
    local skill_id=$(get_test_id "skill")
    
    if [ -n "$tag_id" ] && [ "$tag_id" != "null" ] && [ -n "$skill_id" ] && [ "$skill_id" != "null" ]; then
        test_case "获取标签关联的实体"
        http_request "GET" "/api/v1/admin/tags/$tag_id/entities" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取标签关联的实体成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取标签关联的实体失败"
        fi
        
        test_case "获取实体的标签"
        http_request "GET" "/api/v1/admin/entities/skill/$skill_id/tags" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取实体的标签成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取实体的标签失败"
        fi
        
        test_case "为实体添加标签"
        local add_tag_data="{\"tag_id\": \"$tag_id\"}"
        http_request "POST" "/api/v1/admin/entities/skill/$skill_id/tags" "$add_tag_data" true
        
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 为实体添加标签成功"
            
            # 验证标签已添加
            test_case "验证标签已添加"
            http_request "GET" "/api/v1/admin/entities/skill/$skill_id/tags" "" true
            if [ "$LAST_HTTP_CODE" = "200" ]; then
                local tag_count=$(get_array_length ".data")
                log_info "实体关联的标签数量: $tag_count"
                if [ "$tag_count" -gt 0 ]; then
                    SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                    log_success "[$TEST_CASE_NUMBER] 标签关联已建立"
                else
                    SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                    log_error "[$TEST_CASE_NUMBER] 标签关联未找到"
                fi
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 验证标签关联失败"
            fi
            
            # 测试批量设置标签
            test_case "批量设置实体标签"
            local batch_tag_data="{\"tag_ids\": [\"$tag_id\"]}"
            http_request "POST" "/api/v1/admin/entities/skill/$skill_id/tags/batch" "$batch_tag_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 批量设置实体标签成功"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_warning "[$TEST_CASE_NUMBER] 批量设置实体标签失败"
            fi
            
            # 移除标签
            test_case "移除实体标签"
            http_request "DELETE" "/api/v1/admin/entities/skill/$skill_id/tags/$tag_id" "" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 移除实体标签成功"
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_error "[$TEST_CASE_NUMBER] 移除实体标签失败"
            fi
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 为实体添加标签失败"
        fi
    else
        log_warning "跳过标签关联测试：未找到必需的资源"
    fi
    
    # ===== 职业属性加成 =====
    
    local class_id=$(get_test_id "class")
    local hero_attr_id=$(get_test_id "hero_attribute_type")
    
    if [ -n "$class_id" ] && [ "$class_id" != "null" ]; then
        test_case "获取职业属性加成列表"
        http_request "GET" "/api/v1/admin/classes/$class_id/attribute-bonuses" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取职业属性加成列表成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取职业属性加成列表失败"
        fi
        
        if [ -n "$hero_attr_id" ] && [ "$hero_attr_id" != "null" ]; then
            test_case "创建职业属性加成"
            local bonus_data="{
                \"attribute_type_id\": \"$hero_attr_id\",
                \"bonus_value\": 10,
                \"bonus_type\": \"fixed\"
            }"
            http_request "POST" "/api/v1/admin/classes/$class_id/attribute-bonuses" "$bonus_data" true
            
            if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ]; then
                SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                log_success "[$TEST_CASE_NUMBER] 创建职业属性加成成功"
                
                local bonus_id=$(extract_field '.data.id')
                if [ -n "$bonus_id" ] && [ "$bonus_id" != "null" ]; then
                    # 测试批量设置
                    test_case "批量设置职业属性加成"
                    local batch_bonus_data="{
                        \"bonuses\": [
                            {
                                \"attribute_type_id\": \"$hero_attr_id\",
                                \"bonus_value\": 15,
                                \"bonus_type\": \"fixed\"
                            }
                        ]
                    }"
                    http_request "POST" "/api/v1/admin/classes/$class_id/attribute-bonuses/batch" "$batch_bonus_data" true
                    
                    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "204" ]; then
                        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
                        log_success "[$TEST_CASE_NUMBER] 批量设置职业属性加成成功"
                    else
                        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                        log_warning "[$TEST_CASE_NUMBER] 批量设置职业属性加成失败"
                    fi
                fi
            else
                SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
                log_warning "[$TEST_CASE_NUMBER] 创建职业属性加成失败（可能已存在）"
            fi
        else
            log_warning "跳过职业属性加成创建：未找到属性类型"
        fi
    else
        log_warning "跳过职业属性加成测试：未找到职业"
    fi
    
    # ===== 职业进阶路径 =====
    
    if [ -n "$class_id" ] && [ "$class_id" != "null" ]; then
        test_case "获取职业进阶信息"
        http_request "GET" "/api/v1/admin/classes/$class_id/advancement" "" true
        if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "404" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取职业进阶信息成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取职业进阶信息失败"
        fi
        
        test_case "获取职业进阶路径"
        http_request "GET" "/api/v1/admin/classes/$class_id/advancement-paths" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取职业进阶路径成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取职业进阶路径失败"
        fi
        
        test_case "获取职业进阶来源"
        http_request "GET" "/api/v1/admin/classes/$class_id/advancement-sources" "" true
        if [ "$LAST_HTTP_CODE" = "200" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] 获取职业进阶来源成功"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_error "[$TEST_CASE_NUMBER] 获取职业进阶来源失败"
        fi
    fi
    
    # ===== 职业进阶要求管理 =====
    
    test_case "获取职业进阶要求列表"
    http_request "GET" "/api/v1/admin/advancement-requirements?page=1&page_size=10" "" true
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 获取职业进阶要求列表成功"
        validate_pagination_response
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 获取职业进阶要求列表失败"
    fi
    
    end_test_suite
}
