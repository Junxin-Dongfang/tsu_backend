#!/bin/bash

################################################################################
# 密码重置接口测试脚本
# 用途: 测试 /auth/password/reset-with-code 接口
# 使用: ./password_reset_test.sh [options]
################################################################################

set -e  # 遇到错误立即退出

# ==================== 配置区 ====================
GAME_BASE_URL="${GAME_BASE_URL:-http://localhost:80}"
ADMIN_BASE_URL="${ADMIN_BASE_URL:-http://localhost:80}"
GAME_API_BASE="/api/v1/game"
ADMIN_API_BASE="/api/v1/admin"

# 测试邮箱配置
TEST_EMAIL="${TEST_EMAIL:-test@example.com}"
NEW_PASSWORD="${NEW_PASSWORD:-NewPassword123}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ==================== 工具函数 ====================

# 打印彩色信息
print_header() {
    echo ""
    echo -e "${CYAN}================================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================================${NC}"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

print_error() {
    echo -e "${RED}[✗ FAIL]${NC} $1"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_separator() {
    echo -e "${CYAN}------------------------------------------------${NC}"
}

# 显示帮助信息
show_help() {
    cat << EOF
密码重置接口测试脚本

用法: $0 [选项]

选项:
  --game-url URL         游戏服务基础URL (默认: http://localhost:80)
  --admin-url URL        管理服务基础URL (默认: http://localhost:80)
  --email EMAIL          测试邮箱地址 (默认: test@example.com)
  --new-password PASS    新密码 (默认: NewPassword123)
  -h, --help            显示此帮助信息

示例:
  $0 --email user@test.com --new-password Secret123
  $0 --game-url http://localhost:8081 --admin-url http://localhost:8071

环境变量:
  GAME_BASE_URL         游戏服务基础URL
  ADMIN_BASE_URL        管理服务基础URL
  TEST_EMAIL            测试邮箱地址
  NEW_PASSWORD          新密码

EOF
}

# ==================== API 测试函数 ====================

# 测试游戏端 - 发起密码恢复
test_game_initiate_recovery() {
    print_header "测试 1: 游戏端 - 发起密码恢复"
    
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/recovery/initiate"
    print_info "请求地址: POST $endpoint"
    print_info "邮箱: $TEST_EMAIL"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 200 ]; then
        if echo "$body" | jq -e '.data.code_sent' > /dev/null 2>&1; then
            print_success "密码恢复请求成功发送"
            return 0
        else
            print_error "响应格式不正确"
            return 1
        fi
    else
        print_error "密码恢复请求失败 (HTTP $http_code)"
        return 1
    fi
}

# 测试管理端 - 发起密码恢复
test_admin_initiate_recovery() {
    print_header "测试 2: 管理端 - 发起密码恢复"
    
    local endpoint="${ADMIN_BASE_URL}${ADMIN_API_BASE}/auth/recovery/initiate"
    print_info "请求地址: POST $endpoint"
    print_info "邮箱: $TEST_EMAIL"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 200 ]; then
        if echo "$body" | jq -e '.data.code_sent' > /dev/null 2>&1; then
            print_success "密码恢复请求成功发送"
            return 0
        else
            print_error "响应格式不正确"
            return 1
        fi
    else
        print_error "密码恢复请求失败 (HTTP $http_code)"
        return 1
    fi
}

# 测试游戏端 - 使用验证码重置密码（完整流程）
test_game_reset_with_code() {
    print_header "测试 3: 游戏端 - 使用验证码重置密码（完整流程）"
    
    # 步骤1: 先发起密码恢复（这会在Redis中存储flow_id）
    print_info "步骤 1/3: 发起密码恢复..."
    local initiate_endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/recovery/initiate"
    local initiate_response
    local initiate_http_code
    
    initiate_response=$(curl -s -w "\n%{http_code}" -X POST "$initiate_endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\"}")
    
    initiate_http_code=$(echo "$initiate_response" | tail -n1)
    local initiate_body=$(echo "$initiate_response" | sed '$d')
    
    if [ "$initiate_http_code" -ne 200 ]; then
        print_error "发起密码恢复失败 (HTTP $initiate_http_code): $initiate_body"
        return 1
    fi
    
    print_success "密码恢复请求已发送"
    
    # 步骤2: 获取验证码
    echo ""
    print_info "步骤 2/3: 获取验证码"
    print_warning "请检查邮箱 $TEST_EMAIL 获取验证码"
    read -p "请输入6位验证码: " verification_code
    
    if [ -z "$verification_code" ]; then
        print_error "验证码不能为空"
        return 1
    fi
    
    # 步骤3: 使用验证码重置密码
    print_info "步骤 3/3: 使用验证码重置密码"
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "邮箱: $TEST_EMAIL"
    print_info "验证码: $verification_code"
    print_info "新密码: $NEW_PASSWORD"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"code\":\"$verification_code\",\"new_password\":\"$NEW_PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 200 ]; then
        if echo "$body" | jq -e '.data.success' > /dev/null 2>&1; then
            print_success "密码重置成功"
            return 0
        else
            print_error "响应格式不正确"
            return 1
        fi
    elif [ "$http_code" -eq 401 ]; then
        print_error "验证码错误或已过期 (HTTP $http_code)"
        return 1
    elif [ "$http_code" -eq 400 ]; then
        print_error "请求参数错误 (HTTP $http_code)"
        return 1
    elif [ "$http_code" -eq 503 ]; then
        print_error "服务不可用或恢复流程已过期，请重新发起密码恢复 (HTTP $http_code)"
        return 1
    else
        print_error "密码重置失败 (HTTP $http_code)"
        return 1
    fi
}

# 测试管理端 - 使用验证码重置密码（完整流程）
test_admin_reset_with_code() {
    print_header "测试 4: 管理端 - 使用验证码重置密码（完整流程）"
    
    # 步骤1: 先发起密码恢复（这会在Redis中存储flow_id）
    print_info "步骤 1/3: 发起密码恢复..."
    local initiate_endpoint="${ADMIN_BASE_URL}${ADMIN_API_BASE}/auth/recovery/initiate"
    local initiate_response
    local initiate_http_code
    
    initiate_response=$(curl -s -w "\n%{http_code}" -X POST "$initiate_endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\"}")
    
    initiate_http_code=$(echo "$initiate_response" | tail -n1)
    local initiate_body=$(echo "$initiate_response" | sed '$d')
    
    if [ "$initiate_http_code" -ne 200 ]; then
        print_error "发起密码恢复失败 (HTTP $initiate_http_code): $initiate_body"
        return 1
    fi
    
    print_success "密码恢复请求已发送"
    
    # 步骤2: 获取验证码
    echo ""
    print_info "步骤 2/3: 获取验证码"
    print_warning "请检查邮箱 $TEST_EMAIL 获取验证码"
    read -p "请输入6位验证码: " verification_code
    
    if [ -z "$verification_code" ]; then
        print_error "验证码不能为空"
        return 1
    fi
    
    # 步骤3: 使用验证码重置密码
    print_info "步骤 3/3: 使用验证码重置密码"
    local endpoint="${ADMIN_BASE_URL}${ADMIN_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "邮箱: $TEST_EMAIL"
    print_info "验证码: $verification_code"
    print_info "新密码: $NEW_PASSWORD"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"code\":\"$verification_code\",\"new_password\":\"$NEW_PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 200 ]; then
        if echo "$body" | jq -e '.data.success' > /dev/null 2>&1; then
            print_success "密码重置成功"
            return 0
        else
            print_error "响应格式不正确"
            return 1
        fi
    elif [ "$http_code" -eq 401 ]; then
        print_error "验证码错误或已过期 (HTTP $http_code)"
        return 1
    elif [ "$http_code" -eq 400 ]; then
        print_error "请求参数错误 (HTTP $http_code)"
        return 1
    elif [ "$http_code" -eq 503 ]; then
        print_error "服务不可用或恢复流程已过期，请重新发起密码恢复 (HTTP $http_code)"
        return 1
    else
        print_error "密码重置失败 (HTTP $http_code)"
        return 1
    fi
}

# 测试参数验证 - 无效的邮箱
test_invalid_email() {
    print_header "测试 5: 参数验证 - 无效的邮箱格式"
    
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "使用无效邮箱: invalid-email"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"invalid-email\",\"code\":\"123456\",\"new_password\":\"$NEW_PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 400 ]; then
        print_success "正确拒绝了无效的邮箱格式"
        return 0
    else
        print_error "应该返回400错误，但返回了 HTTP $http_code"
        return 1
    fi
}

# 测试参数验证 - 无效的验证码
test_invalid_code() {
    print_header "测试 6: 参数验证 - 无效的验证码长度"
    
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "使用无效验证码: 123 (长度不是6位)"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"code\":\"123\",\"new_password\":\"$NEW_PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 400 ]; then
        print_success "正确拒绝了无效的验证码长度"
        return 0
    else
        print_error "应该返回400错误，但返回了 HTTP $http_code"
        return 1
    fi
}

# 测试参数验证 - 密码太短
test_short_password() {
    print_header "测试 7: 参数验证 - 密码长度不足"
    
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "使用过短密码: 123 (少于6位)"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"code\":\"123456\",\"new_password\":\"123\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 400 ]; then
        print_success "正确拒绝了过短的密码"
        return 0
    else
        print_error "应该返回400错误，但返回了 HTTP $http_code"
        return 1
    fi
}

# 测试错误的验证码
test_wrong_code() {
    print_header "测试 8: 使用错误的验证码"
    
    local endpoint="${GAME_BASE_URL}${GAME_API_BASE}/auth/password/reset-with-code"
    print_info "请求地址: POST $endpoint"
    print_info "使用错误验证码: 000000"
    
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"code\":\"000000\",\"new_password\":\"$NEW_PASSWORD\"}")
    
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    print_info "HTTP 状态码: $http_code"
    print_info "响应内容: $body"
    
    if [ "$http_code" -eq 401 ] || [ "$http_code" -eq 400 ]; then
        print_success "正确拒绝了错误的验证码"
        return 0
    else
        print_error "应该返回401或400错误，但返回了 HTTP $http_code"
        return 1
    fi
}

# ==================== 主测试流程 ====================

run_all_tests() {
    print_header "密码重置接口测试开始"
    print_info "游戏服务: $GAME_BASE_URL"
    print_info "管理服务: $ADMIN_BASE_URL"
    print_info "测试邮箱: $TEST_EMAIL"
    print_info "新密码: $NEW_PASSWORD"
    
    # 检查 jq 是否安装
    if ! command -v jq &> /dev/null; then
        print_error "需要安装 jq 工具来解析JSON"
        print_info "在 macOS 上: brew install jq"
        exit 1
    fi
    
    # 参数验证测试（不需要真实验证码）
    test_invalid_email || true
    test_invalid_code || true
    test_short_password || true
    test_wrong_code || true
    
    # 交互式测试（需要用户获取验证码）
    echo ""
    print_warning "========== 交互式测试 =========="
    print_info "以下测试需要您从邮箱获取验证码"
    echo ""
    read -p "是否继续进行交互式测试？(y/n): " continue_test
    
    if [ "$continue_test" = "y" ] || [ "$continue_test" = "Y" ]; then
        echo ""
        print_info "选择测试模式："
        print_info "1. 测试游戏端接口（完整流程）"
        print_info "2. 测试管理端接口（完整流程）"
        print_info "3. 测试所有接口"
        read -p "请选择 (1/2/3): " test_mode
        
        case $test_mode in
            1)
                test_game_reset_with_code || true
                ;;
            2)
                test_admin_reset_with_code || true
                ;;
            3)
                test_game_reset_with_code || true
                echo ""
                test_admin_reset_with_code || true
                ;;
            *)
                print_warning "无效的选择，跳过交互式测试"
                ;;
        esac
    else
        print_info "跳过交互式测试"
    fi
    
    # 显示测试结果汇总
    print_header "测试结果汇总"
    echo -e "${CYAN}总测试数:${NC} $TOTAL_TESTS"
    echo -e "${GREEN}通过:${NC} $PASSED_TESTS"
    echo -e "${RED}失败:${NC} $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}所有测试通过! ✓${NC}\n"
        return 0
    else
        echo -e "\n${RED}存在测试失败! ✗${NC}\n"
        return 1
    fi
}

# ==================== 命令行参数解析 ====================

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --game-url)
                GAME_BASE_URL="$2"
                shift 2
                ;;
            --admin-url)
                ADMIN_BASE_URL="$2"
                shift 2
                ;;
            --email)
                TEST_EMAIL="$2"
                shift 2
                ;;
            --new-password)
                NEW_PASSWORD="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# ==================== 主程序入口 ====================

main() {
    parse_args "$@"
    run_all_tests
    exit $?
}

# 执行主程序
main "$@"
