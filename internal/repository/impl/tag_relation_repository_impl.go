package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

type tagRelationRepositoryImpl struct {
	exec     boil.ContextExecutor
	beginner boil.ContextBeginner
}

// NewTagRelationRepository 创建标签关联仓储实例
func NewTagRelationRepository(db *sql.DB) interfaces.TagRelationRepository {
	return &tagRelationRepositoryImpl{
		exec:     db,
		beginner: db,
	}
}

// NewTagRelationRepositoryWithExecutor 使用自定义执行器创建仓储实例
func NewTagRelationRepositoryWithExecutor(exec boil.ContextExecutor) interfaces.TagRelationRepository {
	var beginner boil.ContextBeginner
	if b, ok := exec.(boil.ContextBeginner); ok {
		beginner = b
	}
	return &tagRelationRepositoryImpl{
		exec:     exec,
		beginner: beginner,
	}
}

// GetByID 根据ID获取标签关联
func (r *tagRelationRepositoryImpl) GetByID(ctx context.Context, relationID string) (*game_config.TagsRelation, error) {
	relation, err := game_config.TagsRelations(
		qm.Where("id = ? AND deleted_at IS NULL", relationID),
	).One(ctx, r.exec)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("标签关联不存在: %s", relationID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签关联失败: %w", err)
	}

	return relation, nil
}

// List 获取标签关联列表
func (r *tagRelationRepositoryImpl) List(ctx context.Context, params interfaces.TagRelationQueryParams) ([]*game_config.TagsRelation, int64, error) {
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	if params.TagID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("tag_id = ?", *params.TagID))
	}
	if params.EntityType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("entity_type = ?", *params.EntityType))
	}
	if params.EntityID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("entity_id = ?", *params.EntityID))
	}

	count, err := game_config.TagsRelations(baseQueryMods...).Count(ctx, r.exec)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签关联总数失败: %w", err)
	}

	queryMods := append([]qm.QueryMod{}, baseQueryMods...)
	queryMods = append(queryMods, qm.OrderBy("created_at DESC"))
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	relations, err := game_config.TagsRelations(queryMods...).All(ctx, r.exec)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签关联列表失败: %w", err)
	}

	return relations, count, nil
}

// GetEntityTags 获取实体的所有标签
func (r *tagRelationRepositoryImpl) GetEntityTags(ctx context.Context, entityType string, entityID string) ([]*game_config.Tag, error) {
	tags, err := game_config.Tags(
		qm.InnerJoin("game_config.tags_relations tr ON tags.id = tr.tag_id"),
		qm.Where("tr.entity_type = ? AND tr.entity_id = ? AND tr.deleted_at IS NULL", entityType, entityID),
		qm.Where("tags.deleted_at IS NULL"),
		qm.OrderBy("tags.display_order ASC"),
	).All(ctx, r.exec)
	if err != nil {
		return nil, fmt.Errorf("查询实体标签失败: %w", err)
	}

	return tags, nil
}

