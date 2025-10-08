#!/bin/bash

################################################################################
# Admin API 接口自动化测试脚本
# 用途: 验证 Admin 服务所有接口的功能
# 使用: ./admin-api-test.sh [options]
################################################################################

set -e  # 遇到错误立即退出

# ==================== 配置区 ====================
BASE_URL="${BASE_URL:-http://localhost:80}"
API_BASE="/api/v1"
USERNAME="${TEST_USERNAME:-root}"
PASSWORD="${TEST_PASSWORD:-password}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Token 存储
AUTH_TOKEN=""
TEST_LOG_FILE="test_results_$(date +%s)/test_log.txt"

# ==================== 工具函数 ====================

# 打印彩色信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1" | tee -a "$TEST_LOG_FILE"
    ((PASSED_TESTS++))
}

print_error() {
    echo -e "${RED}[✗ FAIL]${NC} $1" | tee -a "$TEST_LOG_FILE"
    ((FAILED_TESTS++))
}

print_skip() {
    echo -e "${YELLOW}[⊘ SKIP]${NC} $1" | tee -a "$TEST_LOG_FILE"
    ((SKIPPED_TESTS++))
}

print_header() {
    echo -e "\n${BLUE}========================================${NC}" | tee -a "$TEST_LOG_FILE"
    echo -e "${BLUE}$1${NC}" | tee -a "$TEST_LOG_FILE"
    echo -e "${BLUE}========================================${NC}\n" | tee -a "$TEST_LOG_FILE"
}

# 初始化测试环境
init_test_env() {
    mkdir -p "$(dirname "$TEST_LOG_FILE")"
    print_header "Admin API 接口自动化测试"
    print_info "测试开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    print_info "API 地址: $BASE_URL"
    print_info "测试账号: $USERNAME"
    print_info ""
}

