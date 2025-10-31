package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/aarondl/null/v8"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ItemDropService 物品掉落服务
type ItemDropService struct {
	db                     *sql.DB
	dropPoolRepo           interfaces.DropPoolRepository
	worldDropConfigRepo    interfaces.WorldDropConfigRepository
	worldDropStatsRepo     interfaces.WorldDropStatsRepository
	itemRepo               interfaces.ItemRepository
	playerItemRepo         interfaces.PlayerItemRepository
	itemDropRecordRepo     interfaces.ItemDropRecordRepository
}

// NewItemDropService 创建物品掉落服务
func NewItemDropService(db *sql.DB) *ItemDropService {
	return &ItemDropService{
		db:                  db,
		dropPoolRepo:        impl.NewDropPoolRepository(db),
		worldDropConfigRepo: impl.NewWorldDropConfigRepository(db),
		worldDropStatsRepo:  impl.NewWorldDropStatsRepository(db),
		itemRepo:            impl.NewItemRepository(db),
		playerItemRepo:      impl.NewPlayerItemRepository(db),
		itemDropRecordRepo:  impl.NewItemDropRecordRepository(db),
	}
}

// DropFromMonsterRequest 从怪物掉落请求
type DropFromMonsterRequest struct {
	MonsterID    string `json:"monster_id"`     // 怪物ID
	PlayerID     string `json:"player_id"`      // 玩家ID
	PlayerLevel  int    `json:"player_level"`   // 玩家等级
	TeamID       string `json:"team_id"`        // 队伍ID(可选)
	DungeonID    string `json:"dungeon_id"`     // 地城ID(可选)
	DungeonLevel int    `json:"dungeon_level"`  // 地城等级(可选)
}

// DropFromMonsterResponse 从怪物掉落响应
type DropFromMonsterResponse struct {
	DroppedItems []*game_runtime.PlayerItem `json:"dropped_items"`
	Message      string                     `json:"message"`
}

// DropFromMonster 从怪物掉落池掉落物品
func (s *ItemDropService) DropFromMonster(ctx context.Context, req *DropFromMonsterRequest) (*DropFromMonsterResponse, error) {
	// 1. 验证参数
	if req.MonsterID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "怪物ID不能为空")
	}
	if req.PlayerID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "玩家ID不能为空")
	}

	// 2. TODO: 查询怪物的掉落池
	// 这需要一个怪物配置表,关联到掉落池
	// 暂时使用示例掉落池代码
	poolCode := fmt.Sprintf("monster_%s", req.MonsterID)

	// 3. 查询掉落池
	pool, err := s.dropPoolRepo.GetByCode(ctx, poolCode)
	if err != nil {
		// 如果没有配置掉落池,返回空掉落
		return &DropFromMonsterResponse{
			DroppedItems: []*game_runtime.PlayerItem{},
			Message:      "该怪物没有掉落",
		}, nil
	}

	// 4. 查询掉落池中符合等级要求的物品
	poolItems, err := s.dropPoolRepo.GetPoolItemsByLevel(ctx, pool.ID, req.PlayerLevel)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询掉落池物品失败")
	}

	if len(poolItems) == 0 {
		return &DropFromMonsterResponse{
			DroppedItems: []*game_runtime.PlayerItem{},
			Message:      "没有符合等级要求的掉落物品",
		}, nil
	}

	// 5. 确定掉落数量
	dropCount := s.determineDropCount(pool)

	// 6. 从掉落池中随机选择物品
	selectedItems := s.selectItemsFromPool(poolItems, dropCount)

	// 7. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	// 8. 创建物品实例
	droppedItems := make([]*game_runtime.PlayerItem, 0, len(selectedItems))

	for _, poolItem := range selectedItems {
		// 获取物品配置
		itemConfig, err := s.itemRepo.GetByID(ctx, poolItem.ItemID)
		if err != nil {
			continue
		}

		// 随机生成品质
		quality := s.randomQuality(poolItem.QualityWeights, itemConfig.ItemQuality)

		// 创建物品实例
		playerItem := &game_runtime.PlayerItem{
			ItemID:       poolItem.ItemID,
			OwnerID:      req.PlayerID,
			SourceType:   "dungeon",
			SourceID:     null.StringFrom(req.DungeonID),
			ItemLocation: "backpack",
			StackCount:   null.IntFrom(s.randomQuantity(poolItem.MinQuantity, poolItem.MaxQuantity)),
		}

		// 如果是装备,初始化耐久度
		if itemConfig.ItemType == "equipment" && itemConfig.MaxDurability.Valid {
			playerItem.CurrentDurability = null.IntFrom(int(itemConfig.MaxDurability.Int))
			playerItem.MaxDurabilityOverride = itemConfig.MaxDurability
		}

		// 创建物品实例
		if err := s.playerItemRepo.Create(ctx, tx, playerItem); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建物品实例失败")
		}

		// 记录掉落历史
		dropRecord := &game_runtime.ItemDropRecord{
			ItemInstanceID: playerItem.ID,
			ItemConfigID:   poolItem.ItemID,
			DropSource:     "dungeon",
			SourceID:       null.StringFrom(req.DungeonID),
			ReceiverID:     req.PlayerID,
			TeamID:         null.StringFrom(req.TeamID),
			PlayerLevel:    null.Int16From(int16(req.PlayerLevel)),
			ItemQuality:    null.StringFrom(quality),
		}

		if err := s.itemDropRecordRepo.Create(ctx, tx, dropRecord); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "记录掉落历史失败")
		}

		droppedItems = append(droppedItems, playerItem)
	}

	// 9. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return &DropFromMonsterResponse{
		DroppedItems: droppedItems,
		Message:      fmt.Sprintf("掉落了%d个物品", len(droppedItems)),
	}, nil
}

