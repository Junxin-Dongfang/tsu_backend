package handler

import (
	"database/sql"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// DungeonBattleHandler 战斗配置 HTTP 处理器
type DungeonBattleHandler struct {
	service    *service.DungeonBattleService
	respWriter response.Writer
}

// NewDungeonBattleHandler 创建战斗配置处理器
func NewDungeonBattleHandler(db *sql.DB, respWriter response.Writer) *DungeonBattleHandler {
	return &DungeonBattleHandler{
		service:    service.NewDungeonBattleService(db),
		respWriter: respWriter,
	}
}

// CreateBattle 创建战斗配置
// @Summary 创建战斗配置
// @Description 创建新的战斗配置,包含场景/环境参数、全程 Buff、怪物阵容以及三段描述。
// @Description
// @Description **字段说明**:
// @Description - `battle_code`: 唯一代码,与房间序列关联
// @Description - `location_config`: JSON对象,可配置场景、光照、背景音乐、地形,前端原样读取
// @Description - `global_buffs`: 全局 Buff 列表,每项包含 `buff_code` 与 `target`(allies/enemies/all)
// @Description - `monster_setup`: 阵容列表,含 `monster_code`, `position(1-21)` 和可选 `level_override`
// @Description - `battle_start_desc` / `battle_success_desc` / `battle_failure_desc`: 文案描述
// @Description - `is_active`: 控制该战斗是否可被房间引用
// @Tags 地城战斗管理
// @Accept json
// @Produce json
// @Param request body dto.CreateBattleRequest true "创建战斗配置请求"
// @Success 201 {object} response.Response{data=dto.BattleResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "战斗配置代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-battles [post]
func (h *DungeonBattleHandler) CreateBattle(c echo.Context) error {
	var req dto.CreateBattleRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	battle, err := h.service.CreateBattle(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toBattleResponse(battle)
	return response.EchoOK(c, h.respWriter, resp)
}

// GetBattle 获取战斗配置详情
// @Summary 获取战斗配置详情
// @Description 根据ID获取战斗配置详细信息,包含场景配置、全程 Buff、怪物阵容以及三段描述文案,用于编辑或复用。
// @Tags 地城战斗管理
// @Accept json
// @Produce json
// @Param id path string true "战斗配置ID"
// @Success 200 {object} response.Response{data=dto.BattleResponse} "查询成功"
// @Failure 404 {object} response.Response "战斗配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-battles/{id} [get]
func (h *DungeonBattleHandler) GetBattle(c echo.Context) error {
	battleID := c.Param("id")

	battle, err := h.service.GetBattleByID(c.Request().Context(), battleID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toBattleResponse(battle)
	return response.EchoOK(c, h.respWriter, resp)
}

// UpdateBattle 更新战斗配置
// @Summary 更新战斗配置
// @Description 更新战斗配置信息,支持只调整部分字段(例如替换场景、增减怪物或修改文案)。
// @Tags 地城战斗管理
// @Accept json
// @Produce json
// @Param id path string true "战斗配置ID"
// @Param request body dto.UpdateBattleRequest true "更新战斗配置请求"
// @Success 200 {object} response.Response{data=dto.BattleResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "战斗配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-battles/{id} [put]
func (h *DungeonBattleHandler) UpdateBattle(c echo.Context) error {
	battleID := c.Param("id")

	var req dto.UpdateBattleRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	battle, err := h.service.UpdateBattle(c.Request().Context(), battleID, &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	resp := h.toBattleResponse(battle)
	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteBattle 删除战斗配置
// @Summary 删除战斗配置
// @Description 软删除战斗配置(原房间若引用需要手动调整),常用于清理废弃战斗。
// @Tags 地城战斗管理
// @Accept json
// @Produce json
// @Param id path string true "战斗配置ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "战斗配置不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/dungeon-battles/{id} [delete]
func (h *DungeonBattleHandler) DeleteBattle(c echo.Context) error {
	battleID := c.Param("id")

	if err := h.service.DeleteBattle(c.Request().Context(), battleID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "删除成功"})
}

// toBattleResponse 转换为响应格式
func (h *DungeonBattleHandler) toBattleResponse(battle *game_config.DungeonBattle) dto.BattleResponse {
	resp := dto.BattleResponse{
		ID:         battle.ID,
		BattleCode: battle.BattleCode,
		IsActive:   battle.IsActive,
		CreatedAt:  battle.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  battle.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 解析场地配置
	var locationConfig map[string]interface{}
	if err := json.Unmarshal(battle.LocationConfig, &locationConfig); err == nil {
		resp.LocationConfig = locationConfig
	}

	// 解析全程Buff
	var globalBuffs []dto.GlobalBuffItem
	if err := json.Unmarshal(battle.GlobalBuffs, &globalBuffs); err == nil {
		resp.GlobalBuffs = globalBuffs
	}

	// 解析怪物配置
	var monsterSetup []dto.MonsterSetupItem
	if err := json.Unmarshal(battle.MonsterSetup, &monsterSetup); err == nil {
		resp.MonsterSetup = monsterSetup
	}

	if battle.BattleStartDesc.Valid {
		desc := battle.BattleStartDesc.String
		resp.BattleStartDesc = &desc
	}

	if battle.BattleSuccessDesc.Valid {
		desc := battle.BattleSuccessDesc.String
		resp.BattleSuccessDesc = &desc
	}

	if battle.BattleFailureDesc.Valid {
		desc := battle.BattleFailureDesc.String
		resp.BattleFailureDesc = &desc
	}

	return resp
}
