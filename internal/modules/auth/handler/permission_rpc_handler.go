package handler

import (
	"context"
	"database/sql"

	pb "tsu-self/internal/pb/auth"
	commonpb "tsu-self/internal/pb/common"
	"tsu-self/internal/modules/auth/service"
	authModels "tsu-self/internal/entity/auth"
	"tsu-self/internal/pkg/xerrors"

	"google.golang.org/protobuf/proto"
)

// PermissionRPCHandler 权限管理 RPC 处理器
type PermissionRPCHandler struct {
	db      *sql.DB
	service *service.PermissionService
}

// NewPermissionRPCHandler 创建权限 RPC Handler
func NewPermissionRPCHandler(db *sql.DB, permService *service.PermissionService) *PermissionRPCHandler {
	return &PermissionRPCHandler{
		db:      db,
		service: permService,
	}
}

// ==================== 权限检查 ====================

// CheckUserPermission 检查用户是否拥有指定权限
func (h *PermissionRPCHandler) CheckUserPermission(data []byte) ([]byte, error) {
	req := &pb.CheckUserPermissionRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	allowed, err := h.service.CheckUserPermission(context.Background(), req.UserId, req.PermissionCode)
	if err != nil {
		return nil, err
	}

	resp := &pb.CheckUserPermissionResponse{
		Allowed: allowed,
	}

	return proto.Marshal(resp)
}

// ==================== 角色管理 ====================

// GetRoles 获取角色列表
func (h *PermissionRPCHandler) GetRoles(data []byte) ([]byte, error) {
	req := &pb.GetRolesRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	// 处理分页参数
	offset := 0
	limit := 20 // 默认每页 20 条
	if req.Pagination != nil {
		if req.Pagination.Page > 0 && req.Pagination.PageSize > 0 {
			offset = int(req.Pagination.Page-1) * int(req.Pagination.PageSize)
			limit = int(req.Pagination.PageSize)
		}
	}

	roles, total, err := h.service.GetRoles(context.Background(), req.Keyword, offset, limit)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbRoles := make([]*pb.RoleInfo, len(roles))
	for i, role := range roles {
		pbRoles[i] = roleToProto(role)
	}

	// 计算分页元数据
	totalPages := int32(total) / int32(limit)
	if int32(total)%int32(limit) > 0 {
		totalPages++
	}

	resp := &pb.GetRolesResponse{
		Roles: pbRoles,
		Pagination: &commonpb.PaginationMetadata{
			Page:       req.Pagination.GetPage(),
			PageSize:   req.Pagination.GetPageSize(),
			Total:      int32(total),
			TotalPages: totalPages,
		},
	}

	return proto.Marshal(resp)
}

// CreateRole 创建角色
func (h *PermissionRPCHandler) CreateRole(data []byte) ([]byte, error) {
	req := &pb.CreateRoleRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	role, err := h.service.CreateRole(
		context.Background(),
		req.Code,
		req.Name,
		req.Description,
		req.IsSystem,
		req.IsDefault,
	)
	if err != nil {
		return nil, err
	}

	resp := &pb.CreateRoleResponse{
		Role: roleToProto(role),
	}

	return proto.Marshal(resp)
}

// UpdateRole 更新角色
func (h *PermissionRPCHandler) UpdateRole(data []byte) ([]byte, error) {
	req := &pb.UpdateRoleRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	role, err := h.service.UpdateRole(
		context.Background(),
		req.RoleId,
		req.Name,
		req.Description,
	)
	if err != nil {
		return nil, err
	}

	resp := &pb.UpdateRoleResponse{
		Role: roleToProto(role),
	}

	return proto.Marshal(resp)
}

// DeleteRole 删除角色
func (h *PermissionRPCHandler) DeleteRole(data []byte) ([]byte, error) {
	req := &pb.DeleteRoleRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.DeleteRole(context.Background(), req.RoleId)
	if err != nil {
		return nil, err
	}

	resp := &pb.DeleteRoleResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "角色删除成功",
		},
	}

	return proto.Marshal(resp)
}

// ==================== 权限管理 ====================

