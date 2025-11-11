#!/bin/bash
# 分析性能测试结果脚本

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <timestamp>"
  echo "Example: $0 20251111_153000"
  echo ""
  echo "Available test results:"
  ls -1 test/results/performance/*.txt 2>/dev/null | sed 's/.*_\([0-9_]*\)\.txt/\1/' | sort -u || echo "  No results found"
  exit 1
fi

TIMESTAMP=$1
RESULTS_DIR="test/results/performance"

echo "==================================="
echo "Performance Analysis Report"
echo "Timestamp: $TIMESTAMP"
echo "==================================="

# 分析 Game Server 健康检查结果
echo ""
echo "1. Game Server Health Check Performance:"
echo "   (不记录 Prometheus 指标)"
echo "-----------------------------------"
if [ -f "$RESULTS_DIR/game_health_$TIMESTAMP.txt" ]; then
  grep -E "Requests/sec|Latency|Transfer" "$RESULTS_DIR/game_health_$TIMESTAMP.txt" | head -5 || true
  echo "  Latency Distribution:"
  grep -A 4 "Latency Distribution" "$RESULTS_DIR/game_health_$TIMESTAMP.txt" | tail -4 || true
else
  echo "  Result file not found"
fi

# 分析 Game Server Metrics 结果
echo ""
echo "2. Game Server Metrics Endpoint Performance:"
echo "   (包含指标序列化开销)"
echo "-----------------------------------"
if [ -f "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt" ]; then
  grep -E "Requests/sec|Latency|Transfer" "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt" | head -5 || true
  echo "  Latency Distribution:"
  grep -A 4 "Latency Distribution" "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt" | tail -4 || true
else
  echo "  Result file not found"
fi

# 分析 Admin Server 健康检查结果
echo ""
echo "3. Admin Server Health Check Performance:"
echo "-----------------------------------"
if [ -f "$RESULTS_DIR/admin_health_$TIMESTAMP.txt" ]; then
  grep -E "Requests/sec|Latency|Transfer" "$RESULTS_DIR/admin_health_$TIMESTAMP.txt" | head -5 || true
  echo "  Latency Distribution:"
  grep -A 4 "Latency Distribution" "$RESULTS_DIR/admin_health_$TIMESTAMP.txt" | tail -4 || true
else
  echo "  Result file not found"
fi

# 分析 Admin Server Metrics 结果
echo ""
echo "4. Admin Server Metrics Endpoint Performance:"
echo "-----------------------------------"
if [ -f "$RESULTS_DIR/admin_metrics_$TIMESTAMP.txt" ]; then
  grep -E "Requests/sec|Latency|Transfer" "$RESULTS_DIR/admin_metrics_$TIMESTAMP.txt" | head -5 || true
  echo "  Latency Distribution:"
  grep -A 4 "Latency Distribution" "$RESULTS_DIR/admin_metrics_$TIMESTAMP.txt" | tail -4 || true
else
  echo "  Result file not found"
fi

# 性能对比总结
echo ""
echo "==================================="
echo "Performance Comparison Summary"
echo "==================================="

extract_qps() {
  grep "Requests/sec" "$1" | awk '{print $2}' | tr -d 'k' 2>/dev/null || echo "0"
}

extract_p95() {
  grep -A 3 "Latency Distribution" "$1" | tail -1 | awk '{print $2}' 2>/dev/null || echo "0ms"
}

if [ -f "$RESULTS_DIR/game_health_$TIMESTAMP.txt" ] && [ -f "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt" ]; then
  health_qps=$(extract_qps "$RESULTS_DIR/game_health_$TIMESTAMP.txt")
  metrics_qps=$(extract_qps "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt")
  health_p95=$(extract_p95 "$RESULTS_DIR/game_health_$TIMESTAMP.txt")
  metrics_p95=$(extract_p95 "$RESULTS_DIR/game_metrics_$TIMESTAMP.txt")

  echo "Game Server:"
  echo "  Health Check QPS: $health_qps"
  echo "  Metrics QPS:      $metrics_qps"
  echo "  Health p95:       $health_p95"
  echo "  Metrics p95:      $metrics_p95"
fi

# CPU/Memory Profile 分析提示
echo ""
echo "==================================="
echo "Profile Analysis Commands"
echo "==================================="

if [ -f "$RESULTS_DIR/game_cpu_$TIMESTAMP.prof" ]; then
  echo "CPU Profile:"
  echo "  go tool pprof -http=:8080 $RESULTS_DIR/game_cpu_$TIMESTAMP.prof"
  echo ""
  echo "Top CPU consumers:"
  go tool pprof -top -nodecount=10 "$RESULTS_DIR/game_cpu_$TIMESTAMP.prof" 2>/dev/null | head -15 || echo "  (go tool pprof not available)"
fi

if [ -f "$RESULTS_DIR/game_mem_$TIMESTAMP.prof" ]; then
  echo ""
  echo "Memory Profile:"
  echo "  go tool pprof -http=:8081 $RESULTS_DIR/game_mem_$TIMESTAMP.prof"
  echo ""
  echo "Top memory allocators:"
  go tool pprof -top -nodecount=10 "$RESULTS_DIR/game_mem_$TIMESTAMP.prof" 2>/dev/null | head -15 || echo "  (go tool pprof not available)"
fi

echo ""
echo "==================================="
echo "Performance Acceptance Criteria"
echo "==================================="
echo "Based on CLAUDE.md standards:"
echo "  ✓ API p95 latency     < 200ms"
echo "  ✓ CPU overhead        < 5%"
echo "  ✓ Memory overhead     < 50MB"
echo "  ✓ p95 latency delta   < 10ms"
echo ""
echo "Review the results above to confirm compliance."
echo "==================================="
