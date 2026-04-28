#!/bin/bash
BASE="http://localhost:8080/api"

# Get token
LOGIN=$(curl -s -X POST "$BASE/auth/login" -H "Content-Type: application/json" -d '{"email":"e2e@test.com","password":"Test123!@#"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Try non-existent task
curl -s -X PUT "$BASE/tasks/99999/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"assigned_to":"user-id"}'
