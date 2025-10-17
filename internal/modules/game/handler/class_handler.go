package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

// ClassHandler handles class HTTP requests
type ClassHandler struct {
	classService *service.ClassService
	respWriter   response.Writer
}

// NewClassHandler creates a new class handler
func NewClassHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *ClassHandler {
	return &ClassHandler{
		classService: serviceContainer.GetClassService(),
		respWriter:   respWriter,
	}
}

// ==================== HTTP Response Models ====================

// ClassResponse HTTP class response
type ClassResponse struct {
	ID              string  `json:"id" example:"class-warrior-001"`                                        // 职业ID
	ClassCode       string  `json:"class_code" example:"WARRIOR"`                                          // 职业代码
	ClassName       string  `json:"class_name" example:"战士"`                                               // 职业名称
	Tier            string  `json:"tier" example:"basic" enums:"basic,advanced,master"`                    // 职业等阶：basic=基础职业，advanced=进阶职业，master=大师职业
	Description     *string `json:"description,omitempty" example:"擅长近战和物理攻击的职业"`                          // 职业描述（可选）
	LoreText        *string `json:"lore_text,omitempty" example:"在古老的战场上磨练技艺"`                            // 职业背景故事（可选）
	Specialty       *string `json:"specialty,omitempty" example:"高护甲、高生命值、近战控制"`                            // 职业特长（可选）
	Playstyle       *string `json:"playstyle,omitempty" example:"前排坦克，保护队友"`                              // 玩法风格（可选）
	PromotionCount  int     `json:"promotion_count" example:"0"`                                           // 转职次数加成
	Icon            *string `json:"icon,omitempty" example:"https://example.com/warrior.png"`              // 职业图标URL（可选）
	Color           *string `json:"color,omitempty" example:"#FF5733"`                                     // 职业代表颜色（可选）
	IsActive        bool    `json:"is_active" example:"true"`                                              // 是否启用
	IsVisible       bool    `json:"is_visible" example:"true"`                                             // 是否在UI中显示
	DisplayOrder    int     `json:"display_order" example:"1"`                                             // 显示顺序
	CreatedAt       string  `json:"created_at" example:"2025-10-17 10:30:00"`                              // 创建时间
	UpdatedAt       string  `json:"updated_at" example:"2025-10-17 12:30:00"`                              // 更新时间
}

// AdvancementOptionResponse 进阶选项响应
type AdvancementOptionResponse struct {
	FromClassID        string         `json:"from_class_id"`
	ToClassID          string         `json:"to_class_id"`
	ToClassName        string         `json:"to_class_name"`
	RequiredLevel      int            `json:"required_level"`
	RequiredHonor      int            `json:"required_honor"`
	RequiredAttributes map[string]int `json:"required_attributes,omitempty"`
	RequiredSkills     []string       `json:"required_skills,omitempty"`
	RequiredItems      map[string]int `json:"required_items,omitempty"`
	DisplayOrder       int            `json:"display_order"`
}

// ==================== HTTP Handlers ====================

