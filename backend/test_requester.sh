#!/bin/bash
BASE_URL="http://localhost:8080/api"

# Register new user (not member of any project)
NEW_USER=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"outsider@test.com","password":"Test123!@#","name":"Outsider"}')
echo "New user: $NEW_USER"

LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"outsider@test.com","password":"Test123!@#"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

# Try to assign task (should fail - not project member)
echo -e "\nTrying to assign (outsider)..."
ASSIGN=$(curl -s -X PUT "$BASE_URL/tasks/17/assign" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"assigned_to":"some-user"}')
echo "Response: $ASSIGN"

# Should get ACCESS_DENIED (403) or INVALID_ASSIGNEE (400)
