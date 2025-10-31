#!/bin/bash
# Validate OpenSpec change proposal completeness
# Usage: ./scripts/validation/validate-change-proposal.sh <change-id>

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

echo "🔍 Validating change proposal: $CHANGE_ID"
echo ""

ERRORS=0

# Check 1: specs/ directory exists
echo "Checking specs/ directory..."
if [ ! -d "$CHANGE_DIR/specs" ]; then
    echo "  ❌ FAIL: specs/ directory not found"
    echo "     Create specs/ directory and add spec deltas for affected capabilities"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✅ PASS: specs/ directory exists"
    
    # Check if there are any spec files
    SPEC_COUNT=$(find "$CHANGE_DIR/specs" -name "spec.md" | wc -l)
    if [ "$SPEC_COUNT" -eq 0 ]; then
        echo "  ❌ FAIL: No spec.md files found in specs/"
        echo "     Add at least one spec delta file"
        ERRORS=$((ERRORS + 1))
    else
        echo "  ✅ PASS: Found $SPEC_COUNT spec delta file(s)"
    fi
fi

# Check 2: proposal.md exists
echo ""
echo "Checking proposal.md..."
if [ ! -f "$CHANGE_DIR/proposal.md" ]; then
    echo "  ❌ FAIL: proposal.md not found"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✅ PASS: proposal.md exists"
fi

# Check 3: tasks.md exists
echo ""
echo "Checking tasks.md..."
if [ ! -f "$CHANGE_DIR/tasks.md" ]; then
    echo "  ❌ FAIL: tasks.md not found"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✅ PASS: tasks.md exists"
fi

# Check 4: Spec deltas have correct format
if [ -d "$CHANGE_DIR/specs" ]; then
    echo ""
    echo "Checking spec delta format..."
    
    for spec_file in $(find "$CHANGE_DIR/specs" -name "spec.md"); do
        echo "  Checking $spec_file..."
        
        # Check for delta operations
        if ! grep -q "## ADDED\|## MODIFIED\|## REMOVED" "$spec_file"; then
            echo "    ❌ FAIL: No delta operations found (## ADDED|MODIFIED|REMOVED)"
            ERRORS=$((ERRORS + 1))
        else
            echo "    ✅ PASS: Delta operations found"
        fi
        
        # Check for Requirements
        if ! grep -q "### Requirement:" "$spec_file"; then
            echo "    ❌ FAIL: No Requirements found"
            ERRORS=$((ERRORS + 1))
        else
            echo "    ✅ PASS: Requirements found"
        fi
        
        # Check for Scenarios
        if ! grep -q "#### Scenario:" "$spec_file"; then
            echo "    ❌ FAIL: No Scenarios found"
            ERRORS=$((ERRORS + 1))
        else
            echo "    ✅ PASS: Scenarios found"
        fi
    done
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✅ Validation PASSED"
    echo ""
    echo "Next steps:"
    echo "  1. Run: openspec validate $CHANGE_ID --strict"
    echo "  2. Request proposal review and approval"
    exit 0
else
    echo "❌ Validation FAILED with $ERRORS error(s)"
    echo ""
    echo "Please fix the errors above before proceeding."
    exit 1
fi