// GetBasicClasses handles getting basic class list (for character creation)
// @Summary 获取基础职业列表
// @Description 获取可供创建角色时选择的基础职业列表（仅 tier=basic）
// @Tags 职业
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]ClassResponse} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/classes/basic [get]
func (h *ClassHandler) GetBasicClasses(c echo.Context) error {
	// 构建查询参数：仅查询基础职业
	isActive := true
	isVisible := true
	params := interfaces.ClassQueryParams{
		Page:      1,
		PageSize:  100, // 基础职业数量不会很多
		Tier:      "basic",
		IsActive:  &isActive,
		IsVisible: &isVisible,
		SortBy:    "display_order",
		SortDir:   "ASC",
	}

	// 调用 Service
	classes, _, err := h.classService.GetClassList(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*ClassResponse, len(classes))
	for i, class := range classes {
		promotionCount := 0
		if !class.PromotionCount.IsZero() {
			promotionCount = int(class.PromotionCount.Int16)
		}

		displayOrder := 0
		if !class.DisplayOrder.IsZero() {
			displayOrder = int(class.DisplayOrder.Int16)
		}

		resp := &ClassResponse{
			ID:             class.ID,
			ClassCode:      class.ClassCode,
			ClassName:      class.ClassName,
			Tier:           class.Tier,
			PromotionCount: promotionCount,
			IsActive:       class.IsActive.Bool,
			IsVisible:      class.IsVisible.Bool,
			DisplayOrder:   displayOrder,
			CreatedAt:      class.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      class.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if !class.Description.IsZero() {
			desc := class.Description.String
			resp.Description = &desc
		}

		if !class.LoreText.IsZero() {
			loreText := class.LoreText.String
			resp.LoreText = &loreText
		}

		if !class.Specialty.IsZero() {
			specialty := class.Specialty.String
			resp.Specialty = &specialty
		}

		if !class.Playstyle.IsZero() {
			playstyle := class.Playstyle.String
			resp.Playstyle = &playstyle
		}

		if !class.Icon.IsZero() {
			icon := class.Icon.String
			resp.Icon = &icon
		}

		if !class.Color.IsZero() {
			color := class.Color.String
			resp.Color = &color
		}

		respList[i] = resp
	}

	return response.EchoOK(c, h.respWriter, respList)
}

// GetClasses handles getting class list
// @Summary 获取职业列表
// @Description 获取游戏职业列表，支持分页和筛选
// @Tags 职业
// @Accept json
// @Produce json
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认20）"
// @Param tier query string false "职业等级 (basic, advanced, master)"
// @Param is_active query bool false "是否激活"
// @Param is_visible query bool false "是否可见"
// @Param sort_by query string false "排序字段 (class_name, tier, display_order, created_at)"
// @Param sort_dir query string false "排序方向 (ASC, DESC)"
// @Success 200 {object} response.Response{data=object{list=[]ClassResponse,total=int64,page=int,page_size=int}} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/classes [get]
func (h *ClassHandler) GetClasses(c echo.Context) error {
	// 解析查询参数
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询参数
	params := interfaces.ClassQueryParams{
		Page:     page,
		PageSize: pageSize,
		Tier:     c.QueryParam("tier"),
		SortBy:   c.QueryParam("sort_by"),
		SortDir:  c.QueryParam("sort_dir"),
	}

	// 解析布尔值参数
	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive, _ := strconv.ParseBool(isActiveStr)
		params.IsActive = &isActive
	}

	if isVisibleStr := c.QueryParam("is_visible"); isVisibleStr != "" {
		isVisible, _ := strconv.ParseBool(isVisibleStr)
		params.IsVisible = &isVisible
	}

	// 调用 Service
	classes, total, err := h.classService.GetClassList(c.Request().Context(), params)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*ClassResponse, len(classes))
	for i, class := range classes {
		promotionCount := 0
		if !class.PromotionCount.IsZero() {
			promotionCount = int(class.PromotionCount.Int16)
		}

		displayOrder := 0
		if !class.DisplayOrder.IsZero() {
			displayOrder = int(class.DisplayOrder.Int16)
		}

		resp := &ClassResponse{
			ID:             class.ID,
			ClassCode:      class.ClassCode,
			ClassName:      class.ClassName,
			Tier:           class.Tier,
			PromotionCount: promotionCount,
			IsActive:       class.IsActive.Bool,
			IsVisible:      class.IsVisible.Bool,
			DisplayOrder:   displayOrder,
			CreatedAt:      class.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      class.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if !class.Description.IsZero() {
			desc := class.Description.String
			resp.Description = &desc
		}

		if !class.LoreText.IsZero() {
			loreText := class.LoreText.String
			resp.LoreText = &loreText
		}

		if !class.Specialty.IsZero() {
			specialty := class.Specialty.String
			resp.Specialty = &specialty
		}

		if !class.Playstyle.IsZero() {
			playstyle := class.Playstyle.String
			resp.Playstyle = &playstyle
		}

		if !class.Icon.IsZero() {
			icon := class.Icon.String
			resp.Icon = &icon
		}

		if !class.Color.IsZero() {
			color := class.Color.String
			resp.Color = &color
		}

		respList[i] = resp
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":      respList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetClass handles getting class details
// @Summary 获取职业详情
// @Description 根据职业ID获取职业详细信息
// @Tags 职业
// @Accept json
// @Produce json
// @Param class_id path string true "职业ID"
// @Success 200 {object} response.Response{data=ClassResponse} "获取成功"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/classes/{class_id} [get]
func (h *ClassHandler) GetClass(c echo.Context) error {
	classID := c.Param("class_id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	class, err := h.classService.GetClassByID(c.Request().Context(), classID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	promotionCount := 0
	if !class.PromotionCount.IsZero() {
		promotionCount = int(class.PromotionCount.Int16)
	}

	displayOrder := 0
	if !class.DisplayOrder.IsZero() {
		displayOrder = int(class.DisplayOrder.Int16)
	}

	resp := &ClassResponse{
		ID:             class.ID,
		ClassCode:      class.ClassCode,
		ClassName:      class.ClassName,
		Tier:           class.Tier,
		PromotionCount: promotionCount,
		IsActive:       class.IsActive.Bool,
		IsVisible:      class.IsVisible.Bool,
		DisplayOrder:   displayOrder,
		CreatedAt:      class.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      class.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !class.Description.IsZero() {
		desc := class.Description.String
		resp.Description = &desc
	}

	if !class.LoreText.IsZero() {
		loreText := class.LoreText.String
		resp.LoreText = &loreText
	}

	if !class.Specialty.IsZero() {
		specialty := class.Specialty.String
		resp.Specialty = &specialty
	}

	if !class.Playstyle.IsZero() {
		playstyle := class.Playstyle.String
		resp.Playstyle = &playstyle
	}

	if !class.Icon.IsZero() {
		icon := class.Icon.String
		resp.Icon = &icon
	}

	if !class.Color.IsZero() {
		color := class.Color.String
		resp.Color = &color
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetAdvancementOptions handles getting advancement options for a class
// @Summary 获取职业可进阶选项
// @Description 获取指定职业的所有可进阶选项及其要求
// @Tags 职业
// @Accept json
// @Produce json
// @Param class_id path string true "职业ID"
// @Success 200 {object} response.Response{data=[]AdvancementOptionResponse} "获取成功"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /game/classes/{class_id}/advancement-options [get]
func (h *ClassHandler) GetAdvancementOptions(c echo.Context) error {
	classID := c.Param("class_id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	// 获取进阶选项
	options, err := h.classService.GetAdvancementOptions(c.Request().Context(), classID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为 HTTP 响应
	respList := make([]*AdvancementOptionResponse, len(options))
	for i, opt := range options {
		// 获取目标职业名称
		toClass, err := h.classService.GetClassByID(c.Request().Context(), opt.ToClassID)
		toClassName := ""
		if err == nil && toClass != nil {
			toClassName = toClass.ClassName
		}

		displayOrder := 0
		if !opt.DisplayOrder.IsZero() {
			displayOrder = int(opt.DisplayOrder.Int16)
		}

		resp := &AdvancementOptionResponse{
			FromClassID:   opt.FromClassID,
			ToClassID:     opt.ToClassID,
			ToClassName:   toClassName,
			RequiredLevel: opt.RequiredLevel,
			RequiredHonor: opt.RequiredHonor,
			DisplayOrder:  displayOrder,
		}

		// 解析 JSONB 字段
		if !opt.RequiredAttributes.IsZero() {
			var attrs map[string]int
			if err := opt.RequiredAttributes.Unmarshal(&attrs); err == nil {
				resp.RequiredAttributes = attrs
			}
		}

		if !opt.RequiredSkills.IsZero() {
			var skills []string
			if err := opt.RequiredSkills.Unmarshal(&skills); err == nil {
				resp.RequiredSkills = skills
			}
		}

		if !opt.RequiredItems.IsZero() {
			var items map[string]int
			if err := opt.RequiredItems.Unmarshal(&items); err == nil {
				resp.RequiredItems = items
			}
		}

		respList[i] = resp
	}

	return response.EchoOK(c, h.respWriter, respList)
}
