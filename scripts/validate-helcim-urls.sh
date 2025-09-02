#!/bin/bash

# Helcim URL Validation Script
# This script checks for outdated or incorrect Helcim script URLs in the codebase
# Run this script to ensure all Helcim integrations use the canonical URL

set -e

echo "🔍 Validating Helcim URLs in codebase..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Forbidden URLs that should not exist
FORBIDDEN_URLS=(
    "gateway.helcim.com/js/helcim.js"
    "/js/helcim-pay.min.js"
)

# Canonical URL that should be used
CANONICAL_URL="https://secure.helcim.app/helcim-pay/services/start.js"

echo "📋 Checking for forbidden URLs..."

FOUND_ISSUES=0

# Check each forbidden URL
for url in "${FORBIDDEN_URLS[@]}"; do
    if [[ "$url" == "$CANONICAL_URL" ]]; then
        continue  # Skip the canonical URL
    fi

    echo "🔎 Searching for: $url"
    RESULTS=$(grep -r "$url" --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=public/assets 2>/dev/null || true)

    if [[ -n "$RESULTS" ]]; then
        echo -e "${RED}❌ FOUND FORBIDDEN URL: $url${NC}"
        echo "$RESULTS"
        echo ""
        FOUND_ISSUES=$((FOUND_ISSUES + 1))
    else
        echo -e "${GREEN}✅ No instances found: $url${NC}"
    fi
done

echo ""
echo "🔍 Checking for canonical URL usage..."

# Check that canonical URL exists
CANONICAL_RESULTS=$(grep -r "$CANONICAL_URL" --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=public/assets 2>/dev/null || true)

if [[ -n "$CANONICAL_RESULTS" ]]; then
    echo -e "${GREEN}✅ Canonical URL found in expected locations:${NC}"
    echo "$CANONICAL_RESULTS"
else
    echo -e "${YELLOW}⚠️  Canonical URL not found. This might be expected if no Helcim integration exists yet.${NC}"
fi

echo ""
echo "📊 Validation Summary:"

if [[ $FOUND_ISSUES -eq 0 ]]; then
    echo -e "${GREEN}✅ SUCCESS: No forbidden URLs found!${NC}"
    echo "🎉 Helcim URL validation passed."
    exit 0
else
    echo -e "${RED}❌ FAILURE: Found $FOUND_ISSUES forbidden URL(s)!${NC}"
    echo "🔧 Please update the codebase to use the canonical Helcim URL:"
    echo "   $CANONICAL_URL"
    echo ""
    echo "📖 For more information, see:"
    echo "   - docs/payment-system/helcim-integration.md"
    echo "   - docs/payment-system/validation-checklist.md"
    exit 1
fi