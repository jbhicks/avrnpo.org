#!/bin/bash

# Script for testing Helcim API authentication in local environment
echo "=== Helcim API Test - Local Environment ==="
echo "Testing API with different authentication headers"
echo ""

HOST="http://localhost:3001"
ENDPOINT="/api/diagnostics/helcim"
DEBUG_KEY="4c8b8a75-13d3-45c4-8f54-3d82830c16df"

echo "Starting test..."
echo "Using endpoint: ${HOST}${ENDPOINT}"
echo "Using debug key: ${DEBUG_KEY}"

# Check if DEBUG_ACCESS_KEY environment variable is set correctly
echo ""
echo "Environment check:"
echo "DEBUG_ACCESS_KEY environment variable is $(if [ -n "$DEBUG_ACCESS_KEY" ]; then echo "set to: ${DEBUG_ACCESS_KEY}"; else echo "not set"; fi)"
echo ""

# First request with detailed output and explicit verbose debug
echo "Running first curl command with verbose output..."
RESPONSE=$(curl -v "${HOST}${ENDPOINT}" \
  -H "X-Debug-Access: ${DEBUG_KEY}" 2>&1)
echo "Curl verbose output:"
echo "$RESPONSE"

# Process with jq if response is valid JSON
echo ""
echo "Processing response with jq..."
echo "$RESPONSE" | jq '
  .diagnostics.tokenInfo,
  .diagnostics.requestTests | 
  map_values({
    header: .requestInfo.headerName,
    status: .statusCode,
    success: .success,
    duration: .requestDuration
  })
' 2>/dev/null || echo "Error: Unable to process response with jq. Check if response is valid JSON."

echo ""
echo "Writing complete test output to test-results-local.json"
# Second request with file output and debug info
SAVE_RESPONSE=$(curl -v "${HOST}${ENDPOINT}" \
  -H "X-Debug-Access: ${DEBUG_KEY}" -o test-results-local.json 2>&1)
echo "Curl file save output:"
echo "$SAVE_RESPONSE"

# Check if file was created and has content
echo ""
echo "Checking result file:"
if [ -f "test-results-local.json" ]; then
  FILE_SIZE=$(wc -c < test-results-local.json)
  echo "File exists. Size: $FILE_SIZE bytes"
  if [ $FILE_SIZE -eq 0 ]; then
    echo "WARNING: File is empty!"
  else
    echo "File contains data."
    echo "First 100 characters of file content:"
    head -c 100 test-results-local.json
  fi
else
  echo "ERROR: File was not created!"
fi

echo ""
echo "=== Test complete ==="