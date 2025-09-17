// File: internal/app/admin/controller/user_controller.go
package controller

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"tsu-self/internal/app/admin/service"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// UserController 用户管理控制器
type UserController struct {
	userService *service.UserService
	respWriter  response.Writer
	logger      log.Logger
}

// NewUserController 创建用户控制器
func NewUserController(userService *service.UserService, respWriter response.Writer, logger log.Logger) *UserController {
	return &UserController{
		userService: userService,
		respWriter:  respWriter,
		logger:      logger,
	}
}

// ListUsers 列出用户（简化版，映射到身份）
func (uc *UserController) ListUsers(c echo.Context) error {
	return uc.ListIdentities(c)
}

// GetUser 获取用户（简化版，映射到身份）
func (uc *UserController) GetUser(c echo.Context) error {
	return uc.GetIdentity(c)
}

// UpdateUser 更新用户（简化版，映射到身份）
func (uc *UserController) UpdateUser(c echo.Context) error {
	return uc.UpdateIdentity(c)
}

// DeleteUser 删除用户（简化版，映射到身份）
func (uc *UserController) DeleteUser(c echo.Context) error {
	return uc.DeleteIdentity(c)
}

// DisableUser 禁用用户
func (uc *UserController) DisableUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		appErr := xerrors.NewValidationError("id", "用户ID不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	identity, appErr := uc.userService.DisableIdentity(ctx, id)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	respData := map[string]interface{}{
		"message":  "用户已禁用",
		"identity": identity,
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// EnableUser 启用用户
func (uc *UserController) EnableUser(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		appErr := xerrors.NewValidationError("id", "用户ID不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	identity, appErr := uc.userService.EnableIdentity(ctx, id)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	respData := map[string]interface{}{
		"message":  "用户已启用",
		"identity": identity,
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// ListIdentities 列出身份
func (uc *UserController) ListIdentities(c echo.Context) error {
	ctx := c.Request().Context()

	// 解析查询参数
	query := &service.ListIdentitiesQuery{
		Page:    1,
		PerPage: 20,
	}

	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.ParseInt(pageStr, 10, 64); err == nil && page > 0 {
			query.Page = page
		}
	}

	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if perPage, err := strconv.ParseInt(perPageStr, 10, 64); err == nil && perPage > 0 && perPage <= 1000 {
			query.PerPage = perPage
		}
	}

	query.Ids = c.QueryParam("ids")

	// 获取身份列表
	identities, total, appErr := uc.userService.ListIdentities(ctx, query)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	respData := map[string]interface{}{
		"identities": identities,
		"pagination": map[string]interface{}{
			"page":     query.Page,
			"per_page": query.PerPage,
			"total":    total,
			"count":    len(identities),
		},
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// GetIdentity 获取身份
func (uc *UserController) GetIdentity(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	identity, appErr := uc.userService.GetIdentity(ctx, id)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, identity)
}

// CreateIdentity 创建身份
func (uc *UserController) CreateIdentity(c echo.Context) error {
	ctx := c.Request().Context()

	var req service.CreateIdentityRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	// 基础验证
	if req.Email == "" {
		appErr := xerrors.NewValidationError("email", "邮箱不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	if req.Username == "" {
		appErr := xerrors.NewValidationError("username", "用户名不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	identity, appErr := uc.userService.CreateIdentity(ctx, &req)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	uc.logger.InfoContext(ctx, "创建身份成功",
		log.String("identity_id", identity.Id),
		log.String("email", req.Email),
	)

	respData := map[string]interface{}{
		"message":  "身份创建成功",
		"identity": identity,
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// UpdateIdentity 更新身份
func (uc *UserController) UpdateIdentity(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	var req service.UpdateIdentityRequest
	if err := c.Bind(&req); err != nil {
		appErr := xerrors.NewValidationError("request", "请求格式错误")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	identity, appErr := uc.userService.UpdateIdentity(ctx, id, &req)
	if appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	uc.logger.InfoContext(ctx, "更新身份成功",
		log.String("identity_id", id),
	)

	respData := map[string]interface{}{
		"message":  "身份更新成功",
		"identity": identity,
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}

// DeleteIdentity 删除身份
func (uc *UserController) DeleteIdentity(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	if id == "" {
		appErr := xerrors.NewValidationError("id", "身份ID不能为空")
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	if appErr := uc.userService.DeleteIdentity(ctx, id); appErr != nil {
		return uc.respWriter.WriteError(ctx, c.Response().Writer, appErr)
	}

	uc.logger.InfoContext(ctx, "删除身份成功",
		log.String("identity_id", id),
	)

	respData := map[string]string{
		"message": "身份删除成功",
	}

	return uc.respWriter.WriteSuccess(ctx, c.Response().Writer, respData)
}
