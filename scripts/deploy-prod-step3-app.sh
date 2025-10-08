#!/bin/bash

# ==========================================
# TSU 项目生产环境部署 - 步骤 3: 主服务
# ==========================================
# 部署内容：
#   1. 构建 Docker 镜像
#   2. 推送镜像到 Docker Hub
#   3. 部署 Admin Server
#   4. 执行数据库迁移
#   5. 初始化 root 用户

set -e

# 加载通用函数库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/deploy-common.sh"

print_step "步骤 3: 部署主应用服务"

# ==========================================
# 1. 检查依赖服务
# ==========================================
print_step "[1/11] 检查依赖服务"

print_info "检查基础设施服务..."
if ! check_container_running "tsu_postgres_main"; then
    print_error "主数据库未运行，请先执行步骤 1"
    exit 1
fi

if ! check_container_running "tsu_redis"; then
    print_error "Redis 未运行，请先执行步骤 1"
    exit 1
fi

print_info "检查 Ory 服务..."
if ! check_container_running "tsu_kratos"; then
    print_error "Kratos 未运行，请先执行步骤 2"
    exit 1
fi

print_success "依赖服务检查通过"

# ==========================================
# 2. 检查 Docker Hub 配置
# ==========================================
print_step "[2/11] 检查 Docker Hub 配置"

if [ ! -f "$PROJECT_DIR/.registry.conf" ]; then
    print_error "未找到 .registry.conf 文件"
    print_info "请执行以下步骤："
    print_info "  1. cp .registry.conf.example .registry.conf"
    print_info "  2. vim .registry.conf  # 填写 Docker Hub 用户名和密码"
    exit 1
fi

source "$PROJECT_DIR/.registry.conf"

if [ -z "$DOCKERHUB_USERNAME" ] || [ "$DOCKERHUB_USERNAME" = "your-dockerhub-username" ]; then
    print_error "请在 .registry.conf 中配置 Docker Hub 用户名"
    exit 1
fi

print_success "Docker Hub 配置检查通过"

# 构建镜像标签
IMAGE_TAG="${DOCKERHUB_USERNAME}/${IMAGE_NAME}:${IMAGE_VERSION}"

# ==========================================
# 3. 构建 Docker 镜像
# ==========================================
print_step "[3/11] 构建 Docker 镜像"

print_info "开始构建镜像: $IMAGE_TAG"
print_info "这可能需要几分钟时间..."

cd "$PROJECT_DIR"
docker build \
    --platform linux/amd64 \
    -f deployments/docker/Dockerfile.admin.prod \
    -t "$IMAGE_TAG" \
    --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    --build-arg VERSION="${IMAGE_VERSION}" \
    .

print_success "镜像构建完成"

# ==========================================
# 4. 登录 Docker Hub
# ==========================================
print_step "[4/11] 保存镜像到文件"

print_info "保存镜像为 tar.gz..."
TEMP_IMAGE_FILE="/tmp/tsu-admin-server.tar.gz"
docker save "$IMAGE_TAG" | gzip > "$TEMP_IMAGE_FILE"

print_success "镜像已保存"

# ==========================================
# 5. 上传镜像到服务器
# ==========================================
print_step "[5/11] 上传镜像到服务器"

print_info "上传镜像（约 30MB，需要几分钟）..."
sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no "$TEMP_IMAGE_FILE" "$SERVER_USER@$SERVER_HOST:/tmp/"

print_info "在服务器加载镜像..."
ssh_exec "docker load < /tmp/tsu-admin-server.tar.gz && rm /tmp/tsu-admin-server.tar.gz && docker images | grep tsu-admin-server"

print_success "镜像已加载到服务器"
rm -f "$TEMP_IMAGE_FILE"

# ==========================================
# 6. 上传配置文件
# ==========================================
print_step "[6/11] 上传配置文件到服务器"

print_info "上传 docker-compose 配置..."
ssh_copy "$PROJECT_DIR/deployments/docker-compose/docker-compose.prod.3-app.yml" "$SERVER_DEPLOY_DIR/"

print_info "上传 migrations 目录..."
ssh_copy "$PROJECT_DIR/migrations" "$SERVER_DEPLOY_DIR/"

print_info "上传 configs 目录..."
ssh_copy "$PROJECT_DIR/configs" "$SERVER_DEPLOY_DIR/"

print_info "上传 root 用户初始化脚本..."
ssh_copy "$PROJECT_DIR/scripts/init-root-user.sh" "$SERVER_DEPLOY_DIR/"

print_success "配置文件上传完成"

# ==========================================
# 7. 跳过（镜像已在第5步加载）
# ==========================================
print_step "[7/11] 镜像已就绪"

print_info "镜像已在第5步加载，跳过拉取"
print_success "✓ 镜像就绪"

# ==========================================
# 8. 启动主服务
# ==========================================
print_step "[8/11] 启动主服务"

print_info "设置镜像环境变量..."
ssh_exec "cd $SERVER_DEPLOY_DIR && echo 'DOCKER_IMAGE_FULL=$IMAGE_TAG' >> .env.prod"

print_info "启动 Admin Server..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.3-app.yml --env-file .env.prod up -d"

print_info "等待服务启动..."
sleep 15

# ==========================================
# 9. 安装数据库扩展
# ==========================================
print_step "[9/12] 安装数据库扩展"

print_info "安装 pg_uuidv7 扩展..."

# 检查扩展是否已安装
if ssh_exec "docker exec tsu_postgres_main psql -U tsu_user -d tsu_db -tAc \"SELECT 1 FROM pg_extension WHERE extname='pg_uuidv7'\" | grep -q 1"; then
    print_success "pg_uuidv7 扩展已安装"
