-- =============================================================================
-- Create Core Infrastructure
-- 创建核心基础设施：扩展、枚举类型、触发器函数等
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 数据库扩展
-- --------------------------------------------------------------------------------

-- UUID 生成扩展 (v7 优先，回退到 v4)
DO $$
BEGIN
    -- 尝试创建 uuid-ossp 扩展，用于 UUID 生成
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- 如果支持，可以考虑使用更现代的 pg_uuidv7 扩展
    -- CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";

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
        CREATE TYPE class_tier_enum AS ENUM ('1', '2', '3', '4', '5');
    END IF;
END $$;

-- 属性数据类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attribute_data_type_enum') THEN
        CREATE TYPE attribute_data_type_enum AS ENUM (
            'integer',     -- 整数类型
            'decimal',     -- 小数类型
            'percentage',  -- 百分比类型
            'boolean'      -- 布尔类型
        );
    END IF;
END $$;

-- 属性分类枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'attribute_category_enum') THEN
        CREATE TYPE attribute_category_enum AS ENUM (
            'basic',      -- 基础属性
            'combat',     -- 战斗属性
            'special',    -- 特殊属性
            'resistance', -- 抗性属性
            'derived'     -- 派生属性
        );
    END IF;
END $$;

-- 技能类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'skill_type_enum') THEN
        CREATE TYPE skill_type_enum AS ENUM (
            'passive',    -- 被动技能
            'active',     -- 主动技能
            'ultimate',   -- 终极技能
            'class'       -- 职业技能
        );
    END IF;
END $$;

-- 动作类型枚举
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'action_type_enum') THEN
        CREATE TYPE action_type_enum AS ENUM (
            'main',       -- 主要动作
            'minor',      -- 次要动作
            'reaction'    -- 反应动作
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

-- 用户财务变更触发器函数
CREATE OR REPLACE FUNCTION handle_user_finance_change()
RETURNS TRIGGER AS $$
BEGIN
    -- 记录财务变更日志
    INSERT INTO financial_transactions (
        user_id,
        transaction_type,
        amount,
        balance_before,
        balance_after,
        description
    ) VALUES (
        COALESCE(NEW.user_id, OLD.user_id),
        CASE
            WHEN TG_OP = 'INSERT' THEN 'system_init'
            WHEN TG_OP = 'UPDATE' THEN 'balance_update'
            ELSE 'unknown'
        END,
        CASE
            WHEN TG_OP = 'INSERT' THEN NEW.diamond_count
            WHEN TG_OP = 'UPDATE' THEN (NEW.diamond_count - OLD.diamond_count)
            ELSE 0
        END,
        CASE
            WHEN TG_OP = 'INSERT' THEN 0
            WHEN TG_OP = 'UPDATE' THEN OLD.diamond_count
            ELSE 0
        END,
        CASE
            WHEN TG_OP = 'INSERT' THEN NEW.diamond_count
            WHEN TG_OP = 'UPDATE' THEN NEW.diamond_count
            ELSE 0
        END,
        'Triggered by ' || TG_OP || ' on user_finances'
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- --------------------------------------------------------------------------------
-- 通用索引模式函数
-- --------------------------------------------------------------------------------

-- 为表创建标准索引的函数
CREATE OR REPLACE FUNCTION create_standard_indexes(table_name TEXT)
RETURNS void AS $$
BEGIN
    -- created_at 索引
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at)', table_name, table_name);

    -- updated_at 索引
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_updated_at ON %s(updated_at)', table_name, table_name);

    -- 软删除索引
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_deleted_at ON %s(deleted_at)', table_name, table_name);

    -- 激活状态索引（如果存在）
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_is_active ON %s(is_active) WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = ''%s'' AND column_name = ''is_active'')', table_name, table_name, table_name);

    RAISE NOTICE 'Standard indexes created for table: %', table_name;
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
-- 版本信息
-- --------------------------------------------------------------------------------

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Core Infrastructure 创建完成';
    RAISE NOTICE '包含: 扩展、枚举类型、触发器函数';
    RAISE NOTICE '============================================';
END $$;