// determineDropCount 确定掉落数量
func (s *ItemDropService) determineDropCount(pool *game_config.DropPool) int {
	// 保底掉落
	if pool.GuaranteedDrops.Valid && pool.GuaranteedDrops.Int16 > 0 {
		return int(pool.GuaranteedDrops.Int16)
	}

	// 随机掉落数量(min_drops ~ max_drops)
	minDrops := 0
	maxDrops := 1

	if pool.MinDrops.Valid {
		minDrops = int(pool.MinDrops.Int16)
	}
	if pool.MaxDrops.Valid {
		maxDrops = int(pool.MaxDrops.Int16)
	}

	if minDrops >= maxDrops {
		return minDrops
	}

	return minDrops + rand.Intn(maxDrops-minDrops+1)
}

// selectItemsFromPool 从掉落池中随机选择物品
func (s *ItemDropService) selectItemsFromPool(poolItems []*game_config.DropPoolItem, count int) []*game_config.DropPoolItem {
	if count <= 0 || len(poolItems) == 0 {
		return []*game_config.DropPoolItem{}
	}

	selected := make([]*game_config.DropPoolItem, 0, count)

	for i := 0; i < count; i++ {
		// 随机选择一个物品
		item := s.selectItemByWeight(poolItems)
		if item != nil {
			selected = append(selected, item)
		}
	}

	return selected
}

// selectItemByWeight 根据权重随机选择物品
func (s *ItemDropService) selectItemByWeight(poolItems []*game_config.DropPoolItem) *game_config.DropPoolItem {
	// 计算总权重
	totalWeight := 0
	for _, item := range poolItems {
		// 如果设置了固定概率,使用固定概率
		if !item.DropRate.IsZero() {
			// 固定概率判定
			dropRate, _ := item.DropRate.Float64()
			if rand.Float64() < dropRate {
				return item
			}
			continue
		}

		// 使用权重
		totalWeight += item.DropWeight
	}

	if totalWeight == 0 {
		return nil
	}

	// 随机选择
	randomValue := rand.Intn(totalWeight)
	currentWeight := 0

	for _, item := range poolItems {
		if !item.DropRate.IsZero() {
			continue // 跳过固定概率的物品
		}

		currentWeight += item.DropWeight
		if randomValue < currentWeight {
			return item
		}
	}

	return nil
}

// randomQuality 随机生成品质
func (s *ItemDropService) randomQuality(qualityWeightsJSON null.JSON, defaultQuality string) string {
	if !qualityWeightsJSON.Valid {
		return defaultQuality
	}

	// 解析品质权重
	var weights map[string]int
	if err := json.Unmarshal(qualityWeightsJSON.JSON, &weights); err != nil {
		return defaultQuality
	}

	// 计算总权重
	totalWeight := 0
	for _, weight := range weights {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return defaultQuality
	}

	// 随机选择品质
	randomValue := rand.Intn(totalWeight)
	currentWeight := 0

	for quality, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return quality
		}
	}

	return defaultQuality
}

// randomQuantity 随机生成数量
func (s *ItemDropService) randomQuantity(minQuantity, maxQuantity null.Int) int {
	min := 1
	max := 1

	if minQuantity.Valid {
		min = minQuantity.Int
	}
	if maxQuantity.Valid {
		max = maxQuantity.Int
	}

	if min >= max {
		return min
	}

	return min + rand.Intn(max-min+1)
}

