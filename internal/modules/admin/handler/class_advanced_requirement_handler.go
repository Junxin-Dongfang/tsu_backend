package handler

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// ClassAdvancedRequirementHandler 职业进阶要求处理器
type ClassAdvancedRequirementHandler struct {
	service    *service.ClassAdvancedRequirementService
	respWriter response.Writer
}

// NewClassAdvancedRequirementHandler 创建职业进阶要求处理器
func NewClassAdvancedRequirementHandler(db *sql.DB, respWriter response.Writer) *ClassAdvancedRequirementHandler {
	return &ClassAdvancedRequirementHandler{
		service:    service.NewClassAdvancedRequirementService(db),
		respWriter: respWriter,
	}
}

// AdvancedRequirementInfo HTTP响应结构
type AdvancedRequirementInfo struct {
	ID                     string          `json:"id"`
	FromClassID            string          `json:"from_class_id"`
	ToClassID              string          `json:"to_class_id"`
	RequiredLevel          int             `json:"required_level"`
	RequiredHonor          int             `json:"required_honor"`
	RequiredJobChangeCount int             `json:"required_job_change_count"`
	RequiredAttributes     json.RawMessage `json:"required_attributes,omitempty"`
	RequiredSkills         json.RawMessage `json:"required_skills,omitempty"`
	RequiredItems          json.RawMessage `json:"required_items,omitempty"`
	IsActive               *bool           `json:"is_active,omitempty"`
	DisplayOrder           *int16          `json:"display_order,omitempty"`
	CreatedAt              int64           `json:"created_at"`
	UpdatedAt              int64           `json:"updated_at"`
}

// CreateAdvancedRequirementRequest 创建进阶要求请求
type CreateAdvancedRequirementRequest struct {
	FromClassID            string          `json:"from_class_id" validate:"required,uuid"`
	ToClassID              string          `json:"to_class_id" validate:"required,uuid"`
	RequiredLevel          int             `json:"required_level" validate:"required,min=1,max=100"`
	RequiredHonor          int             `json:"required_honor" validate:"min=0"`
	RequiredJobChangeCount int             `json:"required_job_change_count" validate:"min=0"`
	RequiredAttributes     json.RawMessage `json:"required_attributes"`
	RequiredSkills         json.RawMessage `json:"required_skills"`
	RequiredItems          json.RawMessage `json:"required_items"`
	IsActive               *bool           `json:"is_active"`
	DisplayOrder           *int16          `json:"display_order"`
}

// UpdateAdvancedRequirementRequest 更新进阶要求请求
type UpdateAdvancedRequirementRequest struct {
	FromClassID            *string          `json:"from_class_id" validate:"omitempty,uuid"`
	ToClassID              *string          `json:"to_class_id" validate:"omitempty,uuid"`
	RequiredLevel          *int             `json:"required_level" validate:"omitempty,min=1,max=100"`
	RequiredHonor          *int             `json:"required_honor" validate:"omitempty,min=0"`
	RequiredJobChangeCount *int             `json:"required_job_change_count" validate:"omitempty,min=0"`
	RequiredAttributes     *json.RawMessage `json:"required_attributes"`
	RequiredSkills         *json.RawMessage `json:"required_skills"`
	RequiredItems          *json.RawMessage `json:"required_items"`
	IsActive               *bool            `json:"is_active"`
	DisplayOrder           *int16           `json:"display_order"`
}

// BatchCreateRequest 批量创建请求
type BatchCreateRequest struct {
	Requirements []CreateAdvancedRequirementRequest `json:"requirements" validate:"required,min=1,max=100,dive"`
}

// ListResponse 列表响应
type ListAdvancedRequirementsResponse struct {
	Requirements []AdvancedRequirementInfo `json:"requirements"`
	Total        int64                     `json:"total"`
	Page         int                       `json:"page"`
	PageSize     int                       `json:"page_size"`
	TotalPages   int                       `json:"total_pages"`
}

