package apitest

import (
	"fmt"
	"strconv"
	"time"
)

// LoginRequest reused by admin/game auth endpoints.
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// LoginResponse captures login payload of admin/game.
type LoginResponse struct {
	Email        string `json:"email"`
	SessionToken string `json:"session_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
}

// RegisterRequest for game registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

// RegisterResponse mirrors game register output.
type RegisterResponse struct {
	Email        string `json:"email"`
	KratosID     string `json:"kratos_id"`
	NeedVerify   bool   `json:"need_verify"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	SessionToken string `json:"session_token,omitempty"`
}

// GetUserResponse for both admin/game user detail endpoints.
type GetUserResponse struct {
	AvatarURL  string `json:"avatar_url"`
	Email      string `json:"email"`
	IsBanned   bool   `json:"is_banned"`
	LastLogin  string `json:"last_login_at"`
	LoginCount int    `json:"login_count"`
	Nickname   string `json:"nickname"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
}

// ClassResponse 描述职业条目。
type ClassResponse struct {
	ID           string `json:"id"`
	ClassCode    string `json:"class_code"`
	ClassName    string `json:"class_name"`
	Description  string `json:"description"`
	Tier         string `json:"tier"`
	DisplayOrder int    `json:"display_order"`
}

// ClassListData 为 GET /game/classes 的 data 字段。
type ClassListData struct {
	List     []ClassResponse `json:"list"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Total    int             `json:"total"`
}

// CreateHeroRequest mirrors /game/heroes create payload.
type CreateHeroRequest struct {
	ClassID     string `json:"class_id"`
	HeroName    string `json:"hero_name"`
	Description string `json:"description,omitempty"`
}

// HeroResponse 描述英雄基础信息。
type HeroResponse struct {
	ID                  string `json:"id"`
	HeroName            string `json:"hero_name"`
	ClassID             string `json:"class_id"`
	CurrentLevel        int    `json:"current_level"`
	ExperienceTotal     int64  `json:"experience_total"`
	ExperienceSpent     int64  `json:"experience_spent"`
	ExperienceAvailable int64  `json:"experience_available"`
}

// CreateTeamRequest matches /game/teams create payload.
type CreateTeamRequest struct {
	HeroID      string `json:"hero_id"`
	TeamName    string `json:"team_name"`
	Description string `json:"description,omitempty"`
}

// UpdateTeamInfoRequest updates name/description.
type UpdateTeamInfoRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// DistributeGoldRequest for warehouse gold distribution.
type DistributeGoldRequest struct {
	DistributorID string           `json:"distributor_id"`
	Distributions map[string]int64 `json:"distributions"`
}

// InviteMemberRequest matches队长邀请成员的请求。
type InviteMemberRequest struct {
	TeamID        string  `json:"team_id"`
	InviterHeroID string  `json:"inviter_hero_id"`
	InviteeHeroID string  `json:"invitee_hero_id"`
	Message       *string `json:"message,omitempty"`
}

// AcceptInvitationRequest for /teams/invite/accept.
type AcceptInvitationRequest struct {
	InvitationID string `json:"invitation_id"`
	HeroID       string `json:"hero_id"`
}

// ApproveInvitationRequest mirrors审批接口 payload.
type ApproveInvitationRequest struct {
	InvitationID string `json:"invitation_id"`
	HeroID       string `json:"hero_id"`
	Approved     bool   `json:"approved"`
}

// PromoteToAdminPayload for /teams/members/promote.
type PromoteToAdminPayload struct {
	TeamID       string `json:"team_id"`
	TargetHeroID string `json:"target_hero_id"`
	LeaderHeroID string `json:"leader_hero_id"`
}

// TeamJoinApprovePayload request body for /teams/join/approve.
type TeamJoinApprovePayload struct {
	RequestID string `json:"request_id"`
	HeroID    string `json:"hero_id"`
	Approved  bool   `json:"approved"`
}

// DungeonProgressResponse 简化团队副本进度数据。
type DungeonProgressResponse struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	DungeonID string `json:"dungeon_id"`
	Status    string `json:"status"`
}

// AssignRolesRequest 用于后台角色分配。
type AssignRolesRequest struct {
	RoleCodes []string `json:"role_codes"`
}

// TeamResponse 描述团队基础信息。
type TeamResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	LeaderHeroID string `json:"leader_hero_id"`
	MaxMembers   int    `json:"max_members"`
}

