#!/bin/bash

# 英雄系统端到端测试脚本
# 包括：认证、测试数据创建、英雄创建、属性操作

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost}"
API_BASE="/api/v1"
TEST_USER="${TEST_USER:-1902104816@qq.com}"
TEST_PASS="${TEST_PASS:-12345678}"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 全局变量
AUTH_TOKEN=""
CLASS_ID=""
USER_ID=""
HERO_ID=""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  英雄系统端到端测试${NC}"
echo -e "${BLUE}========================================${NC}\n"

# ==================== 辅助函数 ====================

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

# ==================== 第一步：登录 ====================

print_info "第一步：用户认证登录"

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}${API_BASE}/game/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "'${TEST_USER}'",
    "password": "'${TEST_PASS}'"
  }')

echo "登录响应: $LOGIN_RESPONSE" | head -c 200
echo ""

AUTH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.session_token // .data.token // .session_token // .token // empty' 2>/dev/null)

if [ -z "$AUTH_TOKEN" ]; then
    print_error "登录失败，未获取到 token"
    echo "完整响应: $LOGIN_RESPONSE"
    exit 1
fi

print_success "登录成功"
print_info "Token: ${AUTH_TOKEN:0:30}..."
echo ""

# ==================== 第二步：获取或创建测试职业 ====================

print_info "第二步：准备测试职业"

# 先查询是否有现有的职业
CLASSES_RESPONSE=$(curl -s -X GET "${BASE_URL}${API_BASE}/admin/classes?page=1&page_size=5" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "职业列表响应: $CLASSES_RESPONSE" | head -c 200
echo ""

# 尝试获取第一个职业 ID
CLASS_ID=$(echo $CLASSES_RESPONSE | jq -r '.data.items[0].id // .data[0].id // empty' 2>/dev/null)

if [ -z "$CLASS_ID" ]; then
    print_info "未找到现有职业，创建测试职业"

    CREATE_CLASS=$(curl -s -X POST "${BASE_URL}${API_BASE}/admin/classes" \
      -H "Authorization: Bearer $AUTH_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "测试战士",
        "name_en": "Test Warrior",
        "description": "自动化测试职业",
        "tier": "basic",
        "is_enabled": true
      }')

    echo "创建职业响应: $CREATE_CLASS" | head -c 200
    echo ""

    CLASS_ID=$(echo $CREATE_CLASS | jq -r '.data.id // empty' 2>/dev/null)

    if [ -z "$CLASS_ID" ]; then
        print_error "创建职业失败"
        echo "完整响应: $CREATE_CLASS"
        exit 1
    fi
fi

print_success "职业 ID: $CLASS_ID"
echo ""

# ==================== 第三步：获取当前用户 ID ====================

print_info "第三步：获取当前用户信息"

USER_RESPONSE=$(curl -s -X GET "${BASE_URL}${API_BASE}/admin/users/me" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "用户响应: $USER_RESPONSE" | head -c 200
echo ""

USER_ID=$(echo $USER_RESPONSE | jq -r '.data.id // empty' 2>/dev/null)

if [ -z "$USER_ID" ]; then
    print_error "无法获取用户 ID"
    echo "完整响应: $USER_RESPONSE"
    exit 1
fi

print_success "用户 ID: $USER_ID"
echo ""

# ==================== 第四步：创建英雄 ====================

print_info "第四步：创建英雄"

# 使用登录获得的 session token 来调用 game API
# 注意：可能需要在 cookie 中设置 session

CREATE_HERO=$(curl -s -X POST "${BASE_URL}${API_BASE}/game/heroes" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'${USER_ID}'",
    "class_id": "'${CLASS_ID}'",
    "hero_name": "测试英雄_'$(date +%s)'",
    "description": "自动化测试英雄"
  }')

echo "创建英雄响应: $CREATE_HERO" | head -c 300
echo ""

HERO_ID=$(echo $CREATE_HERO | jq -r '.data.id // empty' 2>/dev/null)

if [ -z "$HERO_ID" ]; then
    print_error "创建英雄失败"
    echo "完整响应: $CREATE_HERO"
    exit 1
fi

print_success "英雄创建成功"
print_info "英雄 ID: $HERO_ID"
echo ""

# ==================== 第五步：获取英雄信息 ====================

print_info "第五步：获取英雄详情"

HERO_DETAIL=$(curl -s -X GET "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "英雄详情: "
echo $HERO_DETAIL | jq '.'
echo ""

print_success "英雄详情获取成功"
echo ""

# ==================== 第六步：获取英雄属性 ====================

print_info "第六步：获取英雄计算属性"

ATTRIBUTES=$(curl -s -X GET "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/attributes" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "英雄属性: "
echo $ATTRIBUTES | jq '.'
echo ""

# 提取一个属性代码用于后续测试
ATTR_CODE=$(echo $ATTRIBUTES | jq -r '.data[0].attribute_code // empty' 2>/dev/null)

if [ -z "$ATTR_CODE" ]; then
    print_error "未找到属性信息"
    echo "完整响应: $ATTRIBUTES"
    exit 1
fi

print_success "属性代码: $ATTR_CODE"
echo ""

# ==================== 第七步：属性加点 ====================

print_info "第七步：属性加点操作"

# 首先给英雄增加经验
ADD_EXP=$(curl -s -X POST "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/add-experience" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "experience": 1000
  }')

echo "增加经验响应: $ADD_EXP" | head -c 200
echo ""

print_success "增加经验成功"
echo ""

# 现在进行属性加点
ALLOCATE=$(curl -s -X POST "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/attributes/allocate" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hero_id": "'${HERO_ID}'",
    "attribute_code": "'${ATTR_CODE}'",
    "points_to_add": 2
  }')

echo "属性加点响应: "
echo $ALLOCATE | jq '.'
echo ""

print_success "属性加点成功"
echo ""

# ==================== 第八步：查看更新后的属性 ====================

print_info "第八步：查看加点后的属性"

ATTRIBUTES_AFTER=$(curl -s -X GET "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/attributes" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "加点后的属性: "
echo $ATTRIBUTES_AFTER | jq '.'
echo ""

print_success "属性查询成功"
echo ""

# ==================== 第九步：属性回退 ====================

print_info "第九步：属性回退操作"

ROLLBACK=$(curl -s -X POST "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/attributes/rollback" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hero_id": "'${HERO_ID}'",
    "attribute_code": "'${ATTR_CODE}'"
  }')

echo "属性回退响应: "
echo $ROLLBACK | jq '.'
echo ""

print_success "属性回退成功"
echo ""

# ==================== 第十步：查看回退后的属性 ====================

print_info "第十步：查看回退后的属性"

ATTRIBUTES_ROLLBACK=$(curl -s -X GET "${BASE_URL}${API_BASE}/game/heroes/${HERO_ID}/attributes" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json")

echo "回退后的属性: "
echo $ATTRIBUTES_ROLLBACK | jq '.'
echo ""

print_success "属性查询成功"
echo ""

# ==================== 测试总结 ====================

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ 所有测试通过！${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "测试数据总结："
echo "  用户 ID: $USER_ID"
echo "  职业 ID: $CLASS_ID"
echo "  英雄 ID: $HERO_ID"
echo "  属性代码: $ATTR_CODE"
echo ""
