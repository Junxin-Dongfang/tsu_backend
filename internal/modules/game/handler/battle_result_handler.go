package handler

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/response"
)

// BattleResultHandler 接收战斗引擎的回调。
type BattleResultHandler struct {
	battleService *service.BattleResultService
	respWriter    response.Writer
	token         string
}

// NewBattleResultHandler 构造函数。
func NewBattleResultHandler(sc *service.ServiceContainer, respWriter response.Writer) *BattleResultHandler {
	var token string
	if t := os.Getenv("BATTLE_RESULT_TOKEN"); t != "" {
		token = t
	}
	return &BattleResultHandler{
		battleService: sc.GetBattleResultService(),
		respWriter:    respWriter,
		token:         token,
	}
}

type battleResultPayload struct {
	BattleID     string              `json:"battle_id"`
	BattleCode   string              `json:"battle_code"`
	Participants []battleParticipant `json:"participants"`
	Result       battleResultInfo    `json:"result"`
	Loot         *battleLootPayload  `json:"loot"`
	Events       interface{}         `json:"events"`
}

type battleParticipant struct {
	HeroID string `json:"hero_id"`
	TeamID string `json:"team_id"`
	Role   string `json:"role"`
}

type battleResultInfo struct {
	Status      string      `json:"status"`
	LootContext lootContext `json:"loot_context"`
}

type lootContext struct {
	Type      string `json:"type"`
	TeamID    string `json:"team_id"`
	DungeonID string `json:"dungeon_id"`
}

type battleLootPayload struct {
	Gold  int64            `json:"gold"`
	Items []battleLootItem `json:"items"`
}

type battleLootItem struct {
	ItemID   string `json:"item_id"`
	ItemType string `json:"item_type"`
	Quantity int    `json:"quantity"`
}

// ReportResult 处理战斗结果。
func (h *BattleResultHandler) ReportResult(c echo.Context) error {
	if h.battleService == nil {
		return response.EchoError(c, h.respWriter, echo.NewHTTPError(http.StatusInternalServerError, "battle result service unavailable"))
	}
	if h.token != "" && c.Request().Header.Get("X-Battle-Token") != h.token {
		return response.EchoUnauthorized(c, h.respWriter, "battle token invalid")
	}

	var payload battleResultPayload
	if err := c.Bind(&payload); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "invalid battle payload")
	}
	if payload.BattleID == "" {
		return response.EchoBadRequest(c, h.respWriter, "battle_id is required")
	}

	input := payload.toServiceInput()
	progress, err := h.battleService.RecordAndComplete(c.Request().Context(), input)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	if progress == nil {
		return response.EchoOK(c, h.respWriter, map[string]interface{}{"message": "battle result recorded"})
	}
	return response.EchoOK(c, h.respWriter, toDungeonProgressResponse(progress))
}

func (p *battleResultPayload) firstHeroID() string {
	if p == nil {
		return ""
	}
	for _, participant := range p.Participants {
		if participant.HeroID != "" {
			return participant.HeroID
		}
	}
	return ""
}

func (p *battleResultPayload) toLootData() service.LootData {
	loot := service.LootData{}
	if p == nil || p.Loot == nil {
		return loot
	}
	loot.Gold = p.Loot.Gold
	for _, item := range p.Loot.Items {
		if item.ItemID == "" || item.Quantity <= 0 {
			continue
		}
		loot.Items = append(loot.Items, service.LootItem{ItemID: item.ItemID, ItemType: item.ItemType, Quantity: item.Quantity})
	}
	return loot
}

func (p *battleResultPayload) toServiceInput() *service.BattleResultInput {
	if p == nil {
		return &service.BattleResultInput{}
	}
	return &service.BattleResultInput{
		BattleID:     p.BattleID,
		BattleCode:   p.BattleCode,
		TeamID:       p.Result.LootContext.TeamID,
		DungeonID:    p.Result.LootContext.DungeonID,
		HeroID:       p.firstHeroID(),
		ResultStatus: p.Result.Status,
		Participants: p.Participants,
		Events:       p.Events,
		RawPayload:   p,
		Loot:         p.toLootData(),
	}
}
