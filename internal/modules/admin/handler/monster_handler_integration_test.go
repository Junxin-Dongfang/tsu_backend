package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

func TestMonsterHandler_GetMonster_ReturnsAggregatedData(t *testing.T) {
	handler, _ := newTestMonsterHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/monsters/monster-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/admin/monsters/:id")
	c.SetParamNames("id")
	c.SetParamValues("monster-1")

	require.NoError(t, handler.GetMonster(c))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data MonsterInfo
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	require.Equal(t, "M001", resp.Data.MonsterCode)
	require.Len(t, resp.Data.Skills, 1)
	require.Equal(t, "FIREBALL", resp.Data.Skills[0].SkillCode)
	require.Len(t, resp.Data.Drops, 1)
	require.Equal(t, "POOL_MAIN", resp.Data.Drops[0].DropPoolCode)
	require.Len(t, resp.Data.Tags, 1)
	require.Equal(t, "BEAST", resp.Data.Tags[0].TagCode)
	require.NotZero(t, resp.Data.DamageResistances)
	require.NotZero(t, resp.Data.PassiveBuffs)
}

func TestMonsterHandler_GetMonsters_InvalidOrderBy(t *testing.T) {
	handler, _ := newTestMonsterHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/monsters?order_by=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetMonsters(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMonsterHandler_GetMonsters_TagFilterPropagated(t *testing.T) {
	handler, monsterRepo := newTestMonsterHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/monsters?tag_ids=tag-1,tag-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.GetMonsters(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []string{"tag-1", "tag-2"}, monsterRepo.lastParams.TagIDs)
}

func newTestMonsterHandler() (*MonsterHandler, *stubMonsterRepo) {
	monster := &game_config.Monster{
		ID:           "monster-1",
		MonsterCode:  "M001",
		MonsterName:  "测试怪物",
		MonsterLevel: 5,
		MaxHP:        30,
	}
	_ = monster.DamageResistances.Marshal(map[string]interface{}{"FIRE_RESIST": 0.5})
	_ = monster.PassiveBuffs.Marshal([]interface{}{map[string]interface{}{"buff_id": "REGEN"}})

	monsterRepo := &stubMonsterRepo{
		monsters: map[string]*game_config.Monster{"monster-1": monster},
		list:     []*game_config.Monster{monster},
	}
	monsterSkillRepo := &stubMonsterSkillRepo{
		skills: map[string][]*game_config.MonsterSkill{
			"monster-1": {
				{
					ID:         "ms-1",
					MonsterID:  "monster-1",
					SkillID:    "skill-1",
					SkillLevel: 3,
				},
			},
		},
	}
	monsterDropRepo := &stubMonsterDropRepo{
		drops: map[string][]*game_config.MonsterDrop{
			"monster-1": {
				{
					ID:         "md-1",
					MonsterID:  "monster-1",
					DropPoolID: "pool-1",
					DropType:   "team",
					DropChance: types.NewDecimal(decimal.New(8, -1)),
				},
			},
		},
	}
	skillRepo := &stubSkillRepo{
		skills: map[string]*game_config.Skill{
			"skill-1": {
				ID:        "skill-1",
				SkillCode: "FIREBALL",
				SkillName: "火球术",
			},
		},
	}
	dropPoolRepo := &stubDropPoolRepo{
		pools: map[string]*game_config.DropPool{
			"pool-1": {
				ID:       "pool-1",
				PoolCode: "POOL_MAIN",
				PoolName: "主掉落池",
			},
		},
	}
	tagRepo := &stubTagRelationRepo{
		tags: map[string][]*game_config.Tag{
			"monster-1": {
				{
					ID:          "tag-1",
					TagCode:     "BEAST",
					TagName:     "野兽",
					Color:       null.StringFrom("red"),
					Icon:        null.StringFrom("icon.png"),
					Description: null.StringFrom("tag desc"),
					Category:    "type",
				},
			},
		},
	}

	svc := service.NewMonsterServiceWithDeps(service.MonsterServiceDeps{
		MonsterRepo:       monsterRepo,
		MonsterSkillRepo:  monsterSkillRepo,
		MonsterDropRepo:   monsterDropRepo,
		SkillRepo:         skillRepo,
		DropPoolRepo:      dropPoolRepo,
		TagRelationRepo:   tagRepo,
		AttributeTypeRepo: &stubAttributeTypeRepo{},
	})

	return &MonsterHandler{
		service:    svc,
		respWriter: response.DefaultResponseHandler(),
	}, monsterRepo
}

type stubMonsterRepo struct {
	monsters   map[string]*game_config.Monster
	list       []*game_config.Monster
	lastParams interfaces.MonsterQueryParams
}

func (s *stubMonsterRepo) GetByID(ctx context.Context, id string) (*game_config.Monster, error) {
	return s.monsters[id], nil
}
func (s *stubMonsterRepo) GetByCode(ctx context.Context, code string) (*game_config.Monster, error) {
	for _, m := range s.monsters {
		if m.MonsterCode == code {
			return m, nil
		}
	}
	return nil, nil
}
func (s *stubMonsterRepo) List(ctx context.Context, params interfaces.MonsterQueryParams) ([]*game_config.Monster, int64, error) {
	s.lastParams = params
	return s.list, int64(len(s.list)), nil
}
func (s *stubMonsterRepo) Create(ctx context.Context, monster *game_config.Monster) error { return nil }
func (s *stubMonsterRepo) Update(ctx context.Context, monster *game_config.Monster) error { return nil }
func (s *stubMonsterRepo) Delete(ctx context.Context, monsterID string) error             { return nil }
func (s *stubMonsterRepo) Exists(ctx context.Context, code string) (bool, error)          { return false, nil }
func (s *stubMonsterRepo) ExistsExcludingID(ctx context.Context, code, id string) (bool, error) {
	return false, nil
}

type stubMonsterSkillRepo struct {
	skills map[string][]*game_config.MonsterSkill
}

func (s *stubMonsterSkillRepo) Create(ctx context.Context, skill *game_config.MonsterSkill) error {
	return nil
}
func (s *stubMonsterSkillRepo) BatchCreate(ctx context.Context, skills []*game_config.MonsterSkill) error {
	return nil
}
func (s *stubMonsterSkillRepo) GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterSkill, error) {
	return s.skills[monsterID], nil
}
func (s *stubMonsterSkillRepo) GetByMonsterAndSkill(ctx context.Context, monsterID, skillID string) (*game_config.MonsterSkill, error) {
	return nil, nil
}
func (s *stubMonsterSkillRepo) Update(ctx context.Context, skill *game_config.MonsterSkill) error {
	return nil
}
func (s *stubMonsterSkillRepo) Delete(ctx context.Context, monsterID, skillID string) error {
	return nil
}
func (s *stubMonsterSkillRepo) DeleteByMonsterID(ctx context.Context, monsterID string) error {
	return nil
}
func (s *stubMonsterSkillRepo) Exists(ctx context.Context, monsterID, skillID string) (bool, error) {
	return false, nil
}

type stubMonsterDropRepo struct {
	drops map[string][]*game_config.MonsterDrop
}

func (s *stubMonsterDropRepo) Create(ctx context.Context, drop *game_config.MonsterDrop) error {
	return nil
}
func (s *stubMonsterDropRepo) BatchCreate(ctx context.Context, drops []*game_config.MonsterDrop) error {
	return nil
}
func (s *stubMonsterDropRepo) GetByMonsterID(ctx context.Context, monsterID string) ([]*game_config.MonsterDrop, error) {
	return s.drops[monsterID], nil
}
func (s *stubMonsterDropRepo) GetByMonsterAndPool(ctx context.Context, monsterID, poolID string) (*game_config.MonsterDrop, error) {
	return nil, nil
}
func (s *stubMonsterDropRepo) Update(ctx context.Context, drop *game_config.MonsterDrop) error {
	return nil
}
func (s *stubMonsterDropRepo) Delete(ctx context.Context, monsterID, poolID string) error { return nil }
func (s *stubMonsterDropRepo) DeleteByMonsterID(ctx context.Context, monsterID string) error {
	return nil
}
func (s *stubMonsterDropRepo) Exists(ctx context.Context, monsterID, poolID string) (bool, error) {
	return false, nil
}

type stubSkillRepo struct {
	skills map[string]*game_config.Skill
}

func (s *stubSkillRepo) GetByID(ctx context.Context, id string) (*game_config.Skill, error) {
	return s.skills[id], nil
}
func (s *stubSkillRepo) GetByCode(ctx context.Context, code string) (*game_config.Skill, error) {
	return nil, nil
}
func (s *stubSkillRepo) List(ctx context.Context, params interfaces.SkillQueryParams) ([]*game_config.Skill, int64, error) {
	return nil, 0, nil
}
func (s *stubSkillRepo) Create(ctx context.Context, skill *game_config.Skill) error { return nil }
func (s *stubSkillRepo) Update(ctx context.Context, skill *game_config.Skill) error { return nil }
func (s *stubSkillRepo) Delete(ctx context.Context, skillID string) error           { return nil }
func (s *stubSkillRepo) Exists(ctx context.Context, code string) (bool, error)      { return false, nil }

type stubDropPoolRepo struct {
	pools map[string]*game_config.DropPool
}

func (s *stubDropPoolRepo) GetByID(ctx context.Context, id string) (*game_config.DropPool, error) {
	return s.pools[id], nil
}
func (s *stubDropPoolRepo) GetByCode(ctx context.Context, code string) (*game_config.DropPool, error) {
	return nil, nil
}
func (s *stubDropPoolRepo) GetByType(ctx context.Context, poolType string) ([]*game_config.DropPool, error) {
	return nil, nil
}
func (s *stubDropPoolRepo) GetPoolItems(ctx context.Context, poolID string) ([]*game_config.DropPoolItem, error) {
	return nil, nil
}
func (s *stubDropPoolRepo) GetPoolItemsByLevel(ctx context.Context, poolID string, playerLevel int) ([]*game_config.DropPoolItem, error) {
	return nil, nil
}
func (s *stubDropPoolRepo) List(ctx context.Context, params interfaces.ListDropPoolParams) ([]*game_config.DropPool, int64, error) {
	return nil, 0, nil
}
func (s *stubDropPoolRepo) Create(ctx context.Context, pool *game_config.DropPool) error { return nil }
func (s *stubDropPoolRepo) Update(ctx context.Context, pool *game_config.DropPool) error { return nil }
func (s *stubDropPoolRepo) Delete(ctx context.Context, poolID string) error              { return nil }
func (s *stubDropPoolRepo) Count(ctx context.Context, params interfaces.ListDropPoolParams) (int64, error) {
	return 0, nil
}
func (s *stubDropPoolRepo) CreatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error {
	return nil
}
func (s *stubDropPoolRepo) GetPoolItemByID(ctx context.Context, poolID, itemID string) (*game_config.DropPoolItem, error) {
	return nil, nil
}
func (s *stubDropPoolRepo) UpdatePoolItem(ctx context.Context, item *game_config.DropPoolItem) error {
	return nil
}
func (s *stubDropPoolRepo) DeletePoolItem(ctx context.Context, poolID, itemID string) error {
	return nil
}
func (s *stubDropPoolRepo) ListPoolItems(ctx context.Context, params interfaces.ListDropPoolItemParams) ([]*game_config.DropPoolItem, int64, error) {
	return nil, 0, nil
}
func (s *stubDropPoolRepo) CountPoolItems(ctx context.Context, params interfaces.ListDropPoolItemParams) (int64, error) {
	return 0, nil
}

type stubTagRelationRepo struct {
	tags map[string][]*game_config.Tag
}

func (s *stubTagRelationRepo) GetByID(ctx context.Context, id string) (*game_config.TagsRelation, error) {
	return nil, nil
}
func (s *stubTagRelationRepo) List(ctx context.Context, params interfaces.TagRelationQueryParams) ([]*game_config.TagsRelation, int64, error) {
	return nil, 0, nil
}
func (s *stubTagRelationRepo) GetEntityTags(ctx context.Context, entityType string, entityID string) ([]*game_config.Tag, error) {
	return s.tags[entityID], nil
}
func (s *stubTagRelationRepo) GetTagEntities(ctx context.Context, tagID string) ([]*game_config.TagsRelation, error) {
	return nil, nil
}
func (s *stubTagRelationRepo) Create(ctx context.Context, relation *game_config.TagsRelation) error {
	return nil
}
func (s *stubTagRelationRepo) Delete(ctx context.Context, relationID string) error { return nil }
func (s *stubTagRelationRepo) DeleteByTagAndEntity(ctx context.Context, tagID, entityType, entityID string) error {
	return nil
}
func (s *stubTagRelationRepo) Exists(ctx context.Context, tagID, entityType, entityID string) (bool, error) {
	return false, nil
}
func (s *stubTagRelationRepo) BatchCreate(ctx context.Context, relations []*game_config.TagsRelation) error {
	return nil
}
func (s *stubTagRelationRepo) DeleteByEntity(ctx context.Context, entityType string, entityID string) error {
	return nil
}

type stubAttributeTypeRepo struct{}

func (s *stubAttributeTypeRepo) GetByID(ctx context.Context, id string) (*game_config.HeroAttributeType, error) {
	return nil, nil
}
func (s *stubAttributeTypeRepo) GetByCode(ctx context.Context, code string) (*game_config.HeroAttributeType, error) {
	return &game_config.HeroAttributeType{
		AttributeCode: code,
	}, nil
}
func (s *stubAttributeTypeRepo) List(ctx context.Context, params interfaces.HeroAttributeTypeQueryParams) ([]*game_config.HeroAttributeType, int64, error) {
	return nil, 0, nil
}
func (s *stubAttributeTypeRepo) Create(ctx context.Context, attrType *game_config.HeroAttributeType) error {
	return nil
}
func (s *stubAttributeTypeRepo) Update(ctx context.Context, attrType *game_config.HeroAttributeType) error {
	return nil
}
func (s *stubAttributeTypeRepo) Delete(ctx context.Context, attrTypeID string) error { return nil }
func (s *stubAttributeTypeRepo) Exists(ctx context.Context, code string) (bool, error) {
	return false, nil
}
func (s *stubAttributeTypeRepo) ListByCategory(ctx context.Context, category string) ([]*game_config.HeroAttributeType, error) {
	return nil, nil
}
