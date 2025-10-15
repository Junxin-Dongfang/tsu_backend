package handler

import (
	"database/sql"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// ActionEffectHandler 动作效果关联 HTTP 处理器
type ActionEffectHandler struct {
	service    *service.ActionEffectService
	respWriter response.Writer
}

// NewActionEffectHandler 创建动作效果关联处理器
func NewActionEffectHandler(db *sql.DB, respWriter response.Writer) *ActionEffectHandler {
	return &ActionEffectHandler{
		service:    service.NewActionEffectService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// AddActionEffectRequest 添加动作效果请求
type AddActionEffectRequest struct {
	EffectID           string `json:"effect_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 效果ID
	ExecutionOrder     int    `json:"execution_order" example:"1"`                                                  // 执行顺序,数字越小越先执行
	ParameterOverrides string `json:"parameter_overrides" example:"{\"damage\":\"2d6\"}"`                           // 参数覆盖JSON配置
	IsConditional      bool   `json:"is_conditional" example:"false"`                                               // 是否有条件
	ConditionConfig    string `json:"condition_config" example:"{\"type\":\"hp_below\",\"value\":50}"`              // 条件配置JSON
}

// BatchSetActionEffectsRequest 批量设置动作效果请求
type BatchSetActionEffectsRequest struct {
	Effects []AddActionEffectRequest `json:"effects" validate:"required"` // 效果列表
}

// ActionEffectInfo 动作效果信息响应
type ActionEffectInfo struct {
	ID                 string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`               // 动作效果关联ID
	ActionID           string `json:"action_id" example:"550e8400-e29b-41d4-a716-446655440000"`        // 动作ID
	EffectID           string `json:"effect_id" example:"550e8400-e29b-41d4-a716-446655440000"`        // 效果ID
	ExecutionOrder     int    `json:"execution_order" example:"1"`                                     // 执行顺序
	ParameterOverrides string `json:"parameter_overrides" example:"{\"damage\":\"2d6\"}"`              // 参数覆盖JSON配置
	IsConditional      bool   `json:"is_conditional" example:"false"`                                  // 是否有条件
	ConditionConfig    string `json:"condition_config" example:"{\"type\":\"hp_below\",\"value\":50}"` // 条件配置JSON
	CreatedAt          int64  `json:"created_at" example:"1633024800"`                                 // 创建时间戳
}

// ==================== HTTP Handlers ====================

// GetActionEffects 获取动作的所有效果
// @Summary 获取动作的所有效果
// @Description 获取指定动作关联的所有效果列表,按执行顺序排序
// @Tags 动作
// @Accept json
// @Produce json
// @Param action_id path string true "动作的UUID"
// @Success 200 {object} response.Response{data=[]ActionEffectInfo}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{action_id}/effects [get]
func (h *ActionEffectHandler) GetActionEffects(c echo.Context) error {
	ctx := c.Request().Context()
	actionID := c.Param("action_id")

	actionEffects, err := h.service.GetActionEffects(ctx, actionID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]ActionEffectInfo, len(actionEffects))
	for i, ae := range actionEffects {
		result[i] = h.convertToActionEffectInfo(ae)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// AddActionEffect 为动作添加效果
// @Summary 为动作添加效果
// @Description 为指定动作添加单个效果关联,支持参数覆盖和条件配置
// @Tags 动作
// @Accept json
// @Produce json
// @Param action_id path string true "动作的UUID"
// @Param request body AddActionEffectRequest true "添加效果的请求参数,包含效果ID和执行顺序"
// @Success 200 {object} response.Response{data=ActionEffectInfo}
// @Failure 400 {object} response.Response "参数错误或验证失败(包括JSONB格式错误)"
// @Failure 404 {object} response.Response "动作或效果不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{action_id}/effects [post]
func (h *ActionEffectHandler) AddActionEffect(c echo.Context) error {
	ctx := c.Request().Context()
	actionID := c.Param("action_id")

	var req AddActionEffectRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	actionEffect := &game_config.ActionEffect{
		ActionID: actionID,
		EffectID: req.EffectID,
	}
	actionEffect.ExecutionOrder.SetValid(req.ExecutionOrder)
	actionEffect.IsConditional.SetValid(req.IsConditional)

	if req.ParameterOverrides != "" {
		var paramJSON interface{}
		if err := json.Unmarshal([]byte(req.ParameterOverrides), &paramJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "parameter_overrides 必须是有效的 JSON")
		}
		actionEffect.ParameterOverrides.UnmarshalJSON([]byte(req.ParameterOverrides))
	}

	if req.ConditionConfig != "" {
		var condJSON interface{}
		if err := json.Unmarshal([]byte(req.ConditionConfig), &condJSON); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "condition_config 必须是有效的 JSON")
		}
		actionEffect.ConditionConfig.UnmarshalJSON([]byte(req.ConditionConfig))
	}

	if err := h.service.AddEffectToAction(ctx, actionEffect); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionEffectInfo(actionEffect))
}

