package handler

import (
	"database/sql"
	"fmt"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

// ClassHandler 职业管理处理器
type ClassHandler struct {
	classService *service.ClassService
	advService   *service.ClassAdvancedRequirementService
	respWriter   response.Writer
}

// NewClassHandler 创建职业管理处理器
func NewClassHandler(db *sql.DB, respWriter response.Writer) *ClassHandler {
	return &ClassHandler{
		classService: service.NewClassService(db),
		advService:   service.NewClassAdvancedRequirementService(db),
		respWriter:   respWriter,
	}
}

// ClassInfo HTTP 响应结构
type ClassInfo struct {
	ID             string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"` // 职业ID
	ClassCode      string  `json:"class_code" example:"warrior"`                      // 职业唯一代码
	ClassName      string  `json:"class_name" example:"战士"`                           // 职业名称
	Description    *string `json:"description,omitempty" example:"强大的近战职业"`           // 职业描述(可选)
	LoreText       *string `json:"lore_text,omitempty" example:"勇敢的战士"`               // 背景故事(可选)
	Specialty      *string `json:"specialty,omitempty" example:"近战攻击"`                // 职业特长(可选)
	Playstyle      *string `json:"playstyle,omitempty" example:"高伤害、高防御"`             // 游戏风格(可选)
	Tier           string  `json:"tier" example:"basic"`                              // 职业层级(basic/advanced/elite/legendary/mythic)
	PromotionCount *int16  `json:"promotion_count,omitempty" example:"0"`             // 转职次数加成，默认0
	Icon           *string `json:"icon,omitempty" example:"warrior_icon.png"`         // 职业图标(可选)
	Color          *string `json:"color,omitempty" example:"#FF0000"`                 // 职业颜色(可选)
	IsActive       *bool   `json:"is_active,omitempty" example:"true"`                // 是否启用(可选)
	IsVisible      *bool   `json:"is_visible,omitempty" example:"true"`               // 是否可见(可选)
	DisplayOrder   *int16  `json:"display_order,omitempty" example:"1"`               // 显示顺序(可选)
	CreatedAt      int64   `json:"created_at" example:"1633024800"`                   // 创建时间戳
	UpdatedAt      int64   `json:"updated_at" example:"1633024800"`                   // 更新时间戳
}

// ClassListResponse 职业列表响应
type ClassListResponse struct {
	Classes    []ClassInfo `json:"classes"`                 // 职业列表
	Total      int64       `json:"total" example:"100"`     // 总数
	Page       int         `json:"page" example:"1"`        // 当前页码
	PageSize   int         `json:"page_size" example:"20"`  // 每页数量
	TotalPages int         `json:"total_pages" example:"5"` // 总页数
}

// CreateClassRequest 创建职业请求
type CreateClassRequest struct {
	ClassCode      string  `json:"class_code" validate:"required,min=2,max=32" example:"warrior"`                        // 职业唯一代码(必需,2-32字符)
	ClassName      string  `json:"class_name" validate:"required,min=2,max=64" example:"战士"`                             // 职业名称(必需,2-64字符)
	Description    *string `json:"description" example:"强大的近战职业"`                                                        // 职业描述(可选)
	LoreText       *string `json:"lore_text" example:"勇敢的战士"`                                                            // 背景故事(可选)
	Specialty      *string `json:"specialty" example:"近战攻击"`                                                             // 职业特长(可选)
	Playstyle      *string `json:"playstyle" example:"高伤害、高防御"`                                                          // 游戏风格(可选)
	Tier           string  `json:"tier" validate:"required,oneof=basic advanced elite legendary mythic" example:"basic"` // 职业层级(必需,只能是basic/advanced/elite/legendary/mythic之一)
	PromotionCount *int16  `json:"promotion_count" example:"0"`                                                          // 转职次数加成，默认0
	Icon           *string `json:"icon" example:"warrior_icon.png"`                                                      // 职业图标(可选)
	Color          *string `json:"color" example:"#FF0000"`                                                              // 职业颜色(可选)
	IsActive       *bool   `json:"is_active" example:"true"`                                                             // 是否启用(可选)
	IsVisible      *bool   `json:"is_visible" example:"true"`                                                            // 是否可见(可选)
	DisplayOrder   *int16  `json:"display_order" example:"1"`                                                            // 显示顺序(可选)
}

// UpdateClassRequest 更新职业请求
type UpdateClassRequest struct {
	ClassCode      *string `json:"class_code" validate:"omitempty,min=2,max=32" example:"warrior"`                        // 职业唯一代码(可选,2-32字符)
	ClassName      *string `json:"class_name" validate:"omitempty,min=2,max=64" example:"战士"`                             // 职业名称(可选,2-64字符)
	Description    *string `json:"description" example:"强大的近战职业"`                                                         // 职业描述(可选)
	LoreText       *string `json:"lore_text" example:"勇敢的战士"`                                                             // 背景故事(可选)
	Specialty      *string `json:"specialty" example:"近战攻击"`                                                              // 职业特长(可选)
	Playstyle      *string `json:"playstyle" example:"高伤害、高防御"`                                                           // 游戏风格(可选)
	Tier           *string `json:"tier" validate:"omitempty,oneof=basic advanced elite legendary mythic" example:"basic"` // 职业层级(可选,只能是basic/advanced/elite/legendary/mythic之一)
	PromotionCount *int16  `json:"promotion_count" example:"0"`                                                           // 转职次数加成，默认0
	Icon           *string `json:"icon" example:"warrior_icon.png"`                                                       // 职业图标(可选)
	Color          *string `json:"color" example:"#FF0000"`                                                               // 职业颜色(可选)
	IsActive       *bool   `json:"is_active" example:"true"`                                                              // 是否启用(可选)
	IsVisible      *bool   `json:"is_visible" example:"true"`                                                             // 是否可见(可选)
	DisplayOrder   *int16  `json:"display_order" example:"1"`                                                             // 显示顺序(可选)
}

// GetClasses 获取职业列表
// @Summary 获取职业列表
// @Description 获取职业列表，支持分页、筛选和排序
// @Tags 职业
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param tier query string false "职业层级" Enums(basic, advanced, elite, legendary, mythic)
// @Param is_active query bool false "是否激活"
// @Param is_visible query bool false "是否可见"
// @Param sort_by query string false "排序字段" default(display_order)
// @Param sort_dir query string false "排序方向" Enums(ASC, DESC) default(ASC)
// @Success 200 {object} response.Response{data=ClassListResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes [get]
// @Security BearerAuth
func (h *ClassHandler) GetClasses(c echo.Context) error {
	// 1. 解析查询参数
	params := interfaces.ClassQueryParams{
		Page:     parseIntParam(c.QueryParam("page"), 1),
		PageSize: parseIntParam(c.QueryParam("page_size"), 20),
		Tier:     c.QueryParam("tier"),
		SortBy:   c.QueryParam("sort_by"),
		SortDir:  c.QueryParam("sort_dir"),
	}

	// 解析布尔参数
	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		params.IsActive = &isActive
	}

	if isVisibleStr := c.QueryParam("is_visible"); isVisibleStr != "" {
		isVisible := isVisibleStr == "true"
		params.IsVisible = &isVisible
	}

	// 2. 查询职业列表
	classes, total, err := h.classService.GetClasses(c.Request().Context(), params)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "查询职业列表失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 3. 转换为响应格式
	classInfos := make([]ClassInfo, 0, len(classes))
	for _, cls := range classes {
		classInfo := h.convertToClassInfo(cls)
		classInfos = append(classInfos, classInfo)
	}

	// 4. 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	resp := ClassListResponse{
		Classes:    classInfos,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetClass 获取职业详情
// @Summary 获取单个职业
// @Description 获取指定职业的详细信息
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Success 200 {object} response.Response{data=ClassInfo} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id} [get]
// @Security BearerAuth
func (h *ClassHandler) GetClass(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	class, err := h.classService.GetClassByID(c.Request().Context(), classID)
	if err != nil {
		appErr := xerrors.NewNotFoundError("class", classID)
		return response.EchoError(c, h.respWriter, appErr)
	}

	classInfo := h.convertToClassInfo(class)
	return response.EchoOK(c, h.respWriter, classInfo)
}

