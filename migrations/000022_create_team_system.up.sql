-- =============================================================================
-- Create Team System Tables
-- 创建团队系统表
-- =============================================================================

-- --------------------------------------------------------------------------------
-- 运行时表 (game_runtime schema)
-- --------------------------------------------------------------------------------

-- 1. 团队表
CREATE TABLE IF NOT EXISTS game_runtime.teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    max_members INT NOT NULL DEFAULT 12,
    leader_hero_id UUID NOT NULL,  -- 当前队长的英雄ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT check_max_members CHECK (max_members > 0 AND max_members <= 12)
);

COMMENT ON TABLE game_runtime.teams IS '团队表';
COMMENT ON COLUMN game_runtime.teams.id IS '团队ID';
COMMENT ON COLUMN game_runtime.teams.name IS '团队名称';
COMMENT ON COLUMN game_runtime.teams.description IS '团队描述';
COMMENT ON COLUMN game_runtime.teams.max_members IS '最大成员数';
COMMENT ON COLUMN game_runtime.teams.leader_hero_id IS '当前队长的英雄ID';
COMMENT ON COLUMN game_runtime.teams.created_at IS '创建时间';
COMMENT ON COLUMN game_runtime.teams.updated_at IS '更新时间';
COMMENT ON COLUMN game_runtime.teams.deleted_at IS '软删除时间';

-- 团队表索引
CREATE INDEX IF NOT EXISTS idx_teams_leader ON game_runtime.teams(leader_hero_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_created_at ON game_runtime.teams(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_name ON game_runtime.teams(name) WHERE deleted_at IS NULL;

-- 团队表触发器
CREATE TRIGGER update_teams_updated_at
    BEFORE UPDATE ON game_runtime.teams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 2. 团队成员表
CREATE TABLE IF NOT EXISTS game_runtime.team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,  -- 冗余字段，方便查询
    role VARCHAR(20) NOT NULL DEFAULT 'member',  -- 'leader', 'admin', 'member'
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_role CHECK (role IN ('leader', 'admin', 'member')),
    CONSTRAINT unique_team_hero UNIQUE (team_id, hero_id)
);

COMMENT ON TABLE game_runtime.team_members IS '团队成员表';
COMMENT ON COLUMN game_runtime.team_members.id IS '成员记录ID';
COMMENT ON COLUMN game_runtime.team_members.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_members.hero_id IS '英雄ID';
COMMENT ON COLUMN game_runtime.team_members.user_id IS '用户ID（冗余）';
COMMENT ON COLUMN game_runtime.team_members.role IS '角色：leader-队长, admin-管理员, member-成员';
COMMENT ON COLUMN game_runtime.team_members.joined_at IS '加入时间';
COMMENT ON COLUMN game_runtime.team_members.last_active_at IS '最后活跃时间';

-- 团队成员表索引
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON game_runtime.team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_hero_id ON game_runtime.team_members(hero_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON game_runtime.team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON game_runtime.team_members(team_id, role);
CREATE INDEX IF NOT EXISTS idx_team_members_last_active ON game_runtime.team_members(last_active_at);
CREATE INDEX IF NOT EXISTS idx_team_members_joined_at ON game_runtime.team_members(joined_at);

-- 3. 团队加入申请表
CREATE TABLE IF NOT EXISTS game_runtime.team_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    message TEXT,  -- 申请留言
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'approved', 'rejected', 'cancelled'
    reviewed_by_hero_id UUID REFERENCES game_runtime.heroes(id) ON DELETE SET NULL,  -- 审批人英雄ID
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_status CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled'))
);

COMMENT ON TABLE game_runtime.team_join_requests IS '团队加入申请表';
COMMENT ON COLUMN game_runtime.team_join_requests.id IS '申请ID';
COMMENT ON COLUMN game_runtime.team_join_requests.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_join_requests.hero_id IS '申请人英雄ID';
COMMENT ON COLUMN game_runtime.team_join_requests.user_id IS '申请人用户ID';
COMMENT ON COLUMN game_runtime.team_join_requests.message IS '申请留言';
COMMENT ON COLUMN game_runtime.team_join_requests.status IS '状态：pending-待审批, approved-已批准, rejected-已拒绝, cancelled-已取消';
COMMENT ON COLUMN game_runtime.team_join_requests.reviewed_by_hero_id IS '审批人英雄ID';
COMMENT ON COLUMN game_runtime.team_join_requests.reviewed_at IS '审批时间';
COMMENT ON COLUMN game_runtime.team_join_requests.created_at IS '创建时间';

-- 团队加入申请表索引
CREATE INDEX IF NOT EXISTS idx_join_requests_team_id ON game_runtime.team_join_requests(team_id, status);
CREATE INDEX IF NOT EXISTS idx_join_requests_hero_id ON game_runtime.team_join_requests(hero_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_created_at ON game_runtime.team_join_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_join_requests_status ON game_runtime.team_join_requests(status) WHERE status = 'pending';

-- 4. 团队邀请表
CREATE TABLE IF NOT EXISTS game_runtime.team_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    inviter_hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    invitee_hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    message TEXT,  -- 邀请留言
    status VARCHAR(20) NOT NULL DEFAULT 'pending_approval',
    -- 'pending_approval' (等待队长/管理员审批)
    -- 'pending_accept' (等待被邀请人接受)
    -- 'accepted', 'rejected', 'cancelled', 'expired'
    approved_by_hero_id UUID REFERENCES game_runtime.heroes(id) ON DELETE SET NULL,  -- 审批人英雄ID
    approved_at TIMESTAMPTZ,
    responded_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_invitation_status CHECK (status IN (
        'pending_approval', 'pending_accept', 'accepted', 'rejected', 'cancelled', 'expired'
    ))
);

COMMENT ON TABLE game_runtime.team_invitations IS '团队邀请表';
COMMENT ON COLUMN game_runtime.team_invitations.id IS '邀请ID';
COMMENT ON COLUMN game_runtime.team_invitations.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_invitations.inviter_hero_id IS '邀请人英雄ID';
COMMENT ON COLUMN game_runtime.team_invitations.invitee_hero_id IS '被邀请人英雄ID';
COMMENT ON COLUMN game_runtime.team_invitations.message IS '邀请留言';
COMMENT ON COLUMN game_runtime.team_invitations.status IS '状态';
COMMENT ON COLUMN game_runtime.team_invitations.approved_by_hero_id IS '审批人英雄ID';
COMMENT ON COLUMN game_runtime.team_invitations.approved_at IS '审批时间';
COMMENT ON COLUMN game_runtime.team_invitations.responded_at IS '被邀请人响应时间';
COMMENT ON COLUMN game_runtime.team_invitations.expires_at IS '过期时间';
COMMENT ON COLUMN game_runtime.team_invitations.created_at IS '创建时间';

-- 团队邀请表索引
CREATE INDEX IF NOT EXISTS idx_invitations_team_id ON game_runtime.team_invitations(team_id, status);
CREATE INDEX IF NOT EXISTS idx_invitations_invitee ON game_runtime.team_invitations(invitee_hero_id, status);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_at ON game_runtime.team_invitations(expires_at) WHERE status IN ('pending_approval', 'pending_accept');
CREATE INDEX IF NOT EXISTS idx_invitations_created_at ON game_runtime.team_invitations(created_at);

-- 5. 团队踢出记录表
CREATE TABLE IF NOT EXISTS game_runtime.team_kicked_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    kicked_by_hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    reason TEXT,
    kicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cooldown_until TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hours')
);

