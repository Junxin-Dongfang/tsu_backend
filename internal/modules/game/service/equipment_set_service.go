package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// EquipmentSetService 装备套装服务
type EquipmentSetService struct {
	equipmentSetRepo interfaces.EquipmentSetRepository
	equipmentRepo    interfaces.EquipmentRepository
	itemRepo         interfaces.ItemRepository
}

// NewEquipmentSetService 创建装备套装服务
func NewEquipmentSetService(
	equipmentSetRepo interfaces.EquipmentSetRepository,
	equipmentRepo interfaces.EquipmentRepository,
	itemRepo interfaces.ItemRepository,
) *EquipmentSetService {
	return &EquipmentSetService{
		equipmentSetRepo: equipmentSetRepo,
		equipmentRepo:    equipmentRepo,
		itemRepo:         itemRepo,
	}
}

// ActiveSetInfo 激活的套装信息
type ActiveSetInfo struct {
	SetID       string                     `json:"set_id"`
	SetCode     string                     `json:"set_code"`
	SetName     string                     `json:"set_name"`
	PieceCount  int                        `json:"piece_count"`   // 当前穿戴件数
	ActiveTiers []int                      `json:"active_tiers"`  // 激活的档位
	Bonuses     map[string]*AttributeBonus `json:"bonuses"`       // 属性加成
}

// SetInfo 套装详细信息
type SetInfo struct {
	SetID       string       `json:"set_id"`
	SetCode     string       `json:"set_code"`
	SetName     string       `json:"set_name"`
	Description string       `json:"description"`
	SetEffects  []SetEffect  `json:"set_effects"`
	Items       []ItemInfo   `json:"items"` // 套装包含的装备
}

// SetEffect 套装效果
type SetEffect struct {
	PieceCount       int                        `json:"piece_count"`
	RequiredPieces   int                        `json:"required_pieces"`   // 数据库字段名
	BonusDescription string                     `json:"bonus_description"` // 数据库字段名
	Effects          []SetEffectInfo            `json:"effects"`
	BonusEffects     []SetEffectInfo            `json:"bonus_effects"` // 数据库字段名
	Bonuses          map[string]*AttributeBonus `json:"bonuses"`       // 解析后的属性加成
}

// SetEffectInfo 套装效果信息
type SetEffectInfo struct {
	DataType    string `json:"Data_type"`
	DataID      string `json:"Data_ID"`
	BonusType   string `json:"Bouns_type"`
	BonusNumber string `json:"Bouns_Number"`
}

// ItemInfo 装备信息
type ItemInfo struct {
	ItemID      string `json:"item_id"`
	ItemCode    string `json:"item_code"`
	ItemName    string `json:"item_name"`
	ItemType    string `json:"item_type"`
	Quality     string `json:"quality"`
	Description string `json:"description"`
}