else
    print_info "在容器内安装构建工具..."
    ssh_exec "docker exec tsu_postgres_main apk add --no-cache git make gcc musl-dev postgresql16-dev" || true
    
    print_info "克隆 pg_uuidv7 源码..."
    ssh_exec "docker exec tsu_postgres_main sh -c 'cd /tmp && rm -rf pg_uuidv7 && git clone --depth=1 https://github.com/fboulnois/pg_uuidv7.git'" || true
    
    print_info "编译扩展..."
    ssh_exec "docker exec tsu_postgres_main sh -c 'cd /tmp/pg_uuidv7 && make'" || true
    
    print_info "复制扩展文件到PostgreSQL目录..."
    ssh_exec "docker exec tsu_postgres_main sh -c 'cp /tmp/pg_uuidv7/pg_uuidv7.so /usr/local/lib/postgresql/ && cp /tmp/pg_uuidv7/pg_uuidv7.control /usr/local/share/postgresql/extension/ && cp /tmp/pg_uuidv7/sql/*.sql /usr/local/share/postgresql/extension/'"
    
    print_info "在数据库中创建扩展..."
    if ssh_exec "docker exec tsu_postgres_main psql -U tsu_user -d tsu_db -c 'CREATE EXTENSION IF NOT EXISTS pg_uuidv7;'"; then
        print_success "pg_uuidv7 扩展安装完成"
    else
        print_error "扩展创建失败"
        exit 1
    fi
    
    print_info "清理临时文件..."
    ssh_exec "docker exec tsu_postgres_main sh -c 'rm -rf /tmp/pg_uuidv7'" || true
fi

# ==========================================
# 10. 执行数据库迁移
# ==========================================
print_step "[10/12] 执行数据库迁移"

print_info "检查 migrate 工具..."
if ! ssh_exec "command -v migrate > /dev/null 2>&1"; then
    print_info "安装 golang-migrate 工具..."
    ssh_exec "curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && mv migrate /usr/local/bin/ && chmod +x /usr/local/bin/migrate"
    print_success "migrate 工具安装完成"
else
    print_success "migrate 工具已安装"
fi

print_info "执行数据库迁移..."
# 加载环境变量并执行迁移
if ssh_exec "cd $SERVER_DEPLOY_DIR && source .env.prod && migrate -path ./migrations -database \"postgres://\${DB_USER}:\${DB_PASSWORD}@localhost:5432/\${DB_NAME}?sslmode=disable\" up"; then
    print_success "数据库迁移完成"
else
    print_error "数据库迁移失败"
    print_info "查看迁移日志以获取详细信息"
    ssh_exec "cd $SERVER_DEPLOY_DIR && source .env.prod && migrate -path ./migrations -database \"postgres://\${DB_USER}:\${DB_PASSWORD}@localhost:5432/\${DB_NAME}?sslmode=disable\" version"
    exit 1
fi

# ==========================================
# 11. 初始化 root 用户
# ==========================================
print_step "[11/12] 初始化 root 用户"

print_info "执行用户初始化脚本..."
print_info "安装 jq 工具（如果未安装）..."
ssh_exec "command -v jq > /dev/null 2>&1 || (apt-get update && apt-get install -y jq)" || true

print_info "创建 root 用户..."
ssh_exec "cd $SERVER_DEPLOY_DIR && source .env.prod && chmod +x init-root-user.sh && ./init-root-user.sh" || {
    print_warning "root 用户初始化失败，可能已存在或 Kratos 未就绪"
    print_info "您可以稍后手动运行: ssh root@$SERVER_HOST 'cd $SERVER_DEPLOY_DIR && ./init-root-user.sh'"
}

print_success "root 用户初始化完成"

# ==========================================
# 12. 验证服务状态
# ==========================================
print_step "[12/12] 验证服务状态"

print_info "等待服务完全就绪..."
wait_for_container_healthy "tsu_admin" 10

print_info "检查容器状态..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.3-app.yml ps"

echo ""
print_info "测试 Admin Server 健康检查..."
sleep 5
if ssh_exec "curl -sf http://47.239.139.109/api/v1/admin/health > /dev/null"; then
    print_success "Admin Server 服务正常"
else
    print_warning "Admin Server 服务可能未就绪，查看日志："
    ssh_exec "docker logs tsu_admin 2>&1 | tail -30"
fi

# ==========================================
# 部署完成
# ==========================================
print_step "✅ 步骤 3 完成：主服务部署成功！"

echo ""
echo -e "${BLUE}=========================================="
echo -e "  部署信息"
echo -e "==========================================${NC}"
echo -e "${GREEN}已部署的服务：${NC}"
echo "  - Admin Server: tsu_admin (端口 8071)"
echo ""
echo -e "${GREEN}镜像信息：${NC}"
echo "  - 镜像: $IMAGE_TAG"
echo ""
echo -e "${YELLOW}管理员账号：${NC}"
echo "  - 用户名: root"
echo "  - 密码: password"
echo ""
echo -e "${RED}⚠️  重要提示：${NC}"
echo "  请在首次登录后立即修改密码！"
echo ""
echo -e "${BLUE}访问地址（内网）：${NC}"
echo "  - API: http://47.239.139.109:8071/api/"
echo "  - Swagger: http://47.239.139.109:8071/swagger/"
echo "  - Health: http://47.239.139.109:8071/health"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  运行: make deploy-prod-step4"
echo "  或: ./scripts/deploy-prod-step4-nginx.sh"
echo ""

print_success "🎉 主服务部署完成！"
