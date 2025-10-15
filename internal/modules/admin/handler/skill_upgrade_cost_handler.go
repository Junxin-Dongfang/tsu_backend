package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// SkillUpgradeCostHandler 技能升级消耗 HTTP 处理器
type SkillUpgradeCostHandler struct {
	service    *service.SkillUpgradeCostService
	respWriter response.Writer
}

// NewSkillUpgradeCostHandler 创建技能升级消耗处理器
func NewSkillUpgradeCostHandler(db *sql.DB, respWriter response.Writer) *SkillUpgradeCostHandler {
	return &SkillUpgradeCostHandler{
		service:    service.NewSkillUpgradeCostService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateSkillUpgradeCostRequest 创建升级消耗请求
type CreateSkillUpgradeCostRequest struct {
	LevelNumber   int             `json:"level_number" validate:"required,min=2,max=100"`
	CostXP        int             `json:"cost_xp"`
	CostGold      int             `json:"cost_gold"`
	CostMaterials json.RawMessage `json:"cost_materials"` // JSON数组: [{"item_code": "xxx", "count": 5}]
}

// UpdateSkillUpgradeCostRequest 更新升级消耗请求
type UpdateSkillUpgradeCostRequest struct {
	CostXP        int             `json:"cost_xp"`
	CostGold      int             `json:"cost_gold"`
	CostMaterials json.RawMessage `json:"cost_materials"`
}

// SkillUpgradeCostInfo 升级消耗信息响应
type SkillUpgradeCostInfo struct {
	ID            string          `json:"id"`
	LevelNumber   int             `json:"level_number"`
	CostXP        int             `json:"cost_xp"`
	CostGold      int             `json:"cost_gold"`
	CostMaterials json.RawMessage `json:"cost_materials"`
	CreatedAt     int64           `json:"created_at"`
	UpdatedAt     int64           `json:"updated_at"`
}

// ==================== HTTP Handlers ====================

// GetSkillUpgradeCosts 获取所有升级消耗配置
// @Summary 获取所有升级消耗配置
// @Description 获取全局技能升级消耗配置列表，所有技能共享这些升级消耗规则
// @Tags 技能
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]SkillUpgradeCostInfo} "成功返回升级消耗列表"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs [get]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) GetSkillUpgradeCosts(c echo.Context) error {
	ctx := c.Request().Context()

	costs, err := h.service.GetAllSkillUpgradeCosts(ctx)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]SkillUpgradeCostInfo, len(costs))
	for i, cost := range costs {
		result[i] = h.convertToInfo(cost)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// GetSkillUpgradeCost 获取单个升级消耗配置
// @Summary 获取单个升级消耗配置
// @Description 根据ID获取指定的升级消耗配置详情
// @Tags 技能
// @Accept json
// @Produce json
// @Param id path string true "升级消耗配置ID (UUID格式)"
// @Success 200 {object} response.Response{data=SkillUpgradeCostInfo} "成功返回升级消耗详情"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs/{id} [get]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) GetSkillUpgradeCost(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	cost, err := h.service.GetSkillUpgradeCostByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(cost))
}

// GetSkillUpgradeCostByLevel 根据等级获取升级消耗
// @Summary 根据等级获取升级消耗
// @Description 根据目标等级获取对应的升级消耗配置
// @Tags 技能
// @Accept json
// @Produce json
// @Param level path int true "目标等级 (2-100)"
// @Success 200 {object} response.Response{data=SkillUpgradeCostInfo} "成功返回升级消耗详情"
// @Failure 404 {object} response.Response "该等级配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs/level/{level} [get]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) GetSkillUpgradeCostByLevel(c echo.Context) error {
	ctx := c.Request().Context()
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return response.EchoBadRequest(c, h.respWriter, "等级必须是有效的数字")
	}

	cost, err := h.service.GetSkillUpgradeCostByLevel(ctx, level)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(cost))
}

