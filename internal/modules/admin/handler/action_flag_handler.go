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

// ActionFlagHandler 动作Flag HTTP 处理器
type ActionFlagHandler struct {
	service    *service.ActionFlagService
	respWriter response.Writer
}

// NewActionFlagHandler 创建动作Flag处理器
func NewActionFlagHandler(db *sql.DB, respWriter response.Writer) *ActionFlagHandler {
	return &ActionFlagHandler{
		service:    service.NewActionFlagService(db),
		respWriter: respWriter,
	}
}

// ==================== HTTP Models ====================

// CreateActionFlagRequest 创建动作Flag请求
type CreateActionFlagRequest struct {
	FlagCode     string `json:"flag_code" validate:"required,max=50" example:"stunned"` // Flag唯一代码
	FlagName     string `json:"flag_name" validate:"required,max=100" example:"眩晕状态"`   // Flag名称
	Category     string `json:"category" example:"control"`                             // Flag分类
	DurationType string `json:"duration_type" example:"turns"`                          // 持续时间类型(instant/turns/permanent)
	Description  string `json:"description" example:"角色无法行动"`                           // Flag描述
	IsActive     bool   `json:"is_active" example:"true"`                               // 是否启用
}

// UpdateActionFlagRequest 更新动作Flag请求
type UpdateActionFlagRequest struct {
	FlagCode     string `json:"flag_code" example:"stunned"`   // Flag唯一代码
	FlagName     string `json:"flag_name" example:"眩晕状态"`      // Flag名称
	Category     string `json:"category" example:"control"`    // Flag分类
	DurationType string `json:"duration_type" example:"turns"` // 持续时间类型(instant/turns/permanent)
	Description  string `json:"description" example:"角色无法行动"`  // Flag描述
	IsActive     bool   `json:"is_active" example:"true"`      // 是否启用
}

// ActionFlagInfo 动作Flag信息响应
type ActionFlagInfo struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // Flag ID
	FlagCode     string `json:"flag_code" example:"stunned"`                       // Flag唯一代码
	FlagName     string `json:"flag_name" example:"眩晕状态"`                          // Flag名称
	Category     string `json:"category" example:"control"`                        // Flag分类
	DurationType string `json:"duration_type" example:"turns"`                     // 持续时间类型(instant/turns/permanent)
	Description  string `json:"description" example:"角色无法行动"`                      // Flag描述
	IsActive     bool   `json:"is_active" example:"true"`                          // 是否启用
	CreatedAt    int64  `json:"created_at" example:"1633024800"`                   // 创建时间戳
	UpdatedAt    int64  `json:"updated_at" example:"1633024800"`                   // 更新时间戳
}

// ==================== HTTP Handlers ====================

// GetActionFlags 获取动作Flag列表
// @Summary 获取动作Flag列表
// @Description 获取动作Flag列表,支持分页和多条件筛选
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param category query string false "分类筛选"
// @Param duration_type query string false "持续时间类型筛选"
// @Param is_active query bool false "是否启用筛选"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=object{list=[]ActionFlagInfo,total=int}}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-flags [get]
func (h *ActionFlagHandler) GetActionFlags(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	params := interfaces.ActionFlagQueryParams{}

	if category := c.QueryParam("category"); category != "" {
		params.Category = &category
	}

	if durationType := c.QueryParam("duration_type"); durationType != "" {
		params.DurationType = &durationType
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
	flags, total, err := h.service.GetActionFlags(ctx, params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	result := make([]ActionFlagInfo, len(flags))
	for i, flag := range flags {
		result[i] = h.convertToActionFlagInfo(flag)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":  result,
		"total": total,
	})
}

// GetActionFlag 获取动作Flag详情
// @Summary 获取动作Flag详情
// @Description 根据ID获取单个动作Flag的详细信息
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param id path string true "动作Flag的UUID"
// @Success 200 {object} response.Response{data=ActionFlagInfo}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作Flag不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-flags/{id} [get]
func (h *ActionFlagHandler) GetActionFlag(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	flag, err := h.service.GetActionFlagByID(ctx, id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionFlagInfo(flag))
}

// CreateActionFlag 创建动作Flag
// @Summary 创建动作Flag
// @Description 创建新的动作Flag,flag_code必须唯一
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param request body CreateActionFlagRequest true "创建动作Flag的请求参数"
// @Success 200 {object} response.Response{data=ActionFlagInfo}
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-flags [post]
func (h *ActionFlagHandler) CreateActionFlag(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateActionFlagRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造动作Flag实体
	flag := &game_config.ActionFlag{
		FlagCode: req.FlagCode,
		FlagName: req.FlagName,
	}

	if req.Category != "" {
		flag.Category.SetValid(req.Category)
	}

	if req.DurationType != "" {
		flag.DurationType.SetValid(req.DurationType)
	}

	if req.Description != "" {
		flag.Description.SetValid(req.Description)
	}

	flag.IsActive.SetValid(req.IsActive)

	// 创建动作Flag
	if err := h.service.CreateActionFlag(ctx, flag); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToActionFlagInfo(flag))
}

// UpdateActionFlag 更新动作Flag
// @Summary 更新动作Flag
// @Description 更新指定ID的动作Flag信息
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param id path string true "动作Flag的UUID"
// @Param request body UpdateActionFlagRequest true "更新动作Flag的请求参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误或验证失败"
// @Failure 404 {object} response.Response "动作Flag不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-flags/{id} [put]
func (h *ActionFlagHandler) UpdateActionFlag(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	var req UpdateActionFlagRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数错误")
	}

	// 构造更新字段
	updates := make(map[string]interface{})

	if req.FlagCode != "" {
		updates["flag_code"] = req.FlagCode
	}
	if req.FlagName != "" {
		updates["flag_name"] = req.FlagName
	}
	updates["category"] = req.Category
	updates["duration_type"] = req.DurationType
	updates["description"] = req.Description
	updates["is_active"] = req.IsActive

	// 更新动作Flag
	if err := h.service.UpdateActionFlag(ctx, id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "动作Flag更新成功",
	})
}

// DeleteActionFlag 删除动作Flag
// @Summary 删除动作Flag
// @Description 软删除指定ID的动作Flag(设置deleted_at字段)
// @Tags 动作系统
// @Accept json
// @Produce json
// @Param id path string true "动作Flag的UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "动作Flag不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth
// @Router /admin/action-flags/{id} [delete]
func (h *ActionFlagHandler) DeleteActionFlag(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if err := h.service.DeleteActionFlag(ctx, id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": "动作Flag删除成功",
	})
}

// ==================== Helper Functions ====================

func (h *ActionFlagHandler) convertToActionFlagInfo(flag *game_config.ActionFlag) ActionFlagInfo {
	info := ActionFlagInfo{
		ID:       flag.ID,
		FlagCode: flag.FlagCode,
		FlagName: flag.FlagName,
	}

	if flag.Category.Valid {
		info.Category = flag.Category.String
	}

	if flag.DurationType.Valid {
		info.DurationType = flag.DurationType.String
	}

	if flag.Description.Valid {
		info.Description = flag.Description.String
	}

	if flag.IsActive.Valid {
		info.IsActive = flag.IsActive.Bool
	}

	if flag.CreatedAt.Valid {
		info.CreatedAt = flag.CreatedAt.Time.Unix()
	}

	if flag.UpdatedAt.Valid {
		info.UpdatedAt = flag.UpdatedAt.Time.Unix()
	}

	return info
}
