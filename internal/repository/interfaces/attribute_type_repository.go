package interfaces

import (
	"context"

	apiReqAdmin "tsu-self/internal/api/model/request/admin"
	"tsu-self/internal/entity"
)

// AttributeTypeRepository 属性类型仓储接口
type AttributeTypeRepository interface {
	// Create 创建属性类型
	Create(ctx context.Context, attributeType *entity.HeroAttributeType) error

	// GetByID 根据ID获取属性类型
	GetByID(ctx context.Context, id string) (*entity.HeroAttributeType, error)

	// GetByCode 根据属性代码获取属性类型
	GetByCode(ctx context.Context, code string) (*entity.HeroAttributeType, error)

	// List 获取属性类型列表
	List(ctx context.Context, req *apiReqAdmin.GetAttributeTypesRequest) ([]*entity.HeroAttributeType, int64, error)

	// Update 更新属性类型
	Update(ctx context.Context, attributeType *entity.HeroAttributeType) error

	// Delete 软删除属性类型
	Delete(ctx context.Context, id string) error

	// ExistsByCode 检查属性代码是否存在
	ExistsByCode(ctx context.Context, code string, excludeID ...string) (bool, error)

	// GetActiveList 获取启用的属性类型列表（用于选项）
	GetActiveList(ctx context.Context, category string) ([]*entity.HeroAttributeType, error)
}