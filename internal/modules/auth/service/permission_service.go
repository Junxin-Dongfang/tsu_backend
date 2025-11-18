// Package service provides business logic for auth module
package service

import (
	"context"
	"database/sql"
	"fmt"

	"tsu-self/internal/entity/auth"
	"tsu-self/internal/modules/auth/client"
	"tsu-self/internal/pkg/xerrors"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// PermissionService 权限管理服务
type PermissionService struct {
	db         *sql.DB
	ketoClient *client.KetoClient
}

// NewPermissionService 创建权限服务
func NewPermissionService(db *sql.DB, ketoClient *client.KetoClient) *PermissionService {
	return &PermissionService{
		db:         db,
		ketoClient: ketoClient,
	}
}

// ==================== 角色管理 ====================

// GetRoles 获取角色列表
func (s *PermissionService) GetRoles(ctx context.Context, keyword string, offset, limit int) ([]*auth.Role, int64, error) {
	mods := []qm.QueryMod{}

	// 搜索条件
	if keyword != "" {
		mods = append(mods, qm.Where("code ILIKE ? OR name ILIKE ?", "%"+keyword+"%", "%"+keyword+"%"))
	}

	// 总数
	total, err := auth.Roles(mods...).Count(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("count", "roles", err)
	}

	// 分页
	mods = append(mods,
		qm.OrderBy("created_at DESC"),
		qm.Offset(offset),
		qm.Limit(limit),
	)

	roles, err := auth.Roles(mods...).All(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("select", "roles", err)
	}

	return roles, total, nil
}

// GetRoleByID 根据ID获取角色
func (s *PermissionService) GetRoleByID(ctx context.Context, roleID string) (*auth.Role, error) {
	role, err := auth.Roles(qm.Where("id = ?", roleID)).One(ctx, s.db)
	if err == sql.ErrNoRows {
		return nil, xerrors.NewRoleError(roleID)
	}
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "roles", err)
	}
	return role, nil
}

// GetRoleByCode 根据Code获取角色
func (s *PermissionService) GetRoleByCode(ctx context.Context, code string) (*auth.Role, error) {
	role, err := auth.Roles(qm.Where("code = ?", code)).One(ctx, s.db)
	if err == sql.ErrNoRows {
		return nil, xerrors.NewRoleError(code)
	}
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "roles", err)
	}
	return role, nil
}

// CreateRole 创建角色
func (s *PermissionService) CreateRole(ctx context.Context, code, name, description string, isSystem, isDefault bool) (*auth.Role, error) {
	// 检查code是否重复
	exists, err := auth.Roles(qm.Where("code = ?", code)).Exists(ctx, s.db)
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "roles", err)
	}
	if exists {
		return nil, xerrors.NewConflictError("role", fmt.Sprintf("code '%s' already exists", code))
	}

	role := &auth.Role{
		Code:        code,
		Name:        name,
		Description: null.StringFrom(description),
		IsSystem:    isSystem,
		IsDefault:   isDefault,
	}

	if err := role.Insert(ctx, s.db, boil.Infer()); err != nil {
		return nil, xerrors.NewDatabaseError("insert", "roles", err)
	}

	return role, nil
}

// UpdateRole 更新角色
func (s *PermissionService) UpdateRole(ctx context.Context, roleID, name, description string) (*auth.Role, error) {
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 系统角色不允许修改
	if role.IsSystem {
		return nil, xerrors.New(xerrors.CodePermissionDenied, "系统角色不可修改")
	}

	role.Name = name
	role.Description = null.StringFrom(description)

	if _, err := role.Update(ctx, s.db, boil.Infer()); err != nil {
		return nil, xerrors.NewDatabaseError("update", "roles", err)
	}

	return role, nil
}

