package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/repository/interfaces"
)

type ketoCall struct {
	teamID string
	heroID string
	role   string
}

type ketoUpdateCall struct {
	teamID  string
	heroID  string
	oldRole string
	newRole string
}

type ketoCheckCall struct {
	teamID     string
	heroID     string
	permission string
}

type ketoRoleCall struct {
	teamID string
	heroID string
}

type fakeKetoClient struct {
	addCalls      []ketoCall
	removeCalls   []ketoCall
	updateCalls   []ketoUpdateCall
	checkCalls    []ketoCheckCall
	checkResp     bool
	checkErr      error
	addErrForHero map[string]error
	roleCalls     []ketoRoleCall
	roleResp      string
	roleExists    bool
	roleErr       error
}

func (f *fakeKetoClient) AddTeamMember(_ context.Context, teamID, heroID, role string) error {
	f.addCalls = append(f.addCalls, ketoCall{teamID: teamID, heroID: heroID, role: role})
	if f.addErrForHero != nil {
		if err, ok := f.addErrForHero[heroID]; ok {
			return err
		}
	}
	return nil
}

func (f *fakeKetoClient) RemoveTeamMember(_ context.Context, teamID, heroID, role string) error {
	f.removeCalls = append(f.removeCalls, ketoCall{teamID: teamID, heroID: heroID, role: role})
	return nil
}

func (f *fakeKetoClient) UpdateTeamMemberRole(_ context.Context, teamID, heroID, oldRole, newRole string) error {
	f.updateCalls = append(f.updateCalls, ketoUpdateCall{
		teamID: teamID, heroID: heroID, oldRole: oldRole, newRole: newRole,
	})
	return nil
}

func (f *fakeKetoClient) CheckTeamPermission(_ context.Context, teamID, permission, heroID string) (bool, error) {
	f.checkCalls = append(f.checkCalls, ketoCheckCall{
		teamID: teamID, heroID: heroID, permission: permission,
	})
	if f.checkErr != nil {
		return false, f.checkErr
	}
	return f.checkResp, nil
}

func (f *fakeKetoClient) CheckTeamMemberRole(_ context.Context, teamID, heroID string) (string, bool, error) {
	f.roleCalls = append(f.roleCalls, ketoRoleCall{teamID: teamID, heroID: heroID})
	if f.roleErr != nil {
		return "", false, f.roleErr
	}
	return f.roleResp, f.roleExists, nil
}

type fakeTeamMemberRepo struct {
	members    map[string]*game_runtime.TeamMember
	listByTeam map[string][]*game_runtime.TeamMember
	getErr     error
	listErr    error
	getCalls   int
}

func (f *fakeTeamMemberRepo) key(teamID, heroID string) string {
	return fmt.Sprintf("%s:%s", teamID, heroID)
}

func (f *fakeTeamMemberRepo) Create(context.Context, boil.ContextExecutor, *game_runtime.TeamMember) error {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) GetByID(context.Context, string) (*game_runtime.TeamMember, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) GetByTeamAndHero(_ context.Context, teamID, heroID string) (*game_runtime.TeamMember, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	if member, ok := f.members[f.key(teamID, heroID)]; ok {
		return member, nil
	}
	return nil, interfaces.ErrTeamMemberNotFound
}

func (f *fakeTeamMemberRepo) GetLeaderTeam(context.Context, string) (*game_runtime.TeamMember, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) ListByTeam(_ context.Context, teamID string) ([]*game_runtime.TeamMember, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	members := f.listByTeam[teamID]
	result := make([]*game_runtime.TeamMember, len(members))
	copy(result, members)
	return result, nil
}

func (f *fakeTeamMemberRepo) ListByHero(context.Context, string) ([]*game_runtime.TeamMember, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) ListAll(context.Context) ([]*game_runtime.TeamMember, error) {
	result := make([]*game_runtime.TeamMember, 0, len(f.members))
	for _, member := range f.members {
		copyMember := *member
		result = append(result, &copyMember)
	}
	return result, nil
}

