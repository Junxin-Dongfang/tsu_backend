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

-- 2. 修改 skill_unlock_actions 表，添加升级规则配置
-- 每个解锁的动作可以有独立的成长曲线
-- 每个属性可以有不同的成长类型（linear/percentage/fixed），所以不需要顶层的 type 字段
ALTER TABLE game_config.skill_unlock_actions 
ADD COLUMN IF NOT EXISTS level_scaling_config JSONB DEFAULT '{}';

COMMENT ON COLUMN game_config.skill_unlock_actions.level_scaling_config IS '动作升级加成配置，JSON格式存储各属性的基础值和增长规则。每个属性有独立的type(linear/percentage/fixed)、base、value配置';

-- 示例配置格式：
-- {
--   "damage": {"type": "linear", "base": 10, "value": 2},
--   "accuracy": {"type": "percentage", "base": 0, "value": 1},
--   "cooldown": {"type": "linear", "base": 3, "value": -0.1}
-- }

-- 3. 删除旧的 skill_level_configs 表
DROP TABLE IF EXISTS game_config.skill_level_configs CASCADE;

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
-- 升级规则配置示例（针对 skill_unlock_actions 表）
-- ============================================

-- 示例1：剑术精通 - 基础攻击动作 - 线性增长
-- level_scaling_config: {
--   "damage": {"base": 10, "type": "linear", "value": 2},      // 1级10，每级+2
--   "accuracy": {"base": 0, "type": "percentage", "value": 1}  // 命中率每级+1%
-- }

-- 示例2：剑术精通 - 强力一击动作 - 线性增长+冷却递减
-- level_scaling_config: {
--   "damage": {"base": 20, "type": "linear", "value": 5},      // 1级20，每级+5
--   "cooldown": {"base": 3, "type": "linear", "value": -0.1}   // 冷却每级-0.1回合
-- }

-- 示例3：火球术动作 - 百分比增长  
-- level_scaling_config: {
--   "damage": {"base": 30, "type": "percentage", "value": 15}, // 1级30，每级+15%
--   "mana_cost": {"base": 10, "type": "linear", "value": 2},   // 1级10法力，每级+2
--   "range": {"base": 6, "type": "fixed", "value": 0}          // 固定6格射程
-- }

-- 示例4：旋风斩动作 - 固定值+范围成长
-- level_scaling_config: {
--   "damage": {"base": 15, "type": "linear", "value": 3},      // 每级+3伤害
--   "range": {"base": 2, "type": "fixed", "value": 0}          // 固定范围2格
-- }
