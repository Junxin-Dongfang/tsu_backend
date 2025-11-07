-- =============================================================================
-- Create Dungeon System Tables
-- 创建地城系统表
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 配置表 (game_config schema)
-- --------------------------------------------------------------------------------

-- 1. 地城配置表
CREATE TABLE IF NOT EXISTS game_config.dungeons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dungeon_code VARCHAR(50) UNIQUE NOT NULL,
    dungeon_name VARCHAR(100) NOT NULL,
    
    -- 等级限制
    min_level SMALLINT NOT NULL,
    max_level SMALLINT NOT NULL,
    
    -- 描述
    description TEXT,
    
    -- 限时设置
    is_time_limited BOOLEAN NOT NULL DEFAULT FALSE,
    time_limit_start TIMESTAMPTZ,
    time_limit_end TIMESTAMPTZ,
    
    -- 挑战次数限制
    requires_attempts BOOLEAN NOT NULL DEFAULT FALSE,
    max_attempts_per_day SMALLINT,
    
    -- 房间序列配置 (JSONB)
    room_sequence JSONB NOT NULL DEFAULT '[]',
    
    -- 状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT check_level_range CHECK (min_level <= max_level),
    CONSTRAINT check_time_limit CHECK (
        NOT is_time_limited OR 
        (time_limit_start IS NOT NULL AND time_limit_end IS NOT NULL)
    ),
    CONSTRAINT check_attempts CHECK (
        NOT requires_attempts OR 
        (max_attempts_per_day IS NOT NULL AND max_attempts_per_day > 0)
    )
);

