package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// MonsterService 怪物服务
type MonsterService struct {
	monsterRepo       interfaces.MonsterRepository
	monsterSkillRepo  interfaces.MonsterSkillRepository
	monsterDropRepo   interfaces.MonsterDropRepository
	skillRepo         interfaces.SkillRepository
	dropPoolRepo      interfaces.DropPoolRepository
	tagRelationRepo   interfaces.TagRelationRepository
	attributeTypeRepo interfaces.HeroAttributeTypeRepository
	db                *sql.DB
}

// NewMonsterService 创建怪物服务
func NewMonsterService(db *sql.DB) *MonsterService {
	return &MonsterService{
		monsterRepo:       impl.NewMonsterRepository(db),
		monsterSkillRepo:  impl.NewMonsterSkillRepository(db),
		monsterDropRepo:   impl.NewMonsterDropRepository(db),
		skillRepo:         impl.NewSkillRepository(db),
		dropPoolRepo:      impl.NewDropPoolRepository(db),
		tagRelationRepo:   impl.NewTagRelationRepository(db),
		attributeTypeRepo: impl.NewHeroAttributeTypeRepository(db),
		db:                db,
	}
}

// GetMonsters 获取怪物列表
func (s *MonsterService) GetMonsters(ctx context.Context, params interfaces.MonsterQueryParams) ([]*game_config.Monster, int64, error) {
	return s.monsterRepo.List(ctx, params)
}

// GetMonsterByID 根据ID获取怪物详情
func (s *MonsterService) GetMonsterByID(ctx context.Context, monsterID string) (*game_config.Monster, error) {
	return s.monsterRepo.GetByID(ctx, monsterID)
}

// GetMonsterByCode 根据代码获取怪物
func (s *MonsterService) GetMonsterByCode(ctx context.Context, code string) (*game_config.Monster, error) {
	return s.monsterRepo.GetByCode(ctx, code)
}

// CreateMonster 创建怪物
func (s *MonsterService) CreateMonster(ctx context.Context, monster *game_config.Monster) error {
	// 业务验证：检查怪物代码是否已存在
	exists, err := s.monsterRepo.Exists(ctx, monster.MonsterCode)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("怪物代码已存在: %s", monster.MonsterCode))
	}

	// 验证必填字段
	if monster.MonsterCode == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "怪物代码不能为空")
	}
	if monster.MonsterName == "" {
		return xerrors.New(xerrors.CodeInvalidParams, "怪物名称不能为空")
	}
	if monster.MonsterLevel < 1 || monster.MonsterLevel > 100 {
		return xerrors.New(xerrors.CodeInvalidParams, "怪物等级必须在1-100之间")
	}
	if monster.MaxHP <= 0 {
		return xerrors.New(xerrors.CodeInvalidParams, "最大生命值必须大于0")
	}

	// 验证基础属性范围
	if err := s.validateBaseAttributes(monster); err != nil {
		return err
	}

	// 验证属性类型代码
	if err := s.validateAttributeCodes(ctx, monster); err != nil {
		return err
	}

	return s.monsterRepo.Create(ctx, monster)
}

