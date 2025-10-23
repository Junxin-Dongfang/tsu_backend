# 数据模型: 技术债务审计

**功能**: 技术债务审计
**阶段**: Phase 1 - 数据建模
**创建日期**: 2025-10-21

## 模型概述

本文档定义技术债务审计报告中使用的核心数据实体及其关系。虽然这是一个文档生成项目(不涉及数据库),但清晰的数据模型有助于确保报告结构的一致性和完整性。

---

## 实体 1: 审计报告 (AuditReport)

代表完整的技术债务审计报告文档。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `report_id` | String | ✅ | 报告唯一标识符(如 "tech-debt-2025-10-21") |
| `title` | String | ✅ | 报告标题(如 "TSU 项目技术债务审计报告") |
| `audit_date` | Date | ✅ | 审计执行日期 |
| `audited_directories` | List<String> | ✅ | 被审计的目录列表 |
| `total_issues_count` | Integer | ✅ | 发现的问题总数 |
| `severity_distribution` | SeverityStats | ✅ | 按严重程度的问题分布 |
| `principle_distribution` | PrincipleStats | ✅ | 按宪法原则的问题分布 |
| `executive_summary` | String | ✅ | 执行摘要文本 |
| `key_findings` | List<String> | ✅ | 关键发现列表(5-10 条) |
| `issues` | List<TechDebtItem> | ✅ | 所有技术债务问题条目 |
| `appendices` | List<Appendix> | ⚪ | 附录章节 |

### 验证规则

- 问题总数必须在 20-100 之间
- 每个宪法原则至少有 3 个问题
- 执行摘要长度 200-500 字

### 示例

```yaml
report_id: "tech-debt-2025-10-21"
title: "TSU 项目技术债务审计报告"
audit_date: 2025-10-21
audited_directories:
  - "internal/game/"
  - "internal/auth/"
  - "cmd/"
  - "scripts/"
  - "migrations/"
total_issues_count: 45
severity_distribution:
  critical: 8
  moderate: 22
  minor: 15
principle_distribution:
  code_quality: 12
  test_driven: 10
  user_experience: 8
  performance: 9
  observability: 6
```

---

## 实体 2: 技术债务项 (TechDebtItem)

代表单个技术债务问题条目。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `issue_id` | String | ✅ | 问题唯一标识符(如 "CQ-001") |
| `title` | String | ✅ | 问题简短标题(10-50 字) |
| `principle` | PrincipleCategory | ✅ | 违反的宪法原则 |
| `severity` | SeverityLevel | ✅ | 严重程度(严重/中等/轻微) |
| `file_path` | String | ✅ | 文件路径(相对于项目根目录) |
| `line_range` | LineRange | ⚪ | 代码行号范围 |
| `violated_clause` | String | ✅ | 违反的具体宪法条款 |
| `description` | String | ✅ | 问题详细描述(100-300 字) |
| `code_snippet` | String | ⚪ | 问题代码片段(建议提供) |
| `impact_scope` | String | ⚪ | 影响范围说明 |
| `improvement_suggestion` | String | ✅ | 改进建议(50-200 字) |
| `related_issues` | List<String> | ⚪ | 相关问题 ID 列表 |
| `priority` | Integer | ⚪ | 修复优先级(1-10,数字越小越优先) |

### 验证规则

- `issue_id` 必须遵循格式: `{原则缩写}-{序号}` (如 CQ-001, TD-002)
- `file_path` 必须是有效的相对路径
- `description` 必须清晰解释问题及其违反宪法的原因
- `improvement_suggestion` 必须提供可操作的改进建议

### 示例

```yaml
issue_id: "CQ-001"
title: "game_service.go 中缺少错误处理"
principle: CODE_QUALITY
severity: CRITICAL
file_path: "internal/game/game_service.go"
line_range:
  start: 145
  end: 158
violated_clause: "所有错误处理必须显式且有意义(不允许静默失败)"
description: |
  在 ProcessGameAction 函数中,调用 db.Query() 后没有检查错误。
  如果数据库查询失败,错误会被静默忽略,导致后续逻辑基于空结果集执行,
  可能产生难以调试的 bug 和数据不一致问题。
code_snippet: |
  ```go
  func ProcessGameAction(actionID string) {
      rows, _ := db.Query("SELECT * FROM actions WHERE id = ?", actionID)
      // 错误被忽略,使用 _ 丢弃
      defer rows.Close()
      // ...
  }
  ```
impact_scope: "影响所有游戏动作处理流程,可能导致玩家数据丢失"
improvement_suggestion: |
  添加显式错误处理和日志记录:
  1. 检查 db.Query() 返回的错误
  2. 记录错误日志并包含上下文(actionID)
  3. 返回错误给调用者,避免静默失败
related_issues: ["OB-003"]
priority: 1
```

