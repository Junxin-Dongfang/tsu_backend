package dto

// RoomSequenceItem 房间序列项
type RoomSequenceItem struct {
	RoomID            string                 `json:"room_id" binding:"required" example:"c5e6fb80-0d44-4e41-85eb-3e52fa45209d"`                                 // 关联的房间ID(UUID)，按顺序排列
	Sort              int                    `json:"sort" binding:"required,min=1" example:"1"`                                                                 // 房间顺序，从1开始
	ConditionalSkip   map[string]interface{} `json:"conditional_skip,omitempty" swaggertype:"object" example:"{\"condition\":\"has_key\",\"jump_to\":3}"`       // 满足条件时跳过当前房间的规则
	ConditionalReturn map[string]interface{} `json:"conditional_return,omitempty" swaggertype:"object" example:"{\"condition\":\"fail_boss\",\"return_to\":1}"` // 失败后返回到某个房间的规则
}

// CreateDungeonRequest 创建地城请求
type CreateDungeonRequest struct {
	DungeonCode       string             `json:"dungeon_code" binding:"required,max=50" example:"dungeon_forest_beginner"` // 地城唯一代码，推荐使用蛇形或驼峰命名
	DungeonName       string             `json:"dungeon_name" binding:"required,max=100" example:"初心者森林"`                  // 展示给运营或玩家的地城名称
	MinLevel          int16              `json:"min_level" binding:"required,min=1" example:"5"`                           // 进入地城的最低英雄等级
	MaxLevel          int16              `json:"max_level" binding:"required,min=1" example:"20"`                          // 进入地城的最高英雄等级
	Description       *string            `json:"description" example:"探索密林，击败驻守的兽人首领"`                                     // 地城介绍或策划备注
	IsTimeLimited     bool               `json:"is_time_limited" example:"false"`                                          // 是否限定开放时间
	TimeLimitStart    *string            `json:"time_limit_start" example:"2025-05-01T00:00:00+08:00"`                     // 限时开始时间(ISO8601)，仅在限时地城使用
	TimeLimitEnd      *string            `json:"time_limit_end" example:"2025-05-07T23:59:59+08:00"`                       // 限时结束时间(ISO8601)
	RequiresAttempts  bool               `json:"requires_attempts" example:"true"`                                         // 是否需要扣除挑战次数
	MaxAttemptsPerDay *int16             `json:"max_attempts_per_day" example:"3"`                                         // 每日最大挑战次数，requires_attempts=true 时必填
	RoomSequence      []RoomSequenceItem `json:"room_sequence" binding:"required,min=1"`                                   // 房间流程配置，至少包含1个房间
	IsActive          bool               `json:"is_active" example:"true"`                                                 // 是否启用
}

// UpdateDungeonRequest 更新地城请求
type UpdateDungeonRequest struct {
	DungeonName       *string            `json:"dungeon_name" binding:"omitempty,max=100" example:"初心者森林·困难"` // 新的地城名称
	MinLevel          *int16             `json:"min_level" binding:"omitempty,min=1" example:"10"`            // 更新后的最低等级
	MaxLevel          *int16             `json:"max_level" binding:"omitempty,min=1" example:"25"`            // 更新后的最高等级
	Description       *string            `json:"description" example:"增加精英怪和陷阱"`                              // 描述更新
	IsTimeLimited     *bool              `json:"is_time_limited" example:"true"`                              // 是否切换为限时地城
	TimeLimitStart    *string            `json:"time_limit_start" example:"2025-05-10T00:00:00+08:00"`        // 限时开始时间
	TimeLimitEnd      *string            `json:"time_limit_end" example:"2025-05-15T23:59:59+08:00"`          // 限时结束时间
	RequiresAttempts  *bool              `json:"requires_attempts" example:"true"`                            // 是否需要挑战次数
	MaxAttemptsPerDay *int16             `json:"max_attempts_per_day" example:"5"`                            // 新的每日挑战次数
	RoomSequence      []RoomSequenceItem `json:"room_sequence"`                                               // 新的房间流程配置
	IsActive          *bool              `json:"is_active" example:"false"`                                   // 是否下线地城
}

