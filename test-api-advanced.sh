#!/bin/bash

# Advanced test script for Helcim API authentication debugging

# The token format seen in the logs
TOKEN_FROM_LOGS="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM\\*RFVwOpQjD_"

# A corrected token with $ character instead of \\*
TOKEN_WITH_DOLLAR="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM$apJQc*RFVwOpQjD_"

# The token without any escapes
TOKEN_CLEAN="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM*RFVwOpQjD_"

# The API endpoint
API_URL="https://api.helcim.com/v2/helcim-pay/initialize"

# JSON request body - add companyName to match real request
REQUEST_BODY='{"paymentType":"purchase","amount":1,"currency":"USD","companyName":"American Veterans Rebuilding"}'

echo "HTTP REQUEST DETAILS THAT WOULD BE USED:"
echo "POST $API_URL"
echo "Content-Type: application/json"
echo "Accept: application/json"
echo "Request Body: $REQUEST_BODY"
echo ""

echo "==== Testing token from logs ===="
echo "Token length: ${#TOKEN_FROM_LOGS} characters"
echo "First 10 chars: ${TOKEN_FROM_LOGS:0:10}..."
echo "Last 10 chars: ...${TOKEN_FROM_LOGS:${#TOKEN_FROM_LOGS}-10}"

RESPONSE=$(curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "api-token: $TOKEN_FROM_LOGS" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY")
echo "$RESPONSE" | jq || echo "$RESPONSE"

echo -e "\n==== Testing token with $ character (not truncated) ===="
echo "Token length: ${#TOKEN_WITH_DOLLAR} characters"
echo "First 10 chars: ${TOKEN_WITH_DOLLAR:0:10}..."
echo "Last 10 chars: ...${TOKEN_WITH_DOLLAR:${#TOKEN_WITH_DOLLAR}-10}"

RESPONSE=$(curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "api-token: $TOKEN_WITH_DOLLAR" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY")
echo "$RESPONSE" | jq || echo "$RESPONSE"

echo -e "\n==== Testing token with lowercase header name ===="
echo "Using token from logs but with lowercase 'api-token' header"

RESPONSE=$(curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "api-token: $TOKEN_FROM_LOGS" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY")
echo "$RESPONSE" | jq || echo "$RESPONSE"

echo -e "\n==== Testing if Helcim API is accessible at all ===="
echo "Making a basic OPTIONS request to check API availability"

curl -s -i -X OPTIONS "$API_URL" | head -20