---

## 实体 3: 宪法原则类别 (PrincipleCategory)

枚举类型,表示宪法原则和项目管理类别。

### 枚举值

| 值 | 缩写 | 中文名称 | 描述 |
|---|------|----------|------|
| `CODE_QUALITY` | CQ | 代码质量优先 | 代码规范、文档、错误处理、安全性 |
| `TEST_DRIVEN` | TD | 测试驱动开发 | 单元测试、集成测试、代码覆盖率 |
| `USER_EXPERIENCE` | UX | 用户体验一致性 | API 一致性、错误消息、响应格式 |
| `PERFORMANCE` | PF | 性能与资源效率 | 查询优化、并发控制、资源管理 |
| `OBSERVABILITY` | OB | 可观测性与调试 | 日志记录、监控指标、错误追踪 |
| `MIGRATIONS` | MG | 数据库迁移管理 | 迁移脚本质量、命名规范、文档完整性 |
| `FILE_CLEANUP` | FC | 项目文件组织 | 文件清理、目录结构、配置管理 |

### 使用规则

- 每个 `TechDebtItem` 必须映射到一个类别
- 如果问题涉及多个类别,选择最主要的一个,其他类别通过 `related_issues` 关联
- 前五个为宪法原则类别,后两个为项目管理类别

---

## 实体 4: 严重程度级别 (SeverityLevel)

枚举类型,表示问题的严重程度。

### 枚举值

| 级别 | 标识 | 描述 | 颜色标记 |
|------|------|------|----------|
| `CRITICAL` | 严重 | 严重影响系统安全性、可靠性或性能的问题 | 🔴 |
| `MODERATE` | 中等 | 影响代码质量或可维护性的问题 | 🟡 |
| `MINOR` | 轻微 | 轻微的代码风格或规范问题 | 🟢 |

### 严重程度判定标准

**🔴 严重 (CRITICAL)**:
- 安全漏洞(SQL 注入、XSS、敏感信息泄露)
- 严重性能问题(N+1 查询、goroutine 泄漏、资源未释放)
- 核心业务逻辑无测试覆盖
- 错误被静默吞噬,导致数据不一致风险
- API 行为不一致,影响客户端集成

**🟡 中等 (MODERATE)**:
- 缺少文档注释
- 测试覆盖率 <80%
- 缺少日志记录或日志缺少上下文
- 代码复杂度高(圈复杂度 >15)
- 缺少 Prometheus 指标
- 错误消息不清晰

**🟢 轻微 (MINOR)**:
- 命名不符合 Go 规范
- 代码风格不一致
- 注释拼写错误
- HTTP 状态码使用不规范(但不影响功能)
- 日志级别使用不当

---

## 实体 5: 行号范围 (LineRange)

表示代码片段的行号范围。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `start` | Integer | ✅ | 起始行号 |
| `end` | Integer | ⚪ | 结束行号(如果是单行问题可省略) |

### 验证规则

- `start` 必须 > 0
- 如果提供 `end`,则 `end` 必须 >= `start`

### 示例

```yaml
# 单行问题
line_range:
  start: 145

# 多行问题
line_range:
  start: 145
  end: 158
```

---

## 实体 6: 严重程度统计 (SeverityStats)

表示按严重程度分布的问题统计。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `critical` | Integer | ✅ | 严重问题数量 |
| `moderate` | Integer | ✅ | 中等问题数量 |
| `minor` | Integer | ✅ | 轻微问题数量 |

### 计算规则

- `critical + moderate + minor` 必须等于 `total_issues_count`

---

## 实体 7: 原则统计 (PrincipleStats)

