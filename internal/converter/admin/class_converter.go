package admin

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/google/uuid"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	apiRespAdmin "tsu-self/internal/api/model/response/admin"
	"tsu-self/internal/entity"
	"tsu-self/internal/repository/query"
)

// ConvertToClass 将创建请求转换为数据库实体
func ConvertCreateRequestToClass(req *apiReqAdmin.CreateClassRequest) *entity.Class {
	class := &entity.Class{
		ClassCode: req.Code,
		ClassName: req.Name,
		Tier:      fmt.Sprintf("%d", req.Tier),
		IsActive:  null.BoolFrom(true), // 默认启用
	}

	if req.Description != nil {
		class.Description = null.StringFrom(*req.Description)
	}

	if req.LoreText != nil {
		class.LoreText = null.StringFrom(*req.LoreText)
	}

	if req.JobChangeBonus != nil {
		class.PromotionCount = null.Int16From(int16(*req.JobChangeBonus))
	}

	if req.Icon != nil {
		class.Icon = null.StringFrom(*req.Icon)
	}

	if req.ColorTheme != nil {
		class.Color = null.StringFrom(*req.ColorTheme)
	}

	if req.IsHidden != nil {
		// is_hidden 对应数据库的 is_visible（反向）
		class.IsVisible = null.BoolFrom(!*req.IsHidden)
	} else {
		class.IsVisible = null.BoolFrom(true)
	}

	if req.DisplayOrder != nil {
		class.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}

	return class
}

// UpdateClassFromRequest 使用更新请求更新数据库实体
func UpdateClassFromRequest(class *entity.Class, req *apiReqAdmin.UpdateClassRequest) {
	if req.Code != nil {
		class.ClassCode = *req.Code
	}

	if req.Name != nil {
		class.ClassName = *req.Name
	}

	if req.Description != nil {
		class.Description = null.StringFrom(*req.Description)
	}

	if req.LoreText != nil {
		class.LoreText = null.StringFrom(*req.LoreText)
	}

	if req.Tier != nil {
		class.Tier = fmt.Sprintf("%d", *req.Tier)
	}

	if req.JobChangeBonus != nil {
		class.PromotionCount = null.Int16From(int16(*req.JobChangeBonus))
	}

	if req.Icon != nil {
		class.Icon = null.StringFrom(*req.Icon)
	}

	if req.ColorTheme != nil {
		class.Color = null.StringFrom(*req.ColorTheme)
	}

	if req.IsHidden != nil {
		// is_hidden 对应数据库的 is_visible（反向）
		class.IsVisible = null.BoolFrom(!*req.IsHidden)
	}

	if req.DisplayOrder != nil {
		class.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}
}

// ConvertClassToResponse 将数据库实体转换为API响应
func ConvertClassToResponse(class *entity.Class) *apiRespAdmin.Class {
	// 解析UUID
	classID, _ := uuid.Parse(class.ID)

	// 解析Tier
	tier, _ := strconv.Atoi(class.Tier)

	resp := &apiRespAdmin.Class{
		ID:           classID,
		Code:         class.ClassCode,
		Name:         class.ClassName,
		Tier:         tier,
		IsActive:     class.IsActive.Bool,
		IsHidden:     !class.IsVisible.Bool, // 反向映射
		DisplayOrder: int(class.DisplayOrder.Int16),
		CreatedAt:    class.CreatedAt.Time,
		UpdatedAt:    class.UpdatedAt.Time,
	}

	if class.Description.Valid {
		resp.Description = &class.Description.String
	}

	if class.LoreText.Valid {
		resp.LoreText = &class.LoreText.String
	}

	if class.PromotionCount.Valid {
		resp.JobChangeBonus = int(class.PromotionCount.Int16)
	}

	if class.Icon.Valid {
		resp.Icon = &class.Icon.String
	}

	if class.Color.Valid {
		resp.ColorTheme = &class.Color.String
	}

	return resp
}

// ConvertClassesToResponse 将数据库实体列表转换为API响应列表
func ConvertClassesToResponse(classes []*entity.Class) []apiRespAdmin.Class {
	responses := make([]apiRespAdmin.Class, len(classes))
	for i, class := range classes {
		responses[i] = *ConvertClassToResponse(class)
	}
	return responses
}

// ConvertToClassWithStats 将实体和统计信息转换为带统计的响应
func ConvertToClassWithStats(class *entity.Class, stats *query.ClassHeroStats) *apiRespAdmin.ClassWithStats {
	resp := &apiRespAdmin.ClassWithStats{
		Class: *ConvertClassToResponse(class),
	}

	if stats != nil {
		resp.TotalHeroes = stats.TotalHeroes
		resp.ActiveHeroes = stats.ActiveHeroes
		resp.AverageLevel = stats.AverageLevel
		resp.MaxLevel = stats.MaxLevel
	}

	return resp
}

