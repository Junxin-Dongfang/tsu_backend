package handler

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"google.golang.org/protobuf/proto"

	custommiddleware "tsu-self/internal/middleware"
	authpb "tsu-self/internal/pb/auth"
	"tsu-self/internal/pkg/response"
	"tsu-self/internal/pkg/xerrors"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	rpcCaller  module.RPCModule
	respWriter response.Writer
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(rpcCaller module.RPCModule, respWriter response.Writer) *UserHandler {
	return &UserHandler{
		rpcCaller:  rpcCaller,
		respWriter: respWriter,
	}
}

// ==================== HTTP Request/Response Models ====================

// GetUsersRequest 获取用户列表请求
type GetUsersRequest struct {
	Page     int    `query:"page" validate:"omitempty,min=1"`
	PageSize int    `query:"page_size" validate:"omitempty,min=1,max=100"`
	Keyword  string `query:"keyword"`
	IsBanned *bool  `query:"is_banned"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at login_count last_login_at"`
	SortDir  string `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []UserInfo `json:"users"`
	Total      int64      `json:"total"`
	Page       int32      `json:"page"`
	PageSize   int32      `json:"page_size"`
	TotalPages int32      `json:"total_pages"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Nickname    *string `json:"nickname,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	BirthDate   *string `json:"birth_date,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	Language    *string `json:"language,omitempty"`
	IsBanned    bool    `json:"is_banned"`
	BanUntil    *string `json:"ban_until,omitempty"`
	BanReason   *string `json:"ban_reason,omitempty"`
	LoginCount  int32   `json:"login_count"`
	LastLoginAt *string `json:"last_login_at,omitempty"`
	LastLoginIP *string `json:"last_login_ip,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username    *string `json:"username" validate:"omitempty,min=3,max=50"`
	Email       *string `json:"email" validate:"omitempty,email"`
	Nickname    *string `json:"nickname" validate:"omitempty,max=50"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=10,max=20"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,url,max=500"`
	Bio         *string `json:"bio" validate:"omitempty,max=500"`
	BirthDate   *string `json:"birth_date" validate:"omitempty"`
	Gender      *string `json:"gender" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Timezone    *string `json:"timezone" validate:"omitempty,max=50"`
	Language    *string `json:"language" validate:"omitempty,max=10"`
}

// BanUserRequest 封禁用户请求
type BanUserRequest struct {
	BanUntil  *string `json:"ban_until" validate:"omitempty"`
	BanReason string  `json:"ban_reason" validate:"required,min=1,max=500"`
}

