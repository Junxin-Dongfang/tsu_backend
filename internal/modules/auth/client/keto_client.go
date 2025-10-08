// Package client provides clients for external services
package client

import (
	"context"
	"fmt"

	rts "github.com/ory/keto/proto/ory/keto/relation_tuples/v1alpha2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// KetoClient Keto 客户端封装 (使用 gRPC API)
type KetoClient struct {
	readConn    *grpc.ClientConn
	writeConn   *grpc.ClientConn
	readClient  rts.ReadServiceClient
	writeClient rts.WriteServiceClient
	checkClient rts.CheckServiceClient
}

// NewKetoClient 创建 Keto 客户端
// readAddr: Keto Read gRPC 地址 (例如: "localhost:4466")
// writeAddr: Keto Write gRPC 地址 (例如: "localhost:4467")
func NewKetoClient(readAddr, writeAddr string) (*KetoClient, error) {
	if readAddr == "" || writeAddr == "" {
		return nil, fmt.Errorf("keto addresses cannot be empty")
	}

	// 连接读取服务
	readConn, err := grpc.Dial(readAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to keto read service: %w", err)
	}

	// 连接写入服务
	writeConn, err := grpc.Dial(writeAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		readConn.Close()
		return nil, fmt.Errorf("failed to connect to keto write service: %w", err)
	}

	return &KetoClient{
		readConn:    readConn,
		writeConn:   writeConn,
		readClient:  rts.NewReadServiceClient(readConn),
		writeClient: rts.NewWriteServiceClient(writeConn),
		checkClient: rts.NewCheckServiceClient(readConn),
	}, nil
}

// Close 关闭客户端连接
func (k *KetoClient) Close() error {
	var errs []error
	if k.readConn != nil {
		if err := k.readConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close read connection: %w", err))
		}
	}
	if k.writeConn != nil {
		if err := k.writeConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close write connection: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing keto client: %v", errs)
	}
	return nil
}

// ==================== 数据结构 ====================

// RelationTuple 关系三元组 (业务层使用)
type RelationTuple struct {
	Namespace  string      // 命名空间: "roles", "permissions"
	Object     string      // 对象: "admin", "user:create"
	Relation   string      // 关系: "member", "granted"
	SubjectID  string      // 简单主体: "users:alice"
	SubjectSet *SubjectSet // 复杂主体 (用于角色继承)
}

// SubjectSet 主体集合
type SubjectSet struct {
	Namespace string
	Object    string
	Relation  string
}

// ==================== 关系管理 (Write API) ====================

// CreateRelation 创建关系三元组
func (k *KetoClient) CreateRelation(ctx context.Context, tuple *RelationTuple) error {
	req := &rts.TransactRelationTuplesRequest{
		RelationTupleDeltas: []*rts.RelationTupleDelta{
			{
				Action:        rts.RelationTupleDelta_ACTION_INSERT,
				RelationTuple: k.toProtoTuple(tuple),
			},
		},
	}

	_, err := k.writeClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}

	return nil
}

// DeleteRelation 删除关系三元组
func (k *KetoClient) DeleteRelation(ctx context.Context, tuple *RelationTuple) error {
	req := &rts.TransactRelationTuplesRequest{
		RelationTupleDeltas: []*rts.RelationTupleDelta{
			{
				Action:        rts.RelationTupleDelta_ACTION_DELETE,
				RelationTuple: k.toProtoTuple(tuple),
			},
		},
	}

	_, err := k.writeClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

// BatchCreateRelations 批量创建关系三元组
func (k *KetoClient) BatchCreateRelations(ctx context.Context, tuples []*RelationTuple) error {
	deltas := make([]*rts.RelationTupleDelta, len(tuples))
	for i, tuple := range tuples {
		deltas[i] = &rts.RelationTupleDelta{
			Action:        rts.RelationTupleDelta_ACTION_INSERT,
			RelationTuple: k.toProtoTuple(tuple),
		}
	}

	req := &rts.TransactRelationTuplesRequest{
		RelationTupleDeltas: deltas,
	}

	_, err := k.writeClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to batch create relations: %w", err)
	}

	return nil
}

// BatchDeleteRelations 批量删除关系三元组
func (k *KetoClient) BatchDeleteRelations(ctx context.Context, tuples []*RelationTuple) error {
	deltas := make([]*rts.RelationTupleDelta, len(tuples))
	for i, tuple := range tuples {
		deltas[i] = &rts.RelationTupleDelta{
			Action:        rts.RelationTupleDelta_ACTION_DELETE,
			RelationTuple: k.toProtoTuple(tuple),
		}
	}

	req := &rts.TransactRelationTuplesRequest{
		RelationTupleDeltas: deltas,
	}

	_, err := k.writeClient.TransactRelationTuples(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to batch delete relations: %w", err)
	}

	return nil
}

// ==================== 关系查询 (Read API) ====================

