package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	"google.golang.org/protobuf/proto"

	"tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// PermissionMiddlewareConfig 权限中间件配置
type PermissionMiddlewareConfig struct {
	// RequiredPermissions 需要的权限代码列表（满足任意一个即可）
	RequiredPermissions []string

	// RequireAllPermissions 是否需要满足所有权限（默认 false，满足任意一个即可）
	RequireAllPermissions bool

	// Skipper 跳过中间件的条件函数（可选）
	Skipper func(c echo.Context) bool
}

// PermissionMiddleware 权限检查中间件 - 集成 Keto
// 使用方式：
//
//	admin.POST("/critical-operation", handler,
//	    custommiddleware.PermissionMiddleware(app, m, respWriter, logger, custommiddleware.PermissionMiddlewareConfig{
//	        RequiredPermissions: []string{"admin:write", "super_admin"},
//	    }))
func PermissionMiddleware(
	app module.App,
	thisModule module.RPCModule,
	respWriter response.Writer,
	logger log.Logger,
	config PermissionMiddlewareConfig,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skipper 检查
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			ctx := c.Request().Context()

			// 1. 获取当前用户（必须先经过 AuthMiddleware）
			currentUser, err := GetCurrentUser(c)
			if err != nil {
				logger.WarnContext(ctx, "权限检查失败: 未找到用户信息")
				return respWriter.WriteError(ctx, c.Response().Writer, xerrors.New(
					xerrors.CodeAuthenticationFailed,
					"未授权访问",
				))
			}

			userID := currentUser.UserID

			// 2. 如果没有配置权限要求，直接放行
			if len(config.RequiredPermissions) == 0 {
				return next(c)
			}

			// 3. 检查用户权限
			if config.RequireAllPermissions {
				// 需要满足所有权限
				for _, permCode := range config.RequiredPermissions {
					allowed, err := checkPermission(ctx, app, thisModule, userID, permCode)
					if err != nil {
						logger.ErrorContext(ctx, "权限检查 RPC 调用失败",
							log.String("user_id", userID),
							log.String("permission", permCode),
							log.Any("error", err),
						)
						return respWriter.WriteError(ctx, c.Response().Writer, xerrors.New(
							xerrors.CodeInternalError,
							"权限检查失败",
						))
					}

					if !allowed {
						logger.WarnContext(ctx, "权限不足",
							log.String("user_id", userID),
							log.String("required_permission", permCode),
						)
						return respWriter.WriteError(ctx, c.Response().Writer, xerrors.New(
							xerrors.CodePermissionDenied,
							"权限不足",
						))
					}
				}
			} else {
				// 只需满足任意一个权限
				hasPermission := false
				for _, permCode := range config.RequiredPermissions {
					allowed, err := checkPermission(ctx, app, thisModule, userID, permCode)
					if err != nil {
						logger.ErrorContext(ctx, "权限检查 RPC 调用失败",
							log.String("user_id", userID),
							log.String("permission", permCode),
							log.Any("error", err),
						)
						continue
					}

					if allowed {
						hasPermission = true
						break
					}
				}

				if !hasPermission {
					logger.WarnContext(ctx, "权限不足",
						log.String("user_id", userID),
						log.Any("required_permissions", config.RequiredPermissions),
					)
					return respWriter.WriteError(ctx, c.Response().Writer, xerrors.New(
						xerrors.CodePermissionDenied,
						"权限不足: 需要以下权限之一: "+joinPermissions(config.RequiredPermissions),
					))
				}
			}

			logger.DebugContext(ctx, "权限检查通过",
				log.String("user_id", userID),
				log.Any("required_permissions", config.RequiredPermissions),
			)

			return next(c)
		}
	}
}

// RequirePermission 快捷方法：需要单个权限
func RequirePermission(
	app module.App,
	thisModule module.RPCModule,
	respWriter response.Writer,
	logger log.Logger,
	permissionCode string,
) echo.MiddlewareFunc {
	return PermissionMiddleware(app, thisModule, respWriter, logger, PermissionMiddlewareConfig{
		RequiredPermissions: []string{permissionCode},
	})
}

// RequireAnyPermission 快捷方法：需要任意一个权限
func RequireAnyPermission(
	app module.App,
	thisModule module.RPCModule,
	respWriter response.Writer,
	logger log.Logger,
	permissionCodes ...string,
) echo.MiddlewareFunc {
	return PermissionMiddleware(app, thisModule, respWriter, logger, PermissionMiddlewareConfig{
		RequiredPermissions: permissionCodes,
	})
}

// RequireAllPermissions 快捷方法：需要所有权限
func RequireAllPermissions(
	app module.App,
	thisModule module.RPCModule,
	respWriter response.Writer,
	logger log.Logger,
	permissionCodes ...string,
) echo.MiddlewareFunc {
	return PermissionMiddleware(app, thisModule, respWriter, logger, PermissionMiddlewareConfig{
		RequiredPermissions:   permissionCodes,
		RequireAllPermissions: true,
	})
}

// checkPermission 调用 Auth 模块的 RPC 检查权限
func checkPermission(ctx context.Context, app module.App, thisModule module.RPCModule, userID, permissionCode string) (bool, error) {
	rpcReq := &auth.CheckUserPermissionRequest{
		UserId:         userID,
		PermissionCode: permissionCode,
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		return false, err
	}

	result, errStr := app.Invoke(thisModule, "auth", "CheckUserPermission", rpcReqBytes)
	if errStr != "" {
		return false, xerrors.New(xerrors.CodeExternalServiceError, errStr)
	}

	rpcResp := &auth.CheckUserPermissionResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		return false, xerrors.New(xerrors.CodeInternalError, "RPC 响应类型错误")
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		return false, err
	}

	return rpcResp.Allowed, nil
}

// joinPermissions 辅助函数：拼接权限列表
func joinPermissions(perms []string) string {
	if len(perms) == 0 {
		return ""
	}
	result := perms[0]
	for i := 1; i < len(perms); i++ {
		result += ", " + perms[i]
	}
	return result
}
