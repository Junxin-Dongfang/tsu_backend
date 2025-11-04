package handler

import (
	"database/sql"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// MonsterHandler 怪物 HTTP 处理器
type MonsterHandler struct {
	service    *service.MonsterService
	respWriter response.Writer
}

// NewMonsterHandler 创建怪物处理器
func NewMonsterHandler(db *sql.DB, respWriter response.Writer) *MonsterHandler {
	return &MonsterHandler{
		service:    service.NewMonsterService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateMonsterRequest 创建怪物请求
type CreateMonsterRequest struct {
	// 基础信息
	MonsterCode  string `json:"monster_code" validate:"required,max=64" example:"MAD_DOG"`   // 怪物代码，唯一标识，必填
	MonsterName  string `json:"monster_name" validate:"required,max=128" example:"疯狗"`       // 怪物名称，必填
	MonsterLevel int16  `json:"monster_level" validate:"required,min=1,max=100" example:"2"` // 怪物等级，范围 1-100，必填
	Description  string `json:"description" example:"一只失去理智的野狗"`                             // 怪物描述，可选

	// 生命与法力
	MaxHP      int `json:"max_hp" validate:"required,min=1" example:"20"` // 最大生命值，必须大于0，必填
	HPRecovery int `json:"hp_recovery" example:"2"`                       // 生命恢复，每回合恢复的生命值，默认0
	MaxMP      int `json:"max_mp" example:"0"`                            // 最大法力值，默认0
	MPRecovery int `json:"mp_recovery" example:"0"`                       // 法力恢复，每回合恢复的法力值，默认0

	// 基础属性（0-99）
	BaseSTR int16 `json:"base_str" example:"1"` // 力量，影响物理攻击和负重，范围 0-99
	BaseAgi int16 `json:"base_agi" example:"3"` // 敏捷，影响闪避和先攻，范围 0-99
	BaseVit int16 `json:"base_vit" example:"1"` // 体质，影响生命值和体质豁免，范围 0-99
	BaseWLP int16 `json:"base_wlp" example:"0"` // 意志，影响魔法抗性和精神豁免，范围 0-99
	BaseInt int16 `json:"base_int" example:"1"` // 智力，影响魔法攻击和学习能力，范围 0-99
	BaseWis int16 `json:"base_wis" example:"0"` // 感知，影响洞察和精神豁免，范围 0-99
	BaseCha int16 `json:"base_cha" example:"0"` // 魅力，影响社交和领导力，范围 0-99
	// 战斗属性类型代码（引用 hero_attribute_type 表，不指定则使用默认值）
	AccuracyAttributeCode   string `json:"accuracy_attribute_code" example:"ACCURACY"`     // 精准属性类型代码，决定命中率计算公式，默认 ACCURACY
	DodgeAttributeCode      string `json:"dodge_attribute_code" example:"DODGE"`           // 闪避属性类型代码，决定闪避率计算公式，默认 DODGE
	InitiativeAttributeCode string `json:"initiative_attribute_code" example:"INITIATIVE"` // 先攻属性类型代码，决定行动顺序计算公式，默认 INITIATIVE

	// 豁免属性类型代码（引用 hero_attribute_type 表，不指定则使用默认值）
	BodyResistAttributeCode        string `json:"body_resist_attribute_code" example:"BODY_RESIST"`               // 体质豁免属性类型代码，抵抗毒素、疾病等，默认 BODY_RESIST
	MagicResistAttributeCode       string `json:"magic_resist_attribute_code" example:"MAGIC_RESIST"`             // 魔法豁免属性类型代码，抵抗魔法效果，默认 MAGIC_RESIST
	MentalResistAttributeCode      string `json:"mental_resist_attribute_code" example:"MENTAL_RESIST"`           // 精神豁免属性类型代码，抵抗精神控制、幻觉等，默认 MENTAL_RESIST
	EnvironmentResistAttributeCode string `json:"environment_resist_attribute_code" example:"ENVIRONMENT_RESIST"` // 环境豁免属性类型代码，抵抗极端环境，默认 ENVIRONMENT_RESIST

	// JSON 配置
	DamageResistances map[string]interface{} `json:"damage_resistances" swaggertype:"object,string"` // 伤害抗性，JSON 对象，如 {"SHADOW_DR": 1, "FIRE_RESIST": 0.5}
	PassiveBuffs      []interface{}          `json:"passive_buffs" swaggertype:"array,string"`       // 被动 Buff 列表，JSON 数组

	// 掉落配置
	DropGoldMin int `json:"drop_gold_min" example:"10"` // 最小掉落金币数，默认0
	DropGoldMax int `json:"drop_gold_max" example:"30"` // 最大掉落金币数，默认0
	DropExp     int `json:"drop_exp" example:"20"`      // 掉落经验值，击败怪物获得的经验，默认0

	// 显示配置
	IconURL      string `json:"icon_url" example:"/assets/monsters/mad_dog.png"` // 图标 URL，怪物头像图片路径
	ModelURL     string `json:"model_url" example:"/assets/models/mad_dog.fbx"`  // 模型 URL，3D 模型文件路径
	IsActive     bool   `json:"is_active" example:"true"`                        // 是否启用，false 表示禁用该怪物，默认 true
	DisplayOrder int    `json:"display_order" example:"0"`                       // 显示顺序，用于排序，数字越小越靠前，默认0
}

// UpdateMonsterRequest 更新怪物请求（所有字段可选，只传需要更新的字段）
type UpdateMonsterRequest struct {
	// 基础信息
	MonsterCode  string `json:"monster_code" example:"MAD_DOG"`  // 怪物代码，唯一标识
	MonsterName  string `json:"monster_name" example:"疯狗"`       // 怪物名称
	MonsterLevel int16  `json:"monster_level" example:"2"`       // 怪物等级，范围 1-100
	Description  string `json:"description" example:"一只失去理智的野狗"` // 怪物描述

	// 生命与法力
	MaxHP      int `json:"max_hp" example:"20"`     // 最大生命值，必须大于0
	HPRecovery int `json:"hp_recovery" example:"2"` // 生命恢复，每回合恢复的生命值
	MaxMP      int `json:"max_mp" example:"0"`      // 最大法力值
	MPRecovery int `json:"mp_recovery" example:"0"` // 法力恢复，每回合恢复的法力值

	// 基础属性（0-99）
	BaseSTR int16 `json:"base_str" example:"1"` // 力量，影响物理攻击和负重
	BaseAgi int16 `json:"base_agi" example:"3"` // 敏捷，影响闪避和先攻
	BaseVit int16 `json:"base_vit" example:"1"` // 体质，影响生命值和体质豁免
	BaseWLP int16 `json:"base_wlp" example:"0"` // 意志，影响魔法抗性和精神豁免
	BaseInt int16 `json:"base_int" example:"1"` // 智力，影响魔法攻击和学习能力
	BaseWis int16 `json:"base_wis" example:"0"` // 感知，影响洞察和精神豁免
	BaseCha int16 `json:"base_cha" example:"0"` // 魅力，影响社交和领导力
	// 战斗属性类型代码（引用 hero_attribute_type 表）
	AccuracyAttributeCode   string `json:"accuracy_attribute_code" example:"ACCURACY"`     // 精准属性类型代码，决定命中率计算公式
	DodgeAttributeCode      string `json:"dodge_attribute_code" example:"DODGE"`           // 闪避属性类型代码，决定闪避率计算公式
	InitiativeAttributeCode string `json:"initiative_attribute_code" example:"INITIATIVE"` // 先攻属性类型代码，决定行动顺序计算公式

	// 豁免属性类型代码（引用 hero_attribute_type 表）
	BodyResistAttributeCode        string `json:"body_resist_attribute_code" example:"BODY_RESIST"`               // 体质豁免属性类型代码，抵抗毒素、疾病等
	MagicResistAttributeCode       string `json:"magic_resist_attribute_code" example:"MAGIC_RESIST"`             // 魔法豁免属性类型代码，抵抗魔法效果
	MentalResistAttributeCode      string `json:"mental_resist_attribute_code" example:"MENTAL_RESIST"`           // 精神豁免属性类型代码，抵抗精神控制、幻觉等
	EnvironmentResistAttributeCode string `json:"environment_resist_attribute_code" example:"ENVIRONMENT_RESIST"` // 环境豁免属性类型代码，抵抗极端环境

	// JSON 配置
	DamageResistances map[string]interface{} `json:"damage_resistances" swaggertype:"object,string"` // 伤害抗性，JSON 对象
	PassiveBuffs      []interface{}          `json:"passive_buffs" swaggertype:"array,string"`       // 被动 Buff 列表，JSON 数组

	// 掉落配置
	DropGoldMin int `json:"drop_gold_min" example:"10"` // 最小掉落金币数
	DropGoldMax int `json:"drop_gold_max" example:"30"` // 最大掉落金币数
	DropExp     int `json:"drop_exp" example:"20"`      // 掉落经验值，击败怪物获得的经验

	// 显示配置
	IconURL      string `json:"icon_url" example:"/assets/monsters/mad_dog.png"` // 图标 URL，怪物头像图片路径
	ModelURL     string `json:"model_url" example:"/assets/models/mad_dog.fbx"`  // 模型 URL，3D 模型文件路径
	IsActive     bool   `json:"is_active" example:"true"`                        // 是否启用，false 表示禁用该怪物
	DisplayOrder int    `json:"display_order" example:"0"`                       // 显示顺序，用于排序，数字越小越靠前
}

// MonsterInfo 怪物信息响应
type MonsterInfo struct {
	// 基础信息
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // 怪物 ID，UUID 格式
	MonsterCode  string `json:"monster_code" example:"MAD_DOG"`                    // 怪物代码，唯一标识
	MonsterName  string `json:"monster_name" example:"疯狗"`                         // 怪物名称
	MonsterLevel int16  `json:"monster_level" example:"2"`                         // 怪物等级，范围 1-100
	Description  string `json:"description" example:"一只失去理智的野狗"`                   // 怪物描述

	// 生命与法力
	MaxHP      int `json:"max_hp" example:"20"`     // 最大生命值
	HPRecovery int `json:"hp_recovery" example:"2"` // 生命恢复，每回合恢复的生命值
	MaxMP      int `json:"max_mp" example:"0"`      // 最大法力值
	MPRecovery int `json:"mp_recovery" example:"0"` // 法力恢复，每回合恢复的法力值

	// 基础属性（0-99）
	BaseSTR int16 `json:"base_str" example:"1"` // 力量，影响物理攻击和负重
	BaseAgi int16 `json:"base_agi" example:"3"` // 敏捷，影响闪避和先攻
	BaseVit int16 `json:"base_vit" example:"1"` // 体质，影响生命值和体质豁免
	BaseWLP int16 `json:"base_wlp" example:"0"` // 意志，影响魔法抗性和精神豁免
	BaseInt int16 `json:"base_int" example:"1"` // 智力，影响魔法攻击和学习能力
	BaseWis int16 `json:"base_wis" example:"0"` // 感知，影响洞察和精神豁免
	BaseCha int16 `json:"base_cha" example:"0"` // 魅力，影响社交和领导力
	// 战斗属性类型代码（引用 hero_attribute_type 表）
	AccuracyAttributeCode   string `json:"accuracy_attribute_code" example:"ACCURACY"`     // 精准属性类型代码，决定命中率计算公式
	DodgeAttributeCode      string `json:"dodge_attribute_code" example:"DODGE"`           // 闪避属性类型代码，决定闪避率计算公式
	InitiativeAttributeCode string `json:"initiative_attribute_code" example:"INITIATIVE"` // 先攻属性类型代码，决定行动顺序计算公式

	// 豁免属性类型代码（引用 hero_attribute_type 表）
	BodyResistAttributeCode        string `json:"body_resist_attribute_code" example:"BODY_RESIST"`               // 体质豁免属性类型代码，抵抗毒素、疾病等
	MagicResistAttributeCode       string `json:"magic_resist_attribute_code" example:"MAGIC_RESIST"`             // 魔法豁免属性类型代码，抵抗魔法效果
	MentalResistAttributeCode      string `json:"mental_resist_attribute_code" example:"MENTAL_RESIST"`           // 精神豁免属性类型代码，抵抗精神控制、幻觉等
	EnvironmentResistAttributeCode string `json:"environment_resist_attribute_code" example:"ENVIRONMENT_RESIST"` // 环境豁免属性类型代码，抵抗极端环境

	// JSON 配置
	DamageResistances map[string]interface{} `json:"damage_resistances" swaggertype:"object,string"` // 伤害抗性，JSON 对象，如 {"SHADOW_DR": 1, "FIRE_RESIST": 0.5}
	PassiveBuffs      []interface{}          `json:"passive_buffs" swaggertype:"array,string"`       // 被动 Buff 列表，JSON 数组

	// 掉落配置
	DropGoldMin int `json:"drop_gold_min" example:"10"` // 最小掉落金币数
	DropGoldMax int `json:"drop_gold_max" example:"30"` // 最大掉落金币数
	DropExp     int `json:"drop_exp" example:"20"`      // 掉落经验值，击败怪物获得的经验

	// 显示配置
	IconURL      string `json:"icon_url" example:"/assets/monsters/mad_dog.png"` // 图标 URL，怪物头像图片路径
	ModelURL     string `json:"model_url" example:"/assets/models/mad_dog.fbx"`  // 模型 URL，3D 模型文件路径
	IsActive     bool   `json:"is_active" example:"true"`                        // 是否启用，false 表示禁用该怪物
	DisplayOrder int    `json:"display_order" example:"0"`                       // 显示顺序，用于排序，数字越小越靠前

	// 时间戳
	CreatedAt int64 `json:"created_at" example:"1633024800"` // 创建时间，Unix 时间戳（秒）
	UpdatedAt int64 `json:"updated_at" example:"1633024800"` // 更新时间，Unix 时间戳（秒）
}

// AddMonsterSkillRequest 添加怪物技能请求
type AddMonsterSkillRequest struct {
	SkillID     string   `json:"skill_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkillLevel  int16    `json:"skill_level" validate:"required,min=1,max=20" example:"2"`
	GainActions []string `json:"gain_actions" example:"[\"MAD_DOG_BITE\"]"`
}

// UpdateMonsterSkillRequest 更新怪物技能请求
type UpdateMonsterSkillRequest struct {
	SkillLevel  int16    `json:"skill_level" validate:"required,min=1,max=20" example:"2"`
	GainActions []string `json:"gain_actions" example:"[\"MAD_DOG_BITE\"]"`
}

// MonsterSkillInfo 怪物技能信息响应
type MonsterSkillInfo struct {
	ID          string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MonsterID   string   `json:"monster_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkillID     string   `json:"skill_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SkillLevel  int16    `json:"skill_level" example:"2"`
	GainActions []string `json:"gain_actions" example:"[\"MAD_DOG_BITE\"]"`
	CreatedAt   int64    `json:"created_at" example:"1633024800"`
	UpdatedAt   int64    `json:"updated_at" example:"1633024800"`
}

// AddMonsterDropRequest 添加怪物掉落请求
type AddMonsterDropRequest struct {
	DropPoolID  string  `json:"drop_pool_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	DropType    string  `json:"drop_type" validate:"required,oneof=team personal" example:"team"`
	DropChance  float64 `json:"drop_chance" validate:"required,gt=0,lte=1" example:"1.0"`
	MinQuantity int     `json:"min_quantity" validate:"required,min=1" example:"1"`
	MaxQuantity int     `json:"max_quantity" validate:"required,min=1" example:"3"`
}

// UpdateMonsterDropRequest 更新怪物掉落请求
type UpdateMonsterDropRequest struct {
	DropType    string  `json:"drop_type" validate:"required,oneof=team personal" example:"team"`
	DropChance  float64 `json:"drop_chance" validate:"required,gt=0,lte=1" example:"1.0"`
	MinQuantity int     `json:"min_quantity" validate:"required,min=1" example:"1"`
	MaxQuantity int     `json:"max_quantity" validate:"required,min=1" example:"3"`
}

// MonsterDropInfo 怪物掉落信息响应
type MonsterDropInfo struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MonsterID   string  `json:"monster_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DropPoolID  string  `json:"drop_pool_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DropType    string  `json:"drop_type" example:"team"`
	DropChance  float64 `json:"drop_chance" example:"1.0"`
	MinQuantity int     `json:"min_quantity" example:"1"`
	MaxQuantity int     `json:"max_quantity" example:"3"`
	CreatedAt   int64   `json:"created_at" example:"1633024800"`
	UpdatedAt   int64   `json:"updated_at" example:"1633024800"`
}

// ==================== HTTP Handlers ====================

// GetMonsters 获取怪物列表
// @Summary 查询怪物配置列表
// @Description 分页查询怪物配置列表，支持多维度筛选、排序。适用于怪物管理界面、怪物选择器等场景。
// @Description
// @Description **筛选条件**:
// @Description - monster_code: 按怪物代码筛选(模糊匹配，如"MAD"可匹配"MAD_DOG")
// @Description   - 用途: 快速查找特定代码的怪物
// @Description - monster_name: 按怪物名称筛选(模糊匹配，如"狗"可匹配"疯狗")
// @Description   - 用途: 按中文名称搜索怪物
// @Description - min_level: 最小等级(1-100，筛选等级≥此值的怪物)
// @Description   - 用途: 查找适合特定等级玩家的怪物
// @Description - max_level: 最大等级(1-100，筛选等级≤此值的怪物)
// @Description   - 用途: 配合min_level实现等级区间筛选
// @Description - is_active: 是否启用(true=仅启用, false=仅禁用, 不传=全部)
// @Description   - 用途: 区分正在使用和已废弃的怪物
// @Description
// @Description **分页参数**:
// @Description - limit: 每页数量(默认10，最大100)
// @Description   - 建议: 列表页用20-50，下拉选择器用10-20
// @Description - offset: 偏移量(默认0，用于翻页)
// @Description   - 计算: offset = (page - 1) * limit
// @Description
// @Description **排序规则**:
// @Description - order_by: 排序字段
// @Description   - monster_level: 按等级排序(适合按难度浏览)
// @Description   - created_at: 按创建时间排序(适合查看最新添加)
// @Description   - updated_at: 按更新时间排序(适合查看最近修改)
// @Description - order_desc: 是否降序
// @Description   - false: 升序(等级从低到高，时间从旧到新)
// @Description   - true: 降序(等级从高到低，时间从新到旧)
// @Description - 默认: 按monster_level升序(从低级到高级)
// @Description
// @Description **返回数据**:
// @Description - list: 怪物配置列表(包含完整的怪物信息)
// @Description - total: 符合条件的总数量(用于计算总页数)
// @Description
// @Description **使用场景示例**:
// @Description - 查找1-10级怪物: ?min_level=1&max_level=10&order_by=monster_level
// @Description - 搜索"狗"类怪物: ?monster_name=狗&limit=20
// @Description - 查看最新添加: ?order_by=created_at&order_desc=true&limit=10
// @Description - 获取所有启用怪物: ?is_active=true&limit=100
// @Tags 怪物配置管理
// @Accept json
// @Produce json
// @Param monster_code query string false "怪物代码(模糊搜索)" example("MAD_DOG")
// @Param monster_name query string false "怪物名称(模糊搜索)" example("疯狗")
// @Param min_level query int false "最小等级(筛选等级≥此值)" minimum(1) maximum(100) example(1)
// @Param max_level query int false "最大等级(筛选等级≤此值)" minimum(1) maximum(100) example(50)
// @Param is_active query bool false "是否启用(true=仅启用, false=仅禁用)" example(true)
// @Param limit query int false "每页数量(建议10-50)" default(10) minimum(1) maximum(100)
// @Param offset query int false "偏移量(翻页用)" default(0) minimum(0)
// @Param order_by query string false "排序字段" Enums(monster_level, created_at, updated_at) default(monster_level)
// @Param order_desc query bool false "是否降序(false=升序)" default(false)
// @Success 200 {object} response.Response{data=object{list=[]MonsterInfo,total=int64}} "查询成功，返回怪物列表和总数"
// @Failure 400 {object} response.Response "参数错误(100400): 等级范围无效、分页参数错误等"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters [get]
// @Security BearerAuth
func (h *MonsterHandler) GetMonsters(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.MonsterQueryParams{
		Limit:  10,
		Offset: 0,
	}

	if monsterCode := c.QueryParam("monster_code"); monsterCode != "" {
		params.MonsterCode = &monsterCode
	}
	if monsterName := c.QueryParam("monster_name"); monsterName != "" {
		params.MonsterName = &monsterName
	}
	if minLevel := c.QueryParam("min_level"); minLevel != "" {
		if level, err := strconv.ParseInt(minLevel, 10, 16); err == nil {
			l := int16(level)
			params.MinLevel = &l
		}
	}
	if maxLevel := c.QueryParam("max_level"); maxLevel != "" {
		if level, err := strconv.ParseInt(maxLevel, 10, 16); err == nil {
			l := int16(level)
			params.MaxLevel = &l
		}
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			params.IsActive = &active
		}
	}
	if limit := c.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			params.Limit = l
		}
	}
	if offset := c.QueryParam("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			params.Offset = o
		}
	}
	if orderBy := c.QueryParam("order_by"); orderBy != "" {
		params.OrderBy = orderBy
	}
	if orderDesc := c.QueryParam("order_desc"); orderDesc != "" {
		if desc, err := strconv.ParseBool(orderDesc); err == nil {
			params.OrderDesc = desc
		}
	}

	// 查询怪物列表
	monsters, total, err := h.service.GetMonsters(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	monsterInfos := make([]MonsterInfo, 0, len(monsters))
	for _, monster := range monsters {
		monsterInfos = append(monsterInfos, h.toMonsterInfo(monster))
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  monsterInfos,
		"total": total,
	})
}

