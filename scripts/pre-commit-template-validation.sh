#!/bin/bash

# Pre-commit hook for template validation
# This runs before each commit to ensure templates are valid

echo "🔍 Running pre-commit template validation..."

# Run enhanced template validation
if ! ./scripts/validate-templates-enhanced.sh; then
    echo ""
    echo "❌ Template validation failed!"
    echo "Please fix the template issues before committing."
    echo ""
    echo "💡 Common fixes:"
    echo "   • Add missing c.Set() calls in controllers"
    echo "   • Use default values: <%= variable || \"default\" %>"
    echo "   • Check conditional variable setting"
    exit 1
fi

echo "✅ All templates validated successfully!"
exit 0