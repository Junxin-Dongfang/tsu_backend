package dto

// RoomSequenceItem 房间序列项
type RoomSequenceItem struct {
	RoomID            string                 `json:"room_id" binding:"required"`
	Sort              int                    `json:"sort" binding:"required,min=1"`
	ConditionalSkip   map[string]interface{} `json:"conditional_skip,omitempty"`
	ConditionalReturn map[string]interface{} `json:"conditional_return,omitempty"`
}

// CreateDungeonRequest 创建地城请求
type CreateDungeonRequest struct {
	DungeonCode        string             `json:"dungeon_code" binding:"required,max=50"`
	DungeonName        string             `json:"dungeon_name" binding:"required,max=100"`
	MinLevel           int16              `json:"min_level" binding:"required,min=1"`
	MaxLevel           int16              `json:"max_level" binding:"required,min=1"`
	Description        *string            `json:"description"`
	IsTimeLimited      bool               `json:"is_time_limited"`
	TimeLimitStart     *string            `json:"time_limit_start"`
	TimeLimitEnd       *string            `json:"time_limit_end"`
	RequiresAttempts   bool               `json:"requires_attempts"`
	MaxAttemptsPerDay  *int16             `json:"max_attempts_per_day"`
	RoomSequence       []RoomSequenceItem `json:"room_sequence" binding:"required,min=1"`
	IsActive           bool               `json:"is_active"`
}

// UpdateDungeonRequest 更新地城请求
type UpdateDungeonRequest struct {
	DungeonName        *string            `json:"dungeon_name" binding:"omitempty,max=100"`
	MinLevel           *int16             `json:"min_level" binding:"omitempty,min=1"`
	MaxLevel           *int16             `json:"max_level" binding:"omitempty,min=1"`
	Description        *string            `json:"description"`
	IsTimeLimited      *bool              `json:"is_time_limited"`
	TimeLimitStart     *string            `json:"time_limit_start"`
	TimeLimitEnd       *string            `json:"time_limit_end"`
	RequiresAttempts   *bool              `json:"requires_attempts"`
	MaxAttemptsPerDay  *int16             `json:"max_attempts_per_day"`
	RoomSequence       []RoomSequenceItem `json:"room_sequence"`
	IsActive           *bool              `json:"is_active"`
}

