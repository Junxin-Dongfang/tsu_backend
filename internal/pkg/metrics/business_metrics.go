// File: internal/pkg/metrics/business_metrics.go
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BusinessMetrics 游戏业务指标收集器
type BusinessMetrics struct {
	// 玩家总数（Gauge 类型，可增可减）
	PlayersTotal *prometheus.GaugeVec

	// 战斗次数（按结果分组：victory/defeat/draw）
	BattlesTotal *prometheus.CounterVec

	// 战斗耗时直方图
	BattleDuration *prometheus.HistogramVec

	// 完成任务数
	QuestsCompletedTotal *prometheus.CounterVec

	// 装备获取数（按装备品质分组）
	EquipmentObtainedTotal *prometheus.CounterVec
}

var (
	// DefaultBusinessMetrics 默认的业务指标实例
	DefaultBusinessMetrics *BusinessMetrics
)

// BattleBuckets 是针对战斗耗时优化的 buckets
// 游戏战斗预期时长: 1-30秒为主
// 单位：秒
var BattleBuckets = []float64{
	0.5, // 0.5s
	1,   // 1s
	2,   // 2s
	5,   // 5s
	10,  // 10s
	30,  // 30s
	60,  // 1分钟
}

// init 初始化默认指标
func init() {
	DefaultBusinessMetrics = NewBusinessMetrics("tsu")
}

// NewBusinessMetrics 创建新的业务指标收集器
func NewBusinessMetrics(namespace string) *BusinessMetrics {
	return NewBusinessMetricsWithRegistry(namespace, GetRegisterer())
}

// NewBusinessMetricsWithRegistry 创建新的业务指标收集器（使用自定义注册表）
func NewBusinessMetricsWithRegistry(namespace string, registerer prometheus.Registerer) *BusinessMetrics {
	factory := promauto.With(registerer)

	return &BusinessMetrics{
		PlayersTotal: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "game",
				Name:      "players_total",
				Help:      "Current number of active players",
			},
			[]string{"service"},
		),

		BattlesTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "game",
				Name:      "battles_total",
				Help:      "Total number of battles by result (victory/defeat/draw)",
			},
			[]string{"result", "service"},
		),

		BattleDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "game",
				Name:      "battle_duration_seconds",
				Help:      "Battle duration in seconds",
				Buckets:   BattleBuckets,
			},
			[]string{"service"},
		),

		QuestsCompletedTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "game",
				Name:      "quests_completed_total",
				Help:      "Total number of completed quests",
			},
			[]string{"service"},
		),

		EquipmentObtainedTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "game",
				Name:      "equipment_obtained_total",
				Help:      "Total number of equipment obtained by quality",
			},
			[]string{"quality", "service"},
		),
	}
}

// RecordBattle 记录战斗指标
//
// 参数:
//   - result: 战斗结果 ("victory", "defeat", "draw")
//   - duration: 战斗耗时
//   - service: 服务名称 ("game" 或 "admin")
func (m *BusinessMetrics) RecordBattle(result string, duration time.Duration, service string) {
	service = normalizeServiceName(service)
	m.BattlesTotal.WithLabelValues(result, service).Inc()
	m.BattleDuration.WithLabelValues(service).Observe(duration.Seconds())
}

// RecordQuestCompleted 记录完成任务
func (m *BusinessMetrics) RecordQuestCompleted(service string) {
	service = normalizeServiceName(service)
	m.QuestsCompletedTotal.WithLabelValues(service).Inc()
}

// RecordEquipmentObtained 记录装备获取
//
// 参数:
//   - quality: 装备品质 ("common", "uncommon", "rare", "epic", "legendary")
//   - service: 服务名称
func (m *BusinessMetrics) RecordEquipmentObtained(quality, service string) {
	service = normalizeServiceName(service)
	m.EquipmentObtainedTotal.WithLabelValues(quality, service).Inc()
}

// SetPlayersTotal 设置当前玩家总数
func (m *BusinessMetrics) SetPlayersTotal(count int, service string) {
	service = normalizeServiceName(service)
	m.PlayersTotal.WithLabelValues(service).Set(float64(count))
}

// IncPlayers 增加玩家数
func (m *BusinessMetrics) IncPlayers(service string) {
	service = normalizeServiceName(service)
	m.PlayersTotal.WithLabelValues(service).Inc()
}

// DecPlayers 减少玩家数
func (m *BusinessMetrics) DecPlayers(service string) {
	service = normalizeServiceName(service)
	m.PlayersTotal.WithLabelValues(service).Dec()
}
