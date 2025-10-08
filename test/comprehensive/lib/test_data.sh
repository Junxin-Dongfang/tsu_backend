#!/usr/bin/env bash

################################################################################
# 测试数据管理 - 创建、追踪、清理
################################################################################

# 测试数据前缀
TEST_DATA_PREFIX="[TEST-$(date +%s)]"

# ID 追踪（使用简单变量，兼容旧版 bash）
TEST_ID_class=""
TEST_ID_skill_category=""
TEST_ID_action_category=""
TEST_ID_damage_type=""
TEST_ID_hero_attribute_type=""
TEST_ID_tag=""
TEST_ID_skill=""
TEST_ID_effect=""
TEST_ID_buff=""
TEST_ID_action=""
TEST_ID_action_flag=""
TEST_ID_role=""
TEST_ID_skill_upgrade_cost=""

# 资源清理列表（格式: "type:id"）
CLEANUP_RESOURCES=()

################################################################################
# ID 追踪管理
################################################################################

# 保存测试资源ID
save_test_id() {
    local resource_type="$1"
    local resource_id="$2"
    
    # 使用动态变量名
    eval "TEST_ID_${resource_type}=\"$resource_id\""
    CLEANUP_RESOURCES+=("$resource_type:$resource_id")
    
    log_debug "Saved test ID: $resource_type = $resource_id"
}

# 获取测试资源ID
get_test_id() {
    local resource_type="$1"
    eval "echo \"\$TEST_ID_${resource_type}\""
}

# 追加资源ID到列表（用于批量资源）
append_test_id() {
    local resource_type="$1"
    local resource_id="$2"
    
    local existing=$(get_test_id "$resource_type")
    if [ -z "$existing" ]; then
        eval "TEST_ID_${resource_type}=\"$resource_id\""
    else
        eval "TEST_ID_${resource_type}=\"$existing,$resource_id\""
    fi
    
    CLEANUP_RESOURCES+=("$resource_type:$resource_id")
}

################################################################################
# 通用数据创建工具
################################################################################

# 创建测试职业
create_test_class() {
    local class_name="${1:-${TEST_DATA_PREFIX} 测试职业}"
    local class_name_en="${2:-TestClass}"
    local class_code="TEST_$(date +%s)"
    
    local data="{
        \"class_code\": \"$class_code\",
        \"class_name\": \"$class_name\",
        \"description\": \"自动化测试创建的职业\",
        \"tier\": \"basic\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/classes" "$data" true
    
    if assert_success "创建测试职业"; then
        local class_id=$(extract_field '.data.id')
        save_test_id "class" "$class_id"
        echo "$class_id"
        return 0
    fi
    
    return 1
}

# 创建测试技能分类
create_test_skill_category() {
    local category_code="TEST_SC_$(date +%s)"
    local category_name="${1:-${TEST_DATA_PREFIX} 技能分类}"
    
    local data="{
        \"category_code\": \"$category_code\",
        \"category_name\": \"$category_name\",
        \"description\": \"自动化测试创建的技能分类\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/skill-categories" "$data" true
    
    if assert_success "创建测试技能分类"; then
        local category_id=$(extract_field '.data.id')
        save_test_id "skill_category" "$category_id"
        echo "$category_id"
        return 0
    fi
    
    return 1
}

# 创建测试动作分类
create_test_action_category() {
    local category_code="${1:-TEST_ACTION_CAT}"
    local category_name="${2:-${TEST_DATA_PREFIX} 动作分类}"
    
    local data="{
        \"category_code\": \"$category_code\",
        \"category_name\": \"$category_name\",
        \"description\": \"自动化测试创建的动作分类\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/action-categories" "$data" true
    
    if assert_success "创建测试动作分类"; then
        local category_id=$(extract_field '.data.id')
        save_test_id "action_category" "$category_id"
        echo "$category_id"
        return 0
    fi
    
    return 1
}

# 创建测试伤害类型
create_test_damage_type() {
    local type_code="TEST_DMG_$(date +%s)"
    local type_name="${1:-${TEST_DATA_PREFIX} 伤害类型}"
    
    local data="{
        \"code\": \"$type_code\",
        \"name\": \"$type_name\",
        \"category\": \"physical\",
        \"description\": \"自动化测试创建的伤害类型\",
        \"color\": \"#FF0000\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/damage-types" "$data" true
    
    if assert_success "创建测试伤害类型"; then
        local type_id=$(extract_field '.data.id')
        save_test_id "damage_type" "$type_id"
        echo "$type_id"
        return 0
    fi
    
    return 1
}

