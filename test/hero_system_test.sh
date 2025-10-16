#!/bin/bash

# 英雄系统集成测试脚本
# 测试英雄创建、属性加点、技能学习等功能

BASE_URL="http://localhost:8072/api/v1/game"
USER_ID="test_user_$(date +%s)"

echo "========================================="
echo "  英雄系统集成测试"
echo "========================================="
echo ""

# 1. 创建英雄
echo "【测试1】创建英雄..."
HERO_RESPONSE=$(curl -s -X POST "${BASE_URL}/heroes" \
  -H "Content-Type: application/json" \
  -d '{
    "class_id": "basic_warrior_class_id",
    "hero_name": "测试战士",
    "description": "这是一个测试英雄"
  }')

HERO_ID=$(echo $HERO_RESPONSE | jq -r '.data.id')
echo "创建的英雄ID: $HERO_ID"
echo "响应: $HERO_RESPONSE"
echo ""

# 2. 获取英雄详情
echo "【测试2】获取英雄详情..."
curl -s -X GET "${BASE_URL}/heroes/${HERO_ID}" | jq '.'
echo ""

# 3. 获取英雄属性
echo "【测试3】获取英雄属性..."
curl -s -X GET "${BASE_URL}/heroes/${HERO_ID}/attributes" | jq '.'
echo ""

# 4. 属性加点
echo "【测试4】属性加点 (STR +2)..."
curl -s -X POST "${BASE_URL}/heroes/${HERO_ID}/attributes/allocate" \
  -H "Content-Type: application/json" \
  -d '{
    "attribute_code": "STR",
    "points": 2
  }' | jq '.'
echo ""

# 5. 再次查看属性
echo "【测试5】查看加点后的属性..."
curl -s -X GET "${BASE_URL}/heroes/${HERO_ID}/attributes" | jq '.'
echo ""

# 6. 回退属性
echo "【测试6】回退属性加点..."
curl -s -X POST "${BASE_URL}/heroes/${HERO_ID}/attributes/rollback" \
  -H "Content-Type: application/json" \
  -d '{
    "attribute_code": "STR"
  }' | jq '.'
echo ""

# 7. 获取可学习技能
echo "【测试7】获取可学习技能..."
curl -s -X GET "${BASE_URL}/heroes/${HERO_ID}/skills/available" | jq '.'
echo ""

# 8. 学习技能
echo "【测试8】学习技能..."
curl -s -X POST "${BASE_URL}/heroes/${HERO_ID}/skills/learn" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "basic_attack_skill_id"
  }' | jq '.'
echo ""

# 9. 升级技能
echo "【测试9】升级技能..."
curl -s -X POST "${BASE_URL}/heroes/${HERO_ID}/skills/basic_attack_skill_id/upgrade" \
  -H "Content-Type: application/json" \
  -d '{
    "levels": 1
  }' | jq '.'
echo ""

# 10. 回退技能
echo "【测试10】回退技能操作..."
curl -s -X POST "${BASE_URL}/heroes/${HERO_ID}/skills/basic_attack_skill_id/rollback" \
  -H "Content-Type: application/json" | jq '.'
echo ""

echo "========================================="
echo "  测试完成！"
echo "========================================="

