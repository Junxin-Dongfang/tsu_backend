package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/repository/interfaces"
)

func TestDungeonHandler_GetDungeons_DefaultOrder(t *testing.T) {
	handler, repo := newTestDungeonHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/dungeons?min_level=5&max_level=15", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.GetDungeons(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.EqualValues(t, 5, *repo.lastParams.MinLevel)
	require.EqualValues(t, 15, *repo.lastParams.MaxLevel)
	require.Equal(t, "created_at", repo.lastParams.OrderBy)

	var resp struct {
		Data dto.DungeonListResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "created_at", resp.Data.OrderBy)
	require.True(t, resp.Data.OrderDesc)
}

func TestDungeonHandler_GetDungeons_InvalidOrder(t *testing.T) {
	handler, _ := newTestDungeonHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/dungeons?order_by=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, handler.GetDungeons(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func newTestDungeonHandler() (*DungeonHandler, *stubDungeonRepo) {
	repo := &stubDungeonRepo{
		list: []*game_config.Dungeon{
			{
				ID:          "dungeon-1",
				DungeonCode: "D001",
				DungeonName: "测试地城",
				MinLevel:    1,
				MaxLevel:    10,
			},
		},
	}

	svc := service.NewDungeonServiceWithDeps(service.DungeonServiceDeps{
		DungeonRepo: repo,
		DungeonRoomRepo: &stubDungeonRoomRepo{
			rooms: map[string]*game_config.DungeonRoom{
				"room-1": {ID: "room-1"},
			},
		},
	})

	return &DungeonHandler{
		service:    svc,
		respWriter: response.DefaultResponseHandler(),
	}, repo
}

type stubDungeonRepo struct {
	list       []*game_config.Dungeon
	lastParams interfaces.DungeonQueryParams
}

func (s *stubDungeonRepo) List(ctx context.Context, params interfaces.DungeonQueryParams) ([]*game_config.Dungeon, int64, error) {
	s.lastParams = params
	if params.MinLevel == nil {
		min := int16(0)
		params.MinLevel = &min
	}
	return s.list, int64(len(s.list)), nil
}
func (s *stubDungeonRepo) GetByID(ctx context.Context, id string) (*game_config.Dungeon, error) {
	return nil, nil
}
func (s *stubDungeonRepo) GetByCode(ctx context.Context, code string) (*game_config.Dungeon, error) {
	return nil, nil
}
func (s *stubDungeonRepo) Exists(ctx context.Context, code string) (bool, error) { return false, nil }
func (s *stubDungeonRepo) Create(ctx context.Context, dungeon *game_config.Dungeon) error {
	return nil
}
func (s *stubDungeonRepo) Update(ctx context.Context, dungeon *game_config.Dungeon) error {
	return nil
}
func (s *stubDungeonRepo) Delete(ctx context.Context, dungeonID string) error { return nil }

type stubDungeonRoomRepo struct {
	rooms map[string]*game_config.DungeonRoom
}

func (s *stubDungeonRoomRepo) GetByID(ctx context.Context, roomID string) (*game_config.DungeonRoom, error) {
	return s.rooms[roomID], nil
}
func (s *stubDungeonRoomRepo) GetByCode(ctx context.Context, code string) (*game_config.DungeonRoom, error) {
	return nil, nil
}
func (s *stubDungeonRoomRepo) GetByCodes(ctx context.Context, codes []string) ([]*game_config.DungeonRoom, error) {
	return nil, nil
}
func (s *stubDungeonRoomRepo) GetByIDs(ctx context.Context, ids []string) ([]*game_config.DungeonRoom, error) {
	result := make([]*game_config.DungeonRoom, 0, len(ids))
	for _, id := range ids {
		if room, ok := s.rooms[id]; ok {
			result = append(result, room)
		}
	}
	return result, nil
}
func (s *stubDungeonRoomRepo) List(ctx context.Context, params interfaces.DungeonRoomQueryParams) ([]*game_config.DungeonRoom, int64, error) {
	return nil, 0, nil
}
func (s *stubDungeonRoomRepo) Create(ctx context.Context, room *game_config.DungeonRoom) error {
	return nil
}
func (s *stubDungeonRoomRepo) Update(ctx context.Context, room *game_config.DungeonRoom) error {
	return nil
}
func (s *stubDungeonRoomRepo) Delete(ctx context.Context, roomID string) error { return nil }
func (s *stubDungeonRoomRepo) Exists(ctx context.Context, code string) (bool, error) {
	return false, nil
}
func (s *stubDungeonRoomRepo) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	return false, nil
}
