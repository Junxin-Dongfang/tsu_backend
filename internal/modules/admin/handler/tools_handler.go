package handler

import (
	"os"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/modules/admin/dto"
	"tsu-self/internal/modules/admin/service"
	"tsu-self/internal/pkg/response"
)

// ToolsHandler 管理/测试工具接口（仅受控环境启用）
type ToolsHandler struct {
	service    *service.ToolsService
	respWriter response.Writer
	enabled    bool
}

func NewToolsHandler(s *service.ToolsService, respWriter response.Writer) *ToolsHandler {
	// 默认开启，只有显式设置 ENABLE_TEST_TOOLS=false 时关闭
	enable := os.Getenv("ENABLE_TEST_TOOLS") != "false"
	return &ToolsHandler{service: s, respWriter: respWriter, enabled: enable}
}

// GrantItem 发放物品到背包或团队仓库
// @Summary 管理员发放物品（测试工具）
// @Tags Tools
// @Accept json
// @Produce json
// @Param request body dto.GrantItemRequest true "发放请求"
// @Success 200 {object} response.Response{data=dto.GrantItemResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/tools/grant-item [post]
func (h *ToolsHandler) GrantItem(c echo.Context) error {
	if !h.enabled {
		return response.EchoForbidden(c, h.respWriter, "tools", "grant_item")
	}

	var req dto.GrantItemRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	respData, err := h.service.GrantItem(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, respData)
}

// GrantGold 向仓库添加金币
// @Summary 管理员加金币到团队仓库（测试工具）
// @Tags Tools
// @Accept json
// @Produce json
// @Param request body dto.GrantGoldRequest true "加金币请求"
// @Success 200 {object} response.Response{data=dto.GrantGoldResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/tools/grant-gold [post]
func (h *ToolsHandler) GrantGold(c echo.Context) error {
	if !h.enabled {
		return response.EchoForbidden(c, h.respWriter, "tools", "grant_gold")
	}

	var req dto.GrantGoldRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	respData, err := h.service.GrantGold(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, respData)
}

// GrantExperience 给英雄加经验
// @Summary 管理员给英雄加经验（测试工具）
// @Tags Tools
// @Accept json
// @Produce json
// @Param request body dto.GrantExperienceRequest true "加经验请求"
// @Success 200 {object} response.Response{data=dto.GrantExperienceResponse}
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/tools/grant-experience [post]
func (h *ToolsHandler) GrantExperience(c echo.Context) error {
	if !h.enabled {
		return response.EchoForbidden(c, h.respWriter, "tools", "grant_experience")
	}

	var req dto.GrantExperienceRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	respData, err := h.service.GrantExperience(c.Request().Context(), &req)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}
	return response.EchoOK(c, h.respWriter, respData)
}
