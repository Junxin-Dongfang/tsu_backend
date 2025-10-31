#!/bin/bash

# 集成测试运行脚本
# 使用方式: ./scripts/development/run-integration-tests.sh

set -e

echo "🧪 运行集成测试..."
echo ""

# 检查数据库连接
echo "📊 检查数据库连接..."
if ! psql -h localhost -p 5432 -U postgres -d tsu_db -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ 无法连接到数据库"
    echo "请确保数据库正在运行: make dev-up"
    exit 1
fi

echo "✅ 数据库连接正常"
echo ""

# 设置测试数据库URL
export TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=tsu_db sslmode=disable"

# 运行集成测试
echo "🚀 运行集成测试..."
go test -tags=integration -v ./internal/modules/admin/service/... -run Integration

echo ""
echo "✅ 集成测试完成！"

