package middleware

import (
	"github.com/labstack/echo/v4"

	"tsu-self/internal/pkg/ctxkey"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// CurrentUser 当前请求的用户信息（从 Oathkeeper 传递）
type CurrentUser struct {
	UserID       string // Kratos Identity ID (从 X-User-ID header)
	SessionToken string // Kratos Session Token (从 X-Session-Token header)
}

// AuthMiddleware 认证中间件 - 从 Oathkeeper 传递的 Header 提取用户信息
// 这个中间件假设请求已经通过 Oathkeeper 验证，只需从 Header 提取用户信息
func AuthMiddleware(respWriter response.Writer, logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// 从 Oathkeeper 传递的 Header 中提取用户信息
			userID := c.Request().Header.Get("X-User-ID")
			sessionToken := c.Request().Header.Get("X-Session-Token")

			// 验证必要信息是否存在
			if userID == "" {
				logger.WarnContext(ctx, "认证失败: 缺少 X-User-ID header")
				err := xerrors.New(
					xerrors.CodeAuthenticationFailed,
					"未授权访问: 缺少用户身份信息",
				).WithService("middleware", "auth")

				return respWriter.WriteError(ctx, c.Response().Writer, err)
			}

			// 构建当前用户对象
			currentUser := &CurrentUser{
				UserID:       userID,
				SessionToken: sessionToken,
			}

			// 将用户信息注入到 Context（使用统一的 ctxkey）
			ctx = ctxkey.WithValue(ctx, ctxkey.UserID, userID)
			ctx = ctxkey.WithValue(ctx, ctxkey.SessionID, sessionToken)
			c.SetRequest(c.Request().WithContext(ctx))

			// 也可以设置到 Echo Context，便于直接访问
			c.Set(string(ctxkey.CurrentUser), currentUser)
			// 设置 user_id 以供 handler 使用
			c.Set(string(ctxkey.UserID), userID)

			logger.DebugContext(ctx,
				"用户认证成功",
				log.String("user_id", userID),
				log.Bool("has_session_token", sessionToken != ""),
			)

			return next(c)
		}
	}
}

// GetCurrentUser 从 Echo Context 中获取当前用户
func GetCurrentUser(c echo.Context) (*CurrentUser, error) {
	user := c.Get(string(ctxkey.CurrentUser))
	if user == nil {
		return nil, xerrors.New(
			xerrors.CodeAuthenticationFailed,
			"未找到用户信息",
		)
	}

	currentUser, ok := user.(*CurrentUser)
	if !ok {
		return nil, xerrors.New(
			xerrors.CodeInternalError,
			"用户信息类型错误",
		)
	}

	return currentUser, nil
}

// GetCurrentUserID 从 Echo Context 中获取当前用户 ID（快捷方法）
func GetCurrentUserID(c echo.Context) (string, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return user.UserID, nil
}

// MustGetCurrentUser 获取当前用户，如果不存在则 panic（用于明确需要认证的地方）
func MustGetCurrentUser(c echo.Context) *CurrentUser {
	user, err := GetCurrentUser(c)
	if err != nil {
		panic(err)
	}
	return user
}
