#!/bin/bash
# Validate OpenSpec change completion and file cleanup
# Usage: ./scripts/validation/validate-change-completion.sh <change-id>

set -e

CHANGE_ID="$1"

if [ -z "$CHANGE_ID" ]; then
    echo "Usage: $0 <change-id>"
    echo "Example: $0 add-equipment-system"
    exit 1
fi

CHANGE_DIR="openspec/changes/$CHANGE_ID"

if [ ! -d "$CHANGE_DIR" ]; then
    echo "❌ Error: Change directory not found: $CHANGE_DIR"
    exit 1
fi

echo "🔍 Validating change completion: $CHANGE_ID"
echo ""

ERRORS=0
WARNINGS=0

# Check 1: All specs are created
echo "Checking specs completion..."
if [ ! -d "$CHANGE_DIR/specs" ]; then
    echo "  ❌ FAIL: specs/ directory not found"
    ERRORS=$((ERRORS + 1))
else
    SPEC_COUNT=$(find "$CHANGE_DIR/specs" -name "spec.md" | wc -l)
    if [ "$SPEC_COUNT" -eq 0 ]; then
        echo "  ❌ FAIL: No spec files found"
        ERRORS=$((ERRORS + 1))
    else
        echo "  ✅ PASS: Found $SPEC_COUNT spec file(s)"
    fi
fi

# Check 2: Run openspec validate
echo ""
echo "Running openspec validate --strict..."
if command -v openspec &> /dev/null; then
    if openspec validate "$CHANGE_ID" --strict 2>&1 | tee /tmp/openspec-validate.log; then
        echo "  ✅ PASS: openspec validate passed"
    else
        echo "  ❌ FAIL: openspec validate failed"
        echo "     See errors above"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "  ⚠️  SKIP: openspec command not found"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 3: File cleanup - temporary files in change directory
echo ""
echo "Checking for temporary files in change directory..."
TEMP_FILES=$(find "$CHANGE_DIR" -type f \( \
    -name "test-*.sh" -o \
    -name "*-test.sh" -o \
    -name "curl-*.sh" -o \
    -name "debug-*.sh" -o \
    -name "temp-*" -o \
    -name "*.tmp" -o \
    -name "*.bak" -o \
    -name "*.log" -o \
    -name "*.out" -o \
    -name "output.txt" -o \
    -name "result.txt" -o \
    -name "notes.md" -o \
    -name "scratch.md" -o \
    -name "todo.md" \
\) 2>/dev/null || true)

if [ -n "$TEMP_FILES" ]; then
    echo "  ⚠️  WARNING: Found temporary files:"
    echo "$TEMP_FILES" | sed 's/^/     /'
    echo "     Consider removing these files or moving important tests to tests/ directory"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✅ PASS: No temporary files found"
fi

# Check 4: File cleanup - project root directory
echo ""
echo "Checking for temporary files in project root..."
ROOT_TEMP_FILES=$(find . -maxdepth 1 -type f \( \
    -name "test.sh" -o \
    -name "test-*.sh" -o \
    -name "*-test.sh" -o \
    -name "curl-*.sh" -o \
    -name "debug*.sh" -o \
    -name "*.log" -o \
    -name "*.out" -o \
    -name "output.txt" -o \
    -name "result.txt" -o \
    -name "temp*.json" -o \
    -name "*.tmp" -o \
    -name "notes.md" -o \
    -name "scratch.md" \
\) 2>/dev/null || true)

if [ -n "$ROOT_TEMP_FILES" ]; then
    echo "  ⚠️  WARNING: Found temporary files in project root:"
    echo "$ROOT_TEMP_FILES" | sed 's/^/     /'
    echo "     Please remove these files or move them to temp/ directory"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✅ PASS: No temporary files in project root"
fi

# Check 5: README.md status
echo ""
echo "Checking README.md status..."
if [ -f "$CHANGE_DIR/README.md" ]; then
    if grep -q "状态.*已完成\|Status.*Complete\|✅.*完成" "$CHANGE_DIR/README.md"; then
        echo "  ✅ PASS: README.md shows completed status"
    else
        echo "  ⚠️  WARNING: README.md status not updated to 'completed'"
        echo "     Update README.md to reflect final status"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "  ⚠️  WARNING: README.md not found"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 6: Directory structure
echo ""
echo "Checking directory structure..."
REQUIRED_FILES=("proposal.md" "tasks.md" "README.md")
for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$CHANGE_DIR/$file" ]; then
        echo "  ✅ $file exists"
    else
        echo "  ❌ $file missing"
        ERRORS=$((ERRORS + 1))
    fi
done

if [ -d "$CHANGE_DIR/specs" ]; then
    echo "  ✅ specs/ directory exists"
else
    echo "  ❌ specs/ directory missing"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "✅ Validation PASSED"
    echo ""
    echo "Change is ready for archiving!"
    echo ""
    echo "Next steps:"
    echo "  1. Review all changes one final time"
    echo "  2. Archive: openspec/changes/$CHANGE_ID → openspec/changes/archive/YYYY-MM-DD-$CHANGE_ID"
    echo "  3. Apply specs to global specs/"
    echo "  4. Run: openspec validate --strict"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo "⚠️  Validation PASSED with $WARNINGS warning(s)"
    echo ""
    echo "Consider addressing the warnings above before archiving."
    exit 0
else
    echo "❌ Validation FAILED with $ERRORS error(s) and $WARNINGS warning(s)"
    echo ""
    echo "Please fix the errors above before archiving."
    exit 1
fi