// GetAdvancedRequirements godoc
// @Summary      获取职业进阶要求列表
// @Description  获取系统中的所有职业进阶要求，支持按源职业、目标职业筛选。进阶要求定义了职业间的转换条件，包括等级、属性、技能、物品等多维度要求
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        from_class_id  query     string  false  "源职业ID，筛选从指定职业开始的进阶路径"  Format(uuid)  Example(30000000-0000-0000-0000-000000000001)
// @Param        to_class_id    query     string  false  "目标职业ID，筛选进阶到指定职业的路径"  Format(uuid)  Example(30000000-0000-0000-0000-000000000010)
// @Param        is_active      query     bool    false  "是否为激活状态的进阶要求"  Example(true)
// @Param        page           query     int     false  "页码，从1开始"  Default(1)  minimum(1)  maximum(1000)  Example(1)
// @Param        page_size      query     int     false  "每页记录数"  Default(20)  minimum(1)  maximum(100)  Example(20)
// @Param        sort_by        query     string  false  "排序字段选择"  Enums(display_order, required_level, created_at)  Example(display_order)
// @Param        sort_dir       query     string  false  "排序方向"  Enums(asc, desc)  Example(asc)
// @Success      200  {object}  response.Response{data=ListAdvancedRequirementsResponse}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/advancement-requirements [get]
func (h *ClassAdvancedRequirementHandler) GetAdvancedRequirements(c echo.Context) error {
	params := interfaces.ListAdvancedRequirementsParams{
		Page:     1,
		PageSize: 20,
		SortBy:   "display_order",
		SortDir:  "asc",
	}

	// 解析查询参数
	if fromClassID := c.QueryParam("from_class_id"); fromClassID != "" {
		params.FromClassID = &fromClassID
	}
	if toClassID := c.QueryParam("to_class_id"); toClassID != "" {
		params.ToClassID = &toClassID
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		active := isActive == "true"
		params.IsActive = &active
	}
	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}
	if pageSize := c.QueryParam("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}
	if sortBy := c.QueryParam("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}
	if sortDir := c.QueryParam("sort_dir"); sortDir != "" {
		params.SortDir = sortDir
	}

	requirements, total, err := h.service.List(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	infos := make([]AdvancedRequirementInfo, len(requirements))
	for i, req := range requirements {
		infos[i] = h.convertToInfo(req)
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	resp := ListAdvancedRequirementsResponse{
		Requirements: infos,
		Total:        total,
		Page:         params.Page,
		PageSize:     params.PageSize,
		TotalPages:   totalPages,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetAdvancedRequirement godoc
// @Summary      获取单个职业进阶要求
// @Description  根据ID获取职业进阶要求的详细信息
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "进阶要求ID"  Format(uuid)
// @Success      200  {object}  response.Response{data=AdvancedRequirementInfo}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/advancement-requirements/{id} [get]
func (h *ClassAdvancedRequirementHandler) GetAdvancedRequirement(c echo.Context) error {
	id := c.Param("id")

	requirement, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(requirement))
}

// CreateAdvancedRequirement godoc
// @Summary      创建职业进阶要求
// @Description  创建新的职业进阶要求配置
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAdvancedRequirementRequest  true  "创建进阶要求请求"
// @Success      200      {object}  response.Response{data=AdvancedRequirementInfo}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /admin/advancement-requirements [post]
func (h *ClassAdvancedRequirementHandler) CreateAdvancedRequirement(c echo.Context) error {
	var req CreateAdvancedRequirementRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	input := service.CreateAdvancedRequirementInput{
		FromClassID:            req.FromClassID,
		ToClassID:              req.ToClassID,
		RequiredLevel:          req.RequiredLevel,
		RequiredHonor:          req.RequiredHonor,
		RequiredJobChangeCount: req.RequiredJobChangeCount,
		RequiredAttributes:     req.RequiredAttributes,
		RequiredSkills:         req.RequiredSkills,
		RequiredItems:          req.RequiredItems,
		IsActive:               true,
		DisplayOrder:           0,
	}

	if req.IsActive != nil {
		input.IsActive = *req.IsActive
	}
	if req.DisplayOrder != nil {
		input.DisplayOrder = *req.DisplayOrder
	}

	requirement, err := h.service.Create(c.Request().Context(), input)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, h.convertToInfo(requirement))
}

// UpdateAdvancedRequirement godoc
// @Summary      更新职业进阶要求
// @Description  更新已存在的职业进阶要求配置
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        id       path      string                          true  "进阶要求ID"  Format(uuid)
// @Param        request  body      UpdateAdvancedRequirementRequest  true  "更新进阶要求请求"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /admin/advancement-requirements/{id} [put]
func (h *ClassAdvancedRequirementHandler) UpdateAdvancedRequirement(c echo.Context) error {
	id := c.Param("id")

	var req UpdateAdvancedRequirementRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	updates := make(map[string]interface{})
	if req.FromClassID != nil {
		updates["from_class_id"] = *req.FromClassID
	}
	if req.ToClassID != nil {
		updates["to_class_id"] = *req.ToClassID
	}
	if req.RequiredLevel != nil {
		updates["required_level"] = *req.RequiredLevel
	}
	if req.RequiredHonor != nil {
		updates["required_honor"] = *req.RequiredHonor
	}
	if req.RequiredJobChangeCount != nil {
		updates["required_job_change_count"] = *req.RequiredJobChangeCount
	}
	if req.RequiredAttributes != nil {
		updates["required_attributes"] = []byte(*req.RequiredAttributes)
	}
	if req.RequiredSkills != nil {
		updates["required_skills"] = []byte(*req.RequiredSkills)
	}
	if req.RequiredItems != nil {
		updates["required_items"] = []byte(*req.RequiredItems)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = int(*req.DisplayOrder)
	}

	if err := h.service.Update(c.Request().Context(), id, updates); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "进阶要求更新成功"})
}

// DeleteAdvancedRequirement godoc
// @Summary      删除职业进阶要求
// @Description  删除指定的职业进阶要求（软删除）
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "进阶要求ID"  Format(uuid)
// @Success      200  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/advancement-requirements/{id} [delete]
func (h *ClassAdvancedRequirementHandler) DeleteAdvancedRequirement(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, map[string]string{"message": "进阶要求删除成功"})
}

