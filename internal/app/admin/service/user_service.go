// File: internal/app/admin/service/user_service.go
package service

import (
	"context"
	"net/http"
	"strings"

	client "github.com/ory/client-go"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
)

// UserService 用户管理服务
type UserService struct {
	adminClient *client.APIClient
	logger      log.Logger
}

// NewUserService 创建用户管理服务
func NewUserService(adminURL string, logger log.Logger) (*UserService, error) {
	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{
		{URL: adminURL},
	}

	return &UserService{
		adminClient: client.NewAPIClient(adminConfig),
		logger:      logger,
	}, nil
}

const (
	// IdentityStateActive indicates that the identity is active.
	IdentityStateActive string = "active"
	// IdentityStateInactive indicates that the identity is inactive.
	IdentityStateInactive string = "inactive"
)

// CreateIdentityRequest 创建身份请求
type CreateIdentityRequest struct {
	Email    string                 `json:"email" binding:"required,email"`
	Username string                 `json:"username" binding:"required,min=3,max=30"`
	Password string                 `json:"password,omitempty" binding:"omitempty,min=8"`
	Traits   map[string]interface{} `json:"traits,omitempty"`
}

// UpdateIdentityRequest 更新身份请求
type UpdateIdentityRequest struct {
	Email    string                 `json:"email,omitempty" binding:"omitempty,email"`
	Username string                 `json:"username,omitempty" binding:"omitempty,min=3,max=30"`
	Traits   map[string]interface{} `json:"traits,omitempty"`
}

// ListIdentitiesQuery 列表查询参数
type ListIdentitiesQuery struct {
	Page    int64  `query:"page" validate:"min=1"`
	PerPage int64  `query:"per_page" validate:"min=1,max=1000"`
	Ids     string `query:"ids"`
}

// ListIdentities 列出身份
func (s *UserService) ListIdentities(ctx context.Context, query *ListIdentitiesQuery) ([]client.Identity, int64, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "列出身份",
		log.Int64("page", query.Page),
		log.Int64("per_page", query.PerPage),
	)

	req := s.adminClient.IdentityAPI.ListIdentities(ctx)

	if query.Page > 0 {
		req = req.Page(query.Page)
	}
	if query.PerPage > 0 {
		req = req.PerPage(query.PerPage)
		if query.Ids != "" {
			ids := strings.Split(query.Ids, ",")
			req = req.Ids(ids)
		}
	}

	identities, resp, err := req.Execute()
	if err != nil {
		return nil, 0, s.handleKratosError(ctx, "list_identities", err, resp)
	}

	// 从响应头获取总数
	total := int64(len(identities))
	if xTotalCount := resp.Header.Get("X-Total-Count"); xTotalCount != "" {
		// 解析总数，如果失败则使用当前数量
		// total = parseHeaderCount(xTotalCount)
	}

	s.logger.InfoContext(ctx, "列出身份成功",
		log.Int64("count", int64(len(identities))),
		log.Int64("total", total),
	)

	return identities, total, nil
}

// GetIdentity 获取身份
func (s *UserService) GetIdentity(ctx context.Context, id string) (*client.Identity, *xerrors.AppError) {
	s.logger.DebugContext(ctx, "获取身份",
		log.String("identity_id", id),
	)

	identity, resp, err := s.adminClient.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, s.handleKratosError(ctx, "get_identity", err, resp)
	}

	return identity, nil
}

// CreateIdentity 创建身份
func (s *UserService) CreateIdentity(ctx context.Context, req *CreateIdentityRequest) (*client.Identity, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "创建身份",
		log.String("email", req.Email),
		log.String("username", req.Username),
	)

	// 构建特征数据
	traits := map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	}

	// 合并额外特征
	for k, v := range req.Traits {
		traits[k] = v
	}

	createReq := client.CreateIdentityBody{
		SchemaId: "default", // 使用默认 schema
		Traits:   traits,
	}

	// 如果提供了密码，设置凭据
	if req.Password != "" {
		createReq.Credentials = &client.IdentityWithCredentials{
			Password: &client.IdentityWithCredentialsPassword{
				Config: &client.IdentityWithCredentialsPasswordConfig{
					Password: &req.Password,
				},
			},
		}
	}

	identity, resp, err := s.adminClient.IdentityAPI.CreateIdentity(ctx).
		CreateIdentityBody(createReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "create_identity", err, resp)
	}

	s.logger.InfoContext(ctx, "创建身份成功",
		log.String("identity_id", identity.Id),
		log.String("email", req.Email),
	)

	return identity, nil
}

