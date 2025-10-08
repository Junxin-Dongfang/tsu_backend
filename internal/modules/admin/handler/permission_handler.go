package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	"google.golang.org/protobuf/proto"

	authpb "tsu-self/internal/pb/auth"
	commonpb "tsu-self/internal/pb/common"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// PermissionHandler 权限管理 HTTP 处理器
type PermissionHandler struct {
	app        module.App
	thisModule module.RPCModule
	respWriter response.Writer
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(app module.App, thisModule module.RPCModule, respWriter response.Writer) *PermissionHandler {
	return &PermissionHandler{
		app:        app,
		thisModule: thisModule,
		respWriter: respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// RoleRequest 角色请求
type RoleRequest struct {
	Code        string `json:"code" validate:"required,min=2"`
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	IsDefault   bool   `json:"is_default"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	IsDefault   bool   `json:"is_default"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	IsSystem    bool   `json:"is_system"`
	CreatedAt   int64  `json:"created_at"`
}

// PermissionGroupResponse 权限分组响应
type PermissionGroupResponse struct {
	ID          string               `json:"id"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Icon        string               `json:"icon"`
	Color       string               `json:"color"`
	SortOrder   int                  `json:"sort_order"`
	Level       int                  `json:"level"`
	Permissions []PermissionResponse `json:"permissions"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1"`
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleCodes []string `json:"role_codes" validate:"required,min=1"`
}

// GrantPermissionsRequest 授予权限请求
type GrantPermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes" validate:"required,min=1"`
}

// PaginatedRolesResponse 分页角色响应
type PaginatedRolesResponse struct {
	Roles      []RoleResponse         `json:"roles"`
	Pagination PaginationMetaResponse `json:"pagination"`
}

// PaginatedPermissionsResponse 分页权限响应
type PaginatedPermissionsResponse struct {
	Permissions []PermissionResponse   `json:"permissions"`
	Pagination  PaginationMetaResponse `json:"pagination"`
}

// PaginatedPermissionGroupsResponse 分页权限分组响应
type PaginatedPermissionGroupsResponse struct {
	Groups     []PermissionGroupResponse `json:"groups"`
	Pagination PaginationMetaResponse    `json:"pagination"`
}

// PaginationMetaResponse 分页元数据
type PaginationMetaResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ==================== 角色管理 HTTP Handlers ====================

// GetRoles 获取角色列表
// @Summary 获取角色列表
// @Description 获取角色列表,支持分页和搜索
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.Response{data=PaginatedRolesResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles [get]
// @Security BearerAuth
func (h *PermissionHandler) GetRoles(c echo.Context) error {
	// 1. 解析查询参数
	keyword := c.QueryParam("keyword")
	page := parseIntParam(c.QueryParam("page"), 1)
	pageSize := parseIntParam(c.QueryParam("page_size"), 20)

	// 2. 构造 RPC 请求
	rpcReq := &authpb.GetRolesRequest{
		Keyword: keyword,
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetRoles", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析 RPC 响应
	var resp authpb.GetRolesResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 转换为 HTTP 响应
	httpResp := PaginatedRolesResponse{
		Roles:      convertRolesToHTTP(resp.Roles),
		Pagination: convertPaginationToHTTP(resp.Pagination),
	}

	return response.EchoOK(c, h.respWriter, httpResp)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新的角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param request body RoleRequest true "角色请求"
// @Success 200 {object} response.Response{data=RoleResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles [post]
// @Security BearerAuth
func (h *PermissionHandler) CreateRole(c echo.Context) error {
	// 1. 绑定和验证请求
	var req RoleRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.CreateRoleRequest{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    req.IsSystem,
		IsDefault:   req.IsDefault,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "CreateRole", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.CreateRoleResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回 HTTP 响应
	httpResp := convertRoleToHTTP(resp.Role)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    httpResp,
	})
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Param request body UpdateRoleRequest true "更新请求"
// @Success 200 {object} response.Response{data=RoleResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles/{id} [put]
// @Security BearerAuth
func (h *PermissionHandler) UpdateRole(c echo.Context) error {
	roleID := c.Param("id")
	if roleID == "" {
		return response.EchoBadRequest(c, h.respWriter, "角色ID不能为空")
	}

	// 1. 绑定和验证请求
	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.UpdateRoleRequest{
		RoleId:      roleID,
		Name:        req.Name,
		Description: req.Description,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "UpdateRole", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.UpdateRoleResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回 HTTP 响应
	httpResp := convertRoleToHTTP(resp.Role)
	return response.EchoOK(c, h.respWriter, httpResp)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response{data=map[string]interface{}} "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles/{id} [delete]
// @Security BearerAuth
func (h *PermissionHandler) DeleteRole(c echo.Context) error {
	roleID := c.Param("id")
	if roleID == "" {
		return response.EchoBadRequest(c, h.respWriter, "角色ID不能为空")
	}

	// 1. 构造 RPC 请求
	rpcReq := &authpb.DeleteRoleRequest{
		RoleId: roleID,
	}

	// 2. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "DeleteRole", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 解析响应
	var resp authpb.DeleteRoleResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// ==================== 权限管理 HTTP Handlers ====================

// GetPermissions 获取权限列表
// @Summary 获取权限列表
// @Description 获取权限列表,支持分页、搜索和筛选
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param resource query string false "资源筛选"
// @Param action query string false "操作筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.Response{data=PaginatedPermissionsResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/permissions [get]
// @Security BearerAuth
func (h *PermissionHandler) GetPermissions(c echo.Context) error {
	// 1. 解析查询参数
	keyword := c.QueryParam("keyword")
	resource := c.QueryParam("resource")
	action := c.QueryParam("action")
	page := parseIntParam(c.QueryParam("page"), 1)
	pageSize := parseIntParam(c.QueryParam("page_size"), 20)

	// 2. 构造 RPC 请求
	rpcReq := &authpb.GetPermissionsRequest{
		Keyword:  keyword,
		Resource: resource,
		Action:   action,
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetPermissions", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.GetPermissionsResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 转换为 HTTP 响应
	httpResp := PaginatedPermissionsResponse{
		Permissions: convertPermissionsToHTTP(resp.Permissions),
		Pagination:  convertPaginationToHTTP(resp.Pagination),
	}

	return response.EchoOK(c, h.respWriter, httpResp)
}

// GetPermissionGroups 获取权限分组列表
// @Summary 获取权限分组列表
// @Description 获取权限分组列表,支持分页和搜索
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} response.Response{data=PaginatedPermissionGroupsResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/permission-groups [get]
// @Security BearerAuth
func (h *PermissionHandler) GetPermissionGroups(c echo.Context) error {
	// 1. 解析查询参数
	keyword := c.QueryParam("keyword")
	page := parseIntParam(c.QueryParam("page"), 1)
	pageSize := parseIntParam(c.QueryParam("page_size"), 20)

	// 2. 构造 RPC 请求
	rpcReq := &authpb.GetPermissionGroupsRequest{
		Keyword: keyword,
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetPermissionGroups", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.GetPermissionGroupsResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 转换为 HTTP 响应
	httpResp := PaginatedPermissionGroupsResponse{
		Groups:     convertPermissionGroupsToHTTP(resp.Groups),
		Pagination: convertPaginationToHTTP(resp.Pagination),
	}

	return response.EchoOK(c, h.respWriter, httpResp)
}

// ==================== 角色-权限管理 HTTP Handlers ====================

// GetRolePermissions 获取角色的权限列表
// @Summary 获取角色权限
// @Description 获取分配给角色的所有权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response{data=[]PermissionResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles/{id}/permissions [get]
// @Security BearerAuth
func (h *PermissionHandler) GetRolePermissions(c echo.Context) error {
	roleID := c.Param("id")
	if roleID == "" {
		return response.EchoBadRequest(c, h.respWriter, "角色ID不能为空")
	}

	// 1. 构造 RPC 请求
	rpcReq := &authpb.GetRolePermissionsRequest{
		RoleId: roleID,
	}

	// 2. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetRolePermissions", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 解析响应
	var resp authpb.GetRolePermissionsResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 转换为 HTTP 响应
	httpResp := convertPermissionsToHTTP(resp.Permissions)
	return response.EchoOK(c, h.respWriter, httpResp)
}

// AssignPermissionsToRole 为角色分配权限
// @Summary 为角色分配权限
// @Description 为指定角色分配权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path string true "角色ID"
// @Param request body AssignPermissionsRequest true "分配请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "角色不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/roles/{id}/permissions [post]
// @Security BearerAuth
func (h *PermissionHandler) AssignPermissionsToRole(c echo.Context) error {
	roleID := c.Param("id")
	if roleID == "" {
		return response.EchoBadRequest(c, h.respWriter, "角色ID不能为空")
	}

	// 1. 绑定和验证请求
	var req AssignPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 获取操作人 ID (从请求头)
	operatorID := c.Request().Header.Get("X-User-ID")
	if operatorID == "" {
		operatorID = "system" // 默认系统操作
	}

	// 3. 构造 RPC 请求
	rpcReq := &authpb.AssignPermissionsToRoleRequest{
		RoleId:        roleID,
		PermissionIds: req.PermissionIDs,
		OperatorId:    operatorID,
	}

	// 4. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "AssignPermissionsToRole", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 5. 解析响应
	var resp authpb.AssignPermissionsToRoleResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 6. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// ==================== 用户-角色管理 HTTP Handlers ====================

// GetUserRoles 获取用户的角色列表
// @Summary 获取用户角色
// @Description 获取分配给用户的所有角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response{data=[]RoleResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/roles [get]
// @Security BearerAuth
func (h *PermissionHandler) GetUserRoles(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 构造 RPC 请求
	rpcReq := &authpb.GetUserRolesRequest{
		UserId: userID,
	}

	// 2. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetUserRoles", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 解析响应
	var resp authpb.GetUserRolesResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 转换为 HTTP 响应
	httpResp := convertRolesToHTTP(resp.Roles)
	return response.EchoOK(c, h.respWriter, httpResp)
}

// AssignRolesToUser 为用户分配角色
// @Summary 为用户分配角色
// @Description 为指定用户分配角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body AssignRolesRequest true "分配请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "分配成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/roles [post]
// @Security BearerAuth
func (h *PermissionHandler) AssignRolesToUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 绑定和验证请求
	var req AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.AssignRolesToUserRequest{
		UserId:    userID,
		RoleCodes: req.RoleCodes,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "AssignRolesToUser", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.AssignRolesToUserResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// RevokeRolesFromUser 撤销用户的角色
// @Summary 撤销用户角色
// @Description 撤销用户的指定角色
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body AssignRolesRequest true "撤销请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "撤销成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/roles [delete]
// @Security BearerAuth
func (h *PermissionHandler) RevokeRolesFromUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 绑定和验证请求
	var req AssignRolesRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.RevokeRolesFromUserRequest{
		UserId:    userID,
		RoleCodes: req.RoleCodes,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "RevokeRolesFromUser", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.RevokeRolesFromUserResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// ==================== 用户-权限管理 HTTP Handlers ====================

// GetUserPermissions 获取用户的所有权限
// @Summary 获取用户权限
// @Description 获取用户的所有权限(包括角色权限和直接权限)
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} response.Response{data=[]PermissionResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/permissions [get]
// @Security BearerAuth
func (h *PermissionHandler) GetUserPermissions(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 构造 RPC 请求
	rpcReq := &authpb.GetUserPermissionsRequest{
		UserId: userID,
	}

	// 2. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GetUserPermissions", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 3. 解析响应
	var resp authpb.GetUserPermissionsResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 转换为 HTTP 响应
	httpResp := convertPermissionsToHTTP(resp.Permissions)
	return response.EchoOK(c, h.respWriter, httpResp)
}

// GrantPermissionsToUser 直接授予用户权限
// @Summary 授予用户权限
// @Description 直接授予用户权限(绕过角色)
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body GrantPermissionsRequest true "授予请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "授予成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/permissions [post]
// @Security BearerAuth
func (h *PermissionHandler) GrantPermissionsToUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 绑定和验证请求
	var req GrantPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.GrantPermissionsToUserRequest{
		UserId:          userID,
		PermissionCodes: req.PermissionCodes,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "GrantPermissionsToUser", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.GrantPermissionsToUserResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// RevokePermissionsFromUser 撤销用户的直接权限
// @Summary 撤销用户权限
// @Description 撤销用户的直接权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body GrantPermissionsRequest true "撤销请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "撤销成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/users/{user_id}/permissions [delete]
// @Security BearerAuth
func (h *PermissionHandler) RevokePermissionsFromUser(c echo.Context) error {
	userID := c.Param("user_id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 1. 绑定和验证请求
	var req GrantPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.RevokePermissionsFromUserRequest{
		UserId:          userID,
		PermissionCodes: req.PermissionCodes,
	}

	// 3. 调用 Auth RPC
	rpcResp, err := h.callAuthRPC(c.Request().Context(), "RevokePermissionsFromUser", rpcReq)
	if err != nil {
		return response.EchoError(c, h.respWriter, err)
	}

	// 4. 解析响应
	var resp authpb.RevokePermissionsFromUserResponse
	if err := proto.Unmarshal(rpcResp, &resp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 返回成功
	return response.EchoOK(c, h.respWriter, map[string]interface{}{
		"message": resp.Status.Message,
	})
}

// ==================== 内部辅助方法 ====================

// callAuthRPC 调用 Auth 模块 RPC
func (h *PermissionHandler) callAuthRPC(_ context.Context, method string, req proto.Message) ([]byte, error) {
	// 1. 序列化请求
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
	}

	// 2. 调用 RPC (使用 Invoke 代替 RpcInvoke)
	result, errStr := h.app.Invoke(h.thisModule, "auth", method, reqBytes)
	if errStr != "" {
		return nil, xerrors.New(xerrors.CodeExternalServiceError, errStr)
	}

	// 3. 类型断言
	respBytes, ok := result.([]byte)
	if !ok {
		return nil, xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
	}

	return respBytes, nil
}

// parseIntParam 解析整数参数
func parseIntParam(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil || result <= 0 {
		return defaultValue
	}
	return result
}

// ==================== Protobuf 转换方法 ====================

func convertRoleToHTTP(role *authpb.RoleInfo) RoleResponse {
	if role == nil {
		return RoleResponse{}
	}
	return RoleResponse{
		ID:          role.Id,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		IsDefault:   role.IsDefault,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func convertRolesToHTTP(roles []*authpb.RoleInfo) []RoleResponse {
	result := make([]RoleResponse, len(roles))
	for i, role := range roles {
		result[i] = convertRoleToHTTP(role)
	}
	return result
}

func convertPermissionToHTTP(perm *authpb.PermissionInfo) PermissionResponse {
	if perm == nil {
		return PermissionResponse{}
	}
	return PermissionResponse{
		ID:          perm.Id,
		Code:        perm.Code,
		Name:        perm.Name,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		IsSystem:    perm.IsSystem,
		CreatedAt:   perm.CreatedAt,
	}
}

func convertPermissionsToHTTP(perms []*authpb.PermissionInfo) []PermissionResponse {
	result := make([]PermissionResponse, len(perms))
	for i, perm := range perms {
		result[i] = convertPermissionToHTTP(perm)
	}
	return result
}

func convertPermissionGroupToHTTP(group *authpb.PermissionGroupInfo) PermissionGroupResponse {
	if group == nil {
		return PermissionGroupResponse{}
	}
	return PermissionGroupResponse{
		ID:          group.Id,
		Code:        group.Code,
		Name:        group.Name,
		Description: group.Description,
		Icon:        group.Icon,
		Color:       group.Color,
		SortOrder:   int(group.SortOrder),
		Level:       int(group.Level),
		Permissions: convertPermissionsToHTTP(group.Permissions),
	}
}

func convertPermissionGroupsToHTTP(groups []*authpb.PermissionGroupInfo) []PermissionGroupResponse {
	result := make([]PermissionGroupResponse, len(groups))
	for i, group := range groups {
		result[i] = convertPermissionGroupToHTTP(group)
	}
	return result
}

func convertPaginationToHTTP(pagination *commonpb.PaginationMetadata) PaginationMetaResponse {
	if pagination == nil {
		return PaginationMetaResponse{}
	}
	return PaginationMetaResponse{
		Page:       int(pagination.Page),
		PageSize:   int(pagination.PageSize),
		Total:      int64(pagination.Total),
		TotalPages: int(pagination.TotalPages),
	}
}