// WarehouseResponse 描述团队仓库摘要。
type WarehouseResponse struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	Gold      int64  `json:"gold_amount"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SelectDungeonRequest chooses a dungeon for hero.
type SelectDungeonRequest struct {
	DungeonID string `json:"dungeon_id"`
	HeroID    string `json:"hero_id"`
}

// EnterDungeonRequest starts run with optional dungeon selection.
type EnterDungeonRequest struct {
	DungeonID string `json:"dungeon_id,omitempty"`
	HeroID    string `json:"hero_id"`
}

// CompleteDungeonRequest notifies completion.
type CompleteDungeonRequest struct {
	DungeonID string       `json:"dungeon_id"`
	HeroID    string       `json:"hero_id"`
	Loot      *LootPayload `json:"loot,omitempty"`
}

// SelectTeamDungeonRequest describes leader selecting a dungeon.
type SelectTeamDungeonRequest struct {
	TeamID    string `json:"team_id"`
	HeroID    string `json:"hero_id"`
	DungeonID string `json:"dungeon_id"`
}

// LootPayload optional structure for rewards.
type LootPayload struct {
	Gold  int64      `json:"gold,omitempty"`
	Items []LootItem `json:"items,omitempty"`
}

// LootItem describes a single loot entry.
type LootItem struct {
	ItemID   string `json:"item_id"`
	ItemType string `json:"item_type,omitempty"`
	Quantity int    `json:"quantity,omitempty"`
}

// AllocateAttributeRequest wraps attribute allocation inputs.
type AllocateAttributeRequest struct {
	AttributeCode string `json:"attribute_code"`
	PointsToAdd   int    `json:"points_to_add"`
}

// AdminUserInfo mirrors单个后台用户信息。
type AdminUserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	IsBanned   bool   `json:"is_banned"`
	LoginCount int    `json:"login_count"`
}

// AdminUserListResponse is /admin/users data payload.
type AdminUserListResponse struct {
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
	Users      []AdminUserInfo `json:"users"`
}

// AdminRole 表示单个角色。
type AdminRole struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// AdminRoleListResponse 为角色列表的 data。
type AdminRoleListResponse struct {
	Roles []AdminRole `json:"roles"`
}

// CreateDropPoolRequest mirrors admin create drop pool API.
type CreateDropPoolRequest struct {
	PoolCode        string  `json:"pool_code"`
	PoolName        string  `json:"pool_name"`
	PoolType        string  `json:"pool_type"`
	Description     *string `json:"description,omitempty"`
	MinDrops        int16   `json:"min_drops"`
	MaxDrops        int16   `json:"max_drops"`
	GuaranteedDrops int16   `json:"guaranteed_drops"`
}

// DropPoolResponse captures drop pool payload.
type DropPoolResponse struct {
	ID              string  `json:"id"`
	PoolCode        string  `json:"pool_code"`
	PoolName        string  `json:"pool_name"`
	PoolType        string  `json:"pool_type"`
	Description     *string `json:"description,omitempty"`
	MinDrops        int16   `json:"min_drops"`
	MaxDrops        int16   `json:"max_drops"`
	GuaranteedDrops int16   `json:"guaranteed_drops"`
}

// AddDropPoolItemRequest adds item into pool.
type AddDropPoolItemRequest struct {
	ItemID         string   `json:"item_id"`
	DropWeight     *int     `json:"drop_weight,omitempty"`
	DropRate       *float64 `json:"drop_rate,omitempty"`
	QualityWeights string   `json:"quality_weights,omitempty"`
	MinQuantity    int16    `json:"min_quantity"`
	MaxQuantity    int16    `json:"max_quantity"`
	MinLevel       *int16   `json:"min_level,omitempty"`
	MaxLevel       *int16   `json:"max_level,omitempty"`
}

// DropPoolItemResponse mirrors admin add/update item response.
type DropPoolItemResponse struct {
	ID          string `json:"id"`
	DropPoolID  string `json:"drop_pool_id"`
	ItemID      string `json:"item_id"`
	ItemCode    string `json:"item_code"`
	ItemName    string `json:"item_name"`
	MinQuantity int16  `json:"min_quantity"`
	MaxQuantity int16  `json:"max_quantity"`
}

// AdminItem represents an item config entry.
type AdminItem struct {
	ID          string `json:"id"`
	ItemCode    string `json:"item_code"`
	ItemName    string `json:"item_name"`
	ItemType    string `json:"item_type"`
	ItemQuality string `json:"item_quality"`
	ItemLevel   int    `json:"item_level"`
}

// AdminItemList is the paged list shape for /admin/items.
type AdminItemList struct {
	Items    []AdminItem `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// UniqueSuffix returns a short unique string for test data.
func UniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// UniqueCredentials generates test user credentials.
func UniqueCredentials(prefix string) (username, email, password string) {
	suffix := UniqueSuffix()
	username = fmt.Sprintf("%s-%s", prefix, suffix)
	email = fmt.Sprintf("%s-%s@example.com", prefix, suffix)
	password = "Admin123!"
	return
}