# 创建测试英雄属性类型
create_test_hero_attribute_type() {
    local attr_code="${1:-TEST_ATTR_$(date +%s)}"
    local attr_name="${2:-${TEST_DATA_PREFIX} 属性}"
    
    local data="{
        \"attribute_code\": \"$attr_code\",
        \"attribute_name\": \"$attr_name\",
        \"category\": \"derived\",
        \"data_type\": \"integer\",
        \"description\": \"自动化测试创建的属性类型\",
        \"is_active\": true,
        \"is_visible\": true,
        \"display_order\": 100
    }"
    
    http_request "POST" "/api/v1/admin/hero-attribute-types" "$data" true
    
    if assert_success "创建测试英雄属性类型"; then
        local attr_id=$(extract_field '.data.id')
        save_test_id "hero_attribute_type" "$attr_id"
        echo "$attr_id"
        return 0
    fi
    
    return 1
}

# 创建测试标签
create_test_tag() {
    local tag_code="${1:-TEST_TAG_$(date +%s)}"
    local tag_name="${2:-${TEST_DATA_PREFIX} 标签}"
    
    local data="{
        \"tag_code\": \"$tag_code\",
        \"tag_name\": \"$tag_name\",
        \"category\": \"skill\",
        \"description\": \"自动化测试创建的标签\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/tags" "$data" true
    
    if assert_success "创建测试标签"; then
        local tag_id=$(extract_field '.data.id')
        save_test_id "tag" "$tag_id"
        echo "$tag_id"
        return 0
    fi
    
    return 1
}

# 创建测试技能
create_test_skill() {
    local category_id="$1"
    local skill_name="${2:-${TEST_DATA_PREFIX} 技能}"
    local skill_code="TEST_SKILL_$(date +%s)"
    
    if [ -z "$category_id" ]; then
        log_error "技能分类ID不能为空"
        return 1
    fi
    
    local data="{
        \"skill_code\": \"$skill_code\",
        \"skill_name\": \"$skill_name\",
        \"skill_type\": \"active\",
        \"category_id\": \"$category_id\",
        \"description\": \"自动化测试创建的技能\",
        \"max_level\": 5,
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/skills" "$data" true
    
    if assert_success "创建测试技能"; then
        local skill_id=$(extract_field '.data.id')
        save_test_id "skill" "$skill_id"
        echo "$skill_id"
        return 0
    fi
    
    return 1
}

# 创建测试效果
create_test_effect() {
    local effect_name="${1:-${TEST_DATA_PREFIX} 效果}"
    
    # 先获取一个效果类型定义
    http_request "GET" "/api/v1/admin/metadata/effect-type-definitions/all" "" true
    local effect_type_id=$(extract_field '.data[0].id')
    
    if [ -z "$effect_type_id" ] || [ "$effect_type_id" = "null" ]; then
        log_warning "未找到效果类型定义，使用默认值"
        effect_type_id="1"
    fi
    
    local effect_code="TEST_EFFECT_$(date +%s)"
    
    local data="{
        \"effect_code\": \"$effect_code\",
        \"effect_name\": \"$effect_name\",
        \"effect_type\": \"damage\",
        \"parameters\": \"{\\\"damage\\\": 100}\",
        \"description\": \"自动化测试创建的效果\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/effects" "$data" true
    
    if assert_success "创建测试效果"; then
        local effect_id=$(extract_field '.data.id')
        save_test_id "effect" "$effect_id"
        echo "$effect_id"
        return 0
    fi
    
    return 1
}

# 创建测试 Buff
create_test_buff() {
    local buff_name="${1:-${TEST_DATA_PREFIX} Buff}"
    local buff_code="TEST_BUFF_$(date +%s)"
    
    local data="{
        \"buff_code\": \"$buff_code\",
        \"buff_name\": \"$buff_name\",
        \"buff_type\": \"positive\",
        \"description\": \"自动化测试创建的Buff\",
        \"is_stackable\": false,
        \"max_stacks\": 1,
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/buffs" "$data" true
    
    if assert_success "创建测试Buff"; then
        local buff_id=$(extract_field '.data.id')
        save_test_id "buff" "$buff_id"
        echo "$buff_id"
        return 0
    fi
    
    return 1
}

# 创建测试动作
create_test_action() {
    local category_id="$1"
    local action_code="${2:-TEST_ACTION_$(date +%s)}"
    local action_name="${3:-${TEST_DATA_PREFIX} 动作}"
    
    if [ -z "$category_id" ]; then
        log_error "动作分类ID不能为空"
        return 1
    fi
    
    local data="{
        \"action_code\": \"$action_code\",
        \"action_name\": \"$action_name\",
        \"action_category_id\": \"$category_id\",
        \"action_type\": \"attack\",
        \"description\": \"自动化测试创建的动作\",
        \"action_point_cost\": 1,
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/actions" "$data" true
    
    if assert_success "创建测试动作"; then
        local action_id=$(extract_field '.data.id')
        save_test_id "action" "$action_id"
        echo "$action_id"
        return 0
    fi
    
    return 1
}

