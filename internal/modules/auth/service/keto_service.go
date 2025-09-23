// internal/modules/auth/service/keto_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	authpb "tsu-self/proto"
)

type KetoService struct {
	readURL    string
	writeURL   string
	httpClient *http.Client
	logger     log.Logger
}

// Keto 权限关系结构
type KetoRelation struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
	Subject   string `json:"subject"`
}

// 权限检查请求
type KetoCheckRequest struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
	Subject   string `json:"subject"`
}

func NewKetoService(readURL, writeURL string, logger log.Logger) *KetoService {
	return &KetoService{
		readURL:    readURL,
		writeURL:   writeURL,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

// CheckPermission 检查用户权限
func (s *KetoService) CheckPermission(ctx context.Context, req *authpb.CheckPermissionRequest) (*authpb.CheckPermissionResponse, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "检查用户权限",
		log.String("user_id", req.UserId),
		log.String("resource", req.Resource),
		log.String("action", req.Action))

	// 将传统权限模型映射到 Keto 关系
	// 例如：admin_users:create -> user:123 can create admin_users
	ketoReq := &KetoCheckRequest{
		Namespace: "tsu_game", // 使用你配置的 namespace
		Object:    req.Resource,
		Relation:  req.Action,
		Subject:   fmt.Sprintf("user:%s", req.UserId),
	}

	allowed, err := s.checkRelation(ctx, ketoReq)
	if err != nil {
		return nil, err
	}

	return &authpb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}

// AssignRole 分配角色给用户
func (s *KetoService) AssignRole(ctx context.Context, req *authpb.AssignRoleRequest) (*authpb.AssignRoleResponse, *xerrors.AppError) {
	// 创建用户-角色关系: user:123 member role:admin
	relation := &KetoRelation{
		Namespace: "tsu_game",
		Object:    fmt.Sprintf("role:%s", req.Role),
		Relation:  "member",
		Subject:   fmt.Sprintf("user:%s", req.UserId),
	}

	if err := s.createRelation(ctx, relation); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "角色分配成功",
		log.String("user_id", req.UserId),
		log.String("role", req.Role))

	return &authpb.AssignRoleResponse{Success: true}, nil
}

// RevokeRole 撤销用户角色
func (s *KetoService) RevokeRole(ctx context.Context, req *authpb.RevokeRoleRequest) (*authpb.RevokeRoleResponse, *xerrors.AppError) {
	relation := &KetoRelation{
		Namespace: "tsu_game",
		Object:    fmt.Sprintf("role:%s", req.Role),
		Relation:  "member",
		Subject:   fmt.Sprintf("user:%s", req.UserId),
	}

	if err := s.deleteRelation(ctx, relation); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "角色撤销成功",
		log.String("user_id", req.UserId),
		log.String("role", req.Role))

	return &authpb.RevokeRoleResponse{Success: true}, nil
}

// CreateRole 创建角色权限关系
func (s *KetoService) CreateRole(ctx context.Context, req *authpb.CreateRoleRequest) (*authpb.CreateRoleResponse, *xerrors.AppError) {
	// 为角色分配权限: role:admin can create admin_users
	for _, permission := range req.Permissions {
		parts := strings.Split(permission, ":")
		if len(parts) != 2 {
			continue
		}
		resource := parts[0]
		action := parts[1]

		relation := &KetoRelation{
			Namespace: "tsu_game",
			Object:    resource,
			Relation:  action,
			Subject:   fmt.Sprintf("role:%s", req.RoleName),
		}

		if err := s.createRelation(ctx, relation); err != nil {
			return nil, err
		}
	}

	s.logger.InfoContext(ctx, "角色权限创建成功",
		log.String("role", req.RoleName),
		log.Any("permissions", req.Permissions))

	return &authpb.CreateRoleResponse{Success: true}, nil
}

// GetUserRoles 获取用户角色
func (s *KetoService) GetUserRoles(ctx context.Context, userID string) ([]string, *xerrors.AppError) {
	// 查询用户的所有角色关系
	// GET /admin/relation-tuples?subject=user:123&relation=member
	url := fmt.Sprintf("%s/admin/relation-tuples?subject=user:%s&relation=member", s.readURL, userID)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, xerrors.NewExternalServiceError("keto", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.NewExternalServiceError("keto", fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	var result struct {
		RelationTuples []KetoRelation `json:"relation_tuples"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, xerrors.NewExternalServiceError("keto", err)
	}

	roles := make([]string, 0, len(result.RelationTuples))
	for _, tuple := range result.RelationTuples {
		if strings.HasPrefix(tuple.Object, "role:") {
			roleName := strings.TrimPrefix(tuple.Object, "role:")
			roles = append(roles, roleName)
		}
	}

	return roles, nil
}

// === 私有辅助方法 ===

func (s *KetoService) checkRelation(ctx context.Context, req *KetoCheckRequest) (bool, *xerrors.AppError) {
	// POST /relation-tuples/check
	url := fmt.Sprintf("%s/relation-tuples/check", s.readURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return false, xerrors.NewExternalServiceError("keto", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return false, xerrors.NewExternalServiceError("keto", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return false, xerrors.NewExternalServiceError("keto", err)
	}
	defer resp.Body.Close()

	// Keto 返回 200 表示允许，403 表示拒绝
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusForbidden:
		return false, nil
	default:
		return false, xerrors.NewExternalServiceError("keto", fmt.Errorf("unexpected status: %d", resp.StatusCode))
	}
}

func (s *KetoService) createRelation(ctx context.Context, relation *KetoRelation) *xerrors.AppError {
	// PUT /admin/relation-tuples
	url := fmt.Sprintf("%s/admin/relation-tuples", s.writeURL)

	reqBody, err := json.Marshal(relation)
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return xerrors.NewExternalServiceError("keto", fmt.Errorf("create relation failed: %d", resp.StatusCode))
	}

	return nil
}

func (s *KetoService) deleteRelation(ctx context.Context, relation *KetoRelation) *xerrors.AppError {
	// DELETE /admin/relation-tuples
	url := fmt.Sprintf("%s/admin/relation-tuples?namespace=%s&object=%s&relation=%s&subject=%s",
		s.writeURL, relation.Namespace, relation.Object, relation.Relation, relation.Subject)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return xerrors.NewExternalServiceError("keto", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return xerrors.NewExternalServiceError("keto", fmt.Errorf("delete relation failed: %d", resp.StatusCode))
	}

	return nil
}