// CreateSkillUpgradeCost 创建升级消耗配置
// @Summary 创建升级消耗配置
// @Description 创建新的等级升级消耗配置
// @Tags 技能
// @Accept json
// @Produce json
// @Param request body CreateSkillUpgradeCostRequest true "创建请求，level_number为必填"
// @Success 200 {object} response.Response{data=SkillUpgradeCostInfo} "成功创建升级消耗配置"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "该等级配置已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs [post]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) CreateSkillUpgradeCost(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateSkillUpgradeCostRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 验证 cost_materials JSON格式
	if len(req.CostMaterials) > 0 {
		var materials interface{}
		if err := json.Unmarshal(req.CostMaterials, &materials); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "cost_materials 必须是有效的 JSON 数组")
		}
	}

	// 构造实体
	cost := &game_config.SkillUpgradeCost{
		LevelNumber: req.LevelNumber,
	}

	if req.CostXP > 0 {
		cost.CostXP.SetValid(req.CostXP)
	}
	if req.CostGold > 0 {
		cost.CostGold.SetValid(req.CostGold)
	}
	if len(req.CostMaterials) > 0 {
		cost.CostMaterials.UnmarshalJSON(req.CostMaterials)
	}

	if err := h.service.CreateSkillUpgradeCost(ctx, cost); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(cost))
}

// UpdateSkillUpgradeCost 更新升级消耗配置
// @Summary 更新升级消耗配置
// @Description 更新指定的升级消耗配置，仅更新提供的字段
// @Tags 技能
// @Accept json
// @Produce json
// @Param id path string true "升级消耗配置ID (UUID格式)"
// @Param request body UpdateSkillUpgradeCostRequest true "更新请求"
// @Success 200 {object} response.Response{data=map[string]string} "成功更新配置"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs/{id} [put]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) UpdateSkillUpgradeCost(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateSkillUpgradeCostRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.CostXP >= 0 {
		updates["cost_xp"] = req.CostXP
	}
	if req.CostGold >= 0 {
		updates["cost_gold"] = req.CostGold
	}
	if len(req.CostMaterials) > 0 {
		// 验证 JSON 格式
		var materials interface{}
		if err := json.Unmarshal(req.CostMaterials, &materials); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "cost_materials 必须是有效的 JSON")
		}
		updates["cost_materials"] = req.CostMaterials
	}

	if err := h.service.UpdateSkillUpgradeCost(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "升级消耗配置更新成功",
	})
}

// DeleteSkillUpgradeCost 删除升级消耗配置
// @Summary 删除升级消耗配置
// @Description 删除指定的升级消耗配置
// @Tags 技能
// @Accept json
// @Produce json
// @Param id path string true "升级消耗配置ID (UUID格式)"
// @Success 200 {object} response.Response{data=map[string]string} "成功删除配置"
// @Failure 404 {object} response.Response "配置不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/skill-upgrade-costs/{id} [delete]
// @Security BearerAuth
func (h *SkillUpgradeCostHandler) DeleteSkillUpgradeCost(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteSkillUpgradeCost(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "升级消耗配置删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *SkillUpgradeCostHandler) convertToInfo(cost *game_config.SkillUpgradeCost) SkillUpgradeCostInfo {
	info := SkillUpgradeCostInfo{
		ID:          cost.ID,
		LevelNumber: cost.LevelNumber,
	}

	// null.Time 类型需要检查
	if !cost.CreatedAt.IsZero() {
		info.CreatedAt = cost.CreatedAt.Time.Unix()
	}
	if !cost.UpdatedAt.IsZero() {
		info.UpdatedAt = cost.UpdatedAt.Time.Unix()
	}

	if cost.CostXP.Valid {
		info.CostXP = cost.CostXP.Int
	}

	if cost.CostGold.Valid {
		info.CostGold = cost.CostGold.Int
	}

	if cost.CostMaterials.Valid {
		jsonBytes, _ := cost.CostMaterials.MarshalJSON()
		info.CostMaterials = json.RawMessage(jsonBytes)
	} else {
		info.CostMaterials = json.RawMessage("[]")
	}

	return info
}
