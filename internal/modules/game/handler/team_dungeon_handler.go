package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/labstack/echo/v4"

	"tsu-self/internal/entity/game_runtime"
	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// TeamDungeonHandler 团队地城 HTTP Handler
type TeamDungeonHandler struct {
	dungeonService *service.TeamDungeonService
	respWriter     response.Writer
}

// NewTeamDungeonHandler 创建 Handler
func NewTeamDungeonHandler(serviceContainer *service.ServiceContainer, respWriter response.Writer) *TeamDungeonHandler {
	return &TeamDungeonHandler{
		dungeonService: serviceContainer.GetTeamDungeonService(),
		respWriter:     respWriter,
	}
}

// ==================== 请求/响应模型 ====================

type selectDungeonRequest struct {
	HeroID    string `json:"hero_id" validate:"required"`
	DungeonID string `json:"dungeon_id" validate:"required"`
}

type enterDungeonRequest struct {
	HeroID    string `json:"hero_id" validate:"required"`
	DungeonID string `json:"dungeon_id,omitempty"`
}

type completeDungeonRequest struct {
	HeroID    string      `json:"hero_id" validate:"required"`
	DungeonID string      `json:"dungeon_id" validate:"required"`
	Loot      lootPayload `json:"loot"`
}

type failDungeonRequest struct {
	HeroID    string `json:"hero_id" validate:"required"`
	DungeonID string `json:"dungeon_id" validate:"required"`
	Reason    string `json:"reason,omitempty"`
}

type abandonDungeonRequest struct {
	HeroID    string `json:"hero_id" validate:"required"`
	DungeonID string `json:"dungeon_id" validate:"required"`
	Reason    string `json:"reason,omitempty"`
}

type lootPayload struct {
	Gold  int64             `json:"gold"`
	Items []lootItemPayload `json:"items"`
}

type lootItemPayload struct {
	ItemID   string `json:"item_id" validate:"required"`
	ItemType string `json:"item_type,omitempty"`
	Quantity int    `json:"quantity" validate:"gte=1"`
}

type dungeonProgressResponse struct {
	ID              string   `json:"id"`
	TeamID          string   `json:"team_id"`
	DungeonID       string   `json:"dungeon_id"`
	Status          string   `json:"status"`
	CurrentRoomID   *string  `json:"current_room_id,omitempty"`
	CompletedRooms  []string `json:"completed_rooms"`
	StartedAt       string   `json:"started_at"`
	CompletedAt     *string  `json:"completed_at,omitempty"`
	LastUpdatedTime string   `json:"updated_at"`
}

type dungeonHistoryItem struct {
	DungeonID     string `json:"dungeon_id"`
	AttemptsCount int    `json:"attempts_count"`
	MaxAttempts   *int   `json:"max_attempts,omitempty"`
	UpdatedAt     string `json:"updated_at"`
}

// ==================== Handlers ====================