// CreateClass 创建职业
// @Summary 创建职业
// @Description 创建新的职业
// @Tags 职业
// @Accept json
// @Produce json
// @Param class body CreateClassRequest true "职业信息"
// @Success 200 {object} response.Response{data=ClassInfo} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes [post]
// @Security BearerAuth
func (h *ClassHandler) CreateClass(c echo.Context) error {
	// 1. 绑定请求
	var req CreateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 3. 构建 Entity
	class := &game_config.Class{
		ClassCode: req.ClassCode,
		ClassName: req.ClassName,
		Tier:      req.Tier,
	}

	if req.Description != nil {
		class.Description.SetValid(*req.Description)
	}
	if req.LoreText != nil {
		class.LoreText.SetValid(*req.LoreText)
	}
	if req.Specialty != nil {
		class.Specialty.SetValid(*req.Specialty)
	}
	if req.Playstyle != nil {
		class.Playstyle.SetValid(*req.Playstyle)
	}
	if req.PromotionCount != nil {
		class.PromotionCount.SetValid(*req.PromotionCount)
	}
	if req.Icon != nil {
		class.Icon.SetValid(*req.Icon)
	}
	if req.Color != nil {
		class.Color.SetValid(*req.Color)
	}
	if req.IsActive != nil {
		class.IsActive.SetValid(*req.IsActive)
	} else {
		class.IsActive.SetValid(true) // 默认激活
	}
	if req.IsVisible != nil {
		class.IsVisible.SetValid(*req.IsVisible)
	} else {
		class.IsVisible.SetValid(true) // 默认可见
	}
	if req.DisplayOrder != nil {
		class.DisplayOrder.SetValid(*req.DisplayOrder)
	}

	// 4. 创建职业
	if err := h.classService.CreateClass(c.Request().Context(), class); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "创建职业失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回结果
	classInfo := h.convertToClassInfo(class)
	return response.EchoOK(c, h.respWriter, classInfo)
}

