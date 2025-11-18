package tasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/robfig/cron/v3"

	"tsu-self/internal/pkg/log"
	"tsu-self/internal/repository/impl"
	"tsu-self/internal/repository/interfaces"
)

// TeamInvitationExpireTask 团队邀请过期定时任务
// 每小时检查一次，将过期的邀请状态更新为 'expired'
type TeamInvitationExpireTask struct {
	invitationRepo interfaces.TeamInvitationRepository
	logger         log.Logger
	cron           *cron.Cron
}

// NewTeamInvitationExpireTask 创建邀请过期任务实例
func NewTeamInvitationExpireTask(db *sql.DB, logger log.Logger) *TeamInvitationExpireTask {
	return &TeamInvitationExpireTask{
		invitationRepo: impl.NewTeamInvitationRepository(db),
		logger:         logger,
	}
}

// Start 启动定时任务
func (t *TeamInvitationExpireTask) Start() {
	// 创建 cron 调度器
	t.cron = cron.New(cron.WithSeconds())

	// 每小时执行一次邀请过期检查
	// Cron 表达式: 秒 分 时 日 月 周
	// "0 30 * * * *" 表示每小时的第30分0秒执行
	_, err := t.cron.AddFunc("0 30 * * * *", func() {
		t.logger.Info("【团队定时任务】开始检查过期邀请")
		t.expireInvitations()
		t.logger.Info("【团队定时任务】过期邀请检查完成")
	})

	if err != nil {
		t.logger.Error("【团队定时任务】添加邀请过期任务失败", err)
		return
	}

	// 启动调度器
	t.cron.Start()
	t.logger.Info("【团队定时任务】邀请过期任务已启动 - 每小时执行一次")
}

// expireInvitations 过期未处理的邀请
func (t *TeamInvitationExpireTask) expireInvitations() {
	ctx := context.Background()

	// 调用 Repository 的过期邀请方法
	expiredCount, err := t.invitationRepo.ExpireInvitations(ctx)
	if err != nil {
		t.logger.Error("【团队定时任务】过期邀请失败", err)
		return
	}

	if expiredCount > 0 {
		t.logger.Info("【团队定时任务】邀请过期成功",
			"expired_count", expiredCount,
			"timestamp", time.Now().Format("2006-01-02 15:04:05"))
	} else {
		t.logger.Debug("【团队定时任务】没有需要过期的邀请")
	}
}

// Stop 停止定时任务（优雅关闭）
func (t *TeamInvitationExpireTask) Stop() {
	if t.cron != nil {
		t.logger.Info("【团队定时任务】正在停止邀请过期任务...")
		ctx := t.cron.Stop()
		<-ctx.Done()
		t.logger.Info("【团队定时任务】邀请过期任务已停止")
	}
}