// DungeonResponse 地城响应
type DungeonResponse struct {
	ID                string             `json:"id"`
	DungeonCode       string             `json:"dungeon_code"`
	DungeonName       string             `json:"dungeon_name"`
	MinLevel          int16              `json:"min_level"`
	MaxLevel          int16              `json:"max_level"`
	Description       *string            `json:"description"`
	IsTimeLimited     bool               `json:"is_time_limited"`
	TimeLimitStart    *string            `json:"time_limit_start"`
	TimeLimitEnd      *string            `json:"time_limit_end"`
	RequiresAttempts  bool               `json:"requires_attempts"`
	MaxAttemptsPerDay *int16             `json:"max_attempts_per_day"`
	RoomSequence      []RoomSequenceItem `json:"room_sequence"`
	IsActive          bool               `json:"is_active"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}

// DungeonListResponse 地城列表响应
type DungeonListResponse struct {
	List      []DungeonResponse `json:"list"`
	Total     int64             `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
	OrderBy   string            `json:"order_by"`
	OrderDesc bool              `json:"order_desc"`
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	RoomCode       string                 `json:"room_code" binding:"required,max=50" example:"room_forest_start"`                                          // 房间唯一代码
	RoomName       *string                `json:"room_name" binding:"omitempty,max=100" example:"密林入口"`                                                     // 展示名称
	RoomType       string                 `json:"room_type" binding:"required,oneof=battle event treasure rest" example:"battle"`                           // 房间类型
	TriggerID      *string                `json:"trigger_id" binding:"omitempty,max=50" example:"trigger_story_001"`                                        // 触发器ID
	OpenConditions map[string]interface{} `json:"open_conditions" swaggertype:"object" example:"{\"type\":\"require_item\",\"item_code\":\"key_ancient\"}"` // 进入条件
	IsActive       bool                   `json:"is_active" example:"true"`                                                                                 // 是否启用
}

// UpdateRoomRequest 更新房间请求
type UpdateRoomRequest struct {
	RoomName       *string                `json:"room_name" binding:"omitempty,max=100" example:"密林第二层"`                                                   // 新名称
	RoomType       *string                `json:"room_type" binding:"omitempty,oneof=battle event treasure rest" example:"treasure"`                       // 新类型
	TriggerID      *string                `json:"trigger_id" binding:"omitempty,max=50" example:"trigger_story_002"`                                       // 更新触发器
	OpenConditions map[string]interface{} `json:"open_conditions" swaggertype:"object" example:"{\"type\":\"require_flag\",\"flag\":\"unlock_treasure\"}"` // 更新后的进入条件
	IsActive       *bool                  `json:"is_active" example:"false"`                                                                               // 是否下线
}

// RoomResponse 房间响应
type RoomResponse struct {
	ID             string                 `json:"id"`
	RoomCode       string                 `json:"room_code"`
	RoomName       *string                `json:"room_name"`
	RoomType       string                 `json:"room_type"`
	TriggerID      *string                `json:"trigger_id"`
	OpenConditions map[string]interface{} `json:"open_conditions"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// RoomListResponse 房间列表响应
type RoomListResponse struct {
	List     []RoomResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// MonsterSetupItem 怪物配置项
type MonsterSetupItem struct {
	MonsterCode   string `json:"monster_code" binding:"required" example:"MONSTER_GOBLIN_BOSS"` // 怪物配置代码
	Position      int    `json:"position" binding:"required,min=1,max=21" example:"5"`          // 阵位1-21
	LevelOverride *int16 `json:"level_override" example:"25"`                                   // 覆盖默认等级
}

// GlobalBuffItem 全程Buff项
type GlobalBuffItem struct {
	BuffCode string `json:"buff_code" binding:"required" example:"BUFF_BLOODLUST"` // Buff代码
	Target   string `json:"target" binding:"required" example:"allies"`            // buff作用对象(allies/enemies/all)
}

// CreateBattleRequest 创建战斗配置请求
type CreateBattleRequest struct {
	BattleCode        string                 `json:"battle_code" binding:"required,max=50" example:"battle_forest_boss"` // 战斗配置代码
	LocationConfig    map[string]interface{} `json:"location_config" swaggertype:"object"`                               // 场景/环境配置(JSON对象)
	GlobalBuffs       []GlobalBuffItem       `json:"global_buffs"`                                                       // 全程 Buff 列表
	MonsterSetup      []MonsterSetupItem     `json:"monster_setup" binding:"required,min=1"`                             // 怪物阵容
	BattleStartDesc   *string                `json:"battle_start_desc" example:"跨过藤蔓，首领现身"`                              // 开场描述
	BattleSuccessDesc *string                `json:"battle_success_desc" example:"成功清理了密林的威胁"`                           // 胜利描述
	BattleFailureDesc *string                `json:"battle_failure_desc" example:"队伍被击退，需要重新集结"`                         // 失败描述
	IsActive          bool                   `json:"is_active" example:"true"`                                           // 是否启用
}

// UpdateBattleRequest 更新战斗配置请求
type UpdateBattleRequest struct {
	LocationConfig    map[string]interface{} `json:"location_config" swaggertype:"object"`   // 新的场景配置(JSON对象)
	GlobalBuffs       []GlobalBuffItem       `json:"global_buffs"`                           // 更新的Buff列表
	MonsterSetup      []MonsterSetupItem     `json:"monster_setup"`                          // 更新的怪物阵容
	BattleStartDesc   *string                `json:"battle_start_desc" example:"首领愤怒地咆哮"`    // 新的开场描述
	BattleSuccessDesc *string                `json:"battle_success_desc" example:"净化了腐化的神木"` // 新的胜利描述
	BattleFailureDesc *string                `json:"battle_failure_desc" example:"被藤蔓困住，撤退"` // 新的失败描述
	IsActive          *bool                  `json:"is_active" example:"false"`              // 是否下线战斗
}

// BattleResponse 战斗配置响应
type BattleResponse struct {
	ID                string                 `json:"id"`
	BattleCode        string                 `json:"battle_code"`
	LocationConfig    map[string]interface{} `json:"location_config"`
	GlobalBuffs       []GlobalBuffItem       `json:"global_buffs"`
	MonsterSetup      []MonsterSetupItem     `json:"monster_setup"`
	BattleStartDesc   *string                `json:"battle_start_desc"`
	BattleSuccessDesc *string                `json:"battle_success_desc"`
	BattleFailureDesc *string                `json:"battle_failure_desc"`
	IsActive          bool                   `json:"is_active"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
}

