# CLAUDE.md

Claude Code AI 助手工作指南 - DnD RPG 游戏服务端

---

## 🎯 核心原则

1. **教学引导式开发** - 先问"为什么"再给方案，展示不同选项的权衡
2. **用中文回答** - 所有响应使用中文
3. **使用 TodoWrite** - 复杂任务必须跟踪进度

---

## 📋 项目概述

**DnD 5e 规则的回合制 RPG**，采用 Go 微服务架构

**技术栈**:
- 框架: mqant (微服务) + Echo (HTTP)
- 数据: PostgreSQL + SQLBoiler (ORM)
- 消息: NATS
- 认证: Ory Kratos + Keto

**快速启动**:
```bash
make dev-up          # 启动 Docker 环境
make migrate-up      # 数据库迁移
make generate        # 生成代码 (Protobuf + ORM)
air -c .air.admin.toml # 热重载启动 admin-server
```

---

## 🏗️ 架构设计

### 模块职责

| 模块 | 职责 | 数据访问 |
|------|------|---------|
| **Admin** | 游戏配置、用户管理 (策划/运营) | `game_config` (读写), `auth` (只读) |
| **Auth** | 认证、权限、Kratos/Keto 同步 | `auth` (读写) |
| **Game** | 战斗、角色、DnD 规则引擎 (玩家) | `game_runtime` (读写), `game_config` (只读) |

### 数据库架构 - Schema 分离

```
PostgreSQL: tsu_db
├─ auth           # 用户账号 (Auth 拥有)
├─ game_config    # 游戏配置 (Admin 管理)
├─ game_runtime   # 运行时数据 (Game 管理)
└─ admin          # 后台数据
```

**黄金规则**: ✅ 跨 schema 写操作必须通过 RPC，只读可直接 SQL

### 数据流 - Protobuf RPC 架构

```
HTTP 请求
  ↓
HTTP Handler (HTTP Models 在 handler 内定义)
  ↓ 转 Protobuf
RPC Handler (使用 internal/pb/*)
  ↓ mqant RPC (Protobuf 序列化)
Service 层 (pb ↔ entity 转换)
  ↓
Repository (使用 internal/entity/*)
  ↓
Database
```

**目录结构**:
```
tsu-self/
├── proto/                   # Protobuf 定义 (RPC 契约)
│   ├── common/             # 跨模块共享 (UserInfo)
│   └── auth/               # Auth RPC 服务
│
├── internal/
│   ├── pb/                 # Protobuf 生成代码
│   ├── entity/             # ORM 模型 (SQLBoiler)
│   │   ├── auth/
│   │   ├── game_config/
│   │   └── game_runtime/
│   ├── repository/         # 数据访问层
│   │   ├── interfaces/    # Repository 接口
│   │   └── impl/          # SQLBoiler 实现
│   ├── modules/
│   │   ├── admin/         # Admin 模块
│   │   │   ├── handler/   # HTTP Handler + RPC Handler
│   │   │   └── service/   # 业务逻辑
│   │   ├── auth/          # Auth 模块
│   │   │   ├── client/    # Kratos/Keto Client
│   │   │   ├── handler/
│   │   │   └── service/
│   │   └── game/          # Game 模块 (待开发)
│   └── pkg/               # 共享组件
│       ├── response/      # HTTP 响应
│       ├── xerrors/       # 错误系统
│       ├── validator/
│       ├── config/
│       └── log/
```

