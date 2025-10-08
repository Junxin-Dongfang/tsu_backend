-- ============================================
-- 技能升级系统重新设计
-- ============================================

-- 1. 全局技能升级消耗表（所有技能共用）
CREATE TABLE IF NOT EXISTS game_config.skill_upgrade_costs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    level_number INTEGER NOT NULL,  -- 升级到第N级
    cost_xp INTEGER DEFAULT 0,
    cost_gold INTEGER DEFAULT 0,
    cost_materials JSONB DEFAULT '[]',  -- [{"item_code": "xxx", "count": 5}]
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(level_number)
);

COMMENT ON TABLE game_config.skill_upgrade_costs IS '全局技能升级消耗配置（所有技能共用）';
COMMENT ON COLUMN game_config.skill_upgrade_costs.level_number IS '升级到第N级所需消耗（例如：2表示从1级升到2级的消耗）';

-- 2. 修改 skills 表，添加升级规则配置
ALTER TABLE game_config.skills 
ADD COLUMN IF NOT EXISTS level_scaling_type VARCHAR(20) DEFAULT 'linear' 
    CHECK (level_scaling_type IN ('linear', 'fixed', 'percentage')),
ADD COLUMN IF NOT EXISTS level_scaling_config JSONB DEFAULT '{}';

COMMENT ON COLUMN game_config.skills.level_scaling_type IS '升级加成类型: linear(线性增长), fixed(固定不变), percentage(百分比增长)';
COMMENT ON COLUMN game_config.skills.level_scaling_config IS '升级加成配置，JSON格式存储各属性的基础值和增长规则';

-- 3. 删除旧的 skill_level_configs 表（如果需要迁移数据，请先备份）
-- DROP TABLE IF EXISTS game_config.skill_level_configs CASCADE;

-- 4. 插入默认的升级消耗配置（示例数据）
INSERT INTO game_config.skill_upgrade_costs (level_number, cost_xp, cost_gold, cost_materials) 
VALUES 
(2, 100, 50, '[]'),
(3, 200, 100, '[]'),
(4, 400, 200, '[]'),
(5, 800, 400, '[{"item_code": "skill_book_basic", "count": 1}]'),
(6, 1200, 600, '[]'),
(7, 1800, 900, '[]'),
(8, 2500, 1200, '[]'),
(9, 3500, 1600, '[]'),
(10, 5000, 2000, '[{"item_code": "skill_book_advanced", "count": 1}]')
ON CONFLICT (level_number) DO NOTHING;

-- 5. 创建触发器
CREATE TRIGGER update_skill_upgrade_costs_updated_at 
    BEFORE UPDATE ON game_config.skill_upgrade_costs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 6. 创建索引
CREATE INDEX IF NOT EXISTS idx_skill_upgrade_costs_level 
    ON game_config.skill_upgrade_costs(level_number);

-- ============================================
-- 升级规则配置示例
-- ============================================

-- 示例1：火球术 - 线性增长
-- level_scaling_config: {
--   "damage": {"base": 20, "type": "linear", "value": 8},      // 1级20，每级+8
--   "mana_cost": {"base": 10, "type": "linear", "value": 2},   // 1级10法力，每级+2
--   "range": {"base": 6, "type": "fixed", "value": 0}          // 固定6格射程
-- }
-- 结果：1级伤害20，2级28，3级36...

-- 示例2：治疗术 - 百分比增长  
-- level_scaling_config: {
--   "healing": {"base": 30, "type": "percentage", "value": 15}, // 1级30，每级+15%
--   "mana_cost": {"base": 8, "type": "linear", "value": 1}
-- }
-- 结果：1级治疗30，2级34.5(30*1.15)，3级39.7(30*1.15^2)...

-- 示例3：被动技能 - 固定值
-- level_scaling_config: {
--   "bonus_ac": {"base": 2, "type": "fixed", "value": 0},       // 固定+2护甲
--   "bonus_hp": {"base": 5, "type": "linear", "value": 3}       // 每级+3生命
-- }
