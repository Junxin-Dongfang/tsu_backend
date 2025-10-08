#!/bin/bash

################################################################################
# TSU Admin API 全面测试 - 主入口
################################################################################

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 默认配置
BASE_URL="${BASE_URL:-http://localhost:80}"
USERNAME="${USERNAME:-root}"
PASSWORD="${PASSWORD:-password}"
VERBOSE="${VERBOSE:-false}"
NO_CLEANUP="${NO_CLEANUP:-false}"
CONTINUE_ON_FAILURE="${CONTINUE_ON_FAILURE:-true}"
SPECIFIC_SUITE=""

# 全局变量
AUTH_TOKEN=""
CURRENT_USER_ID=""

# 创建报告目录
REPORT_DIR="$SCRIPT_DIR/reports/run_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$REPORT_DIR"

# 日志文件
LOG_DETAILED="$REPORT_DIR/detailed.log"
LOG_API_CALLS="$REPORT_DIR/api_calls.log"
LOG_FAILURES="$REPORT_DIR/failures.log"
LOG_SUMMARY="$REPORT_DIR/summary.log"

# 创建日志文件
touch "$LOG_DETAILED" "$LOG_API_CALLS" "$LOG_FAILURES" "$LOG_SUMMARY"

################################################################################
# 帮助信息
################################################################################

show_help() {
    cat << EOF
TSU Admin API 全面测试框架

用法: $0 [选项]

选项:
  --url URL              API 基础地址 (默认: http://localhost:80)
  --username USER        测试账号用户名 (默认: root)
  --password PASS        测试账号密码 (默认: password)
  --suite SUITE          只运行指定的测试套件 (例如: 01, 02, skill)
  --verbose              详细输出模式
  --no-cleanup           测试后不清理测试数据
  --continue-on-failure  失败后继续执行 (默认: true)
  --help                 显示此帮助信息

测试套件列表:
  01  - 系统健康检查
  02  - 认证流程
  03  - 用户管理
  04  - RBAC 权限系统
  05  - 基础游戏配置
  06  - 元数据定义
  07  - 技能系统
  08  - 效果系统
  09  - 动作系统
  10  - 关联关系
  11  - 边界条件

示例:
  $0                                    # 运行所有测试
  $0 --suite 07                         # 只运行技能系统测试
  $0 --verbose --no-cleanup             # 详细模式且不清理数据
  $0 --url http://test.example.com      # 使用自定义 API 地址

报告输出:
  报告目录: $REPORT_DIR
  - detailed.log    : 详细测试日志
  - api_calls.log   : 所有 API 调用记录
  - failures.log    : 失败用例详情
  - summary.log     : 测试摘要
  - test_data.json  : 测试数据快照

EOF
}

################################################################################
# 参数解析
################################################################################

parse_arguments() {
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
            --suite)
                SPECIFIC_SUITE="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --no-cleanup)
                NO_CLEANUP=true
                shift
                ;;
            --continue-on-failure)
                CONTINUE_ON_FAILURE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                echo "使用 --help 查看帮助信息"
                exit 1
                ;;
        esac
    done
}

################################################################################
# 环境检查
################################################################################

check_dependencies() {
    local missing_deps=()
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo "错误: 缺少必需的依赖工具: ${missing_deps[*]}"
        echo ""
        echo "安装方法:"
        echo "  macOS:  brew install ${missing_deps[*]}"
        echo "  Ubuntu: apt-get install ${missing_deps[*]}"
        echo "  CentOS: yum install ${missing_deps[*]}"
        exit 1
    fi
}

################################################################################
# 加载测试框架和工具
################################################################################

load_framework() {
    # 加载测试框架
    source "$SCRIPT_DIR/lib/test_framework.sh"
    source "$SCRIPT_DIR/lib/test_data.sh"
    source "$SCRIPT_DIR/lib/test_utils.sh"
    
    echo "测试框架加载完成"
}

################################################################################
# 加载测试套件
################################################################################