# 创建测试动作标记
create_test_action_flag() {
    local flag_code="${1:-TEST_FLAG_$(date +%s)}"
    local flag_name="${2:-${TEST_DATA_PREFIX} 动作标记}"
    
    local data="{
        \"flag_code\": \"$flag_code\",
        \"flag_name\": \"$flag_name\",
        \"description\": \"自动化测试创建的动作标记\",
        \"category\": \"test\",
        \"duration_type\": \"instant\",
        \"is_active\": true
    }"
    
    http_request "POST" "/api/v1/admin/action-flags" "$data" true
    
    if assert_success "创建测试动作标记"; then
        local flag_id=$(extract_field '.data.id')
        save_test_id "action_flag" "$flag_id"
        echo "$flag_id"
        return 0
    fi
    
    return 1
}

################################################################################
# 数据清理
################################################################################

# 清理单个资源
cleanup_resource() {
    local resource_type="$1"
    local resource_id="$2"
    
    if [ -z "$resource_id" ] || [ "$resource_id" = "null" ]; then
        return 0
    fi
    
    log_debug "Cleaning up $resource_type: $resource_id"
    
    case "$resource_type" in
        "class")
            http_request "DELETE" "/api/v1/admin/classes/$resource_id" "" true
            ;;
        "skill_category")
            http_request "DELETE" "/api/v1/admin/skill-categories/$resource_id" "" true
            ;;
        "action_category")
            http_request "DELETE" "/api/v1/admin/action-categories/$resource_id" "" true
            ;;
        "damage_type")
            http_request "DELETE" "/api/v1/admin/damage-types/$resource_id" "" true
            ;;
        "hero_attribute_type")
            http_request "DELETE" "/api/v1/admin/hero-attribute-types/$resource_id" "" true
            ;;
        "tag")
            http_request "DELETE" "/api/v1/admin/tags/$resource_id" "" true
            ;;
        "skill")
            http_request "DELETE" "/api/v1/admin/skills/$resource_id" "" true
            ;;
        "effect")
            http_request "DELETE" "/api/v1/admin/effects/$resource_id" "" true
            ;;
        "buff")
            http_request "DELETE" "/api/v1/admin/buffs/$resource_id" "" true
            ;;
        "action")
            http_request "DELETE" "/api/v1/admin/actions/$resource_id" "" true
            ;;
        "action_flag")
            http_request "DELETE" "/api/v1/admin/action-flags/$resource_id" "" true
            ;;
        *)
            log_warning "Unknown resource type: $resource_type"
            ;;
    esac
}

# 清理所有测试数据
cleanup_all_test_data() {
    if [ "$NO_CLEANUP" = "true" ]; then
        log_info "跳过数据清理（--no-cleanup 已设置）"
        return 0
    fi
    
    log_info "开始清理测试数据..."
    
    local cleanup_count=0
    # 反向遍历清理（后创建的先删除）
    for ((i=${#CLEANUP_RESOURCES[@]}-1; i>=0; i--)); do
        local resource="${CLEANUP_RESOURCES[i]}"
        local type="${resource%%:*}"
        local id="${resource##*:}"
        
        cleanup_resource "$type" "$id"
        cleanup_count=$((cleanup_count + 1))
    done
    
    log_success "清理完成，共清理 $cleanup_count 个资源"
}

# 保存测试数据快照
save_test_data_snapshot() {
    local snapshot_file="$1"
    
    log_debug "Saving test data snapshot to $snapshot_file"
    
    # 创建 JSON 格式的快照
    {
        echo "{"
        echo "  \"timestamp\": \"$(date '+%Y-%m-%d %H:%M:%S')\","
        echo "  \"test_ids\": {"
        echo "    \"class\": \"$TEST_ID_class\","
        echo "    \"skill_category\": \"$TEST_ID_skill_category\","
        echo "    \"action_category\": \"$TEST_ID_action_category\","
        echo "    \"damage_type\": \"$TEST_ID_damage_type\","
        echo "    \"hero_attribute_type\": \"$TEST_ID_hero_attribute_type\","
        echo "    \"tag\": \"$TEST_ID_tag\","
        echo "    \"skill\": \"$TEST_ID_skill\","
        echo "    \"effect\": \"$TEST_ID_effect\","
        echo "    \"buff\": \"$TEST_ID_buff\","
        echo "    \"action\": \"$TEST_ID_action\","
        echo "    \"action_flag\": \"$TEST_ID_action_flag\","
        echo "    \"role\": \"$TEST_ID_role\""
        echo "  },"
        echo "  \"cleanup_resources\": ["
        
        local first=true
        for resource in "${CLEANUP_RESOURCES[@]}"; do
            if [ "$first" = true ]; then
                first=false
            else
                echo ","
            fi
            echo -n "    \"$resource\""
        done
        
        echo ""
        echo "  ]"
        echo "}"
    } > "$snapshot_file"
}