// UpdateClass 更新职业
// @Summary 更新职业
// @Description 更新职业信息
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Param class body UpdateClassRequest true "职业信息"
// @Success 200 {object} response.Response{data=map[string]string} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id} [put]
// @Security BearerAuth
func (h *ClassHandler) UpdateClass(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	// 1. 绑定请求
	var req UpdateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 2. 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 3. 构建更新字段
	updates := make(map[string]interface{})

	if req.ClassCode != nil {
		updates["class_code"] = *req.ClassCode
	}
	if req.ClassName != nil {
		updates["class_name"] = *req.ClassName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.LoreText != nil {
		updates["lore_text"] = *req.LoreText
	}
	if req.Specialty != nil {
		updates["specialty"] = *req.Specialty
	}
	if req.Playstyle != nil {
		updates["playstyle"] = *req.Playstyle
	}
	if req.Tier != nil {
		updates["tier"] = *req.Tier
	}
	if req.PromotionCount != nil {
		updates["promotion_count"] = *req.PromotionCount
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}

	// 4. 更新职业
	if err := h.classService.UpdateClass(c.Request().Context(), classID, updates); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "更新职业失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回结果
	resp := map[string]string{
		"message":  "职业更新成功",
		"class_id": classID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteClass 删除职业
// @Summary 删除职业
// @Description 软删除职业
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Success 200 {object} response.Response{data=map[string]string} "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id} [delete]
// @Security BearerAuth
func (h *ClassHandler) DeleteClass(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	if err := h.classService.DeleteClass(c.Request().Context(), classID); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "删除职业失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := map[string]string{
		"message":  "职业删除成功",
		"class_id": classID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// convertToClassInfo 将 Entity 转换为 HTTP 响应格式
func (h *ClassHandler) convertToClassInfo(class *game_config.Class) ClassInfo {
	info := ClassInfo{
		ID:        class.ID,
		ClassCode: class.ClassCode,
		ClassName: class.ClassName,
		Tier:      class.Tier,
		CreatedAt: class.CreatedAt.Unix(),
		UpdatedAt: class.UpdatedAt.Unix(),
	}

	if class.Description.Valid {
		info.Description = &class.Description.String
	}
	if class.LoreText.Valid {
		info.LoreText = &class.LoreText.String
	}
	if class.Specialty.Valid {
		info.Specialty = &class.Specialty.String
	}
	if class.Playstyle.Valid {
		info.Playstyle = &class.Playstyle.String
	}
	if class.PromotionCount.Valid {
		info.PromotionCount = &class.PromotionCount.Int16
	}
	if class.Icon.Valid {
		info.Icon = &class.Icon.String
	}
	if class.Color.Valid {
		info.Color = &class.Color.String
	}
	if class.IsActive.Valid {
		info.IsActive = &class.IsActive.Bool
	}
	if class.IsVisible.Valid {
		info.IsVisible = &class.IsVisible.Bool
	}
	if class.DisplayOrder.Valid {
		info.DisplayOrder = &class.DisplayOrder.Int16
	}

	return info
}

// ==================== 职业属性加成接口 ====================

// AttributeBonusInfo HTTP 响应结构
type AttributeBonusInfo struct {
	ID                 string `json:"id"`
	ClassID            string `json:"class_id"`
	AttributeID        string `json:"attribute_id"`
	BaseBonusValue     string `json:"base_bonus_value"`
	BonusPerLevel      bool   `json:"bonus_per_level"`
	PerLevelBonusValue string `json:"per_level_bonus_value"`
	CreatedAt          int64  `json:"created_at"`
	UpdatedAt          int64  `json:"updated_at"`
}

// CreateAttributeBonusRequest 创建属性加成请求
type CreateAttributeBonusRequest struct {
	AttributeID        string `json:"attribute_id" validate:"required"`
	BaseBonusValue     string `json:"base_bonus_value" validate:"required"`
	BonusPerLevel      bool   `json:"bonus_per_level"`
	PerLevelBonusValue string `json:"per_level_bonus_value"`
}

// UpdateAttributeBonusRequest 更新属性加成请求
type UpdateAttributeBonusRequest struct {
	BaseBonusValue     *string `json:"base_bonus_value"`
	BonusPerLevel      *bool   `json:"bonus_per_level"`
	PerLevelBonusValue *string `json:"per_level_bonus_value"`
}

// BatchSetAttributeBonusesRequest 批量设置属性加成请求
type BatchSetAttributeBonusesRequest struct {
	Bonuses []CreateAttributeBonusRequest `json:"bonuses" validate:"required,min=1"`
}

// GetClassAttributeBonuses 获取职业的所有属性加成
// @Summary 获取职业属性加成列表
// @Description 获取指定职业的所有属性加成配置
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Success 200 {object} response.Response{data=[]AttributeBonusInfo} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id}/attribute-bonuses [get]
// @Security BearerAuth
func (h *ClassHandler) GetClassAttributeBonuses(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	bonuses, err := h.classService.GetClassAttributeBonuses(c.Request().Context(), classID)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "查询属性加成失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为响应格式
	bonusInfos := make([]AttributeBonusInfo, 0, len(bonuses))
	for _, b := range bonuses {
		bonusInfos = append(bonusInfos, h.convertToAttributeBonusInfo(b))
	}

	return response.EchoOK(c, h.respWriter, bonusInfos)
}

// CreateAttributeBonus 为职业添加属性加成
// @Summary 创建属性加成
// @Description 为指定职业添加新的属性加成配置
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Param bonus body CreateAttributeBonusRequest true "属性加成信息"
// @Success 200 {object} response.Response{data=AttributeBonusInfo} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id}/attribute-bonuses [post]
// @Security BearerAuth
func (h *ClassHandler) CreateAttributeBonus(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	// 绑定请求
	var req CreateAttributeBonusRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构建 Entity
	bonus := &game_config.ClassAttributeBonuse{
		AttributeID:   req.AttributeID,
		BonusPerLevel: req.BonusPerLevel,
	}

	// 解析 Decimal 字段
	if err := bonus.BaseBonusValue.UnmarshalText([]byte(req.BaseBonusValue)); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "base_bonus_value 格式错误")
	}
	if err := bonus.PerLevelBonusValue.UnmarshalText([]byte(req.PerLevelBonusValue)); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "per_level_bonus_value 格式错误")
	}

	// 创建属性加成
	if err := h.classService.CreateAttributeBonus(c.Request().Context(), classID, bonus); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "创建属性加成失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 返回结果
	bonusInfo := h.convertToAttributeBonusInfo(bonus)
	return response.EchoOK(c, h.respWriter, bonusInfo)
}

