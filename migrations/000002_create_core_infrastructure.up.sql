-- =============================================================================
-- Create Core Infrastructure
-- 创建核心基础设施：扩展、枚举类型、触发器函数等
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 数据库扩展
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    -- 尝试创建 uuid-ossp 扩展，用于 UUID 生成
    -- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- 使用 pg_uuidv7 扩展
    CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";

EXCEPTION WHEN OTHERS THEN
    -- 记录错误但继续执行
    RAISE NOTICE 'Extension creation skipped: %', SQLERRM;
END $$;

-- --------------------------------------------------------------------------------
-- 枚举类型定义
-- --------------------------------------------------------------------------------

-- 职业阶级枚举：1-基础，2-进阶，3-精英，4-传奇，5-神话
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'class_tier_enum') THEN
        CREATE TYPE class_tier_enum AS ENUM (
            'basic',     -- 基础
            'advanced',  -- 进阶
            'elite',     -- 精英
            'legendary', -- 传奇
            'mythic'     -- 神话
        );
    END IF;
END $$;

-- 属性数据类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_type_enum') THEN
        CREATE TYPE data_type_enum AS ENUM (
            'integer',     -- 整数类型
            'decimal',     -- 小数类型
            'percentage',  -- 百分比类型
            'boolean'      -- 布尔类型
        );
    END IF;
END $$;

-- 技能类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'skill_type_enum') THEN
        CREATE TYPE skill_type_enum AS ENUM (
            'weapon',    -- 武器技能
            'magic',     -- 魔法技能
            'physical',  -- 物理技能
            'usage',     -- 使用物品技能
            'reaction',   -- 反应技能
            'guard'       -- 防御技能
        );
    END IF;
END $$;

-- 标签类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tag_type_enum') THEN
        CREATE TYPE tag_type_enum AS ENUM (
            'class',    -- 职业标签
            'item',     -- 物品标签
            'skill',    -- 技能标签
            'monster'   -- 怪物标签
        );
    END IF;
END $$;

-- --------------------------------------------------------------------------------
-- 触发器函数定义
-- --------------------------------------------------------------------------------

-- 自动更新 updated_at 字段的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- --------------------------------------------------------------------------------
-- 数据完整性约束函数
-- --------------------------------------------------------------------------------

-- 检查枚举值有效性的函数
CREATE OR REPLACE FUNCTION validate_enum_value(enum_type TEXT, value TEXT)
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM pg_enum
        WHERE enumtypid = enum_type::regtype::oid
        AND enumlabel = value
    );
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 完成消息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Core Infrastructure 创建完成';
    RAISE NOTICE '包含: 扩展、枚举类型、触发器函数';
    RAISE NOTICE '============================================';
END $$;
