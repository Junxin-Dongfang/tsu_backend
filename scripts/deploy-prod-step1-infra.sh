#!/bin/bash

# ==========================================
# TSU 项目生产环境部署 - 步骤 1: 基础设施
# ==========================================
# 部署内容：
#   - PostgreSQL 主数据库
#   - PostgreSQL Ory 数据库
#   - Redis
#   - NATS
#   - Consul

set -e

# 加载通用函数库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/deploy-common.sh"

print_step "步骤 1: 部署基础设施服务"

# ==========================================
# 1. 环境检查
# ==========================================
print_step "[1/8] 检查本地环境"

check_sshpass || exit 1
check_docker || exit 1

print_success "本地环境检查通过"

# ==========================================
# 2. 生成或检查环境变量文件
# ==========================================
print_step "[2/8] 准备环境配置"

if [ ! -f "$PROJECT_DIR/.env.prod" ]; then
    print_info ".env.prod 不存在，开始生成..."
    create_env_file
else
    print_success ".env.prod 文件已存在"
fi

# ==========================================
# 3. 测试服务器连接
# ==========================================
print_step "[3/8] 测试服务器连接"

test_ssh_connection || exit 1

# ==========================================
# 4. 初始化服务器目录
# ==========================================
print_step "[4/8] 初始化服务器目录"

init_server_directories

# ==========================================
# 5. 上传配置文件
# ==========================================
print_step "[5/8] 上传配置文件到服务器"

print_info "上传 docker-compose 配置..."
ssh_copy "$PROJECT_DIR/deployments/docker-compose/docker-compose.prod.1-infra.yml" "$SERVER_DEPLOY_DIR/"

print_info "上传环境变量文件..."
ssh_copy "$PROJECT_DIR/.env.prod" "$SERVER_DEPLOY_DIR/"

print_info "上传 Ory 初始化脚本..."
ssh_copy "$PROJECT_DIR/infra/ory/init-schemas.sql" "$SERVER_DEPLOY_DIR/infra/ory/"

print_success "配置文件上传完成"

# ==========================================
# 6. 创建 Docker 网络
# ==========================================
print_step "[6/8] 创建 Docker 网络"

print_info "创建 tsu_network 网络..."
if ssh_exec "docker network inspect tsu_network >/dev/null 2>&1"; then
    print_warning "网络 tsu_network 已存在"
else
    ssh_exec "docker network create tsu_network"
    print_success "网络 tsu_network 创建成功"
fi

# ==========================================
# 7. 启动基础设施服务
# ==========================================
print_step "[7/8] 启动基础设施服务"

print_info "启动服务（这可能需要几分钟）..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.1-infra.yml --env-file .env.prod up -d"

print_info "等待服务启动..."
sleep 10

# ==========================================
# 8. 验证服务状态
# ==========================================
print_step "[8/8] 验证服务状态"

print_info "检查容器状态..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.1-infra.yml ps"

echo ""
print_info "等待数据库健康检查..."
wait_for_container_healthy "tsu_postgres_main" 60
wait_for_container_healthy "tsu_postgres_ory" 60

print_info "等待其他服务就绪..."
wait_for_container_healthy "tsu_redis" 30
wait_for_container_healthy "tsu_nats" 30
wait_for_container_healthy "tsu_consul" 30

# ==========================================
# 部署完成
# ==========================================
print_step "✅ 步骤 1 完成：基础设施部署成功！"

echo ""
echo -e "${BLUE}已部署的服务：${NC}"
echo "  - PostgreSQL 主数据库: tsu_postgres_main (端口 5432)"
echo "  - PostgreSQL Ory数据库: tsu_postgres_ory (端口 5433)"
echo "  - Redis: tsu_redis (端口 6379)"
echo "  - NATS: tsu_nats (端口 4222)"
echo "  - Consul: tsu_consul (端口 8500)"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  运行: make deploy-prod-step2"
echo "  或: ./scripts/deploy-prod-step2-ory.sh"
echo ""

print_success "🎉 基础设施部署完成！"