// GetPermissions 获取权限列表
func (h *PermissionRPCHandler) GetPermissions(data []byte) ([]byte, error) {
	req := &pb.GetPermissionsRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	// 处理分页参数
	offset := 0
	limit := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 && req.Pagination.PageSize > 0 {
			offset = int(req.Pagination.Page-1) * int(req.Pagination.PageSize)
			limit = int(req.Pagination.PageSize)
		}
	}

	permissions, total, err := h.service.GetPermissions(
		context.Background(),
		req.Keyword,
		req.Resource,
		req.Action,
		offset,
		limit,
	)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbPermissions := make([]*pb.PermissionInfo, len(permissions))
	for i, perm := range permissions {
		pbPermissions[i] = permissionToProto(perm)
	}

	// 计算分页元数据
	totalPages := int32(total) / int32(limit)
	if int32(total)%int32(limit) > 0 {
		totalPages++
	}

	resp := &pb.GetPermissionsResponse{
		Permissions: pbPermissions,
		Pagination: &commonpb.PaginationMetadata{
			Page:       req.Pagination.GetPage(),
			PageSize:   req.Pagination.GetPageSize(),
			Total:      int32(total),
			TotalPages: totalPages,
		},
	}

	return proto.Marshal(resp)
}

// GetPermissionGroups 获取权限分组列表
func (h *PermissionRPCHandler) GetPermissionGroups(data []byte) ([]byte, error) {
	req := &pb.GetPermissionGroupsRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	// 处理分页参数
	offset := 0
	limit := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 && req.Pagination.PageSize > 0 {
			offset = int(req.Pagination.Page-1) * int(req.Pagination.PageSize)
			limit = int(req.Pagination.PageSize)
		}
	}

	groups, total, err := h.service.GetPermissionGroups(
		context.Background(),
		req.Keyword,
		offset,
		limit,
	)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbGroups := make([]*pb.PermissionGroupInfo, len(groups))
	for i, group := range groups {
		pbGroups[i] = permissionGroupToProto(group)
	}

	// 计算分页元数据
	totalPages := int32(total) / int32(limit)
	if int32(total)%int32(limit) > 0 {
		totalPages++
	}

	resp := &pb.GetPermissionGroupsResponse{
		Groups: pbGroups,
		Pagination: &commonpb.PaginationMetadata{
			Page:       req.Pagination.GetPage(),
			PageSize:   req.Pagination.GetPageSize(),
			Total:      int32(total),
			TotalPages: totalPages,
		},
	}

	return proto.Marshal(resp)
}

// ==================== 角色-权限管理 ====================

// GetRolePermissions 获取角色的权限列表
func (h *PermissionRPCHandler) GetRolePermissions(data []byte) ([]byte, error) {
	req := &pb.GetRolePermissionsRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	permissions, err := h.service.GetRolePermissions(context.Background(), req.RoleId)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbPermissions := make([]*pb.PermissionInfo, len(permissions))
	for i, perm := range permissions {
		pbPermissions[i] = permissionToProto(perm)
	}

	resp := &pb.GetRolePermissionsResponse{
		Permissions: pbPermissions,
	}

	return proto.Marshal(resp)
}

// AssignPermissionsToRole 为角色分配权限
func (h *PermissionRPCHandler) AssignPermissionsToRole(data []byte) ([]byte, error) {
	req := &pb.AssignPermissionsToRoleRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.AssignPermissionsToRole(
		context.Background(),
		req.RoleId,
		req.PermissionIds,
		req.OperatorId,
	)
	if err != nil {
		return nil, err
	}

	resp := &pb.AssignPermissionsToRoleResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "权限分配成功",
		},
	}

	return proto.Marshal(resp)
}

// ==================== 用户-角色管理 ====================

// GetUserRoles 获取用户的角色列表
func (h *PermissionRPCHandler) GetUserRoles(data []byte) ([]byte, error) {
	req := &pb.GetUserRolesRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	roles, err := h.service.GetUserRoles(context.Background(), req.UserId)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbRoles := make([]*pb.RoleInfo, len(roles))
	for i, role := range roles {
		pbRoles[i] = roleToProto(role)
	}

	resp := &pb.GetUserRolesResponse{
		Roles: pbRoles,
	}

	return proto.Marshal(resp)
}