// BatchCreateAdvancedRequirements godoc
// @Summary      批量创建职业进阶要求
// @Description  批量创建多个职业进阶要求配置
// @Tags         职业管理
// @Accept       json
// @Produce      json
// @Param        request  body      BatchCreateRequest  true  "批量创建请求"
// @Success      200      {object}  response.Response{data=[]AdvancedRequirementInfo}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /admin/advancement-requirements/batch [post]
func (h *ClassAdvancedRequirementHandler) BatchCreateAdvancedRequirements(c echo.Context) error {
	var req BatchCreateRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	inputs := make([]service.CreateAdvancedRequirementInput, len(req.Requirements))
	for i, r := range req.Requirements {
		inputs[i] = service.CreateAdvancedRequirementInput{
			FromClassID:            r.FromClassID,
			ToClassID:              r.ToClassID,
			RequiredLevel:          r.RequiredLevel,
			RequiredHonor:          r.RequiredHonor,
			RequiredJobChangeCount: r.RequiredJobChangeCount,
			RequiredAttributes:     r.RequiredAttributes,
			RequiredSkills:         r.RequiredSkills,
			RequiredItems:          r.RequiredItems,
			IsActive:               true,
			DisplayOrder:           0,
		}
		if r.IsActive != nil {
			inputs[i].IsActive = *r.IsActive
		}
		if r.DisplayOrder != nil {
			inputs[i].DisplayOrder = *r.DisplayOrder
		}
	}

	requirements, err := h.service.BatchCreate(c.Request().Context(), inputs)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	infos := make([]AdvancedRequirementInfo, len(requirements))
	for i, req := range requirements {
		infos[i] = h.convertToInfo(req)
	}

	return response.EchoOK(c, h.respWriter, infos)
}

// convertToInfo 转换为响应信息
func (h *ClassAdvancedRequirementHandler) convertToInfo(req *game_config.ClassAdvancedRequirement) AdvancedRequirementInfo {
	info := AdvancedRequirementInfo{
		ID:                     req.ID,
		FromClassID:            req.FromClassID,
		ToClassID:              req.ToClassID,
		RequiredLevel:          req.RequiredLevel,
		RequiredHonor:          req.RequiredHonor,
		RequiredJobChangeCount: req.RequiredJobChangeCount,
		CreatedAt:              req.CreatedAt.Unix(),
		UpdatedAt:              req.UpdatedAt.Unix(),
	}

	if req.RequiredAttributes.Valid {
		jsonBytes, _ := req.RequiredAttributes.MarshalJSON()
		info.RequiredAttributes = json.RawMessage(jsonBytes)
	}

	if req.RequiredSkills.Valid {
		jsonBytes, _ := req.RequiredSkills.MarshalJSON()
		info.RequiredSkills = json.RawMessage(jsonBytes)
	}

	if req.RequiredItems.Valid {
		jsonBytes, _ := req.RequiredItems.MarshalJSON()
		info.RequiredItems = json.RawMessage(jsonBytes)
	}

	if req.IsActive.Valid {
		isActive := req.IsActive.Bool
		info.IsActive = &isActive
	}

	if req.DisplayOrder.Valid {
		displayOrder := req.DisplayOrder.Int16
		info.DisplayOrder = &displayOrder
	}

	return info
}
