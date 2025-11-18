package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// ketoPermissionClient defines the subset of KetoClient behavior TeamPermissionService depends on.
type ketoPermissionClient interface {
	AddTeamMember(ctx context.Context, teamID, heroID, role string) error
	RemoveTeamMember(ctx context.Context, teamID, heroID, role string) error
	UpdateTeamMemberRole(ctx context.Context, teamID, heroID, oldRole, newRole string) error
	CheckTeamPermission(ctx context.Context, teamID, permission, heroID string) (bool, error)
	CheckTeamMemberRole(ctx context.Context, teamID, heroID string) (string, bool, error)
}

const permissionCacheTTL = 5 * time.Minute

var errPermissionCacheMiss = errors.New("permission cache miss")

type permissionCacheClient interface {
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	DeleteKey(ctx context.Context, keys ...string) error
}

// TeamPermissionService 团队权限服务
type TeamPermissionService struct {
	db              *sql.DB
	teamMemberRepo  interfaces.TeamMemberRepository
	ketoClient      ketoPermissionClient
	permissionCache permissionCacheClient
}

// NewTeamPermissionService 创建团队权限服务
func NewTeamPermissionService(db *sql.DB, ketoClient *client.KetoClient, cache permissionCacheClient) *TeamPermissionService {
	var kc ketoPermissionClient
	if ketoClient != nil {
		kc = ketoClient
	}

	return &TeamPermissionService{
		db:              db,
		teamMemberRepo:  impl.NewTeamMemberRepository(db),
		ketoClient:      kc,
		permissionCache: cache,
	}
}

// SyncMemberToKeto 同步成员权限到 Keto
func (s *TeamPermissionService) SyncMemberToKeto(ctx context.Context, member *game_runtime.TeamMember) error {
	s.invalidatePermissionCache(ctx, member.TeamID, member.HeroID)

	// 如果 ketoClient 未初始化,跳过同步
	if s.ketoClient == nil {
		fmt.Printf("Warning: Keto client not initialized, skipping sync for TeamID: %s, HeroID: %s\n",
			member.TeamID, member.HeroID)
		return nil
	}

	// 添加成员关系到 Keto
	return s.ketoClient.AddTeamMember(ctx, member.TeamID, member.HeroID, member.Role)
}

// DeleteMemberFromKeto 从 Keto 删除成员权限
func (s *TeamPermissionService) DeleteMemberFromKeto(ctx context.Context, teamID, heroID string) error {
	s.invalidatePermissionCache(ctx, teamID, heroID)

	if s.ketoClient == nil {
		return nil
	}

	// 需要知道成员的当前角色才能删除
	// 从数据库查询或尝试删除所有可能的角色
	roles := []string{"leader", "admin", "member"}
	for _, role := range roles {
		_ = s.ketoClient.RemoveTeamMember(ctx, teamID, heroID, role)
	}

	return nil
}

// UpdateMemberRoleInKeto 更新成员角色在 Keto 中的权限
func (s *TeamPermissionService) UpdateMemberRoleInKeto(ctx context.Context, teamID, heroID, oldRole, newRole string) error {
	s.invalidatePermissionCache(ctx, teamID, heroID)

	if s.ketoClient == nil {
		return nil
	}

	// 使用 Keto Client 的更新方法
	return s.ketoClient.UpdateTeamMemberRole(ctx, teamID, heroID, oldRole, newRole)
}

// GetMemberRole 获取成员角色（优先 Keto，失败则回退到数据库）
func (s *TeamPermissionService) GetMemberRole(ctx context.Context, teamID, heroID string) (string, bool, error) {
	if teamID == "" || heroID == "" {
		return "", false, fmt.Errorf("teamID 和 heroID 不能为空")
	}

	if s.ketoClient != nil {
		role, exists, err := s.ketoClient.CheckTeamMemberRole(ctx, teamID, heroID)
		if err == nil {
			if exists {
				return role, true, nil
			}
			fmt.Printf("Info: Keto missing member relation, falling back to DB (team=%s, hero=%s)\n", teamID, heroID)
		} else {
			fmt.Printf("Warning: Keto member role check failed, falling back to database: %v\n", err)
		}
	}

	member, err := s.teamMemberRepo.GetByTeamAndHero(ctx, teamID, heroID)
	if err != nil {
		if errors.Is(err, interfaces.ErrTeamMemberNotFound) {
			return "", false, nil
		}
		return "", false, err
	}

	return member.Role, true, nil
}