**关键原则**:
- ✅ **RPC 必须用 Protobuf** (mqant 官方推荐)
- ✅ **跨模块共享结构在 proto/common/**
- ✅ **HTTP Models 简单时在 Handler 内定义**

---

## ⚠️ mqant 框架关键规则

### 1. Module 初始化 - 值类型嵌入

```go
// ✅ 正确
type AuthModule struct {
    basemodule.BaseModule  // 值类型
    db *sql.DB
}

func (m *AuthModule) OnInit(app module.App, settings *conf.ModuleSettings) {
    // 在每个模块配置服务注册 (不要在 main.go 全局配置!)
    m.BaseModule.OnInit(m, app, settings,
        server.RegisterInterval(15*time.Second),  // 心跳
        server.RegisterTTL(30*time.Second),       // TTL > 心跳
    )
}

// ❌ 错误 - 指针嵌入会 panic
type AuthModule struct {
    *basemodule.BaseModule
}
```

### 2. RPC 方法签名 - 固定格式

```go
// ✅ 正确 - RegisterGO 签名: func([]byte) ([]byte, error)
func (h *RPCHandler) Register(reqBytes []byte) ([]byte, error) {
    ctx := context.Background()  // 内部创建
    req := &authpb.RegisterRequest{}
    proto.Unmarshal(reqBytes, req)
    // ...
    return proto.Marshal(resp)
}

// ❌ 错误 - 带 context 参数会 "params not adapted"
func (h *RPCHandler) Register(ctx context.Context, reqBytes []byte) ([]byte, error)
```

### 3. RPC 调用方法 ⭐ 必须使用 Call

**官方推荐**: 使用 **`Call`** 方法，支持超时和节点选择

**完整示例** (HTTP Handler → RPC):
```go
import (
    "context"
    "time"

    "github.com/liangdas/mqant/module"
    "github.com/liangdas/mqant/rpc"
    "google.golang.org/protobuf/proto"
)

// Handler 结构体 - 使用 rpcCaller 字段
type AuthHandler struct {
    rpcCaller  module.RPCModule  // 用于 RPC 调用
    respWriter response.Writer
}

func NewAuthHandler(rpcCaller module.RPCModule, respWriter response.Writer) *AuthHandler {
    return &AuthHandler{
        rpcCaller:  rpcCaller,
        respWriter: respWriter,
    }
}

func (h *AuthHandler) GetUser(c echo.Context) error {
    // 1. 构造 Protobuf 请求
    rpcReq := &authpb.GetUserRequest{UserId: c.Param("id")}
    rpcReqBytes, _ := proto.Marshal(rpcReq)

    // 2. 调用 RPC (使用 Call 方法)
    ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
    defer cancel()

    result, errStr := h.rpcCaller.Call(
        ctx,
        "auth",                  // 目标模块类型
        "GetUser",               // RPC 方法名
        rpc.Param(rpcReqBytes),  // 参数 (必须用 rpc.Param 包装)
    )

    // 3. 错误处理 (区分超时和业务错误)
    if errStr != "" {
        if ctx.Err() == context.DeadlineExceeded {
            return response.EchoError(c, h.respWriter,
                xerrors.New(xerrors.CodeExternalServiceError, "Auth服务超时"))
        }
        return response.EchoError(c, h.respWriter,
            xerrors.NewUserNotFoundError(userID))
    }

    // 4. 解析响应
    resultBytes, _ := result.([]byte)
    rpcResp := &authpb.GetUserResponse{}
    proto.Unmarshal(resultBytes, rpcResp)

    return response.EchoOK(c, h.respWriter, rpcResp.User)
}
```

**关键要点**:
- ✅ 导入 `"github.com/liangdas/mqant/rpc"` (不是 rpc/util)
- ✅ 参数必须用 `rpc.Param()` 包装 (类型为 `rpc.ParamOption`)
- ✅ Handler 字段命名为 `rpcCaller` (语义明确，只用于 RPC 调用)
- ✅ 始终使用 `context.WithTimeout()` 设置超时
- ✅ 检查 `ctx.Err() == context.DeadlineExceeded` 区分超时

**Call vs Invoke 对比**:

| 特性 | Call | Invoke |
|------|------|--------|
| **超时控制** | ✅ 支持 context | ❌ 不支持 |
| **节点选择** | ✅ 支持过滤器 | ❌ 不支持 |
| **错误处理** | ✅ 可区分超时/取消/业务错误 | ⚠️ 只返回错误字符串 |
| **推荐度** | ⭐⭐⭐⭐⭐ **必须使用** | ⚠️ 已过时 |

**❌ 已废弃的方法**:
```go
// 不要使用 Invoke (无超时控制)
result, errStr := h.app.Invoke(h.thisModule, "auth", "GetUser", rpcReqBytes)

// 不要使用 RpcInvoke (间歇性 "none available" 错误)
result, errStr := h.app.RpcInvoke(...)
```

### 4. 服务注册配置 ⭐ 重要

**参考**: [mqant 官方文档 - 服务注册](https://liangdas.github.io/mqant/server_introduce.html)

```go
// ✅ 在每个 Module 的 OnInit 中配置
m.BaseModule.OnInit(m, app, settings,
    server.RegisterInterval(15*time.Second),
    server.RegisterTTL(30*time.Second),  // 必须 > 心跳
)

// ❌ 不要在 main.go 全局配置 (会导致 RPC 不稳定)
```

---

## 🗄️ 数据库开发规范

### SQLBoiler 多 Schema 配置

```
sqlboiler.auth.toml         → internal/entity/auth/
sqlboiler.game_config.toml  → internal/entity/game_config/
sqlboiler.game_runtime.toml → internal/entity/game_runtime/
```

**使用示例**:
```go
import (
    authModels "tsu-self/internal/entity/auth"
    configModels "tsu-self/internal/entity/game_config"
)

user, _ := authModels.Users().One(ctx, db)
// SELECT * FROM "auth"."users" ...

skill, _ := configModels.Skills(
    configModels.SkillWhere.SkillCode.EQ("FIREBALL"),
).One(ctx, db)
```

### 迁移文件规范

```
migrations/
├── 000001_create_schemas.up.sql              # Schema 和用户
├── 000002_create_core_infrastructure.up.sql  # 枚举、触发器
├── 000003_create_users_system.up.sql         # auth schema
└── {version}_{action}_{object}.{up|down}.sql
```

**黄金规则**:
1. 一个迁移 = 一个原子变更
2. 只包含 DDL，不包含数据 (种子数据另建迁移)
3. 部署后不可修改

### 开发工作流

```bash
# 代码生成
make proto           # 生成 Protobuf
make generate-entity # 生成 SQLBoiler ORM
make generate        # 一键生成所有

# 数据库迁移
make migrate-create  # 创建迁移文件
make migrate-up      # 应用迁移
make migrate-down    # 回滚迁移

# Swagger 文档
make swagger-admin   # 生成 Admin API 文档
```

---

## 🚀 Game Server 架构设计

### 多 Server 部署策略

**架构决策**: ✅ **game-server 启动独立的 Auth Module 实例**

**服务部署拓扑**:
```
admin-server (进程1):
├─ Auth Module (实例1)
└─ Admin Module

game-server (进程2):
├─ Auth Module (实例2)  ← 独立实例
└─ Game Module

Consul 服务注册:
- auth: 2个实例 (自动负载均衡)
- admin: 1个实例
- game: 1个实例
```

**优势**:

| 维度 | 说明 |
|------|------|
| **高可用性** | admin-server 挂了，game-server 的认证仍可用 |
| **性能隔离** | Admin 的认证高峰不影响 Game 玩家登录 |
| **本地 RPC** | Game → Auth 调用在同进程内，延迟更低 |
| **独立扩容** | game-server 可独立水平扩展 |
| **负载均衡** | mqant 自动 Round-Robin 分发 RPC 请求 |

**mqant RPC 机制**:
```go
// mqant 通过 Module Type 标识服务
func (m *AuthModule) GetType() string {
    return "auth"  // ← 所有 Auth 实例共享此 Type
}

// RPC 调用时自动负载均衡
result, _ := m.app.Invoke(m, "auth", "GetUser", reqBytes)
// ↑ Consul 发现所有 "auth" 实例，自动选择一个
```

**关键点**:
- ✅ 服务是逻辑概念，可以有多个物理实例
- ✅ mqant 自动处理服务发现和负载均衡
- ✅ 配置相同的环境变量 (共享 Kratos/Keto/DB)

---

## 🌐 WebSocket 实时通信架构

### 请求流程对比

**HTTP REST API 流程**:
```
Client → Nginx → Oathkeeper (验证) → Game HTTP Handler
         ↑ 每次请求都验证 Session
```

**WebSocket 流程**:
```
Client → Nginx → Oathkeeper (握手验证) → Game WS Handler
         ↑ 只在连接建立时验证一次
         ↓ 连接后透传消息
Client ←────────────────────────────────→ Game
         (Game 内部定期检查 Session 过期)
```

### Oathkeeper WebSocket 支持 ⭐

**重要发现**: Oathkeeper **支持 WebSocket 代理**

**官方文档**: https://www.ory.sh/docs/oathkeeper/guides/proxy-websockets

**限制**:
> "WebSockets bypass Ory Oathkeeper after the first request"

**这意味着**:
- ✅ Oathkeeper 在 WebSocket 握手时验证 Session
- ✅ 连接建立后，消息直接透传到后端
- ⚠️ 后续消息不会再验证，需要 Game Module 自己检查

### WebSocket 架构设计

**统一入口模式** (推荐):

```
internal/modules/game/
├── handler/
│   ├── websocket/
│   │   ├── connection_manager.go   # 连接管理器
│   │   ├── session_checker.go      # Session 定期验证
│   │   ├── message_router.go       # 消息路由
│   │   ├── battle_handler.go       # 战斗事件
│   │   ├── chat_handler.go         # 聊天
│   │   └── team_handler.go         # 组队
│   └── http/
│       ├── hero_handler.go         # 英雄 REST API
│       └── inventory_handler.go
```

**Nginx 配置**:
```nginx
# WebSocket 路由 (与 REST API 一样走 Oathkeeper)
location /ws/ {
    proxy_pass http://tsu_oathkeeper:4456;

    # WebSocket 必需配置
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;

    # 长连接超时
    proxy_read_timeout 3600s;
}
```

**Oathkeeper 规则**:
```yaml
# oathkeeper/rules/websocket.yml
- id: "ws:game"
  upstream:
    url: "http://tsu_game:8072"
  match:
    url: "http://<.*>/ws/game<**>"
    methods:
      - GET
  authenticators:
    - handler: cookie_session  # 验证 Kratos session
  authorizer:
    handler: allow
  mutators:
    - handler: noop
```

**Game Module WebSocket Handler**:
```go
// connection_manager.go
type ConnectionManager struct {
    clients        sync.Map  // userID -> *Client
    sessionChecker *SessionChecker
}

func (m *ConnectionManager) HandleWebSocket(c echo.Context) error {
    // 1. Oathkeeper 已验证，从 header 获取 userID
    userID := c.Request().Header.Get("X-User-ID")

    // 2. 获取 session token (用于后续验证)
    cookie, _ := c.Cookie("ory_kratos_session")
    sessionToken := cookie.Value

    // 3. 升级连接
    conn, _ := m.upgrader.Upgrade(c.Response(), c.Request(), nil)

    // 4. 创建客户端并注册
    client := &Client{
        userID:       userID,
        sessionToken: sessionToken,
        conn:         conn,
        send:         make(chan []byte, 256),
    }

    m.clients.Store(userID, client)
    defer m.clients.Delete(userID)

    // 5. 启动读写协程
    go client.writePump()
    client.readPump()

    return nil
}

// session_checker.go - 定期检查 Session 过期
func (s *SessionChecker) StartSessionCheck() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        s.manager.clients.Range(func(key, value interface{}) bool {
            client := value.(*Client)

            // 验证 Session 是否仍有效
            _, err := s.kratosClient.ValidateSession(ctx, client.sessionToken)
            if err != nil {
                // Session 过期，关闭连接
                client.conn.WriteMessage(websocket.CloseMessage,
                    websocket.FormatCloseMessage(4401, "Session expired"))
                client.conn.Close()
                s.manager.clients.Delete(key)
            }

            return true
        })
    }
}
```

**消息格式**:
```go
// models/ws_message.go
type WSMessage struct {
    Type      string      `json:"type"`      // "battle.action", "chat.send"
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
}

// 战斗动作
type BattleActionData struct {
    BattleID string `json:"battle_id"`
    ActionID string `json:"action_id"`
    TargetID string `json:"target_id"`
}
```

**客户端心跳**:
```javascript
// 前端每30秒发送心跳
setInterval(() => {
    ws.send(JSON.stringify({
        type: 'ping',
        timestamp: Date.now()
    }));
}, 30000);
```

### REST vs WebSocket 使用场景

**使用 REST API**:
- ✅ CRUD 操作 (创建英雄、查询背包)
- ✅ 数据查询 (排行榜、成就列表)
- ✅ 配置更新 (设置、偏好)

**使用 WebSocket**:
- ✅ 战斗实时事件
- ✅ 聊天消息
- ✅ 组队邀请/通知
- ✅ 服务器推送 (系统公告)

**关键特点**:
- 回合制游戏可容忍 100-500ms 延迟
- 使用事件驱动 (玩家操作 → 服务器计算 → 推送结果)
- 不需要帧同步 (像 MOBA/FPS)

---

## 🔒 认证与权限系统

### Kratos 认证架构

**职责分离**:
- **Kratos**: 身份认证 (密码、Session、登录流程)
- **业务 DB (auth.users)**: 用户元数据 (nickname, avatar, is_banned 等)

**核心流程**:
```
Login: KratosClient.LoginWithPassword()
  → CreateNativeLoginFlow()
  → UpdateLoginFlow(credentials)
  → 返回 Session Token

Logout: KratosClient.RevokeSession()
  → ValidateSession(token)
  → DisableSession(sessionID)
```

**API 示例**:
```bash
# 登录 (支持 email/username/phone)
POST /api/v1/auth/login
{"identifier":"user@example.com","password":"xxx"}

# 登出
POST /api/v1/auth/logout
X-Session-Token: ory_st_xxx
```

### Keto 权限架构 (RBAC)

**设计理念**:
- **数据库**: 存储角色/权限**元数据** (用于管理界面)
- **Keto**: 存储用户-角色-权限**关系** (用于运行时检查)

**Relation Tuples 设计**:
```
# 用户-角色
namespace: roles
object: admin
relation: member
subject_id: users:alice

# 角色-权限 (SubjectSet)
namespace: permissions
object: user:create
relation: granted
subject_set: {namespace:roles, object:admin, relation:member}
```

**核心方法**:
```go
ketoClient.AssignRoleToUser(ctx, userID, roleCode)
ketoClient.CheckUserPermission(ctx, userID, permissionCode)
ketoClient.GetUserRoles(ctx, userID)
```

---

## 🎮 游戏配置系统 (Admin Module)

### 已实现功能 (24 个 Repository)

**基础配置** (5):
- SkillCategories, ActionCategories, DamageTypes
- HeroAttributeType, Tags + TagsRelations

**元数据表** (4, 只读):
- EffectTypeDefinitions, FormulaVariables
- RangeConfigRules, ActionTypeDefinitions

**职业系统** (2):
- Classes (CRUD + 软删除)
- ClassAttributeBonuses (一对多关联)

**技能系统** (2):
- Skills, SkillLevelConfigs

**效果和 Buff** (4):
- Effects, Buffs, BuffEffects
- ActionFlags

**动作系统** (3):
- Actions, ActionEffects, SkillUnlockActions

**实现模式** (每个功能):
```
Repository Interface (interfaces/*.go)
     ↓
Repository Impl (impl/*_impl.go)
     ↓
Service (modules/admin/service/*.go)
     ↓
Handler (modules/admin/handler/*.go)
     ↓
注册到 admin_module.go
```

### 技能系统设计理念 ⭐

**原子效果组合模式** (类似 Unreal GAS):
```
Skill → unlocks → Action → composed of → Effects (原子效果)
                                       ↓
                                    Buffs
```

**优点**:
- ✅ 高度可复用 (一个"造成伤害"效果用于多个技能)
- ✅ 策划自主 (组合现有效果创建新技能)
- ✅ 配置驱动 (元数据表定义规范)

**JSONB 灵活参数**:
```sql
effects.parameters JSONB           -- 效果参数
actions.range_config JSONB         -- 射程配置
actions.target_config JSONB        -- 目标选择
buffs.parameter_definitions JSONB -- Buff 参数
```

**注意**: 应用层必须严格验证 JSONB 结构！

---

## 🛠️ 错误处理与响应

### xerrors 错误码体系

```
1xxxxx: 通用系统错误
2xxxxx: 认证相关 (CodeAuthenticationFailed, CodeTokenExpired)
3xxxxx: 权限相关 (CodePermissionDenied)
4xxxxx: 用户管理 (CodeUserNotFound, CodeUserBanned)
5xxxxx: 角色权限 (CodeRoleNotFound)
6xxxxx: 业务逻辑
7xxxxx: 外部服务
8xxxxx: 游戏业务
  80xxxx: 角色相关
  81xxxx: 技能相关
  82xxxx: 职业相关
```

### response 响应处理

```go
// Echo 适配器
return response.EchoOK(c, h.respWriter, data)
return response.EchoError(c, h.respWriter, xerrors.NewUserNotFoundError(id))
return response.EchoBadRequest(c, h.respWriter, "参数错误")

// 统一响应格式
{
  "code": 100000,
  "message": "操作成功",
  "data": {...},
  "timestamp": 1759501201,
  "trace_id": "..."
}
```

---

## 🔧 常见问题排查

### 1. Module panic: nil pointer

```go
// 改为值类型嵌入
type AuthModule struct {
    basemodule.BaseModule  // 不是 *basemodule.BaseModule
}
```

### 2. RPC "params not adapted"

```go
// 移除 context.Context 参数
func (h *RPCHandler) Method(req []byte) ([]byte, error) {
    ctx := context.Background()
    // ...
}
```

### 3. RPC 间歇性 "none available" ⭐

**原因**: 服务注册配置不当，Consul 误判下线

**解决**: 在每个 Module 的 OnInit 配置 (不是 main.go):
```go
m.BaseModule.OnInit(m, app, settings,
    server.RegisterInterval(15*time.Second),
    server.RegisterTTL(30*time.Second),
)
```

**诊断**:
```bash
# 查看 Consul 服务
curl http://localhost:8500/v1/catalog/services | jq

# 服务健康状态
curl http://localhost:8500/v1/health/service/auth | jq
```

### 4. SQLBoiler 类型注意事项

```go
// ⚠️ 复数形式拼写
*game_config.ClassAttributeBonuse  // 不是 Bonus!

// Decimal 类型处理
if err := bonus.BaseBonusValue.UnmarshalText([]byte("2.5")); err != nil {
    // 处理错误
}

// NullDecimal 判断
if !bonus.DamageMultiplier.IsZero() {
    // 有值
}
```

---

## 📚 Make 命令速查

| 命令 | 说明 |
|------|------|
| `make dev-up` | 启动开发环境 |
| `make proto` | 生成 Protobuf 代码 |
| `make generate-entity` | 生成 SQLBoiler ORM |
| `make generate` | 一键生成所有 |
| `make migrate-up` | 应用数据库迁移 |
| `make migrate-create` | 创建新迁移文件 |
| `make swagger-admin` | 生成 Swagger 文档 |
| `make clean` | 清理环境 |

---

## 📖 参考文档

- mqant 官方文档: https://liangdas.github.io/mqant/
- Ory Kratos: https://www.ory.sh/docs/kratos/
- Ory Keto: https://www.ory.sh/docs/keto/
- SQLBoiler: https://github.com/volatiletech/sqlboiler
- Echo Framework: https://echo.labstack.com/

---

---

## 📝 架构决策记录

### ADR-001: game-server 启动独立 Auth Module

**日期**: 2025-10-10

**状态**: ✅ 已采纳

**决策**: game-server 启动独立的 Auth Module 实例，而不是复用 admin-server 的 Auth

**理由**:
1. **高可用性**: admin-server 故障不影响游戏玩家登录
2. **性能隔离**: 运营后台与游戏服务的认证流量完全隔离
3. **本地 RPC 优化**: Game → Auth 调用在同进程内，延迟更低
4. **mqant 天然支持**: 自动负载均衡，无需额外代码
5. **独立扩容**: 游戏服务可独立水平扩展，Auth 实例随之扩展

**替代方案**: 共用 admin-server 的 Auth (单点故障风险高)

### ADR-002: WebSocket 走 Oathkeeper 认证

**日期**: 2025-10-10

**状态**: ✅ 已采纳

**决策**: WebSocket 连接通过 Oathkeeper 进行握手时认证，连接后由 Game Module 定期检查 Session

**理由**:
1. **统一认证入口**: 所有请求 (REST + WebSocket) 都走 Oathkeeper
2. **Oathkeeper 原生支持**: 官方文档确认支持 WebSocket 代理
3. **架构一致性**: Nginx 配置统一，无特殊路由
4. **安全性**: 握手时验证 + 定期 Session 检查

**限制**: Oathkeeper 只在握手时验证一次，需要 Game Module 实现 Session 定期检查

**替代方案**: WebSocket 绕过 Oathkeeper (需要在 Game Module 完全实现认证逻辑)

### ADR-003: 统一 WebSocket 入口 (Connection Manager)

**日期**: 2025-10-10

**状态**: ✅ 已采纳

**决策**: 使用单一 WebSocket 端点 `/ws/game`，通过消息类型路由到不同业务 Handler

**理由**:
1. **客户端简单**: 只需建立一个 WebSocket 连接
2. **连接复用**: 战斗、聊天、组队共用一个连接
3. **认证开销低**: 只需在建立连接时认证一次
4. **符合回合制特点**: 不需要极致性能，架构清晰更重要

**替代方案**: 多个 WebSocket 端点 (客户端管理复杂)

---

**最后更新**: 2025-10-10
