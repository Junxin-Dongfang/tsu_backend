-- 000026_alter_monster_unique_indexes.down.sql
-- 回滚：恢复未过滤 deleted_at 的唯一约束。

DROP INDEX IF EXISTS uq_monster_drop_pool;
ALTER TABLE game_config.monster_drops
    ADD CONSTRAINT uq_monster_drop_pool UNIQUE (monster_id, drop_pool_id);

DROP INDEX IF EXISTS uq_monster_skill;
ALTER TABLE game_config.monster_skills
    ADD CONSTRAINT uq_monster_skill UNIQUE (monster_id, skill_id);
