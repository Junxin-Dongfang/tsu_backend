-- 重命名 formula_variables 表为 metadata_dictionary,使其成为通用字典表
-- 添加 dict_category 字段用于区分不同类型的字典项

-- 1. 重命名表
ALTER TABLE game_config.formula_variables RENAME TO metadata_dictionary;

-- 2. 创建字典分类枚举类型
CREATE TYPE game_config.dict_category_enum AS ENUM ('formula', 'action_attribute', 'effect_param');

-- 3. 添加 dict_category 字段
ALTER TABLE game_config.metadata_dictionary
ADD COLUMN dict_category game_config.dict_category_enum NOT NULL DEFAULT 'formula';

-- 4. 为现有数据设置分类(都是公式变量)
UPDATE game_config.metadata_dictionary
SET dict_category = 'formula'
WHERE dict_category = 'formula';

-- 5. 添加 metadata 字段用于存储扩展信息
ALTER TABLE game_config.metadata_dictionary
ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;

-- 6. 更新索引名称
ALTER INDEX game_config.formula_variables_pkey
RENAME TO metadata_dictionary_pkey;

ALTER INDEX game_config.idx_formula_variables_code_unique
RENAME TO idx_metadata_dictionary_code_category_unique;

ALTER INDEX game_config.idx_formula_variables_scope
RENAME TO idx_metadata_dictionary_scope;

-- 7. 修改唯一索引,将 code 和 category 组合作为唯一键
DROP INDEX game_config.idx_metadata_dictionary_code_category_unique;
CREATE UNIQUE INDEX idx_metadata_dictionary_code_category_unique
ON game_config.metadata_dictionary(variable_code, dict_category)
WHERE deleted_at IS NULL;

-- 8. 添加新索引用于按分类查询
CREATE INDEX idx_metadata_dictionary_category
ON game_config.metadata_dictionary(dict_category)
WHERE deleted_at IS NULL;

-- 9. 重命名触发器
DROP TRIGGER update_formula_variables_updated_at ON game_config.metadata_dictionary;
CREATE TRIGGER update_metadata_dictionary_updated_at
BEFORE UPDATE ON game_config.metadata_dictionary
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 10. 插入动作属性字典数据
INSERT INTO game_config.metadata_dictionary
(variable_code, variable_name, variable_type, scope, data_type, description, dict_category, metadata, is_active)
VALUES
-- 基础属性
('mana_cost', '魔法消耗', '动作基础属性', 'action', 'integer', '使用该动作所需消耗的魔法值', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "basic"}'::jsonb, true),
('mp_cost', '魔法消耗(别名)', '动作基础属性', 'action', 'integer', 'mana_cost 的别名', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "basic", "alias_of": "mana_cost"}'::jsonb, true),
('cooldown', '冷却回合', '动作基础属性', 'action', 'integer', '使用后需要等待的回合数', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "basic"}'::jsonb, true),
('cooldown_turns', '冷却回合(别名)', '动作基础属性', 'action', 'integer', 'cooldown 的别名', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "basic", "alias_of": "cooldown"}'::jsonb, true),
('action_point_cost', '行动点消耗', '动作基础属性', 'action', 'integer', '使用该动作所需消耗的行动点', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "basic"}'::jsonb, true),

-- 命中率属性
('base_hit_rate', '基础命中率', '命中率配置', 'action', 'integer', '动作的基础命中率(%)', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "hit_rate"}'::jsonb, true),
('accuracy_multiplier', '命中率倍数', '命中率配置', 'action', 'float', '命中率的倍数系数', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "hit_rate"}'::jsonb, true),
('min_hit_rate', '最小命中率', '命中率配置', 'action', 'integer', '命中率的下限(%)', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "hit_rate"}'::jsonb, true),
('max_hit_rate', '最大命中率', '命中率配置', 'action', 'integer', '命中率的上限(%)', 'action_attribute',
 '{"scalable": true, "scaling_types": ["linear", "percentage", "fixed"], "category": "hit_rate"}'::jsonb, true);

-- 添加注释
COMMENT ON TABLE game_config.metadata_dictionary IS '通用元数据字典表,存储公式变量、动作属性、效果参数等各类字典数据';
COMMENT ON COLUMN game_config.metadata_dictionary.dict_category IS '字典分类: formula(公式变量), action_attribute(动作属性), effect_param(效果参数)等';
COMMENT ON COLUMN game_config.metadata_dictionary.metadata IS '扩展元数据,JSON格式,可存储如scalable(是否可配置)、scaling_types(支持的成长类型)等信息';
