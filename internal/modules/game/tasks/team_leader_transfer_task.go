package tasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/robfig/cron/v3"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/log"
)

// TeamLeaderTransferTask 队长自动转移定时任务
// 每小时检查一次，将7天未活跃的队长转移给最早的管理员或成员
type TeamLeaderTransferTask struct {
	teamService *service.TeamService
	logger      log.Logger
	cron        *cron.Cron
}

// NewTeamLeaderTransferTask 创建队长自动转移任务实例
func NewTeamLeaderTransferTask(db *sql.DB, teamService *service.TeamService, logger log.Logger) *TeamLeaderTransferTask {
	return &TeamLeaderTransferTask{
		teamService: teamService,
		logger:      logger,
	}
}

// Start 启动定时任务
func (t *TeamLeaderTransferTask) Start() {
	// 创建 cron 调度器
	t.cron = cron.New(cron.WithSeconds())

	// 每小时执行一次队长转移检查
	// Cron 表达式: 秒 分 时 日 月 周
	// "0 0 * * * *" 表示每小时的第0分0秒执行
	_, err := t.cron.AddFunc("0 0 * * * *", func() {
		t.logger.Info("【团队定时任务】开始检查不活跃队长")
		t.transferInactiveLeaders()
		t.logger.Info("【团队定时任务】不活跃队长检查完成")
	})

	if err != nil {
		t.logger.Error("【团队定时任务】添加队长转移任务失败", err)
		return
	}

	// 启动调度器
	t.cron.Start()
	t.logger.Info("【团队定时任务】队长自动转移任务已启动 - 每小时执行一次")
}

// transferInactiveLeaders 转移不活跃队长
func (t *TeamLeaderTransferTask) transferInactiveLeaders() {
	ctx := context.Background()

	// 调用 TeamService 的队长转移方法
	err := t.teamService.TransferInactiveLeaders(ctx)
	if err != nil {
		t.logger.Error("【团队定时任务】队长转移失败", err)
		return
	}

	t.logger.Info("【团队定时任务】队长转移检查完成",
		"timestamp", time.Now().Format("2006-01-02 15:04:05"))
}

// Stop 停止定时任务（优雅关闭）
func (t *TeamLeaderTransferTask) Stop() {
	if t.cron != nil {
		t.logger.Info("【团队定时任务】正在停止队长转移任务...")
		ctx := t.cron.Stop()
		<-ctx.Done()
		t.logger.Info("【团队定时任务】队长转移任务已停止")
	}
}
