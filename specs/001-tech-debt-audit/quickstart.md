# 快速入门: 使用技术债务审计报告

**版本**: 1.0.0
**更新日期**: 2025-10-21
**适用人员**: 全体开发团队成员

## 概述

本指南帮助你快速理解和使用技术债务审计报告(`TECH_DEBT.md`),将其作为日常开发中的参考工具,实践"童子军军规"原则,逐步提升代码质量。

---

## 快速开始

### 1. 找到报告

技术债务审计报告位于项目根目录:

```bash
cd /path/to/tsu-self
cat TECH_DEBT.md
```

### 2. 理解报告结构

报告分为三个主要部分:

```
├── 执行摘要           # 问题统计和关键发现
├── 按宪法原则分类的问题  # 五大章节
│   ├── 一、代码质量优先
│   ├── 二、测试驱动开发
│   ├── 三、用户体验一致性
│   ├── 四、性能与资源效率
│   └── 五、可观测性与调试
└── 附录               # 审计方法和改进建议
```

### 3. 查找相关问题

#### 方法 1: 按文件路径搜索

当你准备修改某个文件时,先查看该文件是否有已知问题:

```bash
# 在报告中搜索文件名
grep "internal/game/hero_service.go" TECH_DEBT.md
```

#### 方法 2: 按问题 ID 查找

如果在代码审查或讨论中提到某个问题 ID(如 CQ-001):

```bash
# 查找特定问题
grep "CQ-001" TECH_DEBT.md -A 20
```

#### 方法 3: 按严重程度过滤

优先关注严重问题:

```bash
# 查找所有严重问题
grep "🔴 严重" TECH_DEBT.md -A 3
```

---

## 核心使用场景

### 场景 1: 修复 Bug 时

**步骤**:

1. 定位到需要修改的文件
2. 在 `TECH_DEBT.md` 中搜索该文件
3. 如果发现相关问题,在修复 Bug 的同时顺手改进:
   - 添加缺失的错误处理
   - 补充单元测试
   - 改善命名和注释

**示例**:

```bash
# 1. 你要修复 hero_service.go 中的一个 bug
# 2. 先查看报告
$ grep "hero_service.go" TECH_DEBT.md

# 3. 发现问题 TD-005: 缺少单元测试
# 4. 在修复 bug 的同时,为修改的函数添加测试
$ cat internal/game/hero_service_test.go
# 新增测试代码...

# 5. 提交时引用问题 ID
$ git commit -m "fix: 修复英雄升级计算错误,补充单元测试 (参考 TD-005)"
```

### 场景 2: 添加新功能时

**步骤**:

1. 查看需要修改的模块是否有已知问题
2. 了解该模块的主要债务类型
3. 确保新代码不引入相似问题:
   - 遵循宪法原则
   - 先写测试再实现
   - 包含完整的错误处理和日志

**示例**:

```bash
# 1. 要在 game_service.go 中添加新功能
# 2. 查看该模块的问题
$ grep "game_service.go" TECH_DEBT.md

# 3. 发现该模块普遍缺少错误处理和日志
# 4. 确保新功能包含这些要素
func NewFeature(ctx context.Context, input Input) (Output, error) {
    // ✅ 显式错误处理
    if err := validate(input); err != nil {
        log.Error(ctx, "validation failed", "input", input, "error", err)
        return Output{}, fmt.Errorf("validate input: %w", err)
    }

    // ✅ 结构化日志
    log.Info(ctx, "processing new feature", "input", input)

    // ...
}

# 5. 先写测试
$ cat internal/game/game_service_test.go
func TestNewFeature(t *testing.T) {
    // 测试代码...
}
```

### 场景 3: 代码审查时

**步骤**:

1. 审查 PR 时检查是否引入了类似的债务
2. 如果发现问题,引用报告中的相关条目
3. 建议作者参考改进建议

**示例**:

