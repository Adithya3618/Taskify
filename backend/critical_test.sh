#!/bin/bash
BASE_URL="http://localhost:8080/api"

LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"Test123!@#"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
AUTH="Authorization: Bearer $TOKEN"

echo "=== Critical Test ==="

# Create project and stage
PROJ=$(curl -s -X POST "$BASE_URL/projects" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"Critical Test"}')
PROJ_ID=$(echo $PROJ | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

FINAL=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"name":"Final","position":1,"is_final":true}')
FINAL_ID=$(echo $FINAL | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "Final stage ID: $FINAL_ID"

# Create task in final stage
TASK=$(curl -s -X POST "$BASE_URL/projects/$PROJ_ID/stages/$FINAL_ID/tasks" \
  -H "Content-Type: application/json" -H "$AUTH" \
  -d '{"title":"Task","description":"Test"}')
echo "Task created: $TASK"

# Delete the stage (correct route: /stages/{id})
echo -e "\nDeleting stage $FINAL_ID..."
DELETE=$(curl -s -X DELETE "$BASE_URL/stages/$FINAL_ID" -H "$AUTH")
echo "Delete response: $DELETE"

# Check if task was deleted via CASCADE
echo -e "\nChecking if task was deleted..."
sqlite3 taskify.db "SELECT id, stage_id, title FROM tasks WHERE id IN (15, 16, 17, 18, 19, 20)" 2>/dev/null

# Check stats
STATS=$(curl -s "$BASE_URL/projects/$PROJ_ID/stats" -H "$AUTH")
echo "Stats: $STATS"

echo "=== Done ==="