// UpdateIdentity 更新身份
func (s *UserService) UpdateIdentity(ctx context.Context, id string, req *UpdateIdentityRequest) (*client.Identity, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "更新身份",
		log.String("identity_id", id),
	)

	// 先获取现有身份
	existing, appErr := s.GetIdentity(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	// 合并特征
	traits := existing.Traits.(map[string]interface{})
	if req.Email != "" {
		traits["email"] = req.Email
	}
	if req.Username != "" {
		traits["username"] = req.Username
	}
	for k, v := range req.Traits {
		traits[k] = v
	}

	updateReq := client.UpdateIdentityBody{
		SchemaId: existing.SchemaId,
		Traits:   traits,
		State:    *existing.State,
	}

	identity, resp, err := s.adminClient.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(updateReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "update_identity", err, resp)
	}

	s.logger.InfoContext(ctx, "更新身份成功",
		log.String("identity_id", id),
	)

	return identity, nil
}

// DeleteIdentity 删除身份
func (s *UserService) DeleteIdentity(ctx context.Context, id string) *xerrors.AppError {
	s.logger.InfoContext(ctx, "删除身份",
		log.String("identity_id", id),
	)

	resp, err := s.adminClient.IdentityAPI.DeleteIdentity(ctx, id).Execute()
	if err != nil {
		return s.handleKratosError(ctx, "delete_identity", err, resp)
	}

	s.logger.InfoContext(ctx, "删除身份成功",
		log.String("identity_id", id),
	)

	return nil
}

// DisableIdentity 禁用身份
func (s *UserService) DisableIdentity(ctx context.Context, id string) (*client.Identity, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "禁用身份",
		log.String("identity_id", id),
	)

	// 获取现有身份
	existing, appErr := s.GetIdentity(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	// 更新状态为禁用
	updateReq := client.UpdateIdentityBody{
		SchemaId: existing.SchemaId,
		Traits:   existing.Traits.(map[string]interface{}),
		State:    IdentityStateInactive,
	}

	identity, resp, err := s.adminClient.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(updateReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "disable_identity", err, resp)
	}

	s.logger.InfoContext(ctx, "禁用身份成功",
		log.String("identity_id", id),
	)

	return identity, nil
}

// EnableIdentity 启用身份
func (s *UserService) EnableIdentity(ctx context.Context, id string) (*client.Identity, *xerrors.AppError) {
	s.logger.InfoContext(ctx, "启用身份",
		log.String("identity_id", id),
	)

	// 获取现有身份
	existing, appErr := s.GetIdentity(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	// 更新状态为活动
	updateReq := client.UpdateIdentityBody{
		SchemaId: existing.SchemaId,
		Traits:   existing.Traits.(map[string]interface{}),
		State:    IdentityStateActive,
	}

	identity, resp, err := s.adminClient.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(updateReq).
		Execute()

	if err != nil {
		return nil, s.handleKratosError(ctx, "enable_identity", err, resp)
	}

	s.logger.InfoContext(ctx, "启用身份成功",
		log.String("identity_id", id),
	)

	return identity, nil
}

// handleKratosError 处理 Kratos 错误
func (s *UserService) handleKratosError(ctx context.Context, operation string, err error, resp *http.Response) *xerrors.AppError {
	s.logger.ErrorContext(ctx, "Kratos Admin API 调用失败",
		log.String("operation", operation),
		log.Any("error", err),
	)

	if resp != nil {
		s.logger.DebugContext(ctx, "Kratos 响应详情",
			log.Int("status_code", resp.StatusCode),
			log.String("status", resp.Status),
		)

		switch resp.StatusCode {
		case http.StatusBadRequest:
			return xerrors.FromCode(xerrors.CodeInvalidParams).
				WithService("user-service", operation)
		case http.StatusUnauthorized:
			return xerrors.FromCode(xerrors.CodeAuthenticationFailed).
				WithService("user-service", operation)
		case http.StatusForbidden:
			return xerrors.FromCode(xerrors.CodePermissionDenied).
				WithService("user-service", operation)
		case http.StatusNotFound:
			return xerrors.FromCode(xerrors.CodeResourceNotFound).
				WithService("user-service", operation)
		case http.StatusConflict:
			return xerrors.FromCode(xerrors.CodeDuplicateResource).
				WithService("user-service", operation)
		default:
			if resp.StatusCode >= 500 {
				return xerrors.NewExternalServiceError("kratos", err).
					WithService("user-service", operation)
			}
		}
	}

	return xerrors.NewExternalServiceError("kratos", err).
		WithService("user-service", operation)
}