// CheckWorldDropRequest 检查世界掉落请求
type CheckWorldDropRequest struct {
	PlayerID      string `json:"player_id"`
	PlayerLevel   int    `json:"player_level"`
	DungeonID     string `json:"dungeon_id"`
	DungeonType   string `json:"dungeon_type"`   // elite/boss/normal
	DungeonLevel  int    `json:"dungeon_level"`
	TeamSize      int    `json:"team_size"`
	IsFirstKill   bool   `json:"is_first_kill"`
	TeamID        string `json:"team_id"`
}

// CheckWorldDropResponse 检查世界掉落响应
type CheckWorldDropResponse struct {
	DroppedItems []*game_runtime.PlayerItem `json:"dropped_items"`
	Message      string                     `json:"message"`
}

// CheckWorldDrop 检查世界掉落
func (s *ItemDropService) CheckWorldDrop(ctx context.Context, req *CheckWorldDropRequest) (*CheckWorldDropResponse, error) {
	// 1. 验证参数
	if req.PlayerID == "" {
		return nil, xerrors.New(xerrors.CodeInvalidParams, "玩家ID不能为空")
	}

	// 2. 查询所有激活的世界掉落配置
	configs, err := s.worldDropConfigRepo.GetActiveConfigs(ctx)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询世界掉落配置失败")
	}

	if len(configs) == 0 {
		return &CheckWorldDropResponse{
			DroppedItems: []*game_runtime.PlayerItem{},
			Message:      "没有世界掉落配置",
		}, nil
	}

	// 3. 筛选符合触发条件的配置
	eligibleConfigs := s.filterConfigsByTriggerCondition(configs, req)

	if len(eligibleConfigs) == 0 {
		return &CheckWorldDropResponse{
			DroppedItems: []*game_runtime.PlayerItem{},
			Message:      "没有符合触发条件的世界掉落",
		}, nil
	}

	// 4. 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "开启事务失败")
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// 仅当 Rollback 失败且不是已提交的事务时，才表示有问题
		}
	}()

	droppedItems := make([]*game_runtime.PlayerItem, 0)

	// 5. 遍历每个配置,检查是否掉落
	for _, config := range eligibleConfigs {
		// 检查全局限制
		canDrop, err := s.checkWorldDropLimits(ctx, tx, config)
		if err != nil {
			return nil, err
		}
		if !canDrop {
			continue
		}

		// 计算最终掉落概率
		finalDropRate := s.calculateFinalDropRate(config, req)

		// 随机判定是否掉落
		if rand.Float64() > finalDropRate {
			continue
		}

		// 创建物品实例
		itemConfig, err := s.itemRepo.GetByID(ctx, config.ItemID)
		if err != nil {
			continue
		}

		playerItem := &game_runtime.PlayerItem{
			ItemID:       config.ItemID,
			OwnerID:      req.PlayerID,
			SourceType:   "dungeon",
			SourceID:     null.StringFrom(req.DungeonID),
			ItemLocation: "backpack",
			StackCount:   null.IntFrom(1),
		}

		// 如果是装备,初始化耐久度
		if itemConfig.ItemType == "equipment" && itemConfig.MaxDurability.Valid {
			playerItem.CurrentDurability = null.IntFrom(int(itemConfig.MaxDurability.Int))
			playerItem.MaxDurabilityOverride = itemConfig.MaxDurability
		}

		// 创建物品实例
		if err := s.playerItemRepo.Create(ctx, tx, playerItem); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "创建物品实例失败")
		}

		// 记录掉落历史
		dropRecord := &game_runtime.ItemDropRecord{
			ItemInstanceID: playerItem.ID,
			ItemConfigID:   config.ItemID,
			DropSource:     "dungeon",
			SourceID:       null.StringFrom(req.DungeonID),
			ReceiverID:     req.PlayerID,
			TeamID:         null.StringFrom(req.TeamID),
			PlayerLevel:    null.Int16From(int16(req.PlayerLevel)),
			ItemQuality:    null.StringFrom(itemConfig.ItemQuality),
		}

		if err := s.itemDropRecordRepo.Create(ctx, tx, dropRecord); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "记录掉落历史失败")
		}

		// 更新世界掉落统计
		if err := s.worldDropStatsRepo.IncrementDropCount(ctx, tx, config.ID); err != nil {
			return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "更新世界掉落统计失败")
		}

		droppedItems = append(droppedItems, playerItem)
	}

	// 6. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "提交事务失败")
	}

	return &CheckWorldDropResponse{
		DroppedItems: droppedItems,
		Message:      fmt.Sprintf("世界掉落了%d个物品", len(droppedItems)),
	}, nil
}

