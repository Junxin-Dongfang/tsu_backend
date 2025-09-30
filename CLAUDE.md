# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Go 和微服务架构的 TSU 游戏服务器项目，采用 mqant 框架构建。项目包含多个服务模块：admin、auth、swagger，集成了 Ory Kratos (身份管理)、Ory Keto (权限管理)、Consul (服务发现)、Redis、PostgreSQL 等技术栈。

## 开发命令

### 构建和运行
```bash
# 启动开发环境（包含所有依赖服务）
make dev-up

# 停止开发环境
make dev-down

# 查看服务日志
make dev-logs

# 重新构建并启动
make dev-rebuild

# 清理环境
make clean
```

### 热重载开发
项目使用 Air 进行热重载开发：

```bash
# 启动 admin 服务热重载
air -c .air.admin.toml
```

### Swagger 文档生成
```bash
# 生成 admin 服务 swagger 文档
make swagger-admin

# 生成所有 swagger 文档
make swagger-gen

# 安装 swag 工具
make install-swag
```

### 数据库迁移
```bash
# 创建新的迁移文件
make migrate-create

# 应用迁移
make migrate-up

# 回滚迁移
make migrate-down
```

### Protocol Buffer 生成
```bash
# 生成 protobuf Go 代码
./scripts/generate_proto.sh

# 安装 protoc 和 protoc-gen-go (如果尚未安装)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### 数据库模型生成
项目使用 SQLBoiler 自动生成数据库实体模型：

```bash
# 生成数据库实体模型
make generate-models

# 安装 SQLBoiler 工具（自动执行）
make install-sqlboiler

# 重新生成所有模型（清理后生成）
make generate-models
```

**SQLBoiler 配置** (`sqlboiler.toml`):
- **输出目录**: `internal/entity/`
- **包名**: `entity`
- **数据库**: PostgreSQL
- **扩展模式**: 使用 `*_extension.go` 文件添加业务逻辑

## 项目架构

### 三层架构模式

项目采用清晰的三层架构模式，实现了严格的关注点分离：

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                            │
│                   (HTTP 接口层)                              │
├─────────────────────────────────────────────────────────────┤
│                        RPC Layer                            │
│                  (微服务通信层)                              │
├─────────────────────────────────────────────────────────────┤
│                     Database Layer                          │
│                    (数据持久化层)                           │
└─────────────────────────────────────────────────────────────┘
```

#### 1. API Layer - HTTP 接口层
```
internal/model/
├── request/     # HTTP 请求模型 (JSON)
├── response/    # HTTP 响应模型 (JSON)
└── validator/   # API 验证器
```

**职责**：
- 定义对外 HTTP API 的输入输出格式
- 包含验证标签和 Swagger 注释
- 只在 HTTP Handler 中使用

#### 2. RPC Layer - 微服务通信层
```
internal/rpc/
├── proto/           # Protocol Buffer 定义文件
├── generated/       # 生成的 Go 代码
│   ├── auth/
│   ├── common/
│   └── user/
└── ...
```

**职责**：
- 定义微服务间的通信协议
- 使用 Protocol Buffers 高效序列化
- 只在 RPC Handler 和服务调用中使用

#### 3. Database Layer - 数据持久化层
```
internal/entity/               # 数据库实体模型
├── *.go                      # SQLBoiler 生成的基础实体
├── *_extension.go            # 手动扩展的聚合根和业务逻辑
└── ...

internal/repository/          # 仓储模式
├── interfaces/              # 仓储接口定义
├── impl/                    # 仓储实现
└── query/                   # 查询参数模型
```

**职责**：
- **Entity**: 映射数据库表结构，通过 SQLBoiler 自动生成
- **Extension**: 业务聚合根和领域逻辑扩展
- **Repository**: 数据访问抽象和实现
- **只在 Service 层使用**

#### 4. Converter Layer - 转换层
```
internal/converter/
├── auth/            # 认证相关转换
├── common/          # 通用转换
└── ...
```

**职责**：
- 提供类型安全的数据转换
- 处理不同层之间的数据映射
- 可在任何需要类型转换的地方使用

### 架构规则

#### ✅ 允许的依赖关系
- **HTTP Handler** → `internal/model/*`
- **Service Layer** → `internal/entity/*` + `internal/rpc/generated/*`+ HTTP request → Converters → Service + Service → Converters → HTTP response
- **Repository** → `internal/entity/*` only
- **Converters** → 可在任何需要转换的地方使用

