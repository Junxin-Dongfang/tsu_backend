package impl

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

// HeroAllocatedAttributeRepositoryImpl 英雄已分配属性仓储实现
type HeroAllocatedAttributeRepositoryImpl struct {
	db *sql.DB
}

// NewHeroAllocatedAttributeRepository 创建英雄已分配属性仓储
func NewHeroAllocatedAttributeRepository(db *sql.DB) interfaces.HeroAllocatedAttributeRepository {
	return &HeroAllocatedAttributeRepositoryImpl{db: db}
}

// GetByHeroAndCode 根据英雄ID和属性代码获取
func (r *HeroAllocatedAttributeRepositoryImpl) GetByHeroAndCode(ctx context.Context, heroID, attributeCode string) (*game_runtime.HeroAllocatedAttribute, error) {
	attr, err := game_runtime.HeroAllocatedAttributes(
		qm.Where("hero_id = ?", heroID),
		qm.Where("attribute_code = ?", attributeCode),
		qm.Where("deleted_at IS NULL"),
	).One(ctx, r.db)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return attr, err
}

// GetByHeroID 获取英雄的所有已分配属性
func (r *HeroAllocatedAttributeRepositoryImpl) GetByHeroID(ctx context.Context, heroID string) ([]*game_runtime.HeroAllocatedAttribute, error) {
	attrs, err := game_runtime.HeroAllocatedAttributes(
		qm.Where("hero_id = ?", heroID),
		qm.Where("deleted_at IS NULL"),
		qm.OrderBy("attribute_code ASC"),
	).All(ctx, r.db)

	return attrs, err
}

// Create 创建新的已分配属性
func (r *HeroAllocatedAttributeRepositoryImpl) Create(ctx context.Context, executor interface{}, attr *game_runtime.HeroAllocatedAttribute) error {
	// 如果没有ID，生成UUID
	if attr.ID == "" {
		attr.ID = uuid.New().String()
	}

	var ex boil.ContextExecutor = r.db
	if executor != nil {
		ex = executor.(boil.ContextExecutor)
	}

	return attr.Insert(ctx, ex, boil.Infer())
}

// Update 更新已分配属性
func (r *HeroAllocatedAttributeRepositoryImpl) Update(ctx context.Context, executor interface{}, attr *game_runtime.HeroAllocatedAttribute) error {
	var ex boil.ContextExecutor = r.db
	if executor != nil {
		ex = executor.(boil.ContextExecutor)
	}

	_, err := attr.Update(ctx, ex, boil.Infer())
	return err
}

// Delete 软删除已分配属性
func (r *HeroAllocatedAttributeRepositoryImpl) Delete(ctx context.Context, executor interface{}, heroID, attributeCode string) error {
	var ex boil.ContextExecutor = r.db
	if executor != nil {
		ex = executor.(boil.ContextExecutor)
	}

	_, err := game_runtime.HeroAllocatedAttributes(
		qm.Where("hero_id = ?", heroID),
		qm.Where("attribute_code = ?", attributeCode),
		qm.Where("deleted_at IS NULL"),
	).UpdateAll(ctx, ex, map[string]interface{}{
		"deleted_at": null.TimeFrom(time.Now()),
	})

	return err
}

// BatchCreateForHero 为英雄批量创建初始属性
func (r *HeroAllocatedAttributeRepositoryImpl) BatchCreateForHero(ctx context.Context, executor interface{}, heroID string, attrs []*game_runtime.HeroAllocatedAttribute) error {
	var ex boil.ContextExecutor = r.db
	if executor != nil {
		ex = executor.(boil.ContextExecutor)
	}

	// 设置 hero_id 和生成 ID
	for _, attr := range attrs {
		attr.HeroID = heroID
		if attr.ID == "" {
			attr.ID = uuid.New().String()
		}
	}

	// 使用 SQLBoiler 的批量插入
	for _, attr := range attrs {
		if err := attr.Insert(ctx, ex, boil.Infer()); err != nil {
			return err
		}
	}

	return nil
}

// GetByHeroIDForUpdate 根据英雄ID获取属性（带锁用于事务）
func (r *HeroAllocatedAttributeRepositoryImpl) GetByHeroIDForUpdate(ctx context.Context, executor interface{}, heroID string) ([]*game_runtime.HeroAllocatedAttribute, error) {
	var ex boil.ContextExecutor = r.db
	if executor != nil {
		ex = executor.(boil.ContextExecutor)
	}

	attrs, err := game_runtime.HeroAllocatedAttributes(
		qm.Where("hero_id = ?", heroID),
		qm.Where("deleted_at IS NULL"),
		qm.OrderBy("attribute_code ASC"),
		qm.For("UPDATE"),
	).All(ctx, ex)

	return attrs, err
}