// ApplyEffectItem 施加效果项
type ApplyEffectItem struct {
	BuffCode    string                 `json:"buff_code" binding:"required" example:"BUFF_GIVE_REWARD"` // 事件中施加的效果或Buff代码
	BuffParams  map[string]interface{} `json:"buff_params" swaggertype:"object"`                        // 效果参数(JSON对象)
	CasterLevel int                    `json:"caster_level" binding:"required,min=1" example:"20"`      // 效果生效的施法者等级
	Target      string                 `json:"target" binding:"required" example:"team"`                // 作用对象(team/player)
}

// GuaranteedItem 保底物品项
type GuaranteedItem struct {
	ItemCode string `json:"item_code" binding:"required" example:"ITEM_KEY_FRAGMENT"` // 保底奖励物品代码
	Quantity int    `json:"quantity" binding:"required,min=1" example:"1"`            // 数量
}

// DropConfig 掉落配置
type DropConfig struct {
	DropPoolID      *string          `json:"drop_pool_id" example:"e10e8400-e29b-41d4-a716-446655440000"` // 可选，关联掉落池
	GuaranteedItems []GuaranteedItem `json:"guaranteed_items"`                                            // 保底奖励列表
}

// CreateEventRequest 创建事件配置请求
type CreateEventRequest struct {
	EventCode        string            `json:"event_code" binding:"required,max=50" example:"event_secret_treasure"` // 事件代码
	EventDescription *string           `json:"event_description" example:"触发隐藏密室，获得额外奖励"`                            // 事件描述
	ApplyEffects     []ApplyEffectItem `json:"apply_effects"`                                                        // 效果列表
	DropConfig       DropConfig        `json:"drop_config"`                                                          // 掉落配置
	RewardExp        int               `json:"reward_exp" binding:"min=0" example:"200"`                             // 事件奖励经验
	EventEndDesc     *string           `json:"event_end_desc" example:"密室关闭，需要返回主线"`                                 // 事件结语
	IsActive         bool              `json:"is_active" example:"true"`                                             // 是否启用
}

// UpdateEventRequest 更新事件配置请求
type UpdateEventRequest struct {
	EventDescription *string           `json:"event_description" example:"密室奖励升级"`                 // 更新后的描述
	ApplyEffects     []ApplyEffectItem `json:"apply_effects"`                                      // 新的效果
	DropConfig       *DropConfig       `json:"drop_config"`                                        // 新的掉落
	RewardExp        *int              `json:"reward_exp" binding:"omitempty,min=0" example:"350"` // 新的经验
	EventEndDesc     *string           `json:"event_end_desc" example:"限时开启完毕"`                    // 新的结语
	IsActive         *bool             `json:"is_active" example:"false"`                          // 是否下线事件
}

// EventResponse 事件配置响应
type EventResponse struct {
	ID               string            `json:"id"`
	EventCode        string            `json:"event_code"`
	EventDescription *string           `json:"event_description"`
	ApplyEffects     []ApplyEffectItem `json:"apply_effects"`
	DropConfig       DropConfig        `json:"drop_config"`
	RewardExp        int               `json:"reward_exp"`
	EventEndDesc     *string           `json:"event_end_desc"`
	IsActive         bool              `json:"is_active"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}
