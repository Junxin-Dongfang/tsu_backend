#!/bin/bash

# ==========================================
# TSU 项目生产环境部署 - 步骤 4: Game Server
# ==========================================
# 部署内容：
#   1. 构建 Game Server Docker 镜像
#   2. 保存并上传镜像到服务器
#   3. 部署 Game Server

set -e

# 加载通用函数库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/deploy-common.sh"

print_step "步骤 4: 部署 Game Server"

# ==========================================
# 1. 检查依赖服务
# ==========================================
print_step "[1/9] 检查依赖服务"

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

print_info "检查 Admin Server..."
if ! check_container_running "tsu_admin"; then
    print_error "Admin Server 未运行，请先执行步骤 3"
    exit 1
fi

print_success "依赖服务检查通过"

# ==========================================
# 2. 检查 Docker Hub 配置
# ==========================================
print_step "[2/9] 检查 Docker Hub 配置"

if [ ! -f "$PROJECT_DIR/.registry.conf" ]; then
    print_warning ".registry.conf 不存在，将使用环境变量中的 DOCKERHUB_USERNAME/DOCKERHUB_TOKEN（若跳过本地构建可不需要 token）"
else
    source "$PROJECT_DIR/.registry.conf"
fi

DOCKERHUB_USERNAME="${DOCKERHUB_USERNAME:-$DOCKER_USER}"

# 检查 IMAGE_VERSION，如果未定义则使用 latest
IMAGE_VERSION="${IMAGE_VERSION:-latest}"

if [ -z "$IMAGE_VERSION" ]; then
    print_error "IMAGE_VERSION 未定义，请检查 .registry.conf 文件"
    exit 1
fi

print_success "Docker Hub 配置检查通过"
print_info "镜像版本: $IMAGE_VERSION"

# 构建镜像标签
GAME_IMAGE_TAG="${DOCKERHUB_USERNAME}/tsu-game-server:${IMAGE_VERSION}"

# ==========================================
# 3. 执行数据库迁移（生产）
# ==========================================
print_step "[3/9] 执行数据库迁移"

ensure_remote_migrate
ssh_exec "cd $SERVER_DEPLOY_DIR && source .env.prod && url=postgres://\\${DB_USER}:\\${DB_PASSWORD}@localhost:5432/\\${DB_NAME}?sslmode=disable && migrate -path ./migrations -database \"\$url\" up"

# ==========================================
# 4. 构建 Game Server Docker 镜像
# ==========================================
print_step "[4/9] 构建 Game Server 镜像"

if [ "${SKIP_BUILD:-true}" = "true" ]; then
    print_info "跳过本地构建（SKIP_BUILD=true），后续将直接拉取/加载镜像"
else
    print_info "开始构建 Game Server 镜像: $GAME_IMAGE_TAG"
    print_info "这可能需要几分钟时间..."

    cd "$PROJECT_DIR"
    export DOCKER_BUILDKIT=1
    docker build \
        --platform linux/amd64 \
        -f deployments/docker/Dockerfile.game.prod \
        -t "$GAME_IMAGE_TAG" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        --build-arg VERSION="${IMAGE_VERSION}" \
        .

    print_success "Game Server 镜像构建完成"
fi

# ==========================================
# 5. 保存镜像到文件
# ==========================================
print_step "[5/9] 保存镜像到文件"

if [ "${SKIP_BUILD:-true}" = "true" ]; then
    print_info "跳过镜像保存（SKIP_BUILD=true）"
else
    print_info "保存 Game Server 镜像为 tar.gz..."
    TEMP_GAME_IMAGE_FILE="/tmp/tsu-game-server.tar.gz"
    docker save "$GAME_IMAGE_TAG" | gzip > "$TEMP_GAME_IMAGE_FILE"
    print_success "镜像已保存"
fi

# ==========================================
# 6. 上传镜像到服务器
# ==========================================
print_step "[6/9] 上传镜像到服务器"

if [ "${SKIP_BUILD:-true}" = "true" ]; then
    print_info "跳过本地构建与镜像上传，直接在服务器拉取镜像: $GAME_IMAGE_TAG"
    ssh_exec "docker pull $GAME_IMAGE_TAG"
else
    print_info "上传 Game Server 镜像（约 30MB，需要几分钟）..."
    ssh_copy "$TEMP_GAME_IMAGE_FILE" "/tmp/"
    print_info "在服务器加载 Game Server 镜像..."
    ssh_exec "docker load < /tmp/tsu-game-server.tar.gz && rm /tmp/tsu-game-server.tar.gz && docker images | grep tsu-game-server"
    rm -f "$TEMP_GAME_IMAGE_FILE"
    print_success "镜像已加载到服务器"
fi

# ==========================================
# 7. 上传配置文件到服务器
# ==========================================
print_step "[7/9] 上传配置文件到服务器"

print_info "上传 docker-compose 配置..."
ssh_copy "$PROJECT_DIR/deployments/docker-compose/docker-compose.prod.4-game.yml" "$SERVER_DEPLOY_DIR/"

print_info "上传 configs 目录..."
ssh_copy "$PROJECT_DIR/configs" "$SERVER_DEPLOY_DIR/"

print_success "配置文件上传完成"

# ==========================================
# 8. 启动 Game Server
# ==========================================
print_step "[8/9] 启动 Game Server"

print_info "停止旧的 Game Server 容器（如果存在）..."
ssh_exec "docker stop tsu_game 2>/dev/null || true"
ssh_exec "docker rm tsu_game 2>/dev/null || true"

print_info "设置镜像环境变量..."
ssh_exec "cd $SERVER_DEPLOY_DIR && sed -i '/^DOCKER_GAME_IMAGE=/d' .env.prod 2>/dev/null || true"
ssh_exec "cd $SERVER_DEPLOY_DIR && echo 'DOCKER_GAME_IMAGE=$GAME_IMAGE_TAG' >> .env.prod"

print_info "启动 Game Server..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.4-game.yml --env-file .env.prod up -d"

print_info "等待服务启动..."
sleep 15

# ==========================================
# 9. 验证服务状态
# ==========================================
print_step "[9/9] 验证 Game Server 服务状态"

print_info "等待服务完全就绪..."
wait_for_container_healthy "tsu_game" 10

print_info "检查容器状态..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.4-game.yml ps"

echo ""
print_info "测试 Game Server 健康检查..."
sleep 5
if ssh_exec "curl -sf http://47.239.139.109:8061/health > /dev/null"; then
    print_success "Game Server 服务正常"
else
    print_warning "Game Server 服务可能未就绪，查看日志："
    ssh_exec "docker logs tsu_game 2>&1 | tail -30"
fi

# ==========================================
# 部署完成
# ==========================================
print_step "✅ 步骤 4 完成：Game Server 部署成功！"

echo ""
echo -e "${BLUE}=========================================="
echo -e "  部署信息"
echo -e "==========================================${NC}"
echo -e "${GREEN}已部署的服务：${NC}"
echo "  - Game Server: tsu_game (端口 8061)"
echo ""
echo -e "${GREEN}镜像信息：${NC}"
echo "  - Game: $GAME_IMAGE_TAG"
echo ""
echo -e "${BLUE}访问地址（内网）：${NC}"
echo "  - Game API: http://47.239.139.109:8061/api/"
echo "  - Game Swagger: http://47.239.139.109:8061/swagger/"
echo "  - Health: http://47.239.139.109:8061/health"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  运行: make deploy-prod-step5"
echo "  或: ./scripts/deployment/deploy-prod-step5-nginx.sh"
echo ""

print_success "🎉 Game Server 部署完成！"
