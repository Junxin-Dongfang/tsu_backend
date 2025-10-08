#!/bin/bash

# ==========================================
# TSU 项目生产环境部署 - 步骤 4: Nginx
# ==========================================
# 部署内容：
#   - Nginx 反向代理

set -e

# 加载通用函数库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/deploy-common.sh"

print_step "步骤 4: 部署 Nginx 反向代理"

# ==========================================
# 1. 检查依赖服务
# ==========================================
print_step "[1/5] 检查依赖服务"

print_info "检查主服务..."
if ! check_container_running "tsu_admin"; then
    print_error "Admin Server 未运行，请先执行步骤 3"
    exit 1
fi

print_info "检查 Oathkeeper..."
if ! check_container_running "tsu_oathkeeper"; then
    print_error "Oathkeeper 未运行，请先执行步骤 2"
    exit 1
fi

print_success "依赖服务检查通过"

# ==========================================
# 2. 上传配置文件
# ==========================================
print_step "[2/5] 上传 Nginx 配置文件"

print_info "上传 docker-compose 配置..."
ssh_copy "$PROJECT_DIR/deployments/docker-compose/docker-compose.prod.4-nginx.yml" "$SERVER_DEPLOY_DIR/"

print_info "上传 Nginx 配置..."
ssh_copy "$PROJECT_DIR/infra/nginx/prod.conf" "$SERVER_DEPLOY_DIR/infra/nginx/"

# 创建 web 目录（如果需要）
print_info "创建 web 目录..."
ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/web/admin"
ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/web/user"

print_success "配置文件上传完成"

# ==========================================
# 3. 启动 Nginx
# ==========================================
print_step "[3/5] 启动 Nginx 服务"

print_info "启动 Nginx..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.4-nginx.yml up -d"

print_info "等待服务启动..."
sleep 10

# ==========================================
# 4. 验证服务状态
# ==========================================
print_step "[4/5] 验证服务状态"

print_info "检查容器状态..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.4-nginx.yml ps"

echo ""
print_info "测试 Nginx 配置..."
if ssh_exec "docker exec tsu_nginx nginx -t"; then
    print_success "Nginx 配置正确"
else
    print_error "Nginx 配置有误，请检查"
    exit 1
fi

print_info "测试反向代理..."
if ssh_exec "curl -sf http://localhost/health > /dev/null"; then
    print_success "Nginx 反向代理正常"
else
    print_warning "Nginx 反向代理可能未就绪"
fi

# ==========================================
# 5. 测试外部访问
# ==========================================
print_step "[5/5] 测试外部访问"

print_info "从本地测试访问..."
if curl -sf http://47.239.139.109/health > /dev/null 2>&1; then
    print_success "外部访问正常"
else
    print_warning "外部访问失败，可能原因："
    print_info "  - 服务器防火墙未开放 80 端口"
    print_info "  - 安全组规则未配置"
    print_info "  - 服务尚未完全启动"
fi

# ==========================================
# 部署完成
# ==========================================
print_step "✅ 步骤 4 完成：Nginx 部署成功！"

echo ""
echo -e "${BLUE}=========================================="
echo -e "  所有服务部署完成！"
echo -e "==========================================${NC}"
echo ""
echo -e "${GREEN}部署层次概览：${NC}"
echo "  第一层 - 基础设施："
echo "    ✓ PostgreSQL 主数据库 (5432)"
echo "    ✓ PostgreSQL Ory数据库 (5433)"
echo "    ✓ Redis (6379)"
echo "    ✓ NATS (4222)"
echo "    ✓ Consul (8500)"
echo ""
echo "  第二层 - Ory 服务："
echo "    ✓ Kratos (4433/4434)"
echo "    ✓ Keto (4466/4467)"
echo "    ✓ Oathkeeper (4456/4457)"
echo ""
echo "  第三层 - 主服务："
echo "    ✓ Admin Server (8071)"
echo ""
echo "  第四层 - 接入层："
echo "    ✓ Nginx (80/443)"
echo ""
echo -e "${BLUE}=========================================="
echo -e "  访问信息"
echo -e "==========================================${NC}"
echo ""
echo -e "${GREEN}服务访问地址：${NC}"
echo "  - 主入口: http://47.239.139.109/"
echo "  - Admin API: http://47.239.139.109/api/"
echo "  - Swagger UI: http://47.239.139.109/swagger/"
echo "  - 健康检查: http://47.239.139.109/health"
echo ""
echo -e "${YELLOW}管理员账号：${NC}"
echo "  - 用户名: root"
echo "  - 密码: password"
echo ""
echo -e "${RED}⚠️  安全提示：${NC}"
echo "  1. 请立即登录并修改 root 密码"
echo "  2. 配置服务器防火墙，限制不必要的端口访问"
echo "  3. 考虑配置 HTTPS 证书"
echo "  4. 定期备份数据库"
echo ""
echo -e "${BLUE}=========================================="
echo -e "  常用运维命令"
echo -e "==========================================${NC}"
echo ""
echo "查看所有服务状态："
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker ps'"
echo ""
echo "查看服务日志："
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker logs -f tsu_admin'"
echo ""
echo "重启服务："
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.3-app.yml restart tsu_admin'"
echo ""
echo "停止所有服务："
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.4-nginx.yml down'"
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.3-app.yml down'"
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.2-ory.yml down'"
echo "  ssh root@47.239.139.109 'cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.1-infra.yml down'"
echo ""

print_success "🎉 完整部署完成！系统已就绪！"
