# Game Module 英雄系统代码审查报告

**审查时间**：2025-10-16
**审查范围**：Game Module 完整实现
**审查内容**：数据库设计、Service 层架构、Repository 层、Handler 层、包结构、并发安全

---

## 📊 执行摘要

经过全面审查，发现了 **12 个主要问题**，其中：
- 🔴 **P0（关键）**：3 个 - 需要立即修复，阻塞核心功能
- ⚠️ **P1（重要）**：5 个 - 影响用户体验和系统稳定性
- 💡 **P2（优化）**：4 个 - 长期改进建议

---

## 🔴 P0 级问题（必须立即修复）

### 1. 经验系统逻辑严重错误 ⚠️ 🔥

**位置**：
- `internal/modules/game/service/hero_attribute_service.go:104`
- `internal/modules/game/service/hero_skill_service.go:111`

**问题代码**：
```go
hero.ExperienceAvailable -= int64(totalCost)
hero.ExperienceSpent += int64(totalCost)
hero.ExperienceTotal += int64(totalCost)  // ❌ 这行错误！
```

**问题描述**：
- `experience_total` 应该是**累计获得的总经验**（来自怪物掉落、任务奖励等）
- 当前实现在每次消费经验时**凭空增加**经验总量
- 导致：玩家花越多经验，经验越多 → **无限升级漏洞**

**影响**：
- 经验计算完全错误
- `AutoLevelUp` 基于错误的数据无法正常工作
- 玩家可以通过加点无限升级

**修复**：
```go
// 删除这一行：
// hero.ExperienceTotal += int64(totalCost)

// 正确的逻辑：
// experience_total = experience_available + experience_spent
```

**同样的错误也存在于回退逻辑**（`hero_attribute_service.go:186`、`hero_skill_service.go:339`）

---

### 2. 属性初始化完全缺失 🔴

**位置**：`internal/modules/game/service/hero_service.go:150-155`

**当前代码**：
```go
func (s *HeroService) initializeAllocatedAttributes(ctx context.Context) ([]byte, error) {
    // TODO: 需要实现 HeroAttributeTypeRepository.ListAll 方法
    // 临时返回空的 JSONB
    allocatedAttrs := make(map[string]map[string]interface{})
    return json.Marshal(allocatedAttrs)
}
```

**问题**：
- 创建英雄时 `allocated_attributes` 为空 `{}`
- 随后调用 `AllocateAttribute` 时会在第 81 行报错："属性不存在"
- **结果：玩家无法为任何属性加点**

**影响**：🔴 **阻塞核心游戏功能**

**修复**：
```go
func (s *HeroService) initializeAllocatedAttributes(ctx context.Context) ([]byte, error) {
    // 查询所有 basic 类型的属性
    attrs, err := s.heroAttributeTypeRepo.ListByCategory(ctx, "basic")
    if err != nil {
        return nil, xerrors.Wrap(err, xerrors.CodeInternalError, "查询属性类型失败")
    }

    allocatedAttrs := make(map[string]map[string]interface{})
    for _, attr := range attrs {
        if attr.IsActive.Bool {  // 只初始化活跃属性
            allocatedAttrs[attr.AttributeCode] = map[string]interface{}{
                "value":    1,   // 初始值为 1（按需求）
                "spent_xp": 0,   // 初始消耗为 0
            }
        }
    }

    return json.Marshal(allocatedAttrs)
}
```

**需要实现**：
- `HeroAttributeTypeRepository.ListByCategory(ctx, category)` 方法

---

### 3. 技能池验证缺失（安全漏洞） 🔓

**位置**：
- `internal/modules/game/service/hero_skill_service.go:70-74`（LearnSkill）
- `internal/modules/game/service/hero_skill_service.go:205-209`（UpgradeSkill）

**当前代码**：
```go
// TODO: 需要实现 GetByClassIDsAndSkillID 方法
// skillPool, err := s.classSkillPoolRepo.GetByClassIDsAndSkillID(ctx, classIDs, req.SkillID)
// if err != nil || skillPool == nil {
//     return xerrors.New(xerrors.CodeSkillNotFound, "技能不在可学习池中")
// }
```

**问题**：
- **任何玩家都可以学习任何技能**（只要知道 skill_id）
- 无职业限制（战士可以学法师技能）
- 技能等级上限验证也被跳过

**影响**：🔓 **安全漏洞 - 破坏游戏平衡**