// AssignRolesToUser 为用户分配角色
func (h *PermissionRPCHandler) AssignRolesToUser(data []byte) ([]byte, error) {
	req := &pb.AssignRolesToUserRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.AssignRolesToUser(context.Background(), req.UserId, req.RoleCodes)
	if err != nil {
		return nil, err
	}

	resp := &pb.AssignRolesToUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "角色分配成功",
		},
	}

	return proto.Marshal(resp)
}

// RevokeRolesFromUser 撤销用户角色
func (h *PermissionRPCHandler) RevokeRolesFromUser(data []byte) ([]byte, error) {
	req := &pb.RevokeRolesFromUserRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.RevokeRolesFromUser(context.Background(), req.UserId, req.RoleCodes)
	if err != nil {
		return nil, err
	}

	resp := &pb.RevokeRolesFromUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "角色撤销成功",
		},
	}

	return proto.Marshal(resp)
}

// ==================== 用户-权限管理 ====================

// GetUserPermissions 获取用户的所有权限(包括角色权限和直接权限)
func (h *PermissionRPCHandler) GetUserPermissions(data []byte) ([]byte, error) {
	req := &pb.GetUserPermissionsRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	permissions, err := h.service.GetUserPermissions(context.Background(), req.UserId)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf
	pbPermissions := make([]*pb.PermissionInfo, len(permissions))
	for i, perm := range permissions {
		pbPermissions[i] = permissionToProto(perm)
	}

	resp := &pb.GetUserPermissionsResponse{
		Permissions: pbPermissions,
	}

	return proto.Marshal(resp)
}

// GrantPermissionsToUser 直接授予用户权限(绕过角色)
func (h *PermissionRPCHandler) GrantPermissionsToUser(data []byte) ([]byte, error) {
	req := &pb.GrantPermissionsToUserRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.GrantPermissionsToUser(context.Background(), req.UserId, req.PermissionCodes)
	if err != nil {
		return nil, err
	}

	resp := &pb.GrantPermissionsToUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "权限授予成功",
		},
	}

	return proto.Marshal(resp)
}

// RevokePermissionsFromUser 撤销用户直接权限
func (h *PermissionRPCHandler) RevokePermissionsFromUser(data []byte) ([]byte, error) {
	req := &pb.RevokePermissionsFromUserRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		return nil, xerrors.NewInvalidArgumentError("request", "invalid protobuf data")
	}

	err := h.service.RevokePermissionsFromUser(context.Background(), req.UserId, req.PermissionCodes)
	if err != nil {
		return nil, err
	}

	resp := &pb.RevokePermissionsFromUserResponse{
		Status: &commonpb.Status{
			Success: true,
			Message: "权限撤销成功",
		},
	}

	return proto.Marshal(resp)
}

// ==================== 内部辅助方法 ====================

// roleToProto 转换 Role Entity 为 Protobuf
func roleToProto(role *authModels.Role) *pb.RoleInfo {
	return &pb.RoleInfo{
		Id:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description.String,
		IsSystem:    role.IsSystem,
		IsDefault:   role.IsDefault,
		CreatedAt:   role.CreatedAt.Unix(),
		UpdatedAt:   role.UpdatedAt.Unix(),
	}
}

// permissionToProto 转换 Permission Entity 为 Protobuf
func permissionToProto(perm *authModels.Permission) *pb.PermissionInfo {
	return &pb.PermissionInfo{
		Id:          perm.ID,
		Code:        perm.Code,
		Name:        perm.Name,
		Description: perm.Description.String,
		Resource:    perm.Resource,
		Action:      perm.Action,
		IsSystem:    perm.IsSystem,
		CreatedAt:   perm.CreatedAt.Unix(),
	}
}

// permissionGroupToProto 转换 PermissionGroup Entity 为 Protobuf
func permissionGroupToProto(group *authModels.PermissionGroup) *pb.PermissionGroupInfo {
	// 注意: 这里的 permissions 需要单独查询,暂时返回空数组
	// 实际使用时需要通过 Service 层预加载权限列表
	return &pb.PermissionGroupInfo{
		Id:          group.ID,
		Code:        group.Code,
		Name:        group.Name,
		Description: group.Description.String,
		Icon:        group.Icon.String,
		Color:       group.Color.String,
		SortOrder:   int32(group.SortOrder),
		Level:       int32(group.Level),
		Permissions: []*pb.PermissionInfo{}, // 需要预加载
	}
}