// ListSetsRequest 查询套装列表请求
type ListSetsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetActiveSets 获取英雄当前激活的套装效果
func (s *EquipmentSetService) GetActiveSets(ctx context.Context, heroID string) ([]*ActiveSetInfo, error) {
	// 1. 查询英雄已装备的物品
	equippedItems, err := s.equipmentRepo.GetEquippedItems(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get equipped items")
	}

	if len(equippedItems) == 0 {
		return []*ActiveSetInfo{}, nil
	}

	// 2. 查询装备配置以获取set_id
	itemIDs := make([]string, len(equippedItems))
	for i, item := range equippedItems {
		itemIDs[i] = item.ItemID
	}

	itemConfigs, err := s.itemRepo.GetByIDs(ctx, itemIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get item configs")
	}

	// 创建itemID到配置的映射
	itemConfigMap := make(map[string]*game_config.Item)
	for _, config := range itemConfigs {
		itemConfigMap[config.ID] = config
	}

	// 3. 提取装备的套装ID并统计件数
	setCountMap := make(map[string]int)
	for _, item := range equippedItems {
		config, ok := itemConfigMap[item.ItemID]
		if !ok {
			continue
		}
		// 检查装备是否属于某个套装
		if config.SetID.Valid {
			setID := config.SetID.String
			setCountMap[setID]++
		}
	}

	// 如果没有穿戴任何套装装备，返回空列表
	if len(setCountMap) == 0 {
		return []*ActiveSetInfo{}, nil
	}

	// 3. 查询套装配置
	setIDs := make([]string, 0, len(setCountMap))
	for setID := range setCountMap {
		setIDs = append(setIDs, setID)
	}

	setConfigs, err := s.equipmentSetRepo.GetByIDs(ctx, setIDs)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get set configs")
	}

	// 4. 判断每个套装的激活档位
	activeSets := make([]*ActiveSetInfo, 0)
	for _, setConfig := range setConfigs {
		pieceCount := setCountMap[setConfig.ID]

		// 解析套装效果
		var setEffects []SetEffect
		if err := json.Unmarshal(setConfig.SetEffects, &setEffects); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, fmt.Sprintf("failed to parse set effects for set %s", setConfig.SetCode))
		}

		// 找出所有满足件数要求的档位
		activeTiers := make([]int, 0)
		allBonuses := make(map[string]*AttributeBonus)

		for _, effect := range setEffects {
			// 使用 RequiredPieces 字段（数据库中的字段名）
			requiredPieces := effect.RequiredPieces
			if requiredPieces == 0 {
				requiredPieces = effect.PieceCount // 兼容旧字段名
			}

			if pieceCount >= requiredPieces {
				activeTiers = append(activeTiers, requiredPieces)

				// 使用 BonusEffects 字段（数据库中的字段名）
				effects := effect.BonusEffects
				if len(effects) == 0 {
					effects = effect.Effects // 兼容旧字段名
				}

				// 计算该档位的属性加成
				bonuses, err := s.parseSetEffectToBonuses(effects)
				if err != nil {
					return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to parse set effect to bonuses")
				}

				// 合并属性加成
				for attrID, bonus := range bonuses {
					if existing, ok := allBonuses[attrID]; ok {
						existing.FlatBonus += bonus.FlatBonus
						existing.PercentBonus += bonus.PercentBonus
					} else {
						allBonuses[attrID] = bonus
					}
				}
			}
		}

		// 如果有激活的档位，添加到结果中
		if len(activeTiers) > 0 {
			activeSets = append(activeSets, &ActiveSetInfo{
				SetID:       setConfig.ID,
				SetCode:     setConfig.SetCode,
				SetName:     setConfig.SetName,
				PieceCount:  pieceCount,
				ActiveTiers: activeTiers,
				Bonuses:     allBonuses,
			})
		}
	}

	return activeSets, nil
}

// CalculateSetBonuses 计算套装属性加成
func (s *EquipmentSetService) CalculateSetBonuses(ctx context.Context, heroID string) (map[string]*AttributeBonus, error) {
	// 获取激活的套装
	activeSets, err := s.GetActiveSets(ctx, heroID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get active sets")
	}

	// 合并所有套装的属性加成
	allBonuses := make(map[string]*AttributeBonus)
	for _, activeSet := range activeSets {
		for attrID, bonus := range activeSet.Bonuses {
			if existing, ok := allBonuses[attrID]; ok {
				existing.FlatBonus += bonus.FlatBonus
				existing.PercentBonus += bonus.PercentBonus
			} else {
				allBonuses[attrID] = &AttributeBonus{
					FlatBonus:    bonus.FlatBonus,
					PercentBonus: bonus.PercentBonus,
				}
			}
		}
	}

	return allBonuses, nil
}

