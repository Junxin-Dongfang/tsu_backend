#!/bin/bash
################################################################################
# 测试快速执行脚本
# 用途: 一键执行 Admin API 测试
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_header() {
    echo -e "\n${BLUE}${BOLD}========================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}========================================${NC}\n"
}

print_info() {
    echo -e "${BLUE}[ℹ]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# 显示菜单
show_menu() {
    print_header "Admin API 测试工具"
    
    echo "请选择测试方式:"
    echo ""
    echo "  ${BOLD}1)${NC} Python 自动化测试 ${GREEN}(推荐)${NC}"
    echo "     - 功能最完整"
    echo "     - 生成详细 JSON 报告"
    echo "     - 彩色输出"
    echo ""
    echo "  ${BOLD}2)${NC} Bash 脚本测试"
    echo "     - 无需 Python 环境"
    echo "     - 适合 CI/CD"
    echo "     - 生成文本日志"
    echo ""
    echo "  ${BOLD}3)${NC} 快速健康检查"
    echo "     - 只测试关键接口"
    echo "     - 快速验证服务状态"
    echo ""
    echo "  ${BOLD}4)${NC} 打开 Swagger UI"
    echo "     - 可视化界面"
    echo "     - 手动测试接口"
    echo ""
    echo "  ${BOLD}5)${NC} 查看测试计划文档"
    echo ""
    echo "  ${BOLD}6)${NC} 退出"
    echo ""
}

# 检查服务状态
check_services() {
    print_info "检查服务状态..."
    
    # 检查健康接口
    if curl -s -f http://localhost:80/health > /dev/null 2>&1; then
        print_success "服务运行正常"
        return 0
    else
        print_error "服务未运行或无法访问"
        print_warning "请先启动服务: make dev-up 或 docker-compose up -d"
        return 1
    fi
}

# Python 测试
run_python_test() {
    print_header "Python 自动化测试"
    
    # 检查 Python
    if ! command -v python3 &> /dev/null; then
        print_error "未找到 Python3"
        return 1
    fi
    
    # 检查依赖
    print_info "检查依赖..."
    if ! python3 -c "import requests" 2>/dev/null; then
        print_warning "未安装 requests 库，正在安装..."
        pip3 install requests
    fi
    
    # 运行测试
    print_info "开始测试..."
    echo ""
    python3 admin-api-test.py "$@"
    
    local exit_code=$?
    echo ""
    
    if [ $exit_code -eq 0 ]; then
        print_success "测试完成"
    else
        print_error "测试失败 (退出码: $exit_code)"
    fi
    
    return $exit_code
}

# Bash 测试
run_bash_test() {
    print_header "Bash 脚本测试"
    
    # 检查依赖
    print_info "检查依赖..."
    
    if ! command -v curl &> /dev/null; then
        print_error "未找到 curl"
        return 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "未找到 jq (JSON 处理工具)"
        print_info "安装命令:"
        echo "  macOS:  brew install jq"
        echo "  Ubuntu: sudo apt-get install jq"
        return 1
    fi
    
    # 运行测试
    print_info "开始测试..."
    echo ""
    ./admin-api-test.sh "$@"
    
    local exit_code=$?
    echo ""
    
    if [ $exit_code -eq 0 ]; then
        print_success "测试完成"
    else
        print_error "测试失败 (退出码: $exit_code)"
    fi
    
    return $exit_code
}

# 快速健康检查
quick_health_check() {
    print_header "快速健康检查"
    
    local failed=0
    
    # 1. 健康检查
    print_info "测试: 健康检查接口..."
    if curl -s -f http://localhost:80/health > /dev/null; then
        print_success "健康检查: OK"
    else
        print_error "健康检查: FAILED"
        ((failed++))
    fi
    
    # 2. Swagger
    print_info "测试: Swagger 文档..."
    if curl -s -f http://localhost:80/swagger/index.html > /dev/null; then
        print_success "Swagger 文档: OK"
    else
        print_error "Swagger 文档: FAILED"
        ((failed++))
    fi
    
    # 3. 登录
    print_info "测试: 用户登录..."
    local response=$(curl -s -X POST http://localhost:80/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"identifier":"root","password":"password"}')
    
    if echo "$response" | jq -e '.data.session_token // .data.token // .session_token // .token' > /dev/null 2>&1; then
        print_success "用户登录: OK"
        local token=$(echo "$response" | jq -r '.data.session_token // .data.token // .session_token // .token')
        
        # 4. 获取用户信息
        print_info "测试: 获取当前用户..."
        if curl -s -f http://localhost:80/api/v1/admin/users/me \
            -H "Authorization: Bearer $token" > /dev/null; then
            print_success "用户信息: OK"
        else
            print_error "用户信息: FAILED"
            ((failed++))
        fi
        
        # 5. 测试基础查询
        print_info "测试: 基础数据查询..."
        if curl -s -f "http://localhost:80/api/v1/admin/users?page=1&page_size=1" \
            -H "Authorization: Bearer $token" > /dev/null; then
            print_success "数据查询: OK"
        else
            print_error "数据查询: FAILED"
            ((failed++))
        fi
    else
        print_error "用户登录: FAILED"
        ((failed++))
    fi
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    if [ $failed -eq 0 ]; then
        print_success "所有检查通过 ✓"
    else
        print_error "发现 $failed 个问题 ✗"
    fi
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    return $failed
}

# 打开 Swagger UI
open_swagger() {
    print_header "打开 Swagger UI"
    
    local url="http://localhost:80/swagger/index.html"
    print_info "Swagger URL: $url"
    
    # 检查服务
    if ! curl -s -f "$url" > /dev/null; then
        print_error "无法访问 Swagger UI"
        print_warning "请确认服务已启动"
        return 1
    fi
    
    # 在浏览器中打开
    if command -v open &> /dev/null; then
        # macOS
        open "$url"
        print_success "已在浏览器中打开"
    elif command -v xdg-open &> /dev/null; then
        # Linux
        xdg-open "$url"
        print_success "已在浏览器中打开"
    else
        print_warning "无法自动打开浏览器"
        print_info "请手动访问: $url"
    fi
    
    echo ""
    print_info "Swagger 使用提示:"
    echo "  1. 先调用 POST /api/v1/auth/login 获取 token"
    echo "  2. 点击右上角 'Authorize' 按钮"
    echo "  3. 输入: Bearer {your_token}"
    echo "  4. 点击 'Authorize'"
    echo "  5. 现在可以测试需要认证的接口了"
}

# 查看文档
view_docs() {
    print_header "测试文档"
    
    if [ -f "README_TEST.md" ]; then
        # 尝试使用 bat 或 cat 显示
        if command -v bat &> /dev/null; then
            bat README_TEST.md
        elif command -v less &> /dev/null; then
            less README_TEST.md
        else
            cat README_TEST.md
        fi
    else
        print_error "未找到文档文件"
    fi
    
    echo ""
    print_info "文档位置:"
    echo "  - 测试指南: test/README_TEST.md"
    echo "  - 测试计划: test/api-test-plan.md"
    echo "  - 认证指南: docs/AUTHENTICATION_GUIDE.md"
}

# 主函数
main() {
    cd "$(dirname "$0")"
    
    # 如果有命令行参数，直接执行
    if [ "$1" == "--python" ]; then
        shift
        check_services && run_python_test "$@"
        exit $?
    elif [ "$1" == "--bash" ]; then
        shift
        check_services && run_bash_test "$@"
        exit $?
    elif [ "$1" == "--quick" ]; then
        check_services && quick_health_check
        exit $?
    elif [ "$1" == "--swagger" ]; then
        open_swagger
        exit $?
    elif [ "$1" == "--help" ]; then
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --python   使用 Python 自动化测试"
        echo "  --bash     使用 Bash 脚本测试"
        echo "  --quick    快速健康检查"
        echo "  --swagger  打开 Swagger UI"
        echo "  --help     显示帮助信息"
        echo ""
        echo "无参数运行时显示交互菜单"
        exit 0
    fi
    
    # 交互式菜单
    while true; do
        show_menu
        read -p "请选择 [1-6]: " choice
        
        case $choice in
            1)
                check_services && run_python_test
                ;;
            2)
                check_services && run_bash_test
                ;;
            3)
                check_services && quick_health_check
                ;;
            4)
                open_swagger
                ;;
            5)
                view_docs
                ;;
            6)
                print_info "再见！"
                exit 0
                ;;
            *)
                print_error "无效选择，请输入 1-6"
                ;;
        esac
        
        echo ""
        read -p "按 Enter 继续..."
    done
}

main "$@"
