#!/bin/bash
# 本地环境怪物配置导入脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "🎮 本地环境怪物配置导入"
echo "================================"
echo ""

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

# 执行导入
python3 scripts/game-config/import_monster_config.py \
    --env local \
    --mode "${1:-incremental}" \
    --config configs/game/monsters/monsters.json

echo ""
echo "✅ 导入完成"

