#!/bin/bash
# 怪物 API 手动测试脚本
# 使用方法: ./test/manual/test_monster_api.sh

set -e

# 配置
BASE_URL="http://localhost:8071/api/v1/admin"
CONTENT_TYPE="Content-Type: application/json"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印函数
print_test() {
    echo -e "${YELLOW}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查服务器是否运行
check_server() {
    if ! curl -s "$BASE_URL/monsters" > /dev/null 2>&1; then
        print_error "服务器未运行，请先启动 admin-server"
        echo "运行: make run-admin"
        exit 1
    fi
    print_success "服务器正在运行"
}

# 测试1: 创建怪物
test_create_monster() {
    print_test "测试1: 创建怪物"
    
    RESPONSE=$(curl -s -X POST "$BASE_URL/monsters" \
        -H "$CONTENT_TYPE" \
        -d '{
            "monster_code": "TEST_MONSTER_API",
            "monster_name": "API测试怪物",
            "monster_level": 10,
            "description": "通过API创建的测试怪物",
            "max_hp": 500,
            "hp_recovery": 10,
            "base_str": 15,
            "base_agi": 20,
            "base_vit": 18,
            "base_int": 12,
            "drop_gold_min": 50,
            "drop_gold_max": 150,
            "drop_exp": 100,
            "is_active": true
        }')
    
    echo "$RESPONSE" | jq '.'
    
    # 提取怪物ID
    MONSTER_ID=$(echo "$RESPONSE" | jq -r '.data.id')
    
    if [ "$MONSTER_ID" != "null" ] && [ -n "$MONSTER_ID" ]; then
        print_success "创建成功，怪物ID: $MONSTER_ID"
        echo "$MONSTER_ID" > /tmp/test_monster_id.txt
        return 0
    else
        print_error "创建失败"
        return 1
    fi
}

# 测试2: 获取怪物列表
test_get_monsters() {
    print_test "测试2: 获取怪物列表"
    
    RESPONSE=$(curl -s "$BASE_URL/monsters?limit=5&offset=0")
    echo "$RESPONSE" | jq '.'
    
    COUNT=$(echo "$RESPONSE" | jq '.data.total')
    if [ "$COUNT" -gt 0 ]; then
        print_success "获取成功，共 $COUNT 个怪物"
        return 0
    else
        print_error "获取失败或无数据"
        return 1
    fi
}

# 测试3: 获取怪物详情
test_get_monster() {
    print_test "测试3: 获取怪物详情"
    
    if [ ! -f /tmp/test_monster_id.txt ]; then
        print_error "未找到测试怪物ID"
        return 1
    fi
    
    MONSTER_ID=$(cat /tmp/test_monster_id.txt)
    RESPONSE=$(curl -s "$BASE_URL/monsters/$MONSTER_ID")
    echo "$RESPONSE" | jq '.'
    
    NAME=$(echo "$RESPONSE" | jq -r '.data.monster_name')
    if [ "$NAME" = "API测试怪物" ]; then
        print_success "获取成功"
        return 0
    else
        print_error "获取失败"
        return 1
    fi
}

# 测试4: 更新怪物
test_update_monster() {
    print_test "测试4: 更新怪物"
    
    if [ ! -f /tmp/test_monster_id.txt ]; then
        print_error "未找到测试怪物ID"
        return 1
    fi
    
    MONSTER_ID=$(cat /tmp/test_monster_id.txt)
    RESPONSE=$(curl -s -X PUT "$BASE_URL/monsters/$MONSTER_ID" \
        -H "$CONTENT_TYPE" \
        -d '{
            "monster_name": "API测试怪物（已更新）",
            "max_hp": 600
        }')
    
    echo "$RESPONSE" | jq '.'
    
    NAME=$(echo "$RESPONSE" | jq -r '.data.monster_name')
    if [ "$NAME" = "API测试怪物（已更新）" ]; then
        print_success "更新成功"
        return 0
    else
        print_error "更新失败"
        return 1
    fi
}

# 测试5: 删除怪物
test_delete_monster() {
    print_test "测试5: 删除怪物"
    
    if [ ! -f /tmp/test_monster_id.txt ]; then
        print_error "未找到测试怪物ID"
        return 1
    fi
    
    MONSTER_ID=$(cat /tmp/test_monster_id.txt)
    RESPONSE=$(curl -s -X DELETE "$BASE_URL/monsters/$MONSTER_ID")
    echo "$RESPONSE" | jq '.'
    
    CODE=$(echo "$RESPONSE" | jq -r '.code')
    if [ "$CODE" = "0" ]; then
        print_success "删除成功"
        rm -f /tmp/test_monster_id.txt
        return 0
    else
        print_error "删除失败"
        return 1
    fi
}

# 主测试流程
main() {
    echo "======================================"
    echo "   怪物 API 手动测试"
    echo "======================================"
    echo ""
    
    check_server
    echo ""
    
    # 运行测试
    PASSED=0
    FAILED=0
    
    if test_create_monster; then ((PASSED++)); else ((FAILED++)); fi
    echo ""
    
    if test_get_monsters; then ((PASSED++)); else ((FAILED++)); fi
    echo ""
    
    if test_get_monster; then ((PASSED++)); else ((FAILED++)); fi
    echo ""
    
    if test_update_monster; then ((PASSED++)); else ((FAILED++)); fi
    echo ""
    
    if test_delete_monster; then ((PASSED++)); else ((FAILED++)); fi
    echo ""
    
    # 测试结果
    echo "======================================"
    echo "   测试结果"
    echo "======================================"
    print_success "通过: $PASSED"
    if [ $FAILED -gt 0 ]; then
        print_error "失败: $FAILED"
    fi
    echo ""
    
    if [ $FAILED -eq 0 ]; then
        print_success "所有测试通过！"
        exit 0
    else
        print_error "部分测试失败"
        exit 1
    fi
}

# 运行主函数
main