// GetMonster 获取怪物详情
// @Summary 获取怪物配置详情
// @Description 根据ID获取怪物的完整配置信息，包括所有属性、公式、抗性等。
// @Description
// @Description **返回信息包括**:
// @Description - 基础信息: ID、代码、名称、等级、描述
// @Description - 生命法力: 最大值和恢复速度
// @Description - 基础属性: 力量、敏捷、体质、意志、智力、感知、魅力
// @Description - 属性类型代码: 精准、闪避、先攻、各类豁免的属性类型代码（引用 hero_attribute_type 表）
// @Description - 伤害抗性: JSON对象格式的抗性配置
// @Description - 被动效果: JSON数组格式的被动buff
// @Description - 掉落配置: 金币范围和经验值
// @Description - 显示配置: 图标、模型、启用状态、排序
// @Description - 时间戳: 创建时间和更新时间
// @Tags 怪物配置管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=MonsterInfo} "查询成功，返回怪物完整配置"
// @Failure 400 {object} response.Response "参数错误(100400): ID格式错误"
// @Failure 404 {object} response.Response "怪物不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id} [get]
// @Security BearerAuth
func (h *MonsterHandler) GetMonster(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	monster, err := h.service.GetMonsterByID(ctx, monsterID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.toMonsterInfo(monster))
}

