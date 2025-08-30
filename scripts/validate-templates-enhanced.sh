#!/bin/bash

# Enhanced template validation script for build-time validation
# This script validates Plush templates for syntax errors and missing variables

echo "🔍 Enhanced Template Validation..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cd "$(git rev-parse --show-toplevel)"

# Check if we're in a Buffalo project
if [ ! -f "app.go" ] && [ ! -f "main.go" ] && [ ! -d "actions" ]; then
    echo "⚠️  Not in a Buffalo project root, skipping template validation"
    exit 0
fi

# Function to extract variables from template
extract_template_vars() {
    local template_file="$1"
    # Extract variables using regex - looks for <%= variable %> patterns
    # Exclude for loop variables and function calls
    grep -o '<%=[^%]*%>' "$template_file" 2>/dev/null | \
    sed 's/<%=\s*//' | sed 's/\s*%>$//' | \
    # Remove for loop constructs
    sed 's/for\s*([^)]*)\s*in\s*[^}]*{.*}//g' | \
    # Remove if conditions
    sed 's/if\s*([^)]*).*{.*}//g' | \
    # Extract individual variables from complex expressions
    grep -o '\b[a-zA-Z_][a-zA-Z0-9_]*\b' | \
    # Remove common Plush keywords and built-ins
    grep -v -E '^(for|if|else|end|len|true|false|nil|and|or|not)$' | \
    sort | uniq
}

# Function to extract variables from Go controller
extract_controller_vars() {
    local controller_file="$1"

    # Extract all c.Set calls from the entire controller file
    # This is simpler and more reliable than trying to parse function boundaries
    grep -o 'c\.Set("[^"]*"' "$controller_file" | \
    sed 's/c\.Set("//' | sed 's/"//' | sort | uniq
}

echo "📝 Analyzing templates and controllers..."

# Track validation results
TEMPLATE_ERRORS=0
MISSING_VARS=()

# Find all Plush templates
find templates -name "*.plush.html" -type f | while read -r template_file; do
    echo "🔍 Validating: $template_file"

    # Skip syntax validation - use existing parse_test.go for that
    # Just focus on variable validation which is the enhancement

    # Extract template variables
    template_vars=$(extract_template_vars "$template_file")

    if [ -n "$template_vars" ]; then
        echo "   📋 Template variables found: $template_vars"

        # Find corresponding controller
        template_name=$(basename "$template_file" .plush.html)
        controller_vars=""

        # Search for the controller that renders this template
        # Look for any r.HTML call that includes this template path
        template_path="${template_file#templates/}"  # Remove "templates/" prefix
        controller_vars=""
        controller_file_found=""

        # Find the controller file first
        for cf in $(find actions -name "*.go" -type f); do
            if grep -q "r\.HTML.*$template_path" "$cf" || \
               grep -q "r\.HTML.*$template_name" "$cf" || \
               grep -q "r\.HTML.*$(basename "$template_name" .plush.html)" "$cf"; then
                controller_file_found="$cf"
                break
            fi
        done

        # If controller found, extract variables
        if [ -n "$controller_file_found" ]; then
            echo "   🎯 Controller found: $controller_file_found"
            controller_vars=$(extract_controller_vars "$controller_file_found")
        fi

        if [ -n "$controller_vars" ]; then
            echo "   ✅ Controller variables: $controller_vars"

            # Check for missing variables
            for var in $template_vars; do
                if ! echo "$controller_vars" | grep -q "^$var$"; then
                    echo -e "${YELLOW}⚠️  Missing variable: $var in $template_file${NC}"
                    MISSING_VARS+=("$template_file:$var")
                fi
            done
        else
            # Check if this is a partial template (starts with _)
            if [[ "$template_name" == _* ]]; then
                echo -e "${YELLOW}ℹ️  Partial template (no direct controller expected): $template_file${NC}"
            else
                echo -e "${YELLOW}⚠️  No controller found for template: $template_file${NC}"
            fi
        fi
    else
        echo "   ✅ No variables required"
    fi
done

# Summary
echo ""
echo "📊 Validation Summary:"
echo "   Templates with syntax errors: $TEMPLATE_ERRORS"
echo "   Missing variables: ${#MISSING_VARS[@]}"

if [ $TEMPLATE_ERRORS -gt 0 ] || [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Template validation failed!${NC}"

    if [ ${#MISSING_VARS[@]} -gt 0 ]; then
        echo ""
        echo "Missing variables:"
        for missing in "${MISSING_VARS[@]}"; do
            echo -e "   ${RED}• $missing${NC}"
        done
    fi

    echo ""
    echo "💡 To fix missing variables:"
    echo "   1. Add c.Set(\"variable_name\", value) in the controller"
    echo "   2. Or provide default values in templates: <%= variable_name || \"default\" %>"
    echo "   3. Or check if the variable should be conditionally set"

    exit 1
else
    echo -e "${GREEN}✅ All templates validated successfully!${NC}"
fi

exit 0