// DeleteRole 删除角色
func (s *PermissionService) DeleteRole(ctx context.Context, roleID string) error {
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return xerrors.New(xerrors.CodePermissionDenied, "系统角色不可删除")
	}

	// 删除角色-权限关联
	if _, err := auth.RolePermissions(qm.Where("role_id = ?", roleID)).DeleteAll(ctx, s.db); err != nil {
		return xerrors.NewDatabaseError("delete", "role_permissions", err)
	}

	// 删除 Keto 中所有与该角色相关的关系
	// 1. 删除用户-角色关系
	userRoleTuples, err := s.ketoClient.ListRelations(ctx, "roles", role.Code, "member", "")
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	if len(userRoleTuples) > 0 {
		if err := s.ketoClient.BatchDeleteRelations(ctx, userRoleTuples); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	// 2. 删除角色-权限关系
	rolePermTuples, err := s.ketoClient.ListRelations(ctx, "permissions", "", "granted", "")
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	var toDelete []*client.RelationTuple
	for _, tuple := range rolePermTuples {
		if tuple.SubjectSet != nil && tuple.SubjectSet.Object == role.Code {
			toDelete = append(toDelete, tuple)
		}
	}
	if len(toDelete) > 0 {
		if err := s.ketoClient.BatchDeleteRelations(ctx, toDelete); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	// 删除角色记录
	if _, err := role.Delete(ctx, s.db); err != nil {
		return xerrors.NewDatabaseError("delete", "roles", err)
	}

	return nil
}

// ==================== 权限管理 ====================

// GetPermissions 获取权限列表
func (s *PermissionService) GetPermissions(ctx context.Context, keyword, resource, action string, offset, limit int) ([]*auth.Permission, int64, error) {
	mods := []qm.QueryMod{}

	// 搜索条件
	if keyword != "" {
		mods = append(mods, qm.Where("code ILIKE ? OR name ILIKE ?", "%"+keyword+"%", "%"+keyword+"%"))
	}
	if resource != "" {
		mods = append(mods, qm.Where("resource = ?", resource))
	}
	if action != "" {
		mods = append(mods, qm.Where("action = ?", action))
	}

	// 总数
	total, err := auth.Permissions(mods...).Count(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("count", "permissions", err)
	}

	// 分页
	mods = append(mods,
		qm.OrderBy("resource, action"),
		qm.Offset(offset),
		qm.Limit(limit),
	)

	permissions, err := auth.Permissions(mods...).All(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("select", "permissions", err)
	}

	return permissions, total, nil
}

// GetPermissionByID 根据ID获取权限
func (s *PermissionService) GetPermissionByID(ctx context.Context, permissionID string) (*auth.Permission, error) {
	permission, err := auth.Permissions(qm.Where("id = ?", permissionID)).One(ctx, s.db)
	if err == sql.ErrNoRows {
		return nil, xerrors.NewNotFoundError("permission", permissionID)
	}
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "permissions", err)
	}
	return permission, nil
}

// GetPermissionByCode 根据Code获取权限
func (s *PermissionService) GetPermissionByCode(ctx context.Context, code string) (*auth.Permission, error) {
	permission, err := auth.Permissions(qm.Where("code = ?", code)).One(ctx, s.db)
	if err == sql.ErrNoRows {
		return nil, xerrors.NewNotFoundError("permission", code)
	}
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "permissions", err)
	}
	return permission, nil
}

// GetPermissionGroups 获取权限分组列表
func (s *PermissionService) GetPermissionGroups(ctx context.Context, keyword string, offset, limit int) ([]*auth.PermissionGroup, int64, error) {
	mods := []qm.QueryMod{}

	if keyword != "" {
		mods = append(mods, qm.Where("code ILIKE ? OR name ILIKE ?", "%"+keyword+"%", "%"+keyword+"%"))
	}

	total, err := auth.PermissionGroups(mods...).Count(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("count", "permission_groups", err)
	}

	mods = append(mods,
		qm.OrderBy("sort_order, created_at"),
		qm.Offset(offset),
		qm.Limit(limit),
	)

	groups, err := auth.PermissionGroups(mods...).All(ctx, s.db)
	if err != nil {
		return nil, 0, xerrors.NewDatabaseError("select", "permission_groups", err)
	}

	return groups, total, nil
}

