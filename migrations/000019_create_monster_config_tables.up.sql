-- 000019_create_monster_config_tables.up.sql
-- 创建怪物配置相关表

-- ============================================================================
-- 1. 怪物配置主表 (game_config.monsters)
-- ============================================================================

CREATE TABLE IF NOT EXISTS game_config.monsters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- 基本信息
    monster_code VARCHAR(64) NOT NULL UNIQUE,
    monster_name VARCHAR(128) NOT NULL,
    monster_level SMALLINT NOT NULL CHECK (monster_level >= 1 AND monster_level <= 100),
    description TEXT,
    
    -- HP/MP 配置
    max_hp INTEGER NOT NULL CHECK (max_hp > 0),
    hp_recovery INTEGER DEFAULT 0 CHECK (hp_recovery >= 0),
    max_mp INTEGER DEFAULT 0 CHECK (max_mp >= 0),
    mp_recovery INTEGER DEFAULT 0 CHECK (mp_recovery >= 0),
    
    -- 基础属性（7大属性）
    base_str SMALLINT DEFAULT 0 CHECK (base_str >= 0 AND base_str <= 99),
    base_agi SMALLINT DEFAULT 0 CHECK (base_agi >= 0 AND base_agi <= 99),
    base_vit SMALLINT DEFAULT 0 CHECK (base_vit >= 0 AND base_vit <= 99),
    base_wlp SMALLINT DEFAULT 0 CHECK (base_wlp >= 0 AND base_wlp <= 99),
    base_int SMALLINT DEFAULT 0 CHECK (base_int >= 0 AND base_int <= 99),
    base_wis SMALLINT DEFAULT 0 CHECK (base_wis >= 0 AND base_wis <= 99),
    base_cha SMALLINT DEFAULT 0 CHECK (base_cha >= 0 AND base_cha <= 99),

    -- 战斗属性类型（引用 hero_attribute_type 表的 attribute_code）
    -- 通过引用属性类型，使用对应的计算公式
    accuracy_attribute_code VARCHAR(32) DEFAULT 'ACCURACY' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    dodge_attribute_code VARCHAR(32) DEFAULT 'DODGE' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    initiative_attribute_code VARCHAR(32) DEFAULT 'INITIATIVE' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,

    -- 豁免属性类型（引用 hero_attribute_type 表的 attribute_code）
    body_resist_attribute_code VARCHAR(32) DEFAULT 'BODY_RESIST' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    magic_resist_attribute_code VARCHAR(32) DEFAULT 'MAGIC_RESIST' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    mental_resist_attribute_code VARCHAR(32) DEFAULT 'MENTAL_RESIST' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    environment_resist_attribute_code VARCHAR(32) DEFAULT 'ENVIRONMENT_RESIST' REFERENCES game_config.hero_attribute_type(attribute_code) ON DELETE RESTRICT,
    
    -- 伤害抗性（JSONB格式）
    -- 格式: {"SLASH_RESIST": 0, "SLASH_DR": 0, "FIRE_RESIST": 0.75, ...}
    damage_resistances JSONB DEFAULT '{}'::jsonb,
    
    -- 被动效果（Buff配置）
    -- 格式: [{"buff_id": "xxx", "params": {...}, "caster_level": 5}]
    passive_buffs JSONB DEFAULT '[]'::jsonb,
    
    -- 掉落配置
    drop_gold_min INTEGER DEFAULT 0 CHECK (drop_gold_min >= 0),
    drop_gold_max INTEGER DEFAULT 0 CHECK (drop_gold_max >= 0),
    drop_exp INTEGER DEFAULT 0 CHECK (drop_exp >= 0),
    
    -- 显示配置
    icon_url VARCHAR(512),
    model_url VARCHAR(512),
    
    -- 状态控制
    is_active BOOLEAN DEFAULT TRUE,
    display_order INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- 约束：金币掉落范围合理
    CONSTRAINT check_gold_range CHECK (drop_gold_max >= drop_gold_min)
);

-- 表注释
COMMENT ON TABLE game_config.monsters IS '怪物配置主表';
COMMENT ON COLUMN game_config.monsters.monster_code IS '怪物代码（唯一标识）';
COMMENT ON COLUMN game_config.monsters.monster_name IS '怪物名称';
COMMENT ON COLUMN game_config.monsters.monster_level IS '怪物等级（1-100）';
COMMENT ON COLUMN game_config.monsters.accuracy_attribute_code IS '精准属性类型代码（引用 hero_attribute_type，决定命中率计算公式）';
COMMENT ON COLUMN game_config.monsters.dodge_attribute_code IS '闪避属性类型代码（引用 hero_attribute_type，决定闪避率计算公式）';
COMMENT ON COLUMN game_config.monsters.initiative_attribute_code IS '先攻属性类型代码（引用 hero_attribute_type，决定行动顺序计算公式）';
COMMENT ON COLUMN game_config.monsters.body_resist_attribute_code IS '体质豁免属性类型代码（引用 hero_attribute_type，决定抗毒素等计算公式）';
COMMENT ON COLUMN game_config.monsters.magic_resist_attribute_code IS '魔法豁免属性类型代码（引用 hero_attribute_type，决定抗魔法计算公式）';
COMMENT ON COLUMN game_config.monsters.mental_resist_attribute_code IS '精神豁免属性类型代码（引用 hero_attribute_type，决定抗精神控制计算公式）';
COMMENT ON COLUMN game_config.monsters.environment_resist_attribute_code IS '环境豁免属性类型代码（引用 hero_attribute_type，决定抗环境伤害计算公式）';
COMMENT ON COLUMN game_config.monsters.damage_resistances IS '伤害抗性配置（JSONB格式）';
COMMENT ON COLUMN game_config.monsters.passive_buffs IS '被动效果列表（JSONB格式）';

