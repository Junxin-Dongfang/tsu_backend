#!/bin/bash
set -e

# ==========================================
# 游戏配置导入到生产服务器
# ==========================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 默认导入模式（改为增量模式）
IMPORT_MODE="${IMPORT_MODE:-incremental}"

# 加载公共函数
source "$SCRIPT_DIR/deploy-common.sh"

# 服务器配置
SERVER_IP="47.239.139.109"
SERVER_USER="root"
SERVER_PASSWORD="J8Do8e8Oiv"
SERVER_DEPLOY_DIR="/opt/tsu"

print_header() {
    echo ""
    echo -e "\033[0;32m=========================================="
    echo "  $1"
    echo "=========================================\033[0m"
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            IMPORT_MODE="$2"
            shift 2
            ;;
        --help|-h)
            echo "使用方法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --mode <truncate|incremental>  导入模式（默认: truncate）"
            echo "    truncate    清空现有数据后导入"
            echo "    incremental 增量导入，保留现有数据"
            echo "  --help, -h                     显示此帮助信息"
            echo ""
            echo "环境变量:"
            echo "  AUTO_CONFIRM=yes              自动确认，跳过确认提示"
            echo ""
            echo "示例:"
            echo "  $0                            # 清空导入（默认）"
            echo "  $0 --mode incremental         # 增量导入"
            echo "  AUTO_CONFIRM=yes $0           # 自动确认并导入"
            echo ""
            exit 0
            ;;
        *)
            print_error "未知选项: $1"
            echo "使用 --help 查看帮助"
            exit 1
            ;;
    esac
done

# 验证导入模式
if [[ "$IMPORT_MODE" != "truncate" && "$IMPORT_MODE" != "incremental" ]]; then
    print_error "无效的导入模式: $IMPORT_MODE"
    echo "有效值: truncate 或 incremental"
    exit 1
fi

print_header() {
    echo ""
    echo -e "\033[0;32m=========================================="
    echo "  $1"
    echo "=========================================\033[0m"
}

# ==========================================
# 1. 检查依赖
# ==========================================
print_step "[1/6] 检查本地依赖"

print_info "导入模式: $IMPORT_MODE"
if [ "$IMPORT_MODE" = "truncate" ]; then
    print_warning "  ⚠️  清空模式：将删除现有配置数据"
else
    print_info "  ℹ️  增量模式：保留现有数据，仅更新或新增"
fi
echo ""

if ! command -v sshpass > /dev/null 2>&1; then
    print_error "未安装 sshpass，请先安装: brew install sshpass"
    exit 1
fi
print_success "✅ sshpass 已安装"

# 检查 Excel 文件
EXCEL_FILE="$PROJECT_ROOT/configs/game/游戏配置表_v1.0.0.0.xlsx"
if [ ! -f "$EXCEL_FILE" ]; then
    print_error "Excel 配置文件不存在: $EXCEL_FILE"
    exit 1
fi
print_success "✅ Excel 配置文件存在"

# 检查 Python 脚本
IMPORT_SCRIPT="$PROJECT_ROOT/scripts/import_game_config.py"
if [ ! -f "$IMPORT_SCRIPT" ]; then
    print_error "导入脚本不存在: $IMPORT_SCRIPT"
    exit 1
fi
print_success "✅ 导入脚本存在"

# ==========================================
# 2. 检查服务器连接
# ==========================================
print_step "[2/6] 检查服务器连接"

if ! sshpass -p "$SERVER_PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_IP "echo '连接成功'" > /dev/null 2>&1; then
    print_error "无法连接到服务器"
    exit 1
fi
print_success "✅ 服务器连接正常"

# ==========================================
# 3. 在服务器上安装 Python 依赖
# ==========================================
print_step "[3/6] 检查并安装 Python 依赖"

print_info "检查 Python3..."
if ! ssh_exec "command -v python3 > /dev/null 2>&1"; then
    print_info "安装 Python3..."
    ssh_exec "apt-get update && apt-get install -y python3 python3-pip"