// CreateMonster 创建怪物
// @Summary 创建怪物配置
// @Description 创建新的怪物配置，支持完整的属性、公式、抗性和掉落配置。
// @Description
// @Description **基础信息**:
// @Description - monster_code: 怪物代码(唯一标识，大写字母+数字+下划线，如"MAD_DOG")
// @Description - monster_name: 怪物名称(显示名称，如"疯狗")
// @Description - monster_level: 怪物等级(1-100，决定怪物强度和玩家可挑战等级)
// @Description - description: 怪物描述(背景故事或特征说明)
// @Description
// @Description **生命与法力**:
// @Description - max_hp: 最大生命值(必填，>0，怪物的血量上限)
// @Description - hp_recovery: 生命恢复(每回合恢复的HP，0表示不恢复)
// @Description - max_mp: 最大法力值(魔法怪物需要，0表示无法力)
// @Description - mp_recovery: 法力恢复(每回合恢复的MP，0表示不恢复)
// @Description
// @Description **基础属性**(0-99，影响战斗计算):
// @Description - base_str: 力量(影响物理攻击伤害和命中，近战怪物主属性)
// @Description - base_agi: 敏捷(影响闪避、先攻顺序和暴击率，敏捷型怪物主属性)
// @Description - base_vit: 体质(影响生命值上限和体质豁免检定，坦克型怪物主属性)
// @Description - base_wlp: 意志(影响精神抗性和意志豁免检定，抵抗心灵控制)
// @Description - base_int: 智力(影响魔法攻击伤害和魔法命中，法系怪物主属性)
// @Description - base_wis: 感知(影响魔法防御、先攻和感知检定，辅助型怪物主属性)
// @Description - base_cha: 魅力(影响社交互动和某些特殊技能，NPC型怪物主属性)
// @Description
// @Description **战斗属性类型代码**(引用 hero_attribute_type 表，用于动态计算战斗数值):
// @Description - accuracy_attribute_code: 精准属性类型代码(默认"ACCURACY"，决定攻击命中率)
// @Description   - 用途: 引用 hero_attribute_type 表中的属性类型，使用对应的计算公式
// @Description   - 示例: "ACCURACY" 使用标准精准公式，"BOSS_ACCURACY" 使用BOSS专用公式
// @Description - dodge_attribute_code: 闪避属性类型代码(默认"DODGE"，决定被攻击时的闪避率)
// @Description   - 用途: 引用对应的闪避计算公式
// @Description   - 示例: "DODGE" 使用标准闪避公式
// @Description - initiative_attribute_code: 先攻属性类型代码(默认"INITIATIVE"，决定战斗回合顺序)
// @Description   - 用途: 引用对应的先攻计算公式
// @Description   - 示例: "INITIATIVE" 使用标准先攻公式
// @Description - body_resist_attribute_code: 体质豁免属性类型代码(默认"BODY_RESIST"，抵抗毒素、疾病等效果)
// @Description   - 用途: 引用对应的体质豁免计算公式
// @Description   - 示例: "BODY_RESIST" 使用标准体质豁免公式
// @Description - magic_resist_attribute_code: 魔法豁免属性类型代码(默认"MAGIC_RESIST"，抵抗魔法效果)
// @Description   - 用途: 引用对应的魔法豁免计算公式
// @Description   - 示例: "MAGIC_RESIST" 使用标准魔法豁免公式
// @Description - mental_resist_attribute_code: 精神豁免属性类型代码(默认"MENTAL_RESIST"，抵抗心灵控制)
// @Description   - 用途: 引用对应的精神豁免计算公式
// @Description   - 示例: "MENTAL_RESIST" 使用标准精神豁免公式
// @Description - environment_resist_attribute_code: 环境豁免属性类型代码(默认"ENVIRONMENT_RESIST"，抵抗环境伤害)
// @Description   - 用途: 引用对应的环境豁免计算公式
// @Description   - 示例: "ENVIRONMENT_RESIST" 使用标准环境豁免公式
// @Description
// @Description **注意**: 所有属性类型代码都引用 hero_attribute_type 表，如果不指定则使用数据库默认值
// @Description
// @Description **伤害抗性**(JSON对象，特殊伤害类型的抗性):
// @Description - 格式: {"SHADOW_DR": 1, "FIRE_RESIST": 0.5, "HOLY_IMMUNE": 1}
// @Description - DR(Damage Reduction): 固定减免，如"SHADOW_DR": 1 表示暗影伤害减少1点
// @Description - RESIST: 百分比抗性，如"FIRE_RESIST": 0.5 表示火焰伤害减少50%
// @Description - IMMUNE: 完全免疫，如"HOLY_IMMUNE": 1 表示完全免疫神圣伤害
// @Description - 示例: {"PHYSICAL_DR": 2, "FIRE_RESIST": 0.3, "POISON_IMMUNE": 1}
// @Description
// @Description **被动效果**(JSON数组，怪物的被动能力):
// @Description - 格式: [{"buff_id": "REGENERATION", "value": 5}]
// @Description - 示例: [{"buff_id": "ARMOR_BOOST", "value": 10}] 表示护甲提升10点
// @Description
// @Description **掉落配置**:
// @Description - drop_gold_min: 最小金币掉落(击败怪物获得的最少金币)
// @Description - drop_gold_max: 最大金币掉落(击败怪物获得的最多金币)
// @Description - drop_exp: 经验值掉落(击败怪物获得的经验值)
// @Description
// @Description **显示配置**:
// @Description - icon_url: 图标URL(怪物头像图片路径)
// @Description - model_url: 模型URL(3D模型文件路径)
// @Description - is_active: 是否启用(false表示禁用，不会在游戏中出现)
// @Description - display_order: 显示排序(数值越小越靠前)
// @Tags 怪物配置管理
// @Accept json
// @Produce json
// @Param request body CreateMonsterRequest true "创建怪物配置请求"
// @Success 200 {object} response.Response{data=MonsterInfo} "创建成功，返回怪物详情"
// @Failure 400 {object} response.Response "参数错误(100400): monster_code重复、等级超出范围、属性值无效等"
// @Failure 409 {object} response.Response "怪物代码已存在(100409)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters [post]
// @Security BearerAuth
func (h *MonsterHandler) CreateMonster(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateMonsterRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 转换为实体
	monster := h.toMonsterEntity(&req)

	// 创建怪物
	if err := h.service.CreateMonster(ctx, monster); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.toMonsterInfo(monster))
}