// GetPermissionGroupByID 根据ID获取权限分组详情(含权限列表)
func (s *PermissionService) GetPermissionGroupByID(ctx context.Context, groupID string) (*auth.PermissionGroup, []*auth.Permission, error) {
	group, err := auth.PermissionGroups(qm.Where("id = ?", groupID)).One(ctx, s.db)
	if err == sql.ErrNoRows {
		return nil, nil, xerrors.NewNotFoundError("permission_group", groupID)
	}
	if err != nil {
		return nil, nil, xerrors.NewDatabaseError("select", "permission_groups", err)
	}

	// 获取分组下的权限
	permissions, err := auth.Permissions(
		qm.InnerJoin("auth.permission_group_members pgm ON auth.permissions.id = pgm.permission_id"),
		qm.Where("pgm.group_id = ?", groupID),
		qm.OrderBy("pgm.sort_order"),
	).All(ctx, s.db)
	if err != nil {
		return nil, nil, xerrors.NewDatabaseError("select", "permissions", err)
	}

	return group, permissions, nil
}

// ==================== 角色-权限管理 ====================

// GetRolePermissions 获取角色的权限列表
func (s *PermissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*auth.Permission, error) {
	// 先获取角色,验证存在性
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 从数据库查询权限(用于管理界面展示)
	permissions, err := auth.Permissions(
		qm.InnerJoin("auth.role_permissions rp ON auth.permissions.id = rp.permission_id"),
		qm.Where("rp.role_id = ?", roleID),
		qm.OrderBy("auth.permissions.resource, auth.permissions.action"),
	).All(ctx, s.db)
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "permissions", err)
	}

	// 同步检查 Keto 中的权限(可选,用于验证一致性)
	ketoPerms, err := s.ketoClient.GetRolePermissions(ctx, role.Code)
	if err != nil {
		// Keto 查询失败不影响返回数据库结果
		fmt.Printf("Warning: failed to get role permissions from Keto: %v\n", err)
	} else {
		// 可以在这里做一致性检查,如果发现不一致可以记录日志
		_ = ketoPerms
	}

	return permissions, nil
}

// AssignPermissionsToRole 为角色分配权限
func (s *PermissionService) AssignPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string, operatorID string) error {
	// 验证角色存在
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// 验证权限存在
	permissions, err := auth.Permissions(qm.WhereIn("id IN ?", toInterfaces(permissionIDs)...)).All(ctx, s.db)
	if err != nil {
		return xerrors.NewDatabaseError("select", "permissions", err)
	}
	if len(permissions) != len(permissionIDs) {
		return xerrors.New(xerrors.CodeInvalidParams, "部分权限不存在")
	}

	// 开启事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return xerrors.NewDatabaseError("begin", "transaction", err)
	}
	defer tx.Rollback()

	// 1. 数据库操作: 先删除旧的关联
	if _, err := auth.RolePermissions(qm.Where("role_id = ?", roleID)).DeleteAll(ctx, tx); err != nil {
		return xerrors.NewDatabaseError("delete", "role_permissions", err)
	}

	// 2. 数据库操作: 插入新的关联
	for _, perm := range permissions {
		rp := &auth.RolePermission{
			RoleID:       role.ID,
			PermissionID: perm.ID,
		}
		// 只有 operatorID 是有效的 UUID 格式时才设置 granted_by
		// 注意: granted_by 是外键,必须关联到 users.id
		// 当前简化实现:不设置 granted_by(允许为 NULL)
		// TODO: 未来需要从请求头获取有效的用户 ID
		_ = operatorID // 暂时忽略

		if err := rp.Insert(ctx, tx, boil.Infer()); err != nil {
			return xerrors.NewDatabaseError("insert", "role_permissions", err)
		}
	}

	// 提交数据库事务
	if err := tx.Commit(); err != nil {
		return xerrors.NewDatabaseError("commit", "transaction", err)
	}

	// 3. Keto 操作: 先删除所有旧的权限关系
	oldPerms, _ := s.ketoClient.GetRolePermissions(ctx, role.Code)
	if len(oldPerms) > 0 {
		if err := s.ketoClient.BatchRevokePermissionsFromRole(ctx, role.Code, oldPerms); err != nil {
			// Keto 操作失败,记录错误但不回滚数据库(因为已经commit)
			return xerrors.NewExternalServiceError("keto", fmt.Errorf("failed to revoke old permissions: %w", err))
		}
	}

	// 4. Keto 操作: 批量添加新的权限关系
	permCodes := make([]string, len(permissions))
	for i, perm := range permissions {
		permCodes[i] = perm.Code
	}
	if err := s.ketoClient.BatchGrantPermissionsToRole(ctx, role.Code, permCodes); err != nil {
		return xerrors.NewExternalServiceError("keto", fmt.Errorf("failed to grant new permissions: %w", err))
	}

	return nil
}

