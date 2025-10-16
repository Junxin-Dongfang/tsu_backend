package handler

import (
	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// HeroAttributeHandler handles hero attribute HTTP requests
type HeroAttributeHandler struct {
	attrService *service.HeroAttributeService
	respWriter  response.Writer
}

// NewHeroAttributeHandler creates a new hero attribute handler
func NewHeroAttributeHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *HeroAttributeHandler {
	return &HeroAttributeHandler{
		attrService: serviceContainer.GetHeroAttributeService(),
		respWriter:  respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// AllocateAttributeRequest HTTP allocate attribute request
type AllocateAttributeRequest struct {
	AttributeCode string `json:"attribute_code" validate:"required" example:"STR"`
	PointsToAdd   int    `json:"points_to_add" validate:"required,min=1" example:"2"`
}

// RollbackAttributeRequest HTTP rollback attribute request
type RollbackAttributeRequest struct {
	AttributeCode string `json:"attribute_code" validate:"required" example:"STR"`
}

// ComputedAttributeResponse HTTP computed attribute response
type ComputedAttributeResponse struct {
	AttributeCode string `json:"attribute_code"`
	AttributeName string `json:"attribute_name"`
	BaseValue     int    `json:"base_value"`      // 基础加点值
	ClassBonus    int    `json:"class_bonus"`     // 职业加成
	FinalValue    int    `json:"final_value"`     // 最终值
}

// ==================== HTTP Handlers ====================

// AllocateAttribute handles attribute allocation
// @Summary 属性加点
// @Description 为英雄分配属性点，消耗经验值提升属性。支持堆栈式回退（1小时内可回退）
// @Tags 英雄属性
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param request body AllocateAttributeRequest true "属性加点请求"
// @Success 200 {object} response.Response{data=object{message=string}} "加点成功，返回 {\"message\": \"属性加点成功\"}"
// @Failure 400 {object} response.Response "请求参数错误或验证失败"
// @Failure 404 {object} response.Response "英雄不存在或属性代码无效（错误码: 100404）"
// @Failure 409 {object} response.Response "经验不足（错误码: 100401）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/attributes/allocate [post]
func (h *HeroAttributeHandler) AllocateAttribute(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req AllocateAttributeRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 调用 Service
	serviceReq := &service.AllocateAttributeRequest{
		HeroID:        heroID,
		AttributeCode: req.AttributeCode,
		PointsToAdd:   req.PointsToAdd,
	}

	if err := h.attrService.AllocateAttribute(c.Request().Context(), serviceReq); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "属性加点成功",
	})
}

// RollbackAttribute handles attribute rollback
// @Summary 回退属性加点
// @Description 回退英雄最近一次属性加点操作。支持堆栈式（LIFO）回退，仅限1小时内的操作
// @Tags 英雄属性
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Param request body RollbackAttributeRequest true "回退请求"
// @Success 200 {object} response.Response{data=object{message=string}} "回退成功，返回 {\"message\": \"属性回退成功\"}"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在或没有可回退的操作（错误码: 100404）"
// @Failure 409 {object} response.Response "回退时间已过期（超过1小时）（错误码: 100409）"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/attributes/rollback [post]
func (h *HeroAttributeHandler) RollbackAttribute(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	var req RollbackAttributeRequest

	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	if err := h.attrService.RollbackAttributeAllocation(c.Request().Context(), heroID, req.AttributeCode); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{
		"message": "属性回退成功",
	})
}

// GetComputedAttributes handles getting computed attributes
// @Summary 获取英雄计算属性
// @Description 获取英雄的计算后属性值列表。包含：基础值（加点值）、职业加成、最终计算值。数据来自计算视图（hero_computed_attributes）
// @Tags 英雄属性
// @Accept json
// @Produce json
// @Param hero_id path string true "英雄ID（UUID格式）"
// @Success 200 {object} response.Response{data=[]ComputedAttributeResponse} "查询成功，返回属性列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "英雄不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/heroes/{hero_id}/attributes [get]
func (h *HeroAttributeHandler) GetComputedAttributes(c echo.Context) error {
	heroID := c.Param("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "英雄ID不能为空")
	}

	attrs, err := h.attrService.GetComputedAttributes(c.Request().Context(), heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*ComputedAttributeResponse, len(attrs))
	for i, attr := range attrs {
		baseValue := 0
		if !attr.BaseValue.IsZero() {
			baseValue = int(attr.BaseValue.Int)
		}

		classBonus := 0
		if !attr.ClassBonus.IsZero() {
			classBonusInt, _ := attr.ClassBonus.Int64()
			classBonus = int(classBonusInt)
		}

		finalValue := 0
		if !attr.FinalValue.IsZero() {
			finalValueInt, _ := attr.FinalValue.Int64()
			finalValue = int(finalValueInt)
		}

		respList[i] = &ComputedAttributeResponse{
			AttributeCode: attr.AttributeCode.String,
			AttributeName: attr.AttributeName.String,
			BaseValue:     baseValue,
			ClassBonus:    classBonus,
			FinalValue:    finalValue,
		}
	}

	return response.EchoOK(c, h.respWriter, respList)
}

