-- 回滚: 将 metadata_dictionary 表恢复为 formula_variables

-- 1. 删除新增的动作属性数据
DELETE FROM game_config.metadata_dictionary
WHERE dict_category = 'action_attribute';

-- 2. 删除触发器
DROP TRIGGER update_metadata_dictionary_updated_at ON game_config.metadata_dictionary;

-- 3. 删除新增的索引
DROP INDEX IF EXISTS game_config.idx_metadata_dictionary_category;
DROP INDEX IF EXISTS game_config.idx_metadata_dictionary_code_category_unique;

-- 4. 删除 metadata 字段
ALTER TABLE game_config.metadata_dictionary
DROP COLUMN IF EXISTS metadata;

-- 5. 删除 dict_category 字段
ALTER TABLE game_config.metadata_dictionary
DROP COLUMN IF EXISTS dict_category;

-- 6. 删除枚举类型
DROP TYPE IF EXISTS game_config.dict_category_enum;

-- 7. 重命名表回原名
ALTER TABLE game_config.metadata_dictionary RENAME TO formula_variables;

-- 8. 恢复原索引名称
ALTER INDEX game_config.metadata_dictionary_pkey
RENAME TO formula_variables_pkey;

ALTER INDEX game_config.idx_metadata_dictionary_scope
RENAME TO idx_formula_variables_scope;

-- 9. 重建原唯一索引
CREATE UNIQUE INDEX idx_formula_variables_code_unique
ON game_config.formula_variables(variable_code)
WHERE deleted_at IS NULL;

-- 10. 恢复原触发器
CREATE TRIGGER update_formula_variables_updated_at
BEFORE UPDATE ON game_config.formula_variables
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
