#!/bin/bash

################################################################################
# 测试套件 06: 元数据定义（只读配置）
################################################################################

test_metadata() {
    start_test_suite "元数据定义"
    
    # ===== 效果类型定义 =====
    
    test_case "获取效果类型定义列表（分页）"
    http_request "GET" "/api/v1/admin/metadata/effect-type-definitions?page=1&page_size=10" "" true
    if assert_success "获取效果类型定义列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "效果类型定义总数: $total"
    fi
    
    test_case "获取所有效果类型定义"
    http_request "GET" "/api/v1/admin/metadata/effect-type-definitions/all" "" true
    if assert_success "获取所有效果类型定义成功"; then
        local count=$(get_array_length ".data")
        log_info "效果类型定义数量: $count"
        
        # 如果有数据，测试详情接口
        if [ "$count" -gt 0 ]; then
            local first_id=$(extract_field ".data[0].id")
            if [ -n "$first_id" ] && [ "$first_id" != "null" ]; then
                test_case "获取效果类型定义详情"
                http_request "GET" "/api/v1/admin/metadata/effect-type-definitions/$first_id" "" true
                assert_success "获取效果类型定义详情成功"
            fi
        fi
    fi
    
    # ===== 公式变量 =====
    
    test_case "获取公式变量列表（分页）"
    http_request "GET" "/api/v1/admin/metadata/formula-variables?page=1&page_size=10" "" true
    if assert_success "获取公式变量列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "公式变量总数: $total"
    fi
    
    test_case "获取所有公式变量"
    http_request "GET" "/api/v1/admin/metadata/formula-variables/all" "" true
    if assert_success "获取所有公式变量成功"; then
        local count=$(get_array_length ".data")
        log_info "公式变量数量: $count"
        
        if [ "$count" -gt 0 ]; then
            local first_id=$(extract_field ".data[0].id")
            if [ -n "$first_id" ] && [ "$first_id" != "null" ]; then
                test_case "获取公式变量详情"
                http_request "GET" "/api/v1/admin/metadata/formula-variables/$first_id" "" true
                assert_success "获取公式变量详情成功"
            fi
        fi
    fi
    
    # ===== 范围配置规则 =====
    
    test_case "获取范围配置规则列表（分页）"
    http_request "GET" "/api/v1/admin/metadata/range-config-rules?page=1&page_size=10" "" true
    if assert_success "获取范围配置规则列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "范围配置规则总数: $total"
    fi
    
    test_case "获取所有范围配置规则"
    http_request "GET" "/api/v1/admin/metadata/range-config-rules/all" "" true
    if assert_success "获取所有范围配置规则成功"; then
        local count=$(get_array_length ".data")
        log_info "范围配置规则数量: $count"
        
        if [ "$count" -gt 0 ]; then
            local first_id=$(extract_field ".data[0].id")
            if [ -n "$first_id" ] && [ "$first_id" != "null" ]; then
                test_case "获取范围配置规则详情"
                http_request "GET" "/api/v1/admin/metadata/range-config-rules/$first_id" "" true
                assert_success "获取范围配置规则详情成功"
            fi
        fi
    fi
    
    # ===== 动作类型定义 =====
    
    test_case "获取动作类型定义列表（分页）"
    http_request "GET" "/api/v1/admin/metadata/action-type-definitions?page=1&page_size=10" "" true
    if assert_success "获取动作类型定义列表成功"; then
        validate_pagination_response
        local total=$(extract_field ".data.total")
        log_info "动作类型定义总数: $total"
    fi
    
    test_case "获取所有动作类型定义"
    http_request "GET" "/api/v1/admin/metadata/action-type-definitions/all" "" true
    if assert_success "获取所有动作类型定义成功"; then
        local count=$(get_array_length ".data")
        log_info "动作类型定义数量: $count"
        
        if [ "$count" -gt 0 ]; then
            local first_id=$(extract_field ".data[0].id")
            if [ -n "$first_id" ] && [ "$first_id" != "null" ]; then
                test_case "获取动作类型定义详情"
                http_request "GET" "/api/v1/admin/metadata/action-type-definitions/$first_id" "" true
                assert_success "获取动作类型定义详情成功"
            fi
        fi
    fi
    
    end_test_suite
}
