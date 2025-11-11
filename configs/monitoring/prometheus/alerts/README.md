# TSU Prometheus 告警规则说明

## 告警阈值调优历史

本文档记录告警阈值的调整历史和决策依据。

### 初始阈值（2025-11-10）

基于设计文档 (`openspec/changes/add-prometheus-monitoring/design.md`) 设定的初始阈值：

| 指标 | 初始阈值 | 级别 | 持续时间 | 说明 |
|------|---------|------|---------|------|
| 错误率 | > 5% | Warning | 5 分钟 | 基线预估 0.1%-1%，5-50倍基线 |
| p99 延迟 | > 1s | Critical | 5 分钟 | SLO: p95 < 200ms，预估 p99 300-500ms |
| p95 延迟 | > 180ms | Warning | 10 分钟 | 接近 SLO (200ms) 的预警 |
| 内存使用 | > 80% | Warning | 10 分钟 | 标准阈值 |
| Goroutine | > 1000 | Warning | 5 分钟 | 待基线测试确定 |
| 慢请求比例 | > 10% | Warning | 5 分钟 | >1s 的请求占比 |
| 玩家数下降 | > 30% | Warning | 5 分钟 | 10分钟内下降幅度 |
| 战斗失败率 | > 70% | Info | 10 分钟 | 游戏平衡性指标 |

### 调优计划

**阶段 1: 基线测量（上线后 1-2 周）**
- [ ] 收集正常运行期间的指标分布
- [ ] 计算 p50/p95/p99 作为基线
- [ ] 记录峰值时段和正常波动范围

**阶段 2: 阈值调整（基线数据收集完成后）**
- [ ] 根据基线调整告警阈值：`告警阈值 = 基线值 × 倍数系数`
- [ ] 每月审查告警触发频率
- [ ] 调整假阳性率高的告警

**阶段 3: 持续优化（长期）**
- [ ] 每季度审查告警质量
- [ ] 根据业务变化更新阈值
- [ ] 删除不再相关的告警规则

---

## 调优记录模板

每次调整阈值时，请在下方添加记录：

### YYYY-MM-DD - 阈值调整

**调整原因**:
- 描述为什么需要调整
- 基线数据或触发频率数据

**调整内容**:
| 告警名称 | 旧阈值 | 新阈值 | 理由 |
|---------|-------|-------|------|
| ... | ... | ... | ... |

**预期效果**:
- 预期告警触发频率变化
- 预期改善的问题

---

## 调优记录

（此处记录实际的调优历史）

### 示例：2025-11-20 - 错误率阈值调整

**调整原因**:
- 上线 2 周后测量的基线错误率为 0.2%
- 当前 5% 阈值从未触发，设置过宽

**调整内容**:
| 告警名称 | 旧阈值 | 新阈值 | 理由 |
|---------|-------|-------|------|
| HighErrorRate | > 5% | > 1% | 5倍基线 (0.2% × 5 = 1%) |

**预期效果**:
- 及时发现异常错误率上升
- 预计每周触发 0-1 次（可接受范围）

---

## 告警响应指南

### Critical 级别告警
- **响应时间**: 立即（5分钟内）
- **处理人员**: 值班工程师
- **升级策略**: 15分钟未解决，升级至技术负责人

### Warning 级别告警
- **响应时间**: 30分钟内
- **处理人员**: 值班工程师
- **升级策略**: 2小时未解决，升级至技术负责人

### Info 级别告警
- **响应时间**: 工作时间内处理
- **处理人员**: 相关开发人员
- **升级策略**: 无需升级

---

## 常用 PromQL 查询

### 错误率
```promql
# 5分钟平均错误率
sum(rate(tsu_errors_total[5m]))
/ sum(rate(tsu_http_requests_total[5m]))
```

### 延迟分位数
```promql
# p50 延迟
histogram_quantile(0.50, rate(tsu_http_request_duration_seconds_bucket[5m]))

# p90 延迟
histogram_quantile(0.90, rate(tsu_http_request_duration_seconds_bucket[5m]))

# p95 延迟
histogram_quantile(0.95, rate(tsu_http_request_duration_seconds_bucket[5m]))

# p99 延迟
histogram_quantile(0.99, rate(tsu_http_request_duration_seconds_bucket[5m]))
```

### 慢请求比例
```promql
# >1s 的慢请求比例
sum(rate(tsu_http_request_duration_seconds_bucket{le="1"}[5m]))
/ sum(rate(tsu_http_request_duration_seconds_count[5m]))

# >2s 的超慢请求比例
sum(rate(tsu_http_request_duration_seconds_bucket{le="2"}[5m]))
/ sum(rate(tsu_http_request_duration_seconds_count[5m]))
```

### QPS
```promql
# 总 QPS
sum(rate(tsu_http_requests_total[1m]))

# 按路径分组的 QPS
sum by (path) (rate(tsu_http_requests_total[1m]))
```

### 业务指标
```promql
# 当前在线玩家数
tsu_game_players_total

# 战斗胜率
sum(rate(tsu_game_battles_total{result="victory"}[10m]))
/ sum(rate(tsu_game_battles_total[10m]))

# 平均战斗耗时
rate(tsu_game_battle_duration_seconds_sum[5m])
/ rate(tsu_game_battle_duration_seconds_count[5m])
```

---

## 参考资料

- Prometheus 告警最佳实践: https://prometheus.io/docs/practices/alerting/
- 项目设计文档: `openspec/changes/add-prometheus-monitoring/design.md`
- 监控使用指南: `docs/monitoring/prometheus-guide.md`
