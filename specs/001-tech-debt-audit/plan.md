# Implementation Plan: 技术债务审计

**Branch**: `001-tech-debt-audit` | **Date**: 2025-10-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-tech-debt-audit/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

本功能旨在对现有代码库进行全面审计,识别与项目宪法五大原则(代码质量、测试驱动、用户体验、性能、可观测性)存在冲突的代码模式。审计结果将生成一份中文技术债务报告文件 `TECH_DEBT.md`,帮助开发团队了解技术债务现状并为未来改进提供指导。

**技术方法**: 采用混合式审计方法(静态分析 + 人工审查),按宪法五大原则系统性审查代码,生成结构化 Markdown 文档。

## Technical Context

**Language/Version**: N/A (文档生成任务,无需编程实现)
**Primary Dependencies**: 无(仅需访问现有 Go 代码库和宪法文档)
**Storage**: 文件系统(生成 TECH_DEBT.md 到项目根目录)
**Testing**: N/A (一次性审计活动,无自动化测试需求)
**Target Platform**: 任意平台(审查 Go 语言项目,生成 Markdown 文档)
**Project Type**: 文档生成(documentation-generation)
**Performance Goals**: 审计完成时间 <10 分钟,报告生成后技术负责人应能在 30 分钟内理解技术债务整体状况
**Constraints**: 必须覆盖至少 80% 的核心源代码文件,识别 20-100 个最严重问题,每个宪法原则至少 3 个示例
**Scale/Scope**: 审计范围包括 internal/、cmd/、scripts/、migrations/ 等核心目录,排除自动生成代码(models/)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

本技术债务审计项目与宪法原则的符合性检查:

### ✅ 代码质量优先 (符合)
- 本审计活动不产生新代码,仅生成 Markdown 文档
- 审计目标是识别违反代码质量标准的现有代码,符合宪法精神
- **参考**: constitution.md 第一章

### ✅ 测试驱动开发 (N/A)
- 本审计是文档生成活动,无需编写代码或测试
- **参考**: constitution.md 第二章

### ✅ 用户体验一致性 (符合)
- 生成的 TECH_DEBT.md 将帮助开发人员理解技术债务,间接改善未来的用户体验
- 报告采用统一的 Markdown 格式,确保阅读体验一致
- **参考**: constitution.md 第三章

### ✅ 性能与资源效率 (符合)
- 审计活动限定在 10 分钟内完成,符合效率要求
- 聚焦最严重问题(20-100 个),避免资源浪费
- **参考**: constitution.md 第四章

### ✅ 可观测性与调试 (符合)
- 审计报告将为每个问题提供清晰的文件路径和行号,增强代码可追溯性
- 按宪法原则分类的组织结构,便于理解技术债务分布
- **参考**: constitution.md 第五章

### ✅ 历史代码处理 (完全符合)
- 本审计严格遵循宪法"历史代码处理"原则:只记录问题,不主动修复
- 符合"为历史定罪"的初衷,为未来应用"童子军军规"提供指导
- **参考**: constitution.md "历史代码处理"章节

**结论**: 无宪法违规,可以继续 Phase 0 研究。Phase 0 已完成(research.md),Phase 1 设计已完成(quickstart.md),宪法检查仍然通过。

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

本功能不涉及源代码结构变更,仅生成文档:

```
# 现有项目结构(待审计)
internal/          # 核心业务逻辑(需要审计)
├── auth/          # 认证授权
├── game/          # 游戏逻辑(注:不考虑战斗系统具体实现)
└── common/        # 共享组件

cmd/               # 命令行入口(需要审计)
scripts/           # 部署脚本(需要审计)
migrations/        # 数据库迁移(需要审计)
models/            # SQLBoiler 生成的模型(排除审计)

# 审计输出
TECH_DEBT.md       # 本功能生成的技术债务报告
```

**Structure Decision**: 本功能是文档生成任务,不创建新的源代码目录。审计范围覆盖现有核心目录(internal/、cmd/、scripts/、migrations/),排除自动生成代码(models/)。根据用户要求,本项目完全不考虑前端和战斗的具体系统,因此审计聚焦于后端核心业务逻辑、认证、通用组件和基础设施代码。审计结果将输出到项目根目录的 TECH_DEBT.md 文件。

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

N/A - 本项目无宪法违规,无需复杂度跟踪。

## Implementation Notes

本计划已完成 Phase 0 和 Phase 1:

- **Phase 0 (研究)**: 已生成 [research.md](./research.md),定义了审计方法、问题分类框架、报告格式等
- **Phase 1 (设计)**: 已生成 [quickstart.md](./quickstart.md),提供了使用技术债务报告的完整指南
- **data-model.md**: N/A (本功能无复杂数据模型,问题条目结构已在 research.md 中定义)
- **contracts/**: N/A (本功能无 API 契约,报告格式规范已在 research.md 中说明)

下一步:执行 `/speckit.tasks` 命令生成 tasks.md,规划具体的审计执行任务。

