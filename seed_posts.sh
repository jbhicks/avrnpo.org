#!/bin/bash
ADMIN_EMAIL=${PB_ADMIN_EMAIL:-admin@avrnpo.org}
ADMIN_PASSWORD=${PB_ADMIN_PASSWORD:-adminpassword123}

AUTH_RESPONSE=$(curl -s -X POST http://localhost:8090/api/collections/users/auth-with-password \
  -H "Content-Type: application/json" \
  -d "{\"identity\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo $AUTH_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
ADMIN_ID=$(echo $AUTH_RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

curl -s -X POST http://localhost:8090/api/collections/posts/records \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Markdown Guide\",\"slug\":\"markdown-guide\",\"excerpt\":\"A guide to Markdown\",\"content\":\"# Markdown Guide\n\nThis is a draft post.\",\"published\":false,\"author\":\"$ADMIN_ID\"}"

curl -s -X POST http://localhost:8090/api/collections/posts/records \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"New Features\",\"slug\":\"new-features\",\"excerpt\":\"Announcing new features\",\"content\":\"# New Features\n\nDraft announcement.\",\"published\":false,\"author\":\"$ADMIN_ID\"}"
