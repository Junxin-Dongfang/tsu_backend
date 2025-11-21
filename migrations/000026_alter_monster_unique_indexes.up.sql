-- 000026_alter_monster_unique_indexes.up.sql
-- 调整怪物掉落与怪物技能的唯一约束，忽略已软删除的记录。

-- 掉落：唯一约束 monster_id + drop_pool_id，仅作用于未删除行
ALTER TABLE game_config.monster_drops DROP CONSTRAINT IF EXISTS uq_monster_drop_pool;
DROP INDEX IF EXISTS uq_monster_drop_pool;
CREATE UNIQUE INDEX uq_monster_drop_pool
    ON game_config.monster_drops (monster_id, drop_pool_id)
    WHERE deleted_at IS NULL;

-- 技能：唯一约束 monster_id + skill_id，仅作用于未删除行
ALTER TABLE game_config.monster_skills DROP CONSTRAINT IF EXISTS uq_monster_skill;
DROP INDEX IF EXISTS uq_monster_skill;
CREATE UNIQUE INDEX uq_monster_skill
    ON game_config.monster_skills (monster_id, skill_id)
    WHERE deleted_at IS NULL;
