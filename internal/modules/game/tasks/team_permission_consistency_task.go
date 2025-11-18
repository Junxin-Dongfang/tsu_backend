package tasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/robfig/cron/v3"

	"tsu-self/internal/modules/game/service"
	"tsu-self/internal/pkg/log"
)

// TeamPermissionConsistencyTask 团队权限一致性检查定时任务
// 每天凌晨执行，检查并修复数据库与 Keto 之间的权限不一致
type TeamPermissionConsistencyTask struct {
	teamPermissionService *service.TeamPermissionService
	logger                log.Logger
	cron                  *cron.Cron
}

// NewTeamPermissionConsistencyTask 创建权限一致性检查任务实例
func NewTeamPermissionConsistencyTask(db *sql.DB, teamPermissionService *service.TeamPermissionService, logger log.Logger) *TeamPermissionConsistencyTask {
	return &TeamPermissionConsistencyTask{
		teamPermissionService: teamPermissionService,
		logger:                logger,
	}
}

// Start 启动定时任务
func (t *TeamPermissionConsistencyTask) Start() {
	// 创建 cron 调度器
	t.cron = cron.New(cron.WithSeconds())

	// 每天凌晨3点执行权限一致性检查
	// Cron 表达式: 秒 分 时 日 月 周
	// "0 0 3 * * *" 表示每天凌晨3点0分0秒执行
	_, err := t.cron.AddFunc("0 0 3 * * *", func() {
		t.logger.Info("【团队定时任务】开始权限一致性检查")
		t.checkConsistency()
		t.logger.Info("【团队定时任务】权限一致性检查完成")
	})

	if err != nil {
		t.logger.Error("【团队定时任务】添加权限一致性检查任务失败", err)
		return
	}

	// 启动调度器
	t.cron.Start()
	t.logger.Info("【团队定时任务】权限一致性检查任务已启动 - 每天凌晨3点执行")
}

// checkConsistency 检查并修复权限一致性
func (t *TeamPermissionConsistencyTask) checkConsistency() {
	ctx := context.Background()

	// 如果 TeamPermissionService 为 nil（Keto 不可用），跳过检查
	if t.teamPermissionService == nil {
		t.logger.Warn("【团队定时任务】TeamPermissionService 未初始化，跳过权限一致性检查")
		return
	}

	// 调用 TeamPermissionService 的一致性检查方法
	err := t.teamPermissionService.CheckConsistency(ctx)
	if err != nil {
		t.logger.Error("【团队定时任务】权限一致性检查失败", err)
		return
	}

	t.logger.Info("【团队定时任务】权限一致性检查成功",
		"timestamp", time.Now().Format("2006-01-02 15:04:05"))
}

// Stop 停止定时任务（优雅关闭）
func (t *TeamPermissionConsistencyTask) Stop() {
	if t.cron != nil {
		t.logger.Info("【团队定时任务】正在停止权限一致性检查任务...")
		ctx := t.cron.Stop()
		<-ctx.Done()
		t.logger.Info("【团队定时任务】权限一致性检查任务已停止")
	}
}
