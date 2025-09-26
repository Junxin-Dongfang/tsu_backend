-- =============================================================================
-- Create Heroes System
-- 英雄系统：英雄实例、属性值、装备等
-- 依赖：000002_create_users_system, 000003_create_attribute_system, 000004_create_classes_system
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 英雄状态枚举
-- --------------------------------------------------------------------------------

-- 英雄状态枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'hero_status_enum') THEN
        CREATE TYPE hero_status_enum AS ENUM (
            'active',     -- 激活状态
            'resting',    -- 休息状态
            'injured',    -- 受伤状态
            'training',   -- 训练状态
            'retired'     -- 退役状态
        );
    END IF;
END $$;

-- 装备槽位枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'equipment_slot_enum') THEN
        CREATE TYPE equipment_slot_enum AS ENUM (
            'weapon',     -- 武器
            'armor',      -- 护甲
            'helmet',     -- 头盔
            'boots',      -- 靴子
            'gloves',     -- 手套
            'ring',       -- 戒指
            'necklace',   -- 项链
            'accessory'   -- 饰品
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 英雄表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS heroes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- 关联信息
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_id          UUID NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,

    -- 英雄基本信息
    hero_name         VARCHAR(64) NOT NULL,                -- 英雄名称
    hero_code         VARCHAR(32) UNIQUE,                  -- 英雄代码（可选，用于特殊英雄）
    description       TEXT,                                -- 英雄描述

    -- 英雄等级和经验
    level             INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience        BIGINT NOT NULL DEFAULT 0 CHECK (experience >= 0),
    experience_to_next BIGINT NOT NULL DEFAULT 100 CHECK (experience_to_next > 0),

    -- 英雄状态
    status            hero_status_enum NOT NULL DEFAULT 'active',
    health_points     INTEGER NOT NULL DEFAULT 100 CHECK (health_points >= 0),
    max_health_points INTEGER NOT NULL DEFAULT 100 CHECK (max_health_points > 0),
    mana_points       INTEGER NOT NULL DEFAULT 50 CHECK (mana_points >= 0),
    max_mana_points   INTEGER NOT NULL DEFAULT 50 CHECK (max_mana_points >= 0),

    -- 英雄外观
    avatar_url        VARCHAR(500),                        -- 头像URL
    skin_id           UUID,                                -- 皮肤ID (可引用皮肤表)

    -- 位置信息（用于战斗系统）
    current_x         INTEGER DEFAULT 0,                   -- 当前X坐标
    current_y         INTEGER DEFAULT 0,                   -- 当前Y坐标
    current_map_id    UUID,                                -- 当前地图ID

    -- 统计信息
    total_battles     INTEGER DEFAULT 0 CHECK (total_battles >= 0),
    victories         INTEGER DEFAULT 0 CHECK (victories >= 0),
    defeats           INTEGER DEFAULT 0 CHECK (defeats >= 0),
    total_damage_dealt BIGINT DEFAULT 0 CHECK (total_damage_dealt >= 0),
    total_damage_taken BIGINT DEFAULT 0 CHECK (total_damage_taken >= 0),

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_battle_at    TIMESTAMPTZ,                         -- 上次战斗时间
    deleted_at        TIMESTAMPTZ                          -- 软删除

    -- 约束：确保胜利数不超过总战斗数
    CONSTRAINT chk_victories_battles CHECK (victories <= total_battles),
    CONSTRAINT chk_defeats_battles CHECK (defeats <= total_battles)
);