// RemoveActionEffect 从动作移除效果
// @Summary 从动作移除效果
// @Description 删除指定的动作-效果关联记录
// @Tags 动作
// @Accept json
// @Produce json
// @Param action_id path string true "动作的UUID"
// @Param effect_id path string true "动作效果关联记录的UUID(不是效果ID)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "关联记录不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{action_id}/effects/{effect_id} [delete]
func (h *ActionEffectHandler) RemoveActionEffect(c echo.Context) error {
	ctx := c.Request().Context()
	actionEffectID := c.Param("effect_id")

	if err := h.service.RemoveEffectFromAction(ctx, actionEffectID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "效果移除成功",
	})
}

// BatchSetActionEffects 批量设置动作效果
// @Summary 批量设置动作效果
// @Description 批量设置动作的所有效果关联(先删除旧关联,再创建新关联,事务保证)
// @Tags 动作
// @Accept json
// @Produce json
// @Param action_id path string true "动作的UUID"
// @Param request body BatchSetActionEffectsRequest true "批量设置效果的请求参数,包含效果列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "动作或效果不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/actions/{action_id}/effects/batch [post]
func (h *ActionEffectHandler) BatchSetActionEffects(c echo.Context) error {
	ctx := c.Request().Context()
	actionID := c.Param("action_id")

	var req BatchSetActionEffectsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	actionEffects := make([]*game_config.ActionEffect, len(req.Effects))
	for i, e := range req.Effects {
		actionEffect := &game_config.ActionEffect{
			ActionID: actionID,
			EffectID: e.EffectID,
		}
		actionEffect.ExecutionOrder.SetValid(e.ExecutionOrder)
		actionEffect.IsConditional.SetValid(e.IsConditional)

		if e.ParameterOverrides != "" {
			actionEffect.ParameterOverrides.UnmarshalJSON([]byte(e.ParameterOverrides))
		}

		if e.ConditionConfig != "" {
			actionEffect.ConditionConfig.UnmarshalJSON([]byte(e.ConditionConfig))
		}

		actionEffects[i] = actionEffect
	}

	if err := h.service.BatchSetActionEffects(ctx, actionID, actionEffects); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "批量设置效果成功",
	})
}

// ==================== Helper Functions ====================

func (h *ActionEffectHandler) convertToActionEffectInfo(actionEffect *game_config.ActionEffect) ActionEffectInfo {
	info := ActionEffectInfo{
		ID:       actionEffect.ID,
		ActionID: actionEffect.ActionID,
		EffectID: actionEffect.EffectID,
	}

	if actionEffect.ExecutionOrder.Valid {
		info.ExecutionOrder = actionEffect.ExecutionOrder.Int
	}

	if actionEffect.IsConditional.Valid {
		info.IsConditional = actionEffect.IsConditional.Bool
	}

	if actionEffect.ParameterOverrides.Valid {
		jsonBytes, _ := actionEffect.ParameterOverrides.MarshalJSON()
		info.ParameterOverrides = string(jsonBytes)
	}

	if actionEffect.ConditionConfig.Valid {
		jsonBytes, _ := actionEffect.ConditionConfig.MarshalJSON()
		info.ConditionConfig = string(jsonBytes)
	}

	if actionEffect.CreatedAt.Valid {
		info.CreatedAt = actionEffect.CreatedAt.Time.Unix()
	}

	return info
}