// UserOperationResponse 用户操作响应
type UserOperationResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// ==================== HTTP Handlers ====================

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取系统中的用户列表，支持分页、搜索和筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param keyword query string false "搜索关键词"
// @Param is_banned query bool false "筛选封禁状态"
// @Param sort_by query string false "排序字段" Enums(created_at, updated_at, login_count, last_login_at)
// @Param sort_dir query string false "排序方向" Enums(asc, desc)
// @Success 200 {object} response.Response{data=UserListResponse} "用户列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users [get]
func (h *UserHandler) GetUsers(c echo.Context) error {
	// 1. 绑定和验证请求参数
	var req GetUsersRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求参数格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	// 2. 构造 RPC 请求
	rpcReq := &authpb.GetUsersRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Keyword:  req.Keyword,
		SortBy:   req.SortBy,
		SortDir:  req.SortDir,
	}

	if req.IsBanned != nil {
		rpcReq.IsBanned = req.IsBanned
	}

	// 3. 调用 Auth RPC
	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"GetUsers",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 4. 解析响应
	rpcResp := &authpb.GetUsersResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 5. 转换为 HTTP 响应
	users := make([]UserInfo, 0, len(rpcResp.Users))
	for _, u := range rpcResp.Users {
		user := UserInfo{
			ID:         u.UserId,
			Username:   u.Username,
			Email:      u.Email,
			IsBanned:   u.IsBanned,
			LoginCount: u.LoginCount,
			CreatedAt:  u.CreatedAt,
			UpdatedAt:  u.UpdatedAt,
		}

		if u.Nickname != "" {
			user.Nickname = &u.Nickname
		}
		if u.PhoneNumber != "" {
			user.PhoneNumber = &u.PhoneNumber
		}
		if u.AvatarUrl != "" {
			user.AvatarURL = &u.AvatarUrl
		}
		if u.Bio != "" {
			user.Bio = &u.Bio
		}
		if u.BirthDate != "" {
			user.BirthDate = &u.BirthDate
		}
		if u.Gender != "" {
			user.Gender = &u.Gender
		}
		if u.Timezone != "" {
			user.Timezone = &u.Timezone
		}
		if u.Language != "" {
			user.Language = &u.Language
		}
		if u.BanUntil != "" {
			user.BanUntil = &u.BanUntil
		}
		if u.BanReason != "" {
			user.BanReason = &u.BanReason
		}
		if u.LastLoginAt != "" {
			user.LastLoginAt = &u.LastLoginAt
		}
		if u.LastLoginIp != "" {
			user.LastLoginIP = &u.LastLoginIp
		}

		users = append(users, user)
	}

	resp := UserListResponse{
		Users:      users,
		Total:      rpcResp.Total,
		Page:       rpcResp.Page,
		PageSize:   rpcResp.PageSize,
		TotalPages: rpcResp.TotalPages,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// GetUser 获取用户详情
// @Summary 获取单个用户
// @Description 获取指定用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=UserInfo} "用户信息"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 构造 RPC 请求
	rpcReq := &authpb.GetUserRequest{
		UserId: userID,
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"GetUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.NewUserNotFoundError(userID)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 解析响应
	rpcResp := &authpb.GetUserResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换响应
	u := rpcResp.User
	user := UserInfo{
		ID:         u.UserId,
		Username:   u.Username,
		Email:      u.Email,
		IsBanned:   u.IsBanned,
		LoginCount: u.LoginCount,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}

	if u.Nickname != "" {
		user.Nickname = &u.Nickname
	}
	if u.PhoneNumber != "" {
		user.PhoneNumber = &u.PhoneNumber
	}
	if u.AvatarUrl != "" {
		user.AvatarURL = &u.AvatarUrl
	}
	if u.Bio != "" {
		user.Bio = &u.Bio
	}
	if u.BirthDate != "" {
		user.BirthDate = &u.BirthDate
	}
	if u.Gender != "" {
		user.Gender = &u.Gender
	}
	if u.Timezone != "" {
		user.Timezone = &u.Timezone
	}
	if u.Language != "" {
		user.Language = &u.Language
	}
	if u.BanUntil != "" {
		user.BanUntil = &u.BanUntil
	}
	if u.BanReason != "" {
		user.BanReason = &u.BanReason
	}
	if u.LastLoginAt != "" {
		user.LastLoginAt = &u.LastLoginAt
	}
	if u.LastLoginIp != "" {
		user.LastLoginIP = &u.LastLoginIp
	}

	return response.EchoOK(c, h.respWriter, user)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param user body UpdateUserRequest true "用户更新信息"
// @Success 200 {object} response.Response{data=UserOperationResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 绑定和验证请求
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造 RPC 请求
	rpcReq := &authpb.UpdateUserRequest{
		UserId: userID,
	}

	if req.Username != nil {
		rpcReq.Username = req.Username
	}
	if req.Email != nil {
		rpcReq.Email = req.Email
	}
	if req.Nickname != nil {
		rpcReq.Nickname = req.Nickname
	}
	if req.PhoneNumber != nil {
		rpcReq.PhoneNumber = req.PhoneNumber
	}
	if req.AvatarURL != nil {
		rpcReq.AvatarUrl = req.AvatarURL
	}
	if req.Bio != nil {
		rpcReq.Bio = req.Bio
	}
	if req.BirthDate != nil {
		rpcReq.BirthDate = req.BirthDate
	}
	if req.Gender != nil {
		rpcReq.Gender = req.Gender
	}
	if req.Timezone != nil {
		rpcReq.Timezone = req.Timezone
	}
	if req.Language != nil {
		rpcReq.Language = req.Language
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"UpdateUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 解析响应
	rpcResp := &authpb.UpdateUserResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := UserOperationResponse{
		Message: "用户信息更新成功",
		UserID:  userID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// BanUser 封禁用户
// @Summary 封禁用户
// @Description 封禁指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param ban body BanUserRequest true "封禁信息"
// @Success 200 {object} response.Response{data=UserOperationResponse} "封禁成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users/{id}/ban [post]
func (h *UserHandler) BanUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	var req BanUserRequest
	if err := c.Bind(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, "请求格式错误")
	}

	if err := c.Validate(&req); err != nil {
		return response.EchoBadRequest(c, h.respWriter, err.Error())
	}

	// 构造 RPC 请求
	rpcReq := &authpb.BanUserRequest{
		UserId:    userID,
		BanReason: req.BanReason,
	}

	if req.BanUntil != nil {
		rpcReq.BanUntil = req.BanUntil
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"BanUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 解析响应
	rpcResp := &authpb.BanUserResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := UserOperationResponse{
		Message: "用户封禁成功",
		UserID:  userID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// UnbanUser 解禁用户
// @Summary 解禁用户
// @Description 解除指定用户的封禁状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=UserOperationResponse} "解禁成功"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users/{id}/unban [post]
func (h *UserHandler) UnbanUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return response.EchoBadRequest(c, h.respWriter, "用户ID不能为空")
	}

	// 构造 RPC 请求
	rpcReq := &authpb.UnbanUserRequest{
		UserId: userID,
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"UnbanUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 解析响应
	rpcResp := &authpb.UnbanUserResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	resp := UserOperationResponse{
		Message: "用户解禁成功",
		UserID:  userID,
	}

	return response.EchoOK(c, h.respWriter, resp)
}

// ==================== 示例：如何获取当前登录用户 ====================

// GetCurrentUserProfile 获取当前登录用户的个人资料
// 这是一个示例方法，展示如何从认证中间件中获取当前用户信息
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的详细资料（示例）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=UserInfo} "当前用户信息"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /admin/users/me [get]
// @Security BearerAuth
func (h *UserHandler) GetCurrentUserProfile(c echo.Context) error {
	// 从认证中间件获取当前用户信息
	// Oathkeeper 已经验证了 Session，并将用户 ID 注入到 X-User-ID header
	currentUser, err := custommiddleware.GetCurrentUser(c)
	if err != nil {
		return response.EchoUnauthorized(c, h.respWriter, "获取用户信息失败")
	}

	// 当前用户的 Kratos Identity ID
	userID := currentUser.UserID

	// 构造 RPC 请求，获取用户详细信息
	rpcReq := &authpb.GetUserRequest{
		UserId: userID,
	}

	rpcReqBytes, err := proto.Marshal(rpcReq)
	if err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "序列化RPC请求失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 调用 Auth RPC
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	result, errStr := h.rpcCaller.Call(
		ctx,
		"auth",
		"GetUser",
		mqrpc.Param(rpcReqBytes),
	)

	if errStr != "" {
		if ctx.Err() == context.DeadlineExceeded {
			appErr := xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时")
			return response.EchoError(c, h.respWriter, appErr)
		}
		appErr := xerrors.New(xerrors.CodeExternalServiceError, errStr)
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 解析响应
	rpcResp := &authpb.GetUserResponse{}
	resultBytes, ok := result.([]byte)
	if !ok {
		appErr := xerrors.New(xerrors.CodeInternalError, "RPC响应类型错误")
		return response.EchoError(c, h.respWriter, appErr)
	}

	if err := proto.Unmarshal(resultBytes, rpcResp); err != nil {
		appErr := xerrors.Wrap(err, xerrors.CodeInternalError, "解析RPC响应失败")
		return response.EchoError(c, h.respWriter, appErr)
	}

	// 转换为 HTTP 响应
	user := rpcResp.User
	userInfo := UserInfo{
		ID:          user.UserId,
		Username:    user.Username,
		Email:       user.Email,
		Nickname:    stringToPtr(user.Nickname),
		PhoneNumber: stringToPtr(user.PhoneNumber),
		AvatarURL:   stringToPtr(user.AvatarUrl),
		Bio:         stringToPtr(user.Bio),
		BirthDate:   stringToPtr(user.BirthDate),
		Gender:      stringToPtr(user.Gender),
		Timezone:    stringToPtr(user.Timezone),
		Language:    stringToPtr(user.Language),
		IsBanned:    user.IsBanned,
		BanUntil:    stringToPtr(user.BanUntil),
		BanReason:   stringToPtr(user.BanReason),
		LoginCount:  user.LoginCount,
		LastLoginAt: stringToPtr(user.LastLoginAt),
		LastLoginIP: stringToPtr(user.LastLoginIp),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return response.EchoOK(c, h.respWriter, userInfo)
}

// stringToPtr 将空字符串转换为 nil，非空字符串转换为指针
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