```markdown
<!-- 在 PR 评论中 -->
这段代码缺少错误处理,可能会导致静默失败。

参考 `TECH_DEBT.md` 中的 CQ-001,建议:
1. 检查 db.Query() 的错误
2. 添加日志记录
3. 向上传递错误而非忽略
```

### 场景 4: 计划重构时

**步骤**:

1. 查看执行摘要,了解整体债务状况
2. 阅读"关键发现"部分,识别高优先级领域
3. 参考"附录 C: 改进优先级建议"制定计划

**示例**:

```bash
# 1. 查看执行摘要
$ head -50 TECH_DEBT.md

# 2. 识别严重问题集中的模块
# 假设发现 hero_service.go 有 5 个严重问题

# 3. 创建重构分支
$ git checkout -b refactor/hero-service-improvements

# 4. 系统性解决该模块的所有问题
# 5. 在 PR 中说明解决了哪些债务问题
```

---

## "童子军军规"实践指南

### 原则回顾

> "让代码比你发现时更干净"

每次触碰代码时,至少做一件小改进,不要求全面重构。

### 推荐改进清单

当你修改某段代码时,检查以下快捷改进机会:

#### ✅ 代码质量(1-2 分钟)

- [ ] 函数/变量命名是否清晰?
- [ ] 是否有明显的拼写错误?
- [ ] 是否可以提取重复代码为函数?
- [ ] 是否可以简化复杂的条件判断?

#### ✅ 错误处理(2-3 分钟)

- [ ] 是否所有错误都被检查了?
- [ ] 错误消息是否包含足够的上下文?
- [ ] 是否正确使用了 `fmt.Errorf` 包装错误?

#### ✅ 日志记录(1-2 分钟)

- [ ] 关键操作是否有日志?
- [ ] 日志是否包含请求 ID/用户 ID?
- [ ] 日志级别是否合适(Info/Warn/Error)?

#### ✅ 测试(5-10 分钟)

- [ ] 修改的函数是否有单元测试?
- [ ] 是否可以快速添加一个测试用例?
- [ ] 边界情况是否被覆盖?

#### ✅ 文档(1-2 分钟)

- [ ] 公共函数是否有文档注释?
- [ ] 复杂逻辑是否有注释说明?
- [ ] 是否可以用更清晰的注释替换不明确的旧注释?

### 实践示例

**修改前**(发现问题 CQ-003):

```go
func ProcessAction(a string) {
    r, _ := db.Query("SELECT * FROM actions WHERE id = ?", a)
    defer r.Close()
    // ...
}
```

**修改后**(应用童子军军规):

```go
// ProcessAction processes the specified game action and returns the result.
// It returns an error if the action cannot be found or processed.
func ProcessAction(ctx context.Context, actionID string) (*ActionResult, error) {
    log.Info(ctx, "processing action", "action_id", actionID)

    rows, err := db.Query("SELECT * FROM actions WHERE id = ?", actionID)
    if err != nil {
        log.Error(ctx, "failed to query action", "action_id", actionID, "error", err)
        return nil, fmt.Errorf("query action %s: %w", actionID, err)
    }
    defer rows.Close()

    // ...
}
```

**改进清单**:
- ✅ 添加了函数文档注释
- ✅ 参数名更清晰(a → actionID)
- ✅ 添加了上下文参数(支持分布式追踪)
- ✅ 显式错误处理
- ✅ 结构化日志记录
- ✅ 错误包装和传递

---

## 团队协作

### 在每日站会中

- 分享你昨天修复的技术债务问题
- 讨论遇到的共性问题
- 识别需要团队协作解决的严重问题

### 在 Sprint 计划中

- 每个 Sprint 预留 10-20% 时间用于债务改进
- 从"关键发现"中选择 2-3 个问题作为 Sprint 目标
- 跟踪已修复问题的数量

### 在回顾会议中

- 审查本 Sprint 修复的债务问题
- 更新 `TECH_DEBT.md` ,标记已修复的问题
- 识别阻碍债务改进的系统性问题

