package middleware

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// HeroMiddleware 英雄上下文中间件 - 自动从数据库获取当前操作英雄ID
// 这个中间件应该在 AuthMiddleware 之后使用，因为需要 user_id
//
// 工作流程：
// 1. 从 context 获取 user_id（由 AuthMiddleware 设置）
// 2. 查询 current_hero_contexts 表获取当前英雄ID
// 3. 将 hero_id 设置到 context 供后续 handler 使用
//
// 使用场景：
// - 需要英雄上下文的 API（如团队操作、战斗、装备等）
// - 不需要应用在创建英雄、获取英雄列表等不依赖当前英雄的 API
func HeroMiddleware(db *sql.DB, respWriter response.Writer, logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// 1. 获取 user_id（由 AuthMiddleware 设置）
			userID := c.Get(string(ctxkey.UserID))
			if userID == nil {
				logger.WarnContext(ctx, "HeroMiddleware: 未找到 user_id，请确保 AuthMiddleware 在之前执行")
				err := xerrors.New(
					xerrors.CodeAuthenticationFailed,
					"未找到用户信息",
				).WithService("middleware", "hero")
				return respWriter.WriteError(ctx, c.Response().Writer, err)
			}

			userIDStr, ok := userID.(string)
			if !ok {
				logger.ErrorContext(ctx, "HeroMiddleware: user_id 类型错误")
				err := xerrors.New(
					xerrors.CodeInternalError,
					"用户信息类型错误",
				).WithService("middleware", "hero")
				return respWriter.WriteError(ctx, c.Response().Writer, err)
			}

			// 2. 查询当前操作英雄ID
			var heroID string
			err := db.QueryRowContext(ctx,
				`SELECT hero_id
				 FROM game_runtime.current_hero_contexts
				 WHERE user_id = $1`,
				userIDStr,
			).Scan(&heroID)

			if err != nil {
				// 严格模式：必须有激活的英雄才能进行游戏操作
				if err == sql.ErrNoRows {
					logger.WarnContext(ctx, "用户未激活英雄", log.String("user_id", userIDStr))
					appErr := xerrors.New(
						xerrors.CodeBusinessLogicError,
						"请先激活一个英雄",
					).WithService("middleware", "hero")
					return respWriter.WriteError(ctx, c.Response().Writer, appErr)
				}

				// 数据库查询错误
				logger.Error("查询当前英雄失败", err, log.String("user_id", userIDStr))
				appErr := xerrors.Wrap(err, xerrors.CodeDatabaseError, "查询当前英雄失败").
					WithService("middleware", "hero")
				return respWriter.WriteError(ctx, c.Response().Writer, appErr)
			}

			// 3. 设置 hero_id 到 context
			ctx = ctxkey.WithValue(ctx, ctxkey.HeroID, heroID)
			c.SetRequest(c.Request().WithContext(ctx))

			// 也设置到 Echo Context，便于直接访问
			c.Set(string(ctxkey.HeroID), heroID)

			logger.DebugContext(ctx,
				"英雄上下文设置成功",
				log.String("user_id", userIDStr),
				log.String("hero_id", heroID),
			)

			return next(c)
		}
	}
}

// GetCurrentHeroID 从 Echo Context 中获取当前英雄 ID（快捷方法）
func GetCurrentHeroID(c echo.Context) (string, error) {
	heroID := c.Get(string(ctxkey.HeroID))
	if heroID == nil {
		return "", xerrors.New(
			xerrors.CodeBusinessLogicError,
			"未找到当前英雄信息",
		)
	}

	heroIDStr, ok := heroID.(string)
	if !ok {
		return "", xerrors.New(
			xerrors.CodeInternalError,
			"英雄信息类型错误",
		)
	}

	return heroIDStr, nil
}