-- 英雄表索引
CREATE INDEX IF NOT EXISTS idx_heroes_user_id ON heroes(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_class_id ON heroes(class_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_level ON heroes(level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_status ON heroes(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_heroes_created_at ON heroes(created_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_heroes_hero_code ON heroes(hero_code) WHERE hero_code IS NOT NULL AND deleted_at IS NULL;

-- 英雄表触发器
CREATE TRIGGER update_heroes_updated_at
    BEFORE UPDATE ON heroes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 英雄属性值表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_attributes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    hero_id           UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    attribute_type_id UUID NOT NULL REFERENCES hero_attribute_type(id) ON DELETE CASCADE,

    -- 属性值详细信息
    base_value        DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 基础值
    bonus_value       DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 加成值
    equipment_bonus   DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 装备加成
    temporary_bonus   DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 临时加成（buff等）

    -- 计算值（冗余存储以提高性能）
    final_value       DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 最终值 = base + bonus + equipment + temporary

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个英雄的每个属性类型只能有一条记录
    UNIQUE(hero_id, attribute_type_id)
);

-- 英雄属性值索引
CREATE INDEX IF NOT EXISTS idx_hero_attributes_hero_id ON hero_attributes(hero_id);
CREATE INDEX IF NOT EXISTS idx_hero_attributes_attribute_type_id ON hero_attributes(attribute_type_id);
CREATE INDEX IF NOT EXISTS idx_hero_attributes_final_value ON hero_attributes(final_value);

-- 英雄属性值表触发器
CREATE TRIGGER update_hero_attributes_updated_at
    BEFORE UPDATE ON hero_attributes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 自动计算最终属性值的触发器
CREATE OR REPLACE FUNCTION calculate_hero_final_attribute()
RETURNS TRIGGER AS $$
BEGIN
    NEW.final_value = NEW.base_value + NEW.bonus_value + NEW.equipment_bonus + NEW.temporary_bonus;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_hero_final_attribute_trigger
    BEFORE INSERT OR UPDATE ON hero_attributes
    FOR EACH ROW EXECUTE FUNCTION calculate_hero_final_attribute();

-- --------------------------------------------------------------------------------
-- 英雄装备表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_equipment (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    hero_id           UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    equipment_slot    equipment_slot_enum NOT NULL,       -- 装备槽位

    -- 装备信息 (暂时简化，实际可能引用装备表)
    equipment_id      UUID,                                -- 装备ID
    equipment_name    VARCHAR(100),                        -- 装备名称
    equipment_level   INTEGER DEFAULT 1 CHECK (equipment_level >= 1),
    equipment_quality INTEGER DEFAULT 1 CHECK (equipment_quality >= 1 AND equipment_quality <= 5),

    -- 装备属性加成 (JSON格式存储)
    attribute_bonuses JSONB,                               -- 属性加成，格式: {"attribute_code": bonus_value, ...}

    -- 时间戳
    equipped_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个英雄的每个装备槽位只能装备一件装备
    UNIQUE(hero_id, equipment_slot)
);

-- 英雄装备索引
CREATE INDEX IF NOT EXISTS idx_hero_equipment_hero_id ON hero_equipment(hero_id);
CREATE INDEX IF NOT EXISTS idx_hero_equipment_slot ON hero_equipment(equipment_slot);
CREATE INDEX IF NOT EXISTS idx_hero_equipment_equipped_at ON hero_equipment(equipped_at);

-- 英雄装备表触发器
CREATE TRIGGER update_hero_equipment_updated_at
    BEFORE UPDATE ON hero_equipment
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 英雄技能表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_skills (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    hero_id           UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    skill_id          UUID NOT NULL,                       -- 技能ID (将来引用技能表)
    skill_code        VARCHAR(32) NOT NULL,                -- 技能代码

    -- 技能等级信息
    skill_level       INTEGER NOT NULL DEFAULT 1 CHECK (skill_level >= 1),
    skill_experience  INTEGER NOT NULL DEFAULT 0 CHECK (skill_experience >= 0),
    max_level         INTEGER NOT NULL DEFAULT 10 CHECK (max_level >= 1),

    -- 技能状态
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否激活
    is_equipped       BOOLEAN NOT NULL DEFAULT FALSE,      -- 是否装备（用于战斗）

    -- 获得信息
    learned_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- 学习时间
    learned_method    VARCHAR(50) DEFAULT 'class_unlock',  -- 学习方式

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- 唯一约束：每个英雄的每个技能只能有一条记录
    UNIQUE(hero_id, skill_id),
    UNIQUE(hero_id, skill_code)
);

-- 英雄技能索引
CREATE INDEX IF NOT EXISTS idx_hero_skills_hero_id ON hero_skills(hero_id);
CREATE INDEX IF NOT EXISTS idx_hero_skills_skill_id ON hero_skills(skill_id);
CREATE INDEX IF NOT EXISTS idx_hero_skills_skill_code ON hero_skills(skill_code);
CREATE INDEX IF NOT EXISTS idx_hero_skills_is_equipped ON hero_skills(is_equipped) WHERE is_equipped = TRUE;

-- 英雄技能表触发器
CREATE TRIGGER update_hero_skills_updated_at
    BEFORE UPDATE ON hero_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- --------------------------------------------------------------------------------
-- 英雄经验日志表
-- --------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hero_experience_logs (
    id                BIGSERIAL PRIMARY KEY,

    hero_id           UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,

    -- 经验变化信息
    experience_gained INTEGER NOT NULL,                    -- 获得经验值
    experience_source VARCHAR(50) NOT NULL,               -- 经验来源
    source_id         UUID,                                -- 来源ID（如战斗ID、任务ID等）

    -- 等级变化
    level_before      INTEGER NOT NULL,                    -- 升级前等级
    level_after       INTEGER NOT NULL,                    -- 升级后等级

    -- 描述信息
    description       TEXT,                                -- 详细描述

    -- 时间戳
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 英雄经验日志索引
CREATE INDEX IF NOT EXISTS idx_hero_experience_logs_hero_id ON hero_experience_logs(hero_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_hero_experience_logs_source ON hero_experience_logs(experience_source);

-- --------------------------------------------------------------------------------
-- 英雄相关函数
-- --------------------------------------------------------------------------------

-- 初始化英雄属性的函数
CREATE OR REPLACE FUNCTION initialize_hero_attributes(hero_id_param UUID, class_id_param UUID)
RETURNS void AS $$
DECLARE
    attr_record RECORD;
    class_bonus_record RECORD;
    base_val DECIMAL(10,2);
    bonus_val DECIMAL(10,2);
BEGIN
    -- 为英雄初始化所有激活的属性类型
    FOR attr_record IN
        SELECT id, attribute_code, default_value
        FROM hero_attribute_type
        WHERE is_active = TRUE AND deleted_at IS NULL
    LOOP
        -- 获取基础值（使用属性类型的默认值）
        base_val := COALESCE(attr_record.default_value, 0);

        -- 获取职业加成
        SELECT COALESCE(base_bonus_value, 0)
        INTO bonus_val
        FROM class_attribute_bonuses
        WHERE class_id = class_id_param AND attribute_id = attr_record.id;

        bonus_val := COALESCE(bonus_val, 0);

        -- 插入英雄属性记录
        INSERT INTO hero_attributes (
            hero_id, attribute_type_id, base_value, bonus_value, equipment_bonus, temporary_bonus
        ) VALUES (
            hero_id_param, attr_record.id, base_val, bonus_val, 0, 0
        ) ON CONFLICT (hero_id, attribute_type_id) DO NOTHING;
    END LOOP;

    RAISE NOTICE 'Hero attributes initialized for hero %', hero_id_param;
END;
$$ LANGUAGE plpgsql;

-- 计算英雄战力的函数
CREATE OR REPLACE FUNCTION calculate_hero_power(hero_id_param UUID)
RETURNS INTEGER AS $$
DECLARE
    total_power INTEGER := 0;
    attr_record RECORD;
BEGIN
    -- 基于所有属性计算战力（简化算法）
    FOR attr_record IN
        SELECT ha.final_value, hat.attribute_code
        FROM hero_attributes ha
        JOIN hero_attribute_type hat ON ha.attribute_type_id = hat.id
        WHERE ha.hero_id = hero_id_param AND hat.is_active = TRUE
    LOOP
        -- 不同属性有不同的战力权重
        CASE attr_record.attribute_code
            WHEN 'STRENGTH', 'AGILITY', 'INTELLIGENCE', 'CONSTITUTION' THEN
                total_power := total_power + (attr_record.final_value * 10)::INTEGER;
            WHEN 'ATTACK_POWER', 'MAGIC_POWER' THEN
                total_power := total_power + (attr_record.final_value * 5)::INTEGER;
            WHEN 'DEFENSE', 'MAGIC_RESISTANCE' THEN
                total_power := total_power + (attr_record.final_value * 3)::INTEGER;
            ELSE
                total_power := total_power + (attr_record.final_value * 1)::INTEGER;
        END CASE;
    END LOOP;

    RETURN total_power;
END;
$$ LANGUAGE plpgsql;

-- 获取英雄详细信息的函数
CREATE OR REPLACE FUNCTION get_hero_details(hero_id_param UUID)
RETURNS TABLE (
    hero_id UUID,
    hero_name VARCHAR,
    class_name VARCHAR,
    level INTEGER,
    status hero_status_enum,
    total_power INTEGER,
    attribute_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        h.id,
        h.hero_name,
        c.class_name,
        h.level,
        h.status,
        calculate_hero_power(h.id),
        COUNT(ha.id)
    FROM heroes h
    JOIN classes c ON h.class_id = c.id
    LEFT JOIN hero_attributes ha ON h.id = ha.hero_id
    WHERE h.id = hero_id_param AND h.deleted_at IS NULL
    GROUP BY h.id, h.hero_name, c.class_name, h.level, h.status;
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 创建英雄后自动初始化属性的触发器
-- --------------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION trigger_initialize_hero_attributes()
RETURNS TRIGGER AS $$
BEGIN
    -- 异步初始化属性（避免在事务中执行复杂操作）
    PERFORM initialize_hero_attributes(NEW.id, NEW.class_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER initialize_hero_attributes_trigger
    AFTER INSERT ON heroes
    FOR EACH ROW EXECUTE FUNCTION trigger_initialize_hero_attributes();

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Heroes System 创建完成';
    RAISE NOTICE '包含: 英雄实例、属性值、装备、技能、经验日志';
    RAISE NOTICE '============================================';
END $$;