// UpdateMonster 更新怪物信息
func (s *MonsterService) UpdateMonster(ctx context.Context, monsterID string, updates map[string]interface{}) error {
	// 获取怪物
	monster, err := s.monsterRepo.GetByID(ctx, monsterID)
	if err != nil {
		return err
	}

	// 更新字段
	if monsterCode, ok := updates["monster_code"].(string); ok && monsterCode != "" {
		// 检查怪物代码是否已被使用
		exists, err := s.monsterRepo.ExistsExcludingID(ctx, monsterCode, monsterID)
		if err != nil {
			return err
		}
		if exists {
			return xerrors.New(xerrors.CodeDuplicateResource, fmt.Sprintf("怪物代码已被使用: %s", monsterCode))
		}
		monster.MonsterCode = monsterCode
	}

	if monsterName, ok := updates["monster_name"].(string); ok && monsterName != "" {
		monster.MonsterName = monsterName
	}

	if monsterLevel, ok := updates["monster_level"].(int16); ok {
		if monsterLevel < 1 || monsterLevel > 100 {
			return xerrors.New(xerrors.CodeInvalidParams, "怪物等级必须在1-100之间")
		}
		monster.MonsterLevel = monsterLevel
	}

	if description, ok := updates["description"].(string); ok {
		if description != "" {
			monster.Description.SetValid(description)
		} else {
			monster.Description.Valid = false
		}
	}

	if maxHP, ok := updates["max_hp"].(int); ok {
		if maxHP <= 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "最大生命值必须大于0")
		}
		monster.MaxHP = maxHP
	}

	if hpRecovery, ok := updates["hp_recovery"].(int); ok {
		if hpRecovery < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "生命恢复不能为负数")
		}
		monster.HPRecovery.SetValid(hpRecovery)
	}

	if maxMP, ok := updates["max_mp"].(int); ok {
		if maxMP < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "最大魔法值不能为负数")
		}
		monster.MaxMP.SetValid(maxMP)
	}

	if mpRecovery, ok := updates["mp_recovery"].(int); ok {
		if mpRecovery < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "魔法恢复不能为负数")
		}
		monster.MPRecovery.SetValid(mpRecovery)
	}

	// 更新基础属性
	s.updateBaseAttribute(&monster.BaseSTR, updates, "base_str")
	s.updateBaseAttribute(&monster.BaseAgi, updates, "base_agi")
	s.updateBaseAttribute(&monster.BaseVit, updates, "base_vit")
	s.updateBaseAttribute(&monster.BaseWLP, updates, "base_wlp")
	s.updateBaseAttribute(&monster.BaseInt, updates, "base_int")
	s.updateBaseAttribute(&monster.BaseWis, updates, "base_wis")
	s.updateBaseAttribute(&monster.BaseCha, updates, "base_cha")

	// 验证基础属性范围
	if err := s.validateBaseAttributes(monster); err != nil {
		return err
	}

	// 更新属性类型代码字段
	s.updateStringField(&monster.AccuracyAttributeCode, updates, "accuracy_attribute_code")
	s.updateStringField(&monster.DodgeAttributeCode, updates, "dodge_attribute_code")
	s.updateStringField(&monster.InitiativeAttributeCode, updates, "initiative_attribute_code")
	s.updateStringField(&monster.BodyResistAttributeCode, updates, "body_resist_attribute_code")
	s.updateStringField(&monster.MagicResistAttributeCode, updates, "magic_resist_attribute_code")
	s.updateStringField(&monster.MentalResistAttributeCode, updates, "mental_resist_attribute_code")
	s.updateStringField(&monster.EnvironmentResistAttributeCode, updates, "environment_resist_attribute_code")

	// 验证属性类型代码
	if err := s.validateAttributeCodes(ctx, monster); err != nil {
		return err
	}

	// 更新掉落配置
	if dropGoldMin, ok := updates["drop_gold_min"].(int); ok {
		if dropGoldMin < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "最小金币掉落不能为负数")
		}
		monster.DropGoldMin.SetValid(dropGoldMin)
	}

	if dropGoldMax, ok := updates["drop_gold_max"].(int); ok {
		if dropGoldMax < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "最大金币掉落不能为负数")
		}
		monster.DropGoldMax.SetValid(dropGoldMax)
	}

	if dropExp, ok := updates["drop_exp"].(int); ok {
		if dropExp < 0 {
			return xerrors.New(xerrors.CodeInvalidParams, "经验值掉落不能为负数")
		}
		monster.DropExp.SetValid(dropExp)
	}

	// 更新显示配置
	if iconURL, ok := updates["icon_url"].(string); ok {
		if iconURL != "" {
			monster.IconURL.SetValid(iconURL)
		} else {
			monster.IconURL.Valid = false
		}
	}

	if modelURL, ok := updates["model_url"].(string); ok {
		if modelURL != "" {
			monster.ModelURL.SetValid(modelURL)
		} else {
			monster.ModelURL.Valid = false
		}
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		monster.IsActive.SetValid(isActive)
	}

	if displayOrder, ok := updates["display_order"].(int); ok {
		monster.DisplayOrder.SetValid(displayOrder)
	}

	return s.monsterRepo.Update(ctx, monster)
}

