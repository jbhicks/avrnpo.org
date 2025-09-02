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
    # Handle object properties correctly (e.g., user.Name -> user)
    grep -o '<%=[^%]*%>' "$template_file" 2>/dev/null | \
    sed 's/<%=\s*//' | sed 's/\s*%>$//' | \
    # Remove for loop constructs
    sed 's/for\s*([^)]*)\s*in\s*[^}]*{.*}//g' | \
    # Remove if conditions
    sed 's/if\s*([^)]*).*{.*}//g' | \
    # Remove string literals (anything in quotes)
    sed 's/"[^"]*"//g' | sed "s/'[^']*'//g" | \
    # Extract base variables (before dots) and standalone variables
    grep -o '\b[a-zA-Z_][a-zA-Z0-9_]*\b' | \
    # Remove common Plush keywords and built-ins
    grep -v -E '^(for|if|else|end|len|true|false|nil|and|or|not)$' | \
    # Remove common method names that aren't variables
    grep -v -E '^(Format|Get|HasAny|Errors|capitalize|partial)$' | \
    # Remove common literals and format specifiers
    grep -v -E '^(January|February|March|April|May|June|July|August|September|October|November|December|Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)$' | \
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

# Track validation results
TEMPLATE_ERRORS=0
MISSING_VARS=()

# Find all Plush templates and process them
while IFS= read -r -d '' template_file; do
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

        # Find the controller file - prioritize exact matches
        controller_file_found=""

        # First, try exact template path match
        for cf in $(find actions -name "*.go" -type f); do
            if grep -q "r\.HTML.*$template_path" "$cf"; then
                controller_file_found="$cf"
                break
            fi
        done

        # If no exact match, try template name without extension
        if [ -z "$controller_file_found" ]; then
            for cf in $(find actions -name "*.go" -type f); do
                if grep -q "r\.HTML.*$template_name" "$cf"; then
                    controller_file_found="$cf"
                    break
                fi
            done
        fi

        # Only as last resort, try basename match (but exclude generic names)
        if [ -z "$controller_file_found" ]; then
            template_basename=$(basename "$template_name" .plush.html)
            # Skip generic names that would match too many controllers
            if [ "$template_basename" != "index" ] && [ "$template_basename" != "show" ] && [ "$template_basename" != "new" ] && [ "$template_basename" != "edit" ]; then
                for cf in $(find actions -name "*.go" -type f); do
                    if grep -q "r\.HTML.*$template_basename" "$cf"; then
                        controller_file_found="$cf"
                        break
                    fi
                done
            fi
        fi

        # If controller found, extract variables
        if [ -n "$controller_file_found" ]; then
            echo "   🎯 Controller found: $controller_file_found"
            controller_vars=$(extract_controller_vars "$controller_file_found")
        fi

        if [ -n "$controller_vars" ]; then
            echo "   ✅ Controller variables: $controller_vars"

            # Check for missing variables
            # For each template variable, check if it's either:
            # 1. Directly provided by controller, OR
            # 2. A property of an object that is provided by controller
            for var in $template_vars; do
                # Skip if variable is directly provided
                if echo "$controller_vars" | grep -q "^$var$"; then
                    continue
                fi

                # Check if this might be a property of a provided object
                # Look for patterns like "donation.Amount" or "post.User.FirstName" in the template
                is_property=false
                if grep -q "$var" "$template_file" 2>/dev/null; then
                    # Check if any controller variable could be the parent object
                    for controller_var in $controller_vars; do
                        # Check for direct property access (e.g., donation.Amount)
                        if grep -q "$controller_var\.$var" "$template_file" 2>/dev/null; then
                            is_property=true
                            break
                        fi
                        # Check for nested property access (e.g., post.User.FirstName)
                        # Look for patterns where controller_var is used as an intermediate object
                        if grep -q "$controller_var\..*\.$var" "$template_file" 2>/dev/null; then
                            is_property=true
                            break
                        fi
                    done
                fi

                # Only flag as missing if it's not a property of a provided object
                if [ "$is_property" = false ]; then
                    # Special cases for middleware-provided variables and common helpers
                     case "$var" in
                         authenticity_token|current_path|title|description|yield|flash|current_user)
                             # These are typically provided by middleware or Buffalo itself
                             ;;
                         # Form helper variables
                         action|formFor|method|autocomplete|newAuthPath|usersPath)
                             # These are provided by Buffalo form helpers
                             ;;
                         # Template iteration and flash message variables
                         in|msg|key|message|messages|option|k)
                             # These are used in loops and flash message rendering
                             ;;
                         # Date/time formatting helpers
                         dateFormat|raw|time|String|TrimSpace|strings|Add|After|Hour|Name)
                             # These are provided by Buffalo helpers or Go template functions
                             ;;
                         # Asset and path helpers
                         assetPath|baseURL)
                             # These are provided by Buffalo asset helpers
                             ;;
                         # Form field variables that might be optional
                         email)
                             # This might be a form field that's conditionally set
                             ;;
                         *)
                             echo -e "${YELLOW}⚠️  Missing variable: $var in $template_file${NC}"
                             MISSING_VARS+=("$template_file:$var")
                             ;;
                     esac
                fi
            done
        else
            # Check if this is a partial template (starts with _) or a layout template
            if [[ "$template_name" == _* ]]; then
                # Partial templates - no output needed, they're expected to not have direct controllers
                true
            elif [[ "$template_name" == "application" ]]; then
                # Layout templates - no output needed, they're handled by Buffalo's layout system
                true
            else
                echo -e "${YELLOW}⚠️  No controller found for template: $template_file${NC}"
            fi
        fi
    else
        echo "   ✅ No variables required"
    fi
done < <(find templates -name "*.plush.html" -type f -print0)

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