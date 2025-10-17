# Admin Server 与 Game Server 独立性分析

## 📊 独立性评估总览

| 维度 | 状态 | 说明 |
|---|---|---|
| **进程隔离** | ✅ 完全独立 | 两个独立的可执行文件 |
| **路由隔离** | ✅ 完全独立 | `/api/v1/admin/*` vs `/api/v1/game/*` |
| **端口隔离** | ✅ 完全独立 | 8071 (admin) vs 8072 (game) |
| **数据库连接** | ✅ 完全独立 | 不同的环境变量 |
| **Schema 隔离** | ✅ 架构设计完善 | admin/game_config/game_runtime/auth |
| **Swagger 文档** | ✅ 完全独立 | 独立生成和托管 |
| **Auth 模块** | ⚠️ 共享实例 | 两个 server 都启动独立的 Auth 实例 |
| **代码依赖** | ✅ 无直接依赖 | 无跨模块 import |
| **配置隔离** | ⚠️ 有小问题 | game-server.json 注释有误 |

---

## 1️⃣ 进程隔离 ✅

### 独立的可执行文件
```
cmd/admin-server/main.go  → 编译为 admin-server
cmd/game-server/main.go   → 编译为 game-server
```

### 独立的启动流程
**Admin Server**:
```go
app.Run(
    auth.Module(),   // Admin 专用的 Auth 实例
    admin.Module(),  // Admin 业务模块
)
```

**Game Server**:
```go
app.Run(
    auth.Module(),   // Game 专用的 Auth 实例（高可用设计）
    game.Module(),   // Game 业务模块
)
```

**评估**：✅ 完全独立
- 两个服务可以独立部署、独立重启
- Admin 崩溃不影响 Game，反之亦然

---

## 2️⃣ 路由隔离 ✅

### Admin Server 路由
```
/api/v1/admin/auth/*              # 认证（注册、登录）
/api/v1/admin/users/*             # 用户管理
/api/v1/admin/permissions/*       # 权限管理
/api/v1/admin/classes/*           # 职业配置管理
/api/v1/admin/skills/*            # 技能配置管理
/api/v1/admin/actions/*           # 动作配置管理
/api/v1/admin/effects/*           # 效果配置管理
...（约60+个配置管理接口）
```

### Game Server 路由
```
/api/v1/game/auth/*               # 认证（注册、登录）
/api/v1/game/heroes/*             # 英雄管理
/api/v1/game/classes/*            # 职业查询（只读）
/api/v1/game/skills/*             # 技能查询（只读）
/api/v1/game/hero-level-requirements/*  # 配置查询
/api/v1/game/skill-upgrade-costs/*      # 配置查询
/api/v1/game/attribute-upgrade-costs/*  # 配置查询
...（40个玩家接口）
```

**评估**：✅ 完全独立
- 路由前缀明确区分：`/admin` vs `/game`
- 无路由冲突
- Nginx 可以基于前缀路由到不同的服务

---

## 3️⃣ 端口隔离 ✅

### 端口配置

| 服务 | HTTP 端口 | Auth 端口 | Swagger 地址 |
|---|---|---|---|
| Admin Server | 8071 | 8072 | http://localhost:8071/swagger/ |
| Game Server | 8061 | 8062 | http://localhost:8072/swagger/ |

**注意**：
- Admin 的 HTTP 端口是 8071，Auth 实例端口是 8072
- Game 的 HTTP 端口是 8061，Auth 实例端口是 8062
- 通过 Nginx 反向代理后，外部统一通过 80/443 端口访问

**评估**：✅ 完全独立
- 端口无冲突
- 可以在同一台机器上运行

---

## 4️⃣ 数据库连接隔离 ✅

### 环境变量
```bash
# Admin Server
TSU_ADMIN_DATABASE_URL="postgresql://tsu_admin:password@localhost:5432/tsu_game?sslmode=disable&search_path=admin,game_config,public"

# Game Server
TSU_GAME_DATABASE_URL="postgresql://tsu_game:password@localhost:5432/tsu_game?sslmode=disable&search_path=game_runtime,game_config,auth,public"

# Auth Module
TSU_AUTH_DATABASE_URL="postgresql://tsu_auth:password@localhost:5432/tsu_game?sslmode=disable&search_path=auth,public"
```