// ListRelations 列出关系三元组
func (k *KetoClient) ListRelations(ctx context.Context, namespace, object, relation, subjectID string) ([]*RelationTuple, error) {
	query := &rts.RelationQuery{
		Namespace: &namespace,
	}

	if object != "" {
		query.Object = &object
	}
	if relation != "" {
		query.Relation = &relation
	}
	if subjectID != "" {
		query.Subject = &rts.Subject{
			Ref: &rts.Subject_Id{
				Id: subjectID,
			},
		}
	}

	req := &rts.ListRelationTuplesRequest{
		RelationQuery: query,
	}

	resp, err := k.readClient.ListRelationTuples(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list relations: %w", err)
	}

	tuples := make([]*RelationTuple, len(resp.RelationTuples))
	for i, protoTuple := range resp.RelationTuples {
		tuples[i] = k.fromProtoTuple(protoTuple)
	}

	return tuples, nil
}

// ==================== 权限检查 (Check API) ====================

// CheckPermission 检查权限
func (k *KetoClient) CheckPermission(ctx context.Context, namespace, object, relation, subjectID string) (bool, error) {
	req := &rts.CheckRequest{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		Subject: &rts.Subject{
			Ref: &rts.Subject_Id{
				Id: subjectID,
			},
		},
	}

	resp, err := k.checkClient.Check(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return resp.Allowed, nil
}

// CheckPermissionWithSubjectSet 检查权限 (使用主体集合)
func (k *KetoClient) CheckPermissionWithSubjectSet(ctx context.Context, namespace, object, relation string, subjectSet *SubjectSet) (bool, error) {
	req := &rts.CheckRequest{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		Subject: &rts.Subject{
			Ref: &rts.Subject_Set{
				Set: &rts.SubjectSet{
					Namespace: subjectSet.Namespace,
					Object:    subjectSet.Object,
					Relation:  subjectSet.Relation,
				},
			},
		},
	}

	resp, err := k.checkClient.Check(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check permission with subject set: %w", err)
	}

	return resp.Allowed, nil
}

// ==================== 业务层便捷方法 ====================

// AssignRoleToUser 分配角色给用户
func (k *KetoClient) AssignRoleToUser(ctx context.Context, userID, roleCode string) error {
	return k.CreateRelation(ctx, &RelationTuple{
		Namespace: "roles",
		Object:    roleCode,
		Relation:  "member",
		SubjectID: fmt.Sprintf("users:%s", userID),
	})
}

// RevokeRoleFromUser 撤销用户角色
func (k *KetoClient) RevokeRoleFromUser(ctx context.Context, userID, roleCode string) error {
	return k.DeleteRelation(ctx, &RelationTuple{
		Namespace: "roles",
		Object:    roleCode,
		Relation:  "member",
		SubjectID: fmt.Sprintf("users:%s", userID),
	})
}

// GetUserRoles 获取用户的所有角色
func (k *KetoClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	tuples, err := k.ListRelations(ctx, "roles", "", "member", fmt.Sprintf("users:%s", userID))
	if err != nil {
		return nil, err
	}

	roles := make([]string, len(tuples))
	for i, tuple := range tuples {
		roles[i] = tuple.Object // "admin", "normal_user"
	}

	return roles, nil
}

// GrantPermissionToRole 为角色授予权限
func (k *KetoClient) GrantPermissionToRole(ctx context.Context, roleCode, permissionCode string) error {
	return k.CreateRelation(ctx, &RelationTuple{
		Namespace: "permissions",
		Object:    permissionCode,
		Relation:  "granted",
		SubjectSet: &SubjectSet{
			Namespace: "roles",
			Object:    roleCode,
			Relation:  "member",
		},
	})
}

// RevokePermissionFromRole 撤销角色权限
func (k *KetoClient) RevokePermissionFromRole(ctx context.Context, roleCode, permissionCode string) error {
	return k.DeleteRelation(ctx, &RelationTuple{
		Namespace: "permissions",
		Object:    permissionCode,
		Relation:  "granted",
		SubjectSet: &SubjectSet{
			Namespace: "roles",
			Object:    roleCode,
			Relation:  "member",
		},
	})
}

// GrantPermissionToUser 直接授予用户权限 (绕过角色)
func (k *KetoClient) GrantPermissionToUser(ctx context.Context, userID, permissionCode string) error {
	return k.CreateRelation(ctx, &RelationTuple{
		Namespace: "permissions",
		Object:    permissionCode,
		Relation:  "granted",
		SubjectID: fmt.Sprintf("users:%s", userID),
	})
}

// RevokePermissionFromUser 撤销用户直接权限
func (k *KetoClient) RevokePermissionFromUser(ctx context.Context, userID, permissionCode string) error {
	return k.DeleteRelation(ctx, &RelationTuple{
		Namespace: "permissions",
		Object:    permissionCode,
		Relation:  "granted",
		SubjectID: fmt.Sprintf("users:%s", userID),
	})
}

// CheckUserPermission 检查用户是否拥有权限 (会检查角色和直接权限)
func (k *KetoClient) CheckUserPermission(ctx context.Context, userID, permissionCode string) (bool, error) {
	return k.CheckPermission(ctx, "permissions", permissionCode, "granted", fmt.Sprintf("users:%s", userID))
}

// GetRolePermissions 获取角色的所有权限
func (k *KetoClient) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	// 查询所有权限,然后过滤出属于该角色的
	tuples, err := k.ListRelations(ctx, "permissions", "", "granted", "")
	if err != nil {
		return nil, err
	}

	var permissions []string
	for _, tuple := range tuples {
		// 检查是否是目标角色的权限
		if tuple.SubjectSet != nil &&
			tuple.SubjectSet.Namespace == "roles" &&
			tuple.SubjectSet.Object == roleCode &&
			tuple.SubjectSet.Relation == "member" {
			permissions = append(permissions, tuple.Object)
		}
	}

	return permissions, nil
}

