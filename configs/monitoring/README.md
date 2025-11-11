# TSU 监控系统配置

本目录包含 TSU 游戏服务器的完整监控系统配置，包括 Prometheus 和 Grafana 的所有配置文件。

## 目录结构

```
configs/monitoring/
├── prometheus/           # Prometheus 配置
│   ├── prometheus.yml    # 主配置文件
│   ├── prometheus.prod.yml # 生产环境配置
│   ├── alertmanager.yml   # Alertmanager 配置
│   ├── alerts/           # 告警规则目录
│   │   ├── tsu-alerts.yml
│   │   └── recording.yml
│   └── targets/          # 监控目标配置
│       └── tsu-targets.yml
├── grafana/              # Grafana 配置
│   ├── datasources/      # 数据源配置
│   │   └── prometheus.yml
│   ├── dashboards/       # 仪表盘配置
│   └── provisioning/     # 自动配置
│       ├── datasources.yml
│       └── dashboards.yml
└── README.md            # 本文档
```

## 快速开始

### 1. 启动监控系统

使用提供的脚本快速启动：

```bash
# 本地开发环境
cd deployments/scripts
./setup-monitoring.sh local start

# 生产环境
./setup-monitoring.sh production start
```

### 2. 手动启动

```bash
# 进入对应环境目录
cd deployments/docker-compose/environments/local

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 3. 访问服务

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
  - 默认用户名: `admin`
  - 默认密码: `admin`

## 配置说明

### Prometheus 配置

#### 主要配置项

- **scrape_interval**: 数据抓取间隔 (默认 15s)
- **evaluation_interval**: 规则评估间隔 (默认 15s)
- **storage.tsdb.retention.time**: 数据保留时间
  - 开发环境: 15天
  - 生产环境: 30天

#### 监控目标

1. **Prometheus 自身** (`localhost:9090`)
2. **Game Server** (`host.docker.internal:8072`)
3. **Admin Server** (`host.docker.internal:8071`)

#### 告警规则

位于 `prometheus/alerts/` 目录：
- `tsu-alerts.yml`: TSU 服务告警规则
- `recording.yml`: 预计算规则

### Grafana 配置

#### 数据源

- 自动配置 Prometheus 数据源
- 连接到 `http://prometheus:9090`
- 15秒查询间隔

#### 仪表盘

仪表盘配置文件位于 `grafana/dashboards/` 目录，包含：
- TSU 服务概览仪表盘
- 性能监控仪表盘
- 业务指标仪表盘

## 环境配置

### 本地开发环境

```yaml
# configs/monitoring/prometheus/prometheus.yml
global:
  external_labels:
    cluster: 'tsu-dev'
    environment: 'development'
```

特点：
- 较短的数据保留时间 (15天)
- 详细的服务发现配置
- 调试友好的日志级别

### 生产环境

```yaml
# configs/monitoring/prometheus/prometheus.prod.yml
global:
  external_labels:
    cluster: 'tsu-prod'
    environment: 'production'
```

特点：
- 较长的数据保留时间 (30天)
- 启用管理 API
- 结构化日志
- 健康检查增强

## 数据管理

### 数据目录

监控数据存储在以下位置：
- **开发环境**: `./data/prometheus`, `./data/grafana`
- **生产环境**: `/data/prometheus`, `/data/grafana`

### 数据清理

使用清理脚本管理数据：

```bash
# 完全重置监控系统
cd deployments/scripts
./cleanup-data.sh local reset

# 仅清理数据目录
./cleanup-data.sh local data

# 停止服务
./cleanup-data.sh local stop
```

### 数据备份

生产环境数据备份建议：

```bash
# 备份 Prometheus 数据
sudo tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz /data/prometheus

# 备份 Grafana 数据
sudo tar -czf grafana-backup-$(date +%Y%m%d).tar.gz /data/grafana
```

## 故障排查

### 常见问题

1. **Prometheus 无法抓取指标**
   - 检查目标服务是否运行
   - 验证网络连接
   - 查看抓取配置

2. **Grafana 无法连接 Prometheus**
   - 确认 Prometheus 服务正常
   - 检查数据源配置
   - 验证网络连通性

3. **数据目录权限问题**
   - 确保目录权限正确 (755)
   - 检查 Docker 容器用户权限

### 日志查看

```bash
# 查看所有服务日志
docker-compose -f environments/local/docker-compose.yml logs -f

# 查看特定服务日志
docker-compose -f environments/local/docker-compose.yml logs -f prometheus
docker-compose -f environments/local/docker-compose.yml logs -f grafana
```

### 健康检查

```bash
# 检查 Prometheus
curl http://localhost:9090/-/healthy

# 检查 Grafana
curl http://localhost:3000/api/health

# 检查脚本健康检查
./setup-monitoring.sh local health
```

## 性能优化

### Prometheus 优化

1. **调整抓取间隔**: 根据需要调整 `scrape_interval`
2. **数据保留策略**: 根据存储容量调整 `retention.time`
3. **查询优化**: 使用记录规则预计算常用指标

### Grafana 优化

1. **查询缓存**: 启用查询结果缓存
2. **仪表盘优化**: 减少复杂查询
3. **数据刷新**: 调整仪表盘刷新间隔

## 安全配置

### 生产环境安全

1. **修改默认密码**: 更改 Grafana 管理员密码
2. **网络隔离**: 使用防火墙限制访问
3. **SSL/TLS**: 配置 HTTPS 访问
4. **认证集成**: 集成企业认证系统

### 环境变量

生产环境需要配置以下环境变量：

```bash
# Grafana 配置
GRAFANA_ADMIN_USER=your_admin_user
GRAFANA_ADMIN_PASSWORD=your_secure_password
GRAFANA_ROOT_URL=https://your-domain.com

# 数据目录
DATA_PATH=/data
```

## 扩展配置

### 添加新服务监控

1. 在 `prometheus.yml` 中添加新的 `scrape_config`
2. 配置适当的服务发现
3. 添加相应的标签
4. 创建对应的告警规则

### 自定义告警

1. 在 `prometheus/alerts/` 目录添加新的规则文件
2. 编写 PromQL 查询
3. 配置告警阈值和持续时间
4. 测试告警规则

## 相关文档

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [TSU 项目监控指南](../../docs/monitoring/prometheus-guide.md)
- [告警配置说明](../../../docs/alerting-guide.md)