### 代码实现
**Admin Module**:
```go
// internal/modules/admin/admin_module.go:115
dbURL := os.Getenv("TSU_ADMIN_DATABASE_URL")
```

**Game Module**:
```go
// internal/modules/game/game_module.go:102
dbURL := os.Getenv("TSU_GAME_DATABASE_URL")
```

**评估**：✅ 完全独立
- 使用不同的环境变量
- 可以配置不同的数据库用户
- 可以配置不同的 `search_path`

---

## 5️⃣ Schema 隔离 ✅

### Schema 职责划分

```
PostgreSQL: tsu_game
├─ auth             # 用户认证数据（Auth 模块拥有）
│  ├─ users         # 用户表
│  ├─ sessions      # 会话表
│  └─ ...
├─ game_config      # 游戏配置数据（Admin 管理）
│  ├─ classes       # 职业配置
│  ├─ skills        # 技能配置
│  ├─ actions       # 动作配置
│  ├─ effects       # 效果配置
│  └─ ...
├─ game_runtime     # 运行时数据（Game 管理）
│  ├─ hero_base     # 英雄数据
│  ├─ hero_attributes  # 英雄属性
│  ├─ hero_skills   # 英雄技能
│  └─ ...
└─ admin            # 后台数据（Admin 专用）
   └─ ...（待实现）
```

### 访问权限矩阵

| 模块 | auth.* | game_config.* | game_runtime.* | admin.* |
|---|---|---|---|---|
| **Auth** | ✅ 读写 | ❌ 无权限 | ❌ 无权限 | ❌ 无权限 |
| **Admin** | 👁️ 只读 | ✅ 读写 | 👁️ 只读（查询）| ✅ 读写 |
| **Game** | 👁️ 只读 | 👁️ 只读 | ✅ 读写 | ❌ 无权限 |

### 跨 Schema 访问规则

**✅ 允许的操作**：
```go
// Game 模块读取 game_config（查询职业、技能配置）
db.QueryRow("SELECT * FROM game_config.classes WHERE id = $1", classID)

// Admin 模块读取 auth（查看用户列表）
db.QueryRow("SELECT username FROM auth.users WHERE id = $1", userID)
```

**❌ 禁止的操作**（必须通过 RPC）：
```go
// ❌ Game 模块不能直接修改 auth
db.Exec("UPDATE auth.users SET is_banned = true WHERE id = $1", userID)

// ✅ 应该通过 RPC 调用 Auth 模块
result, _ := app.Invoke(gameModule, "auth", "BanUser", reqBytes)
```

**评估**：✅ 架构设计完善
- Schema 职责清晰
- 通过 PostgreSQL schema 和用户权限实现物理隔离
- 跨 schema 修改强制使用 RPC

---

## 6️⃣ Swagger 文档隔离 ✅

### 文档生成
```bash
# Admin Swagger
swag init --generalInfo cmd/admin-server/main.go --output docs/admin

# Game Swagger
swag init --generalInfo cmd/game-server/main.go --output docs/game
```

### 文档托管
```
Admin: http://localhost:8071/swagger/index.html
Game:  http://localhost:8072/swagger/index.html
```

### 文档内容
- **Admin Swagger**：60+ 配置管理接口
- **Game Swagger**：40 玩家接口

**评估**：✅ 完全独立
- 独立的 Swagger 文件
- 独立的 UI 托管
- 无交叉引用

---

## 7️⃣ Auth 模块共享 ⚠️

### 当前设计

两个 Server 都启动独立的 Auth 模块实例：

```go
// cmd/admin-server/main.go
app.Run(
    auth.Module(),   // Admin 的 Auth 实例
    admin.Module(),
)

// cmd/game-server/main.go
app.Run(
    auth.Module(),   // Game 的 Auth 实例
    game.Module(),
)
```

### 特点

**优点**：
- ✅ **高可用性**：Admin 崩溃不影响玩家登录
- ✅ **独立扩展**：可以为 Game 配置更多 Auth 实例
- ✅ **服务发现共享**：通过 Consul 自动负载均衡

**潜在问题**：
- ⚠️ **数据一致性**：两个实例操作同一个 `auth` schema
- ⚠️ **Session 管理**：需要共享 Redis 保证 session 一致性
- ⚠️ **资源占用**：每个实例都占用端口和内存

