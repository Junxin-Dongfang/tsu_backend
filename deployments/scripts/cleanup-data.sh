#!/bin/bash

# TSU 监控数据清理脚本
# 用于清理 Prometheus 和 Grafana 的运行时数据

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 确认函数
confirm_action() {
    local message="$1"
    while true; do
        read -p "$message [y/N]: " -n 1 -r
        echo
        case $REPLY in
            [Yy]) return 0 ;;
            [Nn]|'') return 1 ;;
            *) echo "请输入 y 或 n" ;;
        esac
    done
}

# 停止监控服务
stop_services() {
    local environment="${1:-local}"

    log_info "停止 $environment 环境的监控服务..."

    if docker-compose -f "environments/$environment/docker-compose.yml" ps -q | grep -q .; then
        docker-compose -f "environments/$environment/docker-compose.yml" down
        log_success "监控服务已停止"
    else
        log_info "没有运行中的监控服务"
    fi
}

# 清理数据目录
cleanup_data() {
    local data_dir="${1:-./data}"

    log_info "清理数据目录: $data_dir"

    if [[ ! -d "$data_dir" ]]; then
        log_warning "数据目录不存在: $data_dir"
        return 0
    fi

    # 显示数据目录大小
    if command -v du >/dev/null 2>&1; then
        local size=$(du -sh "$data_dir" 2>/dev/null | cut -f1)
        log_info "当前数据目录大小: $size"
    fi

    # 确认清理操作
    if confirm_action "确定要清理所有监控数据吗？此操作不可逆！"; then
        # 删除数据目录内容
        rm -rf "$data_dir"/prometheus/*
        rm -rf "$data_dir"/grafana/*
        rm -rf "$data_dir"/alertmanager/*

        log_success "数据清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 清理 Docker 资源
cleanup_docker() {
    log_info "清理相关的 Docker 资源..."

    # 清理停止的容器
    local stopped_containers=$(docker ps -a --filter "name=tsu_" --format "{{.Names}}" | tr '\n' ' ')
    if [[ -n "$stopped_containers" ]]; then
        log_info "删除停止的 TSU 容器: $stopped_containers"
        docker rm $stopped_containers
    fi

    # 清理未使用的 volumes
    local unused_volumes=$(docker volume ls -q --filter "name=prometheus\|grafana\|alertmanager" 2>/dev/null || true)
    if [[ -n "$unused_volumes" ]]; then
        log_info "删除未使用的监控 volumes"
        docker volume rm $unused_volumes 2>/dev/null || true
    fi

    log_success "Docker 资源清理完成"
}

# 重置监控系统
reset_monitoring() {
    local environment="${1:-local}"

    log_warning "即将重置整个监控系统（包括停止服务、清理数据、清理 Docker 资源）"

    if confirm_action "确定要重置监控系统吗？这将删除所有监控数据！"; then
        stop_services "$environment"
        cleanup_data
        cleanup_docker
        log_success "监控系统重置完成"

        log_info "如需重新启动，请运行: ./setup-monitoring.sh $environment start"
    else
        log_info "取消重置操作"
    fi
}

# 显示使用帮助
show_help() {
    echo "TSU 监控数据清理脚本"
    echo
    echo "用法: $0 [environment] [action]"
    echo
    echo "环境:"
    echo "  local       本地开发环境 (默认)"
    echo "  production  生产环境"
    echo
    echo "操作:"
    echo "  stop        停止监控服务"
    echo "  data        仅清理数据目录"
    echo "  docker      仅清理 Docker 资源"
    echo "  reset       完全重置 (停止服务 + 清理数据 + 清理资源)"
    echo "  help        显示此帮助信息"
    echo
    echo "示例:"
    echo "  $0 local data        # 清理本地环境数据"
    echo "  $0 production reset  # 重置生产环境监控"
    echo
}

# 主函数
main() {
    local environment="${1:-local}"
    local action="${2:-help}"

    log_info "TSU 监控数据清理脚本"
    log_info "环境: $environment, 操作: $action"

    # 切换到脚本所在目录
    cd "$(dirname "$0")"

    case "$action" in
        "stop")
            stop_services "$environment"
            ;;
        "data")
            cleanup_data
            ;;
        "docker")
            cleanup_docker
            ;;
        "reset")
            reset_monitoring "$environment"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "未知操作: $action"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"