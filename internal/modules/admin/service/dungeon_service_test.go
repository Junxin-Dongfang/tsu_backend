package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_config"
	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/pkg/xerrors"
	"tsu-self/internal/repository/interfaces"
)

type mockDungeonRoomRepo struct {
	rooms map[string]*game_config.DungeonRoom
}

func (m *mockDungeonRoomRepo) GetByID(ctx context.Context, roomID string) (*game_config.DungeonRoom, error) {
	return m.rooms[roomID], nil
}

func (m *mockDungeonRoomRepo) GetByCode(ctx context.Context, code string) (*game_config.DungeonRoom, error) {
	return nil, nil
}

func (m *mockDungeonRoomRepo) GetByCodes(ctx context.Context, codes []string) ([]*game_config.DungeonRoom, error) {
	return nil, nil
}

func (m *mockDungeonRoomRepo) GetByIDs(ctx context.Context, ids []string) ([]*game_config.DungeonRoom, error) {
	result := make([]*game_config.DungeonRoom, 0, len(ids))
	for _, id := range ids {
		if room, ok := m.rooms[id]; ok {
			result = append(result, room)
		}
	}
	return result, nil
}

func (m *mockDungeonRoomRepo) List(ctx context.Context, params interfaces.DungeonRoomQueryParams) ([]*game_config.DungeonRoom, int64, error) {
	return nil, 0, nil
}

func (m *mockDungeonRoomRepo) Create(ctx context.Context, room *game_config.DungeonRoom) error {
	return nil
}

func (m *mockDungeonRoomRepo) Update(ctx context.Context, room *game_config.DungeonRoom) error {
	return nil
}

func (m *mockDungeonRoomRepo) Delete(ctx context.Context, roomID string) error {
	return nil
}

func (m *mockDungeonRoomRepo) Exists(ctx context.Context, code string) (bool, error) {
	return false, nil
}

func (m *mockDungeonRoomRepo) ExistsExcludingID(ctx context.Context, code string, excludeID string) (bool, error) {
	return false, nil
}

func TestDungeonService_validateRoomSequenceMissingSort(t *testing.T) {
	svc := &DungeonService{
		dungeonRoomRepo: &mockDungeonRoomRepo{
			rooms: map[string]*game_config.DungeonRoom{
				"room-1": {ID: "room-1"},
				"room-2": {ID: "room-2"},
			},
		},
	}

	sequence := []dto.RoomSequenceItem{
		{RoomID: "room-1", Sort: 1},
		{RoomID: "room-2", Sort: 3},
	}

	err := svc.validateRoomSequence(context.Background(), sequence)
	require.Error(t, err)
	appErr, ok := err.(*xerrors.AppError)
	require.True(t, ok, "error should be AppError")
	require.Equal(t, "房间序列缺少排序值: 2", appErr.Context.Metadata["user_message"])
}