### 解决方案

**当前架构已解决**：
```
1. 共享 Redis（session 存储）
   ├─ admin-server 的 auth → redis:6379
   └─ game-server 的 auth → redis:6379

2. 共享数据库（用户数据）
   ├─ admin-server 的 auth → TSU_AUTH_DATABASE_URL
   └─ game-server 的 auth → TSU_AUTH_DATABASE_URL

3. 通过 Consul 服务发现
   ├─ auth@admin-server → Consul
   └─ auth@game-server → Consul
```

**评估**：⚠️ 设计合理，但需注意一致性
- 当前通过共享存储（Redis + PostgreSQL）保证一致性
- 建议：未来可以考虑独立 Auth Service

---

## 8️⃣ 代码依赖 ✅

### 依赖分析

**Admin 模块**：
```go
import (
    "tsu-self/internal/modules/admin/handler"
    "tsu-self/internal/pkg/..."
    // ✅ 无 game 相关 import
)
```

**Game 模块**：
```go
import (
    "tsu-self/internal/modules/game/handler"
    "tsu-self/internal/modules/game/service"
    "tsu-self/internal/pkg/..."
    // ✅ 无 admin 相关 import
)
```

**共享代码**：
```
internal/pkg/          # 工具包（response、log、metrics）
internal/repository/   # 数据访问层（两者可能共享部分 repository）
internal/entity/       # SQLBoiler 生成的实体（两者都使用）
internal/middleware/   # 中间件（auth、logging）
```

**评估**：✅ 无直接依赖
- Admin 和 Game 模块之间无直接 import
- 只共享基础设施代码（`internal/pkg/`）
- 数据层通过 repository 接口隔离

---

## 9️⃣ 配置文件 ⚠️

### 配置隔离

| 配置文件 | 用途 |
|---|---|
| `configs/server/admin-server.json` | Admin + Auth 配置 |
| `configs/server/game-server.json` | Game + Auth 配置 |

### 发现的问题

**game-server.json 第17行**：
```json
"_comment_database_url": "从环境变量 TSU_ADMIN_DATABASE_URL 读取..."
```

❌ **错误**：应该是 `TSU_GAME_DATABASE_URL`

**实际代码是正确的**：
```go
// internal/modules/game/game_module.go:102
dbURL := os.Getenv("TSU_GAME_DATABASE_URL")  // ✅ 正确
```

**评估**：⚠️ 注释错误，但不影响运行
- 建议：修正配置文件注释

---

## 🎯 总结与建议

### ✅ 独立性评估：优秀（9/10）

| 评分项 | 得分 |
|---|---|
| 进程隔离 | ⭐⭐⭐⭐⭐ |
| 路由隔离 | ⭐⭐⭐⭐⭐ |
| 端口隔离 | ⭐⭐⭐⭐⭐ |
| 数据库连接隔离 | ⭐⭐⭐⭐⭐ |
| Schema 隔离 | ⭐⭐⭐⭐⭐ |
| Swagger 文档隔离 | ⭐⭐⭐⭐⭐ |
| Auth 模块设计 | ⭐⭐⭐⭐ |
| 代码依赖 | ⭐⭐⭐⭐⭐ |
| 配置管理 | ⭐⭐⭐⭐ |

**总分**：45/50 ⭐

---

### 🎉 做得好的地方

1. ✅ **完全独立的可执行文件**
   - 可以独立部署、独立扩展
   - 崩溃互不影响

2. ✅ **清晰的路由前缀**
   - `/api/v1/admin/*` vs `/api/v1/game/*`
   - Nginx 路由配置简单

3. ✅ **Schema 职责明确**
   - `game_config`（Admin 写，Game 读）
   - `game_runtime`（Game 写，Admin 读）
   - 物理隔离 + 权限控制

4. ✅ **独立的 Swagger 文档**
   - 方便前端团队分别对接

5. ✅ **高可用的 Auth 设计**
   - 每个 Server 都有独立的 Auth 实例
   - Admin 故障不影响玩家登录

---

### ⚠️ 需要注意的地方

#### 1. Auth 模块的共享依赖

**现状**：
- 两个 Auth 实例共享 Redis（session）
- 两个 Auth 实例共享数据库（用户表）

