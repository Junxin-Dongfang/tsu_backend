#!/bin/bash
# 运行完整的 Prometheus 性能测试套件

set -e

OUTPUT_DIR="test/results/performance"
mkdir -p $OUTPUT_DIR

TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "==================================="
echo "Prometheus Performance Test Suite"
echo "Timestamp: $TIMESTAMP"
echo "==================================="

# 检查服务是否运行
echo "Checking if services are running..."
if ! curl -s http://localhost:8072/health > /dev/null; then
    echo "Error: Game server is not running on port 8072"
    echo "Start with: make dev-up"
    exit 1
fi

if ! curl -s http://localhost:8071/health > /dev/null; then
    echo "Error: Admin server is not running on port 8071"
    echo "Start with: make dev-up"
    exit 1
fi

echo "Services are ready!"
echo ""

# 测试 1: 健康检查端点(Game Server)
echo "Test 1: Game Server Health Check Endpoint"
echo "-------------------------------------------"
./test/scripts/benchmark-api.sh "http://localhost:8072/health" 60s 100 4 \
  | tee "$OUTPUT_DIR/game_health_$TIMESTAMP.txt"

sleep 5

# 测试 2: Metrics 端点(Game Server)
echo ""
echo "Test 2: Game Server Metrics Endpoint"
echo "-------------------------------------------"
./test/scripts/benchmark-api.sh "http://localhost:8072/metrics" 60s 100 4 \
  | tee "$OUTPUT_DIR/game_metrics_$TIMESTAMP.txt"

sleep 5

# 测试 3: 健康检查端点(Admin Server)
echo ""
echo "Test 3: Admin Server Health Check Endpoint"
echo "-------------------------------------------"
./test/scripts/benchmark-api.sh "http://localhost:8071/health" 60s 100 4 \
  | tee "$OUTPUT_DIR/admin_health_$TIMESTAMP.txt"

sleep 5

# 测试 4: Metrics 端点(Admin Server)
echo ""
echo "Test 4: Admin Server Metrics Endpoint"
echo "-------------------------------------------"
./test/scripts/benchmark-api.sh "http://localhost:8071/metrics" 60s 100 4 \
  | tee "$OUTPUT_DIR/admin_metrics_$TIMESTAMP.txt"

sleep 5

# 测试 5: CPU Profile (Game Server)
echo ""
echo "Test 5: Collecting CPU profile (Game Server)"
echo "-------------------------------------------"
if curl -s http://localhost:8072/debug/pprof/profile?seconds=30 \
  > "$OUTPUT_DIR/game_cpu_$TIMESTAMP.prof"; then
    echo "CPU profile saved: $OUTPUT_DIR/game_cpu_$TIMESTAMP.prof"
else
    echo "Warning: Could not collect CPU profile (pprof may not be enabled)"
fi

sleep 2

# 测试 6: Memory Profile (Game Server)
echo ""
echo "Test 6: Collecting memory profile (Game Server)"
echo "-------------------------------------------"
if curl -s http://localhost:8072/debug/pprof/heap \
  > "$OUTPUT_DIR/game_mem_$TIMESTAMP.prof"; then
    echo "Memory profile saved: $OUTPUT_DIR/game_mem_$TIMESTAMP.prof"
else
    echo "Warning: Could not collect memory profile (pprof may not be enabled)"
fi

echo ""
echo "==================================="
echo "All tests completed!"
echo "Results saved to: $OUTPUT_DIR"
echo "==================================="
echo ""
echo "Next steps:"
echo "1. Analyze results: ./test/scripts/analyze-results.sh $TIMESTAMP"
echo "2. View CPU profile: go tool pprof -http=:8080 $OUTPUT_DIR/game_cpu_$TIMESTAMP.prof"
echo "3. View memory profile: go tool pprof -http=:8081 $OUTPUT_DIR/game_mem_$TIMESTAMP.prof"