// UpdateMonster 更新怪物
// @Summary 更新怪物配置
// @Description 更新怪物配置信息，支持部分字段更新。只需传入要修改的字段，未传入的字段保持不变。
// @Description
// @Description **可更新字段分类**:
// @Description
// @Description **1. 基础信息**:
// @Description - monster_code: 怪物代码(更新时会检查唯一性)
// @Description - monster_name: 怪物名称
// @Description - monster_level: 怪物等级(1-100)
// @Description - description: 怪物描述
// @Description
// @Description **2. 生命与法力**:
// @Description - max_hp: 最大生命值(调整怪物血量)
// @Description - hp_recovery: 生命恢复(调整每回合恢复量)
// @Description - max_mp: 最大法力值(调整法力上限)
// @Description - mp_recovery: 法力恢复(调整每回合恢复量)
// @Description
// @Description **3. 基础属性**(0-99):
// @Description - base_str: 力量(调整物理攻击能力)
// @Description - base_agi: 敏捷(调整闪避和先攻)
// @Description - base_vit: 体质(调整生命和防御)
// @Description - base_wlp: 意志(调整精神抗性)
// @Description - base_int: 智力(调整魔法攻击)
// @Description - base_wis: 感知(调整魔法防御)
// @Description - base_cha: 魅力(调整社交能力)
// @Description
// @Description **4. 战斗公式**(调整战斗计算方式):
// @Description - accuracy_formula: 精准值公式(如"STR*2+AGI")
// @Description - dodge_formula: 闪避值公式(如"AGI*2+WIS")
// @Description - initiative_formula: 先攻值公式(如"AGI*2+WIS")
// @Description - body_resist_formula: 体质豁免公式(如"VIT*2+WLP")
// @Description - magic_resist_formula: 魔法豁免公式(如"WLP*2+WIS")
// @Description - mental_resist_formula: 精神豁免公式(如"WIS*2+WLP")
// @Description - environment_resist_formula: 环境豁免公式(如"VIT*2+WIS")
// @Description
// @Description **5. 抗性配置**(JSON对象):
// @Description - damage_resistances: 伤害抗性配置
// @Description   - 示例: {"PHYSICAL_DR": 2, "FIRE_RESIST": 0.3}
// @Description
// @Description **6. 被动效果**(JSON数组):
// @Description - passive_buffs: 被动buff列表
// @Description   - 示例: [{"buff_id": "REGENERATION", "value": 5}]
// @Description
// @Description **7. 掉落配置**:
// @Description - drop_gold_min: 最小金币掉落
// @Description - drop_gold_max: 最大金币掉落
// @Description - drop_exp: 经验值掉落
// @Description
// @Description **8. 显示配置**:
// @Description - icon_url: 图标URL
// @Description - model_url: 模型URL
// @Description - is_active: 是否启用
// @Description - display_order: 显示排序
// @Description
// @Description **使用示例**:
// @Description - 只调整血量: {"max_hp": 1000}
// @Description - 调整多个属性: {"max_hp": 1000, "base_str": 20, "drop_exp": 500}
// @Description - 修改公式: {"accuracy_formula": "STR*3+AGI*2"}
// @Tags 怪物配置管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param request body UpdateMonsterRequest true "更新怪物配置请求(只传入要修改的字段)"
// @Success 200 {object} response.Response{data=MonsterInfo} "更新成功，返回更新后的怪物详情"
// @Failure 400 {object} response.Response "参数错误(100400): ID格式错误、字段值无效等"
// @Failure 404 {object} response.Response "怪物不存在(100404)"
// @Failure 409 {object} response.Response "怪物代码已被使用(100409)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id} [put]
// @Security BearerAuth
func (h *MonsterHandler) UpdateMonster(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	var req UpdateMonsterRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.MonsterCode != "" {
		updates["monster_code"] = req.MonsterCode
	}
	if req.MonsterName != "" {
		updates["monster_name"] = req.MonsterName
	}
	if req.MonsterLevel > 0 {
		updates["monster_level"] = req.MonsterLevel
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.MaxHP > 0 {
		updates["max_hp"] = req.MaxHP
	}
	if req.HPRecovery >= 0 {
		updates["hp_recovery"] = req.HPRecovery
	}
	if req.MaxMP >= 0 {
		updates["max_mp"] = req.MaxMP
	}
	if req.MPRecovery >= 0 {
		updates["mp_recovery"] = req.MPRecovery
	}
	if req.BaseSTR >= 0 {
		updates["base_str"] = req.BaseSTR
	}
	if req.BaseAgi >= 0 {
		updates["base_agi"] = req.BaseAgi
	}
	if req.BaseVit >= 0 {
		updates["base_vit"] = req.BaseVit
	}
	if req.BaseWLP >= 0 {
		updates["base_wlp"] = req.BaseWLP
	}
	if req.BaseInt >= 0 {
		updates["base_int"] = req.BaseInt
	}
	if req.BaseWis >= 0 {
		updates["base_wis"] = req.BaseWis
	}
	if req.BaseCha >= 0 {
		updates["base_cha"] = req.BaseCha
	}
	// 属性类型代码
	if req.AccuracyAttributeCode != "" {
		updates["accuracy_attribute_code"] = req.AccuracyAttributeCode
	}
	if req.DodgeAttributeCode != "" {
		updates["dodge_attribute_code"] = req.DodgeAttributeCode
	}
	if req.InitiativeAttributeCode != "" {
		updates["initiative_attribute_code"] = req.InitiativeAttributeCode
	}
	if req.BodyResistAttributeCode != "" {
		updates["body_resist_attribute_code"] = req.BodyResistAttributeCode
	}
	if req.MagicResistAttributeCode != "" {
		updates["magic_resist_attribute_code"] = req.MagicResistAttributeCode
	}
	if req.MentalResistAttributeCode != "" {
		updates["mental_resist_attribute_code"] = req.MentalResistAttributeCode
	}
	if req.EnvironmentResistAttributeCode != "" {
		updates["environment_resist_attribute_code"] = req.EnvironmentResistAttributeCode
	}
	if req.DropGoldMin >= 0 {
		updates["drop_gold_min"] = req.DropGoldMin
	}
	if req.DropGoldMax >= 0 {
		updates["drop_gold_max"] = req.DropGoldMax
	}
	if req.DropExp >= 0 {
		updates["drop_exp"] = req.DropExp
	}
	if req.IconURL != "" {
		updates["icon_url"] = req.IconURL
	}
	if req.ModelURL != "" {
		updates["model_url"] = req.ModelURL
	}
	updates["is_active"] = req.IsActive
	if req.DisplayOrder >= 0 {
		updates["display_order"] = req.DisplayOrder
	}

	// 更新怪物
	if err := h.service.UpdateMonster(ctx, monsterID, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 返回更新后的怪物信息
	monster, err := h.service.GetMonsterByID(ctx, monsterID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.toMonsterInfo(monster))
}

// DeleteMonster 删除怪物
// @Summary 删除怪物配置
// @Description 软删除怪物配置，同时级联删除关联的技能和掉落配置。
// @Description
// @Description **删除行为**:
// @Description - 软删除: 设置 deleted_at 字段，数据仍保留在数据库中
// @Description - 级联删除: 自动删除该怪物的所有技能配置和掉落配置
// @Description - 不可恢复: 删除后无法通过 API 恢复，需要数据库操作
// @Description
// @Description **注意事项**:
// @Description - 删除操作不可逆，请谨慎操作
// @Description - 删除后该怪物将不再出现在列表查询中
// @Description - 已删除的怪物代码可以被重新使用
// @Tags 怪物配置管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=object{message=string}} "删除成功"
// @Failure 400 {object} response.Response "参数错误(100400): ID格式错误"
// @Failure 404 {object} response.Response "怪物不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id} [delete]
// @Security BearerAuth
func (h *MonsterHandler) DeleteMonster(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	if err := h.service.DeleteMonster(ctx, monsterID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// ==================== 怪物技能管理 ====================

// GetMonsterSkills 获取怪物技能列表
// @Summary 获取怪物技能列表
// @Description 获取指定怪物的所有技能配置，包括技能等级和获得动作。
// @Description
// @Description **返回信息包括**:
// @Description - skill_id: 技能ID
// @Description - skill_level: 技能等级
// @Description - gain_actions: 获得该技能的动作列表(JSON数组)
// @Description - created_at/updated_at: 时间戳
// @Tags 怪物技能管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=[]MonsterSkillInfo} "查询成功，返回技能列表"
// @Failure 400 {object} response.Response "参数错误(100400): ID格式错误"
// @Failure 404 {object} response.Response "怪物不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/skills [get]
// @Security BearerAuth
func (h *MonsterHandler) GetMonsterSkills(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	skills, err := h.service.GetMonsterSkills(ctx, monsterID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	skillInfos := make([]MonsterSkillInfo, 0, len(skills))
	for _, skill := range skills {
		skillInfos = append(skillInfos, h.toMonsterSkillInfo(skill))
	}

	return response.EchoOK(c, h.respWriter, skillInfos)
}

// AddMonsterSkill 为怪物添加技能
// @Summary 为怪物添加技能
// @Description 为指定怪物添加新的技能配置，包括技能等级和获得动作。怪物可以拥有多个技能，在战斗中使用。
// @Description
// @Description **请求参数详解**:
// @Description - skill_id: 技能ID(必填，UUID格式)
// @Description   - 说明: 必须是已存在的技能配置ID
// @Description   - 用途: 关联到具体的技能定义(技能名称、效果、消耗等)
// @Description   - 示例: "660e8400-e29b-41d4-a716-446655440000"
// @Description - skill_level: 技能等级(必填，≥1，通常1-10)
// @Description   - 说明: 决定技能的威力和效果强度
// @Description   - 用途: 等级越高，技能伤害/治疗/效果越强
// @Description   - 示例: 1级火球术造成10伤害，5级火球术造成50伤害
// @Description - gain_actions: 获得该技能的动作列表(可选，字符串数组)
// @Description   - 说明: 怪物通过哪些动作可以获得/使用这个技能
// @Description   - 用途: 控制技能的触发条件或使用场景
// @Description   - 示例: ["ATTACK", "DEFEND"] 表示攻击或防御时可以使用
// @Description   - 示例: ["RAGE"] 表示进入狂暴状态时获得
// @Description
// @Description **使用场景**:
// @Description - 为近战怪物添加"重击"技能: {"skill_id": "xxx", "skill_level": 3, "gain_actions": ["ATTACK"]}
// @Description - 为法师怪物添加"火球术": {"skill_id": "xxx", "skill_level": 5, "gain_actions": ["CAST_SPELL"]}
// @Description - 为BOSS添加"狂暴"技能: {"skill_id": "xxx", "skill_level": 1, "gain_actions": ["HP_LOW"]}
// @Description
// @Description **注意事项**:
// @Description - 同一个怪物不能重复添加相同的技能(skill_id唯一)
// @Description - 技能ID必须是已存在的技能配置，否则返回404
// @Description - 技能等级影响战斗平衡，建议根据怪物等级合理设置
// @Description - gain_actions为空数组表示该技能始终可用
// @Tags 怪物技能管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param request body AddMonsterSkillRequest true "添加技能配置请求"
// @Success 200 {object} response.Response{data=object{message=string}} "添加成功"
// @Failure 400 {object} response.Response "参数错误(100400): ID格式错误、技能等级无效等"
// @Failure 404 {object} response.Response "怪物或技能不存在(100404)"
// @Failure 409 {object} response.Response "技能已添加(100409)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/skills [post]
// @Security BearerAuth
func (h *MonsterHandler) AddMonsterSkill(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	var req AddMonsterSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := h.service.AddMonsterSkill(ctx, monsterID, req.SkillID, req.SkillLevel, req.GainActions); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// UpdateMonsterSkill 更新怪物技能
// @Summary 更新怪物技能配置
// @Description 更新怪物已有技能的等级和获得动作。
// @Description
// @Description **可更新字段**:
// @Description - skill_level: 技能等级(必填，≥1)
// @Description - gain_actions: 获得该技能的动作列表(可选，字符串数组)
// @Tags 怪物技能管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param skill_id path string true "技能ID(UUID格式)" example("660e8400-e29b-41d4-a716-446655440000")
// @Param request body UpdateMonsterSkillRequest true "更新技能配置请求"
// @Success 200 {object} response.Response{data=object{message=string}} "更新成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 404 {object} response.Response "怪物技能不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/skills/{skill_id} [put]
// @Security BearerAuth
func (h *MonsterHandler) UpdateMonsterSkill(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")
	skillID := c.Param("skill_id")

	var req UpdateMonsterSkillRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := h.service.UpdateMonsterSkill(ctx, monsterID, skillID, req.SkillLevel, req.GainActions); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// RemoveMonsterSkill 移除怪物技能
// @Summary 移除怪物技能
// @Description 从怪物移除指定技能配置。
// @Description
// @Description **注意事项**:
// @Description - 移除后该怪物将无法使用该技能
// @Description - 操作不可逆，请谨慎操作
// @Tags 怪物技能管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param skill_id path string true "技能ID(UUID格式)" example("660e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=object{message=string}} "移除成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 404 {object} response.Response "怪物技能不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/skills/{skill_id} [delete]
// @Security BearerAuth
func (h *MonsterHandler) RemoveMonsterSkill(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")
	skillID := c.Param("skill_id")

	if err := h.service.RemoveMonsterSkill(ctx, monsterID, skillID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// ==================== 怪物掉落管理 ====================

// GetMonsterDrops 获取怪物掉落列表
// @Summary 获取怪物掉落配置列表
// @Description 获取指定怪物的所有掉落配置，包括掉落池、掉落类型、概率和数量。
// @Description
// @Description **返回信息包括**:
// @Description - drop_pool_id: 掉落池ID
// @Description - drop_type: 掉落类型(team=队伍掉落, personal=个人掉落)
// @Description - drop_chance: 掉落概率(0.0-1.0)
// @Description - min_quantity/max_quantity: 掉落数量范围
// @Description - created_at/updated_at: 时间戳
// @Tags 怪物掉落管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=[]MonsterDropInfo} "查询成功，返回掉落配置列表"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 404 {object} response.Response "怪物不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/drops [get]
// @Security BearerAuth
func (h *MonsterHandler) GetMonsterDrops(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	drops, err := h.service.GetMonsterDrops(ctx, monsterID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	dropInfos := make([]MonsterDropInfo, 0, len(drops))
	for _, drop := range drops {
		dropInfos = append(dropInfos, h.toMonsterDropInfo(drop))
	}

	return response.EchoOK(c, h.respWriter, dropInfos)
}

// AddMonsterDrop 为怪物添加掉落配置
// @Summary 为怪物添加掉落配置
// @Description 为指定怪物添加新的掉落配置，关联掉落池并设置掉落规则。怪物可以有多个掉落池，每个掉落池包含不同的物品。
// @Description
// @Description **请求参数详解**:
// @Description - drop_pool_id: 掉落池ID(必填，UUID格式)
// @Description   - 说明: 必须是已存在的掉落池配置ID
// @Description   - 用途: 关联到具体的掉落池(包含多个物品及其掉落权重)
// @Description   - 示例: "770e8400-e29b-41d4-a716-446655440000"
// @Description - drop_type: 掉落类型(必填，team或personal)
// @Description   - team: 队伍掉落，整个队伍共享一次掉落
// @Description     - 用途: 适合稀有物品、任务道具
// @Description     - 示例: BOSS掉落的传说装备，队伍只掉一件
// @Description   - personal: 个人掉落，每个玩家独立掉落
// @Description     - 用途: 适合消耗品、货币、经验
// @Description     - 示例: 每个玩家都能获得金币和药水
// @Description - drop_chance: 掉落概率(必填，0.0-1.0)
// @Description   - 说明: 该掉落池的触发概率
// @Description   - 0.0: 永不掉落(0%)
// @Description   - 0.5: 50%概率掉落
// @Description   - 1.0: 必定掉落(100%)
// @Description   - 示例: 0.8 表示80%概率触发该掉落池
// @Description - min_quantity: 最小掉落数量(必填，≥1)
// @Description   - 说明: 触发掉落时，从掉落池中抽取的最少物品数
// @Description   - 用途: 保证最低掉落数量
// @Description   - 示例: 1 表示至少掉1个物品
// @Description - max_quantity: 最大掉落数量(必填，≥min_quantity)
// @Description   - 说明: 触发掉落时，从掉落池中抽取的最多物品数
// @Description   - 用途: 限制最高掉落数量
// @Description   - 示例: 3 表示最多掉3个物品
// @Description   - 注意: 实际掉落数量在[min_quantity, max_quantity]之间随机
// @Description
// @Description **掉落机制说明**:
// @Description 1. 击败怪物时，按drop_chance概率判断是否触发该掉落池
// @Description 2. 如果触发，随机抽取[min_quantity, max_quantity]个物品
// @Description 3. 从掉落池中按权重随机选择物品
// @Description 4. 根据drop_type决定是队伍共享还是个人独立
// @Description
// @Description **使用场景示例**:
// @Description - 普通怪物掉落消耗品(个人):
// @Description   {"drop_pool_id": "xxx", "drop_type": "personal", "drop_chance": 0.6, "min_quantity": 1, "max_quantity": 3}
// @Description - BOSS掉落稀有装备(队伍):
// @Description   {"drop_pool_id": "xxx", "drop_type": "team", "drop_chance": 0.1, "min_quantity": 1, "max_quantity": 1}
// @Description - 精英怪掉落材料(个人):
// @Description   {"drop_pool_id": "xxx", "drop_type": "personal", "drop_chance": 1.0, "min_quantity": 2, "max_quantity": 5}
// @Description
// @Description **注意事项**:
// @Description - 同一个怪物不能重复添加相同的掉落池(drop_pool_id唯一)
// @Description - 掉落池ID必须是已存在的掉落池配置，否则返回404
// @Description - drop_chance建议根据物品稀有度设置: 普通0.6-0.8, 稀有0.2-0.4, 传说0.05-0.1
// @Description - 数量范围影响掉落丰富度，建议普通怪1-3，精英2-5，BOSS3-10
// @Tags 怪物掉落管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param request body AddMonsterDropRequest true "添加掉落配置请求"
// @Success 200 {object} response.Response{data=object{message=string}} "添加成功"
// @Failure 400 {object} response.Response "参数错误(100400): 概率超出范围、数量无效等"
// @Failure 404 {object} response.Response "怪物或掉落池不存在(100404)"
// @Failure 409 {object} response.Response "掉落池已添加(100409)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/drops [post]
// @Security BearerAuth
func (h *MonsterHandler) AddMonsterDrop(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")

	var req AddMonsterDropRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := h.service.AddMonsterDrop(ctx, monsterID, req.DropPoolID, req.DropType, req.DropChance, req.MinQuantity, req.MaxQuantity); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// UpdateMonsterDrop 更新怪物掉落配置
// @Summary 更新怪物掉落配置
// @Description 更新怪物已有掉落配置的类型、概率和数量。
// @Description
// @Description **可更新字段**:
// @Description - drop_type: 掉落类型(必填，team或personal)
// @Description - drop_chance: 掉落概率(必填，0.0-1.0)
// @Description - min_quantity: 最小掉落数量(必填，≥1)
// @Description - max_quantity: 最大掉落数量(必填，≥min_quantity)
// @Tags 怪物掉落管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param drop_pool_id path string true "掉落池ID(UUID格式)" example("770e8400-e29b-41d4-a716-446655440000")
// @Param request body UpdateMonsterDropRequest true "更新掉落配置请求"
// @Success 200 {object} response.Response{data=object{message=string}} "更新成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 404 {object} response.Response "怪物掉落配置不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/drops/{drop_pool_id} [put]
// @Security BearerAuth
func (h *MonsterHandler) UpdateMonsterDrop(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")
	dropPoolID := c.Param("drop_pool_id")

	var req UpdateMonsterDropRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := h.service.UpdateMonsterDrop(ctx, monsterID, dropPoolID, req.DropType, req.DropChance, req.MinQuantity, req.MaxQuantity); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// RemoveMonsterDrop 移除怪物掉落配置
// @Summary 移除怪物掉落配置
// @Description 从怪物移除指定掉落配置。
// @Description
// @Description **注意事项**:
// @Description - 移除后该怪物将不再掉落该掉落池的物品
// @Description - 操作不可逆，请谨慎操作
// @Tags 怪物掉落管理
// @Accept json
// @Produce json
// @Param id path string true "怪物ID(UUID格式)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param drop_pool_id path string true "掉落池ID(UUID格式)" example("770e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} response.Response{data=object{message=string}} "移除成功"
// @Failure 400 {object} response.Response "参数错误(100400)"
// @Failure 404 {object} response.Response "怪物掉落配置不存在(100404)"
// @Failure 500 {object} response.Response "服务器错误(100500)"
// @Router /admin/monsters/{id}/drops/{drop_pool_id} [delete]
// @Security BearerAuth
func (h *MonsterHandler) RemoveMonsterDrop(c echo.Context) error {
	ctx := c.Request().Context()
	monsterID := c.Param("id")
	dropPoolID := c.Param("drop_pool_id")

	if err := h.service.RemoveMonsterDrop(ctx, monsterID, dropPoolID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "操作成功"})
}

// ==================== 辅助方法 ====================

// toMonsterEntity 将请求转换为实体
func (h *MonsterHandler) toMonsterEntity(req *CreateMonsterRequest) *game_config.Monster {
	monster := &game_config.Monster{
		MonsterCode:  req.MonsterCode,
		MonsterName:  req.MonsterName,
		MonsterLevel: req.MonsterLevel,
		MaxHP:        req.MaxHP,
	}

	if req.Description != "" {
		monster.Description.SetValid(req.Description)
	}
	if req.HPRecovery > 0 {
		monster.HPRecovery.SetValid(req.HPRecovery)
	}
	if req.MaxMP > 0 {
		monster.MaxMP.SetValid(req.MaxMP)
	}
	if req.MPRecovery > 0 {
		monster.MPRecovery.SetValid(req.MPRecovery)
	}
	if req.BaseSTR > 0 {
		monster.BaseSTR.SetValid(req.BaseSTR)
	}
	if req.BaseAgi > 0 {
		monster.BaseAgi.SetValid(req.BaseAgi)
	}
	if req.BaseVit > 0 {
		monster.BaseVit.SetValid(req.BaseVit)
	}
	if req.BaseWLP > 0 {
		monster.BaseWLP.SetValid(req.BaseWLP)
	}
	if req.BaseInt > 0 {
		monster.BaseInt.SetValid(req.BaseInt)
	}
	if req.BaseWis > 0 {
		monster.BaseWis.SetValid(req.BaseWis)
	}
	if req.BaseCha > 0 {
		monster.BaseCha.SetValid(req.BaseCha)
	}
	// 属性类型代码（引用 hero_attribute_type 表）
	if req.AccuracyAttributeCode != "" {
		monster.AccuracyAttributeCode.SetValid(req.AccuracyAttributeCode)
	}
	if req.DodgeAttributeCode != "" {
		monster.DodgeAttributeCode.SetValid(req.DodgeAttributeCode)
	}
	if req.InitiativeAttributeCode != "" {
		monster.InitiativeAttributeCode.SetValid(req.InitiativeAttributeCode)
	}
	if req.BodyResistAttributeCode != "" {
		monster.BodyResistAttributeCode.SetValid(req.BodyResistAttributeCode)
	}
	if req.MagicResistAttributeCode != "" {
		monster.MagicResistAttributeCode.SetValid(req.MagicResistAttributeCode)
	}
	if req.MentalResistAttributeCode != "" {
		monster.MentalResistAttributeCode.SetValid(req.MentalResistAttributeCode)
	}
	if req.EnvironmentResistAttributeCode != "" {
		monster.EnvironmentResistAttributeCode.SetValid(req.EnvironmentResistAttributeCode)
	}
	if req.DropGoldMin > 0 {
		monster.DropGoldMin.SetValid(req.DropGoldMin)
	}
	if req.DropGoldMax > 0 {
		monster.DropGoldMax.SetValid(req.DropGoldMax)
	}
	if req.DropExp > 0 {
		monster.DropExp.SetValid(req.DropExp)
	}
	if req.IconURL != "" {
		monster.IconURL.SetValid(req.IconURL)
	}
	if req.ModelURL != "" {
		monster.ModelURL.SetValid(req.ModelURL)
	}
	monster.IsActive.SetValid(req.IsActive)
	if req.DisplayOrder > 0 {
		monster.DisplayOrder.SetValid(req.DisplayOrder)
	}

	return monster
}

// toMonsterInfo 将实体转换为响应
func (h *MonsterHandler) toMonsterInfo(monster *game_config.Monster) MonsterInfo {
	info := MonsterInfo{
		ID:           monster.ID,
		MonsterCode:  monster.MonsterCode,
		MonsterName:  monster.MonsterName,
		MonsterLevel: monster.MonsterLevel,
		MaxHP:        monster.MaxHP,
		CreatedAt:    monster.CreatedAt.Unix(),
		UpdatedAt:    monster.UpdatedAt.Unix(),
	}

	if monster.Description.Valid {
		info.Description = monster.Description.String
	}
	if monster.HPRecovery.Valid {
		info.HPRecovery = monster.HPRecovery.Int
	}
	if monster.MaxMP.Valid {
		info.MaxMP = monster.MaxMP.Int
	}
	if monster.MPRecovery.Valid {
		info.MPRecovery = monster.MPRecovery.Int
	}
	if monster.BaseSTR.Valid {
		info.BaseSTR = monster.BaseSTR.Int16
	}
	if monster.BaseAgi.Valid {
		info.BaseAgi = monster.BaseAgi.Int16
	}
	if monster.BaseVit.Valid {
		info.BaseVit = monster.BaseVit.Int16
	}
	if monster.BaseWLP.Valid {
		info.BaseWLP = monster.BaseWLP.Int16
	}
	if monster.BaseInt.Valid {
		info.BaseInt = monster.BaseInt.Int16
	}
	if monster.BaseWis.Valid {
		info.BaseWis = monster.BaseWis.Int16
	}
	if monster.BaseCha.Valid {
		info.BaseCha = monster.BaseCha.Int16
	}
	// 属性类型代码
	if monster.AccuracyAttributeCode.Valid {
		info.AccuracyAttributeCode = monster.AccuracyAttributeCode.String
	}
	if monster.DodgeAttributeCode.Valid {
		info.DodgeAttributeCode = monster.DodgeAttributeCode.String
	}
	if monster.InitiativeAttributeCode.Valid {
		info.InitiativeAttributeCode = monster.InitiativeAttributeCode.String
	}
	if monster.BodyResistAttributeCode.Valid {
		info.BodyResistAttributeCode = monster.BodyResistAttributeCode.String
	}
	if monster.MagicResistAttributeCode.Valid {
		info.MagicResistAttributeCode = monster.MagicResistAttributeCode.String
	}
	if monster.MentalResistAttributeCode.Valid {
		info.MentalResistAttributeCode = monster.MentalResistAttributeCode.String
	}
	if monster.EnvironmentResistAttributeCode.Valid {
		info.EnvironmentResistAttributeCode = monster.EnvironmentResistAttributeCode.String
	}
	if monster.DamageResistances.Valid {
		var resistances map[string]interface{}
		if err := monster.DamageResistances.Unmarshal(&resistances); err == nil {
			info.DamageResistances = resistances
		}
	}
	if monster.PassiveBuffs.Valid {
		var buffs []interface{}
		if err := monster.PassiveBuffs.Unmarshal(&buffs); err == nil {
			info.PassiveBuffs = buffs
		}
	}
	if monster.DropGoldMin.Valid {
		info.DropGoldMin = monster.DropGoldMin.Int
	}
	if monster.DropGoldMax.Valid {
		info.DropGoldMax = monster.DropGoldMax.Int
	}
	if monster.DropExp.Valid {
		info.DropExp = monster.DropExp.Int
	}
	if monster.IconURL.Valid {
		info.IconURL = monster.IconURL.String
	}
	if monster.ModelURL.Valid {
		info.ModelURL = monster.ModelURL.String
	}
	if monster.IsActive.Valid {
		info.IsActive = monster.IsActive.Bool
	}
	if monster.DisplayOrder.Valid {
		info.DisplayOrder = monster.DisplayOrder.Int
	}

	return info
}

// toMonsterSkillInfo 将技能实体转换为响应
func (h *MonsterHandler) toMonsterSkillInfo(skill *game_config.MonsterSkill) MonsterSkillInfo {
	info := MonsterSkillInfo{
		ID:          skill.ID,
		MonsterID:   skill.MonsterID,
		SkillID:     skill.SkillID,
		SkillLevel:  skill.SkillLevel,
		GainActions: []string(skill.GainActions),
		CreatedAt:   skill.CreatedAt.Unix(),
		UpdatedAt:   skill.UpdatedAt.Unix(),
	}

	return info
}

// toMonsterDropInfo 将掉落实体转换为响应
func (h *MonsterHandler) toMonsterDropInfo(drop *game_config.MonsterDrop) MonsterDropInfo {
	info := MonsterDropInfo{
		ID:         drop.ID,
		MonsterID:  drop.MonsterID,
		DropPoolID: drop.DropPoolID,
		DropType:   drop.DropType,
		CreatedAt:  drop.CreatedAt.Unix(),
		UpdatedAt:  drop.UpdatedAt.Unix(),
	}

	// DropChance 是 types.Decimal，直接转换
	chance, _ := drop.DropChance.Float64()
	info.DropChance = chance

	if drop.MinQuantity.Valid {
		info.MinQuantity = drop.MinQuantity.Int
	}
	if drop.MaxQuantity.Valid {
		info.MaxQuantity = drop.MaxQuantity.Int
	}

	return info
}
