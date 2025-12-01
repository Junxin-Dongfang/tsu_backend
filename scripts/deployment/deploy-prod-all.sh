#!/bin/bash

# ==========================================
# TSU 项目生产环境一键部署脚本
# ==========================================
# 按顺序执行所有部署步骤：
#   步骤 1: 基础设施
#   步骤 2: Ory 服务
#   步骤 3: Admin Server
#   步骤 4: Game Server
#   步骤 5: Nginx

set -e

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 加载通用函数库
source "$SCRIPT_DIR/deploy-common.sh"

# 命令行参数
AUTO_MODE=false
SKIP_CONFIRM=false
SKIP_ENV_UPLOAD=false
ALLOW_RECREATE=false

for arg in "$@"; do
    case $arg in
        --auto)
            AUTO_MODE=true
            SKIP_CONFIRM=true
            shift
            ;;
        --yes|-y)
            SKIP_CONFIRM=true
            shift
            ;;
        --skip-env-upload)
            SKIP_ENV_UPLOAD=true
            shift
            ;;
        --allow-recreate)
            ALLOW_RECREATE=true
            shift
            ;;
        --auth-mode=*)
            SSH_AUTH_MODE="${arg#*=}"
            shift
            ;;
        --no-color)
            NO_COLOR=1
            shift
            ;;
        *)
            ;;
    esac
done

print_step "TSU 项目生产环境一键部署"

echo ""
echo -e "${BLUE}部署计划：${NC}"
echo "  步骤 1: 基础设施（PostgreSQL、Redis、NATS、Consul）"
echo "  步骤 2: Ory 服务（Kratos、Keto、Oathkeeper）"
echo "  步骤 3: Admin Server（后台管理服务 + 数据库迁移）"
echo "  步骤 4: Game Server（游戏服务）"
echo "  步骤 5: Nginx（反向代理）"
echo ""
echo -e "${YELLOW}目标服务器：${NC}$SERVER_HOST"
echo -e "${YELLOW}部署目录：${NC}$SERVER_DEPLOY_DIR"
echo ""

if [ "$SKIP_CONFIRM" = false ]; then
    read -p "确认开始部署？(y/n): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "部署已取消"
        exit 0
    fi
fi

# ==========================================
# 记录开始时间
# ==========================================
START_TIME=$(date +%s)

# ==========================================
# 步骤 1: 基础设施
# ==========================================
print_step "执行步骤 1: 基础设施"

if SKIP_ENV_UPLOAD=$SKIP_ENV_UPLOAD ALLOW_RECREATE=$ALLOW_RECREATE SSH_AUTH_MODE=$SSH_AUTH_MODE NO_COLOR=$NO_COLOR bash "$SCRIPT_DIR/deploy-prod-step1-infra.sh"; then
    print_success "步骤 1 完成"
else
    print_error "步骤 1 失败，部署中止"
    exit 1
fi

if [ "$AUTO_MODE" = false ]; then
    echo ""
    read -p "按回车键继续到步骤 2..."
fi

# ==========================================
# 步骤 2: Ory 服务
# ==========================================
print_step "执行步骤 2: Ory 服务"

if ALLOW_RECREATE=$ALLOW_RECREATE SSH_AUTH_MODE=$SSH_AUTH_MODE NO_COLOR=$NO_COLOR bash "$SCRIPT_DIR/deploy-prod-step2-ory.sh"; then
    print_success "步骤 2 完成"
else
    print_error "步骤 2 失败，部署中止"
    exit 1
fi

if [ "$AUTO_MODE" = false ]; then
    echo ""
    read -p "按回车键继续到步骤 3..."
fi

# ==========================================
# 步骤 3: Admin Server
# ==========================================
print_step "执行步骤 3: Admin Server"

if ALLOW_RECREATE=$ALLOW_RECREATE SSH_AUTH_MODE=$SSH_AUTH_MODE NO_COLOR=$NO_COLOR bash "$SCRIPT_DIR/deploy-prod-step3-admin.sh"; then
    print_success "步骤 3 完成"
else
    print_error "步骤 3 失败，部署中止"
    exit 1
fi

if [ "$AUTO_MODE" = false ]; then
    echo ""
    read -p "按回车键继续到步骤 4..."
fi

# ==========================================
# 步骤 4: Game Server
# ==========================================
print_step "执行步骤 4: Game Server"

if bash "$SCRIPT_DIR/deploy-prod-step4-game.sh"; then
    print_success "步骤 4 完成"
else
    print_error "步骤 4 失败，但前面的服务已部署"
    exit 1
fi

if [ "$AUTO_MODE" = false ]; then
    echo ""
    read -p "按回车键继续到步骤 5..."
fi

# ==========================================
# 步骤 5: Nginx
# ==========================================
print_step "执行步骤 5: Nginx"

if bash "$SCRIPT_DIR/deploy-prod-step5-nginx.sh"; then
    print_success "步骤 5 完成"
else
    print_error "步骤 5 失败，但前面的服务已部署"
    exit 1
fi

# ==========================================
# 计算部署时间
# ==========================================
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
MINUTES=$((DURATION / 60))
SECONDS=$((DURATION % 60))

# ==========================================
# 部署完成
# ==========================================
print_step "🎉 完整部署成功！"

echo ""
echo -e "${GREEN}=========================================="
echo -e "  部署总结"
echo -e "==========================================${NC}"
echo ""
echo -e "${BLUE}部署时间：${NC}${MINUTES} 分 ${SECONDS} 秒"
echo ""
echo -e "${BLUE}已部署的所有服务：${NC}"
echo ""
echo "  【第一层 - 基础设施】"
echo "    ✓ PostgreSQL 主数据库 (端口 5432)"
echo "    ✓ PostgreSQL Ory数据库 (端口 5433)"
echo "    ✓ Redis 缓存 (端口 6379)"
echo "    ✓ NATS 消息队列 (端口 4222)"
echo "    ✓ Consul 服务发现 (端口 8500)"
echo ""
echo "  【第二层 - Ory 服务】"
echo "    ✓ Kratos 认证服务 (端口 4433/4434)"
echo "    ✓ Keto 权限服务 (端口 4466/4467)"
echo "    ✓ Oathkeeper API网关 (端口 4456/4457)"
echo ""
echo "  【第三层 - 主服务】"
echo "    ✓ Admin Server 后台管理服务 (端口 8071)"
echo "    ✓ Game Server 游戏服务 (端口 8061)"
echo "    ✓ 数据库迁移已完成"
echo "    ✓ Root 用户已初始化"
echo ""
echo "  【第四层 - 接入层】"
echo "    ✓ Nginx 反向代理 (端口 80/443)"
echo ""
echo -e "${GREEN}=========================================="
echo -e "  访问信息"
echo -e "==========================================${NC}"
echo ""
echo "  主入口: http://47.239.139.109/"
echo "  Admin API: http://47.239.139.109/api/"
echo "  Swagger UI: http://47.239.139.109/swagger/"
echo ""
echo -e "${YELLOW}管理员账号：${NC}"
echo "  用户名: root"
echo "  密码: password"
echo ""
echo -e "${RED}⚠️  重要提示：${NC}"
echo "  1. 请立即登录并修改默认密码"
echo "  2. 检查服务器防火墙配置"
echo "  3. 考虑配置 HTTPS"
echo "  4. 设置定期数据库备份"
echo ""

print_success "部署完成！系统已就绪！"