**修复**：
```go
// 在 LearnSkill 中
// 获取当前职业
currentClass, err := s.heroClassHistoryRepo.GetCurrentClass(ctx, req.HeroID)
if err != nil {
    return xerrors.Wrap(err, xerrors.CodeResourceNotFound, "职业信息不存在")
}

// 验证技能是否在职业技能池中
skillPool, err := s.classSkillPoolRepo.GetByClassIDAndSkillID(ctx, currentClass.ClassID, req.SkillID)
if err != nil || skillPool == nil {
    return xerrors.New(xerrors.CodeSkillNotFound, "该职业无法学习此技能")
}

// 验证技能等级上限
maxLearnableLevel := int(skillPool.MaxLearnableLevel)
if maxLearnableLevel <= 0 {
    return xerrors.New(xerrors.CodeInvalidParams, "该技能无法学习")
}
```

**需要实现**：
- `ClassSkillPoolRepository.GetByClassIDAndSkillID()`
- `HeroClassHistoryRepository.GetCurrentClass()`

---

## ⚠️ P1 级问题（应该尽快修复）

### 4. 技能升级等级数据不一致

**位置**：`internal/modules/game/service/hero_skill_service.go:273`

**问题代码**：
```go
operation := &game_runtime.HeroSkillOperation{
    LevelsAdded: req.Levels,  // ❌ 使用请求值
    // 但实际只升级了 1 级
    LevelBefore: int(oldLevel),
    LevelAfter:  nextLevel,
    // ...
}
```

**问题**：
- 请求中 `req.Levels` 可能是 `5`
- 但实际逻辑只升级了 `1` 级（`nextLevel := int(heroSkill.SkillLevel) + 1`）
- `LevelsAdded` 数据和实际升级等级不符

**影响**：
- 操作历史记录错误
- 回退功能会读到错误的升级等级

**修复**：
```go
// 选项 A：固定升级 1 级（当前实现的真实逻辑）
operation := &game_runtime.HeroSkillOperation{
    LevelsAdded: 1,  // ✅ 实际升级量
    LevelBefore: int(oldLevel),
    LevelAfter:  nextLevel,
    // ...
}

// 选项 B：如果要支持批量升级，需要完整实现
// for i := 0; i < req.Levels && heroSkill.SkillLevel < maxLevel; i++ {
//     // 升级逻辑
// }
```

---

### 5. 循环依赖风险和架构问题

**位置**：
- `internal/modules/game/service/hero_attribute_service.go:26`
- `internal/modules/game/service/hero_skill_service.go:27`

**当前结构**：
```go
// HeroAttributeService 包含 HeroService
type HeroAttributeService struct {
    heroService *HeroService  // 包含引用
}

// HeroSkillService 包含 HeroService
type HeroSkillService struct {
    heroService *HeroService  // 包含引用
}
```

**问题**：
1. **内存浪费**：每个 Service 都创建自己的 Repository 实例（3 个 Service × 8 个 Repo = 24 个实例）
2. **循环依赖风险**：如果 `HeroService` 需要调用 `HeroAttributeService`，就形成循环
3. **事务隔离不佳**：各 Service 创建独立事务，难以协调
4. **难以测试**：无法轻松替换 Repository 为 Mock

**建议**：改用**函数回调**或**事件驱动**模式

**方案 A：函数回调（推荐，最小改动）**
```go
// hero_attribute_service.go
type HeroAttributeService struct {
    db              *sql.DB
    // ... repositories ...
    // ❌ 移除: heroService *HeroService
    autoLevelUpFunc func(context.Context, *sql.Tx, string) error
}

// 在调用 AllocateAttribute 时通过回调
func (s *HeroAttributeService) AllocateAttribute(
    ctx context.Context,
    req *AllocateAttributeRequest,
) error {
    // ... 业务逻辑 ...

    // 提交事务前调用回调
    if s.autoLevelUpFunc != nil {
        if err := s.autoLevelUpFunc(ctx, tx, req.HeroID); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

**方案 B：事件驱动（更解耦，但需要额外基础设施）**
```go
// 定义事件
type HeroXPChangedEvent struct {
    HeroID    string
    XPAdded   int
    Source    string  // "attribute_allocation", "skill_upgrade", etc
}

// Service 发布事件
func (s *HeroAttributeService) AllocateAttribute(...) {
    // ... 逻辑 ...
    eventBus.Publish(HeroXPChangedEvent{
        HeroID: req.HeroID,
        XPAdded: totalCost,
        Source: "attribute_allocation",
    })
}

// HeroService 监听事件
func (s *HeroService) OnXPChanged(event HeroXPChangedEvent) {
    s.AutoLevelUp(ctx, event.HeroID)
}
```

---

### 6. 并发安全问题

**问题 A：JSONB 更新非原子**

**位置**：`hero_attribute_service.go:107-117`

```go
// 读取
var allocatedAttrs map[string]map[string]interface{}
json.Unmarshal(hero.AllocatedAttributes.JSON, &allocatedAttrs)
currentValue := int(allocatedAttrs[req.AttributeCode]["value"].(float64))

// 修改
allocatedAttrs[req.AttributeCode]["value"] = toPoint