COMMENT ON TABLE game_runtime.team_kicked_records IS '团队踢出记录表（用于冷却期）';
COMMENT ON COLUMN game_runtime.team_kicked_records.id IS '记录ID';
COMMENT ON COLUMN game_runtime.team_kicked_records.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_kicked_records.hero_id IS '被踢出的英雄ID';
COMMENT ON COLUMN game_runtime.team_kicked_records.kicked_by_hero_id IS '踢出者英雄ID';
COMMENT ON COLUMN game_runtime.team_kicked_records.reason IS '踢出理由';
COMMENT ON COLUMN game_runtime.team_kicked_records.kicked_at IS '踢出时间';
COMMENT ON COLUMN game_runtime.team_kicked_records.cooldown_until IS '冷却期结束时间';

-- 团队踢出记录表索引
CREATE INDEX IF NOT EXISTS idx_kicked_records_team_hero ON game_runtime.team_kicked_records(team_id, hero_id);
CREATE INDEX IF NOT EXISTS idx_kicked_records_cooldown ON game_runtime.team_kicked_records(cooldown_until);
CREATE INDEX IF NOT EXISTS idx_kicked_records_kicked_at ON game_runtime.team_kicked_records(kicked_at);

-- 6. 团队仓库表
CREATE TABLE IF NOT EXISTS game_runtime.team_warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    gold_amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_team_warehouse UNIQUE (team_id),
    CONSTRAINT check_gold_amount CHECK (gold_amount >= 0)
);

