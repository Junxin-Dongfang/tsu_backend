-- ============================================
-- 英雄系统配置表（game_config schema）
-- ============================================

-- 1. 属性加点消耗表
CREATE TABLE game_config.attribute_upgrade_costs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    point_number INTEGER NOT NULL UNIQUE,
    cost_xp INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_attribute_upgrade_costs_point ON game_config.attribute_upgrade_costs(point_number);

CREATE TRIGGER update_attribute_upgrade_costs_updated_at 
    BEFORE UPDATE ON game_config.attribute_upgrade_costs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.attribute_upgrade_costs IS '属性加点经验消耗配置（全局共用）';
COMMENT ON COLUMN game_config.attribute_upgrade_costs.point_number IS '第N点属性加点';
COMMENT ON COLUMN game_config.attribute_upgrade_costs.cost_xp IS '该点所需经验值';

-- 插入示例数据（每点100经验，可根据需求调整）
INSERT INTO game_config.attribute_upgrade_costs (point_number, cost_xp) VALUES
(1, 100), (2, 100), (3, 100), (4, 100), (5, 100),
(6, 100), (7, 100), (8, 100), (9, 100), (10, 100),
(11, 100), (12, 100), (13, 100), (14, 100), (15, 100),
(16, 100), (17, 100), (18, 100), (19, 100), (20, 100);
-- 继续插入更多等级...

-- 2. 英雄升级经验需求表
CREATE TABLE game_config.hero_level_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    level INTEGER NOT NULL UNIQUE CHECK (level >= 2 AND level <= 40),
    required_xp INTEGER NOT NULL CHECK (required_xp > 0),
    cumulative_xp INTEGER NOT NULL CHECK (cumulative_xp > 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_hero_level_requirements_level ON game_config.hero_level_requirements(level);

CREATE TRIGGER update_hero_level_requirements_updated_at 
    BEFORE UPDATE ON game_config.hero_level_requirements 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_config.hero_level_requirements IS '英雄升级经验需求配置';
COMMENT ON COLUMN game_config.hero_level_requirements.level IS '目标等级（2-40）';
COMMENT ON COLUMN game_config.hero_level_requirements.required_xp IS '升到该等级需要的增量经验';
COMMENT ON COLUMN game_config.hero_level_requirements.cumulative_xp IS '升到该等级需要的累计总经验';

-- 插入2-40级的经验需求（示例数据，采用递增策略）
INSERT INTO game_config.hero_level_requirements (level, required_xp, cumulative_xp) VALUES
(2, 100, 100),
(3, 150, 250),
(4, 200, 450),
(5, 300, 750),
(6, 400, 1150),
(7, 500, 1650),
(8, 650, 2300),
(9, 800, 3100),
(10, 1000, 4100),
(11, 1200, 5300),
(12, 1400, 6700),
(13, 1600, 8300),
(14, 1850, 10150),
(15, 2100, 12250),
(16, 2400, 14650),
(17, 2700, 17350),
(18, 3000, 20350),
(19, 3350, 23700),
(20, 3700, 27400),
(21, 4100, 31500),
(22, 4500, 36000),
(23, 4950, 40950),
(24, 5400, 46350),
(25, 5900, 52250),
(26, 6450, 58700),
(27, 7000, 65700),
(28, 7600, 73300),
(29, 8250, 81550),
(30, 9000, 90550),
(31, 9800, 100350),
(32, 10650, 111000),
(33, 11550, 122550),
(34, 12500, 135050),
(35, 13550, 148600),
(36, 14700, 163300),
(37, 15950, 179250),
(38, 17300, 196550),
(39, 18800, 215350),
(40, 20500, 235850);

-- 3. 修改职业技能池表，添加初始技能标记
ALTER TABLE game_config.class_skill_pools 
ADD COLUMN IF NOT EXISTS is_initial_skill BOOLEAN DEFAULT FALSE;

COMMENT ON COLUMN game_config.class_skill_pools.is_initial_skill 
IS '是否为职业初始技能（创建英雄或获得职业时自动学习）';

