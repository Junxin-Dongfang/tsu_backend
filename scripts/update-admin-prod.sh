#!/bin/bash

# ==========================================
# 快速更新生产环境 Admin Server
# 仅重新构建和部署 admin-server，用于文档更新等小改动
# ==========================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# 加载通用函数库
source "$SCRIPT_DIR/deploy-common.sh"

print_step "快速更新生产环境 Admin Server"

# 读取镜像配置
if [ -f "$PROJECT_DIR/.registry.conf" ]; then
    source "$PROJECT_DIR/.registry.conf"
    DOCKER_REGISTRY="${DOCKERHUB_REGISTRY:-docker.io}"
    DOCKER_USERNAME="${DOCKERHUB_USERNAME:-lilonyon}"
    DOCKER_ADMIN_IMAGE="${IMAGE_NAME:-tsu-admin-server}"
else
    print_error ".registry.conf 文件不存在"
    exit 1
fi

IMAGE_VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_TAG="${DOCKER_USERNAME}/${DOCKER_ADMIN_IMAGE}:${IMAGE_VERSION}"

echo ""
echo "镜像信息："
echo "  用户名: ${DOCKER_USERNAME}"
echo "  镜像: ${DOCKER_ADMIN_IMAGE}"
echo "  版本: ${IMAGE_VERSION}"
echo "  完整标签: ${IMAGE_TAG}"
echo ""

# ==========================================
# 1. 重新生成 Swagger 文档
# ==========================================
print_step "[1/6] 重新生成 Swagger 文档"

print_info "生成最新的 Swagger 文档..."
cd "$PROJECT_DIR"
swag init -g cmd/admin-server/main.go -o docs --parseDependency --parseInternal

print_success "Swagger 文档已更新"

# ==========================================
# 2. 构建新镜像
# ==========================================
print_step "[2/6] 构建 Admin Server 镜像"

print_info "构建镜像: $IMAGE_TAG"
docker build \
    --platform linux/amd64 \
    -f deployments/docker/Dockerfile.admin.prod \
    -t "$IMAGE_TAG" \
    --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    --build-arg VERSION="${IMAGE_VERSION}" \
    .

print_success "镜像构建完成"

# ==========================================
# 3. 保存镜像到文件
# ==========================================
print_step "[3/6] 保存镜像到文件"

TEMP_IMAGE_FILE="/tmp/tsu-admin-server-update.tar.gz"
print_info "保存镜像为 tar.gz..."
docker save "$IMAGE_TAG" | gzip > "$TEMP_IMAGE_FILE"

print_success "镜像已保存: $TEMP_IMAGE_FILE"

# ==========================================
# 4. 上传镜像到服务器
# ==========================================
print_step "[4/6] 上传镜像到服务器"

print_info "上传镜像到服务器..."
sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no "$TEMP_IMAGE_FILE" "$SERVER_USER@$SERVER_HOST:/tmp/"

print_success "镜像已上传"

# ==========================================
# 5. 在服务器加载镜像
# ==========================================
print_step "[5/6] 在服务器加载镜像"

print_info "加载镜像..."
ssh_exec "docker load < /tmp/tsu-admin-server-update.tar.gz"

print_info "清理临时文件..."
ssh_exec "rm /tmp/tsu-admin-server-update.tar.gz"
rm -f "$TEMP_IMAGE_FILE"

print_info "验证镜像..."
ssh_exec "docker images | grep tsu-admin-server | head -3"

print_success "镜像已加载到服务器"

# ==========================================
# 6. 重启 Admin Server 容器
# ==========================================
print_step "[6/6] 重启 Admin Server"

print_info "标记新镜像为 latest..."
ssh_exec "docker tag $IMAGE_TAG lilonyon/tsu-admin-server:latest"

print_info "停止当前容器..."
ssh_exec "docker stop tsu_admin || true"

print_info "删除旧容器..."
ssh_exec "docker rm tsu_admin || true"

print_info "启动新容器..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.3-admin.yml --env-file .env.prod up -d tsu_admin"

print_success "Admin Server 已重启"

# ==========================================
# 验证部署
# ==========================================
print_step "验证部署"

sleep 3
print_info "检查容器状态..."
ssh_exec "docker ps | grep tsu_admin"

print_info "检查容器日志..."
ssh_exec "docker logs tsu_admin --tail 20"

print_step "✅ Admin Server 更新完成"

echo ""
echo "访问地址："
echo "  Swagger: http://$SERVER_HOST/swagger/index.html"
echo ""
echo "如需查看日志："
echo "  ssh $SERVER_USER@$SERVER_HOST"
echo "  docker logs -f tsu_admin"
echo ""
