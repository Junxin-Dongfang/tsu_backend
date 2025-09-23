# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Go 和微服务架构的 TSU 项目，采用 mqant 框架构建。项目包含多个服务模块：admin、auth、swagger，集成了 Ory Kratos (身份管理)、Ory Keto (权限管理)、Consul (服务发现)、Redis、PostgreSQL 等技术栈。

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

# 启动 auth 服务热重载
air -c .air.auth.toml
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

## 项目架构

### 核心模块结构
- **cmd/**: 服务入口点
  - `admin-server/`: 管理后台服务
  - `auth-server/`: 认证授权服务  
  - `swagger-server/`: API 文档服务
  
- **internal/modules/**: 业务模块
  - `admin/`: 管理模块，提供用户管理、身份管理等功能
  - `auth/`: 认证模块，集成 Ory Kratos/Keto，提供认证授权服务
  - `swagger/`: API 文档模块
  
- **internal/middleware/**: 中间件层
  - 日志、鉴权、限流、错误处理、安全等中间件

- **internal/pkg/**: 公共包
  - `log/`: 统一日志处理
  - `response/`: 统一响应处理

### 服务发现和注册
项目使用 Consul 进行服务发现，每个模块会自动注册 HTTP 服务到 Consul，包含健康检查。

### 配置文件结构
- **configs/base/**: 基础配置
- **configs/environments/**: 环境配置 (local.yaml, dev.yaml 等)
- **configs/server/**: 服务配置 (admin-server.json, auth-server.json)

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
- **实现**: `internal/modules/admin/service/transaction_service.go`
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
- **业务字段**: is_premium, diamond_count, 用户设置等
- **认证字段**: username, email (从 Kratos 同步)

#### 登录历史表 (user_login_history)
- **用途**: 安全审计和用户行为分析
- **字段**: 登录时间、IP地址、设备信息、地理位置等

#### 用户设置表 (user_settings)
- **用途**: 用户偏好和隐私设置
- **字段**: 通知设置、隐私设置、主题偏好等

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
curl -X POST http://127.0.0.1:8081/auth/register \
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
curl -X POST http://127.0.0.1:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "user@example.com",
    "password": "password123",
    "client_ip": "127.0.0.1",
    "user_agent": "curl"
  }'
```

### 故障排除

#### 常见问题
1. **Kratos 服务不可用**: 检查 docker-compose-ory.local.yml 是否正常运行
2. **RPC 调用失败**: 确认 NATS 服务正常，服务间能正常通信
3. **数据不一致**: 检查事务服务日志，确认补偿机制是否触发
4. **迁移问题**: 使用 `make migrate-down` 和 `make migrate-up` 重新应用

#### 日志查看
```bash
# 查看特定服务日志
docker logs tsu_admin --tail 50
docker logs tsu_auth_server --tail 50
docker logs tsu_kratos_service --tail 50

# 查看数据库连接
docker exec tsu_postgres psql -U tsu_user -d tsu_db -c "\dt"
docker exec tsu_ory_postgres psql -U ory_user -d ory_db -c "\dt kratos.*"
```