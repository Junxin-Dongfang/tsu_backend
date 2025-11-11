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

详细的项目架构信息请参考:
- **目录结构**: 查看项目根目录的目录组织
- **数据库架构**: 参考 `openspec/project.md` 的 Tech Stack 部分
- **服务架构**: 参考 `openspec/project.md`

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

## 开发规范

**所有开发规范都在 `openspec/` 目录中管理。** 请参考:

- **代码质量规范**: `openspec/specs/code-quality/spec.md`
- **项目约定和技术栈**: `openspec/project.md`
- **变更管理流程**: `openspec/AGENTS.md`
- **技术债务清单**: `TECH_DEBT.md` (清单) + `openspec/specs/code-quality/spec.md` (偿还规范)
- **游戏设计质量规范**: `openspec/specs/game-design-quality/spec.md`

### 核心原则（快速参考）

1. **渐进式质量改进契约** (详见 `openspec/specs/code-quality/spec.md`)
   - **童子军军规**: 每次代码变更必须让代码库比修改前更好
   - **修改现有代码**: 必须偿还至少一处技术债务
   - **新增代码**: 必须满足 100% 质量标准(测试覆盖率 80%+、完整文档注释、通过 linter)

2. **测试驱动开发(TDD)**
   - 必须先写测试,后写实现
   - 单元测试覆盖率至少 80%
   - 使用 table-driven tests 处理多个测试场景

3. **提交规范**
   - 遵循约定式提交(Conventional Commits)
   - feat, fix, docs, refactor, test, chore

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

## 重要文档

### 开发指南
- **代码质量规范**: `openspec/specs/code-quality/spec.md`
- **技术债务清单**: `TECH_DEBT.md`
- **技术债务偿还指南**: `docs/development/TECH_DEBT_GUIDE.md`
- **Swagger 文档指南**: `docs/development/SWAGGER_GUIDE.md`

### 游戏设计
- **游戏设计文档**: `docs/game-design/`
- **游戏设计质量规范**: `openspec/specs/game-design-quality/spec.md`

### 监控与运维
- **Prometheus 监控指南**: `docs/monitoring/prometheus-guide.md`
- **监控系统设置指南**: `docs/monitoring/setup-guide.md`

### API 文档
- **统一 Swagger 入口**: http://localhost/swagger
- **Game Swagger**: http://localhost/game/swagger/index.html
- **Admin Swagger**: http://localhost/admin/swagger/index.html

## 其他资源

详细的调试技巧、性能优化、故障排查等内容，请参考:
- **开发指南**: `docs/development/`
- **监控指南**: `docs/monitoring/`
- **测试指南**: `test/README.md`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
