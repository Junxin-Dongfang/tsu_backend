#!/bin/bash

# TSU 监控系统设置脚本
# 用于初始化和配置 Prometheus 与 Grafana

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

# 检查 Docker 是否运行
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi
    log_success "Docker 运行正常"
}

# 创建数据目录
create_data_directories() {
    local data_dir="${1:-./data}"

    log_info "创建数据目录: $data_dir"

    mkdir -p "$data_dir"/{prometheus,grafana,alertmanager}

    # 设置正确的权限
    chmod 755 "$data_dir"
    chmod 755 "$data_dir"/prometheus
    chmod 755 "$data_dir"/grafana
    chmod 755 "$data_dir"/alertmanager

    log_success "数据目录创建完成"
}

# 验证配置文件
validate_configs() {
    local config_dir="${1:-../../configs/monitoring}"

    log_info "验证配置文件..."

    # 验证 Prometheus 配置
    if [[ -f "$config_dir/prometheus/prometheus.yml" ]]; then
        if docker run --rm -v "$config_dir/prometheus:/etc/prometheus" \
            prom/prometheus:v2.48.0 \
            --config.file=/etc/prometheus/prometheus.yml \
            --dry-run; then
            log_success "Prometheus 配置验证通过"
        else
            log_error "Prometheus 配置验证失败"
            return 1
        fi
    else
        log_warning "Prometheus 配置文件不存在: $config_dir/prometheus/prometheus.yml"
    fi

    # 验证 Grafana 配置目录结构
    if [[ -d "$config_dir/grafana/provisioning" ]]; then
        log_success "Grafana 配置目录存在"
    else
        log_warning "Grafana 配置目录不存在: $config_dir/grafana/provisioning"
    fi
}

# 启动监控服务
start_monitoring() {
    local environment="${1:-local}"
    local compose_file="environments/$environment/docker-compose.yml"

    log_info "启动 $environment 环境的监控服务..."

    if [[ ! -f "$compose_file" ]]; then
        log_error "Compose 文件不存在: $compose_file"
        return 1
    fi

    # 启动服务
    docker-compose -f "$compose_file" up -d

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10

    # 检查服务状态
    check_service_health

    log_success "监控服务启动完成"
}

# 检查服务健康状态
check_service_health() {
    log_info "检查服务健康状态..."

    # 检查 Prometheus
    if curl -s http://localhost:9090/-/healthy >/dev/null; then
        log_success "✓ Prometheus 运行正常 (http://localhost:9090)"
    else
        log_warning "✗ Prometheus 未就绪"
    fi

    # 检查 Grafana
    if curl -s http://localhost:3000/api/health >/dev/null; then
        log_success "✓ Grafana 运行正常 (http://localhost:3000)"
        log_info "  默认登录: admin/admin"
    else
        log_warning "✗ Grafana 未就绪"
    fi
}

# 显示访问信息
show_access_info() {
    echo
    log_info "监控系统访问信息:"
    echo "  Prometheus: http://localhost:9090"
    echo "  Grafana:    http://localhost:3000 (admin/admin)"
    echo
    log_info "常用命令:"
    echo "  查看日志: docker-compose -f environments/local/docker-compose.yml logs -f"
    echo "  停止服务: docker-compose -f environments/local/docker-compose.yml down"
    echo "  重启服务: docker-compose -f environments/local/docker-compose.yml restart"
    echo
}

# 主函数
main() {
    local environment="${1:-local}"
    local action="${2:-start}"

    log_info "TSU 监控系统设置脚本"
    log_info "环境: $environment, 操作: $action"

    check_docker

    case "$action" in
        "start")
            create_data_directories
            validate_configs
            start_monitoring "$environment"
            show_access_info
            ;;
        "validate")
            validate_configs
            ;;
        "health")
            check_service_health
            ;;
        *)
            echo "用法: $0 [environment] [action]"
            echo "  environment: local (默认) | production"
            echo "  action: start (默认) | validate | health"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"