**风险**：
- Redis 故障会同时影响两个服务
- 数据库连接池竞争

**建议**：
```
方案1：维持现状
  ✓ 实现简单
  ✓ 资源利用率高
  ✗ 单点故障风险

方案2：独立 Auth Service（推荐）
  ✓ 真正的服务解耦
  ✓ 可独立扩展
  ✗ 增加部署复杂度
```

#### 2. Repository 层共享

**现状**：
```
internal/repository/
├─ impl/
│  ├─ class_repository_impl.go       # Admin 写，Game 读
│  ├─ skill_repository_impl.go       # Admin 写，Game 读
│  ├─ hero_repository_impl.go        # Game 专用
│  └─ ...
```

**风险**：
- Admin 和 Game 共享同一个 repository 实现
- 修改 repository 可能同时影响两个服务

**建议**：
- ✅ 当前设计合理（避免重复代码）
- ⚠️ 注意在修改时的影响范围
- 💡 可以考虑按 schema 拆分：
  ```
  internal/repository/
  ├─ config/          # game_config 相关（Admin 写，Game 读）
  ├─ runtime/         # game_runtime 相关（Game 专用）
  └─ auth/            # auth 相关（Auth 模块）
  ```

#### 3. 配置文件注释错误

**问题**：
```json
// configs/server/game-server.json:17
"_comment_database_url": "从环境变量 TSU_ADMIN_DATABASE_URL 读取..."
```

**修正**：
```json
"_comment_database_url": "从环境变量 TSU_GAME_DATABASE_URL 读取..."
```

---

## 📋 部署独立性验证清单

### ✅ 可以独立部署
- [x] Admin Server 可以单独启动
- [x] Game Server 可以单独启动
- [x] Admin 崩溃不影响 Game
- [x] Game 崩溃不影响 Admin

### ✅ 可以独立扩展
- [x] Admin 可以单独扩容（增加实例）
- [x] Game 可以单独扩容（增加实例）
- [x] 通过 Consul 自动负载均衡

### ✅ 可以独立升级
- [x] 修改 Admin 代码不需要重启 Game
- [x] 修改 Game 代码不需要重启 Admin
- [x] 独立的版本号管理

### ⚠️ 共享依赖
- [x] 共享 Redis（Session 存储）
- [x] 共享 PostgreSQL（不同 schema）
- [x] 共享 Consul（服务发现）
- [x] 共享 NATS（消息队列）

---

## 🚀 最佳实践建议

### 1. 保持当前架构
当前设计已经非常优秀，建议：
- ✅ 维持进程隔离
- ✅ 维持路由隔离
- ✅ 维持 Schema 隔离
- ✅ 维持 Auth 双实例（高可用）

### 2. 可选优化
如果未来遇到以下问题，考虑优化：

**问题1：Auth 模块成为瓶颈**
```
解决方案：独立 Auth Service
├─ 独立部署 auth-server
├─ admin-server 和 game-server 通过 RPC 调用
└─ 真正的服务解耦
```

**问题2：Repository 修改影响范围大**
```
解决方案：按 Schema 拆分 Repository
├─ internal/repository/config/    # game_config
├─ internal/repository/runtime/   # game_runtime
└─ internal/repository/auth/      # auth
```

### 3. 监控建议
```
监控指标：
├─ Admin Server 健康状态
├─ Game Server 健康状态
├─ Auth 实例健康状态
├─ Redis 连接池
├─ PostgreSQL 连接池
└─ RPC 调用延迟
```

---

## 🎬 结论

**Admin Server 和 Game Server 已经做到了很好的独立性**：

✅ **架构层面**：进程、路由、端口、数据库连接完全独立  
✅ **数据层面**：Schema 隔离，访问权限清晰  
✅ **代码层面**：无直接依赖，只共享基础工具  
✅ **部署层面**：可独立部署、扩展、升级  
⚠️ **共享依赖**：Redis、PostgreSQL、Consul（这是微服务架构的常见设计）

**总评**：⭐⭐⭐⭐⭐ (9/10)

当前架构足以支持：
- 独立团队开发
- 独立发布上线
- 独立故障隔离
- 独立性能优化

---

*分析时间：2025-10-17*  
*架构版本：v1.0*