load_test_suite() {
    local suite_file="$1"
    
    if [ ! -f "$suite_file" ]; then
        echo "警告: 测试套件文件不存在: $suite_file"
        return 1
    fi
    
    source "$suite_file"
    return 0
}

################################################################################
# 运行测试套件
################################################################################

run_test_suites() {
    # 定义所有测试套件
    local suites=(
        "01_system_health:test_system_health"
        "02_authentication:test_authentication"
        "03_user_management:test_user_management"
        "04_rbac_system:test_rbac_system"
        "05_game_config_basic:test_game_config_basic"
        "06_metadata:test_metadata"
        "07_skill_system:test_skill_system"
        "08_effect_system:test_effect_system"
        "09_action_system:test_action_system"
        "10_relations:test_relations"
        "11_edge_cases:test_edge_cases"
    )
    
    for suite_info in "${suites[@]}"; do
        local suite_file="${suite_info%%:*}"
        local suite_func="${suite_info##*:}"
        
        # 如果指定了特定套件，跳过其他套件
        if [ -n "$SPECIFIC_SUITE" ]; then
            if [[ ! "$suite_file" =~ $SPECIFIC_SUITE ]]; then
                continue
            fi
        fi
        
        # 加载并运行测试套件
        if load_test_suite "$SCRIPT_DIR/suites/${suite_file}.sh"; then
            # 执行测试套件函数
            $suite_func
            
            # 如果失败且不继续执行，则退出
            if [ $SUITE_TESTS_FAILED -gt 0 ] && [ "$CONTINUE_ON_FAILURE" != "true" ]; then
                echo ""
                echo "测试套件失败，停止执行"
                return 1
            fi
        fi
    done
    
    return 0
}

################################################################################
# 生成测试报告
################################################################################

generate_summary_report() {
    {
        echo "╔════════════════════════════════════════╗"
        echo "║     TSU Admin API 测试摘要报告     ║"
        echo "╚════════════════════════════════════════╝"
        echo ""
        echo "测试时间: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "API 地址: $BASE_URL"
        echo "测试账号: $USERNAME"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  测试结果"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "  总测试数:   $GLOBAL_TESTS_TOTAL"
        echo "  通过:       $GLOBAL_TESTS_PASSED"
        echo "  失败:       $GLOBAL_TESTS_FAILED"
        
        if [ $GLOBAL_TESTS_TOTAL -gt 0 ]; then
            local pass_rate=$((GLOBAL_TESTS_PASSED * 100 / GLOBAL_TESTS_TOTAL))
            echo "  通过率:     ${pass_rate}%"
        fi
        
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "报告文件:"
        echo "  详细日志: $LOG_DETAILED"
        echo "  API 调用: $LOG_API_CALLS"
        echo "  失败记录: $LOG_FAILURES"
        echo "  测试数据: $REPORT_DIR/test_data.json"
        echo ""
    } > "$LOG_SUMMARY"
    
    cat "$LOG_SUMMARY"
}

################################################################################
# 主程序
################################################################################

main() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║     TSU Admin API 全面测试框架     ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    # 解析参数
    parse_arguments "$@"
    
    # 检查依赖
    check_dependencies
    
    # 加载框架
    load_framework
    
    # 设置退出陷阱（在加载框架后）
    trap cleanup_all_test_data EXIT
    
    # 开始测试
    start_global_test
    
    # 运行测试套件
    if run_test_suites; then
        # 保存测试数据快照
        save_test_data_snapshot "$REPORT_DIR/test_data.json"
        
        # 清理测试数据
        cleanup_all_test_data
        
        # 结束测试并生成报告
        end_global_test
        
        # 生成摘要报告
        generate_summary_report
        
        # 返回结果
        if [ $GLOBAL_TESTS_FAILED -eq 0 ]; then
            echo ""
            echo "✅ 所有测试通过！"
            exit 0
        else
            echo ""
            echo "❌ 存在失败的测试，请查看报告"
            exit 1
        fi
    else
        echo ""
        echo "❌ 测试执行失败"
        
        # 尝试清理
        cleanup_all_test_data
        
        exit 1
    fi
}

# 运行主程序
main "$@"