// ConvertToAttributeBonusEntity 将请求转换为属性加成实体
func ConvertToAttributeBonusEntity(classID string, req *apiReqAdmin.CreateClassAttributeBonusRequest) *entity.ClassAttributeBonuse {
	// 使用 ericlagergren/decimal 创建 Big 值
	var baseBonus decimal.Big
	baseBonus.SetFloat64(req.BaseBonus)

	var perLevelBonus decimal.Big
	perLevelBonus.SetFloat64(req.PerLevelBonus)

	return &entity.ClassAttributeBonuse{
		ClassID:            classID,
		AttributeID:        req.AttributeID.String(),
		BaseBonusValue:     types.NewDecimal(&baseBonus),
		PerLevelBonusValue: types.NewDecimal(&perLevelBonus),
	}
}

// UpdateAttributeBonusFromRequest 使用请求更新属性加成实体
func UpdateAttributeBonusFromRequest(bonus *entity.ClassAttributeBonuse, req *apiReqAdmin.UpdateClassAttributeBonusRequest) {
	if req.BaseBonus != nil {
		var big decimal.Big
		big.SetFloat64(*req.BaseBonus)
		bonus.BaseBonusValue = types.NewDecimal(&big)
	}

	if req.PerLevelBonus != nil {
		var big decimal.Big
		big.SetFloat64(*req.PerLevelBonus)
		bonus.PerLevelBonusValue = types.NewDecimal(&big)
	}
}

// ConvertAttributeBonusToResponse 将属性加成详情转换为响应
func ConvertAttributeBonusToResponse(bonus *query.ClassAttributeBonusWithDetails) *apiRespAdmin.ClassAttributeBonus {
	return &apiRespAdmin.ClassAttributeBonus{
		ID:            bonus.ID,
		ClassID:       bonus.ClassID,
		AttributeID:   bonus.AttributeID,
		AttributeCode: bonus.AttributeCode,
		AttributeName: bonus.AttributeName,
		BaseBonus:     bonus.BaseBonus,
		PerLevelBonus: bonus.PerLevelBonus,
		CreatedAt:     bonus.CreatedAt,
		UpdatedAt:     bonus.UpdatedAt,
	}
}

// ConvertAttributeBonusesToResponse 批量转换属性加成
func ConvertAttributeBonusesToResponse(bonuses []*query.ClassAttributeBonusWithDetails) []*apiRespAdmin.ClassAttributeBonus {
	responses := make([]*apiRespAdmin.ClassAttributeBonus, len(bonuses))
	for i, bonus := range bonuses {
		responses[i] = ConvertAttributeBonusToResponse(bonus)
	}
	return responses
}

// ConvertToAdvancementRequirementEntity 将请求转换为进阶要求实体
func ConvertToAdvancementRequirementEntity(req *apiReqAdmin.CreateClassAdvancementRequest) *entity.ClassAdvancedRequirement {
	entity := &entity.ClassAdvancedRequirement{
		FromClassID:   req.FromClassID.String(),
		ToClassID:     req.ToClassID.String(),
		RequiredLevel: req.RequiredLevel,
	}

	if req.RequiredHonor != nil {
		entity.RequiredHonor = *req.RequiredHonor
	}

	if req.RequiredJobChangeCount != nil {
		entity.RequiredJobChangeCount = *req.RequiredJobChangeCount
	} else {
		entity.RequiredJobChangeCount = 1 // 默认值
	}

	if req.RequiredAttributes != nil {
		if attrs, err := json.Marshal(req.RequiredAttributes); err == nil {
			entity.RequiredAttributes = null.JSONFrom(attrs)
		}
	}

	if req.RequiredSkills != nil {
		if skills, err := json.Marshal(req.RequiredSkills); err == nil {
			entity.RequiredSkills = null.JSONFrom(skills)
		}
	}

	if req.DisplayOrder != nil {
		entity.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}

	return entity
}