// 写回
updatedAttrsJSON, _ := json.Marshal(allocatedAttrs)
hero.AllocatedAttributes.UnmarshalJSON(updatedAttrsJSON)
hero.Update(ctx, tx, hero)
```

**风险场景**：
- 玩家 A 和 B 同时为同一英雄的不同属性加点
- A 读取整个 JSONB：`{STR: 10, DEX: 8}`
- B 读取同一 JSONB：`{STR: 10, DEX: 8}`
- A 修改 STR → 11，写回：`{STR: 11, DEX: 8}`
- B 修改 DEX → 9，写回：`{STR: 10, DEX: 9}`  ← A 的修改被覆盖！

**建议修复**：使用 PostgreSQL 的原子 JSONB 更新
```sql
UPDATE game_runtime.heroes
SET allocated_attributes = jsonb_set(
    allocated_attributes,
    '{STR,value}',
    to_jsonb(
        COALESCE((allocated_attributes->'STR'->>'value')::int, 1) + $1
    )
),
experience_available = experience_available - $2
WHERE id = $3 AND experience_available >= $2
RETURNING *;
```

**问题 B：AutoLevelUp 的多级升级**

**位置**：`hero_service.go:202-228`

```go
// 当前实现一次性升到最高等级
hero.CurrentLevel = int16(targetLevel)
```

**潜在问题**：
- 如果某些等级有特殊事件（如 3 级解锁新技能），会被漏掉
- 没有升级日志，用户体验不好

**建议**：分级升级并触发事件
```go
func (s *HeroService) AutoLevelUp(ctx context.Context, tx *sql.Tx, heroID string) (bool, int, error) {
    // ...
    if targetLevel > int(hero.CurrentLevel) {
        oldLevel := hero.CurrentLevel
        hero.CurrentLevel = int16(targetLevel)

        // 逐级触发升级事件和奖励
        for level := oldLevel + 1; level <= targetLevel; level++ {
            log.Infof("Hero %s leveled up: %d -> %d", heroID, level-1, level)
            // TODO: 触发升级奖励（技能点、属性点等）
            // s.grantLevelRewards(ctx, tx, heroID, int(level))
        }
    }
    // ...
}
```

---

### 7. 属性视图中的数据不安全

**位置**：`migrations/000010_hero_system_runtime_tables.up.sql:94`

```sql
COALESCE((h.allocated_attributes->hat.attribute_code->>'value')::INTEGER, 1) as base_value
```

**问题**：
- 如果 `allocated_attributes` 为空或格式错误，会返回默认值 `1`
- 无法区分"真实的初始值 1"和"数据损坏被填充的 1"

**建议**：
1. 在创建英雄时**强制初始化**所有属性（见 P0 问题 2）
2. 添加视图的 WHERE 条件
```sql
WHERE h.deleted_at IS NULL
  AND hat.is_active = TRUE
  AND hat.category = 'basic'
  AND h.allocated_attributes IS NOT NULL
```

---

### 8. 事务处理的隐藏问题

**位置**：所有 Service 方法的事务处理

**当前模式**：
```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()  // ❌ 不够优雅

// ... 业务逻辑 ...

if err := tx.Commit(); err != nil {
    return err
}
```

**问题**：
- `Commit()` 成功后，`defer tx.Rollback()` 会返回 `sql.ErrTxDone`
- 虽然标准库会忽略，但不符合最佳实践

**修复**：
```go
defer func() {
    if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
        logger.Errorf("transaction rollback failed: %v", err)
    }
}()
```

---

## 💡 P2 级问题（优化建议）

### 9. allocated_attributes 的 JSONB 结构过于复杂

**当前结构**：
```json
{
  "STR": {
    "value": 10,
    "spent_xp": 500
  }
}
```

**问题**：
- 嵌套深、类型不安全
- 每次加点都要解析整个 JSONB
- `spent_xp` 可以通过 `hero_attribute_operations` 表计算得出

**建议简化**：
```json
{
  "STR": 10,
  "DEX": 8
}
```

**或者使用拆表方案**（推荐）：
```sql
CREATE TABLE game_runtime.hero_attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    hero_id UUID NOT NULL REFERENCES heroes(id) ON DELETE CASCADE,
    attribute_code VARCHAR(32) NOT NULL,
    current_value INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(hero_id, attribute_code)
);
```

**优势**：
- 原子更新：`UPDATE hero_attributes SET current_value = current_value + 1 WHERE hero_id = ? AND attribute_code = ?`
- 类型安全、可索引
- 支持数据库级约束

---

### 10. Repository 实例的重复创建

**问题**：每个 Service 创建自己的 Repository 实例导致内存浪费

**建议**：使用依赖注入容器
```go
// 在 game_module.go 的初始化中
type ServiceContainer struct {
    // Repository 单例
    heroRepo                 interfaces.HeroRepository
    classRepo                interfaces.ClassRepository
    classSkillPoolRepo       interfaces.ClassSkillPoolRepository
    // ... 其他 repo ...

    // Service 单例
    heroService             *service.HeroService
    attributeService        *service.HeroAttributeService
    skillService            *service.HeroSkillService
}