// UpdateAttributeBonus 更新属性加成
// @Summary 更新属性加成
// @Description 更新指定的属性加成配置
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Param bonus_id path string true "属性加成ID"
// @Param bonus body UpdateAttributeBonusRequest true "属性加成信息"
// @Success 200 {object} response.Response{data=map[string]string} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "属性加成不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id}/attribute-bonuses/{bonus_id} [put]
// @Security BearerAuth
func (h *ClassHandler) UpdateAttributeBonus(c echo.Context) error {
	bonusID := c.Param("bonus_id")
	if bonusID == "" {
		return response.EchoBadRequest(c, h.respWriter, "属性加成ID不能为空")
	}

	// 绑定请求
	var req UpdateAttributeBonusRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.BaseBonusValue != nil {
		updates["base_bonus_value"] = *req.BaseBonusValue
	}
	if req.BonusPerLevel != nil {
		updates["bonus_per_level"] = *req.BonusPerLevel
	}
	if req.PerLevelBonusValue != nil {
		updates["per_level_bonus_value"] = *req.PerLevelBonusValue
	}

	// 更新属性加成
	if err := h.classService.UpdateAttributeBonus(c.Request().Context(), bonusID, updates); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "更新属性加成失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 返回结果
	resp := map[string]string{
		"message":  "属性加成更新成功",
		"bonus_id": bonusID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// DeleteAttributeBonus 删除属性加成
// @Summary 删除属性加成
// @Description 删除指定的属性加成配置
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Param bonus_id path string true "属性加成ID"
// @Success 200 {object} response.Response{data=map[string]string} "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "属性加成不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id}/attribute-bonuses/{bonus_id} [delete]
// @Security BearerAuth
func (h *ClassHandler) DeleteAttributeBonus(c echo.Context) error {
	bonusID := c.Param("bonus_id")
	if bonusID == "" {
		return response.EchoBadRequest(c, h.respWriter, "属性加成ID不能为空")
	}

	if err := h.classService.DeleteAttributeBonus(c.Request().Context(), bonusID); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "删除属性加成失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := map[string]string{
		"message":  "属性加成删除成功",
		"bonus_id": bonusID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// BatchSetAttributeBonuses 批量设置职业属性加成
// @Summary 批量设置属性加成
// @Description 批量设置职业的属性加成（先删除旧的，再创建新的）
// @Tags 职业
// @Accept json
// @Produce json
// @Param id path string true "职业ID"
// @Param request body BatchSetAttributeBonusesRequest true "属性加成列表"
// @Success 200 {object} response.Response{data=map[string]string} "批量设置成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "职业不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/classes/{id}/attribute-bonuses/batch [post]
// @Security BearerAuth
func (h *ClassHandler) BatchSetAttributeBonuses(c echo.Context) error {
	classID := c.Param("id")
	if classID == "" {
		return response.EchoBadRequest(c, h.respWriter, "职业ID不能为空")
	}

	// 绑定请求
	var req BatchSetAttributeBonusesRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	// 验证请求
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 转换为 Entity
	bonuses := make([]*game_config.ClassAttributeBonuse, 0, len(req.Bonuses))
	for _, b := range req.Bonuses {
		bonus := &game_config.ClassAttributeBonuse{
			AttributeID:   b.AttributeID,
			BonusPerLevel: b.BonusPerLevel,
		}

		// 解析 Decimal 字段
		if err := bonus.BaseBonusValue.UnmarshalText([]byte(b.BaseBonusValue)); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "base_bonus_value 格式错误")
		}
		if err := bonus.PerLevelBonusValue.UnmarshalText([]byte(b.PerLevelBonusValue)); err != nil {
			return response.EchoBadRequest(c, h.respWriter, "per_level_bonus_value 格式错误")
		}

		bonuses = append(bonuses, bonus)
	}

	// 批量设置
	if err := h.classService.BatchSetAttributeBonuses(c.Request().Context(), classID, bonuses); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "批量设置属性加成失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := map[string]string{
		"message":  "批量设置属性加成成功",
		"class_id": classID,
		"count":    fmt.Sprintf("%d", len(bonuses)),
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// convertToAttributeBonusInfo 转换为 HTTP 响应格式
func (h *ClassHandler) convertToAttributeBonusInfo(bonus *game_config.ClassAttributeBonuse) AttributeBonusInfo {
	baseValue, _ := bonus.BaseBonusValue.MarshalText()
	perLevelValue, _ := bonus.PerLevelBonusValue.MarshalText()

	return AttributeBonusInfo{
		ID:                 bonus.ID,
		ClassID:            bonus.ClassID,
		AttributeID:        bonus.AttributeID,
		BaseBonusValue:     string(baseValue),
		BonusPerLevel:      bonus.BonusPerLevel,
		PerLevelBonusValue: string(perLevelValue),
		CreatedAt:          bonus.CreatedAt.Unix(),
		UpdatedAt:          bonus.UpdatedAt.Unix(),
	}
}

// ==================== 职业进阶路径查询接口 ====================

// GetClassAdvancement godoc
// @Summary      获取职业可进阶的目标职业列表
// @Description  查询指定职业可以进阶到哪些职业，返回目标职业列表及进阶要求
// @Tags 职业
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "职业ID"  Format(uuid)
// @Success      200  {object}  response.Response{data=[]AdvancedRequirementInfo}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/classes/{id}/advancement [get]
func (h *ClassHandler) GetClassAdvancement(c echo.Context) error {
	classID := c.Param("id")

	requirements, err := h.advService.GetByFromClass(c.Request().Context(), classID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	infos := make([]AdvancedRequirementInfo, len(requirements))
	for i, req := range requirements {
		infos[i] = convertAdvancedRequirementToInfo(req)
	}

	return response.EchoOK(c, h.respWriter, infos)
}

// AdvancementPathResponse 进阶路径响应
type AdvancementPathResponse struct {
	Path  []AdvancedRequirementInfo `json:"path"`
	Depth int                       `json:"depth"`
}

// GetClassAdvancementPaths godoc
// @Summary      获取职业的完整进阶路径树
// @Description  查询职业的完整进阶树（包括多级进阶），支持指定最大深度
// @Tags 职业
// @Accept       json
// @Produce      json
// @Param        id         path   string  true   "职业ID"  Format(uuid)
// @Param        max_depth  query  int     false  "最大深度"  Default(5)  minimum(1)  maximum(10)
// @Success      200  {object}  response.Response{data=[]AdvancementPathResponse}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/classes/{id}/advancement-paths [get]
func (h *ClassHandler) GetClassAdvancementPaths(c echo.Context) error {
	classID := c.Param("id")
	maxDepth := 5

	if maxDepthStr := c.QueryParam("max_depth"); maxDepthStr != "" {
		if depth, err := fmt.Sscanf(maxDepthStr, "%d", &maxDepth); err == nil && depth > 0 {
			if maxDepth > 10 {
				maxDepth = 10
			}
		}
	}

	paths, err := h.advService.GetAdvancementPaths(c.Request().Context(), classID, maxDepth)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	responses := make([]AdvancementPathResponse, len(paths))
	for i, path := range paths {
		infos := make([]AdvancedRequirementInfo, len(path))
		for j, req := range path {
			infos[j] = convertAdvancedRequirementToInfo(req)
		}
		responses[i] = AdvancementPathResponse{
			Path:  infos,
			Depth: len(path),
		}
	}

	return response.EchoOK(c, h.respWriter, responses)
}

// GetClassAdvancementSources godoc
// @Summary      获取可进阶到指定职业的源职业列表
// @Description  查询可以进阶到指定职业的源职业列表及进阶要求
// @Tags 职业
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "职业ID"  Format(uuid)
// @Success      200  {object}  response.Response{data=[]AdvancedRequirementInfo}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /admin/classes/{id}/advancement-sources [get]
func (h *ClassHandler) GetClassAdvancementSources(c echo.Context) error {
	classID := c.Param("id")

	requirements, err := h.advService.GetByToClass(c.Request().Context(), classID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 转换为响应格式
	infos := make([]AdvancedRequirementInfo, len(requirements))
	for i, req := range requirements {
		infos[i] = convertAdvancedRequirementToInfo(req)
	}

	return response.EchoOK(c, h.respWriter, infos)
}

// convertAdvancedRequirementToInfo 转换进阶要求为响应格式（复用AdvancedRequirementHandler的结构）
func convertAdvancedRequirementToInfo(req *game_config.ClassAdvancedRequirement) AdvancedRequirementInfo {
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
		info.RequiredAttributes = jsonBytes
	}

	if req.RequiredSkills.Valid {
		jsonBytes, _ := req.RequiredSkills.MarshalJSON()
		info.RequiredSkills = jsonBytes
	}

	if req.RequiredItems.Valid {
		jsonBytes, _ := req.RequiredItems.MarshalJSON()
		info.RequiredItems = jsonBytes
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
