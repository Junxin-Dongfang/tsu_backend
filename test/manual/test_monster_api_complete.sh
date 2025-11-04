#!/bin/bash
# 怪物 API 完整测试脚本
# 服务器: localhost:80
# 账号: root
# 密码: password

set -e

# 配置
BASE_URL="http://localhost:80/api/v1/admin"
USERNAME="root"
PASSWORD="password"
CONTENT_TYPE="Content-Type: application/json"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 全局变量
TOKEN=""
MONSTER_ID=""
SKILL_ID=""
DROP_POOL_ID=""

# 打印函数
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_test() {
    echo -e "${YELLOW}>>> $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 登录获取 Token
login() {
    print_test "登录获取 Token"

    RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    echo "$RESPONSE" | jq '.'

    TOKEN=$(echo "$RESPONSE" | jq -r '.data.session_token // .data.token // .data.access_token // empty')

    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        print_error "登录失败，无法获取 Token"
        echo "响应: $RESPONSE"
        exit 1
    fi

    print_success "登录成功，Token: ${TOKEN:0:20}..."
    echo ""
}

# 测试1: 创建怪物
test_create_monster() {
    print_test "测试1: 创建怪物"

    # 使用时间戳生成唯一代码
    TIMESTAMP=$(date +%s)
    MONSTER_CODE="TEST_API_MONSTER_$TIMESTAMP"

    RESPONSE=$(curl -s -X POST "$BASE_URL/monsters" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"monster_code\": \"$MONSTER_CODE\",
            \"monster_name\": \"API完整测试怪物\",
            \"monster_level\": 15,
            \"description\": \"通过完整API测试创建的怪物\",
            \"max_hp\": 800,
            \"hp_recovery\": 20,
            \"max_mp\": 200,
            \"mp_recovery\": 10,
            \"base_str\": 20,
            \"base_agi\": 25,
            \"base_vit\": 22,
            \"base_wlp\": 15,
            \"base_int\": 18,
            \"base_wis\": 16,
            \"base_cha\": 10,
            \"accuracy_formula\": \"STR*2+AGI\",
            \"dodge_formula\": \"AGI*2+WIS\",
            \"initiative_formula\": \"AGI*2+WIS\",
            \"body_resist_formula\": \"VIT*2+WLP\",
            \"magic_resist_formula\": \"WLP*2+WIS\",
            \"mental_resist_formula\": \"WIS*2+WLP\",
            \"environment_resist_formula\": \"VIT*2+WIS\",
            \"drop_gold_min\": 100,
            \"drop_gold_max\": 300,
            \"drop_exp\": 200,
            \"is_active\": true,
            \"display_order\": 10
        }")
    
    echo "$RESPONSE" | jq '.'
    
    MONSTER_ID=$(echo "$RESPONSE" | jq -r '.data.id // empty')
    
    if [ -z "$MONSTER_ID" ] || [ "$MONSTER_ID" = "null" ]; then
        print_error "创建怪物失败"
        return 1
    fi
    
    print_success "创建成功，怪物ID: $MONSTER_ID"
    echo ""
    return 0
}

# 测试2: 获取怪物列表
test_get_monsters() {
    print_test "测试2: 获取怪物列表"
    
    RESPONSE=$(curl -s "$BASE_URL/monsters?limit=10&offset=0" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "$RESPONSE" | jq '.'
    
    TOTAL=$(echo "$RESPONSE" | jq '.data.total // 0')
    
    if [ "$TOTAL" -gt 0 ]; then
        print_success "获取成功，共 $TOTAL 个怪物"
        echo ""
        return 0
    else
        print_error "获取失败或无数据"
        echo ""
        return 1
    fi
}

# 测试3: 获取怪物详情
test_get_monster() {
    print_test "测试3: 获取怪物详情"
    
    if [ -z "$MONSTER_ID" ]; then
        print_error "未找到怪物ID"
        return 1
    fi
    
    RESPONSE=$(curl -s "$BASE_URL/monsters/$MONSTER_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "$RESPONSE" | jq '.'
    
    NAME=$(echo "$RESPONSE" | jq -r '.data.monster_name // empty')
    
    if [ "$NAME" = "API完整测试怪物" ]; then
        print_success "获取成功，怪物名称: $NAME"
        echo ""
        return 0
    else
        print_error "获取失败"
        echo ""
        return 1
    fi
}

# 测试4: 更新怪物
test_update_monster() {
    print_test "测试4: 更新怪物"
    
    if [ -z "$MONSTER_ID" ]; then
        print_error "未找到怪物ID"
        return 1
    fi
    
    RESPONSE=$(curl -s -X PUT "$BASE_URL/monsters/$MONSTER_ID" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "monster_name": "API完整测试怪物（已更新）",
            "max_hp": 1000,
            "description": "更新后的描述"
        }')
    
    echo "$RESPONSE" | jq '.'
    
    NAME=$(echo "$RESPONSE" | jq -r '.data.monster_name // empty')
    HP=$(echo "$RESPONSE" | jq -r '.data.max_hp // 0')
    
    if [ "$NAME" = "API完整测试怪物（已更新）" ] && [ "$HP" = "1000" ]; then
        print_success "更新成功"
        echo ""
        return 0
    else
        print_error "更新失败"
        echo ""
        return 1
    fi
}

# 测试5: 获取怪物技能列表
test_get_monster_skills() {
    print_test "测试5: 获取怪物技能列表"

    if [ -z "$MONSTER_ID" ]; then
        print_error "未找到怪物ID"
        return 1
    fi

    RESPONSE=$(curl -s "$BASE_URL/monsters/$MONSTER_ID/skills" \
        -H "Authorization: Bearer $TOKEN")

    echo "$RESPONSE" | jq '.'

    print_success "获取技能列表成功"
    echo ""
    return 0
}

# 测试6: 获取怪物掉落列表
test_get_monster_drops() {
    print_test "测试6: 获取怪物掉落列表"

    if [ -z "$MONSTER_ID" ]; then
        print_error "未找到怪物ID"
        return 1
    fi

    RESPONSE=$(curl -s "$BASE_URL/monsters/$MONSTER_ID/drops" \
        -H "Authorization: Bearer $TOKEN")

    echo "$RESPONSE" | jq '.'

    print_success "获取掉落列表成功"
    echo ""
    return 0
}

# 测试7: 删除怪物
test_delete_monster() {
    print_test "测试7: 删除怪物"

    if [ -z "$MONSTER_ID" ]; then
        print_error "未找到怪物ID"
        return 1
    fi

    RESPONSE=$(curl -s -X DELETE "$BASE_URL/monsters/$MONSTER_ID" \
        -H "Authorization: Bearer $TOKEN")

    echo "$RESPONSE" | jq '.'

    CODE=$(echo "$RESPONSE" | jq -r '.code // 1')

    if [ "$CODE" = "100000" ] || [ "$CODE" = "0" ]; then
        print_success "删除成功"
        echo ""
        return 0
    else
        print_error "删除失败"
        echo ""
        return 1
    fi
}

# 主测试流程
main() {
    print_header "怪物 API 完整测试流程"
    echo ""

    print_info "服务器: $BASE_URL"
    print_info "用户名: $USERNAME"
    echo ""

    # 统计
    PASSED=0
    FAILED=0

    # 登录
    if ! login; then
        print_error "登录失败，终止测试"
        exit 1
    fi

    # 运行测试
    if test_create_monster; then ((PASSED++)); else ((FAILED++)); fi
    if test_get_monsters; then ((PASSED++)); else ((FAILED++)); fi
    if test_get_monster; then ((PASSED++)); else ((FAILED++)); fi
    if test_update_monster; then ((PASSED++)); else ((FAILED++)); fi
    if test_get_monster_skills; then ((PASSED++)); else ((FAILED++)); fi
    if test_get_monster_drops; then ((PASSED++)); else ((FAILED++)); fi
    if test_delete_monster; then ((PASSED++)); else ((FAILED++)); fi

    # 测试结果
    print_header "测试结果"
    echo ""
    print_success "通过: $PASSED"
    if [ $FAILED -gt 0 ]; then
        print_error "失败: $FAILED"
    else
        echo -e "${GREEN}失败: $FAILED${NC}"
    fi
    echo ""

    if [ $FAILED -eq 0 ]; then
        print_header "🎉 所有测试通过！"
        exit 0
    else
        print_header "⚠️  部分测试失败"
        exit 1
    fi
}

# 运行主函数
main
