<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

TSU 是一个基于 Go 的游戏服务器项目,采用微服务架构,包含:
- **Game Server**: 游戏核心逻辑服务(端口 8072)
- **Admin Server**: 后台管理服务(端口 8071)
- **认证系统**: 使用 Ory Kratos/Keto/Oathkeeper
- **数据库**: PostgreSQL(多 schema 架构)
- **缓存**: Redis
- **消息队列**: NATS
- **服务发现**: Consul

## 常用命令

### 开发环境
```bash
# 启动完整开发环境(包括 Ory、主服务、Nginx)
make dev-up

# 停止开发环境
make dev-down

# 查看服务日志
make dev-logs

# 重新构建并重启
make dev-rebuild

# 清理所有资源
make clean
```

### 代码生成
```bash
# 生成所有代码(Protobuf + 数据库实体)
make generate

# 单独生成 Protobuf 代码
make proto

# 单独生成数据库实体模型(SQLBoiler)
make generate-entity

# 生成前端错误码枚举
make generate-errors

# 生成 Swagger 文档
make swagger-gen          # 生成所有服务文档
make swagger-admin        # 仅生成 Admin 服务文档
make swagger-game         # 仅生成 Game 服务文档
```

### 数据库迁移
```bash
# 创建新迁移文件
make migrate-create       # 会提示输入迁移名称

# 应用所有待执行的迁移
make migrate-up

# 回滚最后一次迁移
make migrate-down
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/modules/game/service/...

# 运行单个测试
go test -run TestHeroAttributeUpdate ./internal/modules/game/service/

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 代码质量检查
```bash
# 运行 golangci-lint
golangci-lint run

# 自动修复可修复的问题
golangci-lint run --fix

# 检查特定目录
golangci-lint run ./internal/modules/game/...
```

### 监控与性能测试
```bash
# 查看 Prometheus 指标
curl http://localhost:8072/metrics  # Game Server
curl http://localhost:8071/metrics  # Admin Server

# 运行性能测试套件(需要安装 wrk)
./test/scripts/run-performance-tests.sh

# 分析性能测试结果
./test/scripts/analyze-results.sh <timestamp>

# 单独测试某个端点
./test/scripts/benchmark-api.sh "http://localhost:8072/health" 60s 100 4
```

### 生产部署
```bash
# 分步部署(推荐)
make deploy-prod-step1    # 部署基础设施(PostgreSQL, Redis, NATS, Consul)
make deploy-prod-step2    # 部署 Ory 服务
make deploy-prod-step3    # 部署 Admin Server
make deploy-prod-step4    # 部署 Game Server
make deploy-prod-step5    # 部署 Nginx

# 一键部署所有服务
make deploy-prod-all              # 自动模式
make deploy-prod-all-interactive  # 交互确认模式

# 导入游戏配置到生产环境
make import-game-config-prod
```

## 项目架构

### 目录结构
```
tsu-self/
├── cmd/                        # 服务入口
│   ├── game-server/           # 游戏服务器主程序
│   └── admin-server/          # 管理服务器主程序
├── internal/                   # 内部包(不对外暴露)
│   ├── entity/                # 数据库实体(SQLBoiler 生成)
│   │   ├── auth/             # auth schema 实体
│   │   ├── game_config/      # game_config schema 实体
│   │   └── game_runtime/     # game_runtime schema 实体
│   ├── modules/              # 业务模块
│   │   ├── auth/             # 认证模块
│   │   ├── admin/            # 管理模块
│   │   └── game/             # 游戏模块
│   │       ├── handler/      # HTTP 处理器
│   │       ├── service/      # 业务逻辑层
│   │       └── tasks/        # 定时任务
│   ├── middleware/           # HTTP 中间件
│   ├── pb/                   # Protobuf 生成的代码
│   ├── pkg/                  # 通用工具包
│   │   ├── auth/            # 认证工具
│   │   ├── config/          # 配置管理
│   │   ├── i18n/            # 国际化
│   │   ├── log/             # 日志
│   │   ├── metrics/         # Prometheus 指标
│   │   ├── redis/           # Redis 客户端
│   │   ├── response/        # 统一响应格式
│   │   ├── security/        # 安全工具
│   │   ├── validation/      # 数据验证
│   │   └── xerrors/         # 错误处理
│   ├── repository/          # 数据访问层
│   │   ├── interfaces/     # 接口定义
│   │   └── impl/           # 接口实现
│   └── test/               # 测试工具
├── configs/                 # 配置文件
│   ├── base/               # 基础配置
│   ├── environments/       # 环境特定配置
│   ├── game/               # 游戏配置(技能、角色等)
│   └── server/             # 服务器配置
├── migrations/             # 数据库迁移文件
├── deployments/            # 部署配置
│   └── docker-compose/    # Docker Compose 文件
├── proto/                  # Protobuf 定义文件
├── docs/                   # 文档
├── scripts/                # 部署和工具脚本
├── test/                   # 测试用例和报告
└── web/                    # 前端资源
```

### 数据库架构

项目使用 PostgreSQL 的多 schema 设计:

1. **auth schema**: 存储认证相关数据
   - 与 Ory Kratos 集成
   - 用户身份信息

2. **game_config schema**: 存储游戏配置数据(只读)
   - 英雄配置 (heroes)
   - 技能配置 (skills, skill_levels, skill_level_attributes)
   - 游戏内容配置
   - 由配置管理系统维护,业务代码只读

3. **game_runtime schema**: 存储游戏运行时数据
   - 玩家英雄实例 (player_heroes)
   - 英雄属性 (hero_attributes)
   - 英雄技能学习记录 (hero_learned_skills)
   - 玩家游戏状态

### 代码生成工具

- **SQLBoiler**: 从数据库 schema 生成类型安全的 ORM 代码
  - 配置文件: `sqlboiler.auth.toml`, `sqlboiler.game_config.toml`, `sqlboiler.game_runtime.toml`
  - 生成目录: `internal/entity/{auth,game_config,game_runtime}/`

- **Protobuf**: 定义服务间通信协议
  - 定义文件: `proto/`
  - 生成目录: `internal/pb/`

- **Swagger**: 生成 API 文档
  - Game Server 文档: `docs/game/`
  - Admin Server 文档: `docs/admin/`

### 服务架构

#### 分层架构
```
Handler Layer (HTTP/RPC 处理)
    ↓
