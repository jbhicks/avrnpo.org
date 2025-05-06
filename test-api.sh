#!/bin/bash

# Test script for Helcim API authentication debugging

# The token with escape characters as seen in logs
TOKEN_WITH_ESCAPES="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM\\*RFVwOpQjD_"

# The token with escapes removed (what we think should work)
TOKEN_WITHOUT_ESCAPES="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM*RFVwOpQjD_"

# The API endpoint
API_URL="https://api.helcim.com/v2/helcim-pay/initialize"

# JSON request body
REQUEST_BODY='{"paymentType":"test","amount":1,"currency":"USD"}'

echo "==== Testing with token containing backslash escapes ===="
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Api-Token: $TOKEN_WITH_ESCAPES" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY" | jq || echo "Error: jq not installed or response not valid JSON"

echo -e "\n==== Testing with token without backslash escapes ===="
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Api-Token: $TOKEN_WITHOUT_ESCAPES" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY" | jq || echo "Error: jq not installed or response not valid JSON"

echo -e "\nNOTE: If you see a success response with a checkoutToken from either request,"
echo "that indicates which token format is correct."