-- 索引
CREATE INDEX idx_monsters_code ON game_config.monsters(monster_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_monsters_level ON game_config.monsters(monster_level) WHERE deleted_at IS NULL;
CREATE INDEX idx_monsters_active ON game_config.monsters(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_monsters_deleted_at ON game_config.monsters(deleted_at);

-- 触发器：自动更新 updated_at
CREATE TRIGGER update_monsters_updated_at
    BEFORE UPDATE ON game_config.monsters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 2. 怪物技能关联表 (game_config.monster_skills)
-- ============================================================================

CREATE TABLE IF NOT EXISTS game_config.monster_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    monster_id UUID NOT NULL REFERENCES game_config.monsters(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES game_config.skills(id) ON DELETE RESTRICT,
    
    skill_level SMALLINT NOT NULL DEFAULT 1 CHECK (skill_level >= 1 AND skill_level <= 20),
    gain_actions TEXT[],
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    display_order INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- 唯一约束：同一怪物不能重复添加同一技能
    CONSTRAINT uq_monster_skill UNIQUE(monster_id, skill_id)
);

-- 表注释
COMMENT ON TABLE game_config.monster_skills IS '怪物技能关联表';
COMMENT ON COLUMN game_config.monster_skills.monster_id IS '怪物ID（外键）';
COMMENT ON COLUMN game_config.monster_skills.skill_id IS '技能ID（外键）';
COMMENT ON COLUMN game_config.monster_skills.skill_level IS '技能等级（1-20）';
COMMENT ON COLUMN game_config.monster_skills.gain_actions IS '获得的动作列表';

-- 索引
CREATE INDEX idx_monster_skills_monster ON game_config.monster_skills(monster_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monster_skills_skill ON game_config.monster_skills(skill_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monster_skills_deleted_at ON game_config.monster_skills(deleted_at);

-- 触发器：自动更新 updated_at
CREATE TRIGGER update_monster_skills_updated_at
    BEFORE UPDATE ON game_config.monster_skills
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 3. 怪物掉落配置表 (game_config.monster_drops)
-- ============================================================================

CREATE TABLE IF NOT EXISTS game_config.monster_drops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    monster_id UUID NOT NULL REFERENCES game_config.monsters(id) ON DELETE CASCADE,
    drop_pool_id UUID NOT NULL REFERENCES game_config.drop_pools(id) ON DELETE RESTRICT,
    
    -- 掉落配置
    drop_type VARCHAR(32) NOT NULL CHECK (drop_type IN ('team', 'personal')),
    drop_chance DECIMAL(5,4) NOT NULL CHECK (drop_chance > 0 AND drop_chance <= 1),
    min_quantity INTEGER DEFAULT 1 CHECK (min_quantity >= 1),
    max_quantity INTEGER DEFAULT 1 CHECK (max_quantity >= 1),
    
    -- 掉落条件（可选）
    drop_conditions JSONB,

    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    display_order INTEGER DEFAULT 0,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 约束：数量范围合理
    CONSTRAINT check_quantity_range CHECK (max_quantity >= min_quantity),
    -- 唯一约束：同一怪物不能重复添加同一掉落池
    CONSTRAINT uq_monster_drop_pool UNIQUE(monster_id, drop_pool_id)
);

-- 表注释
COMMENT ON TABLE game_config.monster_drops IS '怪物掉落配置表';
COMMENT ON COLUMN game_config.monster_drops.monster_id IS '怪物ID（外键）';
COMMENT ON COLUMN game_config.monster_drops.drop_pool_id IS '掉落池ID（外键）';
COMMENT ON COLUMN game_config.monster_drops.drop_type IS '掉落类型（team/personal）';
COMMENT ON COLUMN game_config.monster_drops.drop_chance IS '掉落概率（0-1）';
COMMENT ON COLUMN game_config.monster_drops.drop_conditions IS '掉落条件配置（JSONB格式）';

-- 索引
CREATE INDEX idx_monster_drops_monster ON game_config.monster_drops(monster_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monster_drops_pool ON game_config.monster_drops(drop_pool_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monster_drops_deleted_at ON game_config.monster_drops(deleted_at);

-- 触发器：自动更新 updated_at
CREATE TRIGGER update_monster_drops_updated_at
    BEFORE UPDATE ON game_config.monster_drops
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
