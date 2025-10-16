package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// HeroAttributeService 英雄属性服务
type HeroAttributeService struct {
	db                           *sql.DB
	heroRepo                     interfaces.HeroRepository
	attributeUpgradeCostRepo     interfaces.AttributeUpgradeCostRepository
	attributeOpRepo              interfaces.HeroAttributeOperationRepository
	heroAttributeTypeRepo        interfaces.HeroAttributeTypeRepository
	heroAllocatedAttributeRepo   interfaces.HeroAllocatedAttributeRepository
	heroService                  *HeroService
}

// NewHeroAttributeService 创建英雄属性服务
func NewHeroAttributeService(db *sql.DB) *HeroAttributeService {
	return &HeroAttributeService{
		db:                       db,
		heroRepo:                 impl.NewHeroRepository(db),
		attributeUpgradeCostRepo: impl.NewAttributeUpgradeCostRepository(db),
		attributeOpRepo:          impl.NewHeroAttributeOperationRepository(db),
		heroAttributeTypeRepo:    impl.NewHeroAttributeTypeRepository(db),
		heroService:              NewHeroService(db),
	}
}

// AllocateAttributeRequest 属性加点请求
type AllocateAttributeRequest struct {
	HeroID        string `json:"hero_id"`
	AttributeCode string `json:"attribute_code"`
	PointsToAdd   int    `json:"points_to_add"` // 要加的点数
}

// AllocateAttribute 属性加点
func (s *HeroAttributeService) AllocateAttribute(ctx context.Context, req *AllocateAttributeRequest) error {
	if req.PointsToAdd <= 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "加点数量必须大于0")
	}

	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 3. 验证属性代码有效性
	_, err = s.heroAttributeTypeRepo.GetByCode(ctx, req.AttributeCode)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "属性代码无效")
	}

	// 4. 获取当前属性值和已花费经验（从 hero_allocated_attributes 表）
	allocatedAttr, err := s.heroAllocatedAttributeRepo.GetByHeroAndCode(ctx, req.HeroID, req.AttributeCode)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询属性分配失败")
	}
	if allocatedAttr == nil {
		return xerrors.New(xerrors.CodeInvalidParams, "属性不存在")
	}

	currentValue := allocatedAttr.Value
	currentSpentXP := allocatedAttr.SpentXP

	// 5. 计算本次加点消耗
	fromPoint := currentValue
	toPoint := currentValue + req.PointsToAdd
	totalCost, err := s.attributeUpgradeCostRepo.CalculateCost(ctx, fromPoint, toPoint)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "计算加点消耗失败")
	}

	// 6. 验证经验是否足够
	if hero.ExperienceAvailable < int64(totalCost) {
		return xerrors.New(xerrors.CodeInsufficientExperience,
			fmt.Sprintf("经验不足: 需要 %d，当前 %d", totalCost, hero.ExperienceAvailable))
	}

	// 7. 扣除经验（注：experience_total = experience_available + experience_spent，不应在此增加）
	hero.ExperienceAvailable -= int64(totalCost)
	hero.ExperienceSpent += int64(totalCost)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 8. 更新已分配属性（hero_allocated_attributes 表）
	allocatedAttr.Value = toPoint
	allocatedAttr.SpentXP = currentSpentXP + totalCost
	allocatedAttr.UpdatedAt = time.Now()

	if err := s.heroAllocatedAttributeRepo.Update(ctx, tx, allocatedAttr); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新属性分配失败")
	}

	// 9. 创建操作历史记录（rollback_deadline = now + 1小时）
	operation := &game_runtime.HeroAttributeOperation{
		ID:               uuid.New().String(),
		HeroID:           req.HeroID,
		AttributeCode:    req.AttributeCode,
		PointsAdded:      req.PointsToAdd,
		XPSpent:          totalCost,
		ValueBefore:      currentValue,
		ValueAfter:       toPoint,
		CreatedAt:        time.Now(),
		RollbackDeadline: time.Now().Add(1 * time.Hour),
	}

	if err := s.attributeOpRepo.Create(ctx, tx, operation); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "创建操作历史失败")
	}

	// 10. 调用 AutoLevelUp 检查是否可以升级
	_, _, err = s.heroService.AutoLevelUp(ctx, tx, req.HeroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "自动升级检查失败")
	}

	// 11. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// RollbackAttributeAllocation 回退属性加点（栈式）
func (s *HeroAttributeService) RollbackAttributeAllocation(ctx context.Context, heroID, attributeCode string) error {
	// 1. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 2. 获取该属性最近一次未回退且未过期的操作（栈顶）
	operation, err := s.attributeOpRepo.GetLatestRollbackable(ctx, heroID, attributeCode)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询可回退操作失败")
	}
	if operation == nil {
		return xerrors.New(xerrors.CodeResourceNotFound, "没有可回退的操作")
	}

	// 3. 验证操作存在且未过期
	if time.Now().After(operation.RollbackDeadline) {
		return xerrors.New(xerrors.CodeOperationExpired, "回退时间已过期")
	}

	// 4. 获取英雄信息（加锁）
	hero, err := s.heroRepo.GetByIDForUpdate(ctx, tx, heroID)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeHeroNotFound, "英雄不存在")
	}

	// 5. 返还经验（注：experience_total 不应改变，它是累计获得的总经验）
	hero.ExperienceAvailable += int64(operation.XPSpent)
	hero.ExperienceSpent -= int64(operation.XPSpent)
	hero.UpdatedAt = time.Now()

	if err := s.heroRepo.Update(ctx, tx, hero); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新英雄失败")
	}

	// 6. 获取当前已分配属性（hero_allocated_attributes 表）
	allocatedAttr, err := s.heroAllocatedAttributeRepo.GetByHeroAndCode(ctx, heroID, attributeCode)
	if err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "查询属性分配失败")
	}
	if allocatedAttr == nil {
		return xerrors.New(xerrors.CodeInvalidParams, "属性不存在")
	}

	// 7. 回退属性值
	allocatedAttr.Value = operation.ValueBefore
	allocatedAttr.SpentXP = allocatedAttr.SpentXP - operation.XPSpent
	allocatedAttr.UpdatedAt = time.Now()

	if err := s.heroAllocatedAttributeRepo.Update(ctx, tx, allocatedAttr); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "更新属性分配失败")
	}

	// 8. 标记操作为已回退
	if err := s.attributeOpRepo.MarkAsRolledBack(ctx, tx, operation.ID); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "标记回退失败")
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return nil
}

// GetComputedAttributes 获取英雄的计算属性（通过视图）
func (s *HeroAttributeService) GetComputedAttributes(ctx context.Context, heroID string) ([]*game_runtime.HeroComputedAttribute, error) {
	// 查询视图
	attributes, err := game_runtime.HeroComputedAttributes(
		game_runtime.HeroComputedAttributeWhere.HeroID.EQ(null.StringFrom(heroID)),
	).All(ctx, s.db)

	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询计算属性失败")
	}

	return attributes, nil
}