COMMENT ON TABLE game_runtime.team_warehouses IS '团队仓库表';
COMMENT ON COLUMN game_runtime.team_warehouses.id IS '仓库ID';
COMMENT ON COLUMN game_runtime.team_warehouses.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_warehouses.gold_amount IS '金币数量';
COMMENT ON COLUMN game_runtime.team_warehouses.created_at IS '创建时间';
COMMENT ON COLUMN game_runtime.team_warehouses.updated_at IS '更新时间';

-- 团队仓库表索引
CREATE INDEX IF NOT EXISTS idx_warehouses_team_id ON game_runtime.team_warehouses(team_id);

-- 团队仓库表触发器
CREATE TRIGGER update_team_warehouses_updated_at
    BEFORE UPDATE ON game_runtime.team_warehouses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 7. 团队仓库物品表
CREATE TABLE IF NOT EXISTS game_runtime.team_warehouse_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id UUID NOT NULL REFERENCES game_runtime.team_warehouses(id) ON DELETE CASCADE,
    item_id UUID NOT NULL,  -- 物品配置ID（引用 game_config.items）
    item_type VARCHAR(50) NOT NULL,  -- 'equipment', 'consumable', 'material'
    quantity INT NOT NULL DEFAULT 1,
    source_dungeon_id UUID,  -- 来源地城（引用 game_config.dungeons）
    obtained_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_quantity CHECK (quantity > 0)
);

COMMENT ON TABLE game_runtime.team_warehouse_items IS '团队仓库物品表';
COMMENT ON COLUMN game_runtime.team_warehouse_items.id IS '物品记录ID';
COMMENT ON COLUMN game_runtime.team_warehouse_items.warehouse_id IS '仓库ID';
COMMENT ON COLUMN game_runtime.team_warehouse_items.item_id IS '物品配置ID';
COMMENT ON COLUMN game_runtime.team_warehouse_items.item_type IS '物品类型';
COMMENT ON COLUMN game_runtime.team_warehouse_items.quantity IS '数量';
COMMENT ON COLUMN game_runtime.team_warehouse_items.source_dungeon_id IS '来源地城ID';
COMMENT ON COLUMN game_runtime.team_warehouse_items.obtained_at IS '获得时间';

-- 团队仓库物品表索引
CREATE INDEX IF NOT EXISTS idx_warehouse_items_warehouse_id ON game_runtime.team_warehouse_items(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_items_item_id ON game_runtime.team_warehouse_items(item_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_items_obtained_at ON game_runtime.team_warehouse_items(obtained_at);
CREATE INDEX IF NOT EXISTS idx_warehouse_items_source_dungeon ON game_runtime.team_warehouse_items(source_dungeon_id);

-- 8. 战利品分配历史表
CREATE TABLE IF NOT EXISTS game_runtime.team_loot_distribution_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES game_runtime.teams(id) ON DELETE CASCADE,
    warehouse_id UUID NOT NULL REFERENCES game_runtime.team_warehouses(id) ON DELETE CASCADE,
    distributor_hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    recipient_hero_id UUID NOT NULL REFERENCES game_runtime.heroes(id) ON DELETE CASCADE,
    item_type VARCHAR(20) NOT NULL,  -- 'gold', 'item'
    item_id UUID,  -- 物品ID（如果是物品）
    quantity BIGINT NOT NULL,  -- 数量（金币数量或物品数量）
    distributed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_item_type CHECK (item_type IN ('gold', 'item')),
    CONSTRAINT check_distribution_quantity CHECK (quantity > 0)
);

COMMENT ON TABLE game_runtime.team_loot_distribution_history IS '战利品分配历史表';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.id IS '分配记录ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.team_id IS '团队ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.warehouse_id IS '仓库ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.distributor_hero_id IS '分配者英雄ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.recipient_hero_id IS '接收者英雄ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.item_type IS '类型：gold-金币, item-物品';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.item_id IS '物品ID';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.quantity IS '数量';
COMMENT ON COLUMN game_runtime.team_loot_distribution_history.distributed_at IS '分配时间';

-- 战利品分配历史表索引
CREATE INDEX IF NOT EXISTS idx_loot_history_team_id ON game_runtime.team_loot_distribution_history(team_id);
CREATE INDEX IF NOT EXISTS idx_loot_history_warehouse_id ON game_runtime.team_loot_distribution_history(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_loot_history_distributed_at ON game_runtime.team_loot_distribution_history(distributed_at);
CREATE INDEX IF NOT EXISTS idx_loot_history_recipient ON game_runtime.team_loot_distribution_history(recipient_hero_id);

