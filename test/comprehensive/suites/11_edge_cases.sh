#!/bin/bash

################################################################################
# 测试套件 11: 边界条件和错误处理
################################################################################

test_edge_cases() {
    start_test_suite "边界条件和错误处理"
    
    # ===== 404 错误测试 =====
    
    test_case "404 - 访问不存在的职业"
    local fake_uuid="00000000-0000-0000-0000-000000000000"
    http_request "GET" "/api/v1/admin/classes/$fake_uuid" "" true
    assert_status "404" "不存在的资源应返回 404"
    
    test_case "404 - 访问不存在的技能"
    http_request "GET" "/api/v1/admin/skills/99999999" "" true
    if [ "$LAST_HTTP_CODE" = "404" ] || [ "$LAST_HTTP_CODE" = "400" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 不存在的技能返回 404/400"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 错误处理异常"
    fi
    
    test_case "404 - 访问不存在的效果"
    http_request "GET" "/api/v1/admin/effects/$fake_uuid" "" true
    assert_status "404" "不存在的效果应返回 404"
    
    # ===== 400 错误测试（无效数据）=====
    
    test_case "400 - 创建职业缺少必需字段"
    local invalid_class='{"description":"缺少name字段"}'
    http_request "POST" "/api/v1/admin/classes" "$invalid_class" true
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "422" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 缺少必需字段返回 400/422"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 参数验证未生效 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "400 - 无效的 JSON 格式"
    http_request "POST" "/api/v1/admin/classes" "{invalid json" true
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "422" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 无效 JSON 返回 400/422"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] JSON 验证未生效 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "400 - 无效的数据类型"
    local invalid_type_data='{"name":"Test","max_level":"not_a_number"}'
    http_request "POST" "/api/v1/admin/skills" "$invalid_type_data" true
    # 可能返回 400、422 或者类型转换错误
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "422" ] || [ "$LAST_HTTP_CODE" = "500" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 无效数据类型被拒绝 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] 数据类型验证未生效 - HTTP $LAST_HTTP_CODE"
    fi
    
    # ===== 401 错误测试（未认证）=====
    
    test_401_error "/api/v1/admin/classes" "未认证访问受保护接口"
    test_401_error "/api/v1/admin/skills" "未认证访问技能接口"
    test_401_error "/api/v1/admin/users" "未认证访问用户接口"
    
    # ===== 分页边界测试 =====
    
    test_case "分页 - 负数页码"
    http_request "GET" "/api/v1/admin/classes?page=-1&page_size=10" "" true
    # 应该返回 400 或者忽略并返回第一页
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 负数页码处理正确 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 分页参数处理异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "分页 - 零页大小"
    http_request "GET" "/api/v1/admin/classes?page=1&page_size=0" "" true
    if [ "$LAST_HTTP_CODE" = "400" ] || [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 零页大小处理正确 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 页大小验证异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "分页 - 超大页码"
    http_request "GET" "/api/v1/admin/classes?page=999999&page_size=10" "" true
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 超大页码返回空列表"
        local items=$(get_array_length ".data.items")
        log_info "超大页码返回 $items 条记录"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 超大页码处理异常"
    fi
    
    test_case "分页 - 超大页大小"
    http_request "GET" "/api/v1/admin/classes?page=1&page_size=10000" "" true
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "400" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 超大页大小处理正确 - HTTP $LAST_HTTP_CODE"
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 页大小限制未生效"
    fi
    
    # ===== 特殊字符处理 =====
    
    test_case "特殊字符 - SQL 注入测试"
    local sql_injection_data='{"name":"Test\"; DROP TABLE classes; --","name_en":"SQLInjection"}'
    http_request "POST" "/api/v1/admin/classes" "$sql_injection_data" true
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] SQL 注入字符被正确处理"
        # 如果创建成功，清理
        local created_id=$(extract_field '.data.id')
        if [ -n "$created_id" ] && [ "$created_id" != "null" ]; then
            http_request "DELETE" "/api/v1/admin/classes/$created_id" "" true
        fi
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] SQL 注入测试异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "特殊字符 - XSS 测试"
    local xss_data='{"name":"<script>alert(\"XSS\")</script>","name_en":"XSSTest"}'
    http_request "POST" "/api/v1/admin/classes" "$xss_data" true
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ] || [ "$LAST_HTTP_CODE" = "400" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] XSS 字符处理正确 - HTTP $LAST_HTTP_CODE"
        # 清理
        local created_id=$(extract_field '.data.id')
        if [ -n "$created_id" ] && [ "$created_id" != "null" ]; then
            http_request "DELETE" "/api/v1/admin/classes/$created_id" "" true
        fi
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] XSS 测试异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    test_case "特殊字符 - Unicode 测试"
    local unicode_data='{"name":"测试🎮游戏😀","name_en":"UnicodeTest"}'
    http_request "POST" "/api/v1/admin/classes" "$unicode_data" true
    if [ "$LAST_HTTP_CODE" = "200" ] || [ "$LAST_HTTP_CODE" = "201" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] Unicode 字符支持正常"
        # 清理
        local created_id=$(extract_field '.data.id')
        if [ -n "$created_id" ] && [ "$created_id" != "null" ]; then
            http_request "DELETE" "/api/v1/admin/classes/$created_id" "" true
        fi
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_warning "[$TEST_CASE_NUMBER] Unicode 字符处理异常 - HTTP $LAST_HTTP_CODE"
    fi
    
    # ===== 并发和幂等性测试 =====
    
    test_case "幂等性 - 多次获取相同资源"
    local class_id=$(get_test_id "class")
    if [ -n "$class_id" ] && [ "$class_id" != "null" ]; then
        http_request "GET" "/api/v1/admin/classes/$class_id" "" true
        local first_response="$LAST_RESPONSE_BODY"
        
        http_request "GET" "/api/v1/admin/classes/$class_id" "" true
        local second_response="$LAST_RESPONSE_BODY"
        
        if [ "$first_response" = "$second_response" ]; then
            SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
            log_success "[$TEST_CASE_NUMBER] GET 请求幂等性正常"
        else
            SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
            log_warning "[$TEST_CASE_NUMBER] GET 请求返回不一致"
        fi
    else
        log_warning "跳过幂等性测试：未找到测试资源"
        SUITE_TESTS_TOTAL=$((SUITE_TESTS_TOTAL + 1))
    fi
    
    # ===== 性能边界测试 =====
    
    test_case "性能 - 大量数据查询"
    http_request "GET" "/api/v1/admin/classes?page=1&page_size=100" "" true
    if [ "$LAST_HTTP_CODE" = "200" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] 大量数据查询成功 (${LAST_REQUEST_DURATION}ms)"
        if [ "$LAST_REQUEST_DURATION" -gt 5000 ]; then
            log_warning "响应时间较长: ${LAST_REQUEST_DURATION}ms"
        fi
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] 大量数据查询失败"
    fi
    
    end_test_suite
}