// UpdateAdvancementRequirementFromRequest 使用请求更新进阶要求
func UpdateAdvancementRequirementFromRequest(requirement *entity.ClassAdvancedRequirement, req *apiReqAdmin.UpdateClassAdvancementRequest) {
	if req.RequiredLevel != nil {
		requirement.RequiredLevel = *req.RequiredLevel
	}

	if req.RequiredHonor != nil {
		requirement.RequiredHonor = *req.RequiredHonor
	}

	if req.RequiredJobChangeCount != nil {
		requirement.RequiredJobChangeCount = *req.RequiredJobChangeCount
	}

	if req.RequiredAttributes != nil {
		if attrs, err := json.Marshal(req.RequiredAttributes); err == nil {
			requirement.RequiredAttributes = null.JSONFrom(attrs)
		}
	}

	if req.RequiredSkills != nil {
		if skills, err := json.Marshal(req.RequiredSkills); err == nil {
			requirement.RequiredSkills = null.JSONFrom(skills)
		}
	}

	if req.IsActive != nil {
		requirement.IsActive = null.BoolFrom(*req.IsActive)
	}

	if req.DisplayOrder != nil {
		requirement.DisplayOrder = null.Int16From(int16(*req.DisplayOrder))
	}
}

// ConvertAdvancementRequirementToResponse 将进阶要求转换为响应
func ConvertAdvancementRequirementToResponse(req *query.ClassAdvancementWithDetails) *apiRespAdmin.ClassAdvancementRequirement {
	return &apiRespAdmin.ClassAdvancementRequirement{
		ID:                     req.ID,
		FromClassID:            req.FromClassID,
		FromClassName:          req.FromClassName,
		ToClassID:              req.ToClassID,
		ToClassName:            req.ToClassName,
		RequiredLevel:          req.RequiredLevel,
		RequiredHonor:          req.RequiredHonor,
		RequiredJobChangeCount: req.RequiredJobChangeCount,
		RequiredAttributes:     req.RequiredAttributes,
		RequiredSkills:         req.RequiredSkills,
		IsActive:               req.IsActive,
		DisplayOrder:           req.DisplayOrder,
		CreatedAt:              req.CreatedAt,
		UpdatedAt:              req.UpdatedAt,
	}
}

// ConvertAdvancementRequirementsToResponse 批量转换进阶要求
func ConvertAdvancementRequirementsToResponse(reqs []*query.ClassAdvancementWithDetails) []*apiRespAdmin.ClassAdvancementRequirement {
	responses := make([]*apiRespAdmin.ClassAdvancementRequirement, len(reqs))
	for i, req := range reqs {
		responses[i] = ConvertAdvancementRequirementToResponse(req)
	}
	return responses
}

// ConvertClassTagToResponse 将职业标签转换为响应
func ConvertClassTagToResponse(tag *query.ClassTagWithDetails) *apiRespAdmin.ClassTag {
	resp := &apiRespAdmin.ClassTag{
		ID:           tag.ID,
		Code:         tag.Code,
		Name:         tag.Name,
		DisplayOrder: tag.DisplayOrder,
	}

	if tag.Description != nil {
		resp.Description = tag.Description
	}

	if tag.Color != nil {
		resp.Color = tag.Color
	}

	if tag.Icon != nil {
		resp.Icon = tag.Icon
	}

	return resp
}

// ConvertClassTagsToResponse 批量转换职业标签
func ConvertClassTagsToResponse(tags []*query.ClassTagWithDetails) []*apiRespAdmin.ClassTag {
	responses := make([]*apiRespAdmin.ClassTag, len(tags))
	for i, tag := range tags {
		responses[i] = ConvertClassTagToResponse(tag)
	}
	return responses
}

// ConvertTagToClassTag 将标签实体转换为职业标签响应
func ConvertTagToClassTag(tag *entity.Tag) *apiRespAdmin.ClassTag {
	tagID, _ := uuid.Parse(tag.ID)
	resp := &apiRespAdmin.ClassTag{
		ID:           tagID,
		Code:         tag.TagCode,
		Name:         tag.TagName,
		DisplayOrder: int(tag.DisplayOrder.Int16),
	}

	if tag.Description.Valid {
		resp.Description = &tag.Description.String
	}

	if tag.Color.Valid {
		resp.Color = &tag.Color.String
	}

	if tag.Icon.Valid {
		resp.Icon = &tag.Icon.String
	}

	return resp
}

// ConvertTagsToClassTags 批量转换标签为职业标签
func ConvertTagsToClassTags(tags []*entity.Tag) []*apiRespAdmin.ClassTag {
	responses := make([]*apiRespAdmin.ClassTag, len(tags))
	for i, tag := range tags {
		responses[i] = ConvertTagToClassTag(tag)
	}
	return responses
}

// ConvertToClassHeroStatsResponse 转换职业英雄统计响应
func ConvertToClassHeroStatsResponse(stats *query.ClassHeroStats) *apiRespAdmin.ClassHeroStats {
	return &apiRespAdmin.ClassHeroStats{
		ClassID:      stats.ClassID,
		TotalHeroes:  stats.TotalHeroes,
		ActiveHeroes: stats.ActiveHeroes,
		AverageLevel: stats.AverageLevel,
		MaxLevel:     stats.MaxLevel,
	}
}