// GetSetInfo 获取套装详细信息
func (s *EquipmentSetService) GetSetInfo(ctx context.Context, setID string) (*SetInfo, error) {
	// 查询套装配置
	setConfig, err := s.equipmentSetRepo.GetByID(ctx, setID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get set config")
	}

	// 查询套装包含的装备
	items, err := s.equipmentSetRepo.GetItemsBySetID(ctx, setID)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to get items by set ID")
	}

	// 解析套装效果
	var setEffects []SetEffect
	if err := json.Unmarshal(setConfig.SetEffects, &setEffects); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to parse set effects")
	}

	// 为每个效果计算属性加成
	for i := range setEffects {
		// 使用 BonusEffects 字段（数据库中的字段名）
		effects := setEffects[i].BonusEffects
		if len(effects) == 0 {
			effects = setEffects[i].Effects // 兼容旧字段名
		}

		bonuses, err := s.parseSetEffectToBonuses(effects)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to parse set effect to bonuses")
		}
		setEffects[i].Bonuses = bonuses

		// 设置 PieceCount 以便前端使用
		if setEffects[i].PieceCount == 0 {
			setEffects[i].PieceCount = setEffects[i].RequiredPieces
		}
	}

	// 转换装备信息
	itemInfos := make([]ItemInfo, len(items))
	for i, item := range items {
		itemInfos[i] = ItemInfo{
			ItemID:      item.ID,
			ItemCode:    item.ItemCode,
			ItemName:    item.ItemName,
			ItemType:    item.ItemType,
			Quality:     item.ItemQuality,
			Description: item.Description.String,
		}
	}

	return &SetInfo{
		SetID:       setConfig.ID,
		SetCode:     setConfig.SetCode,
		SetName:     setConfig.SetName,
		Description: setConfig.Description.String,
		SetEffects:  setEffects,
		Items:       itemInfos,
	}, nil
}

// ListAvailableSets 查询可用套装列表
func (s *EquipmentSetService) ListAvailableSets(ctx context.Context, req *ListSetsRequest) ([]*SetInfo, int64, error) {
	// 查询套装列表
	setConfigs, total, err := s.equipmentSetRepo.List(ctx, &interfaces.ListEquipmentSetsRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to list equipment sets")
	}

	// 转换为SetInfo
	setInfos := make([]*SetInfo, len(setConfigs))
	for i, setConfig := range setConfigs {
		// 解析套装效果
		var setEffects []SetEffect
		if err := json.Unmarshal(setConfig.SetEffects, &setEffects); err != nil {
			return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to parse set effects")
		}

		// 为每个效果计算属性加成
		for j := range setEffects {
			// 使用 BonusEffects 字段（数据库中的字段名）
			effects := setEffects[j].BonusEffects
			if len(effects) == 0 {
				effects = setEffects[j].Effects // 兼容旧字段名
			}

			bonuses, err := s.parseSetEffectToBonuses(effects)
			if err != nil {
				return nil, 0, xerrors.Wrap(err, xerrors.CodeInternalError, "failed to parse set effect to bonuses")
			}
			setEffects[j].Bonuses = bonuses

			// 设置 PieceCount 以便前端使用
			if setEffects[j].PieceCount == 0 {
				setEffects[j].PieceCount = setEffects[j].RequiredPieces
			}
		}

		setInfos[i] = &SetInfo{
			SetID:       setConfig.ID,
			SetCode:     setConfig.SetCode,
			SetName:     setConfig.SetName,
			Description: setConfig.Description.String,
			SetEffects:  setEffects,
			Items:       []ItemInfo{}, // 列表查询不包含装备详情
		}
	}

	return setInfos, total, nil
}

// parseSetEffectToBonuses 解析套装效果为属性加成
func (s *EquipmentSetService) parseSetEffectToBonuses(effects []SetEffectInfo) (map[string]*AttributeBonus, error) {
	bonuses := make(map[string]*AttributeBonus)

	for _, effect := range effects {
		// 只处理Status类型的效果
		if effect.DataType != "Status" {
			continue
		}

		attrID := effect.DataID
		bonusValue, err := strconv.ParseFloat(effect.BonusNumber, 64)
		if err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, fmt.Sprintf("invalid bonus number: %s", effect.BonusNumber))
		}

		if _, ok := bonuses[attrID]; !ok {
			bonuses[attrID] = &AttributeBonus{}
		}

		// 根据加成类型设置值
		switch effect.BonusType {
		case "flat":
			bonuses[attrID].FlatBonus += bonusValue
		case "percent":
			bonuses[attrID].PercentBonus += bonusValue
		default:
			return nil, xerrors.New(xerrors.CodeInvalidParams, fmt.Sprintf("unknown bonus type: %s", effect.BonusType))
		}
	}

	return bonuses, nil
}

