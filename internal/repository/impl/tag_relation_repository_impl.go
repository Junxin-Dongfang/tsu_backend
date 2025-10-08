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
	db *sql.DB
}

// NewTagRelationRepository 创建标签关联仓储实例
func NewTagRelationRepository(db *sql.DB) interfaces.TagRelationRepository {
	return &tagRelationRepositoryImpl{db: db}
}

// GetByID 根据ID获取标签关联
func (r *tagRelationRepositoryImpl) GetByID(ctx context.Context, relationID string) (*game_config.TagsRelation, error) {
	relation, err := game_config.TagsRelations(
		qm.Where("id = ? AND deleted_at IS NULL", relationID),
	).One(ctx, r.db)

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
	// 构建基础查询条件
	var baseQueryMods []qm.QueryMod
	baseQueryMods = append(baseQueryMods, qm.Where("deleted_at IS NULL"))

	// 筛选条件
	if params.TagID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("tag_id = ?", *params.TagID))
	}
	if params.EntityType != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("entity_type = ?", *params.EntityType))
	}
	if params.EntityID != nil {
		baseQueryMods = append(baseQueryMods, qm.Where("entity_id = ?", *params.EntityID))
	}

	// 获取总数
	count, err := game_config.TagsRelations(baseQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, fmt.Errorf("查询标签关联总数失败: %w", err)
	}

	// 构建完整查询条件
	var queryMods []qm.QueryMod
	queryMods = append(queryMods, baseQueryMods...)

	// 排序
	queryMods = append(queryMods, qm.OrderBy("created_at DESC"))

	// 分页
	if params.Limit > 0 {
		queryMods = append(queryMods, qm.Limit(params.Limit))
	}
	if params.Offset > 0 {
		queryMods = append(queryMods, qm.Offset(params.Offset))
	}

	// 查询列表
	relations, err := game_config.TagsRelations(queryMods...).All(ctx, r.db)
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
	).All(ctx, r.db)

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
	).All(ctx, r.db)

	if err != nil {
		return nil, fmt.Errorf("查询标签实体失败: %w", err)
	}

	return relations, nil
}

// Create 创建标签关联
func (r *tagRelationRepositoryImpl) Create(ctx context.Context, relation *game_config.TagsRelation) error {
	// 生成 UUID
	if relation.ID == "" {
		relation.ID = uuid.New().String()
	}

	// 设置时间戳
	now := time.Now()
	relation.CreatedAt = now
	relation.UpdatedAt = now

	// 插入数据库
	if err := relation.Insert(ctx, r.db, boil.Infer()); err != nil {
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

	// 软删除
	now := time.Now()
	relation.DeletedAt.SetValid(now)
	relation.UpdatedAt = now

	if _, err := relation.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除标签关联失败: %w", err)
	}

	return nil
}

// DeleteByTagAndEntity 删除指定标签和实体的关联
func (r *tagRelationRepositoryImpl) DeleteByTagAndEntity(ctx context.Context, tagID string, entityType string, entityID string) error {
	relation, err := game_config.TagsRelations(
		qm.Where("tag_id = ? AND entity_type = ? AND entity_id = ? AND deleted_at IS NULL", tagID, entityType, entityID),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return fmt.Errorf("标签关联不存在")
	}
	if err != nil {
		return fmt.Errorf("查询标签关联失败: %w", err)
	}

	// 软删除
	now := time.Now()
	relation.DeletedAt.SetValid(now)
	relation.UpdatedAt = now

	if _, err := relation.Update(ctx, r.db, boil.Whitelist("deleted_at", "updated_at")); err != nil {
		return fmt.Errorf("删除标签关联失败: %w", err)
	}

	return nil
}

// Exists 检查标签和实体的关联是否存在
func (r *tagRelationRepositoryImpl) Exists(ctx context.Context, tagID string, entityType string, entityID string) (bool, error) {
	count, err := game_config.TagsRelations(
		qm.Where("tag_id = ? AND entity_type = ? AND entity_id = ? AND deleted_at IS NULL", tagID, entityType, entityID),
	).Count(ctx, r.db)

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

	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 逐个插入
	now := time.Now()
	for _, relation := range relations {
		// 生成 UUID
		if relation.ID == "" {
			relation.ID = uuid.New().String()
		}

		// 设置时间戳
		relation.CreatedAt = now
		relation.UpdatedAt = now

		// 插入数据库
		if err := relation.Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("批量创建标签关联失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// DeleteByEntity 删除实体的所有标签关联
func (r *tagRelationRepositoryImpl) DeleteByEntity(ctx context.Context, entityType string, entityID string) error {
	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 软删除所有关联
	now := time.Now()
	_, err = game_config.TagsRelations(
		qm.Where("entity_type = ? AND entity_id = ? AND deleted_at IS NULL", entityType, entityID),
	).UpdateAll(ctx, tx, game_config.M{
		"deleted_at": now,
		"updated_at": now,
	})

	if err != nil {
		return fmt.Errorf("删除实体标签关联失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}
