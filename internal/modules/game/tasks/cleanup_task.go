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

// CleanupTask 定时清理任务
type CleanupTask struct {
	attrOpRepo  interfaces.HeroAttributeOperationRepository
	skillOpRepo interfaces.HeroSkillOperationRepository
	logger      log.Logger
	cron        *cron.Cron
}

// NewCleanupTask 创建定时清理任务实例
func NewCleanupTask(db *sql.DB, logger log.Logger) *CleanupTask {
	return &CleanupTask{
		attrOpRepo:  impl.NewHeroAttributeOperationRepository(db),
		skillOpRepo: impl.NewHeroSkillOperationRepository(db),
		logger:      logger,
	}
}

// Start 启动定时任务
func (t *CleanupTask) Start() {
	// 创建 cron 调度器
	t.cron = cron.New(cron.WithSeconds()) // 支持秒级调度（用于测试）

	// 每天凌晨2点执行清理（生产环境）
	// Cron 表达式: 秒 分 时 日 月 周
	_, err := t.cron.AddFunc("0 0 2 * * *", func() {
		t.logger.Info("【定时任务】开始清理过期操作历史")
		t.cleanupExpiredOperations()
		t.logger.Info("【定时任务】过期操作历史清理完成")
	})

	if err != nil {
		t.logger.Error("【定时任务】添加清理任务失败", err)
		return
	}

	// 启动调度器
	t.cron.Start()
	t.logger.Info("【定时任务】已启动 - 每天凌晨2点执行过期历史清理")
}

// cleanupExpiredOperations 清理过期操作历史
func (t *CleanupTask) cleanupExpiredOperations() {
	ctx := context.Background()

	// 清理7天前的已回退/已过期记录
	expiryDate := time.Now().AddDate(0, 0, -7)

	t.logger.Info("【定时任务】开始清理属性操作历史", "expiry_date", expiryDate.Format("2006-01-02 15:04:05"))

	// 1. 清理属性操作历史
	attrCount, err := t.attrOpRepo.DeleteExpiredOperations(ctx, expiryDate)
	if err != nil {
		t.logger.Error("【定时任务】清理属性操作历史失败", err)
	} else {
		t.logger.Info("【定时任务】属性操作历史清理成功", "deleted_count", attrCount)
	}

	t.logger.Info("【定时任务】开始清理技能操作历史", "expiry_date", expiryDate.Format("2006-01-02 15:04:05"))

	// 2. 清理技能操作历史
	skillCount, err := t.skillOpRepo.DeleteExpiredOperations(ctx, expiryDate)
	if err != nil {
		t.logger.Error("【定时任务】清理技能操作历史失败", err)
	} else {
		t.logger.Info("【定时任务】技能操作历史清理成功", "deleted_count", skillCount)
	}

	t.logger.Info("【定时任务】清理任务执行完成",
		"total_deleted", attrCount+skillCount,
		"attr_deleted", attrCount,
		"skill_deleted", skillCount)
}

// Stop 停止定时任务（优雅关闭）
func (t *CleanupTask) Stop() {
	if t.cron != nil {
		t.logger.Info("【定时任务】正在停止定时任务...")
		ctx := t.cron.Stop()
		<-ctx.Done()
		t.logger.Info("【定时任务】定时任务已停止")
	}
}