Service Layer (业务逻辑)
    ↓
Repository Layer (数据访问)
    ↓
Entity Layer (数据模型)
```

#### 关键组件

1. **认证与授权**
   - Ory Kratos: 用户认证
   - Ory Keto: 权限管理(RBAC)
   - Ory Oathkeeper: API 网关和访问控制

2. **服务通信**
   - HTTP/REST: 客户端-服务器通信
   - NATS: 服务间异步消息
   - Protobuf: 结构化数据序列化

3. **数据存储**
   - PostgreSQL: 主数据库(事务性数据)
   - Redis: 缓存和会话存储
   - Consul: 服务配置和键值存储

4. **监控与可观测性**
   - Prometheus: 指标收集和监控
   - Grafana: 指标可视化和告警
   - 指标端点: `/metrics` (Game Server: 8072, Admin Server: 8071)
   - 详细文档: `docs/prometheus-monitoring-guide.md`

## 开发规范

### 代码质量要求

**严格遵循** `.specify/memory/constitution.md` 中定义的项目宪法和 `openspec/specs/code-quality/spec.md` 中的渐进式质量改进契约:

1. **渐进式质量改进契约** - 不可协商
   - **童子军军规**: 每次代码变更必须让代码库比修改前更好
   - **修改现有代码**: 必须偿还至少一处技术债务(添加测试、文档、解决 TODO、改进错误处理等)
   - **新增代码**: 必须满足 100% 质量标准(测试覆盖率 80%+、完整文档注释、通过 linter)
   - **技术债务优先级**: P0(严重) > P1(高) > P2(中) > P3(低)
   - 详见 `openspec/specs/code-quality/spec.md`

2. **测试驱动开发(TDD)** - 不可协商
   - 必须先写测试,后写实现
   - 单元测试覆盖率至少 80%
   - 所有 API 端点必须有集成测试
   - 使用 table-driven tests 处理多个测试场景

3. **代码质量**
   - 必须通过 `golangci-lint` 检查(无警告)
   - 遵循 Go 语言最佳实践
   - 所有公开函数必须有文档注释
   - 错误处理必须显式且有意义(使用 `xerrors.Wrap` 添加上下文)

4. **性能标准**
   - API 端点 p95 延迟 <200ms
   - 数据库查询必须优化索引
   - 使用 Redis 缓存热数据

5. **可观测性**
   - 结构化日志(包含请求 ID、用户 ID)
   - Prometheus 指标
   - 健康检查端点

### 历史代码处理原则

- **不主动大规模重构**: 除非有明确业务需求或严重问题
- **童子军军规**: 修改代码时必须顺手改进周边代码,偿还至少一处技术债务(参考 `TECH_DEBT.md`)
- **新旧隔离**: 新模块必须严格遵守宪法,与旧代码交互的边界明确定义
- **技术债务偿还**: 优先偿还高优先级债务(P0-P1),低优先级债务可以延后

### 提交规范

遵循约定式提交(Conventional Commits):
```
feat: 添加新功能
fix: 修复 bug
docs: 文档更新
refactor: 代码重构
test: 测试相关
chore: 构建/工具链相关
```

### 分支策略
```
main                 # 主分支
feature/xxx          # 新功能分支
bugfix/xxx           # Bug 修复分支
hotfix/xxx           # 紧急修复分支
```

## 环境配置

### 开发环境要求

- Go 1.25.1+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+
- protoc 编译器
- SQLBoiler
- golangci-lint

### 环境变量

主要配置文件:
- `.env`: 本地开发环境配置
- `.env.prod`: 生产环境配置
- `.env.smtp`: SMTP 邮件配置

关键环境变量:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`: 数据库连接
- `REDIS_ADDR`, `REDIS_PASSWORD`: Redis 连接
- `NATS_URL`: NATS 连接
- `CONSUL_ADDR`: Consul 地址
- `KRATOS_PUBLIC_URL`, `KRATOS_ADMIN_URL`: Ory Kratos 地址