func (f *fakeTeamMemberRepo) Update(context.Context, boil.ContextExecutor, *game_runtime.TeamMember) error {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) Delete(context.Context, boil.ContextExecutor, string) error {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) UpdateRole(context.Context, boil.ContextExecutor, string, string, string) error {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) UpdateLastActive(context.Context, boil.ContextExecutor, string, string) error {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) GetEarliestAdmin(context.Context, string) (*game_runtime.TeamMember, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) GetEarliestMember(context.Context, string) (*game_runtime.TeamMember, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) CountByTeam(context.Context, string) (int64, error) {
	panic("not implemented")
}

func (f *fakeTeamMemberRepo) GetByIDForUpdate(context.Context, *sql.Tx, string) (*game_runtime.TeamMember, error) {
	panic("not implemented")
}

type fakePermissionCache struct {
	values    map[string]string
	setErr    error
	getErr    error
	deleteErr error
}

func newFakePermissionCache() *fakePermissionCache {
	return &fakePermissionCache{
		values: make(map[string]string),
	}
}

func (f *fakePermissionCache) SetWithTTL(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	if f.setErr != nil {
		return f.setErr
	}
	switch v := value.(type) {
	case []byte:
		f.values[key] = string(v)
	case string:
		f.values[key] = v
	default:
		f.values[key] = fmt.Sprint(v)
	}
	return nil
}

func (f *fakePermissionCache) GetString(_ context.Context, key string) (string, error) {
	if f.getErr != nil {
		return "", f.getErr
	}
	val, ok := f.values[key]
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (f *fakePermissionCache) DeleteKey(_ context.Context, keys ...string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for _, key := range keys {
		delete(f.values, key)
	}
	return nil
}

func TestTeamPermissionService_SyncMemberToKeto_ClearsCache(t *testing.T) {
	ctx := context.Background()
	cache := newFakePermissionCache()
	cacheKey := buildPermissionCacheKey("team-1", "hero-1")
	cache.values[cacheKey] = `{"view_team_info":true}`

	keto := &fakeKetoClient{}
	svc := &TeamPermissionService{
		permissionCache: cache,
		ketoClient:      keto,
	}

	err := svc.SyncMemberToKeto(ctx, &game_runtime.TeamMember{
		TeamID: "team-1",
		HeroID: "hero-1",
		Role:   "member",
	})
	require.NoError(t, err)
	_, exists := cache.values[cacheKey]
	assert.False(t, exists, "sync 应清除权限缓存以避免脏读")
}

func TestTeamPermissionService_CheckPermissionFallback(t *testing.T) {
	repo := &fakeTeamMemberRepo{
		members: map[string]*game_runtime.TeamMember{
			"team-1:hero-leader": {TeamID: "team-1", HeroID: "hero-leader", Role: "leader"},
			"team-1:hero-admin":  {TeamID: "team-1", HeroID: "hero-admin", Role: "admin"},
			"team-1:hero-member": {TeamID: "team-1", HeroID: "hero-member", Role: "member"},
		},
	}
	svc := &TeamPermissionService{
		teamMemberRepo: repo,
	}
	ctx := context.Background()

	tests := []struct {
		name        string
		heroID      string
		permission  string
		wantAllowed bool
	}{
		{"leader can disband", "hero-leader", "disband_team", true},
		{"leader can view warehouse", "hero-leader", "view_warehouse", true},
		{"admin cannot disband", "hero-admin", "disband_team", false},
		{"admin can view warehouse", "hero-admin", "view_warehouse", true},
		{"member cannot distribute", "hero-member", "distribute_loot", false},
		{"member can view info", "hero-member", "view_team_info", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := svc.CheckPermission(ctx, "team-1", tt.heroID, tt.permission)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}

	repo.getErr = errors.New("db unavailable")
	_, err := svc.CheckPermission(ctx, "team-1", "hero-leader", "select_dungeon")
	assert.Error(t, err)
}

func TestTeamPermissionService_CheckPermission_UsesKeto(t *testing.T) {
	ctx := context.Background()
	repo := &fakeTeamMemberRepo{
		members: map[string]*game_runtime.TeamMember{
			"team-1:hero-admin":  {TeamID: "team-1", HeroID: "hero-admin", Role: "admin"},
			"team-1:hero-member": {TeamID: "team-1", HeroID: "hero-member", Role: "member"},
		},
	}

	t.Run("uses keto result when available", func(t *testing.T) {
		keto := &fakeKetoClient{checkResp: true}
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		allowed, err := svc.CheckPermission(ctx, "team-1", "hero-admin", "kick_member")
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 0, repo.getCalls) // 未触发数据库回退
		assert.Len(t, keto.checkCalls, 1)
	})

	t.Run("falls back when keto fails", func(t *testing.T) {
		keto := &fakeKetoClient{
			checkErr: errors.New("keto down"),
			roleErr:  errors.New("keto membership down"),
		}
		repo.getCalls = 0
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		allowed, err := svc.CheckPermission(ctx, "team-1", "hero-member", "view_team_info")
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, 1, repo.getCalls) // 触发数据库回退
	})
}

