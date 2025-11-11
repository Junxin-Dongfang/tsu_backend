# TSU 测试

本目录包含 TSU 游戏服务器项目的所有测试代码、测试框架和测试文档。

## 📁 目录结构

```
test/
├── api/                    # API 端到端测试
├── comprehensive/          # 综合测试框架（Shell-based Admin API 测试套件）
├── data/                   # 测试数据和配置
├── e2e/                    # 端到端测试
├── integration/            # 集成测试
├── reports/                # 历史测试报告归档
├── QUICK_START.md          # 快速开始指南
├── api-test-plan.md        # API 测试计划
└── prometheus-performance-test.md  # Prometheus 性能测试
```

## 🚀 快速开始

### 运行测试

```bash
# 运行所有 Go 单元测试
go test ./...

# 运行特定包的测试
go test ./internal/modules/game/service/...

# 运行单个测试
go test -run TestHeroAttributeUpdate ./internal/modules/game/service/

# 查看测试覆盖率
go test -cover ./...
```

### 综合测试框架

综合测试框架是一个基于 Shell 的 Admin API 测试套件：

```bash
cd test/comprehensive

# 运行所有测试套件
./main_test.sh

# 运行特定套件
./tests/01_system_health.sh
```

详细说明请参考 `comprehensive/README.md` 和 `comprehensive/QUICKSTART.md`。

## 📖 测试文档

### 快速指南

- **快速开始**: `QUICK_START.md` - 5 分钟开始测试
- **API 测试计划**: `api-test-plan.md` - API 测试策略和用例
- **综合测试指南**: `comprehensive/README.md` - Shell 测试框架完整文档

### 性能测试

- **Prometheus 性能测试**: `prometheus-performance-test.md`
  - 监控系统性能测试报告
  - 包含 wrk 压测结果和性能分析

## 📊 历史测试报告

所有历史测试报告已归档到 `reports/` 目录：

- 装备系统测试报告
- 技能系统测试报告
- API 功能测试报告
- 综合测试结果总结
- Bug 修复总结

报告文件列表：
```bash
$ ls reports/
API_TEST_REPORT.md
COMPLETE_EQUIPMENT_SYSTEM_TEST_REPORT.md
EQUIPMENT_SYSTEM_API_TEST_REPORT.md
FINAL_REPORT.md
FIXES_SUMMARY.md
SKILL_SYSTEM_TEST_REPORT.md
TEST_REPORT.md
TEST_RESULTS_SUMMARY.md
```

## 🎯 测试覆盖范围

### 综合测试框架覆盖

- **130+ 测试用例**
- **110+ API 接口**
- **11 个测试套件**:
  1. 系统健康检查
  2. 认证流程
  3. 用户管理
  4. RBAC 权限系统
  5. 基础游戏配置
  6. 元数据定义
  7. 技能系统
  8. 效果系统
  9. 动作系统
  10. 关联关系
  11. 边界条件

详见 `comprehensive/IMPLEMENTATION_SUMMARY.md`。

### 装备系统测试覆盖

- **42 个测试用例**
- **97.6% 通过率**
- **7 个核心模块**:
  - 物品管理
  - Tag 管理
  - 职业限制
  - 装备套装
  - 掉落池
  - 世界掉落
  - 装备槽位

详见 `reports/COMPLETE_EQUIPMENT_SYSTEM_TEST_REPORT.md`。

## 📝 测试编写规范

### 单元测试

遵循 Go 测试最佳实践：

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {
            name: "valid case",
            input: ...,
            want: ...,
            wantErr: false,
        },
        {
            name: "error case",
            input: ...,
            want: nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr = %v, error = %v", tt.wantErr, err)
            }
            // 断言逻辑...
        })
    }
}
```

### API 测试

使用综合测试框架编写 Shell 测试：

```bash
# test/comprehensive/tests/your_test.sh

test_your_feature() {
    local test_name="测试功能名称"

    # 准备测试数据
    local data=$(cat <<EOF
{
    "field": "value"
}
EOF
)

    # 执行 API 调用
    http_post "/admin/your-endpoint" "$data"

    # 断言结果
    assert_status 200 "$test_name"
    assert_field "data.field" "expected_value" "$test_name"
}
```

详细说明请参考 `comprehensive/README.md`。

## 🔄 持续集成

测试在以下场景自动运行：

- 每次 PR 提交
- 合并到 main 分支前
- 定期夜间构建

## 📞 支持

测试相关问题：
- 查看测试文档: `test/comprehensive/README.md`
- 查看质量指南: `docs/development/TECH_DEBT_GUIDE.md`
- 查看项目规范: `openspec/specs/code-quality/spec.md`

---

**最后更新**: 2025-11-11
**测试框架版本**: v1.0.0