// DeleteMonster 删除怪物
func (s *MonsterService) DeleteMonster(ctx context.Context, monsterID string) error {
	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 删除怪物（软删除）
	if err := s.monsterRepo.Delete(ctx, monsterID); err != nil {
		return err
	}

	// 级联删除怪物技能
	if err := s.monsterSkillRepo.DeleteByMonsterID(ctx, monsterID); err != nil {
		return err
	}

	// 级联删除怪物掉落配置
	if err := s.monsterDropRepo.DeleteByMonsterID(ctx, monsterID); err != nil {
		return err
	}

	// 级联删除怪物标签关联
	if err := s.tagRelationRepo.DeleteByEntity(ctx, "monster", monsterID); err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// validateBaseAttributes 验证基础属性范围
func (s *MonsterService) validateBaseAttributes(monster *game_config.Monster) error {
	attrs := map[string]null.Int16{
		"力量": monster.BaseSTR,
		"敏捷": monster.BaseAgi,
		"体质": monster.BaseVit,
		"意志": monster.BaseWLP,
		"智力": monster.BaseInt,
		"感知": monster.BaseWis,
		"魅力": monster.BaseCha,
	}

	for name, attr := range attrs {
		if attr.Valid && (attr.Int16 < 0 || attr.Int16 > 99) {
			return xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("%s属性必须在0-99之间", name))
		}
	}

	return nil
}

// updateBaseAttribute 更新基础属性
func (s *MonsterService) updateBaseAttribute(attr *null.Int16, updates map[string]interface{}, key string) {
	if value, ok := updates[key].(int16); ok {
		attr.SetValid(value)
	}
}

// updateStringField 更新字符串字段
func (s *MonsterService) updateStringField(field *null.String, updates map[string]interface{}, key string) {
	if value, ok := updates[key].(string); ok {
		if value != "" {
			field.SetValid(value)
		} else {
			field.Valid = false
		}
	}
}

// ===== 怪物技能管理 =====

// AddMonsterSkill 为怪物添加技能
func (s *MonsterService) AddMonsterSkill(ctx context.Context, monsterID, skillID string, skillLevel int16, gainActions []string) error {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return err
	}

	// 验证技能存在性
	if _, err := s.skillRepo.GetByID(ctx, skillID); err != nil {
		return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("技能不存在: %s", skillID))
	}

	// 验证技能等级范围
	if skillLevel < 1 || skillLevel > 20 {
		return xerrors.New(xerrors.CodeInvalidParams, "技能等级必须在1-20之间")
	}

	// 检查是否已添加该技能
	exists, err := s.monsterSkillRepo.Exists(ctx, monsterID, skillID)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, "该技能已添加到怪物")
	}

	// 创建怪物技能关联
	monsterSkill := &game_config.MonsterSkill{
		MonsterID:   monsterID,
		SkillID:     skillID,
		SkillLevel:  skillLevel,
		GainActions: gainActions,
	}

	return s.monsterSkillRepo.Create(ctx, monsterSkill)
}

// BatchSetMonsterSkills 批量设置怪物技能列表
func (s *MonsterService) BatchSetMonsterSkills(ctx context.Context, monsterID string, skills []struct {
	SkillID     string
	SkillLevel  int16
	GainActions []string
}) error {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return err
	}

	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 删除旧的技能配置
	if err := s.monsterSkillRepo.DeleteByMonsterID(ctx, monsterID); err != nil {
		return err
	}

	// 批量创建新的技能配置
	if len(skills) > 0 {
		monsterSkills := make([]*game_config.MonsterSkill, 0, len(skills))
		for _, skill := range skills {
			// 验证技能存在性
			if _, err := s.skillRepo.GetByID(ctx, skill.SkillID); err != nil {
				return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("技能不存在: %s", skill.SkillID))
			}

			// 验证技能等级范围
			if skill.SkillLevel < 1 || skill.SkillLevel > 20 {
				return xerrors.New(xerrors.CodeInvalidParams, "技能等级必须在1-20之间")
			}

			monsterSkills = append(monsterSkills, &game_config.MonsterSkill{
				MonsterID:   monsterID,
				SkillID:     skill.SkillID,
				SkillLevel:  skill.SkillLevel,
				GainActions: skill.GainActions,
			})
		}

		if err := s.monsterSkillRepo.BatchCreate(ctx, monsterSkills); err != nil {
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetMonsterSkills 获取怪物的技能列表
func (s *MonsterService) GetMonsterSkills(ctx context.Context, monsterID string) ([]*game_config.MonsterSkill, error) {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return nil, err
	}

	return s.monsterSkillRepo.GetByMonsterID(ctx, monsterID)
}