// SelectDungeon 选择地城
// @Summary 选择地城
// @Description 由队长或管理员为团队选择一个地城
// @Tags 地城
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body selectDungeonRequest true "选择地城请求"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/select [post]
func (h *TeamDungeonHandler) SelectDungeon(c echo.Context) error {
	teamID := c.Param("team_id")
	var req selectDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}
	progress, err := h.dungeonService.SelectDungeon(c.Request().Context(), &service.SelectDungeonRequest{
		TeamID:    teamID,
		HeroID:    req.HeroID,
		DungeonID: req.DungeonID,
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// EnterDungeon 进入地城
// @Summary 进入地城
// @Description 队长/管理员发起进入地城
// @Tags 地城
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body enterDungeonRequest true "进入地城请求"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/enter [post]
func (h *TeamDungeonHandler) EnterDungeon(c echo.Context) error {
	teamID := c.Param("team_id")
	var req enterDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	progress, err := h.dungeonService.EnterDungeon(c.Request().Context(), &service.EnterDungeonRequest{
		TeamID:    teamID,
		HeroID:    req.HeroID,
		DungeonID: req.DungeonID,
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// GetDungeonProgress 查看进度
// @Summary 查看地城进度
// @Tags 地城
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "英雄ID"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/progress [get]
func (h *TeamDungeonHandler) GetDungeonProgress(c echo.Context) error {
	teamID := c.Param("team_id")
	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "hero_id 不能为空")
	}

	progress, err := h.dungeonService.GetDungeonProgress(c.Request().Context(), teamID, heroID)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// CompleteDungeon 完成地城
// @Summary 完成地城
// @Tags 地城
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body completeDungeonRequest true "完成请求"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/complete [post]
func (h *TeamDungeonHandler) CompleteDungeon(c echo.Context) error {
	teamID := c.Param("team_id")
	var req completeDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	progress, err := h.dungeonService.CompleteDungeon(c.Request().Context(), &service.CompleteDungeonRequest{
		TeamID:    teamID,
		HeroID:    req.HeroID,
		DungeonID: req.DungeonID,
		Loot: service.LootData{
			Gold:  req.Loot.Gold,
			Items: convertLootItems(req.Loot.Items),
		},
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// FailDungeon 标记失败
// @Summary 地城失败
// @Tags 地城
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body failDungeonRequest true "失败请求"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/fail [post]
func (h *TeamDungeonHandler) FailDungeon(c echo.Context) error {
	teamID := c.Param("team_id")
	var req failDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	progress, err := h.dungeonService.FailDungeon(c.Request().Context(), &service.FailDungeonRequest{
		TeamID:    teamID,
		HeroID:    req.HeroID,
		DungeonID: req.DungeonID,
		Reason:    req.Reason,
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// AbandonDungeon 放弃地城
// @Summary 放弃地城
// @Tags 地城
// @Accept json
// @Produce json
// @Param team_id path string true "团队ID"
// @Param request body abandonDungeonRequest true "放弃请求"
// @Success 200 {object} response.Response{data=dungeonProgressResponse}
// @Router /game/teams/{team_id}/dungeons/abandon [post]
func (h *TeamDungeonHandler) AbandonDungeon(c echo.Context) error {
	teamID := c.Param("team_id")
	var req abandonDungeonRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	progress, err := h.dungeonService.AbandonDungeon(c.Request().Context(), &service.AbandonDungeonRequest{
		TeamID:    teamID,
		HeroID:    req.HeroID,
		DungeonID: req.DungeonID,
		Reason:    req.Reason,
	})
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

// GetDungeonHistory 历史记录
// @Summary 地城挑战历史
// @Tags 地城
// @Produce json
// @Param team_id path string true "团队ID"
// @Param hero_id query string true "英雄ID"
// @Param limit query int false "数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} response.Response{data=object{list=[]dungeonHistoryItem,total=int64,limit=int,offset=int}}
// @Router /game/teams/{team_id}/dungeons/history [get]
func (h *TeamDungeonHandler) GetDungeonHistory(c echo.Context) error {
	teamID := c.Param("team_id")
	heroID := c.QueryParam("hero_id")
	if heroID == "" {
		return response.EchoBadRequest(c, h.respWriter, "hero_id 不能为空")
	}

	limit, offset := parsePagination(c, 20)
	list, total, err := h.dungeonService.GetChallengeHistory(c.Request().Context(), teamID, heroID, limit, offset)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	items := make([]dungeonHistoryItem, len(list))
	for i, record := range list {
		var maxAttempts *int
		if record.MaxAttempts.Valid {
			value := record.MaxAttempts.Int
			maxAttempts = &value
		}
		items[i] = dungeonHistoryItem{
			DungeonID:     record.DungeonID,
			AttemptsCount: record.AttemptsCount,
			MaxAttempts:   maxAttempts,
			UpdatedAt:     record.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"list":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ==================== 辅助函数 ====================

func toDungeonProgressResponse(progress *game_runtime.TeamDungeonProgress) *dungeonProgressResponse {
	if progress == nil {
		return nil
	}
	var currentRoom *string
	if !progress.CurrentRoomID.IsZero() {
		value := progress.CurrentRoomID.String
		currentRoom = &value
	}

	var completedAt *string
	if !progress.CompletedAt.IsZero() {
		value := progress.CompletedAt.Time.Format(time.RFC3339)
		completedAt = &value
	}

	return &dungeonProgressResponse{
		ID:              progress.ID,
		TeamID:          progress.TeamID,
		DungeonID:       progress.DungeonID,
		Status:          progress.Status,
		CurrentRoomID:   currentRoom,
		CompletedRooms:  decodeCompletedRooms(progress.CompletedRooms),
		StartedAt:       progress.StartedAt.Format(time.RFC3339),
		CompletedAt:     completedAt,
		LastUpdatedTime: progress.UpdatedAt.Format(time.RFC3339),
	}
}

func decodeCompletedRooms(data types.JSON) []string {
	if len(data) == 0 {
		return nil
	}
	var rooms []string
	if err := json.Unmarshal(data, &rooms); err != nil {
		return nil
	}
	return rooms
}

func convertLootItems(items []lootItemPayload) []service.LootItem {
	if len(items) == 0 {
		return nil
	}
	result := make([]service.LootItem, 0, len(items))
	for _, item := range items {
		if item.ItemID == "" || item.Quantity <= 0 {
			continue
		}
		result = append(result, service.LootItem{
			ItemID:   item.ItemID,
			ItemType: item.ItemType,
			Quantity: item.Quantity,
		})
	}
	return result
}

// helper: parse limit/offset with defaults
func parsePagination(c echo.Context, defaultLimit int) (int, int) {
	limit := defaultLimit
	offset := 0
	if v := c.QueryParam("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}