func TestTeamPermissionService_KetoSyncOperations(t *testing.T) {
	ctx := context.Background()
	member := &game_runtime.TeamMember{
		TeamID: "team-1",
		HeroID: "hero-1",
		Role:   "admin",
	}

	t.Run("no keto client gracefully returns", func(t *testing.T) {
		svc := &TeamPermissionService{}
		assert.NoError(t, svc.SyncMemberToKeto(ctx, member))
		assert.NoError(t, svc.DeleteMemberFromKeto(ctx, "team-1", "hero-1"))
		assert.NoError(t, svc.UpdateMemberRoleInKeto(ctx, "team-1", "hero-1", "member", "admin"))
	})

	t.Run("keto client receives all mutations", func(t *testing.T) {
		keto := &fakeKetoClient{}
		svc := &TeamPermissionService{ketoClient: keto}

		require.NoError(t, svc.SyncMemberToKeto(ctx, member))
		require.NoError(t, svc.DeleteMemberFromKeto(ctx, "team-1", "hero-1"))
		require.NoError(t, svc.UpdateMemberRoleInKeto(ctx, "team-1", "hero-1", "member", "admin"))

		assert.Equal(t, []ketoCall{{teamID: "team-1", heroID: "hero-1", role: "admin"}}, keto.addCalls)
		assert.Equal(t, []ketoCall{
			{teamID: "team-1", heroID: "hero-1", role: "leader"},
			{teamID: "team-1", heroID: "hero-1", role: "admin"},
			{teamID: "team-1", heroID: "hero-1", role: "member"},
		}, keto.removeCalls)
		assert.Equal(t, []ketoUpdateCall{
			{teamID: "team-1", heroID: "hero-1", oldRole: "member", newRole: "admin"},
		}, keto.updateCalls)
	})
}

func TestTeamPermissionService_SyncAllMembersToKeto(t *testing.T) {
	ctx := context.Background()

	t.Run("syncs all members", func(t *testing.T) {
		keto := &fakeKetoClient{}
		repo := &fakeTeamMemberRepo{
			listByTeam: map[string][]*game_runtime.TeamMember{
				"team-1": {
					{TeamID: "team-1", HeroID: "hero-1", Role: "leader"},
					{TeamID: "team-1", HeroID: "hero-2", Role: "member"},
				},
			},
		}
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		require.NoError(t, svc.SyncAllMembersToKeto(ctx, "team-1"))
		assert.Len(t, keto.addCalls, 2)
	})

	t.Run("continues when syncing a member fails", func(t *testing.T) {
		keto := &fakeKetoClient{
			addErrForHero: map[string]error{
				"hero-1": errors.New("permission sync failed"),
			},
		}
		repo := &fakeTeamMemberRepo{
			listByTeam: map[string][]*game_runtime.TeamMember{
				"team-1": {
					{TeamID: "team-1", HeroID: "hero-1", Role: "leader"},
					{TeamID: "team-1", HeroID: "hero-2", Role: "member"},
				},
			},
		}
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		require.NoError(t, svc.SyncAllMembersToKeto(ctx, "team-1"))
		// 两个成员都尝试同步，即使第一个失败
		assert.Len(t, keto.addCalls, 2)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		keto := &fakeKetoClient{}
		repo := &fakeTeamMemberRepo{
			listByTeam: map[string][]*game_runtime.TeamMember{},
			listErr:    errors.New("query failed"),
		}
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		err := svc.SyncAllMembersToKeto(ctx, "team-1")
		assert.Error(t, err)
	})
}

