package handler

import (
	"database/sql"
	"encoding/json"

	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// SkillUnlockActionHandler 技能解锁动作 HTTP 处理器
type SkillUnlockActionHandler struct {
	service    *service.SkillUnlockActionService
	respWriter response.Writer
}

// NewSkillUnlockActionHandler 创建技能解锁动作处理器
func NewSkillUnlockActionHandler(db *sql.DB, respWriter response.Writer) *SkillUnlockActionHandler {
	return &SkillUnlockActionHandler{
		service:    service.NewSkillUnlockActionService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// AddSkillUnlockActionRequest 添加技能解锁动作请求
type AddSkillUnlockActionRequest struct {
	ActionID           string                 `json:"action_id" validate:"required"`
	UnlockLevel        int                    `json:"unlock_level" validate:"required,min=1"`
	IsDefault          bool                   `json:"is_default"`
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"` // 动作等级成长配置
}

// BatchSetSkillUnlockActionsRequest 批量设置技能解锁动作请求
type BatchSetSkillUnlockActionsRequest struct {
	Actions []AddSkillUnlockActionRequest `json:"actions" validate:"required"`
}

// SkillUnlockActionInfo 技能解锁动作信息响应
type SkillUnlockActionInfo struct {
	ID                 string                 `json:"id"`
	SkillID            string                 `json:"skill_id"`
	ActionID           string                 `json:"action_id"`
	UnlockLevel        int                    `json:"unlock_level"`
	IsDefault          bool                   `json:"is_default"`
	LevelScalingConfig map[string]interface{} `json:"level_scaling_config,omitempty"` // 动作等级成长配置
	CreatedAt          int64                  `json:"created_at"`
}

// ==================== HTTP Handlers ====================

// GetSkillUnlockActions 获取技能的所有解锁动作
// @Summary 获取技能的所有解锁动作
// @Description 获取指定技能关联的所有解锁动作列表,按解锁等级排序
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param skill_id path string true "技能的UUID"
// @Success 200 {object} response.Response{data=[]SkillUnlockActionInfo}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "技能不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/skills/{skill_id}/unlock-actions [get]
func (h *SkillUnlockActionHandler) GetSkillUnlockActions(c echo.Context) error {
	ctx := c.Request().Context()
	skillID := c.Param("skill_id")

	unlockActions, err := h.service.GetSkillUnlockActions(ctx, skillID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]SkillUnlockActionInfo, len(unlockActions))
	for i, ua := range unlockActions {
		result[i] = h.convertToSkillUnlockActionInfo(ua)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// AddSkillUnlockAction 为技能添加解锁动作
// @Summary 为技能添加解锁动作
// @Description 为指定技能添加单个解锁动作关联,指定解锁等级、是否为默认动作以及动作的等级成长配置
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param skill_id path string true "技能的UUID"
// @Param request body AddSkillUnlockActionRequest true "添加解锁动作的请求参数,包含动作ID、解锁等级和成长配置(level_scaling_config格式: {\"damage\":{\"type\":\"linear\",\"base\":10,\"value\":2}})"
// @Success 200 {object} response.Response{data=SkillUnlockActionInfo}
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "技能或动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/skills/{skill_id}/unlock-actions [post]
func (h *SkillUnlockActionHandler) AddSkillUnlockAction(c echo.Context) error {
	ctx := c.Request().Context()
	skillID := c.Param("skill_id")

	var req AddSkillUnlockActionRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	unlockAction := &game_config.SkillUnlockAction{
		SkillID:  skillID,
		ActionID: req.ActionID,
	}
	unlockAction.UnlockLevel = req.UnlockLevel
	unlockAction.IsDefault.SetValid(req.IsDefault)

	// 处理成长配置
	if req.LevelScalingConfig != nil && len(req.LevelScalingConfig) > 0 {
		configJSON, err := json.Marshal(req.LevelScalingConfig)
		if err != nil {
			return response.EchoBadRequest(c, h.respWriter, "level_scaling_config 格式错误")
		}
		unlockAction.LevelScalingConfig = null.JSONFrom(configJSON)
	}

	if err := h.service.AddUnlockActionToSkill(ctx, unlockAction); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToSkillUnlockActionInfo(unlockAction))
}

// RemoveSkillUnlockAction 从技能移除解锁动作
// @Summary 从技能移除解锁动作
// @Description 删除指定的技能-解锁动作关联记录
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param skill_id path string true "技能的UUID"
// @Param action_id path string true "技能解锁动作关联记录的UUID(不是动作ID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "关联记录不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/skills/{skill_id}/unlock-actions/{action_id} [delete]
func (h *SkillUnlockActionHandler) RemoveSkillUnlockAction(c echo.Context) error {
	ctx := c.Request().Context()
	unlockActionID := c.Param("action_id")

	if err := h.service.RemoveUnlockActionFromSkill(ctx, unlockActionID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "解锁动作移除成功",
	})
}

// BatchSetSkillUnlockActions 批量设置技能解锁动作
// @Summary 批量设置技能解锁动作
// @Description 批量设置技能的所有解锁动作关联(先删除旧关联,再创建新关联,事务保证),每个动作可配置独立的成长曲线
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param skill_id path string true "技能的UUID"
// @Param request body BatchSetSkillUnlockActionsRequest true "批量设置解锁动作的请求参数,包含动作列表及其成长配置"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "技能或动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/skills/{skill_id}/unlock-actions/batch [post]
func (h *SkillUnlockActionHandler) BatchSetSkillUnlockActions(c echo.Context) error {
	ctx := c.Request().Context()
	skillID := c.Param("skill_id")

	var req BatchSetSkillUnlockActionsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	unlockActions := make([]*game_config.SkillUnlockAction, len(req.Actions))
	for i, a := range req.Actions {
		unlockAction := &game_config.SkillUnlockAction{
			SkillID:  skillID,
			ActionID: a.ActionID,
		}
		unlockAction.UnlockLevel = a.UnlockLevel
		unlockAction.IsDefault.SetValid(a.IsDefault)

		// 处理成长配置
		if a.LevelScalingConfig != nil && len(a.LevelScalingConfig) > 0 {
			configJSON, err := json.Marshal(a.LevelScalingConfig)
			if err != nil {
				return response.EchoBadRequest(c, h.respWriter, "level_scaling_config 格式错误")
			}
			unlockAction.LevelScalingConfig = null.JSONFrom(configJSON)
		}

		unlockActions[i] = unlockAction
	}

	if err := h.service.BatchSetSkillUnlockActions(ctx, skillID, unlockActions); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "批量设置解锁动作成功",
	})
}

// ==================== Helper Functions ====================

func (h *SkillUnlockActionHandler) convertToSkillUnlockActionInfo(unlockAction *game_config.SkillUnlockAction) SkillUnlockActionInfo {
	info := SkillUnlockActionInfo{
		ID:          unlockAction.ID,
		SkillID:     unlockAction.SkillID,
		ActionID:    unlockAction.ActionID,
		UnlockLevel: unlockAction.UnlockLevel,
	}

	if unlockAction.IsDefault.Valid {
		info.IsDefault = unlockAction.IsDefault.Bool
	}

	if unlockAction.CreatedAt.Valid {
		info.CreatedAt = unlockAction.CreatedAt.Time.Unix()
	}

	// 处理成长配置
	if !unlockAction.LevelScalingConfig.IsZero() {
		var scalingConfig map[string]interface{}
		if err := json.Unmarshal(unlockAction.LevelScalingConfig.JSON, &scalingConfig); err == nil {
			info.LevelScalingConfig = scalingConfig
		}
	}

	return info
}