-- 地城表索引
CREATE INDEX IF NOT EXISTS idx_dungeons_code ON game_config.dungeons(dungeon_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeons_level ON game_config.dungeons(min_level, max_level) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeons_active ON game_config.dungeons(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeons_created_at ON game_config.dungeons(created_at) WHERE deleted_at IS NULL;

-- 地城表触发器
CREATE TRIGGER update_dungeons_updated_at
    BEFORE UPDATE ON game_config.dungeons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 地城表注释
COMMENT ON TABLE game_config.dungeons IS '地城配置表 - 存储地城（副本）的基础配置信息';
COMMENT ON COLUMN game_config.dungeons.dungeon_code IS '地城代码 - 全局唯一标识符';
COMMENT ON COLUMN game_config.dungeons.dungeon_name IS '地城名称 - 显示给玩家的名称';
COMMENT ON COLUMN game_config.dungeons.min_level IS '最小等级 - 进入地城的最低等级要求';
COMMENT ON COLUMN game_config.dungeons.max_level IS '最大等级 - 进入地城的最高等级要求';
COMMENT ON COLUMN game_config.dungeons.description IS '地城描述 - 地城的背景故事和说明';
COMMENT ON COLUMN game_config.dungeons.is_time_limited IS '是否限时 - 地城是否只在特定时间段开放';
COMMENT ON COLUMN game_config.dungeons.time_limit_start IS '限时开始时间 - 地城开放的开始时间';
COMMENT ON COLUMN game_config.dungeons.time_limit_end IS '限时结束时间 - 地城开放的结束时间';
COMMENT ON COLUMN game_config.dungeons.requires_attempts IS '是否需要挑战次数 - 是否限制每日挑战次数';
COMMENT ON COLUMN game_config.dungeons.max_attempts_per_day IS '每日最大挑战次数 - 每天可以挑战的最大次数';
COMMENT ON COLUMN game_config.dungeons.room_sequence IS '房间序列 - JSONB格式,定义房间的顺序和条件分支';
COMMENT ON COLUMN game_config.dungeons.is_active IS '是否启用 - 地城是否对玩家开放';

-- 2. 房间配置表
CREATE TABLE IF NOT EXISTS game_config.dungeon_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_code VARCHAR(50) UNIQUE NOT NULL,
    room_name VARCHAR(100),
    
    -- 房间类型
    room_type VARCHAR(20) NOT NULL,
    
    -- 触发配置
    trigger_id VARCHAR(50),
    
    -- 开启条件 (JSONB)
    open_conditions JSONB NOT NULL DEFAULT '{}',
    
    -- 状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT check_room_type CHECK (room_type IN ('battle', 'event', 'treasure', 'rest'))
);

-- 房间表索引
CREATE INDEX IF NOT EXISTS idx_dungeon_rooms_code ON game_config.dungeon_rooms(room_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeon_rooms_type ON game_config.dungeon_rooms(room_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeon_rooms_active ON game_config.dungeon_rooms(is_active) WHERE deleted_at IS NULL;

-- 房间表触发器
CREATE TRIGGER update_dungeon_rooms_updated_at
    BEFORE UPDATE ON game_config.dungeon_rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 房间表注释
COMMENT ON TABLE game_config.dungeon_rooms IS '地城房间配置表 - 存储地城中各个房间的配置';
COMMENT ON COLUMN game_config.dungeon_rooms.room_code IS '房间代码 - 全局唯一标识符';
COMMENT ON COLUMN game_config.dungeon_rooms.room_name IS '房间名称 - 显示给玩家的房间名称';
COMMENT ON COLUMN game_config.dungeon_rooms.room_type IS '房间类型 - battle(战斗)/event(事件)/treasure(宝箱)/rest(休息)';
COMMENT ON COLUMN game_config.dungeon_rooms.trigger_id IS '触发ID - 战斗ID或事件ID';
COMMENT ON COLUMN game_config.dungeon_rooms.open_conditions IS '开启条件 - JSONB格式,定义进入房间的条件';

-- 3. 战斗配置表
CREATE TABLE IF NOT EXISTS game_config.dungeon_battles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_code VARCHAR(50) UNIQUE NOT NULL,
    
    -- 场地设置 (JSONB)
    location_config JSONB NOT NULL DEFAULT '{}',
    
    -- 全程状态 (JSONB)
    global_buffs JSONB NOT NULL DEFAULT '[]',
    
    -- 怪物配置 (JSONB)
    monster_setup JSONB NOT NULL DEFAULT '[]',
    
    -- 描述文本
    battle_start_desc TEXT,
    battle_success_desc TEXT,
    battle_failure_desc TEXT,
    
    -- 状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 战斗表索引
CREATE INDEX IF NOT EXISTS idx_dungeon_battles_code ON game_config.dungeon_battles(battle_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeon_battles_active ON game_config.dungeon_battles(is_active) WHERE deleted_at IS NULL;

-- 战斗表触发器
CREATE TRIGGER update_dungeon_battles_updated_at
    BEFORE UPDATE ON game_config.dungeon_battles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 战斗表注释
COMMENT ON TABLE game_config.dungeon_battles IS '地城战斗配置表 - 存储地城战斗场景的配置';
COMMENT ON COLUMN game_config.dungeon_battles.battle_code IS '战斗代码 - 全局唯一标识符';
COMMENT ON COLUMN game_config.dungeon_battles.location_config IS '场地配置 - JSONB格式,定义战斗场地和场地事件';
COMMENT ON COLUMN game_config.dungeon_battles.global_buffs IS '全程状态 - JSONB格式,定义战斗全程的Buff/Debuff';
COMMENT ON COLUMN game_config.dungeon_battles.monster_setup IS '怪物配置 - JSONB格式,定义怪物种类、位置、等级';
COMMENT ON COLUMN game_config.dungeon_battles.battle_start_desc IS '战斗开始描述 - 战斗开始时显示的文本';
COMMENT ON COLUMN game_config.dungeon_battles.battle_success_desc IS '战斗成功描述 - 战斗胜利时显示的文本';
COMMENT ON COLUMN game_config.dungeon_battles.battle_failure_desc IS '战斗失败描述 - 战斗失败时显示的文本';

-- 4. 事件配置表
CREATE TABLE IF NOT EXISTS game_config.dungeon_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_code VARCHAR(50) UNIQUE NOT NULL,

    -- 事件描述
    event_description TEXT,

    -- 施加效果 (JSONB)
    apply_effects JSONB NOT NULL DEFAULT '[]',

    -- 掉落配置 (JSONB)
    drop_config JSONB NOT NULL DEFAULT '{}',

    -- 经验奖励
    reward_exp INT DEFAULT 0,

    -- 事件结束描述
    event_end_desc TEXT,

    -- 状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT check_reward_exp CHECK (reward_exp >= 0)
);

-- 事件表索引
CREATE INDEX IF NOT EXISTS idx_dungeon_events_code ON game_config.dungeon_events(event_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dungeon_events_active ON game_config.dungeon_events(is_active) WHERE deleted_at IS NULL;

-- 事件表触发器
CREATE TRIGGER update_dungeon_events_updated_at
    BEFORE UPDATE ON game_config.dungeon_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 事件表注释
COMMENT ON TABLE game_config.dungeon_events IS '地城事件配置表 - 存储地城事件的配置';
COMMENT ON COLUMN game_config.dungeon_events.event_code IS '事件代码 - 全局唯一标识符';
COMMENT ON COLUMN game_config.dungeon_events.event_description IS '事件描述 - 事件发生时显示的文本';
COMMENT ON COLUMN game_config.dungeon_events.apply_effects IS '施加效果 - JSONB格式,定义事件施加的Buff/Debuff';
COMMENT ON COLUMN game_config.dungeon_events.drop_config IS '掉落配置 - JSONB格式,定义掉落池和保底物品';
COMMENT ON COLUMN game_config.dungeon_events.reward_exp IS '经验奖励 - 事件完成后获得的经验值';
COMMENT ON COLUMN game_config.dungeon_events.event_end_desc IS '事件结束描述 - 事件结束时显示的文本';

-- --------------------------------------------------------------------------------
-- 运行时表 (game_runtime schema)
-- --------------------------------------------------------------------------------

-- 5. 团队地城挑战记录表
CREATE TABLE IF NOT EXISTS game_runtime.team_dungeon_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL,
    dungeon_id UUID NOT NULL REFERENCES game_config.dungeons(id) ON DELETE RESTRICT,

    -- 挑战次数
    attempts_count INT NOT NULL DEFAULT 0,
    max_attempts INT,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_team_dungeon UNIQUE (team_id, dungeon_id)
);

-- 团队地城记录表索引
CREATE INDEX IF NOT EXISTS idx_team_dungeon_records_team_id ON game_runtime.team_dungeon_records(team_id);
CREATE INDEX IF NOT EXISTS idx_team_dungeon_records_dungeon_id ON game_runtime.team_dungeon_records(dungeon_id);
CREATE INDEX IF NOT EXISTS idx_team_dungeon_records_created_at ON game_runtime.team_dungeon_records(created_at);

-- 团队地城记录表触发器
CREATE TRIGGER update_team_dungeon_records_updated_at
    BEFORE UPDATE ON game_runtime.team_dungeon_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 团队地城记录表注释
COMMENT ON TABLE game_runtime.team_dungeon_records IS '团队地城挑战记录表 - 记录团队对地城的挑战次数';
COMMENT ON COLUMN game_runtime.team_dungeon_records.team_id IS '团队ID - 挑战地城的团队';
COMMENT ON COLUMN game_runtime.team_dungeon_records.dungeon_id IS '地城ID - 被挑战的地城';
COMMENT ON COLUMN game_runtime.team_dungeon_records.attempts_count IS '已挑战次数 - 团队已经挑战该地城的次数';
COMMENT ON COLUMN game_runtime.team_dungeon_records.max_attempts IS '最大挑战次数 - 团队可以挑战该地城的最大次数';

-- 6. 团队地城进度表
CREATE TABLE IF NOT EXISTS game_runtime.team_dungeon_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL,
    dungeon_id UUID NOT NULL REFERENCES game_config.dungeons(id) ON DELETE RESTRICT,

    -- 进度信息
    current_room_id VARCHAR(50),
    completed_rooms JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',

    -- 时间戳
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_status CHECK (status IN ('in_progress', 'completed', 'failed', 'abandoned'))
);

-- 团队地城进度表索引
CREATE INDEX IF NOT EXISTS idx_team_dungeon_progress_team_id ON game_runtime.team_dungeon_progress(team_id);
CREATE INDEX IF NOT EXISTS idx_team_dungeon_progress_dungeon_id ON game_runtime.team_dungeon_progress(dungeon_id);
CREATE INDEX IF NOT EXISTS idx_team_dungeon_progress_status ON game_runtime.team_dungeon_progress(status);
CREATE INDEX IF NOT EXISTS idx_team_dungeon_progress_started_at ON game_runtime.team_dungeon_progress(started_at);

-- 团队地城进度表触发器
CREATE TRIGGER update_team_dungeon_progress_updated_at
    BEFORE UPDATE ON game_runtime.team_dungeon_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 团队地城进度表注释
COMMENT ON TABLE game_runtime.team_dungeon_progress IS '团队地城进度表 - 记录团队在地城中的实时进度';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.team_id IS '团队ID - 挑战地城的团队';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.dungeon_id IS '地城ID - 被挑战的地城';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.current_room_id IS '当前房间ID - 团队当前所在的房间';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.completed_rooms IS '已完成房间 - JSONB格式,记录已完成的房间列表';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.status IS '状态 - in_progress(进行中)/completed(已完成)/failed(失败)/abandoned(放弃)';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.started_at IS '开始时间 - 团队开始挑战的时间';
COMMENT ON COLUMN game_runtime.team_dungeon_progress.completed_at IS '完成时间 - 团队完成挑战的时间';

