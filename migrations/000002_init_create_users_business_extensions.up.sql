-- =============================================================================
-- 用户财务表
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_finances (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    --消费信息
    --累计消费金额
    total_spent_amount DECIMAL(20, 6) DEFAULT 0.0 CHECK (total_spent_amount >= 0),

    --钻石信息
    --累计获得钻石数
    total_diamonds BIGINT DEFAULT 0 CHECK (total_diamonds >= 0),
    --当前钻石数
    current_diamonds BIGINT DEFAULT 0 CHECK (current_diamonds >= 0),
    --累计消耗钻石数
    total_diamonds_spent BIGINT DEFAULT 0 CHECK (total_diamonds_spent >= 0),

    --高级用户信息
    premium_start      TIMESTAMPTZ, --高级用户开始时间
    premium_expiry     TIMESTAMPTZ, --高级用户到期时间

    --时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- 软删除
);

-- 用户财务表索引
CREATE UNIQUE INDEX idx_user_finances_user_id_unique ON user_finances(user_id) WHERE deleted_at IS NULL;

-- 判断用户是否为高级用户的函数，基于 user_finances 表中的 premium_start 和 premium_expiry 字段
CREATE OR REPLACE FUNCTION is_user_premium(user_uuid UUID) RETURNS BOOLEAN AS $$
DECLARE
    premium_status BOOLEAN;
BEGIN
    SELECT
        CASE
            WHEN premium_start IS NOT NULL AND premium_expiry IS NOT NULL AND NOW() BETWEEN premium_start AND premium_expiry THEN TRUE
            ELSE FALSE
        END
    INTO premium_status
    FROM user_finances
    WHERE user_id = user_uuid AND deleted_at IS NULL;

    --没有找到用户财务记录等情况，返回 FALSE
    RETURN COALESCE(premium_status, FALSE);
END;
$$ LANGUAGE plpgsql;

-- 用户财务表触发器
CREATE TRIGGER update_user_finances_updated_at 
    BEFORE UPDATE ON user_finances 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

--创建交易类型枚举，purchase-购买，diamond_topup-钻石充值，refund-退款，subscription_payment-订阅支付
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_type_enum') THEN
        CREATE TYPE transaction_type_enum AS ENUM ('purchase', 'diamond_topup', 'subscription_payment', 'refund');
    END IF;
END;
$$ LANGUAGE plpgsql;

--交易状态枚举，pending-待处理，completed-已完成，failed-失败，refunded-已退款
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_status_enum') THEN
        CREATE TYPE transaction_status_enum AS ENUM ('pending', 'completed', 'failed', 'refunded');
    END IF;
END;
$$ LANGUAGE plpgsql;

--支付方式枚举，wechat-微信，zhifubao-支付宝，credit_card-信用卡，paypal-贝宝，stripe-Stripe，apple_pay-苹果支付，google_pay-谷歌支付
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_method_enum') THEN
        CREATE TYPE payment_method_enum AS ENUM ('wechat', 'zhifubao', 'credit_card', 'paypal', 'stripe', 'apple_pay', 'google_pay');
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 财务交易记录表
-- =============================================================================
CREATE TABLE IF NOT EXISTS financial_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    --交易信息
    transaction_type transaction_type_enum NOT NULL, -- 交易类型，如 'purchase', 'refund', 'diamond_topup', 'subscription_payment' 等
    amount DECIMAL(20, 6) NOT NULL CHECK (amount >= 0), -- 交易金额
    currency VARCHAR(10) NOT NULL DEFAULT 'USD', -- 货币类型，默认 USD
    diamonds BIGINT DEFAULT 0 CHECK (diamonds >= 0), -- 涉及的钻石数量
    description TEXT, -- 交易描述

    --第三方支付信息
    payment_provider payment_method_enum, -- 支付提供商，如 'stripe', 'paypal'
    payment_params JSONB, -- 交易参数

    --时间戳
    transaction_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- 软删除
);

-- 财务交易记录表索引
CREATE INDEX idx_financial_transactions_user_id ON financial_transactions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_financial_transactions_transaction_time ON financial_transactions(transaction_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_financial_transactions_transaction_type ON financial_transactions(transaction_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_financial_transactions_payment_provider ON financial_transactions(payment_provider) WHERE deleted_at IS NULL;

-- 财务交易记录表触发器
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_financial_transactions_updated_at') THEN
        CREATE TRIGGER update_financial_transactions_updated_at 
        BEFORE UPDATE ON financial_transactions 
        FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
    END IF;
END;
$$ LANGUAGE plpgsql;