---

## 常见问题

### Q1: 我是否必须修复我触碰文件中的所有问题?

**A**: 不需要。遵循"童子军军规",每次至少改进一点即可。优先选择:
1. 与你当前修改相关的问题
2. 能在 5-10 分钟内快速修复的问题
3. 严重程度高的问题

### Q2: 如果我不同意报告中的某个问题,怎么办?

**A**:
1. 在团队会议上讨论你的观点
2. 如果达成共识,更新 `TECH_DEBT.md` 删除或修改该问题
3. 记录讨论结果和决策理由
4. 通过 Git 提交跟踪变更

### Q3: 修复问题后如何更新报告?

**A**:
1. 在问题条目开头添加 `✅ 已修复` 标记
2. 添加修复日期和 PR 链接
3. 可选:将问题移至"已修复问题"章节

示例:

```markdown
#### ✅ 已修复 - CQ-001: game_service.go 中缺少错误处理

**修复日期**: 2025-10-25
**相关 PR**: #123

[原问题描述保持不变...]
```

### Q4: 如何避免重复记录相同类型的问题?

**A**:
1. 优先记录代表性问题
2. 使用"相关问题"字段关联相似问题
3. 在问题描述中说明影响范围(如"在 10 个文件中出现")

### Q5: 新发现的问题如何添加到报告?

**A**:
1. 遵循报告格式规范添加新问题
2. 分配唯一的问题 ID
3. 更新执行摘要中的统计数据
4. 通过 PR 审查确保问题描述准确

---

## 进阶使用

### 自动化工具集成

#### 1. Pre-commit Hook

在提交前自动检查是否引入新的代码质量问题:

```bash
# .git/hooks/pre-commit
#!/bin/bash
golangci-lint run --new-from-rev=HEAD~1
```

#### 2. CI/CD 集成

在 CI 流水线中运行静态分析,并与报告中的问题进行对比:

```yaml
# .github/workflows/code-quality.yml
- name: Run golangci-lint
  run: golangci-lint run --out-format=github-actions
```

#### 3. VS Code 集成

安装 Markdown 预览增强插件,方便查看报告:

```json
{
  "markdown-preview-enhanced.enableExtendedTableSyntax": true,
  "markdown-preview-enhanced.enableCriticMarkupSyntax": true
}
```

### 指标跟踪

建议跟踪以下指标:

| 指标 | 说明 | 目标 |
|------|------|------|
| 问题总数 | TECH_DEBT.md 中未修复的问题 | 每季度减少 20% |
| 严重问题数 | 🔴 标记的问题 | 降至 0 |
| 测试覆盖率 | 单元测试代码覆盖率 | 达到 80% |
| 修复速率 | 每 Sprint 修复的问题数 | ≥ 3 个/Sprint |

### 定期审计

建议每季度重新运行完整审计:

```bash
# 1. 创建新的审计分支
git checkout -b audit/2025-Q2

# 2. 重新执行审计(遵循 research.md 中的流程)
# 3. 生成新版本的 TECH_DEBT.md
# 4. 对比与上一版本的差异
git diff audit/2025-Q1:TECH_DEBT.md TECH_DEBT.md

# 5. 更新版本历史章节
# 6. 提交并合并
```

---

## 资源链接

- [项目宪法](../.specify/memory/constitution.md) - 了解宪法原则
- [报告格式规范](contracts/tech-debt-report-schema.md) - 报告结构说明
- [数据模型](data-model.md) - 问题条目数据结构
- [研究文档](research.md) - 审计方法论

---

## 反馈和改进

如果你对技术债务审计流程有任何建议:

1. 在团队会议上讨论
2. 通过 PR 更新相关文档
3. 将改进经验分享给团队

**记住**: 技术债务报告是一个活文档,应该随着项目演进持续更新和改进。

---

**最后更新**: 2025-10-21
**维护团队**: TSU 开发团队