// UpdateMonsterSkill 更新怪物技能配置
func (s *MonsterService) UpdateMonsterSkill(ctx context.Context, monsterID, skillID string, skillLevel int16, gainActions []string) error {
	// 获取怪物技能关联
	monsterSkill, err := s.monsterSkillRepo.GetByMonsterAndSkill(ctx, monsterID, skillID)
	if err != nil {
		return err
	}

	// 验证技能等级范围
	if skillLevel < 1 || skillLevel > 20 {
		return xerrors.New(xerrors.CodeInvalidParams, "技能等级必须在1-20之间")
	}

	// 更新配置
	monsterSkill.SkillLevel = skillLevel
	monsterSkill.GainActions = gainActions

	return s.monsterSkillRepo.Update(ctx, monsterSkill)
}

// RemoveMonsterSkill 移除怪物技能
func (s *MonsterService) RemoveMonsterSkill(ctx context.Context, monsterID, skillID string) error {
	return s.monsterSkillRepo.Delete(ctx, monsterID, skillID)
}

// ===== 怪物掉落管理 =====

// AddMonsterDrop 为怪物添加掉落配置
func (s *MonsterService) AddMonsterDrop(ctx context.Context, monsterID, dropPoolID, dropType string, dropChance float64, minQuantity, maxQuantity int) error {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return err
	}

	// 验证掉落池存在性
	if _, err := s.dropPoolRepo.GetByID(ctx, dropPoolID); err != nil {
		return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("掉落池不存在: %s", dropPoolID))
	}

	// 验证掉落类型
	if dropType != "team" && dropType != "personal" {
		return xerrors.New(xerrors.CodeInvalidParams, "掉落类型必须是 team 或 personal")
	}

	// 验证掉落概率
	if dropChance <= 0 || dropChance > 1 {
		return xerrors.New(xerrors.CodeInvalidParams, "掉落概率必须在0-1之间")
	}

	// 验证数量范围
	if minQuantity < 1 {
		return xerrors.New(xerrors.CodeInvalidParams, "最小数量必须大于等于1")
	}
	if maxQuantity < minQuantity {
		return xerrors.New(xerrors.CodeInvalidParams, "最大数量必须大于等于最小数量")
	}

	// 检查是否已添加该掉落池
	exists, err := s.monsterDropRepo.Exists(ctx, monsterID, dropPoolID)
	if err != nil {
		return err
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, "该掉落池已添加到怪物")
	}

	// 创建怪物掉落配置
	dec := new(decimal.Big).SetFloat64(dropChance)
	monsterDrop := &game_config.MonsterDrop{
		MonsterID:   monsterID,
		DropPoolID:  dropPoolID,
		DropType:    dropType,
		DropChance:  types.NewDecimal(dec),
		MinQuantity: null.IntFrom(minQuantity),
		MaxQuantity: null.IntFrom(maxQuantity),
	}

	return s.monsterDropRepo.Create(ctx, monsterDrop)
}

