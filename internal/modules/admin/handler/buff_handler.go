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

// BuffHandler Buff HTTP 处理器
type BuffHandler struct {
	service    *service.BuffService
	respWriter response.Writer
}

// NewBuffHandler 创建Buff处理器
func NewBuffHandler(db *sql.DB, respWriter response.Writer) *BuffHandler {
	return &BuffHandler{
		service:    service.NewBuffService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateBuffRequest 创建Buff请求
type CreateBuffRequest struct {
	BuffCode    string `json:"buff_code" validate:"required,max=50" example:"bless"`    // Buff唯一代码
	BuffName    string `json:"buff_name" validate:"required,max=100" example:"祝福术"`     // Buff名称
	BuffType    string `json:"buff_type" validate:"required,max=50" example:"positive"` // Buff类型(positive/negative/neutral)
	Category    string `json:"category" example:"blessing"`                             // Buff分类
	Description string `json:"description" example:"提升攻击力和防御力"`                         // Buff描述
	IsActive    bool   `json:"is_active" example:"true"`                                // 是否启用
}

// UpdateBuffRequest 更新Buff请求
type UpdateBuffRequest struct {
	BuffCode    string `json:"buff_code" example:"bless"`       // Buff唯一代码
	BuffName    string `json:"buff_name" example:"祝福术"`         // Buff名称
	BuffType    string `json:"buff_type" example:"positive"`    // Buff类型(positive/negative/neutral)
	Category    string `json:"category" example:"blessing"`     // Buff分类
	Description string `json:"description" example:"提升攻击力和防御力"` // Buff描述
	IsActive    bool   `json:"is_active" example:"true"`        // 是否启用
}

// BuffInfo Buff信息响应
type BuffInfo struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // Buff ID
	BuffCode    string `json:"buff_code" example:"bless"`                         // Buff唯一代码
	BuffName    string `json:"buff_name" example:"祝福术"`                           // Buff名称
	BuffType    string `json:"buff_type" example:"positive"`                      // Buff类型(positive/negative/neutral)
	Category    string `json:"category" example:"blessing"`                       // Buff分类
	Description string `json:"description" example:"提升攻击力和防御力"`                   // Buff描述
	IsActive    bool   `json:"is_active" example:"true"`                          // 是否启用
	CreatedAt   int64  `json:"created_at" example:"1633024800"`                   // 创建时间戳
	UpdatedAt   int64  `json:"updated_at" example:"1633024800"`                   // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetBuffs 获取Buff列表
// @Summary 获取Buff列表
// @Description 分页查询Buff列表，支持按Buff类型、分类和启用状态筛选
// @Tags Buff系统
// @Accept json
// @Produce json
// @Param buff_type query string false "Buff类型，例如: positive, negative, neutral"
// @Param category query string false "Buff分类"
// @Param is_active query bool false "是否启用，true或false"
// @Param limit query int false "每页数量，默认10"
// @Param offset query int false "偏移量，默认0"
// @Success 200 {object} response.Response{data=map[string]interface{}} "成功返回Buff列表和总数"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs [get]
// @Security BearerAuth
func (h *BuffHandler) GetBuffs(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.BuffQueryParams{}

	if buffType := c.QueryParam("buff_type"); buffType != "" {
		params.BuffType = &buffType
	}

	if category := c.QueryParam("category"); category != "" {
		params.Category = &category
	}

	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive, _ := strconv.ParseBool(isActiveStr)
		params.IsActive = &isActive
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		params.Limit, _ = strconv.Atoi(limitStr)
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		params.Offset, _ = strconv.Atoi(offsetStr)
	}

	// 查询列表
	buffs, total, err := h.service.GetBuffs(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]BuffInfo, len(buffs))
	for i, buff := range buffs {
		result[i] = h.convertToBuffInfo(buff)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetBuff 获取Buff详情
// @Summary 获取Buff详情
// @Description 根据Buff ID获取Buff的详细信息
// @Tags Buff系统
// @Accept json
// @Produce json
// @Param id path string true "Buff ID (UUID格式)"
// @Success 200 {object} response.Response{data=BuffInfo} "成功返回Buff详情"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "Buff不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{id} [get]
// @Security BearerAuth
func (h *BuffHandler) GetBuff(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	buff, err := h.service.GetBuffByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToBuffInfo(buff))
}

// CreateBuff 创建Buff
// @Summary 创建Buff
// @Description 创建新的游戏Buff（增益或减益效果）
// @Tags Buff系统
// @Accept json
// @Produce json
// @Param request body CreateBuffRequest true "创建Buff请求参数"
// @Success 200 {object} response.Response{data=BuffInfo} "成功返回创建的Buff信息"
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 409 {object} response.Response "Buff代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs [post]
// @Security BearerAuth
func (h *BuffHandler) CreateBuff(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateBuffRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造Buff实体
	buff := &game_config.Buff{
		BuffCode: req.BuffCode,
		BuffName: req.BuffName,
		BuffType: req.BuffType,
	}

	if req.Category != "" {
		buff.Category.SetValid(req.Category)
	}

	if req.Description != "" {
		buff.Description.SetValid(req.Description)
	}

	buff.IsActive.SetValid(req.IsActive)

	// 创建Buff
	if err := h.service.CreateBuff(ctx, buff); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToBuffInfo(buff))
}

// UpdateBuff 更新Buff
// @Summary 更新Buff
// @Description 更新已有Buff的信息
// @Tags Buff系统
// @Accept json
// @Produce json
// @Param id path string true "Buff ID (UUID格式)"
// @Param request body UpdateBuffRequest true "更新Buff请求参数"
// @Success 200 {object} response.Response "成功返回更新结果"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "Buff不存在"
// @Failure 409 {object} response.Response "Buff代码已存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{id} [put]
// @Security BearerAuth
func (h *BuffHandler) UpdateBuff(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateBuffRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.BuffCode != "" {
		updates["buff_code"] = req.BuffCode
	}
	if req.BuffName != "" {
		updates["buff_name"] = req.BuffName
	}
	if req.BuffType != "" {
		updates["buff_type"] = req.BuffType
	}
	updates["category"] = req.Category
	updates["description"] = req.Description
	updates["is_active"] = req.IsActive

	// 更新Buff
	if err := h.service.UpdateBuff(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "Buff更新成功",
	})
}

// DeleteBuff 删除Buff
// @Summary 删除Buff
// @Description 软删除指定的Buff（设置deleted_at字段）
// @Tags Buff系统
// @Accept json
// @Produce json
// @Param id path string true "Buff ID (UUID格式)"
// @Success 200 {object} response.Response "成功返回删除结果"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "Buff不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/buffs/{id} [delete]
// @Security BearerAuth
func (h *BuffHandler) DeleteBuff(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteBuff(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "Buff删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *BuffHandler) convertToBuffInfo(buff *game_config.Buff) BuffInfo {
	info := BuffInfo{
		ID:       buff.ID,
		BuffCode: buff.BuffCode,
		BuffName: buff.BuffName,
		BuffType: buff.BuffType,
	}

	if buff.Category.Valid {
		info.Category = buff.Category.String
	}

	if buff.Description.Valid {
		info.Description = buff.Description.String
	}

	if buff.IsActive.Valid {
		info.IsActive = buff.IsActive.Bool
	}

	if buff.CreatedAt.Valid {
		info.CreatedAt = buff.CreatedAt.Time.Unix()
	}

	if buff.UpdatedAt.Valid {
		info.UpdatedAt = buff.UpdatedAt.Time.Unix()
	}

	return info
}
