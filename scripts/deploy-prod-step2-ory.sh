#!/bin/bash

# ==========================================
# TSU 项目生产环境部署 - 步骤 2: Ory 服务
# ==========================================
# 部署内容：
#   - Kratos（认证服务）
#   - Keto（权限服务）
#   - Oathkeeper（API 网关）

set -e

# 加载通用函数库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/deploy-common.sh"

print_step "步骤 2: 部署 Ory 认证授权服务"

# ==========================================
# 1. 检查依赖服务
# ==========================================
print_step "[1/6] 检查依赖服务"

print_info "检查基础设施服务..."

if ! check_container_running "tsu_postgres_ory"; then
    print_error "Ory 数据库未运行，请先执行步骤 1"
    print_info "运行: make deploy-prod-step1"
    exit 1
fi

print_success "依赖服务检查通过"

# ==========================================
# 2. 上传配置文件
# ==========================================
print_step "[2/6] 上传 Ory 配置文件"

print_info "上传 docker-compose 配置..."
ssh_copy "$PROJECT_DIR/deployments/docker-compose/docker-compose.prod.2-ory.yml" "$SERVER_DEPLOY_DIR/"

print_info "清理旧配置并重新创建目录..."
ssh_exec "rm -rf $SERVER_DEPLOY_DIR/infra/ory && mkdir -p $SERVER_DEPLOY_DIR/infra/ory/prod"

print_info "上传 Kratos 配置..."
ssh_copy "$PROJECT_DIR/infra/ory/prod/kratos.prod.yml" "$SERVER_DEPLOY_DIR/infra/ory/prod/"
ssh_copy "$PROJECT_DIR/infra/ory/identity.schema.json" "$SERVER_DEPLOY_DIR/infra/ory/"

print_info "上传 Keto 配置..."
ssh_copy "$PROJECT_DIR/infra/ory/prod/keto.prod.yml" "$SERVER_DEPLOY_DIR/infra/ory/prod/"

print_info "上传 Oathkeeper 配置..."
ssh_copy "$PROJECT_DIR/infra/ory/prod/oathkeeper.prod.yml" "$SERVER_DEPLOY_DIR/infra/ory/prod/"
ssh_copy "$PROJECT_DIR/infra/ory/prod/access-rules.prod.json" "$SERVER_DEPLOY_DIR/infra/ory/prod/"

print_success "配置文件上传完成"

# ==========================================
# 3. 启动 Ory 服务
# ==========================================
print_step "[3/6] 启动 Ory 服务"

print_info "启动服务（包含数据库迁移）..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.2-ory.yml --env-file .env.prod up -d"

print_info "等待迁移完成..."
sleep 15

# ==========================================
# 4. 检查迁移状态
# ==========================================
print_step "[4/6] 检查数据库迁移状态"

print_info "检查 Kratos 迁移..."
if ssh_exec "docker ps -a --filter name=tsu_kratos_migrate --format '{{.Status}}' | grep -q 'Exited (0)'"; then
    print_success "Kratos 迁移完成"
else
    print_warning "Kratos 迁移可能未完成，查看日志："
    ssh_exec "docker logs tsu_kratos_migrate 2>&1 | tail -20"
fi

print_info "检查 Keto 迁移..."
if ssh_exec "docker ps -a --filter name=tsu_keto_migrate --format '{{.Status}}' | grep -q 'Exited (0)'"; then
    print_success "Keto 迁移完成"
else
    print_warning "Keto 迁移可能未完成，查看日志："
    ssh_exec "docker logs tsu_keto_migrate 2>&1 | tail -20"
fi

# ==========================================
# 5. 等待服务就绪
# ==========================================
print_step "[5/6] 等待服务就绪"

wait_for_container_healthy "tsu_kratos" 90
wait_for_container_healthy "tsu_keto" 60

# Oathkeeper 没有 healthcheck，直接检查服务是否运行和端口是否响应
print_info "等待 Oathkeeper 服务启动..."
sleep 5
if ssh_exec "docker ps --filter name=tsu_oathkeeper --filter status=running --format '{{.Names}}' | grep -q tsu_oathkeeper"; then
    print_info "Oathkeeper 容器正在运行，检查端口..."
    # 等待端口响应
    for i in {1..30}; do
        if ssh_exec "curl -sf http://localhost:4456/health/ready > /dev/null 2>&1"; then
            print_success "✅ Oathkeeper 服务就绪"
            break
        fi
        if [ $i -eq 30 ]; then
            print_warning "⚠️  Oathkeeper 端口检查超时，但容器正在运行"
        else
            echo -n "."
            sleep 1
        fi
    done
else
    print_error "Oathkeeper 容器未运行"
    ssh_exec "docker logs tsu_oathkeeper --tail 30"
fi

# ==========================================
# 6. 验证服务状态
# ==========================================
print_step "[6/6] 验证服务状态"

print_info "检查容器状态..."
ssh_exec "cd $SERVER_DEPLOY_DIR && docker compose -f docker-compose.prod.2-ory.yml ps"

echo ""
print_info "测试 Kratos 健康检查..."
if ssh_exec "curl -sf http://localhost:4433/health/ready > /dev/null"; then
    print_success "Kratos 服务正常"
else
    print_warning "Kratos 服务可能未就绪"
fi

print_info "测试 Keto 健康检查..."
if ssh_exec "curl -sf http://localhost:4466/health/ready > /dev/null"; then
    print_success "Keto 服务正常"
else
    print_warning "Keto 服务可能未就绪"
fi

print_info "测试 Oathkeeper 健康检查..."
if ssh_exec "curl -sf http://localhost:4456/health/ready > /dev/null"; then
    print_success "Oathkeeper 服务正常"
else
    print_warning "Oathkeeper 服务可能未就绪"
fi

# ==========================================
# 部署完成
# ==========================================
print_step "✅ 步骤 2 完成：Ory 服务部署成功！"

echo ""
echo -e "${BLUE}已部署的服务：${NC}"
echo "  - Kratos (认证): tsu_kratos (端口 4433/4434)"
echo "  - Keto (权限): tsu_keto (端口 4466/4467)"
echo "  - Oathkeeper (网关): tsu_oathkeeper (端口 4456/4457)"
echo ""
echo -e "${BLUE}下一步：${NC}"
echo "  运行: make deploy-prod-step3"
echo "  或: ./scripts/deploy-prod-step3-app.sh"
echo ""

print_success "🎉 Ory 服务部署完成！"
