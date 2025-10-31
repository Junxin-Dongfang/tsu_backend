#!/bin/bash

# 安装 Git hooks 脚本
# 用于在提交前自动运行质量检查

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

echo "🔧 安装 Git hooks..."

# 检查是否在 Git 仓库中
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo "❌ 错误: 不在 Git 仓库中"
    exit 1
fi

# 创建 pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash

# Pre-commit hook: 在提交前运行质量检查
# 根据 openspec/specs/code-quality/spec.md 的要求

set -e

echo "🔍 运行 pre-commit 质量检查..."

# 获取暂存的 Go 文件
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo "✅ 没有 Go 文件变更,跳过检查"
    exit 0
fi

echo "📝 检查的文件:"
echo "$STAGED_GO_FILES"
echo ""

# 1. 运行 gofmt 检查
echo "🔍 检查代码格式 (gofmt)..."
UNFORMATTED=$(gofmt -l $STAGED_GO_FILES)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ 以下文件需要格式化:"
    echo "$UNFORMATTED"
    echo ""
    echo "运行以下命令修复:"
    echo "  gofmt -w $UNFORMATTED"
    exit 1
fi
echo "✅ 代码格式检查通过"
echo ""

# 2. 运行 goimports 检查
echo "🔍 检查导入排序 (goimports)..."
if command -v goimports &> /dev/null; then
    for file in $STAGED_GO_FILES; do
        goimports -l "$file" > /dev/null
    done
    echo "✅ 导入排序检查通过"
else
    echo "⚠️  goimports 未安装,跳过检查"
    echo "   安装命令: go install golang.org/x/tools/cmd/goimports@latest"
fi
echo ""

# 3. 运行 go vet 检查
echo "🔍 运行 go vet 检查..."
go vet ./... || {
    echo "❌ go vet 检查失败"
    exit 1
}
echo "✅ go vet 检查通过"
echo ""

# 4. 运行 golangci-lint (如果安装)
echo "🔍 运行 golangci-lint 检查..."
if command -v golangci-lint &> /dev/null; then
    # 只检查暂存的文件
    golangci-lint run --new-from-rev=HEAD $STAGED_GO_FILES || {
        echo "❌ golangci-lint 检查失败"
        echo ""
        echo "运行以下命令查看详细信息:"
        echo "  make lint"
        echo ""
        echo "运行以下命令自动修复部分问题:"
        echo "  make lint-fix"
        exit 1
    }
    echo "✅ golangci-lint 检查通过"
else
    echo "⚠️  golangci-lint 未安装,跳过检查"
    echo "   安装命令: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
fi
echo ""

# 5. 检查是否有新增的 TODO 标记
echo "🔍 检查新增的 TODO 标记..."
NEW_TODOS=$(git diff --cached | grep "^+.*TODO\|^+.*FIXME" || true)
if [ -n "$NEW_TODOS" ]; then
    echo "⚠️  检测到新增的 TODO/FIXME 标记:"
    echo "$NEW_TODOS"
    echo ""
    echo "根据质量契约,应该优先偿还现有的 TODO,而不是添加新的 TODO。"
    echo "如果确实需要添加,请确保:"
    echo "  1. TODO 注释包含优先级 (P0-P3)"
    echo "  2. TODO 注释包含预估工作量"
    echo "  3. TODO 注释包含计划时间"
    echo ""
    read -p "是否继续提交? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
echo ""

# 6. 检查是否有测试文件
echo "🔍 检查测试文件..."
STAGED_TEST_FILES=$(echo "$STAGED_GO_FILES" | grep '_test\.go$' || true)
STAGED_NON_TEST_FILES=$(echo "$STAGED_GO_FILES" | grep -v '_test\.go$' || true)

if [ -n "$STAGED_NON_TEST_FILES" ] && [ -z "$STAGED_TEST_FILES" ]; then
    echo "⚠️  检测到代码变更但没有测试文件变更"
    echo ""
    echo "根据质量契约:"
    echo "  - 新增代码必须有单元测试"
    echo "  - 修改现有代码时应该添加或更新测试"
    echo ""
    echo "变更的非测试文件:"
    echo "$STAGED_NON_TEST_FILES"
    echo ""
    read -p "是否继续提交? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
echo ""

echo "✅ 所有 pre-commit 检查通过!"
echo ""
echo "💡 提示: 提交后记得更新 TECH_DEBT.md 记录偿还的技术债务"

exit 0
EOF

# 设置 hook 可执行权限
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Pre-commit hook 已安装: $HOOKS_DIR/pre-commit"
echo ""

# 创建 commit-msg hook
cat > "$HOOKS_DIR/commit-msg" << 'EOF'
#!/bin/bash

# Commit-msg hook: 检查提交消息格式
# 遵循 Conventional Commits 规范

set -e

COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# 检查提交消息格式
# 格式: <type>: <subject>
# 类型: feat, fix, docs, refactor, test, chore, perf, style

PATTERN="^(feat|fix|docs|refactor|test|chore|perf|style)(\(.+\))?: .{1,}"

if ! echo "$COMMIT_MSG" | grep -qE "$PATTERN"; then
    echo "❌ 提交消息格式不符合规范"
    echo ""
    echo "正确格式: <type>: <subject>"
    echo ""
    echo "类型 (type):"
    echo "  feat:     新功能"
    echo "  fix:      Bug 修复"
    echo "  docs:     文档更新"
    echo "  refactor: 代码重构"
    echo "  test:     测试相关"
    echo "  chore:    构建/工具链相关"
    echo "  perf:     性能优化"
    echo "  style:    代码格式调整"
    echo ""
    echo "示例:"
    echo "  feat: 添加英雄技能升级功能"
    echo "  fix: 修复英雄属性计算错误"
    echo "  docs: 更新 API 文档"
    echo "  refactor: 重构技能学习逻辑"
    echo ""
    echo "当前提交消息:"
    echo "$COMMIT_MSG"
    exit 1
fi

exit 0
EOF

# 设置 hook 可执行权限
chmod +x "$HOOKS_DIR/commit-msg"

echo "✅ Commit-msg hook 已安装: $HOOKS_DIR/commit-msg"
echo ""

echo "🎉 Git hooks 安装完成!"
echo ""
echo "已安装的 hooks:"
echo "  - pre-commit:  提交前运行质量检查"
echo "  - commit-msg:  检查提交消息格式"
echo ""
echo "💡 提示:"
echo "  - 如果需要跳过 pre-commit 检查,使用: git commit --no-verify"
echo "  - 建议团队所有成员都运行此脚本安装 hooks"