// GetTagEntities 获取使用某个标签的所有实体
func (r *tagRelationRepositoryImpl) GetTagEntities(ctx context.Context, tagID string) ([]*game_config.TagsRelation, error) {
	relations, err := game_config.TagsRelations(
		qm.Where("tag_id = ? AND deleted_at IS NULL", tagID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.exec)
	if err != nil {
		return nil, fmt.Errorf("查询标签实体失败: %w", err)
	}

	return relations, nil
}

// Create 创建标签关联
func (r *tagRelationRepositoryImpl) Create(ctx context.Context, relation *game_config.TagsRelation) error {
	if relation.ID == "" {
		relation.ID = uuid.New().String()
	}

	now := time.Now()
	relation.CreatedAt = now
	relation.UpdatedAt = now

	if err := relation.Insert(ctx, r.exec, boil.Infer()); err != nil {
		return fmt.Errorf("创建标签关联失败: %w", err)
	}

	return nil
}

// Delete 软删除标签关联
func (r *tagRelationRepositoryImpl) Delete(ctx context.Context, relationID string) error {
	relation, err := r.GetByID(ctx, relationID)
	if err != nil {
		return err
	}

	now := time.Now()
	relation.DeletedAt.SetValid(now)
	relation.UpdatedAt = now

	if _, err := relation.Update(ctx, r.exec, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除标签关联失败: %w", err)
	}

	return nil
}

// DeleteByTagAndEntity 删除指定标签和实体的关联
func (r *tagRelationRepositoryImpl) DeleteByTagAndEntity(ctx context.Context, tagID string, entityType string, entityID string) error {
	relation, err := game_config.TagsRelations(
		qm.Where("tag_id = ? AND entity_type = ? AND entity_id = ? AND deleted_at IS NULL", tagID, entityType, entityID),
	).One(ctx, r.exec)
	if err == sql.ErrNoRows {
		return fmt.Errorf("标签关联不存在")
	}
	if err != nil {
		return fmt.Errorf("查询标签关联失败: %w", err)
	}

	now := time.Now()
	relation.DeletedAt.SetValid(now)
	relation.UpdatedAt = now

	if _, err := relation.Update(ctx, r.exec, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除标签关联失败: %w", err)
	}

	return nil
}

// Exists 检查标签和实体的关联是否存在
func (r *tagRelationRepositoryImpl) Exists(ctx context.Context, tagID string, entityType string, entityID string) (bool, error) {
	count, err := game_config.TagsRelations(
		qm.Where("tag_id = ? AND entity_type = ? AND entity_id = ? AND deleted_at IS NULL", tagID, entityType, entityID),
	).Count(ctx, r.exec)
	if err != nil {
		return false, fmt.Errorf("检查标签关联是否存在失败: %w", err)
	}

	return count > 0, nil
}

// BatchCreate 批量创建标签关联
func (r *tagRelationRepositoryImpl) BatchCreate(ctx context.Context, relations []*game_config.TagsRelation) error {
	if len(relations) == 0 {
		return nil
	}

	if tx, ok := r.exec.(*sql.Tx); ok {
		return r.batchCreateWithExecutor(ctx, tx, relations)
	}

	if r.beginner != nil {
		tx, err := r.beginner.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启事务失败: %w", err)
		}
		defer tx.Rollback()

		if err := r.batchCreateWithExecutor(ctx, tx, relations); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
		return nil
	}

	return r.batchCreateWithExecutor(ctx, r.exec, relations)
}

func (r *tagRelationRepositoryImpl) batchCreateWithExecutor(ctx context.Context, exec boil.ContextExecutor, relations []*game_config.TagsRelation) error {
	now := time.Now()
	for _, relation := range relations {
		if relation.ID == "" {
			relation.ID = uuid.New().String()
		}
		relation.CreatedAt = now
		relation.UpdatedAt = now

		if err := relation.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建标签关联失败: %w", err)
		}
	}
	return nil
}

// DeleteByEntity 删除实体的所有标签关联
func (r *tagRelationRepositoryImpl) DeleteByEntity(ctx context.Context, entityType string, entityID string) error {
	if tx, ok := r.exec.(*sql.Tx); ok {
		return r.deleteEntityRelations(ctx, tx, entityType, entityID)
	}

	if r.beginner != nil {
		tx, err := r.beginner.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启事务失败: %w", err)
		}
		defer tx.Rollback()

		if err := r.deleteEntityRelations(ctx, tx, entityType, entityID); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
		return nil
	}

	return r.deleteEntityRelations(ctx, r.exec, entityType, entityID)
}

func (r *tagRelationRepositoryImpl) deleteEntityRelations(ctx context.Context, exec boil.ContextExecutor, entityType string, entityID string) error {
	now := time.Now()
	_, err := game_config.TagsRelations(
		qm.Where("entity_type = ? AND entity_id = ? AND deleted_at IS NULL", entityType, entityID),
	).UpdateAll(ctx, exec, game_config.M{
		"deleted_at": now,
		"updated_at": now,
	})
	if err != nil {
		return fmt.Errorf("删除实体标签关联失败: %w", err)
	}
	return nil
}
