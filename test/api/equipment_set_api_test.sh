#!/bin/bash
# 装备套装管理端API测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost/api/v1/admin"
CONTENT_TYPE="Content-Type: application/json"

# 测试账号
USERNAME="root"
PASSWORD="password"

# 测试计数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 存储token和套装ID
AUTH_TOKEN=""
CREATED_SET_ID=""

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          装备套装管理端API测试                                          ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 登录函数
login() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}步骤 0: 用户登录${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${YELLOW}正在登录...${NC}"
    echo -e "  用户名: $USERNAME"

    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq "200" ]; then
        AUTH_TOKEN=$(echo "$body" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$AUTH_TOKEN" ]; then
            echo -e "${GREEN}  ✓ 登录成功${NC}"
            echo -e "  Token: ${AUTH_TOKEN:0:20}..."
            echo ""
            return 0
        else
            echo -e "${RED}  ✗ 登录失败: 无法提取token${NC}"
            echo -e "  响应: $body"
            exit 1
        fi
    else
        echo -e "${RED}  ✗ 登录失败 (HTTP $http_code)${NC}"
        echo -e "  响应: $body"
        exit 1
    fi
}

# 测试函数
test_api() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -e "${YELLOW}测试 $TOTAL_TESTS: $test_name${NC}"
    echo -e "  方法: $method"
    echo -e "  端点: $endpoint"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $AUTH_TOKEN" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $AUTH_TOKEN")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    echo -e "  响应码: $http_code"

    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}  ✓ 通过${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))

        # 如果是创建套装成功，提取ID
        if [ "$test_name" == "创建套装配置" ] && [ "$http_code" -eq "200" ]; then
            CREATED_SET_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
            echo -e "  创建的套装ID: $CREATED_SET_ID"
        fi

        return 0
    else
        echo -e "${RED}  ✗ 失败 (期望: $expected_code, 实际: $http_code)${NC}"
        echo -e "  响应: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 执行登录
login

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 1: 创建套装配置${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试1: 创建套装配置
test_api "创建套装配置" "POST" "/equipment-sets" '{
  "set_code": "test_flame_set",
  "set_name": "测试烈焰套装",
  "description": "用于API测试的烈焰套装",
  "set_effects": [
    {
      "piece_count": 2,
      "effect_description": "2件套: 攻击力+10%",
      "out_of_combat_effects": [
        {
          "Data_type": "Status",
          "Data_ID": "ATK",
          "Bouns_type": "percent",
          "Bouns_Number": "10"
        }
      ],
      "in_combat_effects": null
    },
    {
      "piece_count": 4,
      "effect_description": "4件套: 攻击力+20%, 暴击率+10%, 20%概率触发火焰爆发",
      "out_of_combat_effects": [
        {
          "Data_type": "Status",
          "Data_ID": "ATK",
          "Bouns_type": "percent",
          "Bouns_Number": "20"
        },
        {
          "Data_type": "Status",
          "Data_ID": "CRIT_RATE",
          "Bouns_type": "percent",
          "Bouns_Number": "10"
        }
      ],
      "in_combat_effects": [
        {
          "Data_type": "Skill",
          "Data_ID": "flame_burst",
          "Trigger_type": "on_attack",
          "Trigger_chance": "20"
        }
      ]
    }
  ],
  "is_active": true
}' 200

echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 2: 查询套装列表${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试2: 查询套装列表
test_api "查询套装列表（默认参数）" "GET" "/equipment-sets" "" 200
echo ""
sleep 1

# 测试3: 查询套装列表（带分页）
test_api "查询套装列表（分页）" "GET" "/equipment-sets?page=1&page_size=10" "" 200
echo ""
sleep 1

# 测试4: 查询套装列表（带搜索）
test_api "查询套装列表（搜索）" "GET" "/equipment-sets?keyword=测试" "" 200
echo ""
sleep 1

# 测试5: 查询套装列表（带筛选）
test_api "查询套装列表（筛选激活状态）" "GET" "/equipment-sets?is_active=true" "" 200
echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 3: 查询套装详情${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试6: 查询套装详情
if [ -n "$CREATED_SET_ID" ]; then
    test_api "查询套装详情" "GET" "/equipment-sets/$CREATED_SET_ID" "" 200
else
    echo -e "${RED}跳过: 未获取到套装ID${NC}"
fi
echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 4: 更新套装配置${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试7: 更新套装配置
if [ -n "$CREATED_SET_ID" ]; then
    test_api "更新套装配置" "PUT" "/equipment-sets/$CREATED_SET_ID" '{
      "set_name": "测试烈焰套装（已更新）",
      "description": "更新后的描述",
      "is_active": true
    }' 200
else
    echo -e "${RED}跳过: 未获取到套装ID${NC}"
fi
echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 5: 查询套装装备${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试8: 查询套装装备
if [ -n "$CREATED_SET_ID" ]; then
    test_api "查询套装包含的装备" "GET" "/equipment-sets/$CREATED_SET_ID/items" "" 200
else
    echo -e "${RED}跳过: 未获取到套装ID${NC}"
fi
echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 6: 查询未关联装备${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试9: 查询未关联装备
test_api "查询未关联套装的装备" "GET" "/equipment-sets/unassigned-items" "" 200
echo ""
sleep 1

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试 7: 删除套装配置${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""

# 测试10: 删除套装配置
if [ -n "$CREATED_SET_ID" ]; then
    test_api "删除套装配置" "DELETE" "/equipment-sets/$CREATED_SET_ID" "" 200
else
    echo -e "${RED}跳过: 未获取到套装ID${NC}"
fi
echo ""
sleep 1

# 测试11: 验证删除后无法查询
if [ -n "$CREATED_SET_ID" ]; then
    test_api "验证删除后无法查询" "GET" "/equipment-sets/$CREATED_SET_ID" "" 404
else
    echo -e "${RED}跳过: 未获取到套装ID${NC}"
fi
echo ""

echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}测试总结${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ 有测试失败${NC}"
    exit 1
fi