// GetUserPermissions 获取用户的所有权限 (包括角色权限和直接权限)
// 注意: 这个方法性能较低,仅用于管理界面展示
func (k *KetoClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// 查询所有权限定义
	allTuples, err := k.ListRelations(ctx, "permissions", "", "granted", "")
	if err != nil {
		return nil, err
	}

	// 逐个检查用户是否有此权限
	var permissions []string
	for _, tuple := range allTuples {
		allowed, err := k.CheckPermission(ctx, "permissions", tuple.Object, "granted", fmt.Sprintf("users:%s", userID))
		if err != nil {
			continue // 忽略错误,继续检查
		}
		if allowed {
			permissions = append(permissions, tuple.Object)
		}
	}

	return permissions, nil
}

// BatchGrantPermissionsToRole 批量为角色授予权限
func (k *KetoClient) BatchGrantPermissionsToRole(ctx context.Context, roleCode string, permissionCodes []string) error {
	tuples := make([]*RelationTuple, len(permissionCodes))
	for i, permCode := range permissionCodes {
		tuples[i] = &RelationTuple{
			Namespace: "permissions",
			Object:    permCode,
			Relation:  "granted",
			SubjectSet: &SubjectSet{
				Namespace: "roles",
				Object:    roleCode,
				Relation:  "member",
			},
		}
	}

	return k.BatchCreateRelations(ctx, tuples)
}

// BatchRevokePermissionsFromRole 批量撤销角色权限
func (k *KetoClient) BatchRevokePermissionsFromRole(ctx context.Context, roleCode string, permissionCodes []string) error {
	tuples := make([]*RelationTuple, len(permissionCodes))
	for i, permCode := range permissionCodes {
		tuples[i] = &RelationTuple{
			Namespace: "permissions",
			Object:    permCode,
			Relation:  "granted",
			SubjectSet: &SubjectSet{
				Namespace: "roles",
				Object:    roleCode,
				Relation:  "member",
			},
		}
	}

	return k.BatchDeleteRelations(ctx, tuples)
}

// ==================== 内部辅助方法 ====================

// toProtoTuple 转换为 Proto RelationTuple
func (k *KetoClient) toProtoTuple(tuple *RelationTuple) *rts.RelationTuple {
	protoTuple := &rts.RelationTuple{
		Namespace: tuple.Namespace,
		Object:    tuple.Object,
		Relation:  tuple.Relation,
	}

	// 简单主体
	if tuple.SubjectID != "" {
		protoTuple.Subject = &rts.Subject{
			Ref: &rts.Subject_Id{
				Id: tuple.SubjectID,
			},
		}
	}

	// 主体集合
	if tuple.SubjectSet != nil {
		protoTuple.Subject = &rts.Subject{
			Ref: &rts.Subject_Set{
				Set: &rts.SubjectSet{
					Namespace: tuple.SubjectSet.Namespace,
					Object:    tuple.SubjectSet.Object,
					Relation:  tuple.SubjectSet.Relation,
				},
			},
		}
	}

	return protoTuple
}

// fromProtoTuple 从 Proto RelationTuple 转换
func (k *KetoClient) fromProtoTuple(protoTuple *rts.RelationTuple) *RelationTuple {
	tuple := &RelationTuple{
		Namespace: protoTuple.Namespace,
		Object:    protoTuple.Object,
		Relation:  protoTuple.Relation,
	}

	if protoTuple.Subject != nil {
		switch sub := protoTuple.Subject.Ref.(type) {
		case *rts.Subject_Id:
			tuple.SubjectID = sub.Id
		case *rts.Subject_Set:
			tuple.SubjectSet = &SubjectSet{
				Namespace: sub.Set.Namespace,
				Object:    sub.Set.Object,
				Relation:  sub.Set.Relation,
			}
		}
	}

	return tuple
}