func (m *GameModule) initServices() {
    container := &ServiceContainer{
        heroRepo:         impl.NewHeroRepository(m.db),
        classRepo:        impl.NewClassRepository(m.db),
        classSkillPoolRepo: impl.NewClassSkillPoolRepository(m.db),
        // ... 其他 repo ...
    }

    container.heroService = service.NewHeroService(
        container.heroRepo,
        container.classRepo,
        container.classSkillPoolRepo,
        // ... 等等 ...
    )
}
```

---

### 11. Handler 中缺少用户身份验证

**位置**：`hero_handler.go:76-81`

```go
// TODO: 从认证中间件获取用户ID
userID := c.Get("user_id")
if userID == nil {
    return response.EchoUnauthorized(c, h.respWriter, "未登录")
}
```

**问题**：
- 认证中间件还未实现
- 建议在 game_module 的中间件配置中添加

**建议**：
```go
// game_module.go 的中间件配置中添加
// 6. Auth 中间件（在 Error 中间件之前）
m.httpServer.Use(custommiddleware.AuthMiddleware(logger))
```

---

### 12. 缺失的 Repository 接口和实现

以下 Repository 需要实现或补全：

**必须实现**：
- [ ] `HeroAttributeTypeRepository.ListByCategory(ctx, category)` - 用于属性初始化
- [ ] `ClassSkillPoolRepository.GetByClassIDAndSkillID()` - 用于技能验证

**应该实现**：
- [ ] `HeroClassHistoryRepository.GetCurrentClass()` - 获取当前职业
- [ ] `HeroAttributeOperationRepository.DeleteExpiredOperations()` - 定时清理
- [ ] `HeroSkillOperationRepository.DeleteExpiredOperations()` - 定时清理

---

## 📋 修复优先级和时间表

### 🚨 **第一批（本日完成）**：P0 问题 - 2-3 小时
1. 修复经验系统逻辑（删除错误的 `experience_total += ...`）
2. 实现属性初始化函数
3. 添加技能池验证逻辑

### ⏰ **第二批（1-2 天内）**：P1 问题 - 4-6 小时
4. 修复技能升级等级数据
5. 重构 Service 依赖关系（改用函数回调）
6. 优化并发安全问题

### 📅 **第三批（本周完成）**：补充功能
7. 实现职业进阶和转职（新增代码）
8. 添加升级奖励机制
9. 完善错误处理和日志

### 💡 **第四批（可选、后续优化）**：P2 问题
10. 优化 JSONB 结构或拆表
11. 实现依赖注入容器
12. 添加单元测试和集成测试

---

## 📝 快速修复清单

- [ ] **hero_attribute_service.go:104** - 删除 `experience_total += ...`
- [ ] **hero_skill_service.go:111** - 删除 `experience_total += ...`
- [ ] **hero_attribute_service.go:186** - 删除 `experience_total -= ...`
- [ ] **hero_skill_service.go:339** - 删除 `experience_total -= ...`
- [ ] **hero_service.go:150-155** - 实现 `initializeAllocatedAttributes`
- [ ] **hero_skill_service.go:70-74** - 添加技能池验证
- [ ] **hero_skill_service.go:205-209** - 添加技能池验证
- [ ] **hero_skill_service.go:273** - 修改 `LevelsAdded: req.Levels` → `LevelsAdded: 1`
- [ ] 实现缺失的 Repository 接口方法

---

## 🎯 总体评分

| 维度 | 评分 | 说明 |
|------|------|------|
| **数据库设计** | ⭐⭐⭐⭐ | 结构合理，表设计良好，存在逻辑错误 |
| **Service 架构** | ⭐⭐⭐ | 功能完整度 60%，存在依赖关系问题 |
| **并发安全** | ⭐⭐ | JSONB 更新非原子，需改进 |
| **代码规范** | ⭐⭐⭐⭐ | 命名清晰，注释完善，风格一致 |
| **文件组织** | ⭐⭐⭐⭐ | 符合架构规范，逻辑清晰 |
| **整体质量** | ⭐⭐⭐ | 框架合理，需修复关键问题后才能上线 |

---

## 建议

1. **立即修复** P0 和 P1 的 8 个问题，预计 6-8 小时
2. **补全功能** 后在测试环境进行压力测试
3. **添加测试** 特别是并发场景下的测试
4. **代码审查** 修复后再请 Senior Review 一遍
5. **上线前检查** 确保所有 TODO 都被处理

---

**生成时间**：2025-10-16
**审查工具**：Claude Code AI
**下一步**：等待修复后的代码审查反馈