func TestTeamPermissionService_CheckPermissionWithCacheAndHelpers(t *testing.T) {
	ctx := context.Background()
	repo := &fakeTeamMemberRepo{
		members: map[string]*game_runtime.TeamMember{
			"team-1:hero-1": {TeamID: "team-1", HeroID: "hero-1", Role: "member"},
		},
	}
	cache := newFakePermissionCache()
	svc := &TeamPermissionService{
		teamMemberRepo:  repo,
		permissionCache: cache,
	}

	allowed, err := svc.CheckPermissionWithCache(ctx, "team-1", "hero-1", "view_team_info")
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 1, repo.getCalls)

	cacheKey := buildPermissionCacheKey("team-1", "hero-1")
	if _, ok := cache.values[cacheKey]; !ok {
		t.Fatalf("expected cache to contain key %s", cacheKey)
	}

	// 第二次调用应该直接命中缓存, 即使仓储返回错误
	repo.getErr = errors.New("repo should not be called when cache hit")
	allowed, err = svc.CheckPermissionWithCache(ctx, "team-1", "hero-1", "view_team_info")
	require.NoError(t, err)
	assert.True(t, allowed)
	repo.getErr = nil
	assert.Equal(t, 1, repo.getCalls)

	require.NoError(t, svc.ClearPermissionCache(ctx, "team-1", "hero-1"))
	if _, ok := cache.values[cacheKey]; ok {
		t.Fatalf("cache key %s should have been cleared", cacheKey)
	}

	_, err = svc.CheckPermissionWithCache(ctx, "team-1", "hero-1", "view_team_info")
	require.NoError(t, err)
	assert.Greater(t, repo.getCalls, 1)

	assert.NoError(t, svc.CheckConsistency(ctx))
}

func TestTeamPermissionService_GetMemberRole(t *testing.T) {
	ctx := context.Background()

	t.Run("uses keto result when available", func(t *testing.T) {
		keto := &fakeKetoClient{
			roleResp:   "admin",
			roleExists: true,
		}
		svc := &TeamPermissionService{
			ketoClient:     keto,
			teamMemberRepo: &fakeTeamMemberRepo{},
		}

		role, exists, err := svc.GetMemberRole(ctx, "team-1", "hero-1")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "admin", role)
		assert.Len(t, keto.roleCalls, 1)
	})

	t.Run("falls back to repository when keto fails", func(t *testing.T) {
		keto := &fakeKetoClient{
			roleErr: errors.New("keto unavailable"),
		}
		repo := &fakeTeamMemberRepo{
			members: map[string]*game_runtime.TeamMember{
				"team-1:hero-1": {TeamID: "team-1", HeroID: "hero-1", Role: "member"},
			},
		}
		svc := &TeamPermissionService{
			teamMemberRepo: repo,
			ketoClient:     keto,
		}

		role, exists, err := svc.GetMemberRole(ctx, "team-1", "hero-1")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "member", role)
	})

	t.Run("returns not found without error", func(t *testing.T) {
		svc := &TeamPermissionService{
			teamMemberRepo: &fakeTeamMemberRepo{},
		}

		role, exists, err := svc.GetMemberRole(ctx, "team-1", "hero-missing")
		require.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, "", role)
	})
}

func TestTeamPermissionService_NewServiceAllowsNilDeps(t *testing.T) {
	svc := NewTeamPermissionService(nil, nil, nil)
	require.NotNil(t, svc)
	assert.Nil(t, svc.ketoClient)
	assert.Nil(t, svc.permissionCache)
	assert.NotNil(t, svc.teamMemberRepo)
}