// DungeonResponse 地城响应
type DungeonResponse struct {
	ID                 string             `json:"id"`
	DungeonCode        string             `json:"dungeon_code"`
	DungeonName        string             `json:"dungeon_name"`
	MinLevel           int16              `json:"min_level"`
	MaxLevel           int16              `json:"max_level"`
	Description        *string            `json:"description"`
	IsTimeLimited      bool               `json:"is_time_limited"`
	TimeLimitStart     *string            `json:"time_limit_start"`
	TimeLimitEnd       *string            `json:"time_limit_end"`
	RequiresAttempts   bool               `json:"requires_attempts"`
	MaxAttemptsPerDay  *int16             `json:"max_attempts_per_day"`
	RoomSequence       []RoomSequenceItem `json:"room_sequence"`
	IsActive           bool               `json:"is_active"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
}

// DungeonListResponse 地城列表响应
type DungeonListResponse struct {
	List     []DungeonResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	RoomCode       string                 `json:"room_code" binding:"required,max=50"`
	RoomName       *string                `json:"room_name" binding:"omitempty,max=100"`
	RoomType       string                 `json:"room_type" binding:"required,oneof=battle event treasure rest"`
	TriggerID      *string                `json:"trigger_id" binding:"omitempty,max=50"`
	OpenConditions map[string]interface{} `json:"open_conditions"`
	IsActive       bool                   `json:"is_active"`
}

// UpdateRoomRequest 更新房间请求
type UpdateRoomRequest struct {
	RoomName       *string                `json:"room_name" binding:"omitempty,max=100"`
	RoomType       *string                `json:"room_type" binding:"omitempty,oneof=battle event treasure rest"`
	TriggerID      *string                `json:"trigger_id" binding:"omitempty,max=50"`
	OpenConditions map[string]interface{} `json:"open_conditions"`
	IsActive       *bool                  `json:"is_active"`
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
	MonsterCode   string `json:"monster_code" binding:"required"`
	Position      int    `json:"position" binding:"required,min=1,max=21"`
	LevelOverride *int16 `json:"level_override"`
}

// GlobalBuffItem 全程Buff项
type GlobalBuffItem struct {
	BuffCode string `json:"buff_code" binding:"required"`
	Target   string `json:"target" binding:"required"`
}

// CreateBattleRequest 创建战斗配置请求
type CreateBattleRequest struct {
	BattleCode         string                 `json:"battle_code" binding:"required,max=50"`
	LocationConfig     map[string]interface{} `json:"location_config"`
	GlobalBuffs        []GlobalBuffItem       `json:"global_buffs"`
	MonsterSetup       []MonsterSetupItem     `json:"monster_setup" binding:"required,min=1"`
	BattleStartDesc    *string                `json:"battle_start_desc"`
	BattleSuccessDesc  *string                `json:"battle_success_desc"`
	BattleFailureDesc  *string                `json:"battle_failure_desc"`
	IsActive           bool                   `json:"is_active"`
}

// UpdateBattleRequest 更新战斗配置请求
type UpdateBattleRequest struct {
	LocationConfig     map[string]interface{} `json:"location_config"`
	GlobalBuffs        []GlobalBuffItem       `json:"global_buffs"`
	MonsterSetup       []MonsterSetupItem     `json:"monster_setup"`
	BattleStartDesc    *string                `json:"battle_start_desc"`
	BattleSuccessDesc  *string                `json:"battle_success_desc"`
	BattleFailureDesc  *string                `json:"battle_failure_desc"`
	IsActive           *bool                  `json:"is_active"`
}

// BattleResponse 战斗配置响应
type BattleResponse struct {
	ID                 string                 `json:"id"`
	BattleCode         string                 `json:"battle_code"`
	LocationConfig     map[string]interface{} `json:"location_config"`
	GlobalBuffs        []GlobalBuffItem       `json:"global_buffs"`
	MonsterSetup       []MonsterSetupItem     `json:"monster_setup"`
	BattleStartDesc    *string                `json:"battle_start_desc"`
	BattleSuccessDesc  *string                `json:"battle_success_desc"`
	BattleFailureDesc  *string                `json:"battle_failure_desc"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

// ApplyEffectItem 施加效果项
type ApplyEffectItem struct {
	BuffCode    string                 `json:"buff_code" binding:"required"`
	BuffParams  map[string]interface{} `json:"buff_params"`
	CasterLevel int                    `json:"caster_level" binding:"required,min=1"`
	Target      string                 `json:"target" binding:"required"`
}

// GuaranteedItem 保底物品项
type GuaranteedItem struct {
	ItemCode string `json:"item_code" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

// DropConfig 掉落配置
type DropConfig struct {
	DropPoolID       *string          `json:"drop_pool_id"`
	GuaranteedItems  []GuaranteedItem `json:"guaranteed_items"`
}

// CreateEventRequest 创建事件配置请求
type CreateEventRequest struct {
	EventCode        string            `json:"event_code" binding:"required,max=50"`
	EventDescription *string           `json:"event_description"`
	ApplyEffects     []ApplyEffectItem `json:"apply_effects"`
	DropConfig       DropConfig        `json:"drop_config"`
	RewardExp        int               `json:"reward_exp" binding:"min=0"`
	EventEndDesc     *string           `json:"event_end_desc"`
	IsActive         bool              `json:"is_active"`
}

// UpdateEventRequest 更新事件配置请求
type UpdateEventRequest struct {
	EventDescription *string           `json:"event_description"`
	ApplyEffects     []ApplyEffectItem `json:"apply_effects"`
	DropConfig       *DropConfig       `json:"drop_config"`
	RewardExp        *int              `json:"reward_exp" binding:"omitempty,min=0"`
	EventEndDesc     *string           `json:"event_end_desc"`
	IsActive         *bool             `json:"is_active"`
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

