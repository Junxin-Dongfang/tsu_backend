#!/bin/bash
set -e

# ==========================================
# 游戏配置导入到本地开发环境
# ==========================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 默认导入模式（改为增量模式）
IMPORT_MODE="${IMPORT_MODE:-incremental}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo ""
    echo -e "${GREEN}=========================================="
    echo "  $1"
    echo "==========================================${NC}"
}

print_step() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

print_warning() {
    echo -e "${YELLOW}$1${NC}"
}

print_info() {
    echo -e "${BLUE}$1${NC}"
}

# ==========================================
# 1. 检查依赖
# ==========================================
print_header "游戏配置导入 - 本地开发环境"

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
            echo "示例:"
            echo "  $0                            # 清空导入（默认）"
            echo "  $0 --mode truncate            # 清空导入"
            echo "  $0 --mode incremental         # 增量导入"
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

print_step "[1/4] 检查依赖"

print_info "导入模式: $IMPORT_MODE"
if [ "$IMPORT_MODE" = "truncate" ]; then
    print_warning "  ⚠️  清空模式：将删除现有配置数据"
else
    print_info "  ℹ️  增量模式：保留现有数据，仅更新或新增"
fi
echo ""

# 检查 Docker
if ! command -v docker > /dev/null 2>&1; then
    print_error "未安装 Docker，请先安装"
    exit 1
fi
print_success "✅ Docker 已安装"

# 检查 Python3
if ! command -v python3 > /dev/null 2>&1; then
    print_error "未安装 Python3，请先安装"
    exit 1
fi
print_success "✅ Python3 已安装"

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
# 2. 检查 Python 依赖包
# ==========================================
print_step "[2/4] 检查 Python 依赖包"

# 检查 openpyxl
if ! python3 -c "import openpyxl" 2>/dev/null; then
    print_warning "⚠️  未安装 openpyxl，正在安装..."
    pip3 install openpyxl
fi
print_success "✅ openpyxl 已安装"

# 检查 psycopg2
if ! python3 -c "import psycopg2" 2>/dev/null; then
    print_warning "⚠️  未安装 psycopg2，正在安装..."
    pip3 install psycopg2-binary
fi
print_success "✅ psycopg2 已安装"

# ==========================================
# 3. 检查数据库容器
# ==========================================
print_step "[3/4] 检查数据库容器"

# 检查 postgres 容器是否运行
if ! docker ps | grep -q tsu_postgres; then
    print_error "数据库容器未运行，请先启动: docker compose -f deployments/docker-compose/docker-compose-main.local.yml up -d"
    exit 1
fi
print_success "✅ 数据库容器正在运行"

# ==========================================
# 4. 执行导入
# ==========================================
print_step "[4/4] 执行游戏配置导入"

# 从 .env 文件加载数据库配置
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
fi

# 设置数据库连接参数
# 从 MIGRATION_DATABASE_URL 中提取配置
# 格式：postgres://user:password@host:port/database?options
MIGRATION_URL=${MIGRATION_DATABASE_URL:-postgres://tsu_user:tsu_test@localhost:5432/tsu_db}

# 提取数据库参数
DB_USER=$(echo $MIGRATION_URL | sed -n 's|.*://\([^:]*\):.*|\1|p')
DB_PASSWORD=$(echo $MIGRATION_URL | sed -n 's|.*://[^:]*:\([^@]*\)@.*|\1|p')
DB_HOST=$(echo $MIGRATION_URL | sed -n 's|.*@\([^:]*\):.*|\1|p')
DB_PORT=$(echo $MIGRATION_URL | sed -n 's|.*:\([0-9]*\)/.*|\1|p')
DB_NAME=$(echo $MIGRATION_URL | sed -n 's|.*/\([^?]*\).*|\1|p')

# 如果 host 是容器名，转换为 localhost
if [ "$DB_HOST" = "tsu_postgres" ]; then
    DB_HOST="localhost"
fi

print_info() {
    echo -e "${BLUE}$1${NC}"
}

print_info "数据库配置:"
print_info "  主机: $DB_HOST:$DB_PORT"
print_info "  数据库: $DB_NAME"
print_info "  用户: $DB_USER"
print_info "  导入模式: $IMPORT_MODE"
echo ""

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
fi

# 执行导入脚本
echo ""
python3 "$IMPORT_SCRIPT" \
    --file "$EXCEL_FILE" \
    --host "$DB_HOST" \
    --port "$DB_PORT" \
    --user "$DB_USER" \
    --password "$DB_PASSWORD" \
    --database "$DB_NAME" \
    --mode "$IMPORT_MODE"

if [ $? -eq 0 ]; then
    print_header "✅ 🎉 游戏配置导入成功！"
    echo ""
    echo -e "${BLUE}导入信息：${NC}"
    echo "  Excel 文件: 游戏配置表_v1.0.0.0.xlsx"
    echo "  数据库: $DB_NAME @ $DB_HOST:$DB_PORT"
    echo ""
    echo -e "${YELLOW}📋 下一步：${NC}"
    echo "  1. 访问 Swagger 查看配置 API: http://localhost/swagger/index.html"
    echo "  2. 或使用数据库客户端连接查看: postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
    echo ""
else
    print_error "❌ 导入失败，请检查错误信息"
    exit 1
fi