fi
print_success "✅ Python3 已安装"

print_info "检查并安装 Python 依赖包..."
# 使用 apt 安装系统包（避免 externally-managed-environment 错误）
ssh_exec "apt-get update > /dev/null 2>&1 && apt-get install -y python3-openpyxl python3-psycopg2 > /dev/null 2>&1 || true"
print_success "✅ Python 依赖已安装"

# ==========================================
# 4. 上传文件到服务器
# ==========================================
print_step "[4/6] 上传文件到服务器"

# 创建目录
ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/scripts"
ssh_exec "mkdir -p $SERVER_DEPLOY_DIR/configs/game"

print_info "上传 Excel 配置文件..."
sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$EXCEL_FILE" \
    $SERVER_USER@$SERVER_IP:$SERVER_DEPLOY_DIR/configs/game/

print_info "上传导入脚本..."
sshpass -p "$SERVER_PASSWORD" scp -o StrictHostKeyChecking=no \
    "$IMPORT_SCRIPT" \
    $SERVER_USER@$SERVER_IP:$SERVER_DEPLOY_DIR/scripts/

ssh_exec "chmod +x $SERVER_DEPLOY_DIR/scripts/import_game_config.py"

print_success "✅ 文件上传完成"

# ==========================================
# 5. 加载数据库配置
# ==========================================
print_step "[5/6] 获取数据库配置"

# 从容器中获取数据库密码
DB_PASSWORD=$(ssh_exec "docker exec tsu_postgres_main env | grep POSTGRES_PASSWORD | cut -d'=' -f2")
DB_USER="tsu_user"
DB_NAME="tsu_db"
DB_HOST="localhost"
DB_PORT="5432"

if [ -z "$DB_PASSWORD" ]; then
    print_error "无法获取数据库密码"
    exit 1
fi

print_success "✅ 数据库配置已获取"
print_info "  数据库: $DB_NAME"
print_info "  用户: $DB_USER"
print_info "  主机: $DB_HOST:$DB_PORT"

# ==========================================
# 6. 执行导入
# ==========================================
print_step "[6/6] 执行游戏配置导入"

print_info "开始导入游戏配置..."
print_info "导入模式: $IMPORT_MODE"

if [ "$IMPORT_MODE" = "truncate" ]; then
    print_warning "⚠️  清空模式：这将删除现有配置数据并重新导入"
else
    print_warning "⚠️  增量模式：将保留现有数据，仅更新或新增记录"
fi

# 确认导入
if [ "${AUTO_CONFIRM:-no}" != "yes" ]; then
    read -p "确认导入? (yes/no): " -r CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
        print_warning "取消导入"
        exit 0
    fi
else
    print_info "自动确认模式，跳过确认提示"
fi

# 执行导入脚本（使用命令行参数传递数据库配置）
print_info "正在执行导入..."
ssh_exec "cd $SERVER_DEPLOY_DIR && python3 scripts/import_game_config.py \
    --file configs/game/游戏配置表_v1.0.0.0.xlsx \
    --host $DB_HOST \
    --port $DB_PORT \
    --user $DB_USER \
    --password '$DB_PASSWORD' \
    --database $DB_NAME \
    --mode $IMPORT_MODE"

if [ $? -eq 0 ]; then
    print_success "✅ 游戏配置导入完成"
else
    print_error "❌ 导入失败，请检查错误信息"
    exit 1
fi

# ==========================================
# 完成
# ==========================================
print_header "✅ 🎉 游戏配置已成功导入到生产服务器！"

echo ""
echo -e "\033[0;34m导入信息：\033[0m"
echo "  Excel 文件: 游戏配置表_v1.0.0.0.xlsx"
echo "  数据库: $DB_NAME"
echo "  服务器: $SERVER_IP"
echo ""
echo -e "\033[0;33m⚠️  提示：\033[0m"
echo "  1. 配置已导入到 game_config schema"
echo "  2. 可以通过 API 查询配置数据"
echo "  3. 如需重新导入，再次运行此脚本即可"
echo ""
