#!/bin/bash
# API 性能基准测试脚本
# 使用 wrk 工具进行 HTTP 压力测试

set -e

TARGET_URL=${1:-"http://localhost:8072/health"}
DURATION=${2:-"60s"}
CONNECTIONS=${3:-"100"}
THREADS=${4:-"4"}

echo "==================================="
echo "API Performance Benchmark"
echo "==================================="
echo "Target URL: $TARGET_URL"
echo "Duration: $DURATION"
echo "Connections: $CONNECTIONS"
echo "Threads: $THREADS"
echo "==================================="

# 检查 wrk 是否安装
if ! command -v wrk &> /dev/null; then
    echo "Error: wrk is not installed"
    echo "Install with: brew install wrk (macOS) or sudo apt-get install wrk (Linux)"
    exit 1
fi

wrk -t$THREADS -c$CONNECTIONS -d$DURATION --latency $TARGET_URL

echo ""
echo "==================================="
echo "Test completed at $(date)"
echo "==================================="