### 服务端口

本地开发:
- Game Server: 8072
- Admin Server: 8071
- PostgreSQL: 5432
- Redis: 6379
- NATS: 4222
- Consul: 8500
- Nginx: 80

访问地址:
- 统一 Swagger 入口: http://localhost/swagger
- Game Swagger: http://localhost/game/swagger/index.html

## 常见任务

### 添加新的 API 端点

1. 在 `internal/modules/{module}/handler/` 添加处理函数
2. 在 `internal/modules/{module}/service/` 实现业务逻辑
3. 在 `cmd/{server}/main.go` 注册路由
4. 添加 Swagger 注释
5. 编写单元测试和集成测试
6. 运行 `make swagger-gen` 更新文档

### 添加新的数据库表

1. 创建迁移文件: `make migrate-create`
2. 编写 up/down SQL
3. 应用迁移: `make migrate-up`
4. 更新相应的 `sqlboiler.*.toml` 配置
5. 重新生成实体: `make generate-entity`
6. 编写 repository 接口和实现
7. 添加单元测试

### 添加新的游戏技能

1. 在 `configs/game/` 添加技能配置 JSON
2. 导入到 `game_config.skills` 表
3. 在 `internal/modules/game/service/` 实现技能逻辑
4. 添加技能效果计算函数
5. 编写技能测试用例
6. 更新 `configs/技能配置规范.md`

## 重要文档

- `TECH_DEBT.md`: 技术债务审计报告
- `.specify/memory/constitution.md`: 项目宪法(开发规范)
- `docs/error-response-guide.md`: 错误响应规范
- `docs/i18n-guide.md`: 国际化指南
- `docs/prometheus-monitoring-guide.md`: Prometheus 监控完整指南(30+ PromQL 查询示例)
- `test/prometheus-performance-test.md`: 监控系统性能测试报告
- `configs/prometheus/alerts/tsu-alerts.yml`: Prometheus 告警规则配置
- `configs/技能配置规范.md`: 游戏技能配置规范
- `test/TEST_RESULTS_SUMMARY.md`: 测试结果总结

## 调试技巧

### 查看容器日志
```bash
# 查看特定服务日志
docker logs tsu_game_server_local
docker logs tsu_postgres

# 实时跟踪日志
docker logs -f tsu_game_server_local
```

### 数据库调试
```bash
# 连接到开发数据库
psql -h localhost -p 5432 -U postgres -d tsu_db

# 查看迁移状态
SELECT * FROM schema_migrations;
```

### Redis 调试
```bash
# 连接到 Redis
docker exec -it tsu_redis redis-cli

# 查看所有 key
KEYS *

# 查看特定 key
GET key_name
```

## 性能优化

### 数据库优化
- 使用 `EXPLAIN ANALYZE` 分析慢查询
- 为频繁查询的列添加索引
- 使用 connection pooling
- 避免 N+1 查询问题

### Redis 缓存策略
- 缓存热数据(用户信息、配置数据)
- 使用 pipeline 批量操作
- 设置合理的过期时间
- 使用 Redis 集群提高吞吐量

### Go 性能优化
- 使用 `pprof` 分析性能瓶颈
- 避免 goroutine 泄漏
- 合理使用 `sync.Pool` 减少内存分配
- 使用 `context` 控制超时和取消

## 故障排查

### 服务无法启动
1. 检查 Docker 容器状态: `docker ps -a`
2. 查看服务日志: `docker logs <container_name>`
3. 确认环境变量配置: `.env` 文件
4. 确认依赖服务(PostgreSQL, Redis)是否就绪

### 数据库连接失败
1. 检查数据库容器是否运行
2. 确认数据库连接字符串
3. 检查网络连接: `docker network ls`
4. 查看数据库日志

### 认证失败
1. 检查 Ory Kratos 服务状态
2. 确认 Kratos 配置正确
3. 检查认证 token 是否有效
4. 查看 Oathkeeper 访问规则

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