表示按宪法原则和项目管理类别分布的问题统计。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `code_quality` | Integer | ✅ | 代码质量原则相关问题数 |
| `test_driven` | Integer | ✅ | 测试驱动原则相关问题数 |
| `user_experience` | Integer | ✅ | 用户体验原则相关问题数 |
| `performance` | Integer | ✅ | 性能原则相关问题数 |
| `observability` | Integer | ✅ | 可观测性原则相关问题数 |
| `migrations` | Integer | ⚪ | 数据库迁移管理相关问题数(可选) |
| `file_cleanup` | Integer | ⚪ | 项目文件组织相关问题数(可选) |

### 验证规则

- 所有类别的问题数之和必须等于 `total_issues_count`
- 宪法原则类别(前五个)每个至少 3 个问题
- 项目管理类别(migrations、file_cleanup)可以为 0

---

## 实体 8: 附录 (Appendix)

表示报告末尾的附录章节。

### 属性

| 属性名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `id` | String | ✅ | 附录标识符(如 "A", "B") |
| `title` | String | ✅ | 附录标题 |
| `content` | String | ✅ | 附录内容 |

### 标准附录

1. **附录 A: 审计方法说明** - 解释审计流程和工具
2. **附录 B: 问题严重程度定义** - 详细说明严重程度判定标准
3. **附录 C: 改进优先级建议** - 基于宪法原则和严重程度的改进路线图

---

## 实体关系图

```
AuditReport (1)
  │
  ├── has many ──> TechDebtItem (*)
  │                   │
  │                   ├── belongs to ──> PrincipleCategory (1)
  │                   ├── belongs to ──> SeverityLevel (1)
  │                   └── has optional ──> LineRange (0..1)
  │
  ├── has one ──> SeverityStats (1)
  ├── has one ──> PrincipleStats (1)
  └── has many ──> Appendix (*)
```

---

## 数据完整性约束

### 全局约束

1. **问题数量约束**: 20 <= `total_issues_count` <= 100
2. **原则覆盖约束**: 每个 `PrincipleCategory` 至少 3 个问题
3. **ID 唯一性约束**: 所有 `issue_id` 必须唯一
4. **统计一致性约束**:
   - `severity_distribution` 总和 = `total_issues_count`
   - `principle_distribution` 总和 = `total_issues_count`

### 问题条目约束

1. **文件路径有效性**: `file_path` 必须指向项目中实际存在的文件
2. **描述完整性**: `description` 必须包含问题原因和影响
3. **建议可操作性**: `improvement_suggestion` 必须包含具体步骤

---

## 使用示例

### 完整的问题条目示例

```markdown
### 🔴 严重问题

#### TD-005: hero_service.go 核心业务逻辑缺少单元测试

- **文件**: `internal/game/hero_service.go:200-350`
- **违反条款**: 所有新功能必须包含关键路径的单元测试
- **问题描述**:
  HeroGrowthService 是游戏核心功能,负责英雄升级、属性计算和技能解锁。
  该服务包含复杂的业务逻辑(150 行代码),但完全没有单元测试覆盖。
  这意味着任何修改都可能引入 bug 而无法及时发现,严重影响游戏稳定性。

- **代码示例**:
  ```go
  // HeroGrowthService 处理英雄成长逻辑
  type HeroGrowthService struct {
      db *sql.DB
  }

  // ProcessLevelUp 处理英雄升级(无测试)
  func (s *HeroGrowthService) ProcessLevelUp(heroID int) error {
      // 150 行复杂业务逻辑
      // 包括属性计算、技能解锁、数据库更新
      // ...
  }
  ```

- **影响范围**:
  影响所有英雄升级功能,涉及数千名玩家的游戏体验。
  历史上已发生 2 次因缺少测试导致的升级 bug。

- **改进建议**:
  1. 为 HeroGrowthService 创建测试文件 `hero_service_test.go`
  2. 编写单元测试覆盖关键路径:
     - 正常升级流程
     - 属性计算正确性
     - 技能解锁条件
     - 边界情况(满级、经验不足等)
  3. 使用测试数据库或 mock,避免依赖真实数据库
  4. 目标:达到至少 80% 代码覆盖率

- **优先级**: 1 (最高)
```

---

## 后续步骤

基于此数据模型,下一步将:
1. 创建 `contracts/tech-debt-report-schema.md` - 定义报告的 Markdown 格式规范
2. 创建 `quickstart.md` - 编写如何使用和维护审计报告的指南
