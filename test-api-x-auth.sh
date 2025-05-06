#!/bin/bash

# Test script for Helcim API authentication with X-Auth-Token header

# The token from logs with backslash escapes
TOKEN_FROM_LOGS="ae!z7R_Yw9GiId!tq7FLwMccVpNz.Rh.hTGsGQ6Gk4Ui.bM\\*RFVwOpQjD_"

# The API endpoint
API_URL="https://api.helcim.com/v2/helcim-pay/initialize"

# JSON request body - minimal test
REQUEST_BODY='{"paymentType":"purchase","amount":1,"currency":"USD","companyName":"American Veterans Rebuilding"}'

echo "Testing Helcim API with X-Auth-Token header (as shown in OPTIONS response)"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-Auth-Token: $TOKEN_FROM_LOGS" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY" | jq || echo "Response is not valid JSON"

echo -e "\nTesting with js-token header (another allowed header from OPTIONS response)"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "js-token: $TOKEN_FROM_LOGS" \
  -H "Accept: application/json" \
  -d "$REQUEST_BODY" | jq || echo "Response is not valid JSON"

echo -e "\nNOTE: These tests use the token with escapes as seen in logs."