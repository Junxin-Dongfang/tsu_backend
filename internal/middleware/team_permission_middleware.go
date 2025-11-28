// File: internal/middleware/team_permission_middleware.go
package middleware

import (
	"context"
	"fmt"

	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"

	"github.com/labstack/echo/v4"
)

type teamPermissionChecker interface {
	CheckPermissionWithCache(ctx context.Context, teamID, heroID, permission string) (bool, error)
	GetMemberRole(ctx context.Context, teamID, heroID string) (string, bool, error)
}

// TeamPermissionMiddleware 团队权限中间件
type TeamPermissionMiddleware struct {
	permissionChecker teamPermissionChecker
	respWriter        response.Writer
}

// NewTeamPermissionMiddleware 创建团队权限中间件
func NewTeamPermissionMiddleware(permissionChecker teamPermissionChecker, respWriter response.Writer) *TeamPermissionMiddleware {
	return &TeamPermissionMiddleware{
		permissionChecker: permissionChecker,
		respWriter:        respWriter,
	}
}

// RequireTeamLeader 要求队长权限
func (m *TeamPermissionMiddleware) RequireTeamLeader(next echo.HandlerFunc) echo.HandlerFunc {
	return m.requireTeamRole("leader", next)
}

// RequireTeamAdmin 要求管理员或队长权限
func (m *TeamPermissionMiddleware) RequireTeamAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := m.ensureChecker(); err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		teamID, err := m.getTeamIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		heroID, err := m.getHeroIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		// 检查是否是 leader 或 admin
		role, exists, err := m.permissionChecker.GetMemberRole(c.Request().Context(), teamID, heroID)
		if err != nil {
			return response.EchoError(c, m.respWriter,
				xerrors.Wrap(err, xerrors.CodeInternalError, "权限检查失败"))
		}

		if !exists {
			return response.EchoError(c, m.respWriter,
				xerrors.New(xerrors.CodePermissionDenied, "您不是该团队成员"))
		}

		if role != "leader" && role != "admin" {
			return response.EchoError(c, m.respWriter,
				xerrors.New(xerrors.CodePermissionDenied, "需要管理员或队长权限"))
		}

		// 将角色存入 context 供后续使用
		c.Set("team_role", role)
		return next(c)
	}
}

// RequireTeamMember 要求是团队成员
func (m *TeamPermissionMiddleware) RequireTeamMember(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := m.ensureChecker(); err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		teamID, err := m.getTeamIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		heroID, err := m.getHeroIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		// 检查是否是团队成员
		role, exists, err := m.permissionChecker.GetMemberRole(c.Request().Context(), teamID, heroID)
		if err != nil {
			return response.EchoError(c, m.respWriter,
				xerrors.Wrap(err, xerrors.CodeInternalError, "权限检查失败"))
		}

		if !exists {
			return response.EchoError(c, m.respWriter,
				xerrors.New(xerrors.CodePermissionDenied, "您不是该团队成员"))
		}

		// 将角色存入 context 供后续使用
		c.Set("team_role", role)
		return next(c)
	}
}

// RequireTeamPermission 要求特定团队权限
// permission: 权限名称 (如: "select_dungeon", "kick_member", "distribute_loot")
func (m *TeamPermissionMiddleware) RequireTeamPermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := m.ensureChecker(); err != nil {
				return response.EchoError(c, m.respWriter, err)
			}

			teamID, err := m.getTeamIDFromContext(c)
			if err != nil {
				return response.EchoError(c, m.respWriter, err)
			}

			heroID, err := m.getHeroIDFromContext(c)
			if err != nil {
				return response.EchoError(c, m.respWriter, err)
			}

			// 检查权限（带缓存）
			allowed, err := m.permissionChecker.CheckPermissionWithCache(
				c.Request().Context(), teamID, heroID, permission,
			)
			if err != nil {
				return response.EchoError(c, m.respWriter,
					xerrors.Wrap(err, xerrors.CodeInternalError, "权限检查失败"))
			}

			if !allowed {
				return response.EchoError(c, m.respWriter,
					xerrors.New(xerrors.CodePermissionDenied, fmt.Sprintf("缺少 %s 权限", permission)))
			}

			return next(c)
		}
	}
}

// requireTeamRole 内部方法: 要求特定角色
func (m *TeamPermissionMiddleware) requireTeamRole(role string, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := m.ensureChecker(); err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		teamID, err := m.getTeamIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		heroID, err := m.getHeroIDFromContext(c)
		if err != nil {
			return response.EchoError(c, m.respWriter, err)
		}

		// 检查角色
		memberRole, exists, err := m.permissionChecker.GetMemberRole(c.Request().Context(), teamID, heroID)
		if err != nil {
			return response.EchoError(c, m.respWriter,
				xerrors.Wrap(err, xerrors.CodeInternalError, "权限检查失败"))
		}

		if !exists {
			return response.EchoError(c, m.respWriter,
				xerrors.New(xerrors.CodePermissionDenied, "您不是该团队成员"))
		}

		if memberRole != role {
			return response.EchoError(c, m.respWriter,
				xerrors.New(xerrors.CodePermissionDenied, fmt.Sprintf("需要 %s 权限", role)))
		}

		// 将角色存入 context 供后续使用
		c.Set("team_role", memberRole)
		return next(c)
	}
}

// getHeroIDFromContext 从 context 中获取 heroID
// 优先级：context (由 HeroMiddleware 设置) > 查询参数 > 表单值
func (m *TeamPermissionMiddleware) getHeroIDFromContext(c echo.Context) (string, error) {
	// 优先从 context 中获取（由 HeroMiddleware 设置）
	heroID := c.Get(string(ctxkey.HeroID))
	if heroID == nil {
		// Fallback: 从请求参数中获取（向后兼容）
		heroIDStr := c.QueryParam("hero_id")
		if heroIDStr == "" {
			heroIDStr = c.FormValue("hero_id")
		}
		if heroIDStr == "" {
			return "", xerrors.New(xerrors.CodeBusinessLogicError, "未找到英雄ID，请先激活一个英雄")
		}
		return heroIDStr, nil
	}

	heroIDStr, ok := heroID.(string)
	if !ok {
		return "", xerrors.New(xerrors.CodeInternalError, "英雄ID格式错误")
	}

	return heroIDStr, nil
}

func (m *TeamPermissionMiddleware) ensureChecker() error {
	if m.permissionChecker == nil {
		return xerrors.New(xerrors.CodeInternalError, "团队权限服务未初始化")
	}
	return nil
}

func (m *TeamPermissionMiddleware) getTeamIDFromContext(c echo.Context) (string, error) {
	teamID := c.Param("team_id")
	if teamID == "" {
		teamID = c.QueryParam("team_id")
	}
	if teamID == "" {
		teamID = c.FormValue("team_id")
	}
	if teamID == "" {
		return "", xerrors.New(xerrors.CodeAuthenticationFailed, "团队ID不能为空")
	}
	return teamID, nil
}

// InitializeTeamPermissions 初始化团队权限定义
// 这个函数应该在系统启动时调用一次
func InitializeTeamPermissions(ctx context.Context, ketoClient *client.KetoClient) error {
	return ketoClient.InitializeTeamPermissions(ctx)
}
