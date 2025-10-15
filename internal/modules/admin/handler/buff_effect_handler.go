package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// BuffEffectHandler Buff效果关联 HTTP 处理器
type BuffEffectHandler struct {
	service    *service.BuffEffectService
	respWriter response.Writer
}

// NewBuffEffectHandler 创建Buff效果关联处理器
func NewBuffEffectHandler(db *sql.DB, respWriter response.Writer) *BuffEffectHandler {
	return &BuffEffectHandler{
		service:    service.NewBuffEffectService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// AddBuffEffectRequest 添加Buff效果请求
type AddBuffEffectRequest struct {
	EffectID           string `json:"effect_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"` // 效果ID(必需)
	TriggerTiming      string `json:"trigger_timing" validate:"required" example:"on_apply"`                        // 触发时机(必需,如on_apply/on_tick/on_remove)
	ExecutionOrder     int    `json:"execution_order" example:"1"`                                                  // 执行顺序,数字越小越先执行
	ParameterOverrides string `json:"parameter_overrides" example:"{\"damage\":\"1d6\"}"`                           // 参数覆盖JSON配置
}

// BatchSetBuffEffectsRequest 批量设置Buff效果请求
type BatchSetBuffEffectsRequest struct {
	Effects []AddBuffEffectRequest `json:"effects" validate:"required"` // 效果列表(必需)
}

// BuffEffectInfo Buff效果信息响应
type BuffEffectInfo struct {
	ID                 string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`        // Buff效果关联ID
	BuffID             string `json:"buff_id" example:"550e8400-e29b-41d4-a716-446655440000"`   // Buff ID
	EffectID           string `json:"effect_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 效果ID
	TriggerTiming      string `json:"trigger_timing" example:"on_apply"`                        // 触发时机
	ExecutionOrder     int    `json:"execution_order" example:"1"`                              // 执行顺序
	ParameterOverrides string `json:"parameter_overrides" example:"{\"damage\":\"1d6\"}"`       // 参数覆盖JSON配置
	CreatedAt          int64  `json:"created_at" example:"1633024800"`                          // 创建时间戳
}

// ==================== HTTP Handlers ====================

// GetBuffEffects 获取Buff的所有效果
// @Summary 获取Buff的所有效果
// @Description 获取指定Buff关联的所有效果列表
// @Tags Buff
// @Accept json
// @Produce json
// @Param buff_id path string true "Buff ID (UUID格式)"
// @Success 200 {object} response.Response{data=[]BuffEffectInfo} "成功返回Buff效果列表"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "Buff不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{buff_id}/effects [get]
// @Security BearerAuth
func (h *BuffEffectHandler) GetBuffEffects(c echo.Context) error {
	ctx := c.Request().Context()
	buffID := c.Param("buff_id")

	buffEffects, err := h.service.GetBuffEffects(ctx, buffID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	result := make([]BuffEffectInfo, len(buffEffects))
	for i, be := range buffEffects {
		result[i] = h.convertToBuffEffectInfo(be)
	}

	return response.EchoOK(c, h.respWriter, result)
}

// AddBuffEffect 为Buff添加效果
// @Summary 为Buff添加效果
// @Description 为指定Buff添加一个效果，包括触发时机和执行顺序配置
// @Tags Buff
// @Accept json
// @Produce json
// @Param buff_id path string true "Buff ID (UUID格式)"
// @Param request body AddBuffEffectRequest true "添加效果请求参数"
// @Success 200 {object} response.Response{data=BuffEffectInfo} "成功返回添加的效果信息"
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "Buff或效果不存在"
// @Failure 409 {object} response.Response "效果已关联到该Buff"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{buff_id}/effects [post]
// @Security BearerAuth
func (h *BuffEffectHandler) AddBuffEffect(c echo.Context) error {
	ctx := c.Request().Context()
	buffID := c.Param("buff_id")

	var req AddBuffEffectRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	buffEffect := &game_config.BuffEffect{
		BuffID:        buffID,
		EffectID:      req.EffectID,
		TriggerTiming: req.TriggerTiming,
	}
	buffEffect.ExecutionOrder.SetValid(req.ExecutionOrder)

	if req.ParameterOverrides != "" {
		buffEffect.ParameterOverrides.UnmarshalJSON([]byte(req.ParameterOverrides))
	}

	if err := h.service.AddEffectToBuff(ctx, buffEffect); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToBuffEffectInfo(buffEffect))
}

// RemoveBuffEffect 从Buff移除效果
// @Summary 从Buff移除效果
// @Description 从指定Buff中移除一个效果关联
// @Tags Buff
// @Accept json
// @Produce json
// @Param buff_id path string true "Buff ID (UUID格式)"
// @Param effect_id path string true "Buff效果关联ID (buff_effects表的ID，非effect ID)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "Buff效果关联不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{buff_id}/effects/{effect_id} [delete]
// @Security BearerAuth
func (h *BuffEffectHandler) RemoveBuffEffect(c echo.Context) error {
	ctx := c.Request().Context()
	buffEffectID := c.Param("effect_id")

	if err := h.service.RemoveEffectFromBuff(ctx, buffEffectID); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "效果移除成功",
	})
}

// BatchSetBuffEffects 批量设置Buff效果
// @Summary 批量设置Buff效果
// @Description 批量设置Buff的效果列表（先删除旧效果，再添加新效果，事务保证）
// @Tags Buff
// @Accept json
// @Produce json
// @Param buff_id path string true "Buff ID (UUID格式)"
// @Param request body BatchSetBuffEffectsRequest true "批量设置效果请求参数"
// @Success 200 {object} response.Response "成功返回批量设置结果"
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "Buff或效果不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{buff_id}/effects/batch [post]
// @Security BearerAuth
func (h *BuffEffectHandler) BatchSetBuffEffects(c echo.Context) error {
	ctx := c.Request().Context()
	buffID := c.Param("buff_id")

	var req BatchSetBuffEffectsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	buffEffects := make([]*game_config.BuffEffect, len(req.Effects))
	for i, e := range req.Effects {
		buffEffect := &game_config.BuffEffect{
			BuffID:        buffID,
			EffectID:      e.EffectID,
			TriggerTiming: e.TriggerTiming,
		}
		buffEffect.ExecutionOrder.SetValid(e.ExecutionOrder)

		if e.ParameterOverrides != "" {
			buffEffect.ParameterOverrides.UnmarshalJSON([]byte(e.ParameterOverrides))
		}

		buffEffects[i] = buffEffect
	}

	if err := h.service.BatchSetBuffEffects(ctx, buffID, buffEffects); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "批量设置效果成功",
	})
}

// ==================== Helper Functions ====================

func (h *BuffEffectHandler) convertToBuffEffectInfo(buffEffect *game_config.BuffEffect) BuffEffectInfo {
	info := BuffEffectInfo{
		ID:            buffEffect.ID,
		BuffID:        buffEffect.BuffID,
		EffectID:      buffEffect.EffectID,
		TriggerTiming: buffEffect.TriggerTiming,
	}

	if buffEffect.ExecutionOrder.Valid {
		info.ExecutionOrder = buffEffect.ExecutionOrder.Int
	}

	if buffEffect.ParameterOverrides.Valid {
		jsonBytes, _ := buffEffect.ParameterOverrides.MarshalJSON()
		info.ParameterOverrides = string(jsonBytes)
	}

	if buffEffect.CreatedAt.Valid {
		info.CreatedAt = buffEffect.CreatedAt.Time.Unix()
	}

	return info
}
