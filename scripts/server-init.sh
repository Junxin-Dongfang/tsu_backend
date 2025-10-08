#!/bin/bash

# ==========================================
# TSU 服务器初始化脚本
# ==========================================
# 功能：清理旧环境，安装必要软件，准备部署环境
# 使用：ssh root@server "bash -s" < server-init.sh

set -e  # 遇到错误立即退出

echo "=========================================="
echo "  TSU 服务器初始化开始"
echo "=========================================="

# ==========================================
# 1. 停止并删除所有 Docker 容器
# ==========================================
echo ""
echo "[1/7] 停止并删除所有 Docker 容器..."
if [ "$(docker ps -aq)" ]; then
    docker stop $(docker ps -aq) || true
    docker rm $(docker ps -aq) || true
    echo "✅ 已删除所有 Docker 容器"
else
    echo "ℹ️  没有运行中的容器"
fi

# ==========================================
# 2. 删除所有 Docker 镜像
# ==========================================
echo ""
echo "[2/7] 删除所有 Docker 镜像..."
if [ "$(docker images -q)" ]; then
    docker rmi $(docker images -q) -f || true
    echo "✅ 已删除所有 Docker 镜像"
else
    echo "ℹ️  没有 Docker 镜像"
fi

# ==========================================
# 3. 清理 Docker 系统（网络、卷等）
# ==========================================
echo ""
echo "[3/7] 清理 Docker 系统资源..."
docker system prune -af --volumes || true
echo "✅ Docker 系统清理完成"

# ==========================================
# 4. 停止并删除系统安装的 PostgreSQL（如果存在）
# ==========================================
echo ""
echo "[4/7] 检查并删除系统 PostgreSQL..."

# 检测操作系统类型
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    OS=$(uname -s)
fi

case "$OS" in
    ubuntu|debian)
        if systemctl list-units --full -all | grep -q postgresql; then
            echo "发现 PostgreSQL 服务，正在停止..."
            systemctl stop postgresql || true
            systemctl disable postgresql || true
            apt-get remove --purge -y postgresql* || true
            apt-get autoremove -y || true
            echo "✅ 已删除系统 PostgreSQL"
        else
            echo "ℹ️  未发现系统 PostgreSQL"
        fi
        ;;
    centos|rhel|fedora)
        if systemctl list-units --full -all | grep -q postgresql; then
            echo "发现 PostgreSQL 服务，正在停止..."
            systemctl stop postgresql || true
            systemctl disable postgresql || true
            yum remove -y postgresql* || true
            echo "✅ 已删除系统 PostgreSQL"
        else
            echo "ℹ️  未发现系统 PostgreSQL"
        fi
        ;;
    *)
        echo "⚠️  未识别的操作系统，跳过系统 PostgreSQL 检查"
        ;;
esac

# 删除 PostgreSQL 数据目录
if [ -d /var/lib/postgresql ]; then
    rm -rf /var/lib/postgresql
    echo "✅ 已删除 PostgreSQL 数据目录"
fi

# ==========================================
# 5. 更新系统并安装必要软件
# ==========================================
echo ""
echo "[5/7] 更新系统并安装必要软件..."

case "$OS" in
    ubuntu|debian)
        apt-get update
        apt-get install -y curl wget git vim net-tools htop
        
        # 安装 Docker（如果未安装）
        if ! command -v docker &> /dev/null; then
            echo "正在安装 Docker..."
            curl -fsSL https://get.docker.com | sh
            systemctl enable docker
            systemctl start docker
            echo "✅ Docker 安装完成"
        else
            echo "ℹ️  Docker 已安装: $(docker --version)"
        fi
        
        # 检查 Docker Compose（V2 已内置在 Docker 中）
        if docker compose version &> /dev/null; then
            echo "ℹ️  Docker Compose 已安装: $(docker compose version)"
        else
            echo "⚠️  Docker Compose V2 未安装，将安装 Docker Compose 插件..."
            DOCKER_COMPOSE_VERSION="2.24.0"
            mkdir -p /usr/local/lib/docker/cli-plugins
            curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/lib/docker/cli-plugins/docker-compose
            chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
            echo "✅ Docker Compose 插件安装完成"
        fi
        ;;
        
    centos|rhel|fedora)
        yum update -y
        yum install -y curl wget git vim net-tools htop
        
        # 安装 Docker
        if ! command -v docker &> /dev/null; then
            echo "正在安装 Docker..."
            curl -fsSL https://get.docker.com | sh
            systemctl enable docker
            systemctl start docker
            echo "✅ Docker 安装完成"
        else
            echo "ℹ️  Docker 已安装: $(docker --version)"
        fi
        
        # 检查 Docker Compose（V2 已内置在 Docker 中）
        if docker compose version &> /dev/null; then
            echo "ℹ️  Docker Compose 已安装: $(docker compose version)"
        else
            echo "⚠️  Docker Compose V2 未安装，将安装 Docker Compose 插件..."
            DOCKER_COMPOSE_VERSION="2.24.0"
            mkdir -p /usr/local/lib/docker/cli-plugins
            curl -L "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/lib/docker/cli-plugins/docker-compose
            chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
            echo "✅ Docker Compose 插件安装完成"
        fi
        ;;
        
    *)
        echo "⚠️  未识别的操作系统，请手动安装 Docker 和 Docker Compose"
        ;;
esac

# ==========================================
# 6. 创建应用目录结构
# ==========================================
echo ""
echo "[6/7] 创建应用目录结构..."
mkdir -p /opt/tsu/{app,backups,logs}
echo "✅ 目录结构创建完成"

# ==========================================
# 7. 配置防火墙（可选）
# ==========================================
echo ""
echo "[7/7] 配置防火墙端口..."

# 如果使用 ufw（Ubuntu/Debian）
if command -v ufw &> /dev/null; then
    ufw allow 8071/tcp comment "TSU Admin Service"
    ufw allow 8070/tcp comment "TSU Swagger UI"
    echo "✅ 防火墙规则已添加（ufw）"
# 如果使用 firewalld（CentOS/RHEL）
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=8071/tcp
    firewall-cmd --permanent --add-port=8070/tcp
    firewall-cmd --reload
    echo "✅ 防火墙规则已添加（firewalld）"
else
    echo "⚠️  未检测到防火墙管理工具，请手动配置端口：8070, 8071"
fi

# ==========================================
# 完成
# ==========================================
echo ""
echo "=========================================="
echo "  ✅ 服务器初始化完成！"
echo "=========================================="
echo ""
echo "系统信息："
echo "  - OS: $OS"
echo "  - Docker: $(docker --version)"
echo "  - Docker Compose: $(docker compose version 2>/dev/null || echo 'Not installed')"
echo "  - 应用目录: /opt/tsu/"
echo ""
echo "下一步："
echo "  1. 上传项目代码到 /opt/tsu/app/"
echo "  2. 配置 .env.prod 文件"
echo "  3. 运行部署脚本"
echo ""
