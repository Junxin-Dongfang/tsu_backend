package impl

import (
	"context"
	"database/sql"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/repository/interfaces"
)

// MetadataDictionaryRepositoryImpl 元数据字典仓储实现
type MetadataDictionaryRepositoryImpl struct {
	db *sql.DB
}

// NewMetadataDictionaryRepository 创建元数据字典仓储
func NewMetadataDictionaryRepository(db *sql.DB) interfaces.MetadataDictionaryRepository {
	return &MetadataDictionaryRepositoryImpl{db: db}
}

// GetByID 根据ID获取字典项
func (r *MetadataDictionaryRepositoryImpl) GetByID(ctx context.Context, id string) (*game_config.MetadataDictionary, error) {
	return game_config.MetadataDictionaries(
		game_config.MetadataDictionaryWhere.ID.EQ(id),
		game_config.MetadataDictionaryWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
}

// GetByCodeAndCategory 根据代码和分类获取字典项
func (r *MetadataDictionaryRepositoryImpl) GetByCodeAndCategory(ctx context.Context, code, category string) (*game_config.MetadataDictionary, error) {
	return game_config.MetadataDictionaries(
		game_config.MetadataDictionaryWhere.VariableCode.EQ(code),
		game_config.MetadataDictionaryWhere.DictCategory.EQ(category),
		game_config.MetadataDictionaryWhere.IsActive.EQ(null.BoolFrom(true)),
		game_config.MetadataDictionaryWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
}

// GetByCategory 根据分类获取所有字典项
func (r *MetadataDictionaryRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*game_config.MetadataDictionary, error) {
	return game_config.MetadataDictionaries(
		game_config.MetadataDictionaryWhere.DictCategory.EQ(category),
		game_config.MetadataDictionaryWhere.IsActive.EQ(null.BoolFrom(true)),
		game_config.MetadataDictionaryWhere.DeletedAt.IsNull(),
		qm.OrderBy("variable_code ASC"),
	).All(ctx, r.db)
}

// GetActionAttributes 获取所有动作属性字典项
func (r *MetadataDictionaryRepositoryImpl) GetActionAttributes(ctx context.Context) ([]*game_config.MetadataDictionary, error) {
	return r.GetByCategory(ctx, interfaces.DictCategoryActionAttribute)
}

// GetFormulaVariables 获取所有公式变量字典项
func (r *MetadataDictionaryRepositoryImpl) GetFormulaVariables(ctx context.Context) ([]*game_config.MetadataDictionary, error) {
	return r.GetByCategory(ctx, interfaces.DictCategoryFormula)
}