// filterConfigsByTriggerCondition 筛选符合触发条件的配置
func (s *ItemDropService) filterConfigsByTriggerCondition(
	configs []*game_config.WorldDropConfig,
	req *CheckWorldDropRequest,
) []*game_config.WorldDropConfig {
	eligible := make([]*game_config.WorldDropConfig, 0)

	for _, config := range configs {
		if !config.TriggerConditions.Valid {
			eligible = append(eligible, config)
			continue
		}

		// 解析触发条件
		var condition map[string]interface{}
		if err := json.Unmarshal(config.TriggerConditions.JSON, &condition); err != nil {
			continue
		}

		conditionType, ok := condition["type"].(string)
		if !ok {
			continue
		}

		// 检查触发条件
		switch conditionType {
		case "level_range":
			minLevel, _ := condition["min_level"].(float64)
			maxLevel, _ := condition["max_level"].(float64)
			if req.PlayerLevel >= int(minLevel) && req.PlayerLevel <= int(maxLevel) {
				eligible = append(eligible, config)
			}

		case "dungeon_type":
			dungeonTypes, _ := condition["dungeon_types"].([]interface{})
			for _, dt := range dungeonTypes {
				if dtStr, ok := dt.(string); ok && dtStr == req.DungeonType {
					eligible = append(eligible, config)
					break
				}
			}

		default:
			// 未知的触发条件类型,跳过
		}
	}

	return eligible
}

// checkWorldDropLimits 检查世界掉落限制
func (s *ItemDropService) checkWorldDropLimits(
	ctx context.Context,
	tx *sql.Tx,
	config *game_config.WorldDropConfig,
) (bool, error) {
	// 获取统计信息(带锁)
	stats, err := s.worldDropStatsRepo.GetByConfigIDForUpdate(ctx, tx, config.ID)
	if err != nil {
		// 如果统计不存在,创建新的统计记录
		stats = &game_runtime.WorldDropStat{
			WorldDropConfigID: config.ID,
			ItemID:            config.ItemID,
			TotalDropped:      0,
			DailyDropped:      0,
			HourlyDropped:     0,
		}
		if err := s.worldDropStatsRepo.Create(ctx, tx, stats); err != nil {
			return false, xerrors.Wrap(err, xerrors.CodeInternalError, "创建世界掉落统计失败")
		}
	}

	// 检查全局掉落总数限制
	if config.TotalDropLimit.Valid && stats.TotalDropped >= config.TotalDropLimit.Int {
		return false, nil
	}

	// 检查每日掉落数量限制
	if config.DailyDropLimit.Valid && stats.DailyDropped >= config.DailyDropLimit.Int {
		return false, nil
	}

	// 检查每小时掉落数量限制
	if config.HourlyDropLimit.Valid && stats.HourlyDropped >= config.HourlyDropLimit.Int {
		return false, nil
	}

	// 检查掉落间隔
	if config.MinDropInterval.Valid && stats.LastDropAt.Valid {
		elapsed := time.Since(stats.LastDropAt.Time).Seconds()
		if elapsed < float64(config.MinDropInterval.Int) {
			return false, nil
		}
	}

	return true, nil
}

// calculateFinalDropRate 计算最终掉落概率
func (s *ItemDropService) calculateFinalDropRate(
	config *game_config.WorldDropConfig,
	req *CheckWorldDropRequest,
) float64 {
	baseRate, _ := config.BaseDropRate.Float64()

	// 如果没有修正因子,直接返回基础概率
	if !config.DropRateModifiers.Valid {
		return baseRate
	}

	// 解析修正因子
	var modifiers map[string]float64
	if err := json.Unmarshal(config.DropRateModifiers.JSON, &modifiers); err != nil {
		return baseRate
	}

	finalRate := baseRate

	// 应用地城等级加成
	if bonus, exists := modifiers["dungeon_level_bonus"]; exists {
		finalRate += bonus * float64(req.DungeonLevel)
	}

	// 应用队伍人数加成
	if bonus, exists := modifiers["team_size_bonus"]; exists {
		finalRate += bonus * float64(req.TeamSize)
	}

	// 应用首杀加成
	if req.IsFirstKill {
		if bonus, exists := modifiers["first_kill_bonus"]; exists {
			finalRate += bonus
		}
	}

	// 确保概率在0-1之间
	if finalRate < 0 {
		finalRate = 0
	}
	if finalRate > 1 {
		finalRate = 1
	}

	return finalRate
}

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
}

