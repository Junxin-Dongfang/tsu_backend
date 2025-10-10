#!/usr/bin/env python3
"""
测试技能解锁动作 API
"""

import requests
import json

# 配置
BASE_URL = "http://localhost:8071/api/v1/admin"
HEADERS = {
    "Content-Type": "application/json",
    "X-User-ID": "test-user-123"  # 模拟认证后的用户ID
}

# 测试数据 (从数据库查询得到)
SKILL_ID = "0199cd58-bad1-7004-ad00-239177db181a"  # 火球术
ACTION_ID_1 = "0199cd58-bad7-7507-b6a4-f47dad40f29a"  # 火球术动作
ACTION_ID_2 = "0199cd58-bad6-7ead-80f6-ac5f00ac7189"  # 剑类武器攻击


def print_response(title, response):
    """打印响应信息"""
    print(f"\n{'=' * 60}")
    print(f"{title}")
    print(f"{'=' * 60}")
    print(f"状态码: {response.status_code}")
    print(f"响应内容:")
    try:
        print(json.dumps(response.json(), indent=2, ensure_ascii=False))
    except:
        print(response.text)


def test_add_single_unlock_action():
    """测试添加单个解锁动作"""
    url = f"{BASE_URL}/skills/{SKILL_ID}/unlock-actions"
    payload = {
        "action_id": ACTION_ID_1,
        "unlock_level": 1,
        "is_default": True,
        "level_scaling_config": {
            "damage": {
                "type": "linear",
                "base": 10,
                "value": 2
            },
            "range": {
                "type": "constant",
                "value": 30
            }
        }
    }

    response = requests.post(url, json=payload, headers=HEADERS)
    print_response("测试1: 添加单个技能解锁动作", response)
    return response


def test_get_unlock_actions():
    """测试查询技能的所有解锁动作"""
    url = f"{BASE_URL}/skills/{SKILL_ID}/unlock-actions"

    response = requests.get(url, headers=HEADERS)
    print_response("测试2: 查询技能的所有解锁动作", response)
    return response


def test_batch_set_unlock_actions():
    """测试批量设置解锁动作"""
    url = f"{BASE_URL}/skills/{SKILL_ID}/unlock-actions/batch"
    payload = {
        "actions": [
            {
                "action_id": ACTION_ID_1,
                "unlock_level": 1,
                "is_default": True,
                "level_scaling_config": {
                    "damage": {
                        "type": "linear",
                        "base": 10,
                        "value": 2
                    }
                }
            },
            {
                "action_id": ACTION_ID_2,
                "unlock_level": 3,
                "is_default": False,
                "level_scaling_config": {
                    "damage": {
                        "type": "exponential",
                        "base": 20,
                        "exponent": 1.5
                    }
                }
            }
        ]
    }

    response = requests.post(url, json=payload, headers=HEADERS)
    print_response("测试3: 批量设置技能解锁动作", response)
    return response


def test_remove_unlock_action(unlock_action_id):
    """测试删除解锁动作"""
    url = f"{BASE_URL}/skills/{SKILL_ID}/unlock-actions/{unlock_action_id}"

    response = requests.delete(url, headers=HEADERS)
    print_response(f"测试4: 删除解锁动作 (ID: {unlock_action_id})", response)
    return response


if __name__ == "__main__":
    print("=" * 60)
    print("技能解锁动作 API 测试")
    print("=" * 60)
    print(f"测试技能ID: {SKILL_ID} (火球术)")
    print(f"测试动作ID 1: {ACTION_ID_1} (火球术动作)")
    print(f"测试动作ID 2: {ACTION_ID_2} (剑类武器攻击)")

    # 测试1: 添加单个解锁动作
    test_add_single_unlock_action()

    # 测试2: 查询解锁动作
    response = test_get_unlock_actions()

    # 测试3: 批量设置解锁动作
    test_batch_set_unlock_actions()

    # 测试4: 再次查询验证批量设置
    test_get_unlock_actions()

    print(f"\n{'=' * 60}")
    print("所有测试完成!")
    print(f"{'=' * 60}\n")
