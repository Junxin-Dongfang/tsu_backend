#!/bin/bash

# ==========================================
# TSU 项目部署通用函数库
# ==========================================
# 提供各部署脚本共用的函数

# ==========================================
# 配置区域
# ==========================================
SERVER_HOST="47.239.139.109"
SERVER_USER="root"
SERVER_PASSWORD="J8Do8e8Oiv"
SERVER_DEPLOY_DIR="/opt/tsu"

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# ==========================================
# 颜色输出
# ==========================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# ==========================================
# 日志函数
# ==========================================
print_step() {
    echo ""
    echo -e "${GREEN}=========================================="
    echo -e "  $1"
    echo -e "==========================================${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# ==========================================
# 环境检查函数
# ==========================================
check_sshpass() {
    if ! command -v sshpass &> /dev/null; then
        print_error "未安装 sshpass"
        
        if [[ "$OSTYPE" == "darwin"* ]]; then
            print_info "正在安装 sshpass（macOS）..."
            if command -v brew &> /dev/null; then
                brew install hudochenkov/sshpass/sshpass
            else
                print_error "请先安装 Homebrew: https://brew.sh"
                return 1
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            print_info "正在安装 sshpass（Linux）..."
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y sshpass
            elif command -v yum &> /dev/null; then
                sudo yum install -y sshpass
            else
                print_error "无法自动安装 sshpass，请手动安装"
                return 1
            fi
        else
            print_error "不支持的操作系统: $OSTYPE"
            return 1
        fi
    fi
    
    print_success "sshpass 已安装"
    return 0
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "未安装 Docker，请先安装 Docker"
        return 1
    fi
    
    print_success "Docker 已安装: $(docker --version)"
    return 0
}

# ==========================================
# SSH 操作函数
# ==========================================
ssh_exec() {
    local command=$1
    sshpass -p "$SERVER_PASSWORD" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$SERVER_USER@$SERVER_HOST" "$command" 2>/dev/null
}

ssh_copy() {
    local source=$1
    local dest=$2
    sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$source" "$SERVER_USER@$SERVER_HOST:$dest" 2>/dev/null
}

test_ssh_connection() {
    print_info "测试 SSH 连接..."
    if ssh_exec "echo 'SSH 连接成功'" > /dev/null 2>&1; then
        print_success "SSH 连接正常"
        return 0
    else
        print_error "无法连接到服务器 $SERVER_HOST"
        print_info "请检查："
        print_info "  - 服务器 IP 是否正确"
        print_info "  - SSH 端口是否开放（默认 22）"
        print_info "  - 用户名和密码是否正确"
        return 1
    fi
}

# ==========================================
# 环境变量生成函数
# ==========================================
generate_random_password() {
    local length=${1:-32}
    openssl rand -base64 48 | tr -d "=+/\n" | cut -c1-$length
}

create_env_file() {
    local env_file="$PROJECT_DIR/.env.prod"
    
    if [ -f "$env_file" ]; then
        print_warning ".env.prod 文件已存在"
        read -p "是否覆盖? (y/n): " overwrite
        if [ "$overwrite" != "y" ]; then
            print_info "使用现有的 .env.prod 文件"
            return 0
        fi
    fi
    
    print_info "生成 .env.prod 文件..."
    
    # 生成随机密码
    DB_PASSWORD=$(generate_random_password 32)
    ORY_DB_PASSWORD=$(generate_random_password 32)
    REDIS_PASSWORD=$(generate_random_password 32)
    JWT_SECRET=$(generate_random_password 64)
    KRATOS_COOKIE_SECRET=$(generate_random_password 32)
    KRATOS_CIPHER_SECRET=$(generate_random_password 32)
    
    cat > "$env_file" <<EOF
# TSU 项目生产环境配置（自动生成）
# 生成时间: $(date)

ENV=prod
SERVER_IP=47.239.139.109

# 主数据库
DB_HOST=tsu_postgres_main
DB_PORT=5432
DB_USER=tsu_user
DB_PASSWORD=$DB_PASSWORD
DB_NAME=tsu_db

# Ory 数据库
ORY_DB_HOST=tsu_postgres_ory
ORY_DB_PORT=5432
ORY_DB_USER=ory_user
ORY_DB_PASSWORD=$ORY_DB_PASSWORD
ORY_DB_NAME=ory_db

# Redis
REDIS_HOST=tsu_redis
REDIS_PORT=6379
REDIS_PASSWORD=$REDIS_PASSWORD

# Consul & NATS
CONSUL_ADDRESS=tsu_consul:8500
NATS_ADDRESS=tsu_nats:4222

# JWT
JWT_SECRET=$JWT_SECRET

# Kratos Secrets
KRATOS_COOKIE_SECRET=$KRATOS_COOKIE_SECRET
KRATOS_CIPHER_SECRET=$KRATOS_CIPHER_SECRET

# 数据库连接字符串
TSU_ADMIN_DATABASE_URL=postgres://\${DB_USER}:\${DB_PASSWORD}@\${DB_HOST}:\${DB_PORT}/\${DB_NAME}?sslmode=disable&search_path=admin,game_config,public
TSU_AUTH_DATABASE_URL=postgres://\${DB_USER}:\${DB_PASSWORD}@\${DB_HOST}:\${DB_PORT}/\${DB_NAME}?sslmode=disable&search_path=auth,public
KRATOS_DSN=postgres://\${ORY_DB_USER}:\${ORY_DB_PASSWORD}@\${ORY_DB_HOST}:\${ORY_DB_PORT}/\${ORY_DB_NAME}?sslmode=disable&search_path=kratos
KETO_DSN=postgres://\${ORY_DB_USER}:\${ORY_DB_PASSWORD}@\${ORY_DB_HOST}:\${ORY_DB_PORT}/\${ORY_DB_NAME}?sslmode=disable&search_path=keto
EOF
    
    print_success ".env.prod 文件已生成"
    print_warning "请妥善保管以下密码："
    echo ""
    echo "  主数据库密码: $DB_PASSWORD"
    echo "  Ory 数据库密码: $ORY_DB_PASSWORD"
    echo "  Redis 密码: $REDIS_PASSWORD"
    echo ""
    
    return 0
}

# ==========================================
# 服务器目录初始化
# ==========================================
init_server_directories() {
    print_info "初始化服务器目录结构..."
    
    ssh_exec "mkdir -p $SERVER_DEPLOY_DIR"
    ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/configs"
    ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/migrations"
    ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/infra/ory/prod"
    ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/infra/nginx"
    
    print_success "服务器目录初始化完成"
    return 0
}

# ==========================================
# 容器状态检查
# ==========================================
check_container_running() {
    local container_name=$1
    if ssh_exec "docker ps --filter name=$container_name --filter status=running --format '{{.Names}}' | grep -q $container_name"; then
        return 0
    else
        return 1
    fi
}

wait_for_container_healthy() {
    local container_name=$1
    local max_wait=${2:-60}
    local waited=0
    
    print_info "等待容器 $container_name 健康检查通过..."
    
    while [ $waited -lt $max_wait ]; do
        if ssh_exec "docker inspect --format='{{.State.Health.Status}}' $container_name 2>/dev/null" | grep -q "healthy"; then
            print_success "容器 $container_name 健康检查通过"
            return 0
        fi
        
        sleep 2
        waited=$((waited + 2))
        echo -n "."
    done
    
    echo ""
    print_warning "容器 $container_name 健康检查超时（${max_wait}秒）"
    print_info "容器可能仍在启动中，请检查日志"
    return 1
}

# ==========================================
# 导出所有变量和函数
# ==========================================
export SERVER_HOST SERVER_USER SERVER_PASSWORD SERVER_DEPLOY_DIR PROJECT_DIR
export GREEN YELLOW RED BLUE NC
export -f print_step print_info print_error print_success print_warning
export -f check_sshpass check_docker
export -f ssh_exec ssh_copy test_ssh_connection
export -f generate_random_password create_env_file
export -f init_server_directories
export -f check_container_running wait_for_container_healthy