// AddPermissionToRole 为角色添加单个权限
func (s *PermissionService) AddPermissionToRole(ctx context.Context, roleID, permissionID, operatorID string) error {
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	permission, err := s.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return err
	}

	// 检查是否已存在
	exists, err := auth.RolePermissions(
		qm.Where("role_id = ? AND permission_id = ?", roleID, permissionID),
	).Exists(ctx, s.db)
	if err != nil {
		return xerrors.NewDatabaseError("select", "role_permissions", err)
	}
	if exists {
		return xerrors.New(xerrors.CodeDuplicateResource, "权限已分配")
	}

	// 数据库操作
	rp := &auth.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	}
	// TODO: 需要从请求头获取有效的用户 UUID
	_ = operatorID // 暂时忽略

	if err := rp.Insert(ctx, s.db, boil.Infer()); err != nil {
		return xerrors.NewDatabaseError("insert", "role_permissions", err)
	}

	// Keto 操作
	if err := s.ketoClient.GrantPermissionToRole(ctx, role.Code, permission.Code); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	return nil
}

// RemovePermissionFromRole 移除角色的单个权限
func (s *PermissionService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	role, err := s.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	permission, err := s.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return err
	}

	// 数据库操作
	if _, err := auth.RolePermissions(
		qm.Where("role_id = ? AND permission_id = ?", roleID, permissionID),
	).DeleteAll(ctx, s.db); err != nil {
		return xerrors.NewDatabaseError("delete", "role_permissions", err)
	}

	// Keto 操作
	if err := s.ketoClient.RevokePermissionFromRole(ctx, role.Code, permission.Code); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	return nil
}

// ==================== 用户-角色管理 ====================

// AssignRolesToUser 为用户分配角色
func (s *PermissionService) AssignRolesToUser(ctx context.Context, userID string, roleCodes []string) error {
	// 验证角色存在
	roles, err := auth.Roles(qm.WhereIn("code IN ?", toInterfaces(roleCodes)...)).All(ctx, s.db)
	if err != nil {
		return xerrors.NewDatabaseError("select", "roles", err)
	}
	if len(roles) != len(roleCodes) {
		return xerrors.New(xerrors.CodeInvalidParams, "部分角色不存在")
	}

	// 先删除用户的所有角色
	oldRoles, _ := s.ketoClient.GetUserRoles(ctx, userID)
	if len(oldRoles) > 0 {
		for _, oldRole := range oldRoles {
			if err := s.ketoClient.RevokeRoleFromUser(ctx, userID, oldRole); err != nil {
				return xerrors.NewExternalServiceError("keto", err)
			}
		}
	}

	// 批量分配新角色
	for _, roleCode := range roleCodes {
		if err := s.ketoClient.AssignRoleToUser(ctx, userID, roleCode); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	return nil
}

// AddRoleToUser 为用户添加单个角色
func (s *PermissionService) AddRoleToUser(ctx context.Context, userID, roleCode string) error {
	// 验证角色存在
	if _, err := s.GetRoleByCode(ctx, roleCode); err != nil {
		return err
	}

	// Keto 操作
	if err := s.ketoClient.AssignRoleToUser(ctx, userID, roleCode); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	return nil
}

// RemoveRoleFromUser 移除用户的单个角色
func (s *PermissionService) RemoveRoleFromUser(ctx context.Context, userID, roleCode string) error {
	// Keto 操作
	if err := s.ketoClient.RevokeRoleFromUser(ctx, userID, roleCode); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	return nil
}

// GetUserRoles 获取用户的角色列表
func (s *PermissionService) GetUserRoles(ctx context.Context, userID string) ([]*auth.Role, error) {
	// 从 Keto 查询角色代码
	roleCodes, err := s.ketoClient.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, xerrors.NewExternalServiceError("keto", err)
	}

	if len(roleCodes) == 0 {
		return []*auth.Role{}, nil
	}

	// 从数据库查询角色详情
	roles, err := auth.Roles(qm.WhereIn("code IN ?", toInterfaces(roleCodes)...)).All(ctx, s.db)
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "roles", err)
	}

	return roles, nil
}

