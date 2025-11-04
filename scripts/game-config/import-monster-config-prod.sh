#!/bin/bash
# 生产环境怪物配置导入脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🎮 生产环境怪物配置导入"
echo "================================"
echo ""
echo "⚠️  警告: 您正在向生产环境导入配置！"
echo ""
read -p "确认继续？(yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "❌ 已取消"
    exit 1
fi

# 检查 Python 环境
if ! command -v python3 &> /dev/null; then
    echo "❌ 错误: 未找到 python3"
    exit 1
fi

# 检查配置文件
if [ ! -f "configs/game/monsters/monsters.json" ]; then
    echo "❌ 错误: 配置文件不存在: configs/game/monsters/monsters.json"
    exit 1
fi

# 检查环境变量
if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
    echo "❌ 错误: 缺少数据库环境变量 (DB_HOST, DB_NAME, DB_USER, DB_PASSWORD)"
    exit 1
fi

# 执行导入
python3 scripts/game-config/import_monster_config.py \
    --env prod \
    --mode "${1:-incremental}" \
    --config configs/game/monsters/monsters.json

echo ""
echo "✅ 导入完成"

