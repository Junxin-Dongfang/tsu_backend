package interfaces

import (
	"context"

	"github.com/google/uuid"

	"tsu-self/internal/entity"
	"tsu-self/internal/repository/query"
)

// ClassRepository 职业数据仓储接口
type ClassRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, class *entity.Class) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Class, error)
	GetByCode(ctx context.Context, code string) (*entity.Class, error)
	Update(ctx context.Context, class *entity.Class) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 列表查询
	List(ctx context.Context, params *query.ClassListParams) ([]*entity.Class, int64, error)

	// 统计信息
	GetHeroStats(ctx context.Context, classID uuid.UUID) (*query.ClassHeroStats, error)

	// 属性加成管理
	CreateAttributeBonus(ctx context.Context, bonus *entity.ClassAttributeBonuse) error
	GetAttributeBonuses(ctx context.Context, classID uuid.UUID) ([]*query.ClassAttributeBonusWithDetails, error)
	GetAttributeBonus(ctx context.Context, classID, attributeID uuid.UUID) (*entity.ClassAttributeBonuse, error)
	UpdateAttributeBonus(ctx context.Context, bonus *entity.ClassAttributeBonuse) error
	DeleteAttributeBonus(ctx context.Context, id uuid.UUID) error
	BatchCreateAttributeBonuses(ctx context.Context, bonuses []*entity.ClassAttributeBonuse) error

	// 进阶要求管理
	CreateAdvancementRequirement(ctx context.Context, req *entity.ClassAdvancedRequirement) error
	GetAdvancementRequirement(ctx context.Context, fromClassID, toClassID uuid.UUID) (*entity.ClassAdvancedRequirement, error)
	GetAdvancementRequirements(ctx context.Context, classID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error)
	GetAdvancementPaths(ctx context.Context, fromClassID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error)
	GetAdvancementSources(ctx context.Context, toClassID uuid.UUID) ([]*query.ClassAdvancementWithDetails, error)
	UpdateAdvancementRequirement(ctx context.Context, req *entity.ClassAdvancedRequirement) error
	DeleteAdvancementRequirement(ctx context.Context, id uuid.UUID) error

	// 标签管理
	GetClassTags(ctx context.Context, classID uuid.UUID) ([]*query.ClassTagWithDetails, error)
	AddClassTag(ctx context.Context, classID, tagID uuid.UUID) error
	RemoveClassTag(ctx context.Context, classID, tagID uuid.UUID) error

	// 标签列表
	GetAllTags(ctx context.Context, tagType *string) ([]*entity.Tag, error)
}