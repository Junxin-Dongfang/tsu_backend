// internal/modules/admin/service/transaction_service.go
package service

import (
	"context"
	"fmt"

	"tsu-self/internal/api/response/auth"
	"tsu-self/internal/converter/common"
	"tsu-self/internal/pkg/log"
	"tsu-self/internal/pkg/xerrors"
	authpb "tsu-self/internal/rpc/generated/auth"

	"github.com/jmoiron/sqlx"
)

// TransactionService 事务协调服务，实现 Saga 模式
type TransactionService struct {
	db          *sqlx.DB
	syncService *SyncService
	logger      log.Logger
}

func NewTransactionService(db *sqlx.DB, syncService *SyncService, logger log.Logger) *TransactionService {
	return &TransactionService{
		db:          db,
		syncService: syncService,
		logger:      logger,
	}
}

// LoginTransaction 登录事务协调
func (t *TransactionService) LoginTransaction(ctx context.Context, authResp *authpb.LoginResponse, clientIP string) (*auth.LoginResult, *xerrors.AppError) {
	t.logger.InfoContext(ctx, "开始登录事务协调",
		log.String("identity_id", authResp.IdentityId),
		log.Bool("auth_success", authResp.Success))

	if !authResp.Success {
		return &auth.LoginResult{
			Success:      false,
			SessionToken: "",
			ErrorMessage: authResp.ErrorMessage,
		}, nil
	}

	// 同步用户信息到主数据库
	userInfo, syncErr := t.syncService.SyncUserAfterLogin(ctx, authResp.IdentityId, clientIP)
	if syncErr != nil {
		t.logger.ErrorContext(ctx, "同步用户信息失败", log.Any("error", syncErr))

		// 如果用户不存在，尝试从 Kratos 用户信息创建
		if syncErr.Code == xerrors.CodeUserNotFound && authResp.UserInfo != nil {
			t.logger.InfoContext(ctx, "用户不存在，尝试创建业务用户")
			userInfo, syncErr = t.syncService.CreateBusinessUser(ctx,
				authResp.IdentityId,
				authResp.UserInfo.Email,
				authResp.UserInfo.Username)

			if syncErr != nil {
				// 创建用户失败，需要回滚 Auth 操作
				t.logger.ErrorContext(ctx, "创建业务用户失败，开始回滚", log.Any("error", syncErr))
				// 注意：这里可以调用 auth 模块的 logout 来回滚会话
				return &auth.LoginResult{
					Success:      false,
					SessionToken: "",
					ErrorMessage: "用户数据同步失败",
				}, syncErr
			}
		} else {
			return &auth.LoginResult{
				Success:      false,
				SessionToken: "",
				ErrorMessage: "用户数据同步失败",
			}, syncErr
		}
	}

	// 记录登录历史
	go func() {
		t.syncService.RecordLoginHistory(context.Background(), authResp.IdentityId, clientIP, "", true)
	}()

	return &auth.LoginResult{
		Success:       true,
		SessionToken:  authResp.Token,
		SessionCookie: "", // 根据需要设置
		UserInfo:      common.UserInfoFromEntity(userInfo),
	}, nil
}

// RegisterTransaction 注册事务协调
func (t *TransactionService) RegisterTransaction(ctx context.Context, authResp *authpb.RegisterResponse) (*auth.RegisterResult, *xerrors.AppError) {
	t.logger.InfoContext(ctx, "开始注册事务协调",
		log.String("identity_id", authResp.IdentityId),
		log.Bool("auth_success", authResp.Success))

	if !authResp.Success {
		return &auth.RegisterResult{
			Success:      false,
			IdentityID:   "",
			SessionToken: "",
			ErrorMessage: authResp.ErrorMessage,
		}, nil
	}

	// 创建业务用户
	userInfo, syncErr := t.syncService.CreateBusinessUser(ctx,
		authResp.IdentityId,
		authResp.UserInfo.Email,
		authResp.UserInfo.Username)

	if syncErr != nil {
		t.logger.ErrorContext(ctx, "创建业务用户失败，需要回滚 Kratos 身份", log.Any("error", syncErr))

		// 这里应该调用 auth 模块删除已创建的 Kratos 身份
		// 但由于当前设计中 auth 不应该提供删除接口，我们记录错误并返回失败
		// 实际生产中可能需要设计补偿机制

		return &auth.RegisterResult{
			Success:      false,
			IdentityID:   authResp.IdentityId,
			SessionToken: "",
			ErrorMessage: "用户数据创建失败",
		}, syncErr
	}

	return &auth.RegisterResult{
		Success:      true,
		IdentityID:   authResp.IdentityId,
		SessionToken: authResp.Token,
		UserInfo:     common.UserInfoFromEntity(userInfo),
	}, nil
}

// CompensateRegister 注册补偿事务（回滚业务数据）
func (t *TransactionService) CompensateRegister(ctx context.Context, identityID string) *xerrors.AppError {
	t.logger.InfoContext(ctx, "开始注册补偿事务", log.String("identity_id", identityID))

	err := t.syncService.DeleteUser(ctx, identityID)
	if err != nil {
		t.logger.ErrorContext(ctx, "补偿事务失败", log.Any("error", err))
		return err
	}

	t.logger.InfoContext(ctx, "注册补偿事务完成", log.String("identity_id", identityID))
	return nil
}

// ValidateTransaction 验证事务的数据一致性
func (t *TransactionService) ValidateTransaction(ctx context.Context, identityID string) error {
	// 检查用户是否存在于主数据库
	_, err := t.syncService.GetUserByID(ctx, identityID)
	if err != nil {
		return fmt.Errorf("用户数据不一致: %w", err)
	}

	t.logger.DebugContext(ctx, "事务数据一致性验证通过", log.String("identity_id", identityID))
	return nil
}
