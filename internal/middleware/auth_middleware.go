// internal/middleware/auth_middleware.go - 修正版本
package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	mqrpc "github.com/liangdas/mqant/rpc"
	"github.com/redis/go-redis/v9"

	"tsu-self/internal/pkg/contextkeys"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
	authpb "tsu-self/internal/rpc/generated/auth"
	commonpb "tsu-self/internal/rpc/generated/common"
)

type AuthMiddleware struct {
	app         module.App // mqant app用于RPC调用
	redis       *redis.Client
	logger      log.Logger
	respHandler response.Writer

	// 权限缓存
	permissionCache map[string]*CachedPermissions
}

type CachedPermissions struct {
	Permissions []string  `json:"permissions"`
	CachedAt    time.Time `json:"cached_at"`
}

func NewAuthMiddleware(app module.App, redis *redis.Client, logger log.Logger) *AuthMiddleware {
	middleware := &AuthMiddleware{
		app:             app,
		redis:           redis,
		logger:          logger,
		respHandler:     response.DefaultResponseHandler(),
		permissionCache: make(map[string]*CachedPermissions),
	}

	return middleware
}

// RequireAuth Token验证中间件
func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := m.extractToken(c)
			if token == "" {
				return m.respHandler.WriteError(c.Request().Context(), c.Response().Writer,
					xerrors.FromCode(xerrors.CodeInvalidToken).WithMetadata("reason", "missing_token"))
			}

			// 调用 auth module 验证 token
			validateReq := &authpb.ValidateTokenRequest{Token: token}

			result, err := m.app.Call(context.Background(), "auth", "ValidateToken", mqrpc.Param(validateReq))
			if err != "" {
				m.logger.ErrorContext(c.Request().Context(), "Token验证失败", log.Any("error", err))
				return m.respHandler.WriteError(c.Request().Context(), c.Response().Writer,
					xerrors.FromCode(xerrors.CodeInvalidToken))
			}

			validateResp, ok := result.(*authpb.ValidateTokenResponse)
			if !ok || !validateResp.Valid {
				return m.respHandler.WriteError(c.Request().Context(), c.Response().Writer,
					xerrors.FromCode(xerrors.CodeInvalidToken))
			}

			// 将用户信息添加到context
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, contextkeys.UserIDKey, validateResp.UserId)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// RequirePermission 权限检查中间件
func (m *AuthMiddleware) RequirePermission(resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Request().Context().Value(contextkeys.UserIDKey).(string)
			if !ok {
				return m.respHandler.WriteError(c.Request().Context(), c.Response().Writer,
					xerrors.FromCode(xerrors.CodeAuthenticationFailed).WithMetadata("reason", "no_user_context"))
			}

			// 检查权限
			if !m.checkPermission(c.Request().Context(), userID, resource, action) {
				return m.respHandler.WriteError(c.Request().Context(), c.Response().Writer,
					xerrors.FromCode(xerrors.CodePermissionDenied).WithMetadata("resource", resource).WithMetadata("action", action))
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) checkPermission(ctx context.Context, userID, resource, action string) bool {
	permission := resource + ":" + action

	// 1. 检查缓存
	if cached, exists := m.permissionCache[userID]; exists {
		if time.Since(cached.CachedAt) < 5*time.Minute { // 5分钟缓存
			for _, perm := range cached.Permissions {
				if perm == permission {
					return true
				}
			}
			return false
		}
	}

	// 2. 调用 auth module 检查权限
	checkReq := &commonpb.CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	}

	result, err := m.app.Call(context.Background(), "auth", "CheckPermission", mqrpc.Param(checkReq))
	if err != "" {
		m.logger.ErrorContext(ctx, "权限检查调用失败", log.Any("error", err))
		return false
	}

	checkResp, ok := result.(*commonpb.CheckPermissionResponse)
	if !ok {
		m.logger.ErrorContext(ctx, "权限检查响应类型错误")
		return false
	}

	return checkResp.Allowed
}

// ClearUserPermissionCache 清理指定用户的权限缓存 - 通过RPC调用
func (m *AuthMiddleware) ClearUserPermissionCache(userID string) {
	delete(m.permissionCache, userID)
	m.logger.InfoContext(context.Background(), "清理用户权限缓存",
		log.String("user_id", userID))
}

func (m *AuthMiddleware) extractToken(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	const bearer = "Bearer "
	if !strings.HasPrefix(auth, bearer) {
		return ""
	}

	return auth[len(bearer):]
}