// ==================== 用户-权限管理 ====================

// GrantPermissionsToUser 直接授予用户权限
func (s *PermissionService) GrantPermissionsToUser(ctx context.Context, userID string, permissionCodes []string) error {
	// 验证权限存在
	permissions, err := auth.Permissions(qm.WhereIn("code IN ?", toInterfaces(permissionCodes)...)).All(ctx, s.db)
	if err != nil {
		return xerrors.NewDatabaseError("select", "permissions", err)
	}
	if len(permissions) != len(permissionCodes) {
		return xerrors.New(xerrors.CodeInvalidParams, "部分权限不存在")
	}

	// Keto 操作
	for _, permCode := range permissionCodes {
		if err := s.ketoClient.GrantPermissionToUser(ctx, userID, permCode); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	return nil
}

// RevokePermissionFromUser 撤销用户的直接权限(单个)
func (s *PermissionService) RevokePermissionFromUser(ctx context.Context, userID, permissionCode string) error {
	// Keto 操作
	if err := s.ketoClient.RevokePermissionFromUser(ctx, userID, permissionCode); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	return nil
}

// RevokePermissionsFromUser 撤销用户的直接权限(批量)
func (s *PermissionService) RevokePermissionsFromUser(ctx context.Context, userID string, permissionCodes []string) error {
	// Keto 批量操作
	for _, code := range permissionCodes {
		if err := s.ketoClient.RevokePermissionFromUser(ctx, userID, code); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	return nil
}

// RevokeRolesFromUser 撤销用户的角色(批量)
func (s *PermissionService) RevokeRolesFromUser(ctx context.Context, userID string, roleCodes []string) error {
	// Keto 批量操作
	for _, roleCode := range roleCodes {
		if err := s.ketoClient.RevokeRoleFromUser(ctx, userID, roleCode); err != nil {
			return xerrors.NewExternalServiceError("keto", err)
		}
	}

	return nil
}

// GetUserPermissions 获取用户的所有权限(包括角色权限和直接权限)
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string) ([]*auth.Permission, error) {
	// 从 Keto 查询权限代码
	permCodes, err := s.ketoClient.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, xerrors.NewExternalServiceError("keto", err)
	}

	if len(permCodes) == 0 {
		return []*auth.Permission{}, nil
	}

	// 从数据库查询权限详情
	permissions, err := auth.Permissions(qm.WhereIn("code IN ?", toInterfaces(permCodes)...)).All(ctx, s.db)
	if err != nil {
		return nil, xerrors.NewDatabaseError("select", "permissions", err)
	}

	return permissions, nil
}

// ==================== 权限检查 ====================

// CheckUserPermission 检查用户是否拥有权限
func (s *PermissionService) CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	allowed, err := s.ketoClient.CheckUserPermission(ctx, userID, permissionCode)
	if err != nil {
		return false, xerrors.NewExternalServiceError("keto", err)
	}
	return allowed, nil
}

// InitializeTeamPermissions 初始化团队权限拓扑,供 Game/Admin 在启动阶段调用
func (s *PermissionService) InitializeTeamPermissions(ctx context.Context) error {
	if s.ketoClient == nil {
		return xerrors.New(xerrors.CodeExternalServiceError, "Keto 客户端未配置")
	}
	if err := s.ketoClient.InitializeTeamPermissions(ctx); err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	return nil
}

// ==================== 辅助函数 ====================

// toInterfaces 将字符串切片转为 interface{} 切片 (用于 WhereIn)
func toInterfaces(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