// BatchSetMonsterDrops 批量设置怪物掉落配置列表
func (s *MonsterService) BatchSetMonsterDrops(ctx context.Context, monsterID string, drops []struct {
	DropPoolID  string
	DropType    string
	DropChance  float64
	MinQuantity int
	MaxQuantity int
}) error {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return err
	}

	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 删除旧的掉落配置
	if err := s.monsterDropRepo.DeleteByMonsterID(ctx, monsterID); err != nil {
		return err
	}

	// 批量创建新的掉落配置
	if len(drops) > 0 {
		monsterDrops := make([]*game_config.MonsterDrop, 0, len(drops))
		for _, drop := range drops {
			// 验证掉落池存在性
			if _, err := s.dropPoolRepo.GetByID(ctx, drop.DropPoolID); err != nil {
				return xerrors.New(xerrors.CodeResourceNotFound, fmt.Sprintf("掉落池不存在: %s", drop.DropPoolID))
			}

			// 验证掉落类型
			if drop.DropType != "team" && drop.DropType != "personal" {
				return xerrors.New(xerrors.CodeInvalidParams, "掉落类型必须是 team 或 personal")
			}

			// 验证掉落概率
			if drop.DropChance <= 0 || drop.DropChance > 1 {
				return xerrors.New(xerrors.CodeInvalidParams, "掉落概率必须在0-1之间")
			}

			// 验证数量范围
			if drop.MinQuantity < 1 {
				return xerrors.New(xerrors.CodeInvalidParams, "最小数量必须大于等于1")
			}
			if drop.MaxQuantity < drop.MinQuantity {
				return xerrors.New(xerrors.CodeInvalidParams, "最大数量必须大于等于最小数量")
			}

			dec := new(decimal.Big).SetFloat64(drop.DropChance)
			monsterDrops = append(monsterDrops, &game_config.MonsterDrop{
				MonsterID:   monsterID,
				DropPoolID:  drop.DropPoolID,
				DropType:    drop.DropType,
				DropChance:  types.NewDecimal(dec),
				MinQuantity: null.IntFrom(drop.MinQuantity),
				MaxQuantity: null.IntFrom(drop.MaxQuantity),
			})
		}

		if err := s.monsterDropRepo.BatchCreate(ctx, monsterDrops); err != nil {
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetMonsterDrops 获取怪物的掉落配置列表
func (s *MonsterService) GetMonsterDrops(ctx context.Context, monsterID string) ([]*game_config.MonsterDrop, error) {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return nil, err
	}

	return s.monsterDropRepo.GetByMonsterID(ctx, monsterID)
}

// UpdateMonsterDrop 更新怪物掉落配置
func (s *MonsterService) UpdateMonsterDrop(ctx context.Context, monsterID, dropPoolID, dropType string, dropChance float64, minQuantity, maxQuantity int) error {
	// 获取怪物掉落配置
	monsterDrop, err := s.monsterDropRepo.GetByMonsterAndPool(ctx, monsterID, dropPoolID)
	if err != nil {
		return err
	}

	// 验证掉落类型
	if dropType != "team" && dropType != "personal" {
		return xerrors.New(xerrors.CodeInvalidParams, "掉落类型必须是 team 或 personal")
	}

	// 验证掉落概率
	if dropChance <= 0 || dropChance > 1 {
		return xerrors.New(xerrors.CodeInvalidParams, "掉落概率必须在0-1之间")
	}

	// 验证数量范围
	if minQuantity < 1 {
		return xerrors.New(xerrors.CodeInvalidParams, "最小数量必须大于等于1")
	}
	if maxQuantity < minQuantity {
		return xerrors.New(xerrors.CodeInvalidParams, "最大数量必须大于等于最小数量")
	}

	// 更新配置
	dec := new(decimal.Big).SetFloat64(dropChance)
	monsterDrop.DropType = dropType
	monsterDrop.DropChance = types.NewDecimal(dec)
	monsterDrop.MinQuantity.SetValid(minQuantity)
	monsterDrop.MaxQuantity.SetValid(maxQuantity)

	return s.monsterDropRepo.Update(ctx, monsterDrop)
}

// RemoveMonsterDrop 移除怪物掉落配置
func (s *MonsterService) RemoveMonsterDrop(ctx context.Context, monsterID, dropPoolID string) error {
	return s.monsterDropRepo.Delete(ctx, monsterID, dropPoolID)
}

// ===== 怪物标签管理 =====

// SetMonsterTags 设置怪物标签
func (s *MonsterService) SetMonsterTags(ctx context.Context, monsterID string, tagIDs []string) error {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return err
	}

	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 删除旧的标签关联
	if err := s.tagRelationRepo.DeleteByEntity(ctx, "monster", monsterID); err != nil {
		return err
	}

	// 批量创建新的标签关联
	if len(tagIDs) > 0 {
		relations := make([]*game_config.TagsRelation, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			relations = append(relations, &game_config.TagsRelation{
				TagID:      tagID,
				EntityType: "monster",
				EntityID:   monsterID,
			})
		}

		if err := s.tagRelationRepo.BatchCreate(ctx, relations); err != nil {
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetMonsterTags 获取怪物标签列表
func (s *MonsterService) GetMonsterTags(ctx context.Context, monsterID string) ([]*game_config.Tag, error) {
	// 验证怪物存在性
	if _, err := s.monsterRepo.GetByID(ctx, monsterID); err != nil {
		return nil, err
	}

	return s.tagRelationRepo.GetEntityTags(ctx, "monster", monsterID)
}

// GetAttributeFormula 获取怪物的属性计算公式
// 通过属性类型代码从 hero_attribute_type 表获取公式
func (s *MonsterService) GetAttributeFormula(ctx context.Context, attributeCode string) (string, error) {
	attributeType, err := s.attributeTypeRepo.GetByCode(ctx, attributeCode)
	if err != nil {
		return "", err
	}
	if attributeType == nil {
		return "", xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("属性类型不存在: %s", attributeCode))
	}

	return attributeType.CalculationFormula.String, nil
}

// validateAttributeCodes 验证属性类型代码
func (s *MonsterService) validateAttributeCodes(ctx context.Context, monster *game_config.Monster) error {
	// 验证战斗属性类型代码
	if monster.AccuracyAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.AccuracyAttributeCode.String); err != nil {
			return err
		}
	}
	if monster.DodgeAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.DodgeAttributeCode.String); err != nil {
			return err
		}
	}
	if monster.InitiativeAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.InitiativeAttributeCode.String); err != nil {
			return err
		}
	}

	// 验证豁免属性类型代码
	if monster.BodyResistAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.BodyResistAttributeCode.String); err != nil {
			return err
		}
	}
	if monster.MagicResistAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.MagicResistAttributeCode.String); err != nil {
			return err
		}
	}
	if monster.MentalResistAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.MentalResistAttributeCode.String); err != nil {
			return err
		}
	}
	if monster.EnvironmentResistAttributeCode.Valid {
		if err := s.validateAttributeCode(ctx, monster.EnvironmentResistAttributeCode.String); err != nil {
			return err
		}
	}

	return nil
}

// validateAttributeCode 验证单个属性类型代码
func (s *MonsterService) validateAttributeCode(ctx context.Context, attributeCode string) error {
	if attributeCode == "" {
		return nil // 允许空值，使用默认值
	}

	attributeType, err := s.attributeTypeRepo.GetByCode(ctx, attributeCode)
	if err != nil {
		// 如果是"不存在"错误，返回参数错误而不是内部错误
		if attributeType == nil {
			userMsg := fmt.Sprintf("属性类型不存在: %s", attributeCode)
			return xerrors.FromCode(xerrors.CodeInvalidParams).
				WithMetadata("user_message", userMsg).
				WithMetadata("attribute_code", attributeCode).
				WithMetadata("validation_failed", "attribute_type_not_found")
		}
		// 其他数据库错误
		return xerrors.Wrap(err, xerrors.CodeDatabaseError, "查询属性类型失败")
	}
	if attributeType == nil {
		userMsg := fmt.Sprintf("属性类型不存在: %s", attributeCode)
		return xerrors.FromCode(xerrors.CodeInvalidParams).
			WithMetadata("user_message", userMsg).
			WithMetadata("attribute_code", attributeCode).
			WithMetadata("validation_failed", "attribute_type_not_found")
	}

	return nil
}