# HTTP 请求封装
http_get() {
    local url="$1"
    local desc="$2"
    local auth_required="${3:-true}"
    
    ((TOTAL_TESTS++))
    
    if [ "$auth_required" = "true" ] && [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    if [ "$auth_required" = "true" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$url" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$url" \
            -H "Content-Type: application/json")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        print_error "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
}

http_post() {
    local url="$1"
    local data="$2"
    local desc="$3"
    local auth_required="${4:-true}"
    
    ((TOTAL_TESTS++))
    
    if [ "$auth_required" = "true" ] && [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    if [ "$auth_required" = "true" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL$url" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        print_error "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
}

http_put() {
    local url="$1"
    local data="$2"
    local desc="$3"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        print_error "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
}

http_delete() {
    local url="$1"
    local desc="$2"
    
    ((TOTAL_TESTS++))
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "$desc - Token 未设置"
        return 1
    fi
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL$url" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        print_error "$desc - HTTP $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
}

# ==================== 测试用例 ====================

# 1. 系统健康检查
test_system_health() {
    print_header "1. 系统健康检查"
    
    http_get "/health" "健康检查接口" false
}

# 2. 认证流程测试
test_authentication() {
    print_header "2. 认证流程测试"
    
    # 2.1 用户登录
    local login_data='{"identifier":"'$USERNAME'","password":"'$PASSWORD'"}'
    local response
    response=$(curl -s -X POST "$BASE_URL$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    ((TOTAL_TESTS++))
    
    AUTH_TOKEN=$(echo "$response" | jq -r '.data.session_token // .data.token // .session_token // .token // empty' 2>/dev/null)
    
    if [ -n "$AUTH_TOKEN" ]; then
        print_success "用户登录成功"
        print_info "Token: ${AUTH_TOKEN:0:20}..."
    else
        print_error "用户登录失败"
        echo "$response" | jq '.'
        exit 1
    fi
    
    # 2.2 获取当前用户信息
    http_get "$API_BASE/admin/users/me" "获取当前用户信息"
}

# 3. 用户管理测试
test_user_management() {
    print_header "3. 用户管理测试"
    
    http_get "$API_BASE/admin/users?page=1&page_size=10" "获取用户列表"
    http_get "$API_BASE/admin/users/1" "获取用户详情"
}

# 4. 角色权限测试
test_rbac() {
    print_header "4. 角色权限管理测试"
    
    # 角色管理
    http_get "$API_BASE/admin/roles?page=1&page_size=10" "获取角色列表"
    
    # 权限管理
    http_get "$API_BASE/admin/permissions?page=1&page_size=10" "获取权限列表"
    http_get "$API_BASE/admin/permission-groups" "获取权限组列表"
    
    # 用户角色关联
    http_get "$API_BASE/admin/users/1/roles" "获取用户角色"
    http_get "$API_BASE/admin/users/1/permissions" "获取用户权限"
}

# 5. 基础游戏配置测试
test_basic_game_config() {
    print_header "5. 基础游戏配置测试"
    
    # 职业
    http_get "$API_BASE/admin/classes?page=1&page_size=10" "获取职业列表"
    
    # 技能分类
    http_get "$API_BASE/admin/skill-categories?page=1&page_size=10" "获取技能分类列表"
    
    # 动作分类
    http_get "$API_BASE/admin/action-categories?page=1&page_size=10" "获取动作分类列表"
    
    # 伤害类型
    http_get "$API_BASE/admin/damage-types?page=1&page_size=10" "获取伤害类型列表"
    
    # 英雄属性类型
    http_get "$API_BASE/admin/hero-attribute-types?page=1&page_size=10" "获取英雄属性类型列表"
    
    # 标签
    http_get "$API_BASE/admin/tags?page=1&page_size=10" "获取标签列表"
    
    # 标签关系
    http_get "$API_BASE/admin/tag-relations?page=1&page_size=10" "获取标签关系列表"
    
    # 动作标记
    http_get "$API_BASE/admin/action-flags?page=1&page_size=10" "获取动作标记列表"
}

# 6. 元数据定义测试
test_metadata_definitions() {
    print_header "6. 元数据定义测试"
    
    # 效果类型定义
    http_get "$API_BASE/admin/metadata/effect-type-definitions?page=1&page_size=10" "获取效果类型定义列表"
    http_get "$API_BASE/admin/metadata/effect-type-definitions/all" "获取所有效果类型定义"
    
    # 公式变量
    http_get "$API_BASE/admin/metadata/formula-variables?page=1&page_size=10" "获取公式变量列表"
    http_get "$API_BASE/admin/metadata/formula-variables/all" "获取所有公式变量"
    
    # 范围配置规则
    http_get "$API_BASE/admin/metadata/range-config-rules?page=1&page_size=10" "获取范围配置规则列表"
    http_get "$API_BASE/admin/metadata/range-config-rules/all" "获取所有范围配置规则"
    
    # 动作类型定义
    http_get "$API_BASE/admin/metadata/action-type-definitions?page=1&page_size=10" "获取动作类型定义列表"
    http_get "$API_BASE/admin/metadata/action-type-definitions/all" "获取所有动作类型定义"
}

# 7. 技能系统测试
test_skill_system() {
    print_header "7. 技能系统测试"
    
    # 获取技能列表
    http_get "$API_BASE/admin/skills?page=1&page_size=10" "获取技能列表"
    
    # 如果有技能，测试详情和等级配置
    local skills_response
    skills_response=$(curl -s -X GET "$BASE_URL$API_BASE/admin/skills?page=1&page_size=1" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    local skill_id
    skill_id=$(echo "$skills_response" | jq -r '.data.items[0].id // empty' 2>/dev/null)
    
    if [ -n "$skill_id" ]; then
        http_get "$API_BASE/admin/skills/$skill_id" "获取技能详情 (ID: $skill_id)"
        http_get "$API_BASE/admin/skills/$skill_id/level-configs" "获取技能等级配置 (ID: $skill_id)"
        http_get "$API_BASE/admin/skills/$skill_id/unlock-actions" "获取技能解锁动作 (ID: $skill_id)"
    else
        print_skip "技能详情测试 - 暂无技能数据"
    fi
}

# 8. 效果系统测试
test_effect_system() {
    print_header "8. 效果系统测试"
    
    # Effects
    http_get "$API_BASE/admin/effects?page=1&page_size=10" "获取效果列表"
    
    # Buffs
    http_get "$API_BASE/admin/buffs?page=1&page_size=10" "获取Buff列表"
    
    # 测试 Buff 效果关联
    local buffs_response
    buffs_response=$(curl -s -X GET "$BASE_URL$API_BASE/admin/buffs?page=1&page_size=1" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    local buff_id
    buff_id=$(echo "$buffs_response" | jq -r '.data.items[0].id // empty' 2>/dev/null)
    
    if [ -n "$buff_id" ]; then
        http_get "$API_BASE/admin/buffs/$buff_id" "获取Buff详情 (ID: $buff_id)"
        http_get "$API_BASE/admin/buffs/$buff_id/effects" "获取Buff关联的效果 (ID: $buff_id)"
    else
        print_skip "Buff详情测试 - 暂无Buff数据"
    fi
}

# 9. 动作系统测试
test_action_system() {
    print_header "9. 动作系统测试"
    
    # Actions
    http_get "$API_BASE/admin/actions?page=1&page_size=10" "获取动作列表"
    
    # 测试动作效果关联
    local actions_response
    actions_response=$(curl -s -X GET "$BASE_URL$API_BASE/admin/actions?page=1&page_size=1" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    local action_id
    action_id=$(echo "$actions_response" | jq -r '.data.items[0].id // empty' 2>/dev/null)
    
    if [ -n "$action_id" ]; then
        http_get "$API_BASE/admin/actions/$action_id" "获取动作详情 (ID: $action_id)"
        http_get "$API_BASE/admin/actions/$action_id/effects" "获取动作关联的效果 (ID: $action_id)"
    else
        print_skip "动作详情测试 - 暂无动作数据"
    fi
}

# 10. CRUD 完整流程测试（可选）
test_crud_workflow() {
    print_header "10. CRUD 完整流程测试"
    
    # 示例：测试职业的完整 CRUD
    print_info "测试职业 CRUD 流程..."
    
    # 创建
    local create_data='{"name":"测试职业","name_en":"TestClass","description":"自动化测试创建","is_enabled":true}'
    local create_response
    create_response=$(curl -s -X POST "$BASE_URL$API_BASE/admin/classes" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$create_data")
    
    local created_id
    created_id=$(echo "$create_response" | jq -r '.data.id // empty' 2>/dev/null)
    
    if [ -n "$created_id" ]; then
        ((TOTAL_TESTS++))
        print_success "创建职业成功 (ID: $created_id)"
        
        # 读取
        http_get "$API_BASE/admin/classes/$created_id" "读取新创建的职业"
        
        # 更新
        local update_data='{"name":"更新测试职业","description":"已更新"}'
        http_put "$API_BASE/admin/classes/$created_id" "$update_data" "更新职业"
        
        # 删除
        http_delete "$API_BASE/admin/classes/$created_id" "删除职业"
    else
        ((TOTAL_TESTS++))
        print_error "创建职业失败"
        echo "$create_response" | jq '.'
    fi
}

# 11. 边界条件和错误处理测试
test_error_handling() {
    print_header "11. 边界条件和错误处理测试"
    
    # 测试不存在的资源
    ((TOTAL_TESTS++))
    local response
    local http_code
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$API_BASE/admin/skills/999999" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "404" ]; then
        print_success "404错误处理 - 访问不存在的资源"
    else
        print_error "404错误处理失败 - 预期404，实际$http_code"
    fi
    
    # 测试无效的分页参数
    ((TOTAL_TESTS++))
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$API_BASE/admin/skills?page=-1&page_size=0" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "400" ] || [ "$http_code" = "200" ]; then
        print_success "分页参数验证 - 处理无效参数"
    else
        print_error "分页参数验证失败 - HTTP $http_code"
    fi
    
    # 测试无 Token 访问受保护接口
    ((TOTAL_TESTS++))
    response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$API_BASE/admin/users")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
        print_success "认证验证 - 无Token访问受保护接口被拒绝"
    else
        print_error "认证验证失败 - 预期401/403，实际$http_code"
    fi
}

# ==================== 测试报告 ====================

generate_report() {
    print_header "测试报告"
    
    local pass_rate=0
    if [ "$TOTAL_TESTS" -gt 0 ]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    echo ""
    echo "测试完成时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  测试统计"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  总测试数:   $TOTAL_TESTS"
    echo -e "  ${GREEN}通过:${NC}       $PASSED_TESTS"
    echo -e "  ${RED}失败:${NC}       $FAILED_TESTS"
    echo -e "  ${YELLOW}跳过:${NC}       $SKIPPED_TESTS"
    echo "  通过率:     ${pass_rate}%"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "详细日志: $TEST_LOG_FILE"
    echo ""
    
    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "${RED}⚠ 存在失败的测试用例，请检查日志${NC}"
        return 1
    else
        echo -e "${GREEN}✓ 所有测试通过！${NC}"
        return 0
    fi
}

# ==================== 主程序 ====================

main() {
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        echo "错误: 需要安装 curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo "错误: 需要安装 jq (JSON 处理工具)"
        echo "安装命令: brew install jq (macOS) 或 apt-get install jq (Linux)"
        exit 1
    fi
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --url)
                BASE_URL="$2"
                shift 2
                ;;
            --username)
                USERNAME="$2"
                shift 2
                ;;
            --password)
                PASSWORD="$2"
                shift 2
                ;;
            --quick)
                QUICK_MODE=true
                shift
                ;;
            --help)
                echo "用法: $0 [options]"
                echo ""
                echo "选项:"
                echo "  --url URL          API 基础地址 (默认: http://localhost:80)"
                echo "  --username USER    测试账号 (默认: root)"
                echo "  --password PASS    测试密码 (默认: password)"
                echo "  --quick            快速模式，只测试核心接口"
                echo "  --help             显示帮助信息"
                echo ""
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 初始化
    init_test_env
    
    # 执行测试套件
    test_system_health
    test_authentication
    test_user_management
    test_rbac
    test_basic_game_config
    test_metadata_definitions
    test_skill_system
    test_effect_system
    test_action_system
    
    # 完整测试模式
    if [ "$QUICK_MODE" != "true" ]; then
        test_crud_workflow
        test_error_handling
    fi
    
    # 生成报告
    generate_report
}

# 执行主程序
main "$@"