// CheckPermission 检查权限
func (s *TeamPermissionService) CheckPermission(ctx context.Context, teamID, heroID, permission string) (bool, error) {
	// 如果 Keto 客户端可用,使用 Keto 检查
	if s.ketoClient != nil {
		allowed, err := s.ketoClient.CheckTeamPermission(ctx, teamID, permission, heroID)
		if err != nil {
			// Keto 检查失败,降级到数据库检查
			fmt.Printf("Warning: Keto permission check failed, falling back to database: %v\n", err)
		} else {
			return allowed, nil
		}
	}

	// 降级方案：查询成员角色
	role, exists, err := s.GetMemberRole(ctx, teamID, heroID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	// 简单的权限映射
	permissions := map[string][]string{
		"leader": {
			"select_dungeon", "enter_dungeon", "abandon_dungeon",
			"kick_member", "kick_admin", "appoint_admin", "demote_admin",
			"view_warehouse", "distribute_loot", "update_team_info", "disband_team",
			"approve_join_request", "approve_invitation",
		},
		"admin": {
			"select_dungeon", "enter_dungeon", "abandon_dungeon",
			"kick_member", "view_warehouse", "distribute_loot",
			"approve_join_request", "approve_invitation",
		},
		"member": {
			"invite_member", "view_team_info", "leave_team",
		},
	}

	rolePermissions, ok := permissions[role]
	if !ok {
		return false, nil
	}

	for _, p := range rolePermissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermissionWithCache 带缓存的权限检查
func (s *TeamPermissionService) CheckPermissionWithCache(ctx context.Context, teamID, heroID, permission string) (bool, error) {
	// 没有缓存时直接降级
	if s.permissionCache == nil {
		return s.CheckPermission(ctx, teamID, heroID, permission)
	}

	cacheKey := buildPermissionCacheKey(teamID, heroID)
	cached, err := s.loadPermissionCache(ctx, cacheKey)
	if err == nil {
		if allowed, ok := cached[permission]; ok {
			return allowed, nil
		}
	} else if !errors.Is(err, errPermissionCacheMiss) {
		fmt.Printf("Warning: Failed to read team permission cache (team=%s hero=%s): %v\n", teamID, heroID, err)
	}

	allowed, err := s.CheckPermission(ctx, teamID, heroID, permission)
	if err != nil {
		return false, err
	}

	if cached == nil {
		cached = make(map[string]bool)
	}
	cached[permission] = allowed

	if err := s.savePermissionCache(ctx, cacheKey, cached); err != nil {
		fmt.Printf("Warning: Failed to update team permission cache (team=%s hero=%s): %v\n", teamID, heroID, err)
	}

	return allowed, nil
}

// ClearPermissionCache 清除权限缓存
func (s *TeamPermissionService) ClearPermissionCache(ctx context.Context, teamID, heroID string) error {
	return s.removePermissionCache(ctx, teamID, heroID)
}

// CheckConsistency 权限一致性检查（定时任务调用）
func (s *TeamPermissionService) CheckConsistency(ctx context.Context) error {
	if s.ketoClient == nil {
		fmt.Println("[TeamPermissionService] Keto 客户端未初始化，跳过一致性检查")
		return nil
	}

	members, err := s.teamMemberRepo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("查询团队成员失败: %w", err)
	}

	if len(members) == 0 {
		return nil
	}

	var repaired, failed int
	for _, member := range members {
		ketoRole, exists, err := s.ketoClient.CheckTeamMemberRole(ctx, member.TeamID, member.HeroID)
		if err != nil {
			fmt.Printf("[TeamPermissionService] 检查 Keto 角色失败 team=%s hero=%s error=%v\n", member.TeamID, member.HeroID, err)
			continue
		}

		if !exists || ketoRole != member.Role {
			fmt.Printf("[TeamPermissionService] 权限不一致 team=%s hero=%s db_role=%s keto_role=%s -> 尝试修复\n",
				member.TeamID, member.HeroID, member.Role, ketoRole)
			if err := s.SyncMemberToKeto(ctx, member); err != nil {
				failed++
				fmt.Printf("[TeamPermissionService] 修复失败 team=%s hero=%s, 请运行 scripts/development/init_keto_from_db.sh 或手动排查: %v\n",
					member.TeamID, member.HeroID, err)
				continue
			}
			repaired++
		}
	}

	if failed > 0 {
		fmt.Printf("[TeamPermissionService] 权限一致性检查完成，成功修复 %d 项，%d 项仍需人工处理。建议执行 scripts/development/init_keto_from_db.sh 以重新生成关系。\n",
			repaired, failed)
	} else if repaired > 0 {
		fmt.Printf("[TeamPermissionService] 权限一致性检查完成，成功修复 %d 项。\n", repaired)
	} else {
		fmt.Println("[TeamPermissionService] 权限一致性检查完成，未发现不一致数据。")
	}

	return nil
}

// SyncAllMembersToKeto 同步所有成员到 Keto（初始化或修复用）
func (s *TeamPermissionService) SyncAllMembersToKeto(ctx context.Context, teamID string) error {
	// 1. 查询团队所有成员
	members, err := s.teamMemberRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("查询团队成员失败: %w", err)
	}

	// 2. 逐个同步到 Keto
	for _, member := range members {
		if err := s.SyncMemberToKeto(ctx, member); err != nil {
			fmt.Printf("同步成员 %s 到 Keto 失败: %v\n", member.HeroID, err)
			continue
		}
	}

	return nil
}

func (s *TeamPermissionService) loadPermissionCache(ctx context.Context, cacheKey string) (map[string]bool, error) {
	if s.permissionCache == nil {
		return nil, errPermissionCacheMiss
	}

	raw, err := s.permissionCache.GetString(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errPermissionCacheMiss
		}
		return nil, err
	}

	if raw == "" {
		return nil, errPermissionCacheMiss
	}

	var payload map[string]bool
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (s *TeamPermissionService) savePermissionCache(ctx context.Context, cacheKey string, payload map[string]bool) error {
	if s.permissionCache == nil {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.permissionCache.SetWithTTL(ctx, cacheKey, data, permissionCacheTTL)
}

func buildPermissionCacheKey(teamID, heroID string) string {
	return fmt.Sprintf("team:perm:%s:%s", teamID, heroID)
}

func (s *TeamPermissionService) removePermissionCache(ctx context.Context, teamID, heroID string) error {
	if s.permissionCache == nil {
		return nil
	}

	cacheKey := buildPermissionCacheKey(teamID, heroID)
	return s.permissionCache.DeleteKey(ctx, cacheKey)
}

func (s *TeamPermissionService) invalidatePermissionCache(ctx context.Context, teamID, heroID string) {
	if err := s.removePermissionCache(ctx, teamID, heroID); err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("Warning: Failed to clear permission cache (team=%s hero=%s): %v\n", teamID, heroID, err)
	}
}
