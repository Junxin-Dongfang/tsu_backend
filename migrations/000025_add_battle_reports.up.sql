-- +migrate Up
CREATE TABLE IF NOT EXISTS game_runtime.battle_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_id VARCHAR(64) NOT NULL,
    battle_code VARCHAR(64),
    team_id UUID,
    dungeon_id UUID,
    result_status VARCHAR(32) NOT NULL,
    loot_gold BIGINT DEFAULT 0,
    loot_items JSONB,
    participants JSONB NOT NULL,
    events JSONB,
    raw_payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (battle_id)
);

CREATE INDEX IF NOT EXISTS idx_battle_reports_team ON game_runtime.battle_reports(team_id);
CREATE INDEX IF NOT EXISTS idx_battle_reports_dungeon ON game_runtime.battle_reports(dungeon_id);
CREATE INDEX IF NOT EXISTS idx_battle_reports_status ON game_runtime.battle_reports(result_status);

CREATE TRIGGER update_battle_reports_updated_at
    BEFORE UPDATE ON game_runtime.battle_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE game_runtime.battle_reports IS '战斗结果回调记录，包括掉落与事件上下文';
COMMENT ON COLUMN game_runtime.battle_reports.battle_id IS '外部战斗引擎提供的唯一战斗 ID';
COMMENT ON COLUMN game_runtime.battle_reports.battle_code IS '战斗模板/房间编码';
COMMENT ON COLUMN game_runtime.battle_reports.team_id IS '触发战斗的团队 ID（可空）';
COMMENT ON COLUMN game_runtime.battle_reports.dungeon_id IS '所属地城 ID（可空）';
COMMENT ON COLUMN game_runtime.battle_reports.result_status IS '战斗结果状态，例如 victory/defeat/draw';
COMMENT ON COLUMN game_runtime.battle_reports.loot_gold IS '战斗奖励金币';
COMMENT ON COLUMN game_runtime.battle_reports.loot_items IS '战斗奖励物品 JSON 列表';
COMMENT ON COLUMN game_runtime.battle_reports.participants IS '战斗参与者 JSON 列表（英雄/怪物）';
COMMENT ON COLUMN game_runtime.battle_reports.events IS '战斗事件/日志 JSON 数据';
COMMENT ON COLUMN game_runtime.battle_reports.raw_payload IS '完整的回调 JSON Payload';