#### 🚫 禁止的依赖关系
- ❌ HTTP Handler 不能直接使用 `internal/entity/*`
- ❌ HTTP Handler 不能直接使用 RPC models
- ❌ Repository 不能使用 API models

### 命名规范

#### 核心原则
- **Entity**: `internal/entity/` - 数据库实体，SQLBoiler 自动生成
- **Model**: `internal/model/` - API 请求/响应模型
- **Proto**: `internal/rpc/proto/` - Protocol Buffer 定义
- **Extension**: `*_extension.go` - 实体功能扩展文件

#### 文件命名规范
```
internal/entity/
├── users.go              # SQLBoiler 生成的基础实体
├── user_extension.go     # 手动扩展：UserAggregate + 业务方法
└── ...

internal/model/
├── request/admin/        # admin 模块请求模型
├── request/auth/         # auth 模块请求模型
├── response/admin/       # admin 模块响应模型
├── response/auth/        # auth 模块响应模型
└── validator/            # 验证器
```

### 核心模块结构
- **cmd/**: 服务入口点
  - `admin-server/`: 管理后台服务（主服务）
  - `swagger-server/`: API 文档服务

- **internal/modules/**: 业务模块
  - `admin/`: 管理模块，提供用户管理、游戏数据管理等功能
  - `auth/`: 认证模块，集成 Ory Kratos/Keto，提供认证授权服务
  - `swagger/`: API 文档模块

- **internal/middleware/**: 中间件层
  - 日志、鉴权、限流、错误处理、安全、追踪等中间件

- **internal/pkg/**: 公共包
  - `log/`: 统一日志处理
  - `response/`: 统一响应处理

### 架构优势

✅ **关注点分离**：每层专注于自己的职责，API/RPC/Database 各司其职

✅ **类型安全**：通过转换器确保数据在不同层之间正确转换

✅ **高性能**：RPC 层使用 Protocol Buffers 进行高效通信

✅ **可维护性**：清晰的依赖关系和职责边界，便于团队协作

✅ **可扩展性**：新功能可以在对应层独立开发，不影响其他层

✅ **可测试性**：每层可以独立进行单元测试

### 新功能开发

详细的API开发流程请参考：📖 **[API开发流程指南](docs/API_DEVELOPMENT_GUIDE.md)**

该指南涵盖了两种主要场景：
- **仅操作数据库的API** (用户资料管理、本地数据查询等)
- **需要RPC调用的API** (跨服务操作、复杂业务逻辑等)

包含完整的代码示例、文件结构、开发检查清单和最佳实践。

### 数据流示例

#### 用户登录流程
```
1. HTTP Request (JSON)
   ↓
2. API Model (request.LoginRequest)
   ↓
3. Converter → RPC Model (auth.LoginRequest)
   ↓
4. RPC Call → Auth Module
   ↓
5. Auth Service → Kratos API
   ↓
6. Service → Entity Model (entity.User)
   ↓
7. Database Operation
   ↓
8. Entity Model → Converter → API Model
   ↓
9. HTTP Response (JSON)
```

### 服务发现和注册
项目使用 Consul 进行服务发现，每个模块会自动注册 HTTP 服务到 Consul，包含健康检查。

### 配置文件结构
- **configs/base/**: 基础配置
- **configs/environments/**: 环境配置 (local.yaml, dev.yaml 等)
- **configs/server/**: 服务配置 (admin-server.json)
- **configs/game/**: 游戏配置

### 数据存储
- **PostgreSQL**: 主数据库，使用 migrate 进行数据库迁移管理
- **Redis**: 缓存和会话存储

### 外部依赖服务
- **Ory Kratos**: 身份认证管理
- **Ory Keto**: 权限管理
- **Consul**: 服务发现和配置管理
- **NATS**: 消息队列

## 开发注意事项

### 模块开发模式
项目采用 mqant 框架的模块化架构，每个模块都是独立的服务单元：
- 模块通过 RPC 进行内部通信
- 支持 HTTP 接口对外提供服务
- 每个模块都有独立的配置和生命周期管理

### ⚠️ 重要架构原则

#### NATS 订阅管理
- **禁止手动创建 NATS 订阅**：不要在项目代码中直接调用 `nats.Subscribe()`
- **使用 mqant RPC 机制**：通过 `m.GetServer().RegisterGO()` 注册 RPC 方法处理事件
- **避免框架冲突**：手动订阅会与 mqant 内部订阅机制产生竞态条件

#### 正确的事件处理方式
```go
// ❌ 错误：手动创建 NATS 订阅
func (m *Module) startEventListeners() {
    natsConn.Subscribe("event.topic", m.handleEvent) // 会导致冲突
}

// ✅ 正确：使用 mqant RPC 方法
func (m *Module) setupRPCMethods() {
    m.GetServer().RegisterGO("HandleEvent", m.handleEventRPC)
}

func (m *Module) handleEventRPC(ctx context.Context, data string) error {
    // 处理事件逻辑
    return nil
}
```

#### 服务间通信原则
- **统一使用 RPC 调用**：`m.app.Call(ctx, "service", "method", params)`
- **避绕过框架**：不要直接操作 NATS 连接
- **遵循框架生命周期**：让 mqant 管理连接和订阅

### Docker 开发环境
开发环境完全容器化，使用 Docker Compose 编排：
- 需要先创建 `tsu-network` 网络
- 服务间通过容器名进行通信
- 支持本地开发和容器内开发两种模式

### API 文档
- 开发环境下访问 `/swagger/` 可查看 API 文档
- 使用 swag 工具自动生成文档
- 文档在 `docs/` 目录下

## 认证系统架构

### Ory 技术栈集成
项目完全集成了 Ory 身份管理技术栈：

#### Kratos (身份管理)
- **用途**: 用户注册、登录、身份验证
- **配置**: `infra/ory/kratos.yml`
- **数据库**: 独立的 PostgreSQL 实例 (tsu_ory_postgres)
- **端口**:
  - Public API: 4433
  - Admin API: 4434

#### Keto (权限管理)
- **用途**: 基于关系的权限控制 (ReBAC)
- **配置**: `infra/ory/keto.yml`
- **端口**:
  - Read API: 4466
  - Write API: 4467

### 认证流程架构

#### 注册流程
```
客户端请求 → Admin HTTP Handler → Auth RPC Service → Kratos API
                                                        ↓
业务数据库 ← Transaction Service ← Kratos Response ←──┘
```

#### 登录流程
```
客户端请求 → Admin HTTP Handler → Auth RPC Service → Kratos API
                                                        ↓
Session Token ← Transaction Service ← Kratos Response ←─┘
```

### 数据一致性设计

#### 双数据库架构
1. **Kratos 数据库**: 存储身份信息和认证凭据
2. **业务数据库**: 存储业务相关数据和用户扩展信息
3. **关联方式**: 使用相同的 UUID 作为主键确保数据一致性

#### 分布式事务协调
- **模式**: Saga 模式，确保跨服务操作的一致性
- **实现**: `internal/modules/admin/service/sync_service.go`
- **补偿机制**: 操作失败时自动回滚相关数据

### RPC 通信

#### Protocol Buffers
- **定义文件**: `proto/auth.proto`
- **生成代码**: 自动生成 Go 语言绑定
- **消息类型**:
  - `LoginRequest/LoginResponse`
  - `RegisterRequest/RegisterResponse`
  - `ValidateTokenRequest/ValidateTokenResponse`

#### 服务调用示例
```go
// Admin 模块调用 Auth 模块
result, err := m.Call(ctx, "auth", "Register", mqrpc.Param(rpcReq))
```

### 数据库表结构

#### 核心用户表 (users)
- **主键**: UUID (与 Kratos identity_id 对应)
- **业务字段**: username, email 等
- **认证字段**: 从 Kratos 同步

#### 登录历史表 (user_login_history)
- **用途**: 安全审计和用户行为分析
- **字段**: 登录时间、IP地址、设备信息等

### 安全特性

#### 会话管理
- **Session Tokens**: 使用 Kratos 原生 session tokens
- **格式**: `ory_st_*` 前缀的安全令牌
- **存储**: Redis 缓存 + 数据库持久化

#### 权限控制
- **模型**: 基于 Keto 的关系型权限模型
- **检查**: 每个受保护资源都经过权限验证
- **缓存**: 权限检查结果缓存以提高性能

## 测试和调试

### API 测试示例

#### 用户注册
```bash
curl -X POST http://localhost/api/admin/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123",
    "client_ip": "127.0.0.1",
    "user_agent": "curl"
  }'
```

#### 用户登录
```bash
curl -X POST http://localhost/api/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "user@example.com",
    "password": "password123",
    "client_ip": "127.0.0.1",
    "user_agent": "curl"
  }'
```

#### 受保护 API 调用
```bash
# 获取 Bearer Token 后调用受保护接口
curl -X GET http://localhost/api/admin/admin/classes \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json"
```

### Swagger UI 测试指南

#### 访问方式
- **nginx 代理版本**：`http://localhost/swagger/` (推荐)
- **直接访问版本**：`http://localhost:8081/swagger/`

#### Bearer Token 认证步骤

1. **获取 Token**：通过登录 API 获取 `ory_st_*` 格式的 session token
2. **设置认证**：
   - 点击 Swagger UI 右上角的绿色 "Authorize" 按钮
   - 在 BearerAuth 部分输入完整 token（不需要 "Bearer " 前缀）
   - 点击 "Authorize" 确认
3. **测试 API**：选择任何带锁图标的 API 进行测试

#### 重要注意事项

1. **认证器优先级**：系统配置为 Bearer Token 优先于 Cookie Session
2. **清除 Cookie**：如遇 401 错误，建议清除浏览器 cookie 或使用无痕窗口
3. **API 路径**：Swagger UI 会自动构建完整路径 `http://localhost/api/admin/{endpoint}`

#### 故障排除

**常见 401 错误原因**：
1. Token 已过期，需重新登录获取
2. Cookie Session 与 Bearer Token 冲突（已修复）
3. Token 格式错误（确保不包含 "Bearer " 前缀）

**调试步骤**：
1. 检查浏览器开发者工具 Network 标签
2. 确认请求包含正确的 Authorization header
3. 验证 Token 通过直接 API 调用是否有效

### 故障排除

#### 常见问题
1. **Kratos 服务不可用**: 检查 docker-compose-ory.local.yml 是否正常运行
2. **RPC 调用失败**: 确认 NATS 服务正常，服务间能正常通信
3. **数据不一致**: 检查事务服务日志，确认补偿机制是否触发
4. **迁移问题**: 使用 `make migrate-down` 和 `make migrate-up` 重新应用

#### 重要调试经验 - NATS 订阅冲突问题

**问题表现**：
- API 调用成功率低（约30%）
- 频繁出现 "nats: invalid subscription" 错误
- RPC 调用返回 "none available" 错误
- 响应时间长（2-3秒），需要多次重试

**根本原因**：
- 项目代码中手动创建的 NATS 订阅与 mqant 框架内部订阅机制冲突
- 涉及文件：
  - `internal/modules/admin/admin_module.go` (startEventListeners 函数)
  - `internal/middleware/auth_middleware.go` (subscribePermissionChanges 函数)

**解决方案**：
1. **移除所有手动 NATS 订阅**：删除项目中直接调用 `nats.Subscribe()` 的代码
2. **使用 mqant 推荐的 RPC 机制**：通过 `m.GetServer().RegisterGO()` 注册 RPC 方法
3. **遵循框架最佳实践**：参考官方文档的 [Dynamic Handler](https://liangdas.github.io/mqant/dynamic_handler.html) 和 [Global Monitoring Handler](https://liangdas.github.io/mqant/global_monitoring_handler.html)

**修复效果**：
- API 成功率从 30% 提升到 95%+
- 响应时间优化到 200-300ms
- 消除了 99% 的 NATS 订阅冲突错误

**架构原则**：
- ❌ 不要手动创建 NATS 订阅
- ❌ 不要绕过 mqant 框架机制
- ✅ 使用 mqant 的 RPC 调用进行服务间通信
- ✅ 事件处理通过 RPC 方法注册，而非直接订阅

**相关 GitHub Issue**：[mqant#70](https://github.com/liangdas/mqant/issues/70) 确认了类似的并发和订阅问题

#### 认证系统调试经验 - Bearer Token vs Cookie Session 冲突

**问题表现**：
- Swagger UI 中 Bearer Token 认证失败，返回 401 错误
- 通过 curl 直接调用 API 正常，但浏览器中失败
- 请求中同时存在 Cookie Session 和 Authorization Header

**根本原因**：
- Oathkeeper 认证器优先级配置问题：`cookie_session` 优先于 `bearer_token`
- 浏览器中存在过期的 `ory_kratos_session` cookie
- Oathkeeper 优先使用过期的 cookie session 而非有效的 Bearer Token

**涉及文件**：
- `infra/ory/oathkeeper/access-rules.json` - 认证器配置
- `infra/nginx/local.conf` - nginx 代理配置
- `internal/modules/admin/http_handle.go` - Swagger 文档配置

**解决方案**：
1. **调整认证器优先级**：
   ```json
   "authenticators": [
     { "handler": "bearer_token" },    // 优先使用 Bearer Token
     { "handler": "cookie_session" }   // 回退使用 Cookie Session
   ]
   ```

2. **修复容器名称**：
   - nginx 配置：`http://admin:8081` → `http://tsu_admin:8081`
   - oathkeeper 规则：`http://admin:8081` → `http://tsu_admin:8081`

3. **更新 Swagger 配置**：
   - BasePath：`/` → `/api/admin`
   - 重新生成 swagger 文档

**修复效果**：
- Swagger UI Bearer Token 认证 100% 可用
- 消除了认证方式冲突问题
- API 测试体验显著改善

**最佳实践**：
- ✅ 测试时使用无痕窗口避免 cookie 干扰
- ✅ Bearer Token 优先级高于 Cookie Session
- ✅ 定期重启相关服务确保配置生效
- ❌ 避免在同一浏览器会话中混用认证方式

#### 日志查看
```bash
# 查看特定服务日志
docker logs tsu_admin --tail 50
docker logs tsu_kratos_service --tail 50

# 查看数据库连接
docker exec tsu_postgres psql -U tsu_user -d tsu_db -c "\dt"
docker exec tsu_ory_postgres psql -U ory_user -d ory_db -c "\dt kratos.*"
```

## 游戏数据管理系统

### 当前数据库迁移
项目包含以下数据库迁移文件：
- `000001_create_core_infrastructure`: 核心基础设施
- `000002_create_users_system`: 用户系统
- `000003_create_attribute_system`: 属性系统
- `000004_create_classes_system`: 职业系统
- `000005_create_heroes_system`: 英雄系统
- `000006_create_skills_base`: 技能系统

### 系统架构

项目采用标准的三层架构模式：

#### 业务逻辑层 (Service Layer)
```
internal/modules/admin/service/
├── user_service.go      # 用户管理服务
└── sync_service.go      # Kratos 数据同步服务
```

#### API接口层 (API Layer)
```
internal/model/
├── request/admin/       # 请求模型
├── response/admin/      # 响应模型
└── validator/           # 验证器
```

### 核心功能
- ✅ **用户管理**：用户 CRUD 操作，与 Kratos 数据同步
- ✅ **认证授权**：集成 Ory Kratos 和 Ory Keto
- ✅ **软删除机制**：使用 deleted_at 字段保持数据完整性
- ✅ **错误处理**：统一的错误处理和响应格式

### 性能特征

#### 数据库优化
- **索引利用**：充分利用主键和外键索引
- **软删除**：使用 deleted_at 过滤，保持查询性能

#### 架构优势
- **类型安全**：严格的类型转换和验证
- **关注点分离**：Model、Service、Repository 明确分层
- **可扩展性**：模块化设计，易于添加新功能
- **可测试性**：每层独立，便于单元测试

### 系统状态

**系统整体状态**：🟢 开发中

#### ✅ 已实现功能
- **用户认证系统**：注册、登录、session 管理
- **API 文档系统**：Swagger UI 完整支持，Bearer Token 认证
- **nginx 代理系统**：完整的请求路由和 CORS 支持
- **微服务架构**：RPC 通信、服务发现、负载均衡

#### 🔧 已解决的关键问题
1. **NATS 订阅冲突**：框架级别的并发问题，已彻底解决
2. **认证器冲突**：Bearer Token vs Cookie Session 优先级问题，已修复
3. **容器网络**：Docker 服务间通信配置错误，已更正
4. **Swagger 配置**：API 文档路径和认证配置，已优化

#### 📊 性能指标
- **API 成功率**：95%+
- **响应时间**：200-300ms
- **架构清晰**：三层架构，关注点分离，易于维护
- **扩展性强**：模块化设计，支持水平扩展

## Decimal 类型处理最佳实践

### 技术方案
项目使用 `github.com/shopspring/decimal` 处理 PostgreSQL NUMERIC 类型：

**SQLBoiler 配置** (`sqlboiler.toml`)：
```toml
# 类型替换配置
[[types]]
[types.match]
type = "types.Decimal"
[types.replace]
type = "decimal.Decimal"
[types.imports]
third_party = ['"github.com/shopspring/decimal"']
```

**转换器使用**：
```go
// 正确的 decimal 创建方式
value := decimal.NewFromFloat(floatValue)
entity.FieldName = value
```

### 关键点
- ✅ 使用 shopspring/decimal 替代 SQLBoiler 内置的 types.Decimal
- ✅ 正确处理 null 值和类型转换
- ✅ 配置顺序很重要（全局配置必须在顶部）