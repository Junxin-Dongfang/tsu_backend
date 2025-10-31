-- 装备套装测试数据
-- 用于单元测试和集成测试

-- 清理现有测试数据
DELETE FROM game_config.equipment_set_configs WHERE set_code LIKE 'test_%';

-- 插入测试套装数据
INSERT INTO game_config.equipment_set_configs (id, set_code, set_name, description, set_effects, is_active, created_at, updated_at, deleted_at)
VALUES
-- 测试套装1: 烈焰套装（2件套、4件套）
('11111111-1111-1111-1111-111111111111', 'test_flame_set', '测试烈焰套装', '用于测试的烈焰套装',
'[
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
    "effect_description": "4件套: 攻击力+20%, 暴击率+10%",
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
]'::jsonb, true, NOW(), NOW(), NULL),

-- 测试套装2: 冰霜套装（2件套、4件套、6件套）
('22222222-2222-2222-2222-222222222222', 'test_frost_set', '测试冰霜套装', '用于测试的冰霜套装',
'[
  {
    "piece_count": 2,
    "effect_description": "2件套: 防御力+10%",
    "out_of_combat_effects": [
      {
        "Data_type": "Status",
        "Data_ID": "DEF",
        "Bouns_type": "percent",
        "Bouns_Number": "10"
      }
    ],
    "in_combat_effects": null
  },
  {
    "piece_count": 4,
    "effect_description": "4件套: 防御力+20%, 生命值+15%",
    "out_of_combat_effects": [
      {
        "Data_type": "Status",
        "Data_ID": "DEF",
        "Bouns_type": "percent",
        "Bouns_Number": "20"
      },
      {
        "Data_type": "Status",
        "Data_ID": "HP",
        "Bouns_type": "percent",
        "Bouns_Number": "15"
      }
    ],
    "in_combat_effects": null
  },
  {
    "piece_count": 6,
    "effect_description": "6件套: 防御力+30%, 生命值+25%, 冰霜护盾",
    "out_of_combat_effects": [
      {
        "Data_type": "Status",
        "Data_ID": "DEF",
        "Bouns_type": "percent",
        "Bouns_Number": "30"
      },
      {
        "Data_type": "Status",
        "Data_ID": "HP",
        "Bouns_type": "percent",
        "Bouns_Number": "25"
      }
    ],
    "in_combat_effects": [
      {
        "Data_type": "Buff",
        "Data_ID": "frost_shield",
        "Trigger_type": "passive",
        "Trigger_chance": null
      }
    ]
  }
]'::jsonb, true, NOW(), NOW(), NULL),

-- 测试套装3: 已停用的套装
('33333333-3333-3333-3333-333333333333', 'test_inactive_set', '测试已停用套装', '用于测试的已停用套装',
'[
  {
    "piece_count": 2,
    "effect_description": "2件套: 速度+10%",
    "out_of_combat_effects": [
      {
        "Data_type": "Status",
        "Data_ID": "SPD",
        "Bouns_type": "percent",
        "Bouns_Number": "10"
      }
    ],
    "in_combat_effects": null
  }
]'::jsonb, false, NOW(), NOW(), NULL);

-- 验证插入
SELECT set_code, set_name, is_active FROM game_config.equipment_set_configs WHERE set_code LIKE 'test_%';

